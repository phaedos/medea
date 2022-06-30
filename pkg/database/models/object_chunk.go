package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type ObjectChunk struct {
	ID        uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	ObjectID  uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:objectId"`
	ChunkID   uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:chunkId"`
	Number    int       `gorm:"type:int;column:number"`
	HashState *string   `gorm:"type:CHAR(64) NOT NULL;UNIQUE;column:hashState"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`

	Object Object `gorm:"foreignkey:objectId;association_autoupdate:false;association_autocreate:false"`
	Chunk  Chunk  `gorm:"foreignkey:chunkId;association_autoupdate:false;association_autocreate:false"`
}

func (oc ObjectChunk) TableName() string {
	return "object_chunk"
}

func CountObjectChunkByChunkID(chunkID uint64, db *gorm.DB) (int, error) {
	var count int
	err := db.Model(&ObjectChunk{}).Where("chunkId = ?", chunkID).Count(&count).Error
	return count, err
}
