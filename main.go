package main

import (
	"gestion-batches/database"
	"gestion-batches/routers"
)

func main() {
	db := database.Init()

	routers.HandleRoutes(db)
}
