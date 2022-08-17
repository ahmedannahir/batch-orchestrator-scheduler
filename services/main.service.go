package services

import (
	"errors"
	"gestion-batches/entities"
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
	fileHeader, err1 := c.FormFile(key)
	if err1 != nil {
		return nil, err1
	}

	file, err2 := fileHeader.Open()
	if err2 != nil {
		return nil, err1
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

	batch := entities.Batch{
		Name:            batchName,
		Description:     batchDesc,
		Url:             batchPath,
		Timing:          config.Cron,
		Independant:     config.Independant,
		PrevBatchInput:  config.PrevBatchInput,
		PreviousBatchID: prevBatchId,
	}

	err := handlers.SaveBatch(&batch, db)
	if err != nil {
		return entities.Batch{}, err
	}

	return batch, nil
}

func SaveConsecBatches(configs []models.Config, batchesPaths []string, db *gorm.DB, c *gin.Context) ([]entities.Batch, error) {
	var batches []entities.Batch

	batchName := c.PostForm("batchName")
	batchDesc := c.PostForm("batchDesc")

	for i := 0; i < len(configs); i++ {
		batch := entities.Batch{
			Timing:         configs[i].Cron,
			Name:           batchName,
			Description:    batchDesc,
			Independant:    configs[i].Independant,
			PrevBatchInput: configs[i].PrevBatchInput,
			Url:            batchesPaths[i],
		}

		batches = append(batches, batch)
	}

	err := handlers.SaveConsecBatches(&batches, batchesPaths, db)
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

	err1 := db.First(&batch, batchId).Error
	if err1 != nil {
		return nil, entities.Batch{}, err1
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

	err1 := db.First(&execution, execId).Error
	if err1 != nil {
		return nil, entities.Execution{}, err1
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
		var err error

		if batch.PreviousBatchID == nil {
			err = jobs.ScheduleBatch(batch, db)
		} else {
			err = jobs.RunAfterBatch(batch.PreviousBatchID, batch, db)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func RunBatchById(batch entities.Batch, db *gorm.DB) error {
	execution := entities.Execution{
		BatchID: &batch.ID,
	}

	err := db.Create(&execution).Error
	if err != nil {
		return err
	}

	return jobs.RunBatch(&execution, entities.Execution{}, batch, db)
}
