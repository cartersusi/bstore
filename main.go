package main

import (
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"

	bs "github.com/cartersusi/bstore/pkg"
	"github.com/gin-gonic/gin"
)

func main() {
	init_file := flag.Bool("init", false, "Create a new configuration file")
	conf_file := flag.String("config", "conf.yml", "Configuration file")

	config_dir, err := bs.ConfDir()
	if err != nil {
		log.Fatal(err)
	}
	*conf_file = filepath.Join(config_dir, "conf.yml")

	flag.Parse()
	if *init_file {
		bs.Init()
		return
	}

	bstore := &bs.ServerCfg{}
	err = bstore.Load(*conf_file)
	if err != nil {
		log.Fatal(err)
	}
	bstore.Print()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	bstore.Cors(r)
	bstore.Middleware(r)

	log_dir, err := bs.LogDir()
	if err != nil {
		log.Fatal(err)
	}
	log_fname := filepath.Join(log_dir, bstore.LogFile)
	f, err := os.OpenFile(log_fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
	r.DELETE("/api/delete/*file_path", bstore.Delete)
	r.GET("/api/list/*file_path", bstore.List)

	r.Run(bstore.Host)
}
