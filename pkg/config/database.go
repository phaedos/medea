package config

import (
	"errors"
	"fmt"
)

type Database struct {
	Driver   string `yaml:"driver,omitempty"`
	Host     string `yaml:"host,omitempty"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	DBName   string `yaml:"dbName,omitempty"`
	Port     uint32 `yaml:"port,omitempty"`
	DBFile   string `yaml:"dbFile,omitempty"`
}

func (d Database) DSN() (string, error) {
	switch d.Driver {
	case "sqlite3":
		return fmt.Sprintf("file:%s?mode=rw&cache=shared", d.DBFile), nil
	case "mysql":
		format := "%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local"
		return fmt.Sprintf(format, d.User, d.Password, d.Host, d.Port, d.DBName), nil
	case "postgres":
		format := "host=%s port=%d user=%s dbname=%s password=%s"
		return fmt.Sprintf(format, d.Host, d.Port, d.User, d.DBName, d.Password), nil
	default:
		return "", errors.New("unsupported database driver")
	}
}
