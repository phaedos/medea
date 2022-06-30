package config

import (
	"github.com/op/go-logging"
)

type ConsoleLog struct {
	Enable bool   `yaml:"enable,omitempty"`
	Level  string `yaml:"level,omitempty"`
	Format string `yaml:"format,omitempty"`
}

type FileLog struct {
	Enable          bool   `yaml:"enable,omitempty"`
	Level           string `yaml:"level,omitempty"`
	Format          string `yaml:"format,omitempty"`
	Path            string `yaml:"path,omitempty"`
	MaxBytesPerFile uint64 `yaml:"maxBytesPerFile,omitempty"`
}

type Log struct {
	Console ConsoleLog `yaml:"console,omitempty"`
	File    FileLog    `yaml:"file,omitempty"`
}

var (
	LevelToName = map[logging.Level]string{
		logging.DEBUG:    "DEBUG",
		logging.INFO:     "INFO",
		logging.NOTICE:   "NOTICE",
		logging.WARNING:  "WARNING",
		logging.ERROR:    "ERROR",
		logging.CRITICAL: "CRITICAL",
	}
	NameToLevel = map[string]logging.Level{
		"DEBUG":    logging.DEBUG,
		"INFO":     logging.INFO,
		"NOTICE":   logging.NOTICE,
		"WARNING":  logging.WARNING,
		"ERROR":    logging.ERROR,
		"CRITICAL": logging.CRITICAL,
	}
)
