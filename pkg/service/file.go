package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math"
	"medea/pkg/database/models"
	"medea/pkg/utils"
	libPath "path"
	"strings"

	"github.com/go-playground/validator"
	"github.com/jinzhu/gorm"
)

var (
	ErrPathExisted                  = errors.New("the path has already existed")
	ErrOnlyOneRenameAppendOverWrite = errors.New("only one of rename, append and overwrite is allowed")
	ErrFileHasBeenDeleted           = errors.New("the file has been deleted")
)

type FileCreate struct {
	BaseService

	Token     *models.Token `validate:"required"`
	Path      string        `validate:"required,max=1000"`
	Hidden    int8          `validate:"oneof=0 1"`
	IP        *string       `validate:"omitempty"`
	Reader    io.Reader     `validate:"omitempty"`
	Overwrite int8          `validate:"oneof=0 1"`
	Rename    int8          `validate:"oneof=0 1"`
	Append    int8          `validate:"oneof=0 1"`
}

func (fc *FileCreate) Validate() ValidateErrors {
	var (
		err            error
		validateErrors ValidateErrors
	)

	if fc.Overwrite+fc.Rename+fc.Append > 1 {
		validateErrors = append(
			validateErrors,
			generateErrorByField("FileCreate.Operate", ErrOnlyOneRenameAppendOverWrite),
		)
	}

	if err = Validate.Struct(fc); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err = ValidateToken(fc.DB, fc.IP, false, fc.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileCreate.Token", err))
	}

	if !ValidatePath(fc.Path) {
		validateErrors = append(validateErrors, generateErrorByField("FileCreate.Path", ErrInvalidPath))
	}

	return validateErrors
}

func (fc *FileCreate) Execute(ctx context.Context) (interface{}, error) {
	var (
		err   error
		path  = fc.Token.PathWithScope(fc.Path)
		file  *models.File
		inTrx = utils.InTransaction(fc.DB)
	)

	if !inTrx {
		fc.DB = fc.DB.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		defer func() {
			if reErr := recover(); reErr != nil {
				fc.DB.Rollback()
			}
		}()
		defer func() { err = fc.DB.Commit().Error }()
	}

	if err = fc.Token.UpdateAvailableTimes(-1, fc.DB); err != nil {
		return nil, err
	}

	if fc.Reader == nil {
		return models.CreateOrGetLastDirectory(&fc.Token.App, path, fc.DB)
	}

	if file, err = models.FindFileByPathWithTrashed(&fc.Token.App, path, fc.DB); err != nil && !utils.IsRecordNotFound(err) {
		return nil, err
	}

	if file == nil || file.ID == 0 {
		return models.CreateFileFromReader(&fc.Token.App, path, fc.Reader, fc.Hidden, fc.RootPath, fc.DB)
	}

	if file.DeletedAt != nil && (fc.Append == 1 || fc.Overwrite == 1) {
		return nil, ErrFileHasBeenDeleted
	}

	if fc.Overwrite == 1 {
		return file, file.OverWriteFromReader(fc.Reader, fc.Hidden, fc.RootPath, fc.DB)
	}

	if fc.Append == 1 {
		return file, file.AppendFromReader(fc.Reader, fc.Hidden, fc.RootPath, fc.DB)
	}

	if fc.Rename == 1 {
		var (
			dir      = libPath.Dir(path)
			basename = libPath.Base(path)
		)
		path = fmt.Sprintf("%s/%s_%s", dir, models.RandomWithMD5(256), basename)
		return models.CreateFileFromReader(&fc.Token.App, path, fc.Reader, fc.Hidden, fc.RootPath, fc.DB)
	}

	return nil, ErrPathExisted
}

var ErrReadHiddenFile = errors.New("try to read the hidden file")

type FileRead struct {
	BaseService

	Token *models.Token `validate:"required"`
	File  *models.File  `validate:"required"`
	IP    *string       `validate:"omitempty"`
}

func (fr *FileRead) Validate() ValidateErrors {
	var (
		validateErrors ValidateErrors
		errs           error
	)
	if errs = Validate.Struct(fr); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err := ValidateToken(fr.DB, fr.IP, true, fr.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileRead.Token", err))
	}

	if err := ValidateFile(fr.DB, fr.File); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileRead.File", err))
	} else {
		if err := fr.File.CanBeAccessedByToken(fr.Token, fr.DB); err != nil {
			validateErrors = append(validateErrors, generateErrorByField("FileRead.Token", err))
		}
	}

	return validateErrors
}

func (fr *FileRead) Execute(ctx context.Context) (interface{}, error) {
	var err error

	if err = fr.Token.UpdateAvailableTimes(-1, fr.DB); err != nil {
		return nil, err
	}

	if fr.File.Hidden == 1 {
		return nil, ErrReadHiddenFile
	}

	return fr.File.Reader(fr.RootPath, fr.DB)
}

var ErrListFile = errors.New("can't list the content of a file")

type DirectoryListResponse struct {
	Total int
	Pages int
	Files []models.File
}

type DirectoryList struct {
	BaseService

	Token  *models.Token `validate:"required"`
	IP     *string       `validate:"omitempty"`
	SubDir string        `validate:"omitempty"`
	Sort   string        `validate:"required,oneof=type -type name -name time -time"`
	Offset int           `validate:"omitempty,min=0"`
	Limit  int           `validate:"required,min=10,max=20"`
}

