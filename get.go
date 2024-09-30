package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func Get(c *gin.Context) {
	fpath := c.GetHeader("X-File-Path")
	if fpath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-File-Path header is required"})
		return
	}
	fpath = filepath.Join(BASEPATH, fpath)

	_, err := os.Stat(fpath)
	file_exist := !os.IsNotExist(err)
	contains_gzip := strings.Contains(fpath, ".gz")

	if file_exist && !contains_gzip {
		c.File(fpath)
		return
	}

	if file_exist && contains_gzip {
		content, err := Decompress(fpath)
		if err != nil {
			log.Printf("Error decompressing file %s: %s", fpath, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/octet-stream", content.Bytes())
		return
	}

	if !file_exist && contains_gzip {
		fpath = strings.TrimSuffix(fpath, ".gz")
		_, err = os.Stat(fpath)
		if !os.IsNotExist(err) {
			c.File(fpath)
			return
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
	}

	if !file_exist && !contains_gzip {
		fpath += ".gz"
		_, err = os.Stat(fpath)
		if !os.IsNotExist(err) {
			content, err := Decompress(fpath)
			if err != nil {
				log.Printf("Error decompressing file %s: %s", fpath, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Data(http.StatusOK, "application/octet-stream", content.Bytes())
			return
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown error"})
	return
}
