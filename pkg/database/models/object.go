package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"time"

	sha2562 "medea/pkg/utils/sha256"

	"github.com/jinzhu/gorm"
)

type Object struct {
	ID        uint64    `gorm:"type:BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT;primary_key"`
	Size      int       `gorm:"type:int;column:size"`
	Hash      string    `gorm:"type:CHAR(64) NOT NULL;UNIQUE;column:hash"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:createdAt"`
	UpdatedAt time.Time `gorm:"type:TIMESTAMP(6) NOT NULL;DEFAULT:CURRENT_TIMESTAMP(6);column:updatedAt"`

	Files        []File        `gorm:"foreignkey:objectId;association_autoupdate:false;association_autocreate:false"`
	Chunks       []Chunk       `gorm:"many2many:object_chunk;association_jointable_foreignkey:chunkId;jointable_foreignkey:objectId;association_autoupdate:false;association_autocreate:false"`
	ObjectChunks []ObjectChunk `gorm:"foreignkey:objectId;association_autoupdate:false;association_autocreate:false"`
	Histories    []History     `gorm:"foreignkey:objectId;association_autoupdate:false;association_autocreate:false"`
}

func (o Object) TableName() string {
	return "objects"
}

func (o *Object) FileCountWithTrashed(db *gorm.DB) int {
	return db.Unscoped().Model(o).Association("Files").Count()
}

func (o *Object) ChunkCount(db *gorm.DB) int {
	return db.Model(o).Association("Chunks").Count()
}

func (o *Object) ChunkWithNumber(number int, db *gorm.DB) (chunk *Chunk, err error) {
	var joinObjectChunk = "join object_chunk on object_chunk.chunkId = chunks.id and object_chunk.objectId = ? and number = ?"
	chunk = &Chunk{}
	err = db.Joins(joinObjectChunk, o.ID, number).First(chunk).Error
	return chunk, err
}

func (o *Object) LastChunk(db *gorm.DB) (*Chunk, error) {
	var (
		joinObjectChunk = "join object_chunk on object_chunk.chunkId = chunks.id and object_chunk.objectId = ?"
		chunk           = &Chunk{}
		err             error
	)
	err = db.Joins(joinObjectChunk, o.ID).Order("chunks.id desc").First(chunk).Error
	return chunk, err
}

func (o *Object) LastChunkNumber(db *gorm.DB) (int, error) {
	var (
		oc     *ObjectChunk
		err    error
		number int
	)
	oc, err = o.LastObjectChunk(db)
	if oc != nil {
		return oc.Number, nil
	}
	return number, err
}

func (o *Object) LastObjectChunk(db *gorm.DB) (*ObjectChunk, error) {
	err := db.Preload("ObjectChunks", func(db *gorm.DB) *gorm.DB {
		return db.Order("object_chunk.id desc").Limit(1)
	}).Find(o).Error
	if len(o.ObjectChunks) == 0 {
		return nil, err
	}
	return &o.ObjectChunks[0], nil
}

func (o *Object) completeLastChunk(
	reader io.Reader,
	object *Object,
	stateHash hash.Hash,
	readerContentLen *int,
	rootPath *string,
	db *gorm.DB,
) (err error) {
	var lastChunk *Chunk
	if lastChunk, err = o.LastChunk(db); err != nil {
		return err
	}
	if lackSize := ChunkSize - lastChunk.Size; lackSize > 0 {
		var (
			chunk          *Chunk
			hashState      string
			lackContentBuf = bytes.NewBuffer(nil)
		)

		for {
			var (
				readCount   int
				lackContent = make([]byte, lackSize-lackContentBuf.Len())
			)
			if readCount, err = reader.Read(lackContent); err != nil {
				if err == io.EOF {
					break
				}
			}
			if _, err = lackContentBuf.Write(lackContent[:readCount]); err != nil {
				return err
			}
			if lackContentBuf.Len() == lackSize {
				break
			}
		}

		lackContent := lackContentBuf.Bytes()
		if chunk, _, err = lastChunk.AppendBytes(lackContent, rootPath, db); err != nil {
			return err
		}
		if chunk.ID != lastChunk.ID {
			object.ObjectChunks[len(object.ObjectChunks)-1].ChunkID = chunk.ID
		}
		if _, err := stateHash.Write(lackContent); err != nil {
			return err
		}
		if hashState, err = sha2562.GetHashStateText(stateHash); err != nil {
			return err
		}
		*readerContentLen += len(lackContent)
		object.ObjectChunks[len(object.ObjectChunks)-1].HashState = &hashState
	}
	return nil
}

