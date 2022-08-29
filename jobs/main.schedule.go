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
	if batch.Args != nil {
		cmdParts = append(cmdParts, *batch.Args)
	}

	if batch.PrevBatchInput && batch.PreviousBatchID != nil {
		cmdParts = append(cmdParts, *lastPrevBatchExec.LogFileUrl)
	}

	log.Println("Running batch : " + batch.Url + "...")
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = logFile
	cmd.Stderr = errLogFile
	errExec := cmd.Run()

	errLogPath := errLogFile.Name()
	if errExec != nil {
		execution.ErrLogFileUrl = &errLogPath
	} else {
		errLogFile.Close()
		err := os.Remove(errLogFile.Name())
		if err != nil {
			log.Println("Error removing Error Log File : ", err)
			return err
		}
	}

	err = handlers.UpdateExecutionAndBatchStatus(&execution, &batch, errExec, db)
	if err != nil {
		return err
	}

	err = handlers.SendExecInfosMail(batch, execution, db)
	if err != nil {
		return err
	}

	return nil
}

func abortBatch(batch entities.Batch, db *gorm.DB) error {
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

	err = handlers.SendExecInfosMail(batch, execution, db)
	if err != nil {
		return err
	}

	return nil
}

func getPermissionToRun(batch entities.Batch, db *gorm.DB) (bool, entities.Execution, error) {
	if batch.PreviousBatchID == nil {
		log.Println("Permission to run : true | Independant : ", batch.Independant)
		return true, entities.Execution{}, nil
	} else {
		var lastPrevBatchExec entities.Execution

		err := db.Where("batchId = ? AND status IN ?", batch.PreviousBatchID, []string{string(ExecutionStatus.COMPLETED), string(ExecutionStatus.FAILED), string(ExecutionStatus.ABORTED)}).Last(&lastPrevBatchExec).Error
		if err != nil {
			if errors.Is(gorm.ErrRecordNotFound, err) {
				log.Println("Previous batch's execution not found !?")
			}
			return false, entities.Execution{}, err
		}

		// comparing duration between execution can be as narrow as we want
		now := time.Now()
		permission := now.Sub(*lastPrevBatchExec.EndTime).Seconds() < 5 && (lastPrevBatchExec.Status == ExecutionStatus.COMPLETED || batch.Independant)
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
		abortBatch(batch, db)
	}

	return nil
}

func twoConsecBatch(batches []entities.Batch, db *gorm.DB) error {
	job, err := scheduler.Cron(batches[0].Timing).Tag(strconv.FormatUint(uint64(batches[0].ID), 10)).Do(func() {
		batchJobFunc(batches[0], db)
	})
	if err != nil {
		return err
	}

	job.SetEventListeners(
		nil,
		func() {
			err := scheduler.RunByTag(strconv.FormatUint(uint64(batches[1].ID), 10))
			if err != nil {
				log.Println("Error running subsequent batch ", batches[1].Url, " : ", err)
			}
		},
	)

	return nil
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
			return err
		}
	}

	err := ScheduleBatch(batches[len(batches)-1], db)
	if err != nil {
		log.Println("error scheduling subsequent batch ", len(batches), " : ", err)
		return err
	}

	return nil
}

func RunAfterBatch(id *uint, batch entities.Batch, db *gorm.DB) error {
	scheduler.Cron("1 1 30 2 1").Tag(strconv.FormatUint(uint64(batch.ID), 10)).Do(func() {
		batchJobFunc(batch, db)
	})

	jobs, err := scheduler.FindJobsByTag(strconv.FormatUint(uint64(*id), 10))
	if err != nil {
		return err
	}

	jobs[0].SetEventListeners(
		nil,
		func() {
			err := scheduler.RunByTag(strconv.FormatUint(uint64(batch.ID), 10))
			if err != nil {
				log.Println("Error running subsequent batch ", batch.Url, " : ", err)
			}
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func RemoveBatches(batches []entities.Batch, db *gorm.DB) error {
	if batches[0].PreviousBatchID != nil {
		var prevBatch entities.Batch
		err := db.First(&prevBatch, batches[0].PreviousBatchID).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("Error retrieving previous batch : ", err)
			return err
		}
		jobs, err := scheduler.FindJobsByTag(strconv.FormatUint(uint64(prevBatch.ID), 10))
		if err != nil {
			log.Print("Error retrieving jobs from the scheduler : ", err)
			return err
		}
		jobs[0].SetEventListeners(nil, nil)
	}

	for _, batch := range batches {
		if batch.Active {
			err := scheduler.RemoveByTag(strconv.FormatUint(uint64(batch.ID), 10))
			if err != nil {
				log.Println("Error removing batch from scheduler : ", batch.Url)
				return err
			}
			log.Println("Removed from the scheduler the batch : ", batch.Url)
		} else {
			log.Println("Batch already inactive and removed from the scheduler : ", batch.Url)
		}
	}
	return nil
}

func EnableBatch(batch entities.Batch, tx *gorm.DB, db *gorm.DB) error {
	err := ScheduleBatch(batch, db)
	if err != nil {
		return err
	}

	var subseqBatch entities.Batch
	err = tx.Where("previousBatchId = ?", batch.ID).First(&subseqBatch).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) && subseqBatch.Active {
		subseqJobs, err := scheduler.FindJobsByTag(strconv.FormatUint(uint64(batch.ID), 10))
		if err != nil {
			return err
		}

		subseqJobs[0].SetEventListeners(
			nil,
			func() {
				err := scheduler.RunByTag(strconv.FormatUint(uint64(subseqBatch.ID), 10))
				if err != nil {
					log.Println("Error running subsequent batch ", subseqBatch.Url, " : ", err)
				}
			},
		)
	}

	if batch.PreviousBatchID != nil {
		var prevBatch entities.Batch
		err := tx.First(&prevBatch, batch.PreviousBatchID).Error
		if err != nil {
			return err
		}

		if prevBatch.Active {
			prevJobs, err := scheduler.FindJobsByTag(strconv.FormatUint(uint64(prevBatch.ID), 10))
			if err != nil {
				return err
			}

			prevJobs[0].SetEventListeners(
				nil,
				func() {
					err := scheduler.RunByTag(strconv.FormatUint(uint64(batch.ID), 10))
					if err != nil {
						log.Println("Error running subsequent batch ", batch.Url, " : ", err)
					}
				},
			)
		}
	}

	return nil
}
