package main

import (
	"fmt"
	"net/http"
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

type BstoreError struct {
	Code    int
	Message string
	Err     error
}

func (e *BstoreError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func NewError(code int, message string, err error) *BstoreError {
	return &BstoreError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func HandleError(c *gin.Context, err error) {
	if bstoreError, ok := err.(*BstoreError); ok {
		c.JSON(bstoreError.Code, gin.H{"error": bstoreError.Message})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}
	c.Abort()
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
