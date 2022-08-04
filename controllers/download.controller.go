package controllers

import (
	"errors"
	"gestion-batches/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DownloadBatch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, batch, err := services.ProcessBatchIdFromParam("id", db, c)
		if err != nil {
			log.Println(err)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Batch not found"})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
			return
		}

		c.Header("Content-Disposition", "attachment; filename="+batch.Url)
		c.Header("Content-Disposition", "inline;filename="+batch.Url)
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Type", "application/octet-stream")
		c.File(batch.Url)
	}
}

func DownloadLog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// logId, err := services.ProcessBatchIdFromParam("id", db, c)
		_, execution, err := services.ProcessExecIdFromParam("id", db, c)
		if err != nil {
			log.Println(err)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
			return
		}

		c.Header("Content-Disposition", "attachment; filename="+execution.LogFileUrl)
		c.Header("Content-Disposition", "inline;filename="+execution.LogFileUrl)
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Type", "application/octet-stream")
		c.File(execution.LogFileUrl)
	}
}

func DownloadConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// configId, err := services.ProcessBatchIdFromParam("id", db, c)
		_, config, err := services.ProcessConfigIdFromParam("id", db, c)
		if err != nil {
			log.Println(err)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
			return
		}

		c.Header("Content-Disposition", "attachment; filename="+config.Url)
		c.Header("Content-Disposition", "inline;filename="+config.Url)
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Type", "application/octet-stream")
		c.File(config.Url)
	}
}