func (dl *DirectoryList) Validate() ValidateErrors {
	var (
		err            error
		validateErrors ValidateErrors
	)

	if err = Validate.Struct(dl); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err = ValidateToken(dl.DB, dl.IP, false, dl.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("DirectoryList.Token", err))
	}

	if !ValidatePath(dl.SubDir) {
		validateErrors = append(validateErrors, generateErrorByField("DirectoryList.SubDir", ErrInvalidPath))
	}

	return validateErrors
}

func (dl *DirectoryList) Execute(ctx context.Context) (interface{}, error) {
	var (
		err     error
		dir     *models.File
		total   int
		pages   int
		dirPath = dl.Token.PathWithScope(dl.SubDir)
	)

	if err = dl.Token.UpdateAvailableTimes(-1, dl.DB); err != nil {
		return nil, err
	}

	if dir, err = models.FindFileByPath(&dl.Token.App, dirPath, dl.DB, false); err != nil {
		return nil, err
	}

	if dir.IsDir == 0 {
		return nil, ErrListFile
	}

	total = dl.DB.Model(dir).Association("Children").Count()
	pages = int(math.Ceil(float64(total) / float64(dl.Limit)))

	if err = dl.DB.Preload("Children", func(db *gorm.DB) *gorm.DB {
		var (
			order = "DESC"
			key   = "isDir"
		)
		if !strings.HasPrefix(dl.Sort, "-") {
			order = "ASC"
		}
		switch strings.TrimPrefix(dl.Sort, "-") {
		case "type":
			key = "isDir"
		case "name":
			key = "name"
		case "time":
			key = "updatedAt"
		}
		return db.Order(key + " " + order).Offset(dl.Offset).Limit(dl.Limit)
	}).First(dir).Error; err != nil {
		return nil, err
	}

	return &DirectoryListResponse{
		Total: total,
		Pages: pages,
		Files: dir.Children,
	}, nil
}

type FileUpdate struct {
	BaseService

	Token  *models.Token `validate:"required"`
	File   *models.File  `validate:"required"`
	IP     *string       `validate:"omitempty"`
	Hidden *int8         `validate:"omitempty,oneof=0 1"`
	Path   *string       `validate:"omitempty,max=1000"`
}

func (fu *FileUpdate) Validate() ValidateErrors {
	var (
		validateErrors ValidateErrors
		errs           error
	)
	if errs = Validate.Struct(fu); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err := ValidateToken(fu.DB, fu.IP, true, fu.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileUpdate.Token", err))
	}

	if err := ValidateFile(fu.DB, fu.File); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileUpdate.File", err))
	} else {
		if err := fu.File.CanBeAccessedByToken(fu.Token, fu.DB); err != nil {
			validateErrors = append(validateErrors, generateErrorByField("FileUpdate.Token", err))
		}
	}

	if fu.Path != nil {
		if !ValidatePath(*fu.Path) {
			validateErrors = append(validateErrors, generateErrorByField("FileUpdate.Path", ErrInvalidPath))
		}
	}

	return validateErrors
}

func (fu *FileUpdate) Execute(ctx context.Context) (interface{}, error) {
	var (
		err   error
		inTrx = utils.InTransaction(fu.DB)
	)

	if !inTrx {
		fu.DB = fu.DB.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		defer func() {
			if reErr := recover(); reErr != nil {
				fu.DB.Rollback()
			}
		}()
		defer func() { err = fu.DB.Commit().Error }()
	}

	if err = fu.Token.UpdateAvailableTimes(-1, fu.DB); err != nil {
		return nil, err
	}

	if fu.Path != nil {
		if err := fu.File.MoveTo(fu.Token.PathWithScope(*fu.Path), fu.DB); err != nil {
			return nil, err
		}
	}

	if fu.Hidden != nil {
		fu.File.Hidden = *fu.Hidden
	}

	return fu.File, fu.DB.Save(fu.File).Error
}

type FileDelete struct {
	BaseService

	Token *models.Token `validate:"required"`
	File  *models.File  `validate:"required"`
	Force *bool         `validate:"omitempty"`
	IP    *string       `validate:"omitempty"`
}

func (fd *FileDelete) Validate() ValidateErrors {
	var (
		err            error
		validateErrors ValidateErrors
	)
	if err = Validate.Struct(fd); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validateErrors = append(validateErrors, PreDefinedValidateErrors[err.Namespace()])
		}
	}

	if err = ValidateToken(fd.DB, fd.IP, true, fd.Token); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileDelete.Token", err))
	}

	if err = ValidateFile(fd.DB, fd.File); err != nil {
		validateErrors = append(validateErrors, generateErrorByField("FileDelete.File", err))
	} else {
		if err = fd.File.CanBeAccessedByToken(fd.Token, fd.DB); err != nil {
			validateErrors = append(validateErrors, generateErrorByField("FileDelete.Token", err))
		}
	}

	return validateErrors
}

func (fd *FileDelete) Execute(ctx context.Context) (interface{}, error) {
	var (
		falseValue = false
		err        error
		inTrx      = utils.InTransaction(fd.DB)
	)

	if !inTrx {
		fd.DB = fd.DB.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		defer func() {
			if reErr := recover(); reErr != nil {
				fd.DB.Rollback()
			}
		}()
		defer func() { err = fd.DB.Commit().Error }()
	}

	if err = fd.Token.UpdateAvailableTimes(-1, fd.DB); err != nil {
		return nil, err
	}

	if fd.Force == nil {
		fd.Force = &falseValue
	}

	if err = fd.File.Delete(*fd.Force, fd.DB); err != nil {
		return fd.File, err
	}

	return fd.File, nil
}
