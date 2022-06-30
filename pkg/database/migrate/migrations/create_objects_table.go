package migrations

import (
	"medea/pkg/database/migrate"

	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateObjectsTable{})
}

type CreateObjectsTable struct{}

func (c *CreateObjectsTable) Name() string {
	return "create_objects_table"
}

func (c *CreateObjectsTable) Up(db *gorm.DB) error {
	return db.Exec(`
	CREATE TABLE IF NOT EXISTS objects (
	  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
	  size INT UNSIGNED NOT NULL,
	  hash CHAR(64) NOT NULL,
	  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	  updatedAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
	  PRIMARY KEY (id),
	  UNIQUE INDEX hash_UNIQUE (hash ASC))
	ENGINE = InnoDB AUTO_INCREMENT=10000 DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci
	`).Error
}

func (c *CreateObjectsTable) Down(db *gorm.DB) error {
	return db.DropTableIfExists("objects").Error
}
