package migrations

import (
	"medea/pkg/database/migrate"

	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateFilesTable{})
}

type CreateFilesTable struct{}

func (c *CreateFilesTable) Name() string {
	return "create_files_table"
}

func (c *CreateFilesTable) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS files (
		  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		  appId BIGINT(20) UNSIGNED NOT NULL default 0,
		  pid BIGINT(20) UNSIGNED NOT NULL DEFAULT 0,
		  uid CHAR(32) NOT NULL,
		  name VARCHAR(255) NOT NULL default '',
		  ext VARCHAR(255) NOT NULL default '',
		  objectId BIGINT(20) UNSIGNED NOT NULL default 0,
		  size BIGINT(20) UNSIGNED NOT NULL default 0,
		  isDir TINYINT UNSIGNED NOT NULL DEFAULT 0,
		  downloadCount BIGINT(20) UNSIGNED NOT NULL DEFAULT 0,
		  hidden TINYINT UNSIGNED NOT NULL DEFAULT 0,
		  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
		  updatedAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
		  deletedAt timestamp(6) NULL DEFAULT NULL,
		  PRIMARY KEY (id),
		  KEY objectId_idx (objectId),
		  KEY appId_idx (appId),
		  KEY deleted_at_idx (deletedAt),
          UNIQUE appId_pid_name_unique (appId, pid, name),
		  UNIQUE INDEX uid_UNIQUE (uid ASC))
		ENGINE = InnoDB DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci
	`).Error
}

func (c *CreateFilesTable) Down(db *gorm.DB) error {
	return db.DropTableIfExists("files").Error
}
