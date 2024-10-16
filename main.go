package main

import (
	"flag"
	"log"
	"path/filepath"

	server "github.com/cartersusi/bstore/pkg"
	bs "github.com/cartersusi/bstore/pkg/bstore"
)

func main() {
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
