package handlers

import (
	"gestion-batches/entities"
	"gestion-batches/entities/BatchStatus"
	"gestion-batches/entities/ExecutionStatus"
	"log"
	"time"

	"gorm.io/gorm"
)

func SaveBatch(batch *entities.Batch, db *gorm.DB) error {
	log.Println("Saving config and batch in the database...")
	if batch.PreviousBatchID != nil {
		batch.Timing = "1 1 30 2 1"
	}

	tx := db.Begin()

	err2 := tx.Create(batch).Error
	if err2 != nil {
		tx.Rollback()
		log.Println("An error has occured. Config and batch not saved to db : ", err2)
		return err2
	}

	tx.Commit()
	log.Println("Batch saved to db : ", *batch)

	return nil
}

func SaveConsecBatches(batches *[]entities.Batch, batchesPaths []string, db *gorm.DB) error {
	log.Println("Saving config and the batches in the database...")
	tx := db.Begin()

	err2 := tx.Create(batches).Error
	if err2 != nil {
		log.Println("An error has occured. Config and batches not saved to db : ", err2)
		tx.Rollback()
		return err2
	}
	for i := 1; i < len(*batches); i++ {
		(*batches)[i].PreviousBatchID = &(*batches)[i-1].ID
	}

	err3 := tx.Save(&batches).Error
	if err3 != nil {
		log.Println("An error has occured. Config and batches not saved to db : ", err3)
		tx.Rollback()
		return err3
	}

	tx.Commit()
	log.Println("Batches saved to db : ", *batches)

	return nil
}

func SaveExecutionAndBatchStatus(execution *entities.Execution, batch *entities.Batch, db *gorm.DB) error {
	tx := db.Begin()

	err := tx.Create(execution).Error
	if err != nil {
		tx.Rollback()
		log.Println("An error occured during updating exec in db : ", err)
		return err
	}

	err = tx.Save(batch).Error
	if err != nil {
		tx.Rollback()
		log.Println("An error occured during updating batch status in db : ", err)
		return err
	}

	tx.Commit()

	return nil
}

func UpdateExecutionAndBatchStatus(execution *entities.Execution, batch *entities.Batch, err error, db *gorm.DB) error {
	batch.Status = BatchStatus.IDLE

	now := time.Now()
	execution.EndTime = &now

	if err == nil {
		log.Println("The batch : ", batch.Url, " is done running.")
		execution.Status = ExecutionStatus.COMPLETED
		execution.ExitCode = "exit status 0"
	} else {
		log.Println("The batch : ", batch.Url, " threw an error.")
		execution.Status = ExecutionStatus.FAILED
		execution.ExitCode = err.Error()
	}

	log.Println("Batch update : ", *batch)
	log.Println("Execution update : ", *execution)

	tx := db.Begin()

	err2 := tx.Save(batch).Error
	if err2 != nil {
		tx.Rollback()
		log.Println("An error occured during updating batch status in db : ", err2)
		return err2
	}

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
