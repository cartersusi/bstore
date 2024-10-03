package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"errors"

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

func (bstore *ServerCfg) Upload(c *gin.Context) {
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		c.JSON(validation.HttpStatus, gin.H{"error": validation.Err.Error()})
		return
	}

	fpath, err := bstore.mkdir(validation.Fpath, validation.BasePath)
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
		err = fops.Compress(&buf, file, bstore.CompressionLevel)
		if err != nil {
			HandleError(c, NewError(http.StatusInternalServerError, "Error writing compressed data", err))
			return
		}
	} else {
		_, err = file.Write(buf.Bytes())
		if err != nil {
			HandleError(c, NewError(http.StatusInternalServerError, "Error writing data", err))
			return
		}
	}

	file.Sync()
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func (bstore *ServerCfg) Get(c *gin.Context) {
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		c.JSON(validation.HttpStatus, gin.H{"error": validation.Err.Error()})
		return
	}

	fpath := filepath.Join(validation.BasePath, validation.Fpath)
	originalPath := fpath
	zstPath := fpath + ".zst"

	if info, err := os.Stat(originalPath); err == nil {
		log.Printf("Returning file %s", originalPath)
		if !info.IsDir() {
			c.File(originalPath)
			return
		}
	}

	if info, err := os.Stat(zstPath); err == nil && !info.IsDir() {
		log.Printf("Decompressing file %s", zstPath)
		content, err := fops.Decompress(zstPath)
		if err != nil {
			log.Printf("Error decompressing file %s: %v", zstPath, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decompressing file: " + err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/octet-stream", content)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
}

func (bstore *ServerCfg) Delete(c *gin.Context) {
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		c.JSON(validation.HttpStatus, gin.H{"error": validation.Err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete directory"})
			return
		}
	}

	if isCompressed {
		fpath = fpath + ".zst"
	}

	if info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete directory"})
		return
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

func (bstore *ServerCfg) mkdir(fpath, base_path string) (string, error) {
	if fpath == "" {
		return "", errors.New("file path is required")
	}

	fpath = filepath.Join(base_path, fpath)
	if bstore.Compress {
		fpath += ".zst"
	}
	log.Printf("Uploading file %s", fpath)

	dir := filepath.Dir(fpath)
	if dir == "" {
		return "", errors.New("error getting directory")
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", errors.New("error creating directory")
	}

	return fpath, nil
}
