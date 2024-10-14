package bstore

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

type StreamingConfig struct {
	Enabled bool   `yaml:"enable"`
	Codec   string `yaml:"codec"`
}

type ServerCfg struct {
	Host             string           `yaml:"host"`
	Keys             string           `yaml:"keys"`
	PublicBasePath   string           `yaml:"public_base_path"`
	PrivateBasePath  string           `yaml:"private_base_path"`
	MaxFileSize      int64            `yaml:"max_file_size"`
	LogFile          string           `yaml:"log_file"`
	Encrypt          bool             `yaml:"encrypt"`
	Compress         bool             `yaml:"compress"`
	CompressionLevel int              `yaml:"compression_lvl"`
	Streaming        StreamingConfig  `yaml:"streaming"`
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
	log.Printf("Error: %v\n", err)
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

	err = cfg.check_keys()
	if err != nil {
		return err
	}

	err = cfg.handle_basepaths()
	if err != nil {
		return err
	}

	if cfg.Host == "" {
		fmt.Println("Warning: Host is not set.")
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
	init_yaml := `host: localhost:8080
keys: "keys.env" # "env" or "<filename>"
public_base_path: pub/bstore
private_base_path: priv/bstore
max_file_size: 100000000 # bytes
max_file_name_length: 256 
log_file: bstore.log
encrypt: true
compress: true
compression_lvl: 2 # 1-4
streaming: 
  enable: true
  codec: "libx264" # See support/README.md for all options
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
	config_dir, err := ConfDir()
	if err != nil {
		log.Fatal(err)
	}

	conf_file := filepath.Join(config_dir, "conf.yml")
	fmt.Printf("Creating configuration file: %s\n", conf_file)
	f, err := os.Create(conf_file)
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

	enc_key = strings.TrimSuffix(enc_key, "\n")
	read_write_key = strings.TrimSuffix(read_write_key, "\n")

	key_file := filepath.Join(config_dir, "keys.env")
	fmt.Printf("Creating encryption and read/write keys file: %s\n", key_file)
	f, err = os.Create(key_file)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.WriteString(fmt.Sprintf("BSTORE_ENC_KEY=\"%s\"\n", enc_key))
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.WriteString(fmt.Sprintf("BSTORE_READ_WRITE_KEY=\"%s\"", read_write_key))
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func Version() {
	info, ok := debug.ReadBuildInfo()
	if ok {
		version = info.Main.Version
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				commit = setting.Value
			case "vcs.time":
				date = setting.Value
			}
		}
	}
	fmt.Printf("bstore %s, commit %s, built at %s\n", version, commit, date)
}

func Update() {
	resp, err := http.Get("https://cartersusi.com/bstore/install")
	if err != nil {
		fmt.Printf("Error downloading script: %v\n", err)
		return
	}
	defer resp.Body.Close()

	tmpfile, err := os.CreateTemp("", "install-*.sh")
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return
	}
	defer os.Remove(tmpfile.Name())

	_, err = io.Copy(tmpfile, resp.Body)
	if err != nil {
		fmt.Printf("Error writing to temporary file: %v\n", err)
		return
	}

	if err := tmpfile.Close(); err != nil {
		fmt.Printf("Error closing temporary file: %v\n", err)
		return
	}

	cmd := exec.Command("bash", tmpfile.Name())

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running script: %v\n", err)
	}
}

func (cfg *ServerCfg) Print() {
	cd, _ := os.Getwd()
	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("Keys: %s\n", cfg.Keys)
	fmt.Printf("PublicBasePath: %s\n", cfg.PublicBasePath)
	fmt.Printf("PrivateBasePath: %s\n", cfg.PrivateBasePath)
	fmt.Printf("MaxFileSize: %d mb\n", cfg.MaxFileSize/1024/1024)
	fmt.Printf("LogFile: %s\n", filepath.Join(cd, cfg.LogFile))
	fmt.Printf("Encrypt: %t\n", cfg.Encrypt)
	fmt.Printf("Compress: %t\n", cfg.Compress)
	fmt.Printf("CompressionLevel: %d\n", cfg.CompressionLevel)
	fmt.Printf("Streaming:\n")
	fmt.Printf("  Enabled: %t\n", cfg.Streaming.Enabled)
	fmt.Printf("  Codec: %s\n", cfg.Streaming.Codec)
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

func (bstore *ServerCfg) GetAccess(c *gin.Context) string {
	return c.Request.Header.Get("X-access")
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
	ret.BasePath = bstore.get_base_path(bstore.GetAccess(c))

	return ret
}

func (bstore *ServerCfg) MakeUrl(c *gin.Context, fpath string) string {
	url := fmt.Sprintf("%s/bstore%s", c.Request.Host, fpath)
	url = strings.Replace(url, "//", "/", -1)

	var is_https = c.Request.TLS
	if is_https != nil {
		return fmt.Sprintf("https://%s", url)
	}
	return fmt.Sprintf("http://%s", url)
}

func ConfDir() (string, error) {
	home_dir := os.Getenv("HOME")
	if home_dir == "" {
		currentUser, err := user.Current()
		if err != nil {
			fmt.Println("Error:", err)
			return "", err
		}

		home_dir := currentUser.HomeDir
		if home_dir == "" {
			fmt.Println("Error: Could not determine home directory")
			return "", err
		}
	}
	config_path := filepath.Join(home_dir, ".bstore")
	// make the directory if it doesn't exist
	if _, err := os.Stat(config_path); os.IsNotExist(err) {
		err := os.Mkdir(config_path, 0755)
		if err != nil {
			return "", err
		}
	}

	return config_path, nil
}

func LogDir() (string, error) {
	config_path, err := ConfDir()
	if err != nil {
		return "", err
	}

	log_path := filepath.Join(config_path, "logs")
	if _, err := os.Stat(log_path); os.IsNotExist(err) {
		err := os.Mkdir(log_path, 0755)
		if err != nil {
			return "", err
		}
	}

	return log_path, nil
}

func (bstore *ServerCfg) get_base_path(x_access string) string {
	if x_access == "public" {
		return bstore.PublicBasePath
	}
	return bstore.PrivateBasePath
}

func (bstore *ServerCfg) keys_in_file() bool {
	if bstore.Keys != "env" {
		return true
	}
	return false
}

func (bstore *ServerCfg) GetRWKey() string {
	//if bstore.keys_in_file() {
	//	fmt.Printf("Loading keys from file: %s\n", bstore.Keys)
	//	err := godotenv.Load(bstore.Keys)
	//	if err != nil {
	//		log.Fatal("Error loading .env file")
	//	}
	//}
	return os.Getenv("BSTORE_READ_WRITE_KEY")
}

func (cfg *ServerCfg) handle_basepaths() error {
	if cfg.PrivateBasePath == "" {
		fmt.Println("Warning: PrivateBasePath is not set. Files will be stored in the root directory.")
	}
	if cfg.PrivateBasePath[0] != '/' {
		conf_dir, err := ConfDir()
		if err != nil {
			return err
		}
		cfg.PrivateBasePath = filepath.Join(conf_dir, cfg.PrivateBasePath)
	}

	if _, err := os.Stat(cfg.PrivateBasePath); os.IsNotExist(err) {
		fmt.Printf("Warning: PrivateBasePath directory `%s` does not exist. Creating it.\n", cfg.PrivateBasePath)
		err := os.MkdirAll(cfg.PrivateBasePath, 0755)
		if err != nil {
			return err
		}
	}

	if cfg.PublicBasePath == "" {
		fmt.Println("Warning: PublicBasePath is not set. Files will be stored in the root directory.")
	}

	if cfg.PublicBasePath[0] != '/' {
		conf_dir, err := ConfDir()
		if err != nil {
			return err
		}
		cfg.PublicBasePath = filepath.Join(conf_dir, cfg.PublicBasePath)
	}

	if _, err := os.Stat(cfg.PublicBasePath); os.IsNotExist(err) {
		fmt.Printf("Warning: PublicBasePath directory `%s` does not exist. Creating it.\n", cfg.PublicBasePath)
		err := os.MkdirAll(cfg.PublicBasePath, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bstore *ServerCfg) check_keys() error {
	if bstore.keys_in_file() {
		fmt.Printf("Loading keys from file: %s\n", bstore.Keys)

		conf_dir, err := ConfDir()
		if err != nil {
			return err
		}

		// pretty sure this is global, GetRWKey() is commented out until certain
		bstore.Keys = filepath.Join(conf_dir, bstore.Keys)
		err = godotenv.Load(bstore.Keys)
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	if os.Getenv("BSTORE_READ_WRITE_KEY") == "" {
		return errors.New("BSTORE_READ_WRITE_KEY environment variable is not set")
	}
	if os.Getenv("BSTORE_ENC_KEY") == "" && bstore.Encrypt {
		return errors.New("BSTORE_ENC_KEY environment variable is not set")
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
