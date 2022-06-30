package models

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"medea/pkg/config"
	"medea/pkg/utils"

	"github.com/jinzhu/gorm"
)

const ChunkSize = 2 << 20

var (
	ErrInvalidChunkID   = errors.New("invalid chunk id")
	ErrChunkExceedLimit = fmt.Errorf("total length exceed limit: %d bytes", ChunkSize)
)

type Chunk struct {
	ID        uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	Size      int       `gorm:"type:int;column:size"`
	Hash      string    `gorm:"type:CHAR(64) NOT NULL;UNIQUE;column:hash"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`
}

func (c Chunk) TableName() string {
	return "chunks"
}

func (c *Chunk) Reader(rootPath *string) (file *os.File, err error) {
	var path string
	if path, err = c.Path(rootPath); err != nil {
		return
	}
	return os.Open(path)
}

func (c Chunk) Path(rootPath *string) (path string, err error) {
	var (
		idStr string
		parts []string
		index int
		dir   string
	)

	if rootPath == nil {
		rootPath = &config.DefaultConfig.Chunk.RootPath
	}
	if c.ID < 10000 {
		return "", ErrInvalidChunkID
	}
	idStr = strconv.FormatUint(c.ID, 10)
	parts = make([]string, (len(idStr)/3)+1)
	for ; len(idStr) > 3; index++ {
		parts[index] = utils.SubStrFromToEnd(idStr, -3)
		idStr = utils.SubStrFromTo(idStr, 0, -3)
	}
	parts[index] = idStr
	parts = parts[1:]
	utils.ReverseSlice(parts)
	dir = filepath.Join(strings.TrimSuffix(*rootPath, string(os.PathSeparator)), filepath.Join(parts...))
	path = filepath.Join(dir, strconv.FormatUint(c.ID, 10))
	if !utils.IsDir(dir) {
		err = os.MkdirAll(dir, os.ModePerm)
	}
	return path, err
}

func (c *Chunk) AppendBytes(p []byte, rootPath *string, db *gorm.DB) (chunk *Chunk, writeCount int, err error) {
	var (
		file       *os.File
		buf        bytes.Buffer
		oldContent []byte
		hash       string
		path       string
	)

	if len(p) > ChunkSize-c.Size {
		return nil, 0, ErrChunkExceedLimit
	}

	if path, err = c.Path(rootPath); err != nil {
		return
	}

	if oldContent, err = ioutil.ReadFile(path); err != nil {
		return
	}
	buf.Write(oldContent)
	buf.Write(p)

	if hash, err = utils.Sha256Hash2String(buf.Bytes()); err != nil {
		return
	}

	if chunk, err = FindChunkByHash(hash, db); err == nil {
		return chunk, len(p), nil
	}

	if count, err := CountObjectChunkByChunkID(c.ID, db); err != nil {
		return nil, 0, err
	} else if count > 1 {
		newChunk, err := CreateChunkFromBytes(buf.Bytes(), rootPath, db)
		if err != nil {
			return nil, 0, err
		}
		return newChunk, len(p), nil
	}

	c.Size = buf.Len()
	c.Hash = hash

	if file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644); err != nil {
		return nil, 0, err
	}
	defer file.Close()

	if writeCount, err = file.Write(p); err != nil {
		return c, 0, err
	}

	return c, writeCount, db.Model(c).Updates(map[string]interface{}{"size": c.Size, "hash": c.Hash}).Error
}

func CreateChunkFromBytes(p []byte, rootPath *string, db *gorm.DB) (chunk *Chunk, err error) {
	var (
		size    int
		path    string
		hashStr string
	)

	if size = len(p); int64(size) > ChunkSize {
		return nil, fmt.Errorf("the size of chunk must be less than %d bytes", ChunkSize)
	}

	if hashStr, err = utils.Sha256Hash2String(p); err != nil {
		return nil, err
	}

	if chunk, err = FindChunkByHash(hashStr, db); err == nil {
		return chunk, nil
	}

	chunk = &Chunk{
		Size: size,
		Hash: hashStr,
	}

	if err = db.Set("gorm:insert_option", "ON DUPLICATE KEY UPDATE id=id").Create(chunk).Error; err != nil {
		return nil, err
	}

	if path, err = chunk.Path(rootPath); err != nil {
		return chunk, err
	}

	if err = ioutil.WriteFile(path, p, 0644); err != nil {
		return nil, err
	}

	return chunk, err
}

func FindChunkByHash(h string, db *gorm.DB) (*Chunk, error) {
	var chunk Chunk
	return &chunk, db.Where("hash = ?", h).First(&chunk).Error
}

func CreateEmptyContentChunk(rootPath *string, db *gorm.DB) (chunk *Chunk, err error) {
	var (
		path             string
		emptyContentHash string
	)
	if emptyContentHash, err = utils.Sha256Hash2String(nil); err != nil {
		return nil, err
	}

	if chunk, err = FindChunkByHash(emptyContentHash, db); err == nil {
		return chunk, nil
	}

	chunk = &Chunk{Size: 0, Hash: emptyContentHash}

	if err = db.Create(chunk).Error; err != nil {
		return nil, err
	}

	if path, err = chunk.Path(rootPath); err != nil {
		return chunk, err
	}

	return chunk, ioutil.WriteFile(path, nil, 0644)
}
