package services

import (
	"errors"
	"fmt"
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

func ExtractMultiLangsPref(configs []models.Config) ([][]string, [][]string, error) {
	depCmds := make([][]string, len(configs))
	batCmds := make([][]string, len(configs))
	for i, config := range configs {
		depPrefix, batPrefix, err := ExtractLanguagePrefixes(config)
		if err != nil {
			return nil, nil, err
		}
		depCmds[i] = append(depCmds[i], depPrefix...)
		batCmds[i] = append(batCmds[i], batPrefix...)
	}

	return depCmds, batCmds, nil
}

func InstallDependencies(dependencies []string, logFile *os.File, installDependencyPrefix []string) error {
	if len(dependencies) > 0 {
		cmdParts := append(installDependencyPrefix, "dep_placeholder")

		for _, dependency := range dependencies {
			cmdParts[len(cmdParts)-1] = dependency

			log.Println("Installing dependency : " + dependency + "...")
			cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
			cmd.Stdout = logFile
			cmd.Stderr = logFile
			err := cmd.Run()
			return err
		}
	}

	return nil
}

func ExtractLanguagePrefixes(config models.Config) ([]string, []string, error) {
	log.Println("Extracting commands prefixes for the script : " + config.Script + "...")
	instDepPrefix := []string{}
	runBatchPrefix := []string{}
	switch config.Language {
	case "GO":
		{
			instDepPrefix = append(instDepPrefix, "go", "get")
			runBatchPrefix = append(runBatchPrefix, "go", "run")
		}
	case "JAVASCRIPT":
		{
			instDepPrefix = append(instDepPrefix, "npm", "install")
			runBatchPrefix = append(runBatchPrefix, "node")
		}
	case "PYTHON":
		{
			instDepPrefix = append(instDepPrefix, "pip", "install")
			runBatchPrefix = append(runBatchPrefix, "python")
		}
	default:
		{
			return nil, nil, errors.New(fmt.Sprint("This language's configuration is not available at the moment."))
		}
	}

	return instDepPrefix, runBatchPrefix, nil
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

func ScheduleBatch(config models.Config, batch entities.Batch, batchPrefix []string, db *gorm.DB) error {
	log.Println("Scheduling the batch : " + batch.Url + "...")
	err := jobs.ScheduleBatch(config, batch, batchPrefix, db)
	return err
}

func CreateLog(batchPath string) (*os.File, error) {
	log.Println("Creating a logfile for : " + batchPath + "...")
	return handlers.CreateLog(batchPath)
}

func ScheduleConsecBatches(configs []models.Config, batches []entities.Batch, db *gorm.DB) error {
	log.Println("Scheduling the consecutive batches...")
	err := jobs.ScheduleConsecBatches(configs, batches, db)
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

func SaveBatch(configPath string, batchPath string, prevBatchId *uint, db *gorm.DB, c *gin.Context) (entities.Batch, entities.Config, error) {
	configName := c.PostForm("configName")
	batchName := c.PostForm("batchName")
	batchDesc := c.PostForm("batchDesc")

	config := entities.Config{
		Name: configName,
		Url:  configPath,
	}

	batch := entities.Batch{
		Name:            batchName,
		Description:     batchDesc,
		Url:             batchPath,
		PreviousBatchID: prevBatchId,
	}

	err := handlers.SaveBatch(&config, &batch, db)
	if err != nil {
		return entities.Batch{}, entities.Config{}, err
	}

	return batch, config, nil
}

func SaveConsecBatches(configs []models.Config, configPath string, batchesPaths []string, db *gorm.DB, c *gin.Context) ([]entities.Batch, entities.Config, error) {
	var batches []entities.Batch

	configName := c.PostForm("configName")
	batchName := c.PostForm("batchName")
	batchDesc := c.PostForm("batchDesc")

	config := entities.Config{
		Name: configName,
		Url:  configPath,
	}

	for i := 0; i < len(configs); i++ {
		batch := entities.Batch{
			Timing:      configs[i].Cron,
			Name:        batchName,
			Description: batchDesc,
			Url:         batchesPaths[i],
			ConfigID:    &config.ID,
		}

		batches = append(batches, batch)
	}

	err := handlers.SaveConsecBatches(&config, &batches, batchesPaths, db)
	if err != nil {
		return nil, entities.Config{}, err
	}

	return batches, config, nil
}

func VerifyConfigsAndBatchesNumber(configs []models.Config, key string, c *gin.Context) error {
	form, _ := c.MultipartForm()
	batches := form.File[key]
	if len(configs) != len(batches) {
		return errors.New("Number of configs and batches is not the same")
	}
	return nil
}

func RunAfterBatch(id string, config models.Config, batch entities.Batch, batchPrefix []string, db *gorm.DB) error {
	return jobs.RunAfterBatch(id, config, batch, batchPrefix, db)
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

func ProcessConfigIdFromParam(key string, db *gorm.DB, c *gin.Context) (*uint, entities.Config, error) {
	configId64, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return nil, entities.Config{}, err
	}

	configId := uint(configId64)
	var config entities.Config

	err1 := db.First(&config, configId).Error
	if err1 != nil {
		return nil, entities.Config{}, err1
	}

	return &configId, config, nil
}

func UnzipBatch(batchPath string) (string, error) {
	log.Println("Unzipping batch : ", batchPath, "...")
	dst := strings.TrimSuffix(batchPath, ".zip")

	err := handlers.UnzipFile(batchPath, dst, os.ModePerm)

	return dst, err
}
