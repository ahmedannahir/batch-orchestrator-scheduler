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

	batchPath, err2 := services.UploadFile("batch", "jobs/scripts/", time.Now().Format("2006-01-02_15-04-05")+"_", c)
	handlers.HandleError(err2, http.StatusInternalServerError, c)

	logFile, err4 := services.CreateLog(batchPath)
	handlers.HandleError(err4, http.StatusInternalServerError, c)

	err5 := services.InstallDependencies(config.Dependencies, logFile, depPrefix)
	handlers.HandleError(err5, http.StatusBadRequest, c)

	services.RunBatch(batchPath, logFile, batPrefix)

	c.Writer.WriteHeader(http.StatusOK)
}

func ScheduleBatch(c *gin.Context) {
	config, err1 := services.GetConfig("config", c)
	handlers.HandleError(err1, http.StatusBadRequest, c)

	depPrefix, batPrefix, err2 := services.ExtractLanguagePrefixes(config)
	handlers.HandleError(err2, http.StatusBadRequest, c)

	batchPath, err3 := services.UploadFile("batch", "jobs/scripts/", time.Now().Format("2006-01-02_15-04-05")+"_", c)
	handlers.HandleError(err3, http.StatusInternalServerError, c)

	err4 := services.InstallDependencies(config.Dependencies, os.Stdout, depPrefix) // to-figure-out-later
	handlers.HandleError(err4, http.StatusBadRequest, c)

	err5 := services.ScheduleBatch(config, batchPath, batPrefix)
	handlers.HandleError(err5, http.StatusBadRequest, c)

	c.Writer.WriteHeader(http.StatusOK)
}

func ConsecutiveBatches(c *gin.Context) {
	configs, err1 := services.GetConsecConfig("config", c)
	handlers.HandleError(err1, http.StatusBadRequest, c)

	batchPaths, err2 := services.UploadMultipleFiles("batches", "jobs/scripts/", time.Now().Format("2006-01-02_15-04-05")+"_", c)
	handlers.HandleError(err2, http.StatusInternalServerError, c)

	depCmds, batCmds, err3 := services.ExtractMultiLangsPref(configs)
	handlers.HandleError(err3, http.StatusBadRequest, c)

	for i := 0; i < len(configs); i++ {
		err := services.InstallDependencies(configs[i].Dependencies, os.Stdout, depCmds[i])
		handlers.HandleError(err, http.StatusBadRequest, c)
	}

	services.MatchBatchPaths(configs, &batchPaths)

	err5 := services.ConsecBatches(configs, batCmds, batchPaths)
	handlers.HandleError(err5, http.StatusBadRequest, c)

	c.Writer.WriteHeader(http.StatusOK)
}
