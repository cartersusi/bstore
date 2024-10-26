package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	server "github.com/cartersusi/bstore/pkg"
	bs "github.com/cartersusi/bstore/pkg/bstore"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal. Cleaning up...")
		cleanup()
		os.Exit(0)
	}()

	update := flag.Bool("update", false, "Update the binary")
	version := flag.Bool("version", false, "Print version")

	init_file := flag.Bool("init", false, "Create a new configuration file")
	conf_file := flag.String("config", "conf.yml", "Configuration file")
	flag.Parse()

	config_dir, err := bs.ConfDir()
	if err != nil {
		log.Fatal(err)
	}

	*conf_file = filepath.Join(config_dir, "conf.yml")

	// TODO: Using install script for updates right now, might change this later
	if *update {
		Update()
		return
	}

	if *version {
		Version()
		return
	}

	if *init_file {
		Init()
		return
	}

	if DidUpdate() {
		return
	}

	server.Start(*conf_file)
}

func cleanup() {
	lockFile := filepath.Join(os.TempDir(), "bstore-update.lock")
	os.Remove(lockFile)
}
