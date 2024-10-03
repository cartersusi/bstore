package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/carter4299/bstore/fops"
	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Serve() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := strings.Replace(c.Request.URL.Path, "bstore", bstore.PublicBasePath, 1)

		if !strings.HasPrefix(path, "/"+bstore.PublicBasePath) {
			c.Next()
			return
		}

		fpath := filepath.Join(bstore.PublicBasePath, strings.TrimPrefix(path, "/"+bstore.PublicBasePath))
		isCompressed := true

		info, err := os.Stat(fpath)
		if err == nil {
			if !info.IsDir() {
				isCompressed = false
			}
		}

		if isCompressed {
			fpath = fpath + ".zst"
		}

		if isCompressed {
			content, err := fops.Decompress(fpath)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			contentType := http.DetectContentType(content)
			c.Header("Content-Type", contentType)
			c.Data(http.StatusOK, contentType, content)
		} else {
			file, err := os.Open(fpath)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			defer file.Close()
			http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), file)
		}
	}
}
