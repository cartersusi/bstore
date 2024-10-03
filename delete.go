package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Delete(c *gin.Context) {
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		c.JSON(validation.HttpStatus, gin.H{"error": validation.Err.Error()})
		return
	}

	fpath := filepath.Join(validation.BasePath, validation.Fpath)
	originalPath := fpath
	zstPath := fpath + ".zst"

	if info, err := os.Stat(originalPath); err == nil {
		log.Printf("Deleting file %s", originalPath)
		if !info.IsDir() {
			err := os.Remove(originalPath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting file: " + err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
			return
		}
	}

	if info, err := os.Stat(zstPath); err == nil && !info.IsDir() {
		log.Printf("Deleting file %s", zstPath)
		err := os.Remove(zstPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting file: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
}
