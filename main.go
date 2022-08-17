package main

import (
	"gestion-batches/database"
	"gestion-batches/routers"
	"gestion-batches/services"
	"log"
)

func main() {
	db := database.Init()

	err := services.LoadBatchesFromDB(db)
	if err != nil {
		log.Panic(err)
	}

	routers.HandleRoutes(db)
}
