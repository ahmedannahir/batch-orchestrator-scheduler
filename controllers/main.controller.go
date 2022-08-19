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
		config, err1 := services.GetConfig("config", c)
		if err1 != nil {
			log.Println(err1)
			c.JSON(http.StatusBadRequest, gin.H{"error": err1})
			return
		}

		now := time.Now()

		batchPath, err4 := services.UploadFile("batch", "jobs/scripts/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err4 != nil {
			log.Println(err4)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err4})
			return
		}

		dst, err := services.UnzipBatch(batchPath)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		batch, err6 := services.SaveBatch(config, dst, nil, db, c)
		if err6 != nil {
			log.Println(err6)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err6})
			return
		}

		err7 := services.ScheduleBatch(batch, db)
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

		now := time.Now()

		batchPaths, err4 := services.UploadMultipleFiles("batches", "jobs/scripts/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err4 != nil {
			log.Println(err4)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err4})
			return
		}

		services.MatchBatchAndConfig(configs, &batchPaths)

		var batchDests []string
		for _, batchPath := range batchPaths {
			dest, err5 := services.UnzipBatch(batchPath)
			if err5 != nil {
				log.Println(err5)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err5})
				return
			}
			batchDests = append(batchDests, dest)
		}

		batches, err6 := services.SaveConsecBatches(configs, batchDests, db, c)
		if err6 != nil {
			log.Println(err6)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err6})
			return
		}

		err7 := services.ScheduleConsecBatches(batches, db)
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

		now := time.Now()

		batchPath, err4 := services.UploadFile("batch", "jobs/scripts/", now.Format("2006-01-02_15-04-05")+"_", c)
		if err4 != nil {
			log.Println(err4)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err4})
			return
		}

		dest, err5 := services.UnzipBatch(batchPath)
		if err5 != nil {
			log.Println(err5)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err5})
			return
		}

		batch, err6 := services.SaveBatch(config, dest, prevBatchId, db, c)
		if err6 != nil {
			log.Println(err6)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err6})
			return
		}

		err7 := services.RunAfterBatch(prevBatchId, config, batch, db)
		if err7 != nil {
			log.Println(err7)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err7})
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

		err1 := services.RunBatchById(batch, db)
		if err1 != nil {
			log.Println(err1)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err1})
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
