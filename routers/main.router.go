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
	MainRouter.POST("/run-batch/:id", controllers.RunBatch(db))
	MainRouter.POST("/disable-batch/:id", controllers.DisableBatch(db))
	MainRouter.POST("/enable-batch/:id", controllers.EnableBatch(db))
	MainRouter.GET("/download/batch/:id", controllers.DownloadBatch(db))
	MainRouter.GET("/download/log/:id", controllers.DownloadLog(db))

	MainRouter.Run()
}
