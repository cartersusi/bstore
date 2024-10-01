package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func Upload(c *gin.Context) {
	fpath := c.Param("file_path")
	if fpath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-File-Path header is required"})
		return
	}
	cfg := GetConfig(c)

	fpath = filepath.Join(cfg.BasePath, fpath)
	log.Printf("Uploading file %s", fpath)

	dir := filepath.Dir(fpath)
	if dir == "" {
		HandleError(c, NewError(http.StatusInternalServerError, "Error getting directory", nil))
		return
	}

	log.Printf("Creating directory %s", dir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		HandleError(c, NewError(http.StatusInternalServerError, "Error creating directory", err))
		return
	}

	var buf bytes.Buffer
	size, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		HandleError(c, NewError(http.StatusInternalServerError, "Error reading request body", err))
		return
	}
	if size > cfg.MaxFileSize {
		HandleError(c, NewError(http.StatusRequestEntityTooLarge, "File too large", nil))
		return
	}
	log.Printf("Request body size: %d bytes", size)

	if cfg.Compress {
		if err := handle_compress(&buf, fpath, size); err != nil {
			HandleError(c, err)
			return
		}
	} else {
		if err := handle_copy(&buf, fpath); err != nil {
			HandleError(c, err)
			return
		}
	}

	log.Printf("File %s uploaded successfully", fpath)
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func handle_compress(buf *bytes.Buffer, fpath string, size int64) error {
	var err error
	if size >= MB2 {
		log.Printf("File size %d exceeds 2MB, concurrent compressing", size)
		err = pCompress(buf, fpath+".gz")
	} else {
		log.Printf("File size %d exceeds 1MB, compressing", size)
		err = Compress(buf, fpath+".gz")
	}

	if err != nil {
		return NewError(http.StatusInternalServerError, "Error compressing file", err)
	}

	return nil
}

func handle_copy(buf *bytes.Buffer, fpath string) error {
	log.Printf("Copying file %s", fpath)
	if err := WriteBuffer(buf, fpath); err != nil {
		return NewError(http.StatusInternalServerError, "Error copying file", err)
	}
	return nil
}
