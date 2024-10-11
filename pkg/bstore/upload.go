package bstore

import (
	"bytes"
	"log"
	"net/http"
	"os"

	"github.com/cartersusi/bstore/pkg/fops"
	"github.com/gin-gonic/gin"
)

type UploadRespone struct {
	Url     string `json:"url"`
	Message string `json:"message"`
}

func (bstore *ServerCfg) Upload(c *gin.Context) {
	log.Println("Valid Upload Request for", c.Request.URL.Path)
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
	log.Println("Creating file at", fpath)

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

	upload_response := &UploadRespone{}
	upload_response.Url = "UNAUTHORIZED"
	if bstore.GetAccess(c) != "private" {
		upload_response.Url = bstore.MakeUrl(c, validation.Fpath)
		upload_response.Message = "Public File Uploaded Successfully"
		log.Printf("Public file (%s) uploaded successfully to: %s\n", upload_response.Url, fpath)
	} else {
		upload_response.Message = "Private File Uploaded Successfully. No URL available"
		log.Printf("Private file (UNAUTHORIZED) uploaded successfully to: %s\n", fpath)
	}

	file.Sync()
	c.JSON(http.StatusOK, upload_response)
}
