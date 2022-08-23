package handlers

import (
	"archive/zip"
	"fmt"
	"gestion-batches/entities"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadFile(key string, dest string, prefix string, c *gin.Context) (*os.File, error) {
	file, header, err := c.Request.FormFile(key)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return nil, err
	}

	out, err := os.Create(dest + prefix + header.Filename)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func UploadFileByFileHeader(fileHeader *multipart.FileHeader, dest string, prefix string, c *gin.Context) (*os.File, error) {
	err := os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return nil, err
	}

	out, err := os.Create(dest + prefix + fileHeader.Filename)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(out, file)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func CreateLog(batch entities.Batch) (*os.File, error) {
	batchPathSlice := strings.Split(batch.Url, "/")
	batchName := batchPathSlice[len(batchPathSlice)-1]
	batchName = batchName[len("2006-01-02_15-04-05"):]
	now := time.Now()

	logPath := "jobs/logs/" + now.Format("2006-01-02_15-04-05") + "_" + strconv.FormatUint(uint64(batch.ID), 10) + batchName + ".log"

	err := os.MkdirAll("jobs/logs/", os.ModePerm)
	if err != nil {
		return nil, err
	}

	return os.Create(logPath)
}

func CreateErrLog(outLogPath string) (*os.File, error) {
	// jobs/logs/2022-08-05_15-19-01_transaction-db.py >> jobs/logs/2022-08-05_15-19-01_transaction-db_err.py
	strSlice := strings.Split(outLogPath, ".")
	strSlice = append(strSlice[:len(strSlice)-1], "_err", strSlice[len(strSlice)-1])
	errLogPath := strings.Join(strSlice, ".")

	errLogFile, err := os.Create(errLogPath)
	if err != nil {
		log.Println("Error creating Error Log File : ", err)
		return nil, err
	}

	return errLogFile, nil
}

func UnzipFile(archivePath string, dest string, perm fs.FileMode) error {
	os.MkdirAll(dest, os.ModePerm)

	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}

	for _, f := range archive.File {
		filePath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		// We add these lines in the begining of script.sh to get the current directory to where the file is
		if f.Name == "script.sh" {
			dstFile.WriteString(fmt.Sprintln("cd dirname \"${BASH_SOURCE[0]}\""))
			dstFile.WriteString(fmt.Sprintln("cd " + dest))
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		dstFile.Close()
		fileInArchive.Close()
	}

	archive.Close()
	err = os.Remove(archivePath)
	if err != nil {
		return err
	}

	return nil
}
