package migrations

import (
	"medea/pkg/database/migrate"

	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateHistoriesTable{})
}

type CreateHistoriesTable struct{}

func (c *CreateHistoriesTable) Name() string {
	return "create_histories_table"
}

func (c *CreateHistoriesTable) Up(db *gorm.DB) error {
	return db.Exec(`
	CREATE TABLE IF NOT EXISTS histories (
	  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
	  fileId BIGINT(20) UNSIGNED NOT NULL,
	  objectId BIGINT(20) UNSIGNED NOT NULL,
	  path VARCHAR(1000) NOT NULL,
	  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	  PRIMARY KEY (id))
	ENGINE = InnoDB DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci`).Error
}

func (c *CreateHistoriesTable) Down(db *gorm.DB) error {
	return db.DropTableIfExists("histories").Error
}
