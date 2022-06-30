package migrations

import (
	"medea/pkg/database/migrate"

	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateChunksTable{})
}

type CreateChunksTable struct{}

func (c *CreateChunksTable) Name() string {
	return "create_chunks_table"
}

func (c *CreateChunksTable) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS chunks (
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

func (c *CreateChunksTable) Down(db *gorm.DB) error {
	return db.DropTableIfExists("chunks").Error
}
