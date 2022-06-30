package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

const TokenReadOnly = int8(1)

const TokenNonReadOnly = int8(0)

type Token struct {
	ID             uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	UID            string     `gorm:"type:CHAR(32) NOT NULL;UNIQUE;column:uid"`
	Secret         *string    `gorm:"type:CHAR(32)"`
	AppID          uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:appId"`
	IP             *string    `gorm:"type:VARCHAR(1500);column:ip"`
	AvailableTimes int        `gorm:"type:int(10);column:availableTimes;DEFAULT:-1"`
	ReadOnly       int8       `gorm:"type:tinyint;column:readOnly;DEFAULT:0"`
	Path           string     `gorm:"type:tinyint;column:path"`
	ExpiredAt      *time.Time `gorm:"type:TIMESTAMP;column:expiredAt"`
	CreatedAt      time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt      time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`
	DeletedAt      *time.Time `gorm:"type:TIMESTAMP(6);INDEX;column:deletedAt"`

	App App `gorm:"association_foreignkey:id;foreignkey:AppID;association_autoupdate:false;association_autocreate:false"`
}

func (t *Token) TableName() string {
	return "tokens"
}

func (t *Token) Scope() string {
	return t.Path
}

func (t *Token) PathWithScope(path string) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimSuffix(t.Path, "/"), strings.Trim(path, "/"),
	)
}

func (t *Token) BeforeSave() (err error) {
	if !strings.HasPrefix(t.Path, "/") {
		t.Path = "/" + t.Path
	}
	return nil
}

func (t *Token) AllowIPAccess(ip string) bool {
	if t.IP == nil {
		return true
	}
	return strings.Contains(*t.IP, ip)
}

func (t *Token) UpdateAvailableTimes(inc int, db *gorm.DB) error {
	if t.AvailableTimes == -1 {
		return nil
	}
	t.AvailableTimes--
	return db.Model(t).Update("availableTimes", t.AvailableTimes).Error
}

func NewToken(
	app *App, path string, expiredAt *time.Time, ip, secret *string, availableTimes int, readOnly int8, db *gorm.DB,
) (*Token, error) {
	var (
		token = &Token{
			UID:            UID(),
			Secret:         secret,
			AppID:          app.ID,
			IP:             ip,
			AvailableTimes: availableTimes,
			ReadOnly:       readOnly,
			Path:           path,
			ExpiredAt:      expiredAt,
			App:            *app,
		}
		err error
	)
	err = db.Create(token).Error
	return token, err
}

func findTokenByUID(uid string, trashed bool, db *gorm.DB) (*Token, error) {
	var (
		token = &Token{}
		err   error
	)
	if trashed {
		db = db.Unscoped()
	}
	if err = db.Preload("App").Where("uid = ?", uid).Find(token).Error; err != nil {
		return token, err
	}
	return token, nil
}

func FindTokenByUID(uid string, db *gorm.DB) (*Token, error) {
	return findTokenByUID(uid, false, db)
}

func FindTokenByUIDWithTrashed(uid string, db *gorm.DB) (*Token, error) {
	return findTokenByUID(uid, true, db)
}
