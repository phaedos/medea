package models

import (
	"errors"
	"fmt"
	"io"
	"medea/pkg/utils"
	"path"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

const IsDir = int8(1)

const Hidden = int8(1)

var (
	ErrFileExisted       = errors.New("file has already existed")
	ErrOverwriteDir      = errors.New("directory can't be overwritten")
	ErrAppendToDir       = errors.New("can't append data to directory")
	ErrReadDir           = errors.New("can't read a directory")
	ErrAccessDenied      = errors.New("file can't be accessed by some tokens")
	ErrDeleteNonEmptyDir = errors.New("delete non-empty directory")
)

type File struct {
	ID            uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	UID           string     `gorm:"type:CHAR(32) NOT NULL;UNIQUE;column:uid"`
	PID           uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:pid"`
	AppID         uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:appId"`
	ObjectID      uint64     `gorm:"type:BIGINT(20) UNSIGNED NOT NULL;column:objectId"`
	Size          int        `gorm:"type:int;column:size"`
	Name          string     `gorm:"type:VARCHAR(255);NOT NULL;column:name"`
	Ext           string     `gorm:"type:VARCHAR(255);NOT NULL;column:ext"`
	IsDir         int8       `gorm:"type:tinyint;column:isDir;DEFAULT:0"`
	Hidden        int8       `gorm:"type:tinyint;column:hidden;DEFAULT:0"`
	DownloadCount uint64     `gorm:"type:BIGINT(20);column:downloadCount;DEFAULT:0"`
	CreatedAt     time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt     time.Time  `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`
	DeletedAt     *time.Time `gorm:"type:TIMESTAMP(6);INDEX;column:deletedAt"`

	App       App       `gorm:"foreignkey:appId;association_autoupdate:false;association_autocreate:false"`
	Object    Object    `gorm:"foreignkey:objectId;association_autoupdate:false;association_autocreate:false"`
	Parent    *File     `gorm:"foreignkey:id;association_foreignkey:pid;association_autoupdate:false;association_autocreate:false"`
	Children  []File    `gorm:"foreignkey:pid;association_foreignkey:id;association_autoupdate:false;association_autocreate:false"`
	Histories []History `gorm:"foreignkey:fileId;association_autoupdate:false;association_autocreate:false"`
}

func pathCacheKey(app *App, path string) string {
	return fmt.Sprintf("%d-%s", app.ID, path)
}

func (f *File) TableName() string {
	return "files"
}

func (f *File) executeDelete(forceDelete bool, db *gorm.DB) error {
	if f.IsDir == 0 {
		return db.Delete(f).Error
	}

	var err error

	if err = db.Preload("Children").Find(f).Error; err != nil {
		return err
	}

	if len(f.Children) == 0 {
		return db.Delete(f).Error
	}

	if forceDelete {
		for _, child := range f.Children {
			if err = child.executeDelete(forceDelete, db); err != nil {
				return err
			}
		}
		db.Model(f).Update("size", 0)
		return db.Delete(f).Error
	}
	return ErrDeleteNonEmptyDir
}

func (f *File) Delete(forceDelete bool, db *gorm.DB) (err error) {

	if f.Parent == nil {
		if err = db.Preload("Parent").Find(f).Error; err != nil {
			return err
		}
	}

	originSize := f.Size
	if err = f.executeDelete(forceDelete, db); err != nil {
		return err
	}

	if originSize != 0 {
		if err = f.Parent.UpdateParentSize(-originSize, db); err != nil {
			return err
		}
	}

	return db.Unscoped().Find(f).Error
}

func (f *File) CanBeAccessedByToken(token *Token, db *gorm.DB) error {
	var (
		err error
		p   string
	)
	if p, err = f.Path(db); err != nil {
		return err
	}
	if !strings.HasPrefix(p, token.Path) {
		return ErrAccessDenied
	}
	return nil
}

func (f *File) Reader(rootPath *string, db *gorm.DB) (io.ReadSeeker, error) {
	if f.IsDir == 1 {
		return nil, ErrReadDir
	}
	if f.Object.ID == 0 {
		db.Preload("Object").Find(&f)
	}
	return (&f.Object).Reader(rootPath, db)
}

func (f *File) Path(db *gorm.DB) (string, error) {

	if f.PID == 0 && f.IsDir == IsDir {
		return "/", nil
	}

	var (
		parts   []string
		current = *f
	)
	for {
		parts = append(parts, current.Name)
		if current.PID == 0 {
			break
		}
		temp := &File{}
		db.Where("id = ?", current.PID).Find(temp)
		current = *temp
	}

	utils.ReverseSlice(parts)

	return strings.Join(parts, "/"), nil
}

func (f *File) UpdateParentSize(size int, db *gorm.DB) error {
	var dirIds []uint64
	current := f
	for {
		current.Size += size
		dirIds = append(dirIds, current.ID)
		if current.PID == 0 {
			break
		}
		if current.Parent == nil {
			current.Parent = &File{}
		}
		if err := db.Model(current).Association("Parent").Find(current.Parent).Error; err != nil {
			return err
		}
		current = current.Parent
	}

	return db.Model(&File{}).Where("id in (?)", dirIds).UpdateColumn("size", gorm.Expr("size + ?", size)).Error
}

func (f *File) createHistory(objectID uint64, path string, db *gorm.DB) error {
	return db.Save(&History{ObjectID: objectID, FileID: f.ID, Path: path}).Error
}

func (f *File) OverWriteFromReader(reader io.Reader, hidden int8, rootPath *string, db *gorm.DB) (err error) {
	if f.IsDir == IsDir {
		return ErrOverwriteDir
	}

	var (
		p        string
		object   *Object
		sizeDiff int
	)

	if p, err = f.Path(db); err != nil {
		return err
	}

	if err := f.createHistory(f.ObjectID, p, db); err != nil {
		return err
	}

	if object, err = CreateObjectFromReader(reader, rootPath, db); err != nil {
		return err
	}

	f.Object = *object
	f.ObjectID = object.ID
	f.Hidden = hidden
	sizeDiff = object.Size - f.Size
	f.Size += sizeDiff

	if err = db.Model(f).Update(map[string]interface{}{
		"objectId": object.ID,
		"hidden":   hidden,
		"size":     f.Size,
	}).Error; err != nil {
		return err
	}
	db.Preload("Parent").Preload("App").Find(f)
	return f.Parent.UpdateParentSize(sizeDiff, db)
}

func (f *File) mustPath(db *gorm.DB) string {
	p, _ := f.Path(db)
	return p
}

func (f *File) MoveTo(newPath string, db *gorm.DB) (err error) {
	var (
		newPathDir      = path.Dir(newPath)
		newPathDirFile  *File
		newPathFileName = path.Base(newPath)
		newPathExt      = strings.TrimPrefix(path.Ext(newPathFileName), ".")
		previousPath    string
	)

	if previousPath, err = f.Path(db); err != nil {
		return err
	}

	if previousPath == newPath {
		return nil
	}

	if f.App.ID == 0 {
		if err = db.Preload("App").Find(f).Error; err != nil {
			return err
		}
	}

	if _, err := FindFileByPathWithTrashed(&f.App, newPath, db); err == nil {
		return ErrFileExisted
	}

	if newPathDirFile, err = CreateOrGetLastDirectory(&f.App, newPathDir, db); err != nil {
		return err
	}

	if f.IsDir == 0 {
		if err = f.createHistory(f.ObjectID, previousPath, db); err != nil {
			return err
		}
	}

	f.Name = newPathFileName
	f.Ext = newPathExt

	defer func() {
		pathToFileCache.Delete(previousPath)
		_ = pathToFileCache.Add(pathCacheKey(&f.App, f.mustPath(db)), f, time.Minute*10)
	}()

	if newPathDirFile.ID == f.PID {
		return db.Model(f).Updates(map[string]interface{}{"name": f.Name, "ext": f.Ext}).Error
	}

	if f.Parent == nil || f.Parent.ID == 0 {
		f.Parent = &File{}
		if err = db.Model(f).Association("Parent").Find(f.Parent).Error; err != nil {
			return err
		}
	}

	if err = newPathDirFile.UpdateParentSize(f.Size, db); err != nil {
		return err
	}

	if err = f.Parent.UpdateParentSize(-f.Size, db); err != nil {
		return err
	}
	f.PID = newPathDirFile.ID
	f.Parent = newPathDirFile
	f.Name = newPathFileName
	f.Ext = newPathExt

	return db.Model(f).Updates(map[string]interface{}{"pid": f.PID, "name": f.Name, "ext": f.Ext}).Error
}

func (f *File) AppendFromReader(reader io.Reader, hidden int8, rootPath *string, db *gorm.DB) (err error) {
	if f.IsDir == IsDir {
		return ErrAppendToDir
	}

	var (
		size   int
		object *Object
	)

	if err = db.Preload("Object").Preload("Parent").Preload("App").First(f).Error; err != nil {
		return err
	}

	if object, size, err = f.Object.AppendFromReader(reader, rootPath, db); err != nil {
		return err
	}

	f.Hidden = hidden
	f.Size += size
	f.Object = *object
	f.ObjectID = object.ID

	if err = db.Model(f).Updates(map[string]interface{}{"hidden": f.Hidden, "size": f.Size, "objectId": f.ObjectID}).Error; err != nil {
		return err
	}

	return f.Parent.UpdateParentSize(size, db)
}

func CreateOrGetLastDirectory(app *App, dirPath string, db *gorm.DB) (*File, error) {
	var (
		parent = &File{ID: 0}
		err    error
		parts  = strings.Split(strings.TrimRight(strings.TrimSpace(dirPath), "/"), "/")
	)

	if parts[0] != "" {
		if parent, err = CreateOrGetRootPath(app, db); err != nil {
			return nil, err
		}
	}

	for _, part := range parts {
		file := &File{}
		if err = db.Where("appId = ? and pid = ? and name = ?", app.ID, parent.ID, part).First(file).Error; err != nil {
			if !utils.IsRecordNotFound(err) {
				return nil, err
			}
			file.AppID = app.ID
			file.PID = parent.ID
			file.Name = part
			file.IsDir = 1
			file.UID = UID()
			if err = db.Save(file).Error; err != nil {
				return nil, err
			}
		}
		parent = file
	}
	parent.App = *app
	return parent, nil
}

func CreateOrGetRootPath(app *App, db *gorm.DB) (*File, error) {
	var (
		file = &File{}
		err  error
	)
	err = db.Where("appId = ? and pid = 0 and name = ''", app.ID).First(file).Error
	file.App = *app
	return file, err
}

func CreateFileFromReader(app *App, savePath string, reader io.Reader, hidden int8, rootPath *string, db *gorm.DB) (file *File, err error) {
	var (
		object    *Object
		parentDir *File
		dirPrefix = path.Dir(savePath)
		fileName  = path.Base(savePath)
	)

	if f, err := FindFileByPathWithTrashed(app, savePath, db); err == nil && f.ID > 0 {
		return nil, ErrFileExisted
	}

	if parentDir, err = CreateOrGetLastDirectory(app, dirPrefix, db); err != nil {
		return nil, err
	}

	if object, err = CreateObjectFromReader(reader, rootPath, db); err != nil {
		return nil, err
	}

	file = &File{
		UID:      UID(),
		PID:      parentDir.ID,
		AppID:    app.ID,
		ObjectID: object.ID,
		Size:     object.Size,
		Name:     fileName,
		Ext:      strings.TrimPrefix(path.Ext(fileName), "."),
		Hidden:   hidden,
		Object:   *object,
		App:      *app,
		Parent:   parentDir,
	}

	if err = db.Create(file).Error; err != nil {
		return nil, err
	}

	return file, parentDir.UpdateParentSize(object.Size, db)
}

func FindFileByUID(uid string, trashed bool, db *gorm.DB) (*File, error) {
	var (
		file = &File{}
		err  error
	)
	if trashed {
		db = db.Unscoped()
	}
	if err = db.Where("uid = ?", uid).Find(file).Error; err != nil {
		return file, err
	}
	return file, nil
}

func FindFileByPathWithTrashed(app *App, path string, db *gorm.DB) (*File, error) {
	return FindFileByPath(app, path, db.Unscoped(), true)
}

func FindFileByPath(app *App, path string, db *gorm.DB, useCache bool) (*File, error) {
	var cacheKey = pathCacheKey(app, path)

	if useCache {
		if fileValue, ok := pathToFileCache.Get(cacheKey); ok {
			file := fileValue.(*File)
			if err := db.Where("id = ?", file.ID).Find(file).Error; err == nil {
				file.App = *app
				return file, nil
			}
		}
	}

	var (
		err    error
		parent = &File{}
		parts  = strings.Split(strings.Trim(strings.TrimSpace(path), "/"), "/")
	)

	if parts[0] != "" {
		if parent, err = CreateOrGetRootPath(app, db); err != nil {
			return nil, err
		}
	}

	for _, part := range parts {
		var file = &File{}
		if err = db.Where("appId = ? and pid = ? and name = ?", app.ID, parent.ID, part).First(file).Error; err != nil {
			return nil, err
		}
		parent = file
	}
	parent.App = *app

	_ = pathToFileCache.Add(cacheKey, parent, time.Minute*10)

	return parent, nil
}
