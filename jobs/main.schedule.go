package jobs

import (
	"fmt"
	"gestion-batches/entities"
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

func runBatch(execution *entities.Execution, lastPrevBatchExec entities.Execution, batch entities.Batch, db *gorm.DB) error {
	log.Println("creating log for batch : ", batch.Url)
	logFile, _ := handlers.CreateLog(batch)
	defer logFile.Close()

	errLogFile, _ := handlers.CreateErrLog(logFile.Name())
	defer errLogFile.Close()

	now := time.Now()

	execution.Status = entities.RUNNING
	execution.StartTime = &now
	execution.LogFileUrl = logFile.Name()
	tx := db.Begin()

	err := tx.Save(execution).Error
	if err != nil {
		tx.Rollback()
		log.Println("An error occured during updating exec in db : ", err)
		return err
	}

	tx.Commit()

	script := filepath.Join(batch.Url, "script.sh")

	var cmdParts []string
	cmdParts = append(cmdParts, "bash", script)

	if batch.PrevBatchInput && batch.PreviousBatchID != nil {
		cmdParts = append(cmdParts, lastPrevBatchExec.LogFileUrl)
	}

	log.Println("Running batch : " + batch.Url + "...")
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = logFile
	cmd.Stderr = errLogFile
	err1 := cmd.Run()

	errLogPath := errLogFile.Name()
	if err1 != nil {
		execution.ErrLogFileUrl = &errLogPath
	} else {
		errLogFile.Close()
		err3 := os.Remove(errLogFile.Name())
		if err3 != nil {
			log.Println("Error removing Error Log File : ", err3)
		}
	}

	err2 := handlers.UpdateExecution(execution, batch.Url, err1, db)
	if err2 != nil {
		return err2
	}

	return err1
}

func getPermissionToRun(batch entities.Batch, db *gorm.DB) (bool, entities.Execution) {
	if batch.PreviousBatchID == nil {
		log.Println("Permission to run : true | Independant : ", batch.Independant)
		return true, entities.Execution{}
	} else {
		var lastPrevBatchExec entities.Execution
		db.Where("status IN ?", []string{string(entities.COMPLETED), string(entities.FAILED)}).Last(&lastPrevBatchExec, "batchId = ?", batch.PreviousBatchID)
		now := time.Now()
		permission := now.Sub(*lastPrevBatchExec.EndTime).Seconds() < 10 && (lastPrevBatchExec.Status == entities.COMPLETED || batch.Independant)
		log.Println("Permission to run : ", permission, " | Independant : ", batch.Independant, " | PrevBatchID : ", batch.PreviousBatchID, " | LastPrevBatchExec : ", lastPrevBatchExec)
		return permission, lastPrevBatchExec
	}

}

func batchJobFunc(execution *entities.Execution, batch entities.Batch, db *gorm.DB) {
	permission, lastPrevBatchExec := getPermissionToRun(batch, db)
	if permission {
		runBatch(execution, lastPrevBatchExec, batch, db)
	} else {
		log.Println("Previous batch threw an error. The batch :", batch, " did not run")
	}
}

func twoConsecBatch(batches []entities.Batch, db *gorm.DB) error {
	execution := entities.Execution{
		Status:  entities.IDLE,
		BatchID: &batches[0].ID,
	}
	err := handlers.SaveExecution(&execution, db)
	if err != nil {
		return err
	}

	job, err := scheduler.Cron(batches[0].Timing).Tag(strconv.FormatUint(uint64(batches[0].ID), 10)).Do(func() {
		batchJobFunc(&execution, batches[0], db)
	})

	job.SetEventListeners(
		func() {

		},
		func() {
			err1 := scheduler.RunByTag(strconv.FormatUint(uint64(batches[1].ID), 10))
			if err1 != nil {
				log.Println("Error running subsequent batch ", batches[1].Url, " : ", err1)
			}

			// If Condition not mandatory but good guard nevertheless
			if job.NextRun().After(time.Now()) {
				execution = entities.Execution{
					Status:  entities.IDLE,
					BatchID: &batches[0].ID,
				}
				err := handlers.SaveExecution(&execution, db)
				if err != nil {
					fmt.Println("An ERROR happened while creating next execution", err)
					return
				}
			}
		})

	if err == nil {
		scheduler.StartAsync()
	}
	return err
}

func ScheduleBatch(batch entities.Batch, db *gorm.DB) error {
	execution := entities.Execution{
		Status:  entities.IDLE,
		BatchID: &batch.ID,
	}
	err := handlers.SaveExecution(&execution, db)
	if err != nil {
		return err
	}

	job, err := scheduler.Cron(batch.Timing).Tag(strconv.FormatUint(uint64(batch.ID), 10)).Do(func() {
		batchJobFunc(&execution, batch, db)
	})

	job.SetEventListeners(
		nil,
		func() {

			// If Condition not mandatory but good guard nevertheless
			if job.NextRun().After(time.Now()) {
				execution = entities.Execution{
					Status:  entities.IDLE,
					BatchID: &batch.ID,
				}
				err := handlers.SaveExecution(&execution, db)
				if err != nil {
					fmt.Println("An ERROR happened while creating next execution", err)
					return
				}
			}
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
	execution := entities.Execution{
		Status:  entities.IDLE,
		BatchID: &batch.ID,
	}
	err := handlers.SaveExecution(&execution, db)
	if err != nil {
		return err
	}

	scheduler.Cron("1 1 30 2 1").Tag(strconv.FormatUint(uint64(batch.ID), 10)).Do(func() {
		batchJobFunc(&execution, batch, db)
	})

	jobs, err := scheduler.FindJobsByTag(strconv.FormatUint(uint64(*id), 10))
	if err != nil {
		return err
	}

	jobs[0].SetEventListeners(
		nil,
		func() {
			err1 := scheduler.RunByTag(strconv.FormatUint(uint64(batch.ID), 10))
			if err1 != nil {
				log.Println("Error running subsequent batch ", batch.Url, " : ", err1)
			}

			// If Condition not mandatory but good guard nevertheless
			if jobs[0].NextRun().After(time.Now()) {
				execution = entities.Execution{
					Status:  entities.IDLE,
					BatchID: &batch.ID,
				}
				err := handlers.SaveExecution(&execution, db)
				if err != nil {
					fmt.Println("An ERROR happened while creating next execution", err)
					return
				}
			}
		})

	return nil
}
