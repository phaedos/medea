package config

import (
	"os"
	"time"

	"github.com/op/go-logging"
)

var DefaultConfig *Configurator

func init() {
	var (
		dbPassword = "character"
		dbUser     = "root"
	)

	if v, ok := os.LookupEnv("MEDEA_DEFAULT_DB_PWD"); ok {
		dbPassword = v
	}

	if v, ok := os.LookupEnv("MEDEA_DEFAULT_DB_USER"); ok {
		dbUser = v
	}

	DefaultConfig = &Configurator{
		Database{
			Driver:   "mysql",
			Host:     "127.0.0.1",
			Port:     3306,
			User:     dbUser,
			Password: dbPassword,
			DBName:   "medeadb",
		},
		Log{
			Console: ConsoleLog{
				Level:  LevelToName[logging.DEBUG],
				Enable: true,
				Format: `%{color:bold}[%{time:2006/01/02 15:04:05.000}] %{pid} %{level:.5s} %{color:reset} %{message}`,
			},
			File: FileLog{
				Enable:          true,
				Level:           LevelToName[logging.WARNING],
				Format:          "[%{time:2006/01/02 15:04:05.000}] %{pid} %{longfile} %{longfunc} %{callpath} ▶ %{level:.4s} %{message}",
				Path:            "storage/logs/medea.log",
				MaxBytesPerFile: 52428800,
			},
		},
		HTTP{
			APIPrefix:             "/api/medea",
			AccessLogFile:         "storage/logs/medea.http.access.log",
			LimitRateByIPEnable:   false,
			LimitRateByIPInterval: 1000,
			LimitRateByIPMaxNum:   100,
			CORSEnable:            false,
			CORSAllowAllOrigins:   false,
			CORSAllowCredentials:  false,
			CORSAllowHeaders:      []string{"Origin"},
			CORSAllowOrigins:      []string{"*"},
			CORSAllowMethods:      []string{"PUT", "DELETE", "PATCH"},
			CORSMaxAge:            3600 * int64(time.Second),
		},
		Chunk{
			RootPath: "storage/chunks",
		},
	}
}
