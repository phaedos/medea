package service

import (
	"context"

	"github.com/jinzhu/gorm"
)

type BeforeHandler = func(ctx context.Context, service Service) error

type AfterHandler = func(ctx context.Context, service Service) error

type Service interface {
	Execute(ctx context.Context) (interface{}, error)

	Validate() ValidateErrors
}

type BaseService struct {
	Before   []BeforeHandler
	After    []AfterHandler
	DB       *gorm.DB
	Value    map[string]interface{}
	RootPath *string
}

func (b *BaseService) CallBefore(ctx context.Context, service Service) error {
	for _, handler := range b.Before {
		if err := handler(ctx, service); err != nil {
			return err
		}
	}
	return nil
}

func (b *BaseService) CallAfter(ctx context.Context, service Service) error {
	for _, handler := range b.After {
		if err := handler(ctx, service); err != nil {
			return err
		}
	}
	return nil
}

func (b *BaseService) Execute(ctx context.Context) (interface{}, error) {
	var err error

	if err = b.CallBefore(ctx, b); err != nil {
		return false, err
	}
	return true, b.CallAfter(ctx, b)
}

func (b *BaseService) Validate() ValidateErrors {
	return nil
}
