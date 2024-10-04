package bstore

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/cartersusi/bstore/fops"
	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Get(c *gin.Context) {
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		HandleError(c, NewError(validation.HttpStatus, validation.Err.Error(), nil))
		return
	}

	fpath := filepath.Join(validation.BasePath, validation.Fpath)
	originalPath := fpath
	zstPath := fpath + ".zst"

	if info, err := os.Stat(originalPath); err == nil {
		if !info.IsDir() {
			c.File(originalPath)
			return
		}
	}

	if info, err := os.Stat(zstPath); err == nil && !info.IsDir() {
		content, err := fops.Decompress(zstPath, bstore.Encrypt)
		if err != nil {
			HandleError(c, NewError(http.StatusInternalServerError, "Error decompressing file", err))
			return
		}
		c.Data(http.StatusOK, "application/octet-stream", content)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
}
