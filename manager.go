package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cartersusi/bstore/pkg/bstore"
	"github.com/cartersusi/bstore/pkg/cmd"
)

var (
	version string
	commit  string
	date    string
)

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
  codec: "auto" # See support/README.md for all options
  bitrate: 1000 # {bitrate}k
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
	config_dir, err := bstore.ConfDir()
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

	enc_key, err := cmd.GetCMD("openssl", "rand", "-hex", "16")
	if err != nil {
		log.Fatal(err)
	}

	read_write_key, err := cmd.GetCMD("openssl", "rand", "-base64", "32")
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

// TODO: curl -s https://api.github.com/repos/cartersusi/bstore/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
func DidUpdate() bool {
	build_date, err := time.Parse("2006-01-02_15:04:05", date)
	if err != nil {
		return false
	}

	one_month_later := build_date.AddDate(0, 1, 0)
	if time.Now().After(one_month_later) {
		var input string

		fmt.Println("bstore is out of date. Would you like to update? (y/n)")
		fmt.Scanln(&input)

		if strings.ToLower(input) == "y" {
			Update()
			fmt.Println("bstore has been updated. Please restart the program.")
			return true
		} else {
			fmt.Println("bstore has not been updated.")
		}
	}

	return false
}

func Version() {
	fmt.Printf("bstore %s\n", version)
}
