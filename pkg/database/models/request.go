package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Request struct {
	ID            uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	Protocol      string    `gorm:"type:CHAR(10) NOT NULL;column:protocol"`
	AppID         *uint64   `gorm:"type:BIGINT(20) UNSIGNED;DEFAULT:NULL;column:appId"`
	Nonce         *string   `gorm:"type:CHAR(48);DEFAULT:NULL;column:nonce"`
	Token         *string   `gorm:"type:CHAR(32);DEFAULT:NULL;column:token"`
	IP            *string   `gorm:"type:CHAR(15);column:ip;DEFAULT:NULL"`
	Method        *string   `gorm:"type:CHAR(10);column:method;DEFAULT:NULL"`
	Service       *string   `gorm:"type:VARCHAR(512);column:service;DEFAULT:NULL"`
	RequestBody   string    `gorm:"type:TEXT;column:requestBody"`
	RequestHeader string    `gorm:"type:TEXT;column:requestHeader"`
	ResponseCode  int       `gorm:"type:int;column:responseCode;DEFAULT:200"`
	ResponseBody  string    `gorm:"type:TEXT;column:responseBody"`
	CreatedAt     time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
}

func (r *Request) Save(db *gorm.DB) error {
	return db.Save(r).Error
}

func NewRequestWithProtocol(protocol string, db *gorm.DB) (*Request, error) {
	var (
		req = &Request{
			Protocol: protocol,
		}
		err error
	)
	err = db.Create(req).Error
	return req, err
}

func MustNewRequestWithProtocol(protocol string, db *gorm.DB) *Request {
	req, _ := NewRequestWithProtocol(protocol, db)
	return req
}

func MustNewHTTPRequest(ip, method, url string, db *gorm.DB) *Request {
	var (
		req = &Request{
			Protocol: "http",
			IP:       &ip,
			Method:   &method,
			Service:  &url,
		}
	)
	db.Create(req)
	return req
}

func FindRequestWithAppAndNonce(app *App, nonce string, db *gorm.DB) (*Request, error) {
	var (
		request = &Request{}
		err     error
	)
	err = db.Where("appId = ? and nonce = ?", app.ID, nonce).Find(request).Error
	return request, err
}
