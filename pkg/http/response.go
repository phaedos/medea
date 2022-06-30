package http

import (
	"medea/pkg/database/models"
	"medea/pkg/service"

	"github.com/jinzhu/gorm"
)

type Response struct {
	RequestID int64               `json:"requestId"`
	Success   bool                `json:"success"`
	Errors    map[string][]string `json:"errors"`
	Data      interface{}         `json:"data"`
}

func generateErrors(err error, key string) map[string][]string {
	if err == nil {
		return nil
	}

	if key == "" {
		key = "system"
	}

	if vErr, ok := err.(service.ValidateErrors); ok {
		return vErr.MapFieldErrors()
	}
	return map[string][]string{
		key: {err.Error()},
	}
}

func tokenResp(token *models.Token) map[string]interface{} {
	var result = map[string]interface{}{
		"token":          token.UID,
		"ip":             token.IP,
		"availableTimes": token.AvailableTimes,
		"readOnly":       token.ReadOnly,
		"expiredAt":      token.ExpiredAt,
		"path":           token.Path,
		"secret":         token.Secret,
	}

	if token.ExpiredAt != nil {
		result["expiredAt"] = token.ExpiredAt.Unix()
	}

	if token.DeletedAt != nil {
		result["deletedAt"] = token.DeletedAt.Unix()
	}

	return result
}

func fileResp(file *models.File, db *gorm.DB) (map[string]interface{}, error) {
	var (
		err    error
		path   string
		result map[string]interface{}
	)

	if path, err = file.Path(db.Unscoped()); err != nil {
		return nil, err
	}

	if file.Object.ID == 0 {
		if err = db.Unscoped().Preload("Object").Find(file).Error; err != nil {
			return nil, err
		}
	}

	result = map[string]interface{}{
		"fileUid": file.UID,
		"path":    path,
		"size":    file.Size,
		"isDir":   file.IsDir,
		"hidden":  file.Hidden,
	}

	if file.IsDir == 0 {
		result["hash"] = file.Object.Hash
		result["ext"] = file.Ext
	}

	if file.DeletedAt != nil {
		result["deletedAt"] = file.DeletedAt.Unix()
	}

	return result, err
}
