package main

import (
	"fmt"
	"log"
	"os"

	_ "batch/main/database"
	"batch/main/jobs"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	finished := make(chan bool)

	if len(os.Args) > 1 {
		if os.Args[1] == "--help" {
			fmt.Println("")
			fmt.Println("List of available modes:")
			fmt.Println("")
			fmt.Println("--js-file : execute batch from a file using JavaScript")
			fmt.Println("--js-db : execute batch from database using JavaScript")
			fmt.Println("--go-file : execute batch from a file using GO")
			fmt.Println("--go-db : execute batch from database using GO")
			fmt.Println("--py-file : execute batch from a file using Python")
			fmt.Println("--py-db : execute batch from database using Python")
		} else {
			for _, arg := range os.Args {
				jobs.SubscribeJob(finished, arg)
			}
			jobs.Start()
			<-finished
		}
	} else {
		log.Fatal("No launch mode argument provided. For more help use the command --help.")
	}

	log.Println("Execution ended")
}
