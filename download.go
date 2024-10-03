package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/carter4299/bstore/fops"
	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Get(c *gin.Context) {
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		c.JSON(validation.HttpStatus, gin.H{"error": validation.Err.Error()})
		return
	}

	fpath := filepath.Join(validation.BasePath, validation.Fpath)
	originalPath := fpath
	zstPath := fpath + ".zst"

	if info, err := os.Stat(originalPath); err == nil {
		log.Printf("Returning file %s", originalPath)
		if !info.IsDir() {
			c.File(originalPath)
			return
		}
	}

	if info, err := os.Stat(zstPath); err == nil && !info.IsDir() {
		log.Printf("Decompressing file %s", zstPath)
		content, err := fops.Decompress(zstPath)
		if err != nil {
			log.Printf("Error decompressing file %s: %v", zstPath, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decompressing file: " + err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/octet-stream", content)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
}
