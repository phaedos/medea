package service

import (
	"errors"
	"regexp"
	"time"

	"medea/pkg/database/models"

	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v9"
)

var (
	Validate                        = validator.New()
	ErrInvalidApplication           = errors.New("invalid application")
	ErrInvalidToken                 = errors.New("invalid token")
	ErrTokenIP                      = errors.New("token can't be used by this ip")
	ErrTokenAvailableTimesExhausted = errors.New("the available times of token has already exhausted")
	ErrTokenReadOnly                = errors.New("this token is read only")
	ErrTokenExpired                 = errors.New("token is expired")
	ErrInvalidFile                  = errors.New("invalid file")
)

func ValidateFile(db *gorm.DB, file *models.File) error {
	if file == nil {
		return ErrInvalidFile
	}
	return db.Where("id = ?", file.ID).Find(file).Error
}

func ValidateApp(db *gorm.DB, app *models.App) error {
	if app == nil {
		return ErrInvalidApplication
	}
	if _, err := models.FindAppByUID(app.UID, db); err != nil {
		return err
	}
	return nil
}

func ValidateToken(db *gorm.DB, ip *string, canReadOnly bool, token *models.Token) error {
	var err error
	if token == nil {
		return ErrInvalidToken
	}
	if token, err = models.FindTokenByUID(token.UID, db); err != nil {
		return err
	}

	if ip != nil && !token.AllowIPAccess(*ip) {
		return ErrTokenIP
	}

	if token.AvailableTimes != -1 && token.AvailableTimes <= 0 {
		return ErrTokenAvailableTimesExhausted
	}

	if !canReadOnly && token.ReadOnly == 1 {
		return ErrTokenReadOnly
	}

	if token.ExpiredAt != nil && token.ExpiredAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	return nil
}

func ValidatePath(path string) bool {
	var (
		regexps = []*regexp.Regexp{
			regexp.MustCompile(`^(?:/[^\^!@%();,\[\]{}<>/\\|:*?"']{1,255})+$`),
			regexp.MustCompile(`^(?:/[^\^!@%();,\[\]{}<>/\\|:*?"']{1,255})+/$`),
			regexp.MustCompile(`^(?:[^\^!@%();,\[\]{}<>/\\|:*?"']{1,255}/|$)+$?`),
			regexp.MustCompile(`^[^\^!@%();,\[\]{}<>/\\|:*?"']{1,255}$`),
			regexp.MustCompile(`^/$`),
		}
	)

	for _, regex := range regexps {
		if regex.MatchString(path) {
			return true
		}
	}

	return false
}
