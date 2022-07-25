package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
)

func HandleError(err error, httpStatus int, c *gin.Context) {
	if err != nil {
		log.Println(err)
		c.JSON(httpStatus, err.Error())
	}
}
