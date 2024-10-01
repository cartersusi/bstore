package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	conf_file := flag.String("config", "conf.yml", "Configuration file")
	bstore := &ServerCfg{}
	err := bstore.Load(*conf_file)
	if err != nil {
		log.Fatal(err)
	}
	bstore.Print()

	f, err := os.OpenFile(bstore.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	multiWriter := io.MultiWriter(f, os.Stdout)
	gin.DefaultWriter = multiWriter
	r.Use(gin.Logger())
	log.SetOutput(multiWriter)

	r.PUT("/api/upload/*file_path", bstore.Upload)
	r.GET("/api/download/*file_path", bstore.Get)
	//r.DELETE("/api/delete/:file_path", Delete)

	r.Run(fmt.Sprintf("%s:%s", bstore.Host, bstore.Port))
}
