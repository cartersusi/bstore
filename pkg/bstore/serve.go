package bstore

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cartersusi/bstore/pkg/fops"
	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Serve() gin.HandlerFunc {
	return func(c *gin.Context) {
		// no rw priv needed for public files
		log.Println("Valid Serve Request for", c.Request.URL.Path)
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

		cached_content, ok := check_OR_get(c, fpath)
		if ok {
			contentType := http.DetectContentType(cached_content)
			c.Header("Content-Type", contentType)
			c.Data(http.StatusOK, contentType, cached_content)
			return
		}

		if isCompressed {
			fpath = fpath + ".zst"
		}

		if isCompressed {
			content, err := fops.Decompress(fpath, bstore.Encrypt)
			if err != nil {
				HandleError(c, NewError(http.StatusInternalServerError, "Error decompressing file", err))
				return
			}
			set_cache(c, strings.TrimSuffix(fpath, ".zst"), content)

			contentType := http.DetectContentType(content)
			c.Header("Content-Type", contentType)
			c.Data(http.StatusOK, contentType, content)
		} else {
			file, err := os.Open(fpath)
			if err != nil {
				HandleError(c, NewError(http.StatusNotFound, "File not found", err))
				return
			}
			defer file.Close()
			http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), file)
		}
	}
}

func check_OR_get(c *gin.Context, key string) ([]byte, bool) {
	cache := GetCache(c)
	if cache != nil {
		if val, ok := cache.Get(key); ok {
			log.Println("Cache hit for", key)
			return val, true
		}
	}

	return nil, false
}

func set_cache(c *gin.Context, key string, val []byte) {
	cache := GetCache(c)
	if cache != nil {
		log.Println("Caching", key)
		cache.Add(key, val)
	}
}
