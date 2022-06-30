package models

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	mrand "math/rand"
	"strconv"
	"time"
)

const SecretLength = 12

func NewSecret() string {
	r := "1234567890abcdefghijklmnopqrstuvwxyzQWERTYUIOPASDFGHJKLZXCVBNM"
	randomBytes := make([]byte, SecretLength)
	rLength := len(r)
	mrand.Seed(time.Now().Unix())
	for i := 0; i < SecretLength; i++ {
		randomBytes[i] = r[mrand.Intn(rLength)]
	}
	return string(randomBytes)
}

func Random(length uint) []byte {
	var r = make([]byte, length)
	_, _ = rand.Reader.Read(r)
	return r
}

func RandomWithMD5(length uint) string {
	var (
		b    = Random(length)
		hash = md5.New()
	)
	_, _ = hash.Write(b)
	return hex.EncodeToString(hash.Sum(nil))
}

func UID() string {
	random := Random(32)
	random = append(random, []byte(strconv.FormatInt(time.Now().UnixNano(), 10))...)
	hash := md5.New()
	_, _ = hash.Write(random)
	return hex.EncodeToString(hash.Sum(nil))
}
