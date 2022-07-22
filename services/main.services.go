package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Cron         string   `json:"cron"`
	Language     string   `json:"language"`
	Dependencies []string `json:"dependencies"`
}

func UploadFile(key string, dest string, prefix string, c *gin.Context) (string, error) {
	file, header, err1 := c.Request.FormFile(key)
	if err1 != nil {
		return "", err1
	}

	err2 := os.MkdirAll(dest, 0777)
	if err2 != nil {
		return "", err2
	}

	out, err3 := os.Create(dest + prefix + header.Filename)
	if err3 != nil {
		return "", err3
	}
	defer out.Close()

	_, err4 := io.Copy(out, file)
	if err4 != nil {
		return "", err4
	}

	return out.Name(), nil
}

func CreateLog(batchPath string) (*os.File, error) {
	logPath := strings.ReplaceAll(batchPath, "/scripts/", "/logs/")
	logPath += ".log"
	logPathSlice := strings.Split(logPath, "/")
	logPathSlice = logPathSlice[:len(logPathSlice)-1] // remove last element i.e leave only directories

	err := os.MkdirAll(strings.Join(logPathSlice, "/"), 0777)
	if err != nil {
		return nil, err
	}

	return os.Create(logPath)
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

func ExtractLanguagePrefixes(config Config) ([]string, []string, error) {
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

func RunBatch(batchPath string, logFile *os.File, batchPrefix []string) error {
	cmdParts := append(batchPrefix, batchPath)

	var cmd *exec.Cmd
	cmd = exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
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
