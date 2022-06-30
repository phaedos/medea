package migrations

import (
	"medea/pkg/database/migrate"

	"github.com/jinzhu/gorm"
)

func init() {
	migrate.DefaultMC.Register(&UpdateRequestsTable{})
}

type UpdateRequestsTable struct{}

func (c *UpdateRequestsTable) Name() string {
	return "update_requests_table"
}

func (c *UpdateRequestsTable) Up(db *gorm.DB) error {
	return db.Exec(`
	alter table requests 
		add column nonce char(48) default null after appId,
		add column requestHeader text after requestBody,
		add index appId_idx (appId)
	`).Error
}

func (c *UpdateRequestsTable) Down(db *gorm.DB) error {
	return db.Exec(`
	alter table requests 
		drop index appId_idx,
		drop column requestHeader, 
		drop column nonce
	`).Error
}
