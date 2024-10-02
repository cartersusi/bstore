package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

type ServerCfg struct {
	Port        string     `yaml:"port"`
	Host        string     `yaml:"host"`
	BasePath    string     `yaml:"base_path"`
	MaxFileSize int64      `yaml:"max_file_size"`
	LogFile     string     `yaml:"log_file"`
	Compress    bool       `yaml:"compress"`
	CORS        CORSConfig `yaml:"cors"`
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
	err = decoder.Decode(cfg)
	if err != nil {
		return err
	}

	// Set default values if not specified in the YAML
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.BasePath == "" {
		cd, _ := os.Getwd()
		cfg.BasePath = cd
	}
	if cfg.MaxFileSize == 0 {
		cfg.MaxFileSize = 1024 * 1024 * 100 // 100 MB
	}
	if cfg.LogFile == "" {
		cfg.LogFile = "bstore.log"
	}

	return nil
}

func (cfg *ServerCfg) Print() {
	cd, _ := os.Getwd()
	fmt.Printf("Port: %s\n", cfg.Port)
	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("BasePath: %s\n", filepath.Join(cd, cfg.BasePath))
	fmt.Printf("MaxFileSize: %d mb\n", cfg.MaxFileSize/1024/1024)
	fmt.Printf("LogFile: %s\n", filepath.Join(cd, cfg.LogFile))
	fmt.Printf("Compress: %t\n", cfg.Compress)
	fmt.Printf("CORS:\n")
	fmt.Printf("  Allow Origins: %v\n", cfg.CORS.AllowOrigins)
	fmt.Printf("  Allow Methods: %v\n", cfg.CORS.AllowMethods)
	fmt.Printf("  Allow Headers: %v\n", cfg.CORS.AllowHeaders)
	fmt.Printf("  Expose Headers: %v\n", cfg.CORS.ExposeHeaders)
	fmt.Printf("  Allow Credentials: %t\n", cfg.CORS.AllowCredentials)
	fmt.Printf("  Max Age: %d\n", cfg.CORS.MaxAge)
}

func GetConfig(c *gin.Context) *ServerCfg {
	return c.MustGet("config").(*ServerCfg)
}
