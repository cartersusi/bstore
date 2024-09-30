package simple

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func server() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.PUT("/upload", Uplaod)
	r.GET("/get", Get)

	r.Run(":8080")
}

func Get(c *gin.Context) {
	fpath := c.GetHeader("X-File-Path")
	fmt.Println(fpath)

	_, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(fpath)
	return
}

func Uplaod(c *gin.Context) {
	fpath := c.GetHeader("X-File-Path")

	dir := filepath.Dir(fpath)
	os.MkdirAll(dir, os.ModePerm)

	file, _ := os.Create(fpath)
	defer file.Close()

	io.Copy(file, c.Request.Body)

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}
