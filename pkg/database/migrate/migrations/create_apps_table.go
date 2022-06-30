package migrations

import (
	"medea/pkg/database/migrate"

	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateAppsTable{})
}

type CreateAppsTable struct{}

func (c *CreateAppsTable) Name() string {
	return "create_apps_table"
}

func (c *CreateAppsTable) Up(db *gorm.DB) error {
	return db.Exec(`
	CREATE TABLE IF NOT EXISTS apps (
	  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
	  uid char(32) not null comment 'Application unique id',
	  secret char(32) NOT NULL COMMENT 'Application Secret',
	  name varchar(100) NOT NULL COMMENT 'Application Name',
	  note varchar(500) DEFAULT NULL COMMENT 'Application Note',
	  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	  updatedAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
	  deletedAt timestamp(6) NULL DEFAULT NULL,
	  PRIMARY KEY (id),
	  UNIQUE INDEX uid_uq_idx (uid),
	  KEY deleted_at_idx (deletedAt)
	) ENGINE=InnoDB DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci`).Error
}

func (c *CreateAppsTable) Down(db *gorm.DB) error {
	return db.DropTableIfExists("apps").Error
}
