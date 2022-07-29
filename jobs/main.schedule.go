package jobs

import (
	"fmt"
	"gestion-batches/handlers"
	"gestion-batches/models"
	"log"
	"os/exec"
	"time"

	"github.com/go-co-op/gocron"
)

var scheduler = gocron.NewScheduler(time.UTC)

func ScheduleBatch(config models.Config, batchPath string, batchPrefix []string) error {
	job := scheduler.Cron(config.Cron)
	_, err := job.Do(func() {
		log.Println("creating log for batch : ", batchPath)
		logFile, _ := handlers.CreateLog(batchPath)
		defer logFile.Close()

		log.Println("Running batch : " + batchPath + "...")
		cmdParts := append(batchPrefix, batchPath)
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		cmd.Run()
		log.Println("The batch : " + batchPath + "is done running.")
	})
	if err == nil {
		scheduler.StartAsync()
	}
	return err
}

func ConsecBatches(configs []models.Config, batchCmds [][]string, batchPaths []string) error {
	var errors []error

	_, err := scheduler.Cron(configs[0].Cron).Do(func() {
		log.Println("Creating logfile for the sonsecutive batches ", batchPaths, "...")

		logFile, _ := handlers.CreateLog(batchPaths[0])
		defer logFile.Close()

		log.Println("Running batch 1 : " + batchPaths[0] + "...")
		logFile.WriteString(fmt.Sprintln("Batch 1 logs :"))

		cmdParts := append(batchCmds[0], batchPaths[0])
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		err1 := cmd.Run()
		log.Println("The batch : " + batchPaths[0] + " is done running.")

		errors = append(errors, err1)

		for i := 1; i < len(configs); i++ {
			logFile.WriteString(fmt.Sprintln("\n========================================="))
			logFile.WriteString(fmt.Sprintln("Batch ", i+1, " logs :"))

			if errors[len(errors)-1] == nil {
				log.Println("Running batch ", i+1, " : "+batchPaths[i]+"...")

				cmdParts := append(batchCmds[i], batchPaths[i])
				cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
				cmd.Stdout = logFile
				cmd.Stderr = logFile
				err := cmd.Run()
				if err == nil {
					log.Println("The batch ", i+1, " : "+batchPaths[i]+" is done running.")
				} else {
					log.Println("The batch ", i+1, " : "+batchPaths[i]+" threw an error.")
				}

				errors = append(errors, err)
			} else {
				logFile.WriteString(fmt.Sprintln("Batch ", i, "threw an error"))
			}
		}
		log.Println("The consecutive batches :", batchPaths, " are donne running.")
	})
	if err == nil {
		scheduler.StartAsync()
	}
	return err
}
