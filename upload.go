package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/klauspost/compress/zstd"
)

func (bstore *ServerCfg) Upload(c *gin.Context) {
	fpath := c.Param("file_path")
	if fpath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-File-Path header is required"})
		return
	}
	fpath = filepath.Join(bstore.BasePath, fpath)
	if bstore.Compress {
		fpath += ".zst"
	}
	log.Printf("Uploading file %s", fpath)

	dir := filepath.Dir(fpath)
	if dir == "" {
		HandleError(c, NewError(http.StatusInternalServerError, "Error getting directory", nil))
		return
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		HandleError(c, NewError(http.StatusInternalServerError, "Error creating directory", err))
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

	opts := []zstd.EOption{zstd.WithEncoderLevel(zstd.SpeedDefault)}

	if bstore.Compress {
		enc, err := zstd.NewWriter(file, opts...)
		if err != nil {
			HandleError(c, NewError(http.StatusInternalServerError, "Error creating zstd writer", err))
			return
		}
		defer enc.Close()

		_, err = enc.Write(buf.Bytes())
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
