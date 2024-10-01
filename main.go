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
	cfg := &ServerCfg{}
	err := cfg.Load(*conf_file)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Print()

	r.Use(func(c *gin.Context) {
		c.Set("config", cfg)
		c.Next()
	})

	f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	multiWriter := io.MultiWriter(f, os.Stdout)
	gin.DefaultWriter = multiWriter
	r.Use(gin.Logger())
	log.SetOutput(multiWriter)

	r.PUT("/api/upload/*file_path", Upload)
	r.GET("/api/download/*file_path", Get)
	//r.DELETE("/api/delete/:file_path", Delete)

	r.Run(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
}
