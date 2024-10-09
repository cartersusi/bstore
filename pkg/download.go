package bstore

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cartersusi/bstore/pkg/fops"
	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Get(c *gin.Context) {
	log.Println("Valid Get Request for", c.Request.URL.Path)
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		HandleError(c, NewError(validation.HttpStatus, validation.Err.Error(), nil))
		return
	}

	fpath := filepath.Join(validation.BasePath, validation.Fpath)
	log.Println("Getting file at", fpath)

	info, err := os.Stat(fpath)
	if err == nil && !info.IsDir() {
		download(c, fpath, false, bstore.Encrypt)
		return
	}

	fpath += ".zst"
	info, err = os.Stat(fpath)
	if err == nil && !info.IsDir() {
		log.Println("Getting compressed file at", fpath)
		download(c, fpath, true, bstore.Encrypt)
		return
	}

	HandleError(c, NewError(http.StatusNotFound, "File not found", nil))
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
