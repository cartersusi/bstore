package main

import (
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

const BASEPATH = "bstore"
const MB = 1e6
const MB2 = 2 * MB

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	f, err := os.Create("bstore.log")
	if err != nil {
		log.Fatal(err)
	}
	gin.DefaultWriter = io.NewOffsetWriter(f, 0)
	r.Use(gin.Logger())
	log.SetOutput(gin.DefaultWriter)

	r.PUT("/upload", Upload)
	r.GET("/get", Get)

	r.Run(":8080")
}