func (o *Object) appendRestContent(
	lastOc *ObjectChunk,
	reader io.Reader,
	object *Object,
	stateHash hash.Hash,
	readerContentLen *int,
	rootPath *string,
	db *gorm.DB,
) (err error) {
	restContentBuf := bytes.NewBuffer(nil)
	restContentReadOver := false
	for index := lastOc.Number; ; {
		var (
			chunk     *Chunk
			content   = make([]byte, ChunkSize-restContentBuf.Len())
			readLen   int
			hashState string
		)

		if readLen, err = reader.Read(content); err != nil {
			if err == io.EOF {
				restContentReadOver = true
			} else {
				return err
			}
		}

		if readLen > 0 {
			if _, err = restContentBuf.Write(content[:readLen]); err != nil {
				return err
			}
		}

		if (restContentReadOver || restContentBuf.Len() == ChunkSize) && restContentBuf.Len() > 0 {
			writeContent := restContentBuf.Bytes()
			if chunk, err = CreateChunkFromBytes(writeContent, rootPath, db); err != nil {
				return err
			}
			if _, err := stateHash.Write(writeContent); err != nil {
				return err
			}
			if hashState, err = sha2562.GetHashStateText(stateHash); err != nil {
				return err
			}
			object.ObjectChunks = append(object.ObjectChunks, ObjectChunk{
				ChunkID:   chunk.ID,
				Number:    index + 1,
				HashState: &hashState,
			})
			*readerContentLen += len(writeContent)
			restContentBuf.Reset()
			index++
		}
		if restContentReadOver {
			break
		}
	}
	return nil
}

func (o *Object) AppendFromReader(reader io.Reader, rootPath *string, db *gorm.DB) (object *Object, readerContentLen int, err error) {
	var (
		lastOc     *ObjectChunk
		stateHash  hash.Hash
		objectSize = o.Size
	)
	if lastOc, err = o.LastObjectChunk(db); err != nil {
		return o, readerContentLen, err
	}
	if stateHash, err = sha2562.NewHashWithStateText(*lastOc.HashState); err != nil {
		return o, readerContentLen, err
	}
	object = &Object{}
	if err = db.Where("objectId = ?", o.ID).Find(&object.ObjectChunks).Error; err != nil {
		return o, readerContentLen, err
	}

	if o.FileCountWithTrashed(db)+db.Model(o).Association("Histories").Count() <= 1 {
		object.ID = o.ID
		object.CreatedAt = o.CreatedAt
		object.UpdatedAt = o.UpdatedAt
	} else {
		for index := range object.ObjectChunks {
			object.ObjectChunks[index].ID = 0
		}
	}

	if err = o.completeLastChunk(reader, object, stateHash, &readerContentLen, rootPath, db); err != nil {
		return o, readerContentLen, err
	}

	if err = o.appendRestContent(lastOc, reader, object, stateHash, &readerContentLen, rootPath, db); err != nil {
		return o, readerContentLen, err
	}

	objectHashValue := hex.EncodeToString(stateHash.Sum(nil))
	if object, err := FindObjectByHash(objectHashValue, db); err == nil && object != nil {
		return object, readerContentLen, nil
	}

	object.Size = objectSize + readerContentLen
	object.Hash = objectHashValue
	if err = db.Save(object).Error; err != nil {
		return
	}

	for _, objectChunk := range object.ObjectChunks {
		objectChunk.ObjectID = object.ID
		if err = db.Save(&objectChunk).Error; err != nil {
			return
		}
	}

	return object, readerContentLen, nil
}

