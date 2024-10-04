package bstore

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Delete(c *gin.Context) {
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		HandleError(c, NewError(validation.HttpStatus, validation.Err.Error(), nil))
		return
	}

	fpath := filepath.Join(validation.BasePath, validation.Fpath)

	isCompressed := true
	info, err := os.Stat(fpath)
	if err == nil {
		if !info.IsDir() {
			isCompressed = false
		} else if strings.HasSuffix(fpath, "/*") {
			status, message := rmdir(fpath)
			c.JSON(status, gin.H{"message": message})
			return
		} else {
			HandleError(c, NewError(http.StatusBadRequest, "Cannot delete directory", nil))
			return
		}
	}

	if isCompressed {
		fpath = fpath + ".zst"
	}

	err = os.Remove(fpath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
	return
}

func rmdir(fpath string) (int, string) {
	del_path := strings.TrimSuffix(fpath, "*")
	err := os.RemoveAll(del_path)
	if err != nil {
		return http.StatusInternalServerError, "Error deleting directory: " + err.Error()
	}
	return http.StatusOK, "Directory deleted successfully"
}
