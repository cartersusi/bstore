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

	info, err := os.Stat(fpath)
	if err == nil && !info.IsDir() {
		download(c, fpath, false, bstore.Encrypt)
		return
	}

	fpath += ".zst"
	info, err = os.Stat(fpath)
	if err == nil && !info.IsDir() {
		download(c, fpath, true, bstore.Encrypt)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
}

func download(c *gin.Context, fpath string, isCompressed bool, isEncrypted bool) {
	if !isEncrypted && !isCompressed {
		c.File(fpath)
		return
	}

	if isEncrypted && !isCompressed {
		content, err := fops.DecryptFile(fpath)
		if err != nil {
			HandleError(c, NewError(http.StatusInternalServerError, "Error decrypting file", err))
			return
		}
		c.Data(http.StatusOK, "application/octet-stream", content)
		return
	}

	content, err := fops.Decompress(fpath, isEncrypted)
	if err != nil {
		HandleError(c, NewError(http.StatusInternalServerError, "Error decompressing file", err))
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
	return

}
