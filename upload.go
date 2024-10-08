package bstore

import (
	"bytes"
	"net/http"
	"os"

	"github.com/cartersusi/bstore/fops"
	"github.com/gin-gonic/gin"
)

type UploadRespone struct {
	Url string `json:"url"`
}

func (bstore *ServerCfg) Upload(c *gin.Context) {
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		HandleError(c, NewError(validation.HttpStatus, validation.Err.Error(), nil))
		return
	}

	fpath, err := fops.MkDir(validation.Fpath, validation.BasePath, bstore.Compress)
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
		err = fops.WriteFile(file, buf.Bytes(), bstore.Encrypt)
		if err != nil {
			HandleError(c, NewError(http.StatusInternalServerError, "Error writing data", err))
			return
		}
	}

	file.Sync()
	c.JSON(http.StatusOK, UploadRespone{Url: bstore.MakeUrl(c, validation.Fpath)})
}
