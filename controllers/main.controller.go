package controllers

import (
	"gestion-batches/services"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RunBatch(c *gin.Context) {
	config, err1 := services.GetConfig("config", c)
	if err1 != nil {
		log.Println(err1)
		c.JSON(http.StatusBadRequest, err1)
	}

	batchPath, err2 := services.UploadFile("batch", "jobs/scripts/", time.Now().Format("2006-01-02_15-04-05")+"_", c)
	if err2 != nil {
		log.Println(err2)
		c.JSON(http.StatusInternalServerError, err2)
	}

	depPrefix, batPrefix, err3 := services.ExtractLanguagePrefixes(config)
	if err3 != nil {
		log.Println(err3)
		c.JSON(http.StatusBadRequest, err3)
	}

	logFile, err4 := services.CreateLog(batchPath)
	if err4 != nil {
		log.Println(err4)
		c.JSON(http.StatusInternalServerError, err4)
	}

	err5 := services.InstallDependencies(config.Dependencies, logFile, depPrefix)
	if err5 != nil {
		log.Println(err5)
		c.JSON(http.StatusBadRequest, err5)
	}

	err6 := services.RunBatch(batchPath, logFile, batPrefix)
	if err6 != nil {
		log.Println(err6)
		c.JSON(http.StatusBadRequest, err6)
	}

	c.Writer.WriteHeader(http.StatusOK)
}
