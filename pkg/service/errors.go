package service

import (
	"bytes"
	"errors"
	"fmt"
)

var (
	ErrInvalidPath = errors.New("path is not a legal unix path")
)

type ValidateError struct {
	Msg       string `json:"msg"`
	Field     string `json:"field"`
	Code      int    `json:"code"`
	Exception error
}

func (v *ValidateError) Error() string {
	if v.Exception != nil {
		return v.Exception.Error()
	}
	return fmt.Sprintf("code: %d, field: %s, validate error: %s", v.Code, v.Field, v.Msg)
}

type ValidateErrors []*ValidateError

func (v ValidateErrors) Error() string {
	var (
		buf = bytes.NewBufferString("")
	)
	for i := 0; i < len(v); i++ {
		buf.WriteString(v[i].Error())
		buf.WriteString("\n")
	}
	return buf.String()
}

func (v ValidateErrors) MapFieldErrors() map[string][]string {
	var m = make(map[string][]string, len(v))
	for i := 0; i < len(v); i++ {
		m[v[i].Field] = []string{v[i].Error()}
	}
	return m
}

func (v ValidateErrors) Map() map[int]string {
	var (
		m = make(map[int]string, len(v))
	)
	for i := 0; i < len(v); i++ {
		m[v[i].Code] = v[i].Error()
	}
	return m
}

func (v ValidateErrors) ContainsErrCode(code int) bool {
	for i := 0; i < len(v); i++ {
		if v[i].Code == code {
			return true
		}
	}
	return false
}

var (
	PreDefinedValidateErrors = map[string]*ValidateError{
		"TokenCreate.App": {
			Code:  10002,
			Field: "TokenCreate.App",
			Msg:   "can't find specific application by input params",
		},
		"TokenCreate.Path": {
			Code:  10003,
			Field: "TokenCreate.Path",
			Msg:   "path of token can't be empty, max of length is 1000, and must be a legal unix path",
		},
		"TokenCreate.IP": {
			Code:  10004,
			Field: "TokenCreate.Ip",
			Msg:   "max length of ip is 1500",
		},
		"TokenCreate.Secret": {
			Code:  10005,
			Field: "TokenCreate.Secret",
			Msg:   "secret of token is 32",
		},
		"TokenCreate.AvailableTimes": {
			Code:  10006,
			Field: "TokenCreate.AvailableTimes",
			Msg:   "availableTimes of token is greater than -1",
		},
		"TokenCreate.ReadOnly": {
			Code:  10007,
			Field: "TokenCreate.ReadOnly",
			Msg:   "readOnly of token is 0 or 1",
		},

		"TokenUpdate.Token": {
			Code:  10008,
			Field: "TokenUpdate.Token",
			Msg:   "token is required",
		},
		"TokenUpdate.IP": {
			Code:  10009,
			Field: "TokenUpdate.IP",
			Msg:   "max length of ip is 1500, it's optional",
		},
		"TokenUpdate.Path": {
			Code:  10010,
			Field: "TokenUpdate.Path",
			Msg:   "max length of ip is 1000, and must be a legal unix path, is's optional",
		},
		"TokenUpdate.Secret": {
			Code:  10011,
			Field: "TokenUpdate.Secret",
			Msg:   "the length of secret is 32, it's optional",
		},
		"TokenUpdate.ReadOnly": {
			Code:  10012,
			Field: "TokenUpdate.ReadOnly",
			Msg:   "readOnly is 1 or 0, it's optional",
		},
		"TokenUpdate.ExpiredAt": {
			Code:  10013,
			Field: "TokenUpdate.ExpiredAt",
			Msg:   "expiredAt must be greater than now, it's optional",
		},
		"TokenUpdate.AvailableTimes": {
			Code:  10014,
			Field: "TokenUpdate.AvailableTimes",
			Msg:   "availableTimes must be a integer, and must be greater than -1, it's optional",
		},

		"FileCreate.App": {
			Code:  10015,
			Field: "FileCreate.App",
			Msg:   "can't find specific application by input params",
		},
		"FileCreate.Token": {
			Code:  10016,
			Field: "FileCreate.Token",
			Msg:   "can't find specific token by input params",
		},
		"FileCreate.Path": {
			Code:  10017,
			Field: "FileCreate.Path",
			Msg:   "path of file or directory can't be empty, max of length is 1000, and must be a legal unix path",
		},
		"FileCreate.Hidden": {
			Code:  10018,
			Field: "FileCreate.Hidden",
			Msg:   "hidden must be 0 or 1",
		},
		"FileCreate.Overwrite": {
			Code:  10019,
			Field: "FileCreate.Overwrite",
			Msg:   "overwrite must be 0 or 1",
		},
		"FileCreate.Rename": {
			Code:  10020,
			Field: "FileCreate.Rename",
			Msg:   "rename must be 0 or 1",
		},
		"FileCreate.Append": {
			Code:  10021,
			Field: "FileCreate.Append",
			Msg:   "append must be 0 or 1",
		},
		"FileCreate.Operate": {
			Code:  10022,
			Field: "FileCreate.Operate",
			Msg:   ErrOnlyOneRenameAppendOverWrite.Error(),
		},

		"FileRead.Token": {
			Code:  10023,
			Field: "FileRead.Token",
			Msg:   "token is required",
		},
		"FileRead.File": {
			Code:  10024,
			Field: "FileRead.File",
			Msg:   "file is required",
		},

		"FileUpdate.Token": {
			Code:  10025,
			Field: "FileUpdate.Token",
			Msg:   "token is required",
		},
		"FileUpdate.File": {
			Code:  10026,
			Field: "FileUpdate.File",
			Msg:   "file is required",
		},
		"FileUpdate.Hidden": {
			Code:  10027,
			Field: "FileUpdate.Hidden",
			Msg:   "file is required",
		},
		"FileUpdate.Path": {
			Code:  10028,
			Field: "FileUpdate.Path",
			Msg:   "file is required",
		},

		"FileDelete.Token": {
			Code:  10029,
			Field: "FileDelete.Token",
			Msg:   "token is required",
		},
		"FileDelete.File": {
			Code:  10030,
			Field: "FileDelete.File",
			Msg:   "file is required",
		},

		"DirectoryList.Token": {
			Code:  10031,
			Field: "DirectoryList.Token",
			Msg:   "token is required",
		},
		"DirectoryList.Sort": {
			Code:  10032,
			Field: "DirectoryList.Sort",
			Msg:   "sort is only allowed to be one of type -type name -name time -time",
		},
		"DirectoryList.SubDir": {
			Code:  10033,
			Field: "DirectoryList.SubDir",
			Msg:   "subDir must be a legal unix path",
		},
		"DirectoryList.Offset": {
			Code:  10034,
			Field: "DirectoryList.Offset",
			Msg:   "the min value of offset is 0",
		},
		"DirectoryList.Limit": {
			Code:  10035,
			Field: "DirectoryList.Limit",
			Msg:   "the min value of limit is 10, and max of limit 20",
		},
	}
)

func generateErrorByField(field string, err error) *ValidateError {
	return &ValidateError{
		Code:      PreDefinedValidateErrors[field].Code,
		Field:     field,
		Exception: err,
	}
}
