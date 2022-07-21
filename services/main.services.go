package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Cron         string   `json:"cron"`
	Language     string   `json:"language"`
	Dependencies []string `json:"dependencies"`
}

func UploadFile(key string, dest string, c *gin.Context) (string, error) {
	file, header, err1 := c.Request.FormFile(key)
	if err1 != nil {
		return "", err1
	}
	filename := header.Filename
	out, err2 := os.Create(dest + filename)
	if err2 != nil {
		return "", err2
	}
	defer out.Close()
	_, err3 := io.Copy(out, file)
	if err3 != nil {
		return "", err3
	}
	return out.Name(), nil
}

func InstallDependencies(installDependencyPrefix []string, dependencies []string) error {
	for _, dependency := range dependencies {
		cmd := exec.Command(installDependencyPrefix[0], installDependencyPrefix[1], dependency)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		return err
	}
	return nil
}

func ExtractLanguagePrefixes(config Config) ([]string, []string, error) {
	installDependencyPrefix := make([]string, 2)
	runBatchPrefix := make([]string, 2)
	switch config.Language {
	case "GO":
		{
			installDependencyPrefix[0] = "go"
			installDependencyPrefix[1] = "get"
			runBatchPrefix[0] = "go"
			runBatchPrefix[1] = "run"
		}
	case "JAVASCRIPT":
		{
			installDependencyPrefix[0] = "npm"
			installDependencyPrefix[1] = "install"
			runBatchPrefix[0] = "node"
		}
	case "PYTHON":
		{
			installDependencyPrefix[0] = "pip"
			installDependencyPrefix[1] = "install"
			runBatchPrefix[0] = "python"
		}
	default:
		{
			return nil, nil, errors.New(fmt.Sprint(config.Language, "configuration is not available at the moment."))
		}
	}

	return installDependencyPrefix, runBatchPrefix, nil
}

func RunBatch(batchPrefix []string, filePath string) error {
	var cmd *exec.Cmd
	if len(batchPrefix[1]) == 0 {
		cmd = exec.Command(batchPrefix[0], filePath)
	} else {
		cmd = exec.Command(batchPrefix[0], batchPrefix[1], filePath)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func GetConfig(key string, c *gin.Context) (Config, error) {
	configBytes, err := ExtractFile(key, c)
	if err != nil {
		return Config{}, err
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

func ParseConfig(configBytes []byte) (Config, error) {
	var config Config
	err := json.Unmarshal(configBytes, &config)
	return config, err
}
