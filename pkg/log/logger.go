package log

import (
	"os"
	"strings"

	"medea/pkg/config"

	"github.com/op/go-logging"
)

var log *logging.Logger

func NewLogger(logConfig *config.Log) (logger *logging.Logger, err error) {
	if log != nil {
		return log, nil
	}

	if logConfig == nil {
		logConfig = &config.DefaultConfig.Log
	}

	var (
		ok                  bool
		module              = "medea"
		level               logging.Level
		backend             []logging.Backend
		consoleBackend      logging.Backend
		fileBackend         logging.Backend
		fileHandler         *AutoRotateWriter
		consoleLevelBackend logging.LeveledBackend
		fileLevelBackend    logging.LeveledBackend
		leveledBackend      logging.LeveledBackend
	)

	log = logging.MustGetLogger(module)

	if logConfig.Console.Enable {
		consoleBackend = logging.NewBackendFormatter(
			logging.NewLogBackend(os.Stdout, "", 0),
			logging.MustStringFormatter(logConfig.Console.Format),
		)
		consoleLevelBackend = logging.AddModuleLevel(consoleBackend)
		if level, ok = config.NameToLevel[strings.ToUpper(logConfig.Console.Level)]; !ok {
			level = logging.DEBUG
		}
		consoleLevelBackend.SetLevel(level, module)
		backend = append(backend, consoleLevelBackend)
	}

	if logConfig.File.Enable {
		fileHandler, err = NewAutoRotateWriter(logConfig.File.Path, logConfig.File.MaxBytesPerFile)
		if err != nil {
			return
		}
		fileBackend = logging.NewBackendFormatter(
			logging.NewLogBackend(fileHandler, "", 0),
			logging.MustStringFormatter(logConfig.File.Format),
		)
		fileLevelBackend = logging.AddModuleLevel(fileBackend)
		if level, ok = config.NameToLevel[strings.ToUpper(logConfig.File.Level)]; !ok {
			level = logging.WARNING
		}
		fileLevelBackend.SetLevel(level, module)
		backend = append(backend, fileLevelBackend)
	}

	leveledBackend = logging.MultiLogger(backend...)
	log.SetBackend(leveledBackend)

	return log, err
}

func MustNewLogger(logConfig *config.Log) (logger *logging.Logger) {
	logger, _ = NewLogger(logConfig)
	return logger
}