func (o *Object) Reader(rootPath *string, db *gorm.DB) (io.ReadSeeker, error) {
	return NewObjectReader(o, rootPath, db)
}

func FindObjectByHash(h string, db *gorm.DB) (*Object, error) {
	var object Object
	var err = db.Where("hash = ?", h).First(&object).Error
	return &object, err
}

func createChunksForObject(
	reader io.Reader,
	rootPath *string,
	objectHash hash.Hash,
	db *gorm.DB,
) (oc []ObjectChunk, size int, err error) {
	var (
		readerOver bool
		readerBuf  = bytes.NewBuffer(nil)
	)
	for index := 0; ; {
		var (
			chunk     *Chunk
			content   = make([]byte, ChunkSize-readerBuf.Len())
			readLen   int
			hashState string
		)
		if readLen, err = reader.Read(content); err != nil {
			if err == io.EOF {
				err = nil
				readerOver = true
			} else {
				return
			}
		}

		if readLen > 0 {
			if _, err = readerBuf.Write(content[:readLen]); err != nil {
				return
			}
		}

		if (readerOver || readerBuf.Len() == ChunkSize) && readerBuf.Len() > 0 {
			writeContent := readerBuf.Bytes()
			if chunk, err = CreateChunkFromBytes(writeContent, rootPath, db); err != nil {
				return
			}
			if _, err = objectHash.Write(writeContent); err != nil {
				return
			}
			if hashState, err = sha2562.GetHashStateText(objectHash); err != nil {
				return
			}
			oc = append(oc, ObjectChunk{
				ChunkID:   chunk.ID,
				Number:    index + 1,
				HashState: &hashState,
			})
			size += len(writeContent)
			readerBuf.Reset()
			index++
		}
		if readerOver {
			break
		}
	}
	return
}

func CreateObjectFromReader(reader io.Reader, rootPath *string, db *gorm.DB) (object *Object, err error) {
	var (
		oc         []ObjectChunk
		size       int
		objectHash = sha256.New()
	)

	if oc, size, err = createChunksForObject(reader, rootPath, objectHash, db); err != nil {
		return nil, err
	}

	if size == 0 {
		return CreateEmptyObject(rootPath, db)
	}

	objectHashValue := hex.EncodeToString(objectHash.Sum(nil))
	if object, err = FindObjectByHash(objectHashValue, db); err == nil && object != nil {
		return object, nil
	}

	object = &Object{Size: size, Hash: objectHashValue}
	if err = db.Save(object).Error; err != nil {
		return
	}

	for _, objectChunk := range oc {
		objectChunk.ObjectID = object.ID
		if err = db.Save(&objectChunk).Error; err != nil {
			return
		}
	}
	return object, nil
}

func CreateEmptyObject(rootPath *string, db *gorm.DB) (*Object, error) {
	var (
		h                = sha256.New()
		err              error
		chunk            *Chunk
		object           *Object
		hashState        string
		emptyContentHash = hex.EncodeToString(h.Sum(nil))
	)

	if object, err = FindObjectByHash(emptyContentHash, db); err == nil && object != nil {
		return object, nil
	}

	if chunk, err = CreateEmptyContentChunk(rootPath, db); err != nil {
		return nil, err
	}

	if hashState, err = sha2562.GetHashStateText(h); err != nil {
		return nil, err
	}

	object = &Object{
		Size: 0,
		Hash: emptyContentHash,
		ObjectChunks: []ObjectChunk{
			{
				ChunkID:   chunk.ID,
				Number:    1,
				HashState: &hashState,
			},
		},
	}

	return object, db.Set("gorm:association_autocreate", true).Save(object).Error
}
