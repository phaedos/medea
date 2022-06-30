package client

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func Random56(length uint) []byte {
	var r = make([]byte, length)
	_, _ = rand.Reader.Read(r)
	return r
}

func RandomWithMD56(length uint) string {
	var (
		b    = Random56(length)
		hash = md5.New()
	)
	_, _ = hash.Write(b)
	return hex.EncodeToString(hash.Sum(nil))
}

func speedTransfer(src int64) string {
	unit := []string{"B/s", "KB/s", "MB/s", "GB/s", "TB/s"}
	i := 0
	now := src
	for {
		if now < 1000 {
			return fmt.Sprintf("%v %s", now, unit[i])
		}

		if (i + 1) >= len(unit) {
			return fmt.Sprintf("%v %s", now, unit[i])
		} else {
			i += 1
			now /= 1000
		}
	}
}
