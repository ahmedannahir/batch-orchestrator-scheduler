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

	batchPath, err2 := services.UploadFile("batch", "jobs/scripts/"+time.Now().Format("2006-01-02_15-04-05")+"_", c)
	if err2 != nil {
		log.Println(err2)
		c.JSON(http.StatusBadRequest, err2)
	}

	depPrefix, batPrefix, err3 := services.ExtractLanguagePrefixes(config)
	if err3 != nil {
		log.Println(err3)
		c.JSON(http.StatusBadRequest, err3)
	}

	err4 := services.InstallDependencies(depPrefix, config.Dependencies)
	if err4 != nil {
		log.Println(err4)
		c.JSON(http.StatusBadRequest, err4)
	}

	err5 := services.RunBatch(batPrefix, batchPath)
	if err5 != nil {
		log.Println(err5)
		c.JSON(http.StatusBadRequest, err5)
	}

	c.Writer.WriteHeader(http.StatusOK)
}
