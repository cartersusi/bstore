package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(serve())
	r.PUT("/api/upload/*file_path", uploader)
	r.GET("/api/download/*file_path", downloader)

	r.Run(":8080")
}

func serve() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		info, err := os.Stat(path)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		if info.IsDir() {
			c.Status(http.StatusForbidden)
			return
		}

		file, err := os.Open(path)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), file)
	}
}

func downloader(c *gin.Context) {
	fpath := c.Param("file_path")[1:]
	fmt.Println(fpath)

	_, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(fpath)
	return
}

func uploader(c *gin.Context) {
	fpath := c.Param("file_path")[1:]
	fmt.Println("Endpoint Hit: Upload", fpath)

	dir := filepath.Dir(fpath)
	os.MkdirAll(dir, os.ModePerm)
	fmt.Println("Directory created:", dir)

	file, _ := os.Create(fpath)
	defer file.Close()

	io.Copy(file, c.Request.Body)

	file.Sync()
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}
