package database

import (
	"time"

	"medea/pkg/config"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var connection *gorm.DB

func NewConnection(dbConfig *config.Database) (*gorm.DB, error) {
	var (
		err error
		dsn string
	)

	if dbConfig == nil {
		dbConfig = &config.DefaultConfig.Database
	}

	if connection == nil {
		if dsn, err = dbConfig.DSN(); err != nil {
			return nil, err
		}
		if connection, err = gorm.Open(dbConfig.Driver, dsn); err != nil {
			return nil, err
		}
		connection.DB().SetConnMaxLifetime(time.Hour)
		connection.DB().SetMaxOpenConns(500)
		connection.DB().SetMaxIdleConns(0)
	}

	return connection, err
}

func MustNewConnection(dbConfig *config.Database) *gorm.DB {
	var (
		conn *gorm.DB
		err  error
	)
	if conn, err = NewConnection(dbConfig); err != nil {
		panic(err)
	}
	return conn
}
