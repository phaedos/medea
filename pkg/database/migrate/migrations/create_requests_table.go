package migrations

import (
	"medea/pkg/database/migrate"

	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateRequestsTable{})
}

type CreateRequestsTable struct{}

func (c *CreateRequestsTable) Name() string {
	return "create_requests_table"
}

func (c *CreateRequestsTable) Up(db *gorm.DB) error {
	return db.Exec(`
	CREATE TABLE IF NOT EXISTS requests (
		  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		  protocol CHAR(10) NOT NULL,
		  appId BIGINT(20) NULL DEFAULT NULL,
		  token CHAR(32) NULL DEFAULT NULL,
		  ip CHAR(15) NULL DEFAULT NULL,
		  method CHAR(10) NULL DEFAULT NULL,
		  service VARCHAR(512) NULL DEFAULT NULL,
		  requestBody text NULL,
		  responseCode int NULL DEFAULT 200,
		  responseBody text NULL,
		  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
		  PRIMARY KEY (id)
      )ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci`).Error
}

func (c *CreateRequestsTable) Down(db *gorm.DB) error {
	return db.DropTableIfExists("requests").Error
}
