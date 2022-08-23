package services

import (
	"errors"
	"gestion-batches/entities"
	"gestion-batches/entities/BatchStatus"
	"gestion-batches/handlers"
	"gestion-batches/jobs"
	"gestion-batches/models"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ExtractFile(key string, c *gin.Context) ([]byte, error) {
	fileHeader, err := c.FormFile(key)
	if err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(file)
}

func RunBatch(batchPath string, logFile *os.File, batchPrefix []string) {
	cmdParts := append(batchPrefix, batchPath)

	log.Println("Running the batch : " + batchPath + "...")
	var cmd *exec.Cmd
	cmd = exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Run()
}

func ScheduleBatch(batch entities.Batch, db *gorm.DB) error {
	log.Println("Scheduling the batch : " + batch.Url + "...")
	err := jobs.ScheduleBatch(batch, db)
	return err
}

func CreateLog(batch entities.Batch) (*os.File, error) {
	log.Println("Creating a logfile for : " + batch.Url + "...")
	return handlers.CreateLog(batch)
}

func ScheduleConsecBatches(batches []entities.Batch, db *gorm.DB) error {
	log.Println("Scheduling the consecutive batches...")
	err := jobs.ScheduleConsecBatches(batches, db)
	return err
}

// TO BE FIXED: Not working as intended - duplicating entries
func MatchBatchAndConfig(configs []models.Config, batchPaths *[]string) {
	log.Println("Matching every batch with its config...")
	var sorted []string
	for i := 0; i < len(configs); i++ {
		for _, batchPath := range *batchPaths {
			if strings.Contains(batchPath, configs[i].Script) {
				sorted = append(sorted, batchPath)
			}
		}
	}
	*batchPaths = sorted
}

func SaveBatch(config models.Config, batchPath string, prevBatchId *uint, db *gorm.DB, c *gin.Context) (entities.Batch, error) {
	batchName := c.PostForm("batchName")
	batchDesc := c.PostForm("batchDesc")
	profileIdStr := c.PostForm("profileId")
	profileId64, err := strconv.ParseUint(profileIdStr, 10, 64)
	if err != nil {
		return entities.Batch{}, err
	}
	profileId := uint(profileId64)

	batch := entities.Batch{
		Name:            batchName,
		Description:     batchDesc,
		Url:             batchPath,
		Timing:          config.Cron,
		Active:          true,
		Status:          BatchStatus.IDLE,
		Independant:     config.Independant,
		PrevBatchInput:  config.PrevBatchInput,
		PreviousBatchID: prevBatchId,
		ProfileID:       &profileId,
	}

	err = handlers.SaveBatch(&batch, db)
	if err != nil {
		return entities.Batch{}, err
	}

	return batch, nil
}

func SaveConsecBatches(configs []models.Config, batchesPaths []string, db *gorm.DB, c *gin.Context) ([]entities.Batch, error) {
	var batches []entities.Batch

	batchName := c.PostForm("batchName")
	batchDesc := c.PostForm("batchDesc")
	profileIdStr := c.PostForm("profileId")
	profileId64, err := strconv.ParseUint(profileIdStr, 10, 64)
	if err != nil {
		return []entities.Batch{}, err
	}
	profileId := uint(profileId64)

	for i := 0; i < len(configs); i++ {
		batch := entities.Batch{
			Name:           batchName,
			Description:    batchDesc,
			Url:            batchesPaths[i],
			Timing:         configs[i].Cron,
			Active:         true,
			Status:         BatchStatus.IDLE,
			Independant:    configs[i].Independant,
			PrevBatchInput: configs[i].PrevBatchInput,
			ProfileID:      &profileId,
		}

		batches = append(batches, batch)
	}

	err = handlers.SaveConsecBatches(&batches, batchesPaths, db)
	if err != nil {
		return nil, err
	}

	return batches, nil
}

func VerifyConfigsAndBatchesNumber(configs []models.Config, key string, c *gin.Context) error {
	form, _ := c.MultipartForm()
	batches := form.File[key]
	if len(configs) != len(batches) {
		return errors.New("Number of configs and batches is not the same")
	}
	return nil
}

func RunAfterBatch(id *uint, config models.Config, batch entities.Batch, db *gorm.DB) error {
	return jobs.RunAfterBatch(id, batch, db)
}

func ProcessBatchIdFromParam(key string, db *gorm.DB, c *gin.Context) (*uint, entities.Batch, error) {
	batchId64, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return nil, entities.Batch{}, err
	}

	batchId := uint(batchId64)
	var batch entities.Batch

	err = db.First(&batch, batchId).Error
	if err != nil {
		return nil, entities.Batch{}, err
	}

	return &batchId, batch, nil
}

