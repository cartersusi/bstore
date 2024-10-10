package bstore

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type ListResponse struct {
	Files   []string `json:"files"`
	Length  int      `json:"length"`
	Message string   `json:"message"`
}

func (bstore *ServerCfg) List(c *gin.Context) {
	log.Println("Valid List Request for", c.Request.URL.Path)
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		HandleError(c, NewError(validation.HttpStatus, validation.Err.Error(), nil))
		return
	}

	dirpath := filepath.Join(validation.BasePath, validation.Fpath)
	info, err := os.Stat(dirpath)
	if err != nil {
		HandleError(c, NewError(http.StatusNotFound, "Directory not found", err))
		return
	}

	if !info.IsDir() {
		HandleError(c, NewError(http.StatusBadRequest, "Path is not a directory", errors.New("Path is not a directory")))
		return
	}

	log.Println("Listing files in", dirpath)

	list_response := &ListResponse{}
	list_response.Files, err = list_files(dirpath)
	if err != nil {
		HandleError(c, NewError(http.StatusInternalServerError, "Error listing files", err))
		return
	}

	list_response.Length = len(list_response.Files)
	list_response.Message = "Files listed successfully from " + validation.Fpath
	c.JSON(http.StatusOK, list_response)
}

func list_files(dirPath string) ([]string, error) {
	var fileList []string

	err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			fileList = append(fileList, strings.TrimSuffix(relPath, ".zst"))
			//fileList = append(fileList, strings.TrimSuffix(strings.TrimPrefix(relPath, basePath), ".zst"))
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through directory: %v", err)
	}

	return fileList, nil
}
