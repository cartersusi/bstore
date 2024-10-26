package server

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	bs "github.com/cartersusi/bstore/pkg/bstore"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

func Start(conf_file string) {
	bstore := &bs.ServerCfg{}
	err := bstore.Load(conf_file)
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

	var cache *expirable.LRU[string, []byte]
	if bstore.Cache.Enabled {
		cache = expirable.NewLRU[string, []byte](bstore.Cache.N, nil, time.Second*time.Duration(bstore.Cache.TTL))
	} else {
		cache = nil
	}

	r.Use(bs.CacheMiddleware(cache))

	r.Use(bstore.Serve())
	r.PUT("/api/upload/*file_path", bstore.Upload)
	r.GET("/api/download/*file_path", bstore.Get)
	r.DELETE("/api/delete/*file_path", bstore.Delete)
	r.GET("/api/list/*file_path", bstore.List)

	r.Run(bstore.Host)
}
