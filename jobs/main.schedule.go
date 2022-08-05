package jobs

import (
	"gestion-batches/entities"
	"gestion-batches/handlers"
	"gestion-batches/models"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/gorm"
)

var scheduler = gocron.NewScheduler(time.UTC)

func runBatch(batchPrefix []string, batch entities.Batch, db *gorm.DB) error {
	log.Println("creating log for batch : ", batch.Url)
	logFile, _ := handlers.CreateLog(batch.Url)
	defer logFile.Close()

	errLogFile, _ := handlers.CreateErrLog(logFile.Name())
	defer errLogFile.Close()

	execution := entities.Execution{
		Status:     entities.RUNNING,
		StartTime:  time.Now(),
		LogFileUrl: logFile.Name(),
		BatchID:    &batch.ID,
	}
	err := handlers.SaveExecution(&execution, db)
	if err != nil {
		return err
	}

	log.Println("Running batch : " + batch.Url + "...")
	cmdParts := append(batchPrefix, batch.Url)
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

	err2 := handlers.UpdateExecution(&execution, batch.Url, err1, db)
	if err2 != nil {
		return err2
	}

	return err1
}

func getPermissionToRun(config models.Config, batch entities.Batch, db *gorm.DB) bool {
	permission := false

	if batch.PreviousBatchID == nil || config.Independant {
		permission = true
	} else {
		var lastPrevBatchExec entities.Execution
		db.Last(&lastPrevBatchExec, "batch_id = ?", batch.PreviousBatchID)
		permission = time.Now().Sub(*lastPrevBatchExec.EndTime).Seconds() < 10 && lastPrevBatchExec.Status == entities.COMPLETED
	}

	return permission
}

func twoConsecBatch(configs []models.Config, batches []entities.Batch, batchPrefixes [][]string, db *gorm.DB) error {
	job, err := scheduler.Cron(configs[0].Cron).Tag(strconv.FormatUint(uint64(batches[0].ID), 10)).Do(func() {
		permission := getPermissionToRun(configs[0], batches[0], db)
		if permission {
			runBatch(batchPrefixes[0], batches[0], db)
		} else {
			log.Println("Previous batch threw an error. The batch :", batches[0].Url, " did not run")
		}
	})

	job.SetEventListeners(nil, func() {
		err1 := scheduler.RunByTag(strconv.FormatUint(uint64(batches[1].ID), 10))
		if err1 != nil {
			log.Println("Error running subsequent batch ", batches[1].Url, " : ", err1)
		}
	})

	if err == nil {
		scheduler.StartAsync()
	}
	return err
}

func ScheduleBatch(config models.Config, batch entities.Batch, batchPrefix []string, db *gorm.DB) error {
	_, err := scheduler.Cron(config.Cron).Tag(strconv.FormatUint(uint64(batch.ID), 10)).Do(func() {
		runBatch(batchPrefix, batch, db)
	})
	if err == nil {
		scheduler.StartAsync()
	}
	return err
}

func ScheduleConsecBatches(configs []models.Config, batchCmds [][]string, batches []entities.Batch, db *gorm.DB) error {
	for i := 0; i < len(configs)-1; i++ {
		err := twoConsecBatch(configs[i:i+2], batches[i:i+2], batchCmds[i:i+2], db) // sends i and i+1 elements of the slice, i+2 not included
		if err != nil {
			log.Println("error scheduling subsequent batch ", i+1, " : ", err)
		}
	}

	err := ScheduleBatch(configs[len(configs)-1], batches[len(configs)-1], batchCmds[len(configs)-1], db)
	if err != nil {
		log.Println("error scheduling subsequent batch ", len(configs), " : ", err)
	}

	return nil
}

func RunAfterBatch(id string, config models.Config, batch entities.Batch, batchPrefix []string, db *gorm.DB) error {
	scheduler.Cron("1 1 30 2 1").Tag(strconv.FormatUint(uint64(batch.ID), 10)).Do(func() {
		permission := getPermissionToRun(config, batch, db)
		if permission {
			runBatch(batchPrefix, batch, db)
		} else {
			log.Println("Previous batch threw an error. The batch :", batch, " did not run")
		}
	})

	jobs, err := scheduler.FindJobsByTag(id)
	if err != nil {
		return err
	}

	jobs[0].SetEventListeners(nil, func() {
		err1 := scheduler.RunByTag(strconv.FormatUint(uint64(batch.ID), 10))
		if err1 != nil {
			log.Println("Error running subsequent batch ", batch.Url, " : ", err1)
		}
	})

	return nil
}
