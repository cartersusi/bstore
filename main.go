package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

const MB = 1e6
const MB2 = 2 * MB

type ServerCfg struct {
	Port        string `yaml:"port"`
	Host        string `yaml:"host"`
	BasePath    string `yaml:"base_path"`
	MaxFileSize int64  `yaml:"max_file_size"`
	LogFile     string `yaml:"log_file"`
	Compress    bool   `yaml:"compress"`
}

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

func (cfg *ServerCfg) Load(conf_file string) error {
	f, err := os.Open(conf_file)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *ServerCfg) Print() {
	cd, _ := os.Getwd()
	fmt.Printf("Port: %s\n", cfg.Port)
	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("BasePath: %s\n", filepath.Join(cd, cfg.BasePath))
	fmt.Printf("MaxFileSize: %d mb\n", cfg.MaxFileSize/MB)
	fmt.Printf("LogFile: %s\n", filepath.Join(cd, cfg.LogFile))
	fmt.Printf("Compress: %t\n", cfg.Compress)
}

func GetConfig(c *gin.Context) *ServerCfg {
	return c.MustGet("config").(*ServerCfg)
}
