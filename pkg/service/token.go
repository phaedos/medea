package service

import (
	"context"
	"time"

	"medea/pkg/database/models"

	"github.com/go-playground/validator"
)

type TokenCreate struct {
	BaseService

	IP             *string     `validate:"omitempty,max=1500"`
	App            *models.App `validate:"required"`
	Path           string      `validate:"required,max=1000"`
	Secret         *string     `validate:"omitempty,min=12,max=32"`
	ReadOnly       int8        `validate:"oneof=0 1"`
	ExpiredAt      *time.Time  `validate:"omitempty,gt"`
	AvailableTimes int         `validate:"omitempty,gte=-1,max=2147483647"`

	token *models.Token
}

func (t *TokenCreate) Validate() ValidateErrors {
	var (
		validateErrors ValidateErrors
		errs           error
	)

	if errs = Validate.Struct(t); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err := ValidateApp(t.DB, t.App); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("TokenCreate.App", err))
	}

	if !ValidatePath(t.Path) {
		validateErrors = append(validateErrors, generateErrorByField("TokenCreate.Path", ErrInvalidPath))
	}

	return validateErrors
}

func (t *TokenCreate) Execute(ctx context.Context) (interface{}, error) {
	var err error
	t.token, err = models.NewToken(t.App, t.Path, t.ExpiredAt, t.IP, t.Secret, t.AvailableTimes, t.ReadOnly, t.DB)
	return t.token, err
}

type TokenUpdate struct {
	BaseService

	Token          string     `validate:"required"`
	IP             *string    `validate:"omitempty,max=1500"`
	Path           *string    `validate:"omitempty,max=1000"`
	Secret         *string    `validate:"omitempty,min=12,max=32"`
	ReadOnly       *int8      `validate:"omitempty,oneof=0 1"`
	ExpiredAt      *time.Time `validate:"omitempty,gt"`
	AvailableTimes *int       `validate:"omitempty,gte=-1,max=2147483647"`
}

func (t *TokenUpdate) Validate() ValidateErrors {

	var (
		validateErrors ValidateErrors
		errs           error
	)

	if errs = Validate.Struct(t); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if t.Path != nil {
		if !ValidatePath(*t.Path) {
			validateErrors = append(validateErrors, generateErrorByField("TokenCreate.Path", ErrInvalidPath))
		}
	}

	return validateErrors
}

func (t *TokenUpdate) Execute(ctx context.Context) (result interface{}, err error) {
	var token *models.Token

	if token, err = models.FindTokenByUID(t.Token, t.DB); err != nil {
		return nil, err
	}

	if t.Path != nil {
		token.Path = *t.Path
	}
	if t.IP != nil {
		token.IP = t.IP
	}
	if t.Secret != nil {
		token.Secret = t.Secret
	}
	if t.ReadOnly != nil {
		token.ReadOnly = *t.ReadOnly
	}
	if t.ExpiredAt != nil {
		token.ExpiredAt = t.ExpiredAt
	}
	if t.AvailableTimes != nil {
		token.AvailableTimes = *t.AvailableTimes
	}

	if t.DB.Save(token).Error != nil {
		return nil, err
	}

	return token, err
}
