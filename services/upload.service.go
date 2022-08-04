package services

import (
	"gestion-batches/handlers"
	"log"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

func UploadFile(key string, dest string, prefix string, c *gin.Context) (string, error) {
	log.Println("Uploading file : " + key + "...")
	filePath, err := handlers.UploadFile(key, dest, prefix, c)
	return filePath.Name(), err
}

func UploadFileByFileHeader(fileHeader *multipart.FileHeader, dest string, prefix string, c *gin.Context) (string, error) {
	log.Println("Uploading file : " + fileHeader.Filename + "...")
	filePath, err := handlers.UploadFileByFileHeader(fileHeader, dest, prefix, c)
	return filePath.Name(), err
}

func UploadMultipleFiles(key string, dest string, prefix string, c *gin.Context) ([]string, error) {
	form, _ := c.MultipartForm()
	batches := form.File[key]
	var batchPaths []string

	for _, batch := range batches {
		batchPath, err := UploadFileByFileHeader(batch, dest, prefix, c)
		if err != nil {
			return nil, err
		}
		batchPaths = append(batchPaths, batchPath)
	}

	return batchPaths, nil
}
