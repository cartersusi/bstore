package bstore

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cartersusi/bstore/pkg/fops"
	"github.com/gin-gonic/gin"
)

func gbytes(fpath string, to_encrypt bool) ([]byte, error) {
	info, err := os.Stat(fpath)
	if err == nil && !info.IsDir() {
		return fops.ReadFile(fpath, to_encrypt)
	}

	fpath += ".zst"
	info, err = os.Stat(fpath)
	if err == nil && !info.IsDir() {
		return fops.Decompress(fpath, to_encrypt)
	}

	return nil, errors.New("File not found")
}

func (bstore *ServerCfg) Stream() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Valid Get Request for", c.Request.URL.Path)
		path := strings.Replace(c.Request.URL.Path, "stream", bstore.PublicBasePath, 1)
		if !strings.HasPrefix(path, "/"+bstore.PublicBasePath) {
			c.Next()
			return
		}

		fpath := filepath.Join(bstore.PublicBasePath, strings.TrimPrefix(path, "/"+bstore.PublicBasePath))
		file_ext := fmt.Sprintf("video/%s", strings.TrimPrefix(filepath.Ext(fpath), "."))

		data, err := gbytes(fpath, bstore.Encrypt)
		if err != nil {
			HandleError(c, NewError(http.StatusNotFound, "File not found", err))
			return
		}

		streamHandler(c, data, file_ext)
	}
}
func streamHandler(c *gin.Context, file_data []byte, file_ext string) {
	fileSize := int64(len(file_data))
	rangeHeader := c.GetHeader("Range")

	if rangeHeader != "" {
		parts := strings.Split(strings.TrimSpace(rangeHeader), "=")
		if len(parts) == 2 {
			rangeParts := strings.Split(parts[1], "-")
			start, err := strconv.ParseInt(rangeParts[0], 10, 64)
			if err != nil {
				HandleError(c, NewError(http.StatusBadRequest, "Invalid range", err))
				return
			}

			var end int64
			if len(rangeParts) == 2 {
				end, err = strconv.ParseInt(rangeParts[1], 10, 64)
				if err != nil {
					end = fileSize - 1
				}
			} else {
				end = fileSize - 1
			}

			if start >= fileSize {
				c.Writer.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
				c.Writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				return
			}

			if end >= fileSize {
				end = fileSize - 1
			}

			chunkSize := end - start + 1

			c.Writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
			c.Writer.Header().Set("Accept-Ranges", "bytes")
			c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", chunkSize))
			c.Writer.Header().Set("Content-Type", file_ext)
			c.Writer.WriteHeader(http.StatusPartialContent)

			_, err = c.Writer.Write(file_data[start : end+1])
			if err != nil {
				log.Printf("Error while writing: %v", err)
			}

			log.Printf("Served bytes %d-%d/%d", start, end, fileSize)
			return
		}
	}

	c.Writer.Header().Set("Accept-Ranges", "bytes")
	c.Writer.Header().Set("Content-Type", file_ext)
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
	c.Writer.WriteHeader(http.StatusOK)

	_, err := c.Writer.Write(file_data)
	if err != nil {
		log.Printf("Error while writing: %v", err)
	}
	log.Printf("Served full file: %d bytes", fileSize)
}
