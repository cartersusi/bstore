package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Get(c *gin.Context) {
	fpath := c.Param("file_path")
	if fpath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_path is required"})
		return
	}

	fpath = filepath.Join(bstore.BasePath, fpath)
	originalPath := fpath
	zstPath := fpath + ".zst"

	// Check for the original file first
	if info, err := os.Stat(originalPath); err == nil {
		log.Printf("Serving file %s", originalPath)
		if !info.IsDir() {
			c.File(originalPath)
			return
		}
	} else {
		log.Printf("No such file: %s", originalPath)
	}

	// Check for the zstped file
	if info, err := os.Stat(zstPath); err == nil && !info.IsDir() {
		log.Printf("Decompressing file %s", zstPath)
		content, err := Decompress(zstPath)
		if err != nil {
			log.Printf("Error decompressing file %s: %v", zstPath, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decompressing file: " + err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/octet-stream", content.Bytes())
		return
	} else {
		log.Printf("No such file: %s", zstPath)
	}

	// If we reach here, the file doesn't exist
	c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
}
