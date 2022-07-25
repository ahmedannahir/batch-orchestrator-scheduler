package handlers

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadFile(key string, dest string, prefix string, c *gin.Context) (*os.File, error) {
	file, header, err1 := c.Request.FormFile(key)
	if err1 != nil {
		return nil, err1
	}

	err2 := os.MkdirAll(dest, 0777)
	if err2 != nil {
		return nil, err2
	}

	out, err3 := os.Create(dest + prefix + header.Filename)
	if err3 != nil {
		return nil, err3
	}
	defer out.Close()

	_, err4 := io.Copy(out, file)
	if err4 != nil {
		return nil, err4
	}

	return out, nil
}

func CreateLog(batchPath string) (*os.File, error) {
	batchPathSlice := strings.Split(batchPath, "/")
	batchName := batchPathSlice[len(batchPathSlice)-1]
	batchName = batchName[len("2006-01-02_15-04-05"):]
	logPath := "jobs/logs/" + time.Now().Format("2006-01-02_15-04-05") + batchName + ".log"

	err := os.MkdirAll("jobs/logs/", 0777)
	if err != nil {
		return nil, err
	}

	return os.Create(logPath)
}
