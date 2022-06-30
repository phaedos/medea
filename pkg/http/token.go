package http

import (
	"context"
	"medea/pkg/database/models"
	"medea/pkg/service"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type tokenCreateInput struct {
	AppUID         string     `form:"appUid" binding:"required"`
	Nonce          string     `form:"nonce" header:"X-Request-Nonce" binding:"required,min=32,max=48"`
	Sign           string     `form:"sign" binding:"required"`
	Path           *string    `form:"path,default=/" binding:"max=1000"`
	IP             *string    `form:"ip" binding:"omitempty,max=1500"`
	ExpiredAt      *time.Time `form:"expiredAt" time_format:"unix" binding:"omitempty,gt"`
	Secret         *string    `form:"secret" binding:"omitempty,min=12,max=32"`
	AvailableTimes *int       `form:"availableTimes,default=-1" binding:"omitempty,max=2147483647"`
	ReadOnly       *bool      `form:"readOnly,default=0"`
}

func TokenCreateHandler(ctx *gin.Context) {
	var (
		input            *tokenCreateInput
		db               *gorm.DB
		app              *models.App
		tokenCreateSrv   *service.TokenCreate
		readOnlyI8       int8
		tokenCreateValue interface{}
		err              error

		code     = 400
		reErrors map[string][]string
		success  bool
		data     interface{}
	)

	defer func() {
		ctx.JSON(code, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   success,
			Errors:    reErrors,
			Data:      data,
		})
	}()

	input = ctx.MustGet("inputParam").(*tokenCreateInput)
	db = ctx.MustGet("db").(*gorm.DB)
	app = ctx.MustGet("app").(*models.App)

	if input.ReadOnly != nil && *input.ReadOnly {
		readOnlyI8 = 1
	}

	tokenCreateSrv = &service.TokenCreate{
		BaseService: service.BaseService{
			DB: db,
		},
		IP:             input.IP,
		App:            app,
		Path:           *input.Path,
		Secret:         input.Secret,
		ReadOnly:       readOnlyI8,
		ExpiredAt:      input.ExpiredAt,
		AvailableTimes: *input.AvailableTimes,
	}

	if err := tokenCreateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err, "")
		return
	}

	if tokenCreateValue, err = tokenCreateSrv.Execute(context.Background()); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	data = tokenResp(tokenCreateValue.(*models.Token))
	success = true
	code = 200
}

type tokenUpdateInput struct {
	AppUID         string     `form:"appUid" binding:"required"`
	Token          string     `form:"token" binding:"required"`
	Nonce          string     `form:"nonce" header:"X-Request-Nonce" binding:"required,min=32,max=48"`
	Sign           string     `form:"sign" binding:"required"`
	Path           *string    `form:"path" binding:"omitempty,max=1000"`
	IP             *string    `form:"ip" binding:"omitempty,max=1500"`
	ExpiredAt      *time.Time `form:"expiredAt" time_format:"unix" binding:"omitempty,gt"`
	Secret         *string    `form:"secret" binding:"omitempty,min=12,max=32"`
	AvailableTimes *int       `form:"availableTimes" binding:"omitempty,max=2147483647"`
	ReadOnly       *bool      `form:"readOnly"`
}

func TokenUpdateHandler(ctx *gin.Context) {
	var (
		input            *tokenUpdateInput
		db               *gorm.DB
		tokenUpdateSrv   *service.TokenUpdate
		readOnlyI8       int8
		err              error
		tokenUpdateValue interface{}

		code     = 400
		reErrors map[string][]string
		success  bool
		data     interface{}
	)

	defer func() {
		ctx.JSON(code, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   success,
			Errors:    reErrors,
			Data:      data,
		})
	}()

	input = ctx.MustGet("inputParam").(*tokenUpdateInput)
	db = ctx.MustGet("db").(*gorm.DB)

	if input.ReadOnly != nil && *input.ReadOnly {
		readOnlyI8 = 1
	}

	tokenUpdateSrv = &service.TokenUpdate{
		BaseService: service.BaseService{
			DB: db,
		},
		Token:          input.Token,
		Secret:         input.Secret,
		Path:           input.Path,
		IP:             input.IP,
		ExpiredAt:      input.ExpiredAt,
		AvailableTimes: input.AvailableTimes,
		ReadOnly:       &readOnlyI8,
	}

	if err = tokenUpdateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err, "")
		return
	}

	if tokenUpdateValue, err = tokenUpdateSrv.Execute(context.TODO()); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	code = 200
	success = true
	data = tokenResp(tokenUpdateValue.(*models.Token))
}

type tokenDeleteInput struct {
	AppUID string `form:"appUid" binding:"required"`
	Token  string `form:"token" binding:"required"`
	Nonce  string `form:"nonce" header:"X-Request-Nonce" binding:"required,min=32,max=48"`
	Sign   string `form:"sign" binding:"required"`
}

func TokenDeleteHandler(ctx *gin.Context) {
	var (
		db    *gorm.DB
		err   error
		token *models.Token
		input *tokenDeleteInput

		code     = 400
		reErrors map[string][]string
		success  bool
		data     interface{}
	)

	defer func() {
		ctx.JSON(code, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   success,
			Errors:    reErrors,
			Data:      data,
		})
	}()

	input = ctx.MustGet("inputParam").(*tokenDeleteInput)
	db = ctx.MustGet("db").(*gorm.DB)

	if token, err = models.FindTokenByUID(input.Token, db); err != nil {
		reErrors = generateErrors(err, "token")
		return
	}
	db.Delete(token)
	db.Unscoped().First(token)
	success = true
	code = 200
	data = tokenResp(token)
}
