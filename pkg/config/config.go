package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Configurator struct {
	Database `yaml:"database,omitempty"`
	Log      `yaml:"log,omitempty"`
	HTTP     `yaml:"http,omitempty"`
	Chunk    `yaml:"chunk,omitempty"`
}

func ParseConfigFile(file string, config *Configurator) error {
	var (
		content []byte
		err     error
	)
	if content, err = ioutil.ReadFile(file); err != nil {
		return err
	}
	return ParseConfig(content, config)
}

func ParseConfig(configText []byte, config *Configurator) error {
	if config == nil {
		config = DefaultConfig
	}
	return yaml.Unmarshal(configText, config)
}
