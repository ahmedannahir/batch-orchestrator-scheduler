package jobs

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/go-co-op/gocron"
)

var taskScheduler = gocron.NewScheduler(time.UTC)

func SubscribeJob(finished chan bool, launchMode string) {
	if launchMode == "--py-file" {
		log.Println("Benchmark started using", launchMode)
		taskScheduler.Every(30).Seconds().Do(func() {
			log.Println("job started")
			cmd := exec.Command("python", "jobs/scripts/transaction-file.py")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			log.Println(cmd.Run())
		})
	}

	if launchMode == "--py-db" {
		log.Println("Benchmark started using", launchMode)
		taskScheduler.Every(1).Minutes().Do(func() {
			log.Println("job started")
			cmd := exec.Command("python", "jobs/scripts/transaction-db.py")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			log.Println(cmd.Run())
		})
	}

	if launchMode == "--go-db" {
		log.Println("Benchmark started using", launchMode)
		taskScheduler.Every(10).Seconds().Do(func() {
			log.Println("job started")
			cmd := exec.Command("go", "run", "jobs/scripts/transaction-db.go")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			log.Println(cmd.Run())
		})
	}

	if launchMode == "--go-file" {
		log.Println("Benchmark started using", launchMode)
		taskScheduler.Every(10).Seconds().Do(func() {
			log.Println("job started")
			cmd := exec.Command("go", "run", "jobs/scripts/transaction-file.go")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			log.Println(cmd.Run())
		})
	}

	if launchMode == "--js-db" {
		log.Println("Benchmark started using", launchMode)
		taskScheduler.Every(30).Seconds().Do(func() {
			log.Println("job started")
			cmd := exec.Command("node", "jobs/scripts/transaction-db.js")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			log.Println(cmd.Run())
		})
	}

	if launchMode == "--js-file" {
		log.Println("Benchmark started using", launchMode)
		taskScheduler.Every(30).Seconds().Do(func() {
			log.Println("job started")
			cmd := exec.Command("node", "jobs/scripts/transaction-file.js")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			log.Println(cmd.Run())
		})
	}
}

func Start() {
	taskScheduler.StartAsync()
}
