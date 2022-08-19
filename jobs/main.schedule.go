package jobs

import (
	"errors"
	"gestion-batches/entities"
	"gestion-batches/entities/BatchStatus"
	"gestion-batches/entities/ExecutionStatus"
	"gestion-batches/handlers"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/gorm"
)

var scheduler = gocron.NewScheduler(time.UTC)

func RunBatch(lastPrevBatchExec entities.Execution, batch entities.Batch, db *gorm.DB) error {
	log.Println("creating log for batch : ", batch.Url)
	logFile, _ := handlers.CreateLog(batch)
	defer logFile.Close()

	errLogFile, _ := handlers.CreateErrLog(logFile.Name())
	defer errLogFile.Close()

	now := time.Now()

	logFileUrl := logFile.Name()

	execution := entities.Execution{
		Status:     ExecutionStatus.RUNNING,
		StartTime:  &now,
		LogFileUrl: &logFileUrl,
		BatchID:    &batch.ID,
	}

	batch.Status = BatchStatus.RUNNING

	err := handlers.SaveExecutionAndBatchStatus(&execution, &batch, db)
	if err != nil {
		log.Println("Error saving execution and updating batch : ", err)
		return err
	}

	script := filepath.Join(batch.Url, "script.sh")

	var cmdParts []string
	cmdParts = append(cmdParts, "bash", script)

	if batch.PrevBatchInput && batch.PreviousBatchID != nil {
		cmdParts = append(cmdParts, *lastPrevBatchExec.LogFileUrl)
	}

	log.Println("Running batch : " + batch.Url + "...")
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = logFile
	cmd.Stderr = errLogFile
	err2 := cmd.Run()

	errLogPath := errLogFile.Name()
	if err2 != nil {
		execution.ErrLogFileUrl = &errLogPath
	} else {
		errLogFile.Close()
		err3 := os.Remove(errLogFile.Name())
		if err3 != nil {
			log.Println("Error removing Error Log File : ", err3)
			return err3
		}
	}

	err4 := handlers.UpdateExecutionAndBatchStatus(&execution, &batch, err2, db)
	if err4 != nil {
		return err4
	}

	return err2
}

func getPermissionToRun(batch entities.Batch, db *gorm.DB) (bool, entities.Execution, error) {
	if batch.PreviousBatchID == nil {
		log.Println("Permission to run : true | Independant : ", batch.Independant)
		return true, entities.Execution{}, nil
	} else {
		var lastPrevBatchExec entities.Execution

		err1 := db.Where("batchId = ? AND status IN ?", batch.PreviousBatchID, []string{string(ExecutionStatus.COMPLETED), string(ExecutionStatus.FAILED), string(ExecutionStatus.ABORTED)}).Last(&lastPrevBatchExec).Error
		if err1 != nil {
			if errors.Is(gorm.ErrRecordNotFound, err1) {
				log.Println("Previous batch's execution not found !?")
			}
			return false, entities.Execution{}, err1
		}

		// comparing duration between execution can be as narrow as we want
		now := time.Now()
		permission := now.Sub(*lastPrevBatchExec.EndTime).Seconds() < 1 && (lastPrevBatchExec.Status == ExecutionStatus.COMPLETED || batch.Independant)
		log.Println("Permission to run : ", permission, " | Independant : ", batch.Independant, " | PrevBatchID : ", batch.PreviousBatchID, " | LastPrevBatchExec : ", lastPrevBatchExec)
		return permission, lastPrevBatchExec, nil
	}

}

func batchJobFunc(batch entities.Batch, db *gorm.DB) error {
	permission, lastPrevBatchExec, err := getPermissionToRun(batch, db)
	if err != nil {
		return err
	}
	if permission {
		RunBatch(lastPrevBatchExec, batch, db)
	} else {
		now := time.Now()
		execution := entities.Execution{
			Status:    ExecutionStatus.ABORTED,
			StartTime: &now,
			EndTime:   &now,
			BatchID:   &batch.ID,
		}

		err := db.Create(&execution).Error
		if err != nil {
			log.Println("Error creating ABORTED execution : ", err)
			return err
		}

		log.Println("Previous batch threw an error. The batch :", batch, " is aborted.")
		log.Println("Execution : ", execution, " saved to the database.")
	}

	return nil
}

func twoConsecBatch(batches []entities.Batch, db *gorm.DB) error {
	job, err := scheduler.Cron(batches[0].Timing).Tag(strconv.FormatUint(uint64(batches[0].ID), 10)).Do(func() {
		batchJobFunc(batches[0], db)
	})

	job.SetEventListeners(
		nil,
		func() {
			err1 := scheduler.RunByTag(strconv.FormatUint(uint64(batches[1].ID), 10))
			if err1 != nil {
				log.Println("Error running subsequent batch ", batches[1].Url, " : ", err1)
			}
		},
	)

	if err == nil {
		scheduler.StartAsync()
	}
	return err
}

func ScheduleBatch(batch entities.Batch, db *gorm.DB) error {
	_, err := scheduler.Cron(batch.Timing).Tag(strconv.FormatUint(uint64(batch.ID), 10)).Do(func() {
		batchJobFunc(batch, db)
	})
	if err == nil {
		scheduler.StartAsync()
	}
	return err
}

func ScheduleConsecBatches(batches []entities.Batch, db *gorm.DB) error {
	for i := 0; i < len(batches)-1; i++ {
		err := twoConsecBatch(batches[i:i+2], db)
		if err != nil {
			log.Println("error scheduling subsequent batch ", i+1, " : ", err)
		}
	}

	err := ScheduleBatch(batches[len(batches)-1], db)
	if err != nil {
		log.Println("error scheduling subsequent batch ", len(batches), " : ", err)
	}

	return nil
}

func RunAfterBatch(id *uint, batch entities.Batch, db *gorm.DB) error {
	scheduler.Cron("1 1 30 2 1").Tag(strconv.FormatUint(uint64(batch.ID), 10)).Do(func() {
		batchJobFunc(batch, db)
	})

	jobs, err := scheduler.FindJobsByTag(strconv.FormatUint(uint64(*id), 10))
	jobs[0].SetEventListeners(
		nil,
		func() {
			err1 := scheduler.RunByTag(strconv.FormatUint(uint64(batch.ID), 10))
			if err1 != nil {
				log.Println("Error running subsequent batch ", batch.Url, " : ", err1)
			}
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func DisableBatch(batch entities.Batch, db *gorm.DB) error {
	return scheduler.RemoveByTag(strconv.FormatUint(uint64(batch.ID), 10))
}
