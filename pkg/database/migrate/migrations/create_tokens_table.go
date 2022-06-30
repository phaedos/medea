package migrations

import (
	"medea/pkg/database/migrate"

	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateTokensTable{})
}

type CreateTokensTable struct{}

func (c *CreateTokensTable) Name() string {
	return "create_tokens_table"
}

func (c *CreateTokensTable) Up(db *gorm.DB) error {
	return db.Exec(`
	CREATE TABLE IF NOT EXISTS tokens (
		  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		  uid CHAR(32) NOT NULL,
		  appId BIGINT(20) UNSIGNED NOT NULL,
		  ip VARCHAR(1500) NULL DEFAULT NULL,
		  availableTimes INT NOT NULL DEFAULT -1,
		  readOnly TINYINT NOT NULL DEFAULT 0,
		  secret CHAR(32) NULL DEFAULT NULL,
		  path VARCHAR(1000) NOT NULL,
		  expiredAt timestamp(6) NULL,
		  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
		  updatedAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
		  deletedAt timestamp(6) NULL DEFAULT NULL,
		  PRIMARY KEY (id),
		  UNIQUE INDEX uid_uq_index (uid ASC),
		  KEY deleted_at_idx (deletedAt)
      )ENGINE=InnoDB DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci`).Error
}

func (c *CreateTokensTable) Down(db *gorm.DB) error {
	return db.DropTableIfExists("tokens").Error
}
