package jobs

import (
	"gestion-batches/handlers"
	"gestion-batches/models"
	"os/exec"
	"time"

	"github.com/go-co-op/gocron"
)

var scheduler = gocron.NewScheduler(time.UTC)

func ScheduleBatch(config models.Config, batchPath string, batchPrefix []string) error {
	_, err := scheduler.Cron(config.Cron).Do(func() {
		logFile, _ := handlers.CreateLog(batchPath)
		cmdParts := append(batchPrefix, batchPath)
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		cmd.Run()
	})
	if err == nil {
		scheduler.StartAsync()
	}
	return err
}
