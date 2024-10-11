package bstore

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Delete(c *gin.Context) {
	log.Println("Valid Delete Request for", c.Request.URL.Path)
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		HandleError(c, NewError(validation.HttpStatus, validation.Err.Error(), nil))
		return
	}

	fpath := filepath.Join(validation.BasePath, validation.Fpath)
	log.Println("Deleting file at", fpath)

	info, err := os.Stat(fpath)
	if err == nil && !info.IsDir() {
		rm(c, fpath)
		return
	}

	if strings.HasSuffix(fpath, "/*") {
		rmdir(c, fpath)
		return
	}

	rm(c, fpath+".zst")
	return
}

func rm(c *gin.Context, fpath string) {
	err := os.Remove(fpath)
	if err != nil {
		HandleError(c, NewError(http.StatusInternalServerError, "Error deleting file: "+err.Error(), nil))
		return
	}

	log.Println("File deleted at", fpath)
	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
	return
}

func rmdir(c *gin.Context, fpath string) {
	del_path := strings.TrimSuffix(fpath, "*")
	info, err := os.Stat(del_path)
	if err != nil {
		HandleError(c, NewError(http.StatusNotFound, "Directory not found", err))
		return
	}

	if !info.IsDir() {
		HandleError(c, NewError(http.StatusBadRequest, "Cannot delete file with wildcard", nil))
		return
	}

	err = os.RemoveAll(del_path)
	if err != nil {
		HandleError(c, NewError(http.StatusInternalServerError, "Error deleting directory: "+err.Error(), nil))
		return
	}

	log.Println("Directory deleted at", del_path)
	c.JSON(http.StatusOK, gin.H{"message": "Directory deleted successfully"})
	return
}
