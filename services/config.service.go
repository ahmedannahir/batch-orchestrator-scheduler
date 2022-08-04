package services

import (
	"encoding/json"
	"gestion-batches/models"
	"log"

	"github.com/gin-gonic/gin"
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
	configs, err1 := ParseConsecConfig(configBytes)
	for i := 1; i < len(configs); i++ {
		configs[i].Cron = "1 1 30 2 1" // temp workaround, cron for feb 30th i.e. never, for batches following scheduled batches
	}

	return configs, err1
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
