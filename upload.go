package bstore

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cartersusi/bstore/fops"
	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Upload(c *gin.Context) {
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		HandleError(c, NewError(validation.HttpStatus, validation.Err.Error(), nil))
		return
	}

	fpath, err := bstore.mkdir(validation.Fpath, validation.BasePath)
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
		err = fops.Compress(&buf, file, bstore.CompressionLevel, bstore.Encrypt)
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

func (bstore *ServerCfg) mkdir(fpath, base_path string) (string, error) {
	if fpath == "" {
		return "", errors.New("file path is required")
	}

	fpath = filepath.Join(base_path, fpath)
	if bstore.Compress {
		fpath += ".zst"
	}

	dir := filepath.Dir(fpath)
	if dir == "" {
		return "", errors.New("error getting directory")
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", errors.New("error creating directory")
	}

	return fpath, nil
}
