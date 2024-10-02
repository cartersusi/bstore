package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/klauspost/compress/zstd"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	conf_file := flag.String("config", "conf.yml", "Configuration file")
	bstore := &ServerCfg{}
	err := bstore.Load(*conf_file)
	if err != nil {
		log.Fatal(err)
	}
	bstore.Print()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     bstore.CORS.AllowOrigins,
		AllowMethods:     bstore.CORS.AllowMethods,
		AllowHeaders:     bstore.CORS.AllowHeaders,
		ExposeHeaders:    bstore.CORS.ExposeHeaders,
		AllowCredentials: bstore.CORS.AllowCredentials,
		MaxAge:           time.Duration(bstore.CORS.MaxAge),
	}))

	f, err := os.OpenFile(bstore.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	multiWriter := io.MultiWriter(f, os.Stdout)
	gin.DefaultWriter = multiWriter
	r.Use(gin.Logger())
	log.SetOutput(multiWriter)

	r.Use(bstore.Serve())
	r.PUT("/api/upload/*file_path", bstore.Upload)
	r.GET("/api/download/*file_path", bstore.Get)
	//r.DELETE("/api/delete/:file_path", Delete)

	r.Run(fmt.Sprintf("%s:%s", bstore.Host, bstore.Port))
}

func (bstore *ServerCfg) Serve() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/"+bstore.BasePath) {
			c.Next()
			return
		}

		filePath := filepath.Join(bstore.BasePath, strings.TrimPrefix(path, "/"+bstore.BasePath))

		info, err := os.Stat(filePath)
		if err != nil {
			zstFilePath := filePath + ".zst"
			zstInfo, zstErr := os.Stat(zstFilePath)
			if zstErr != nil {
				c.Status(http.StatusNotFound)
				return
			}
			info = zstInfo
			filePath = zstFilePath
		}

		if info.IsDir() {
			c.Status(http.StatusForbidden)
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		if strings.HasSuffix(filePath, ".zst") {
			zstdReader, err := zstd.NewReader(file)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			defer zstdReader.Close()

			var buf bytes.Buffer
			_, err = io.Copy(&buf, zstdReader)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}

			contentType := http.DetectContentType(buf.Bytes())
			c.Header("Content-Type", contentType)
			c.Data(http.StatusOK, contentType, buf.Bytes())
		} else {
			http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), file)
		}
	}
}
