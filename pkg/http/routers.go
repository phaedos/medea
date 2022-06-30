package http

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"medea/pkg/config"

	"medea/pkg/utils"

	"medea/pkg/log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var brw = buildRouteWithPrefix

func Routers() *gin.Engine {
	r := gin.New()
	if config.DefaultConfig.HTTP.AccessLogFile != "" {
		setGinLogWriter()
	}

	r.Use(gin.Recovery(), AccessLogMiddleware())

	if config.DefaultConfig.HTTP.CORSEnable {
		r.Use(cors.New(cors.Config{
			AllowAllOrigins:  config.DefaultConfig.CORSAllowAllOrigins,
			AllowOrigins:     config.DefaultConfig.CORSAllowOrigins,
			AllowMethods:     config.DefaultConfig.CORSAllowMethods,
			AllowHeaders:     config.DefaultConfig.CORSAllowHeaders,
			AllowCredentials: config.DefaultConfig.CORSAllowCredentials,
			ExposeHeaders:    config.DefaultConfig.CORSExposeHeaders,
			MaxAge:           time.Duration(config.DefaultConfig.CORSMaxAge * int64(time.Second)),
		}))
	}

	r.Use(ConfigContextMiddleware(nil), RecordRequestMiddleware())

	if config.DefaultConfig.HTTP.LimitRateByIPEnable {
		interval := time.Duration(config.DefaultConfig.HTTP.LimitRateByIPInterval * int64(time.Millisecond))
		maxNumber := config.DefaultConfig.HTTP.LimitRateByIPMaxNum
		r.Use(RateLimitByIPMiddleware(interval, int(maxNumber)))
	}

	requestWithAppGroup := r.Group("", ParseAppMiddleware(), ReplayAttackMiddleware())
	requestWithAppGroup.POST(brw("/token/create"), SignWithAppMiddleware(&tokenCreateInput{}), TokenCreateHandler)
	requestWithAppGroup.PATCH(brw("/token/update"), SignWithAppMiddleware(&tokenUpdateInput{}), TokenUpdateHandler)
	requestWithAppGroup.DELETE(brw("/token/delete"), SignWithAppMiddleware(&tokenDeleteInput{}), TokenDeleteHandler)

	requestWithTokenGroup := r.Group("", ParseTokenMiddleware(), ReplayAttackMiddleware())
	requestWithTokenGroup.POST(brw("/file/create"), SignWithTokenMiddleware(&fileCreateInput{}), FileCreateHandler)
	requestWithTokenGroup.GET(brw("/file/read"), SignWithTokenMiddleware(&fileReadInput{}), FileReadHandler)
	requestWithTokenGroup.GET(brw("/file/info"), SignWithTokenMiddleware(&fileReadInput{}), FileInfoHandler)
	requestWithTokenGroup.PATCH(brw("/file/update"), SignWithTokenMiddleware(&fileUpdateInput{}), FileUpdateHandler)
	requestWithTokenGroup.DELETE(brw("/file/delete"), SignWithTokenMiddleware(&fileDeleteInput{}), FileDeleteHandler)
	requestWithTokenGroup.GET(brw("/directory/list"), SignWithTokenMiddleware(&directoryListInput{}), DirectoryListHandler)

	return r
}

func buildRoute(prefix, route string) string {
	return strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(route, "/")
}

func buildRouteWithPrefix(route string) string {
	return buildRoute(config.DefaultConfig.HTTP.APIPrefix, route)
}

func setGinLogWriter() {
	accessLogFile := config.DefaultConfig.HTTP.AccessLogFile
	dir := filepath.Dir(accessLogFile)
	logger := log.MustNewLogger(&config.DefaultConfig.Log)
	if utils.IsFile(dir) {
		logger.Fatalf("invalid access log file path: %s", accessLogFile)
	}
	if !utils.IsDir(dir) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Fatal(err)
		}
	}

	if f, err := os.OpenFile(
		config.DefaultConfig.HTTP.AccessLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		panic(err)
	} else {
		gin.DefaultWriter = io.MultiWriter(os.Stdout, f)
	}
}
