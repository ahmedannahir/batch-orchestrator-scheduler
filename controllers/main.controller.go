package controllers

import (
	"errors"
	"gestion-batches/services"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ScheduleBatch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		config, err1 := services.GetConfig("config", c)
		if err1 != nil {
			log.Println(err1)
			c.JSON(http.StatusBadRequest, gin.H{"error": err1})
			return
		}

		depPrefix, batPrefix, err2 := services.ExtractLanguagePrefixes(config)
		if err2 != nil {
			log.Println(err2)
			c.JSON(http.StatusBadRequest, gin.H{"error": err2})
			return
		}

		now := time.Now()

		configPath, err3 := services.UploadFile("config", "jobs/configs/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err3 != nil {
			log.Println(err3)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err3})
			return
		}

		batchPath, err4 := services.UploadFile("batch", "jobs/scripts/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err4 != nil {
			log.Println(err4)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err4})
			return
		}

		err5 := services.InstallDependencies(config.Dependencies, os.Stdout, depPrefix) // stdout to-figure-out-later
		if err5 != nil {
			log.Println(err5)
			c.JSON(http.StatusBadRequest, gin.H{"error": err5})
			return
		}

		batch, _, err6 := services.SaveBatch(configPath, batchPath, nil, db, c)
		if err6 != nil {
			log.Println(err6)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err6})
			return
		}

		err7 := services.ScheduleBatch(config, batch, batPrefix, db)
		if err7 != nil {
			log.Println(err7)
			c.JSON(http.StatusBadRequest, gin.H{"error": err7})
			return
		}

		c.JSON(http.StatusCreated, batch)
	}
}

func ConsecutiveBatches(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		configs, err1 := services.GetConsecConfig("config", c)
		if err1 != nil {
			log.Println(err1)
			c.JSON(http.StatusBadRequest, gin.H{"error": err1})
			return
		}

		err2 := services.VerifyConfigsAndBatchesNumber(configs, "batches", c)
		if err2 != nil {
			log.Println(err2)
			c.JSON(http.StatusBadRequest, gin.H{"error": err2})
			return
		}

		depCmds, batCmds, err3 := services.ExtractMultiLangsPref(configs)
		if err3 != nil {
			log.Println(err3)
			c.JSON(http.StatusBadRequest, gin.H{"error": err3})
			return
		}

		now := time.Now()

		configPath, err4 := services.UploadFile("config", "jobs/configs/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err4 != nil {
			log.Println(err4)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err4})
			return
		}

		batchPaths, err5 := services.UploadMultipleFiles("batches", "jobs/scripts/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err1 != nil {
			log.Println(err5)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err5})
			return
		}

		for i := 0; i < len(configs); i++ {
			err := services.InstallDependencies(configs[i].Dependencies, os.Stdout, depCmds[i])
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
		}

		services.MatchBatchAndConfig(configs, &batchPaths)

		batches, _, err6 := services.SaveConsecBatches(configPath, batchPaths, db, c)
		if err6 != nil {
			log.Println(err6)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err6})
			return
		}

		err7 := services.ScheduleConsecBatches(configs, batCmds, batches, db)
		if err7 != nil {
			log.Println(err7)
			c.JSON(http.StatusBadRequest, gin.H{"error": err7})
			return
		}

		c.JSON(http.StatusCreated, batches)
	}
}

func RunAfterBatch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		prevBatchId, _, err := services.ProcessBatchIdFromParam("id", db, c)
		if err != nil {
			log.Println(err)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Batch not found"})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
			return
		}

		config, err1 := services.GetConfig("config", c)
		if err1 != nil {
			log.Println(err1)
			c.JSON(http.StatusBadRequest, gin.H{"error": err1})
			return
		}

		depPrefix, batPrefix, err2 := services.ExtractLanguagePrefixes(config)
		if err2 != nil {
			log.Println(err2)
			c.JSON(http.StatusBadRequest, gin.H{"error": err2})
			return
		}

		now := time.Now()

		configPath, err3 := services.UploadFile("config", "jobs/configs/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err3 != nil {
			log.Println(err3)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err3})
			return
		}

		batchPath, err4 := services.UploadFile("batch", "jobs/scripts/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err4 != nil {
			log.Println(err4)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err4})
			return
		}

		err5 := services.InstallDependencies(config.Dependencies, os.Stdout, depPrefix) // to-figure-out-later
		if err5 != nil {
			log.Println(err5)
			c.JSON(http.StatusBadRequest, gin.H{"error": err5})
			return
		}

		batch, _, err6 := services.SaveBatch(configPath, batchPath, prevBatchId, db, c)
		if err6 != nil {
			log.Println(err6)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err6})
			return
		}

		err7 := services.RunAfterBatch(c.Param("id"), config, batch, batPrefix, db)
		if err7 != nil {
			log.Println(err7)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err7})
			return
		}

		c.JSON(http.StatusCreated, batch)
	}
}
