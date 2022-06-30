package migrate

import (
	"fmt"

	"github.com/gookit/color"
	"github.com/jinzhu/gorm"
)

type Migrator interface {
	Name() string
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
}

type MigrationModel struct {
	ID        uint   `gorm:"primary_key;AUTO_INCREMENT"`
	Migration string `gorm:"type:varchar(255);not null;UNIQUE_INDEX"`
	Batch     uint   `gorm:"type:int unsigned;not null"`
}

func (m MigrationModel) TableName() string {
	return "migrations"
}

type MigrationCollection struct {
	migrations      map[string]Migrator
	migrationOrders []string
	connection      *gorm.DB
}

func (m *MigrationCollection) SetConnection(db *gorm.DB) {
	m.connection = db
}

func (m *MigrationCollection) Register(migrate Migrator) {
	if m.migrations == nil {
		m.migrations = make(map[string]Migrator)
	}
	m.migrations[migrate.Name()] = migrate
	m.migrationOrders = append(m.migrationOrders, migrate.Name())
}

func (m *MigrationCollection) CreateMigrateTable() {
	if !m.connection.HasTable(&MigrationModel{}) {
		m.connection.CreateTable(&MigrationModel{})
	}
}

func (m *MigrationCollection) MaxBatch() uint {
	m.CreateMigrateTable()
	var (
		batch struct {
			Batch uint
		}
		sql = fmt.Sprintf("select max(batch) as batch from %s", MigrationModel{}.TableName())
	)
	m.connection.Raw(sql).Scan(&batch)
	return batch.Batch
}

func (m *MigrationCollection) Upgrade() {
	var (
		migrations         []MigrationModel
		currentBatchNumber uint
		finishedMigrations = make(map[string]struct{}, len(m.migrations))
	)

	currentBatchNumber = m.MaxBatch() + 1

	m.connection.Find(&migrations)
	for _, migration := range migrations {
		finishedMigrations[migration.Migration] = struct{}{}
	}

	for _, name := range m.migrationOrders {
		if _, ok := finishedMigrations[name]; !ok {
			migration := m.migrations[name]
			if err := migration.Up(m.connection); err == nil {
				m.connection.Create(&MigrationModel{
					Migration: name,
					Batch:     currentBatchNumber,
				})
				color.Green.Printf("Migrate: %s\n", name)
			} else {
				color.Red.Printf("Migrate: %s, %s\n", name, err.Error())
				return
			}
		}
	}
}

func (m *MigrationCollection) Rollback(step uint) {
	fallbackTo := m.MaxBatch() - step + 1
	var migrations []MigrationModel
	m.connection.Where("batch >= ?", fallbackTo).Order("id desc").Find(&migrations)
	for _, migration := range migrations {
		if err := m.migrations[migration.Migration].Down(m.connection); err == nil {
			m.connection.Delete(&migration)
			color.Red.Printf("Rollback: %s\n", migration.Migration)
		} else {
			color.Red.Printf("Rollback: %s, %s\n", migration.Migration, err.Error())
			return
		}
	}
}

func (m *MigrationCollection) Refresh() {
	maxBatch := m.MaxBatch()
	if maxBatch > 0 {
		m.Rollback(maxBatch)
	}
	m.Upgrade()
}

var DefaultMC = &MigrationCollection{}
