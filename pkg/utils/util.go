package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"reflect"

	"github.com/jinzhu/gorm"
)

func IsDir(path string) bool {
	var (
		fileInfo os.FileInfo
		err      error
	)
	fileInfo, err = os.Stat(path)

	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func IsFile(path string) bool {
	var (
		fileInfo os.FileInfo
		err      error
	)
	fileInfo, err = os.Stat(path)

	if err != nil {
		return false
	}
	return !fileInfo.IsDir()
}

func SubStrFromTo(s string, from, to int) string {
	if from < 0 {
		from = len(s) + from
	}
	if to < 0 {
		to = len(s) + to
	}
	return s[from:to]
}

func SubStrFromToEnd(s string, from int) string {
	return SubStrFromTo(s, from, len(s))
}

func ReverseSlice(data interface{}) {
	value := reflect.ValueOf(data)
	if value.Kind() != reflect.Slice {
		panic(errors.New("data must be a slice type"))
	}
	valueLen := value.Len()
	swap := reflect.Swapper(data)
	for i := 0; i <= int((valueLen-1)/2); i++ {
		reverseIndex := valueLen - 1 - i
		swap(i, reverseIndex)
	}
}

func Sha256Hash2String(p []byte) (string, error) {
	hash := sha256.New()
	if _, err := hash.Write(p); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func IsRecordNotFound(err error) bool {
	return err != nil && err.Error() == "record not found"
}

func InTransaction(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	return reflect.ValueOf(db).Elem().FieldByName("db").Elem().Type().String() == "*sql.Tx"
}
