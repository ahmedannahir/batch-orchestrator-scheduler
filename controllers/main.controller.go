package controllers

import (
	"gestion-batches/handlers"
	"gestion-batches/services"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func RunBatch(c *gin.Context) {
	config, err1 := services.GetConfig("config", c)
	handlers.HandleError(err1, http.StatusBadRequest, c)

	depPrefix, batPrefix, err3 := services.ExtractLanguagePrefixes(config)
	handlers.HandleError(err3, http.StatusBadRequest, c)

	batchPath, err2 := handlers.UploadFile("batch", "jobs/scripts/", time.Now().Format("2006-01-02_15-04-05")+"_", c)
	handlers.HandleError(err2, http.StatusInternalServerError, c)

	logFile, err4 := handlers.CreateLog(batchPath.Name())
	handlers.HandleError(err4, http.StatusInternalServerError, c)

	err5 := services.InstallDependencies(config.Dependencies, logFile, depPrefix)
	handlers.HandleError(err5, http.StatusBadRequest, c)

	services.RunBatch(batchPath.Name(), logFile, batPrefix)

	c.Writer.WriteHeader(http.StatusOK)
}

func ScheduleBatch(c *gin.Context) {
	config, err1 := services.GetConfig("config", c)
	handlers.HandleError(err1, http.StatusBadRequest, c)

	depPrefix, batPrefix, err3 := services.ExtractLanguagePrefixes(config)
	handlers.HandleError(err3, http.StatusBadRequest, c)

	batchPath, err2 := services.UploadFile("batch", "jobs/scripts/", time.Now().Format("2006-01-02_15-04-05")+"_", c)
	handlers.HandleError(err2, http.StatusInternalServerError, c)

	err5 := services.InstallDependencies(config.Dependencies, os.Stdout, depPrefix) // to-figure-out-later
	handlers.HandleError(err5, http.StatusBadRequest, c)

	err6 := services.ScheduleBatch(config, batchPath, batPrefix)
	handlers.HandleError(err6, http.StatusBadRequest, c)

	c.Writer.WriteHeader(http.StatusOK)
}
