package controllers

import (
	"errors"
	"gestion-batches/services"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ScheduleBatch(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		config, err := services.GetConfig("config", c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		now := time.Now()

		batchPath, err := services.UploadFile("batch", "jobs/scripts/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		dst, err := services.UnzipBatch(batchPath)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		batch, err := services.SaveBatch(config, dst, nil, db, c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		err = services.ScheduleBatch(batch, db)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		c.JSON(http.StatusCreated, batch)
	}
}

func ConsecutiveBatches(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		configs, err := services.GetConsecConfig("config", c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		err = services.VerifyConfigsAndBatchesNumber(configs, "batches", c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		now := time.Now()

		batchPaths, err := services.UploadMultipleFiles("batches", "jobs/scripts/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		services.MatchBatchAndConfig(configs, &batchPaths)

		var batchDests []string
		for _, batchPath := range batchPaths {
			dest, err := services.UnzipBatch(batchPath)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			batchDests = append(batchDests, dest)
		}

		batches, err := services.SaveConsecBatches(configs, batchDests, db, c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		err = services.ScheduleConsecBatches(batches, db)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
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

		config, err := services.GetConfig("config", c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		now := time.Now()

		batchPath, err := services.UploadFile("batch", "jobs/scripts/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		dest, err := services.UnzipBatch(batchPath)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		batch, err := services.SaveBatch(config, dest, prevBatchId, db, c)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		err = services.RunAfterBatch(prevBatchId, config, batch, db)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		c.JSON(http.StatusCreated, batch)
	}
}

func RunBatch(db *gorm.DB) gin.HandlerFunc {
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

		err = services.RunBatchById(batch, db)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
	}
}

func DisableBatch(db *gorm.DB) gin.HandlerFunc {
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

		err = services.DisableBatch(batch, db)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		}
	}
}

func EnableBatch(db *gorm.DB) gin.HandlerFunc {
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

		err = services.EnableBatch(batch, db)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		}
	}
}
