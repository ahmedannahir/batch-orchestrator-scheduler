package handlers

import (
	"gestion-batches/entities"
	"log"
	"time"

	"gorm.io/gorm"
)

func SaveBatch(config *entities.Config, batch *entities.Batch, db *gorm.DB) error {
	log.Println("Saving config and batch in the database...")
	tx := db.Begin()

	err1 := tx.Create(config).Error
	if err1 != nil {
		tx.Rollback()
		log.Println("An error has occured. Config and batch not saved to db : ", err1)
		return err1
	}

	batch.ConfigID = &config.ID

	err2 := tx.Create(batch).Error
	if err2 != nil {
		tx.Rollback()
		log.Println("An error has occured. Config and batch not saved to db : ", err2)
		return err2
	}

	now := time.Now()
	execution := entities.Execution{
		Status:    entities.IDLE,
		StartTime: now,
		EndTime:   &now,
		BatchID:   &batch.ID,
	}

	err3 := tx.Create(&execution).Error
	if err3 != nil {
		tx.Rollback()
		log.Println("An error has occured. Config and batch not saved to db : ", err3)
		return err3
	}

	tx.Commit()
	log.Println("Batch and its config saved to db : ", *config, *batch)

	return nil
}

func SaveConsecBatches(config *entities.Config, batches *[]entities.Batch, batchesPaths []string, db *gorm.DB) error {
	log.Println("Saving config and the batches in the database...")
	tx := db.Begin()

	err1 := tx.Create(config).Error
	if err1 != nil {
		log.Println("An error has occured. Config and batches not saved to db : ", err1)
		tx.Rollback()
		return err1
	}

	err2 := tx.Create(batches).Error
	if err2 != nil {
		log.Println("An error has occured. Config and batches not saved to db : ", err2)
		tx.Rollback()
		return err2
	}

	now := time.Now()
	executions := []entities.Execution{
		{
			Status:    entities.IDLE,
			StartTime: now,
			EndTime:   &now,
			BatchID:   &(*batches)[0].ID,
		},
	}
	for i := 1; i < len(batchesPaths); i++ {
		(*batches)[i].PreviousBatchID = &(*batches)[i-1].ID

		executions = append(executions, entities.Execution{
			Status:    entities.IDLE,
			StartTime: now,
			EndTime:   &now,
			BatchID:   &(*batches)[i].ID,
		})
	}

	err3 := tx.Save(&batches).Error
	if err3 != nil {
		log.Println("An error has occured. Config and batches not saved to db : ", err3)
		tx.Rollback()
		return err3
	}

	err4 := tx.Save(&executions).Error
	if err4 != nil {
		log.Println("An error has occured. Config and batches not saved to db : ", err4)
		tx.Rollback()
		return err4
	}

	tx.Commit()
	log.Println("Batches and config saved to db : ", config, batches)

	return nil
}

func SaveExecution(execution *entities.Execution, db *gorm.DB) error {
	tx := db.Begin()

	err := tx.Create(&execution).Error
	if err != nil {
		tx.Rollback()
		log.Println("An error occured during saving exec to db : ", err)
		return err
	}

	tx.Commit()
	log.Println("Execution : ", execution, " saved to the database")

	return nil
}

func UpdateExecution(execution *entities.Execution, batchUrl string, err error, db *gorm.DB) error {
	now := time.Now()
	execution.EndTime = &now

	if err == nil {
		log.Println("The batch : ", batchUrl, " is done running.")
		execution.Status = entities.COMPLETED
		execution.ExitCode = "exit status 0"
	} else {
		log.Println("The batch : ", batchUrl, " threw an error.")
		execution.Status = entities.FAILED
		execution.ExitCode = err.Error()
	}

	log.Println("Execution update : ", *execution)

	tx := db.Begin()

	err1 := tx.Save(execution).Error
	if err1 != nil {
		tx.Rollback()
		log.Println("An error occured during updating exec in db : ", err1)
		return err1
	}

	tx.Commit()
	log.Println("Batch execution endtime, status and exit code updated.")

	return nil
}
