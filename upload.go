package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"errors"

	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Upload(c *gin.Context) {
	fpath := c.Param("file_path")

	fpath, err := bstore.mkdir(fpath)
	if err != nil {
		HandleError(c, NewError(http.StatusBadRequest, "Error creating directory", err))
		return
	}

	file, err := os.Create(fpath)
	if err != nil {
		HandleError(c, NewError(http.StatusInternalServerError, "Error creating file", err))
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	size, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		HandleError(c, NewError(http.StatusInternalServerError, "Error reading request body", err))
		return
	}
	if size > bstore.MaxFileSize {
		HandleError(c, NewError(http.StatusBadRequest, "File size exceeds maximum allowed size", nil))
		return
	}

	if bstore.Compress {
		err = Compress(&buf, file)
		if err != nil {
			HandleError(c, NewError(http.StatusInternalServerError, "Error writing compressed data", err))
			return
		}
	} else {
		_, err = file.Write(buf.Bytes())
		if err != nil {
			HandleError(c, NewError(http.StatusInternalServerError, "Error writing data", err))
			return
		}
	}

	file.Sync()
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func (bstore *ServerCfg) mkdir(fpath string) (string, error) {
	if fpath == "" {
		return "", errors.New("file path is required")
	}

	fpath = filepath.Join(bstore.BasePath, fpath)
	if bstore.Compress {
		fpath += ".zst"
	}
	log.Printf("Uploading file %s", fpath)

	dir := filepath.Dir(fpath)
	if dir == "" {
		return "", errors.New("error getting directory")
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", errors.New("error creating directory")
	}

	return fpath, nil
}
