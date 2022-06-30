package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type AutoRotateWriter struct {
	mu                      sync.Mutex
	maxBytes                uint64
	dir                     string
	number                  uint32
	basename                string
	ext                     string
	handler                 *os.File
	handlerAlreadyWriteSize uint64
}

func NewAutoRotateWriter(file string, maxBytes uint64) (*AutoRotateWriter, error) {
	var (
		dir                     = filepath.Dir(file)
		err                     error
		stat                    os.FileInfo
		number                  uint32
		handlerAlreadyWriteSize uint64
		completeFileName        string
		basename                string
		ext                     string
		handler                 *os.File
	)
	stat, err = os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
	} else if err == nil && !stat.IsDir() {
		return nil, errors.New("the directory of log file is illegal")
	} else if err != nil {
		return nil, err
	}

	dir = strings.TrimSuffix(dir, "/")
	ext = filepath.Ext(file)
	basename = strings.TrimSuffix(filepath.Base(file), ext)
	ext = strings.TrimPrefix(ext, ".")

	for {
		completeFileName = fmt.Sprintf("%s/%s.%d.%s", dir, basename, number, ext)
		if stat, err = os.Stat(completeFileName); err != nil {
			if os.IsNotExist(err) {
				if handler, err = os.OpenFile(
					completeFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
					return nil, err
				}
				break
			} else {
				return nil, err
			}
		} else {
			handlerAlreadyWriteSize = uint64(stat.Size())
			if handlerAlreadyWriteSize < maxBytes {
				handler, err = os.OpenFile(completeFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
				if err != nil {
					return nil, err
				}
				break
			}
		}
		number++
	}

	return &AutoRotateWriter{
		maxBytes:                maxBytes,
		dir:                     dir,
		number:                  number,
		basename:                basename,
		ext:                     ext,
		handler:                 handler,
		handlerAlreadyWriteSize: handlerAlreadyWriteSize,
	}, nil
}

func (a *AutoRotateWriter) Write(p []byte) (n int, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	n, err = a.handler.Write(p)
	atomic.AddUint64(&a.handlerAlreadyWriteSize, uint64(n))
	if atomic.LoadUint64(&a.handlerAlreadyWriteSize) >= a.maxBytes {
		_ = a.handler.Close()
		atomic.AddUint32(&a.number, 1)
		nextFileName := fmt.Sprintf("%s/%s.%d.%s", a.dir, a.basename, a.number, a.ext)
		a.handler, err = os.OpenFile(nextFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return n, err
		}
		a.handlerAlreadyWriteSize = 0
	}
	return n, err
}

func (a *AutoRotateWriter) Close() error {
	return a.handler.Close()
}
