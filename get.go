package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func Get(c *gin.Context) {
	fpath := c.Param("file_path")
	if fpath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-File-Path header is required"})
		return
	}
	cfg := GetConfig(c)

	fpath = filepath.Join(cfg.BasePath, fpath)

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
			HandleError(c, NewError(http.StatusInternalServerError, "Error decompressing file", err))
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
			HandleError(c, NewError(http.StatusNotFound, "File not found", nil))
			return
		}
	}

	if !file_exist && !contains_gzip {
		fpath += ".gz"
		_, err = os.Stat(fpath)
		if !os.IsNotExist(err) {
			content, err := Decompress(fpath)
			if err != nil {
				HandleError(c, NewError(http.StatusInternalServerError, "Error decompressing file", err))
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
