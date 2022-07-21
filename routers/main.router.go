package routers

import (
	"gestion-batches/controllers"

	"github.com/gin-gonic/gin"
)

var MainRouter *gin.Engine

func HandleRoutes() {
	MainRouter = gin.Default()

	MainRouter.POST("/run-batch", controllers.RunBatch)

	MainRouter.Run("127.0.0.1:8080")
}
