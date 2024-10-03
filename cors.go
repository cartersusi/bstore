package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (bstore *ServerCfg) Cors(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     bstore.CORS.AllowOrigins,
		AllowMethods:     bstore.CORS.AllowMethods,
		AllowHeaders:     bstore.CORS.AllowHeaders,
		ExposeHeaders:    bstore.CORS.ExposeHeaders,
		AllowCredentials: bstore.CORS.AllowCredentials,
		MaxAge:           time.Duration(bstore.CORS.MaxAge),
	}))
}
