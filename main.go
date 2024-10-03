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
	init_file := flag.Bool("init", false, "Create a new configuration file")
	conf_file := flag.String("config", "conf.yml", "Configuration file")
	flag.Parse()
	if *init_file {
		Init()
		return
	}

	bstore := &ServerCfg{}
	err := bstore.Load(*conf_file)
	if err != nil {
		log.Fatal(err)
	}
	bstore.Print()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	bstore.Cors(r)
	bstore.Middleware(r)

	f, err := os.OpenFile(bstore.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	multiWriter := io.MultiWriter(f, os.Stdout)
	gin.DefaultWriter = multiWriter
	r.Use(gin.Logger())
	log.SetOutput(multiWriter)

	r.Use(bstore.Serve())
	r.PUT("/api/upload/*file_path", bstore.Upload)
	r.GET("/api/download/*file_path", bstore.Get)
	r.DELETE("/api/delete/:file_path", bstore.Delete)

	r.Run(fmt.Sprintf("%s:%s", bstore.Host, bstore.Port))
}
