package migrations

import (
	"medea/pkg/database/migrate"

	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&CreateObjectChunkTable{})
}

type CreateObjectChunkTable struct{}

func (c *CreateObjectChunkTable) Name() string {
	return "create_object_chunk_table"
}

func (c *CreateObjectChunkTable) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS object_chunk (
		  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
		  objectId BIGINT(20) NOT NULL,
		  chunkId BIGINT(20) NOT NULL,
		  hashState text NULL,
		  number BIGINT(20) NOT NULL,
		  createdAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
          updatedAt timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
		  PRIMARY KEY (id),
		  UNIQUE INDEX object_chunk_no_uq (objectId, chunkId, number),
          KEY objectId_idx (objectId),
          KEY chunkId_idx (chunkId)
		)ENGINE = InnoDB DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci
	`).Error
}

func (c *CreateObjectChunkTable) Down(db *gorm.DB) error {
	return db.DropTableIfExists("object_chunk").Error
}