func ProcessExecIdFromParam(key string, db *gorm.DB, c *gin.Context) (*uint, entities.Execution, error) {
	execId64, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return nil, entities.Execution{}, err
	}

	execId := uint(execId64)
	var execution entities.Execution

	err = db.First(&execution, execId).Error
	if err != nil {
		return nil, entities.Execution{}, err
	}

	return &execId, execution, nil
}

func UnzipBatch(batchPath string) (string, error) {
	log.Println("Unzipping batch : ", batchPath, "...")
	dst := strings.TrimSuffix(batchPath, ".zip")

	err := handlers.UnzipFile(batchPath, dst, os.ModePerm)

	return dst, err
}

func LoadBatchesFromDB(db *gorm.DB) error {
	var batches []entities.Batch

	err := db.Find(&batches).Error
	if err != nil {
		return err
	}

	for _, batch := range batches {
		if !batch.Active {
			log.Println("Batch : ", batch.Url, "inactive, thus not scheduled.")
			continue
		}

		if batch.PreviousBatchID == nil {
			err = jobs.ScheduleBatch(batch, db)
			if err == nil {
				log.Println("Scheduled batch : ", batch.Url)
			}
		} else {
			err = jobs.RunAfterBatch(batch.PreviousBatchID, batch, db)
			if err == nil {
				log.Println("Scheduled batch : ", batch.Url, " to run after batch ID : ", *batch.PreviousBatchID)
			}
		}
		if err != nil {
			log.Print("Error scheduling batch : ", batch.Url)
		}
	}

	return nil
}

func RunBatchById(batch entities.Batch, db *gorm.DB) error {
	return jobs.RunBatch(entities.Execution{}, batch, db)
}

func DisableBatch(batch entities.Batch, db *gorm.DB) error {
	if batch.Active == false {
		log.Println("Batch already inactive")
		return nil
	}

	batch.Active = false

	var batches []entities.Batch

	tx := db.Begin()

	subseqBatches, err := handlers.GetSubsequentBatches(batch, tx)
	if err != nil {
		return err
	}

	batches = append(batches, batch)
	batches = append(batches, subseqBatches...)

	for _, batch := range batches {
		batch.Active = false
	}
	err = tx.Save(&batches).Error
	if err != nil {
		log.Println("Error updating batches : ", err)
		tx.Rollback()
		return err
	}

	err = jobs.RemoveBatches(batches, db)
	if err != nil {
		log.Println("Error removing jobs from the scheduler : ", err)
		tx.Rollback()
		return err
	}

	log.Println("Jobs removed from the scheduler.")

	tx.Commit()

	return nil
}

func EnableBatch(batch entities.Batch, db *gorm.DB) error {
	if batch.Active == true {
		log.Println("Batch already active")
		return nil
	}

	batch.Active = true

	tx := db.Begin()

	err := tx.Save(&batch).Error
	if err != nil {
		log.Println("Error updating Batch active field : ", err)
		tx.Rollback()
		return err
	}
	log.Println("Batch active field updated : ", batch.Active)

	err = jobs.EnableBatch(batch, tx)
	if err != nil {
		log.Println("Error adding batch to the scheduler : ", err)
		tx.Rollback()
		return err
	}
	log.Println("Job added to the scheduler.")

	tx.Commit()

	return nil
}
