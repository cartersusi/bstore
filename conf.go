package bstore

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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

type RateLimitConfig struct {
	Enabled     bool  `yaml:"enabled"`
	MaxRequests int64 `yaml:"max_requests"`
	Duration    int64 `yaml:"duration"`
}

type MiddlewareConfig struct {
	MaxPathLength     int             `yaml:"max_path_length"`
	OnlyBstorePaths   bool            `yaml:"only_bstore_paths"`
	RateLimitCapacity int64           `yaml:"rate_limit_capacity"`
	RateLimit         RateLimitConfig `yaml:"rate_limit"`
}

type ServerCfg struct {
	Port             string           `yaml:"port"`
	Host             string           `yaml:"host"`
	ReadWriteKey     string           `yaml:"read_write_key"`
	PublicBasePath   string           `yaml:"public_base_path"`
	PrivateBasePath  string           `yaml:"private_base_path"`
	MaxFileSize      int64            `yaml:"max_file_size"`
	LogFile          string           `yaml:"log_file"`
	Encrypt          bool             `yaml:"encrypt"`
	Compress         bool             `yaml:"compress"`
	CompressionLevel int              `yaml:"compression_lvl"`
	CORS             CORSConfig       `yaml:"cors"`
	MWare            MiddlewareConfig `yaml:"middleware"`
}

type BstoreError struct {
	Code    int
	Message string
	Err     error
}

type ReqValidation struct {
	Err        error
	HttpStatus int
	Fpath      string
	BasePath   string
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
	log.Printf("Error: %v", err)
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
		fmt.Printf("Cannot find configuration file: `%s`\n", conf_file)
		fmt.Println("\nLoad a configuration file with:\n\t$bstore -config <path>\n")
		fmt.Println("Initialize a configuration file with:\n\t$bstore -init\n")
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return err
	}

	if cfg.ReadWriteKey == "env" || cfg.ReadWriteKey == "" {
		err = check_rw_key()
		if err != nil {
			return err
		}
	}

	if cfg.Encrypt {
		if os.Getenv("BSTORE_ENC_KEY") == "" {
			return errors.New("BSTORE_ENC_KEY environment variable is not set. Tip: Use $openssl rand -hex 16")
		}

	}

	if cfg.Host == "" {
		fmt.Println("Warning: Host is not set.")
	}

	if cfg.Port == "" {
		fmt.Println("Warning: Port is not set.")
	}

	if cfg.PrivateBasePath == "" {
		fmt.Println("Warning: PrivateBasePath is not set. Files will be stored in the root directory.")
	}
	if cfg.PublicBasePath == "" {
		fmt.Println("Warning: PublicBasePath is not set. Files will be stored in the root directory.")
	}

	if !cfg.Compress || !cfg.Encrypt {
		fmt.Println("Warning: Compression and Encryption are disabled. Files will be stored as is.")
	}

	if cfg.CompressionLevel < 1 || cfg.CompressionLevel > 4 {
		fmt.Println("Warning: Compression level must be between 1 and 4. Defaulting to 2.")
		cfg.CompressionLevel = 2
	}

	if cfg.MaxFileSize < 1 {
		return errors.New("MaxFileSize must be greater than 0")
	}

	if cfg.MaxFileSize <= 100000 {
		fmt.Printf("Warning: MaxFileSize is measured in bytes. The value %d is less than 0.1mb\n", cfg.MaxFileSize)
	}

	if cfg.MWare.MaxPathLength < 1 {
		return errors.New("MaxPathLength must be greater than 0")
	}

	if cfg.MWare.RateLimit.Enabled {
		if cfg.MWare.RateLimit.MaxRequests < 1 {
			return errors.New("Rate Limit MaxRequests must be greater than 0")
		}

		if cfg.MWare.RateLimit.Duration < 1 {
			return errors.New("Rate Limit Duration must be greater than 0")
		}

		if cfg.MWare.RateLimitCapacity < 1 {
			return errors.New("Rate Limit Capacity must be greater than 0")
		}
	}

	return nil
}

