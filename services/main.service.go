package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"gestion-batches/entities"
	"gestion-batches/handlers"
	"gestion-batches/jobs"
	"gestion-batches/models"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetConfig(key string, c *gin.Context) (models.Config, error) {
	log.Println("Reading config file...")
	configBytes, err := ExtractFile(key, c)
	if err != nil {
		return models.Config{}, err
	}

	log.Println("Parsing config file...")
	return ParseConfig(configBytes)
}

func GetConsecConfig(key string, c *gin.Context) ([]models.Config, error) {
	log.Println("Reading config file...")
	configBytes, err := ExtractFile(key, c)
	if err != nil {
		return nil, err
	}

	log.Println("Parsing config file...")
	return ParseConsecConfig(configBytes)
}

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

func UploadFile(key string, dest string, prefix string, c *gin.Context) (string, error) {
	log.Println("Uploading file : " + key + "...")
	filePath, err := handlers.UploadFile(key, dest, prefix, c)
	return filePath.Name(), err
}

func UploadFileByFileHeader(fileHeader *multipart.FileHeader, dest string, prefix string, c *gin.Context) (string, error) {
	log.Println("Uploading file : " + fileHeader.Filename + "...")
	filePath, err := handlers.UploadFileByFileHeader(fileHeader, dest, prefix, c)
	return filePath.Name(), err
}

func UploadMultipleFiles(key string, dest string, prefix string, c *gin.Context) ([]string, error) {
	form, _ := c.MultipartForm()
	batches := form.File[key]
	var batchPaths []string

	for _, batch := range batches {
		batchPath, err := UploadFileByFileHeader(batch, dest, prefix, c)
		if err != nil {
			return nil, err
		}
		batchPaths = append(batchPaths, batchPath)
	}

	return batchPaths, nil
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

func ParseConfig(configBytes []byte) (models.Config, error) {
	var config models.Config
	err := json.Unmarshal(configBytes, &config)
	return config, err
}

func ParseConsecConfig(configBytes []byte) ([]models.Config, error) {
	var config []models.Config
	err := json.Unmarshal(configBytes, &config)
	return config, err
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

func ConsecBatches(configs []models.Config, batchCmds [][]string, batches []entities.Batch, db *gorm.DB) error {
	log.Println("Scheduling the consecutive batches...")
	err := jobs.ConsecBatches(configs, batchCmds, batches, db)
	return err
}

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

func SaveBatch(configPath string, batchPath string, db *gorm.DB, c *gin.Context) (entities.Batch, entities.Config, error) {
	configName := c.PostForm("configName")
	batchName := c.PostForm("batchName")
	batchDesc := c.PostForm("batchDesc")

	config := entities.Config{
		Name: configName,
		Url:  configPath,
	}
	tx := db.Begin()

	result1 := tx.Create(&config)
	if result1.Error != nil {
		tx.Rollback()
		return entities.Batch{}, entities.Config{}, result1.Error
	}

	batch := entities.Batch{
		Name:        batchName,
		Description: batchDesc,
		Url:         batchPath,
		ConfigID:    config.ID,
	}
	result2 := tx.Create(&batch)
	if result2.Error != nil {
		tx.Rollback()
		return entities.Batch{}, entities.Config{}, result2.Error
	}

	tx.Commit()
	return batch, config, nil
}

func SaveMultipleBatches(configPath string, batchesPaths []string, db *gorm.DB, c *gin.Context) ([]entities.Batch, entities.Config, error) {
	var batches []entities.Batch

	configName := c.PostForm("configName")
	batchName := c.PostForm("batchName")
	batchDesc := c.PostForm("batchDesc")

	config := entities.Config{
		Name: configName,
		Url:  configPath,
	}

	tx := db.Begin()

	result1 := tx.Create(&config)
	if result1.Error != nil {
		tx.Rollback()
		return nil, entities.Config{}, result1.Error
	}

	for _, batchPath := range batchesPaths {
		batch := entities.Batch{
			Name:        batchName,
			Description: batchDesc,
			Url:         batchPath,
			ConfigID:    config.ID,
		}
		batches = append(batches, batch)
	}
	result2 := tx.Create(&batches)
	if result2.Error != nil {
		tx.Rollback()
		return nil, entities.Config{}, result2.Error
	}

	tx.Commit()

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