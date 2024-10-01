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
		log.Printf("Error getting directory for file %s", fpath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting directory"})
		return
	}

	log.Printf("Creating directory %s", dir)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Printf("Error creating directory %s: %s", dir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var buf bytes.Buffer
	size, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File too large"})
		} else {
			log.Printf("Error reading request body: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading request body"})
		}
		return
	}
	if size > cfg.MaxFileSize {
		log.Printf("File size %d exceeds maximum allowed size %d", size, cfg.MaxFileSize)
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File too large"})
		return
	}
	log.Printf("Request body size: %d bytes", size)

	if cfg.Compress {
		if size >= MB2 {
			log.Printf("File size %d exceeds 2MB, concurrent compressing", size)
			err = pCompress(&buf, fpath+".gz")
			if err != nil {
				log.Printf("Error compressing file %s: %s", fpath, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			return
		} else {
			log.Printf("File size %d exceeds 1MB, compressing", size)
			err = Compress(&buf, fpath+".gz")
			if err != nil {
				log.Printf("Error compressing file %s: %s", fpath, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	} else {
		log.Printf("Copying file %s", fpath)
		err = WriteBuffer(&buf, fpath)
		if err != nil {
			log.Printf("Error copying file %s: %s", fpath, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	log.Printf("File %s uploaded successfully", fpath)
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}