func Init() {
	init_yaml := `
host: localhost
port: 8080
read_write_key: "env" # "env" or "my_read_write_key", gen key with $openssl rand -base64 32
public_base_path: pub/bstore
private_base_path: priv/bstore
max_file_size: 100000000 # bytes
max_file_name_length: 256 
log_file: bstore.log
encrypt: true
compress: true
compression_lvl: 2 # 1-4
cors:
  allow_origins: 
    - "*"
  allow_methods: 
    - "GET"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allow_headers: 
    - "Content-Type"
    - "Authorization"
    - "X-Access"
  expose_headers: 
    - "Content-Type"
    - "Authorization"
  allow_credentials: true
  max_age: 3600           #seconds
middleware:
  max_path_length: 256
  only_bstore_paths: true
  rate_limit_capacity: 100000 # Max Number of Keys(IP Addr) in Memory
  rate_limit:
    enabled: true
    max_requests: 100
    duration: 60 # seconds
`
	log.Println("Creating configuration file: conf.yml")
	f, err := os.Create("conf.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(init_yaml)
	if err != nil {
		log.Fatal(err)
	}

	enc_key, err := get_cmd("openssl", "rand", "-hex", "16")
	if err != nil {
		log.Fatal(err)
	}

	read_write_key, err := get_cmd("openssl", "rand", "-base64", "32")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("BSTORE_ENC_KEY=", enc_key)
	fmt.Println("BSTORE_READ_WRITE_KEY=", read_write_key)
	fmt.Println("Configuration file created: conf.yml")
	os.Exit(0)
}

func (cfg *ServerCfg) Print() {
	cd, _ := os.Getwd()
	fmt.Printf("Port: %s\n", cfg.Port)
	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("ReadWriteKey: %s\n", cfg.ReadWriteKey)
	fmt.Printf("PublicBasePath: %s\n", cfg.PublicBasePath)
	fmt.Printf("PrivateBasePath: %s\n", cfg.PrivateBasePath)
	fmt.Printf("MaxFileSize: %d mb\n", cfg.MaxFileSize/1024/1024)
	fmt.Printf("LogFile: %s\n", filepath.Join(cd, cfg.LogFile))
	fmt.Printf("Encrypt: %t\n", cfg.Encrypt)
	fmt.Printf("Compress: %t\n", cfg.Compress)
	fmt.Printf("CompressionLevel: %d\n", cfg.CompressionLevel)
	fmt.Printf("CORS:\n")
	fmt.Printf("  Allow Origins: %v\n", cfg.CORS.AllowOrigins)
	fmt.Printf("  Allow Methods: %v\n", cfg.CORS.AllowMethods)
	fmt.Printf("  Allow Headers: %v\n", cfg.CORS.AllowHeaders)
	fmt.Printf("  Expose Headers: %v\n", cfg.CORS.ExposeHeaders)
	fmt.Printf("  Allow Credentials: %t\n", cfg.CORS.AllowCredentials)
	fmt.Printf("  Max Age: %d\n", cfg.CORS.MaxAge)
	fmt.Printf("Middleware:\n")
	fmt.Printf("  Max Path Length: %d\n", cfg.MWare.MaxPathLength)
	fmt.Printf("  Only Bstore Paths: %t\n", cfg.MWare.OnlyBstorePaths)
	fmt.Printf("  Rate Limit Capacity: %d\n", cfg.MWare.RateLimitCapacity)
	fmt.Printf("  Rate Limit:\n")
	fmt.Printf("    Enabled: %t\n", cfg.MWare.RateLimit.Enabled)
	fmt.Printf("    Max Requests: %d\n", cfg.MWare.RateLimit.MaxRequests)
	fmt.Printf("    Duration: %ds\n", cfg.MWare.RateLimit.Duration)
}

func GetConfig(c *gin.Context) *ServerCfg {
	return c.MustGet("config").(*ServerCfg)
}

func (bstore *ServerCfg) ValidateReq(c *gin.Context) ReqValidation {
	var ret ReqValidation
	fpath := c.Param("file_path")
	if fpath == "" {
		ret.Err = errors.New("file_path is required")
		ret.HttpStatus = http.StatusBadRequest
		return ret
	}

	ret.Fpath = fpath
	ret.BasePath = bstore.get_base_path(c.Request.Header.Get("X-access"))

	return ret
}

func (bstore *ServerCfg) get_base_path(x_access string) string {
	if x_access == "public" {
		return bstore.PublicBasePath
	}
	return bstore.PrivateBasePath
}

func (bstore *ServerCfg) GetRWKey() string {
	if bstore.ReadWriteKey == "env" || bstore.ReadWriteKey == "" {
		return os.Getenv("BSTORE_READ_WRITE_KEY")
	}
	return bstore.ReadWriteKey
}

func check_rw_key() error {
	if os.Getenv("BSTORE_READ_WRITE_KEY") == "" {
		return errors.New("BSTORE_READ_WRITE_KEY environment variable is not set")
	}
	return nil
}

func get_cmd(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
