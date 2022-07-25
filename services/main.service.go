package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"gestion-batches/handlers"
	"gestion-batches/jobs"
	"gestion-batches/models"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func GetConfig(key string, c *gin.Context) (models.Config, error) {
	configBytes, err := ExtractFile(key, c)
	if err != nil {
		return models.Config{}, err
	}

	return ParseConfig(configBytes)
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
	filePath, err := handlers.UploadFile(key, dest, prefix, c)
	return filePath.Name(), err
}

func ParseConfig(configBytes []byte) (models.Config, error) {
	var config models.Config
	err := json.Unmarshal(configBytes, &config)
	return config, err
}

func InstallDependencies(dependencies []string, logFile *os.File, installDependencyPrefix []string) error {
	if len(dependencies) > 0 {
		cmdParts := append(installDependencyPrefix, "dep_placeholder")

		for _, dependency := range dependencies {
			cmdParts[len(cmdParts)-1] = dependency

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
			return nil, nil, errors.New(fmt.Sprint(config.Language, "configuration is not available at the moment."))
		}
	}

	return instDepPrefix, runBatchPrefix, nil
}

func RunBatch(batchPath string, logFile *os.File, batchPrefix []string) {
	cmdParts := append(batchPrefix, batchPath)

	var cmd *exec.Cmd
	cmd = exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Run()
}

func ScheduleBatch(config models.Config, batchPath string, batchPrefix []string) error {
	err := jobs.ScheduleBatch(config, batchPath, batchPrefix)
	return err
}
