package client

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"medea/pkg/http"
	"mime"
	libHttp "net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type Partition struct {
	Index int
	From  int
	To    int
	Data  []byte
}

type RangeDownloader struct {
	m              *fileMetaConfig
	fileSize       int
	url            string
	outputFileName string
	totalPart      int
	outputDir      string
	doneFilePart   []Partition
	md5            string
}

func NewRangeDownloader(m *fileMetaConfig, url, outputFileName, outputDir string, totalPart int, md5 string) *RangeDownloader {
	if outputDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}
		outputDir = wd
	}

	return &RangeDownloader{
		m:              m,
		fileSize:       0,
		url:            url,
		outputFileName: outputFileName,
		outputDir:      outputDir,
		totalPart:      totalPart,
		doneFilePart:   make([]Partition, totalPart),
		md5:            md5,
	}
}

func (r *RangeDownloader) Run() error {

	fileTotalSize, err := r.getHeaderInfo(r.m)
	if err != nil {
		return err
	}
	r.fileSize = fileTotalSize

	jobs := make([]Partition, r.totalPart)
	eachSize := fileTotalSize / r.totalPart

	for i := range jobs {
		jobs[i].Index = i
		if i == 0 {
			jobs[i].From = 0
		} else {
			jobs[i].From = jobs[i-1].To + 1
		}
		if i < r.totalPart-1 {
			jobs[i].To = jobs[i].From + eachSize - 1
		} else {
			jobs[i].To = fileTotalSize - 1
		}
	}

	start := time.Now()

	var wg sync.WaitGroup
	for _, j := range jobs {
		wg.Add(1)
		go func(job Partition) {
			defer wg.Done()
			err := r.downloadPart(job)
			if err != nil {
				log.Println("Download file failed:", err, job)
			}
		}(j)
	}
	wg.Wait()

	elapsed := time.Now().Sub(start)
	ms := int64(elapsed / time.Millisecond)
	speed := int64(fileTotalSize) * 1000 / ms
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%v (%v) - file '%v' saved [%v]\n", timeStr, speedTransfer(speed), r.outputFileName, fileTotalSize)

	return r.mergeFileParts()
}

func getNewRequest(url, method string, headers map[string]string) (*libHttp.Request, error) {
	r, err := libHttp.NewRequest(
		method,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}

	return r, nil
}

func (r *RangeDownloader) downloadPart(c Partition) error {
	headers := map[string]string{
		"Range": fmt.Sprintf("bytes=%v-%v", c.From, c.To),
	}
	request, err := getNewRequest(r.url, "GET", headers)
	if err != nil {
		return err
	}

	request.Header.Set("X-Forwarded-For", r.m.host)

	log.Printf("Start [%d] download from:%d to:%d\n", c.Index, c.From, c.To)
	resp, err := libHttp.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Server error status code: %v", resp.StatusCode))
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(bs) != (c.To - c.From + 1) {
		return errors.New(fmt.Sprintf("File partition length error %v", len(bs)))
	}

	c.Data = bs
	r.doneFilePart[c.Index] = c

	return nil
}

func (r *RangeDownloader) mergeFileParts() error {
	path := r.outputFileName

	log.Println("Start to merge files partitions")
	mergedFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer mergedFile.Close()

	fileMd5 := sha256.New()
	totalSize := 0
	for _, s := range r.doneFilePart {
		_, err := mergedFile.Write(s.Data)
		if err != nil {
			fmt.Printf("error when merge file: %v\n", err)
		}
		fileMd5.Write(s.Data)
		totalSize += len(s.Data)
	}
	if totalSize != r.fileSize {
		return errors.New("File incomplete")
	}

	if r.md5 != "" {
		if hex.EncodeToString(fileMd5.Sum(nil)) != r.md5 {
			return errors.New("File corrupted")
		} else {
			log.Println("SHA-256 check succeeded")
		}
	}

	return nil
}

func (r *RangeDownloader) getHeaderInfo(meta *fileMetaConfig) (int, error) {
	token := meta.token
	secret := meta.secret
	uid := meta.uid
	host := meta.host

	qs := http.GetParamsSignBody(map[string]interface{}{
		"token":   token,
		"fileUid": uid,
		"nonce":   RandomWithMD56(333),
	}, secret)

	api := fmt.Sprintf("%s/%s", medeaServer, "api/medea/file/info")

	request, err := libHttp.NewRequest("GET", fmt.Sprintf("%s?%s", api, qs), nil)
	if err != nil {
		return -1, err
	}
	request.Header.Set("X-Forwarded-For", host)
	request.Header.Set("Range", "bytes")
	resp, err := libHttp.DefaultClient.Do(request)
	if err != nil {
		return -1, err
	}

	if resp.Header.Get("Accept-Ranges") != "bytes" {
		return 0, errors.New("服务器不支持文件断点续传")
	}

	outputFileName, err := parseFileInfo(resp)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("get file info err: %v", err))
	}
	if r.outputFileName == "" {
		r.outputFileName = outputFileName
	}

	return strconv.Atoi(resp.Header.Get("Content-Length"))
}

func parseFileInfo(resp *libHttp.Response) (string, error) {
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err != nil {
			return "", err
		}
		return params["filename"], nil
	}

	filename := filepath.Base(resp.Request.URL.Path)
	return filename, nil
}
