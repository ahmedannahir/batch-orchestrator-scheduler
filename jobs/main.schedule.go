package jobs

import (
	"gestion-batches/entities"
	"gestion-batches/handlers"
	"gestion-batches/models"
	"log"
	"os/exec"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/gorm"
)

var scheduler = gocron.NewScheduler(time.UTC)

func runBatch(batchPrefix []string, batch entities.Batch, db *gorm.DB) error {
	log.Println("creating log for batch : ", batch.Url)
	logFile, _ := handlers.CreateLog(batch.Url)
	defer logFile.Close()

	execution := entities.Execution{
		Status:     entities.RUNNING,
		StartTime:  time.Now(),
		LogFileUrl: logFile.Name(),
		BatchID:    batch.ID,
	}

	tx1 := db.Begin()

	log.Println("Saving execution : ", execution, " to the database")
	results1 := tx1.Create(&execution)
	if results1.Error != nil {
		tx1.Rollback()
		log.Println("An error occured during saving exec to db : ", results1.Error)
	}

	tx1.Commit()

	log.Println("Running batch : " + batch.Url + "...")
	cmdParts := append(batchPrefix, batch.Url)
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	err := cmd.Run()

	now := time.Now()
	execution.EndTime = &now

	if err == nil {
		log.Println("The batch : " + batch.Url + " is done running.")
		execution.Status = entities.COMPLETED
		execution.ExitCode = "exit status 0"
	} else {
		log.Println("The batch : " + batch.Url + " threw an error.")
		execution.Status = entities.FAILED
		execution.ExitCode = err.Error()
	}

	tx2 := db.Begin()

	log.Println("Updating batch execution endtime, status and exit code...")
	results2 := tx2.Save(&execution)
	if results2.Error != nil {
		tx2.Rollback()
		log.Println("An error occured during updating exec in db : ", results2.Error)
	}

	tx2.Commit()

	return err
}

func ScheduleBatch(config models.Config, batch entities.Batch, batchPrefix []string, db *gorm.DB) error {
	job := scheduler.Cron(config.Cron)
	_, err := job.Do(func() {
		runBatch(batchPrefix, batch, db)
	})
	if err == nil {
		scheduler.StartAsync()
	}
	return err
}

func ConsecBatches(configs []models.Config, batchCmds [][]string, batches []entities.Batch, db *gorm.DB) error {
	var errors []error

	_, err := scheduler.Cron(configs[0].Cron).Do(func() {
		err1 := runBatch(batchCmds[0], batches[0], db)
		errors = append(errors, err1)

		for i := 1; i < len(configs); i++ {
			if (errors[len(errors)-1] == nil) || (errors[len(errors)-1] != nil && configs[i].Independant == true) {
				err2 := runBatch(batchCmds[i], batches[i], db)
				errors = append(errors, err2)
			}
		}
		log.Println("The consecutive batches are donne running.")
	})
	if err == nil {
		scheduler.StartAsync()
	}
	return err
}
