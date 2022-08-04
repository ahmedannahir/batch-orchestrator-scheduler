package routers

import (
	"gestion-batches/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var MainRouter *gin.Engine

func HandleRoutes(db *gorm.DB) {
	MainRouter = gin.Default()

	MainRouter.POST("/schedule-batch", controllers.ScheduleBatch(db))
	MainRouter.POST("/consecutive-batches", controllers.ConsecutiveBatches(db))
	MainRouter.POST("/run-after-batch/:id", controllers.RunAfterBatch(db))
	MainRouter.GET("/download/batch/:id", controllers.DownloadBatch(db))
	MainRouter.GET("/download/log/:id", controllers.DownloadLog(db))
	MainRouter.GET("/download/config/:id", controllers.DownloadConfig(db))

	MainRouter.Run("127.0.0.1:8080")
}
