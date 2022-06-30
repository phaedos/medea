package models

import "time"

type History struct {
	ID        uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	ObjectID  uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:objectId"`
	FileID    uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:fileId"`
	Path      string    `gorm:"type:tinyint;column:path"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
}

func (h *History) TableName() string {
	return "histories"
}
