package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type App struct {
	ID        uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	UID       string     `gorm:"type:CHAR(32) NOT NULL;UNIQUE;column:uid"`
	Secret    string     `gorm:"type:CHAR(32) NOT NULL"`
	Name      string     `gorm:"type:VARCHAR(100) NOT NULL"`
	Note      *string    `gorm:"type:VARCHAR(500) NULL"`
	CreatedAt time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`
	DeletedAt *time.Time `gorm:"type:TIMESTAMP(6);INDEX;column:deletedAt"`
}

func (app *App) TableName() string {
	return "apps"
}

func (app *App) AfterCreate(tx *gorm.DB) error {
	var file = &File{
		UID:   UID(),
		PID:   0,
		AppID: app.ID,
		Name:  "",
		IsDir: 1,
	}
	return tx.Save(file).Error
}

func NewApp(name string, note *string, db *gorm.DB) (*App, error) {
	var (
		app = &App{
			Name:   name,
			Note:   note,
			UID:    UID(),
			Secret: NewSecret(),
		}
		err error
	)
	err = db.Create(app).Error
	return app, err
}

func deleteApp(app *App, soft bool, db *gorm.DB) error {
	if !soft {
		db = db.Unscoped()
	}
	return db.Delete(app).Error
}

func DeleteAppSoft(app *App, db *gorm.DB) error {
	return deleteApp(app, true, db)
}

func DeleteAppPermanently(app *App, db *gorm.DB) error {
	return deleteApp(app, false, db)
}

func findAppByUID(uid string, trashed bool, db *gorm.DB) (*App, error) {
	var (
		app = &App{}
		err error
	)
	if trashed {
		db = db.Unscoped()
	}
	err = db.Where("uid = ?", uid).Find(app).Error
	if err != nil {
		return app, err
	}
	return app, nil
}

func FindAppByUID(uid string, db *gorm.DB) (*App, error) {
	return findAppByUID(uid, false, db)
}

func FindAppByUIDWithTrashed(uid string, db *gorm.DB) (*App, error) {
	return findAppByUID(uid, true, db)
}

func DeleteAppByUIDSoft(uid string, db *gorm.DB) error {
	app, err := FindAppByUID(uid, db)
	if err != nil {
		return err
	}
	return deleteApp(app, true, db)
}

func DeleteAppByUIDPermanently(uid string, db *gorm.DB) error {
	app, err := FindAppByUID(uid, db)
	if err != nil {
		return err
	}
	return deleteApp(app, false, db)
}
