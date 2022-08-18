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
	} else {
		batch.Independant = true
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

	log.Print("Batch status update : ", (*batch).Status)
	log.Println("Execution created : ", *execution)
	tx.Commit()
	return nil
}

func UpdateExecutionAndBatchStatus(execution *entities.Execution, batch *entities.Batch, err error, db *gorm.DB) error {
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

	tx2 := db.Begin()

	var count int64
	err2 := tx2.Model(&entities.Execution{}).Where("batchId = ? AND status = ?", batch.ID, ExecutionStatus.RUNNING).Count(&count).Error
	if err2 != nil {
		log.Println("Error retrieving number of current executions running for the batch : ", err2)
		return err2
	}
	log.Println("Number of current execution of this batch : ", count)
	if count == 0 {
		batch.Status = BatchStatus.IDLE
		err3 := tx2.Save(batch).Error
		if err3 != nil {
			tx2.Rollback()
			log.Println("An error occured during updating batch status in db : ", err3)
			return err3
		}
	}
	tx2.Commit()

	log.Println("Batch status update : ", (*batch).Status)
	log.Println("Execution update : ", *execution)
	return nil
}
