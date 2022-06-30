package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"medea/pkg/database/models"
	"medea/pkg/http"
	"mime/multipart"
	libHttp "net/http"
	"os"
	"strconv"
	"time"
)

type fileMetaConfig struct {
	token  string
	secret string
	uid    string
	host   string
	dst    string
}

func file_create(val map[string]string) error {
	token := val["token"]
	secret := val["secret"]
	path := val["path"]
	src := val["src"]
	host := val["host"]

	file, err := os.Open(src)
	if err != nil {
		return err
	}

	count := 0
	for index := 0; ; index++ {
		var (
			err            error
			body           = new(bytes.Buffer)
			chunk          = make([]byte, models.ChunkSize)
			request        *libHttp.Request
			readCount      int
			formBodyWriter = multipart.NewWriter(body)
			formFileWriter io.Writer
		)

		if readCount, err = file.Read(chunk); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		params := map[string]interface{}{
			"token": token,
			"path":  path,
			"nonce": models.RandomWithMD5(255),
		}

		if index == 0 {
			params["overwrite"] = "1"
		} else {
			params["append"] = "1"
		}

		params["sign"] = http.GetParamsSignature(params, secret)
		for k, v := range params {
			if err = formBodyWriter.WriteField(k, v.(string)); err != nil {
				return err
			}
		}

		if formFileWriter, err = formBodyWriter.CreateFormFile("file", "random.bytes"); err != nil {
			return err
		}

		if _, err = formFileWriter.Write(chunk[:readCount]); err != nil {
			return err
		}

		if err = formBodyWriter.Close(); err != nil {
			return err
		}

		api := fmt.Sprintf("%s/%s", medeaServer, "api/medea/file/create")
		if request, err = libHttp.NewRequest(libHttp.MethodPost, api, body); err != nil {
			return err
		}

		request.Header.Set("Content-Type", formBodyWriter.FormDataContentType())
		request.Header.Set("X-Forwarded-For", host)
		resp, err := libHttp.DefaultClient.Do(request)
		if err != nil {
			return err
		}

		if bodyBytes, err := ioutil.ReadAll(resp.Body); err != nil {
			return err
		} else {
			p := &http.Response{}
			err := json.Unmarshal(bodyBytes, p)
			if err != nil {
				return err
			}

			resp2, err := json.MarshalIndent(p, "", "    ")
			if err != nil {
				return err
			}

			fmt.Println("resp:\n", string(resp2))
		}

		count++
	}

	fmt.Printf("finished %d partitions upload\n", count)

	return nil
}

func file_info(val map[string]string) error {
	token := val["token"]
	secret := val["secret"]
	uid := val["uid"]
	host := val["host"]

	qs := http.GetParamsSignBody(map[string]interface{}{
		"token":   token,
		"fileUid": uid,
		"nonce":   RandomWithMD56(333),
	}, secret)

	api := fmt.Sprintf("%s/%s", medeaServer, "api/medea/file/info")

	request, err := libHttp.NewRequest("GET", fmt.Sprintf("%s?%s", api, qs), nil)
	if err != nil {
		return err
	}
	request.Header.Set("X-Forwarded-For", host)
	resp, err := libHttp.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	}

	for k, v := range resp.Header {
		fmt.Println(k, ":", v)
	}

	return nil
}

func file_read(val map[string]string) error {
	token := val["token"]
	secret := val["secret"]
	uid := val["uid"]
	dst := val["dst"]
	host := val["host"]

	var concurrence int
	var err error
	concurrence, err = strconv.Atoi(val["range"])
	if err != nil {
		concurrence = 0
	}

	meta := &fileMetaConfig{
		token:  token,
		secret: secret,
		uid:    uid,
		host:   host,
	}

	qs := http.GetParamsSignBody(map[string]interface{}{
		"token":   token,
		"fileUid": uid,
		"nonce":   RandomWithMD56(333),
	}, secret)

	api := fmt.Sprintf("%s/%s", medeaServer, "api/medea/file/read")
	url := fmt.Sprintf("%s?%s", api, qs)

	if concurrence >= 2 {
		fmt.Println("start range download")
		downloader := NewRangeDownloader(meta, url, dst, "", concurrence, "")
		if err := downloader.Run(); err != nil {
			return err
		}

		return nil
	} else {
		fmt.Println("start plain download")
		start := time.Now()
		request, err := libHttp.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		request.Header.Set("X-Forwarded-For", host)
		resp, err := libHttp.DefaultClient.Do(request)
		if err != nil {
			return err
		}

		out, err := os.Create(dst)
		if err != nil {
			panic(err)
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			panic(err)
		}

		fi, err := os.Stat(dst)
		if err == nil {
			elapsed := time.Now().Sub(start)
			ms := int64(elapsed / time.Millisecond)
			speed := fi.Size() * 1000 / ms
			timeStr := time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("%v (%v) - file '%v' saved [%v]\n", timeStr, speedTransfer(speed), dst, fi.Size())
		} else {
			fmt.Printf("file saved to %v failed, %v\n", dst, err)
		}

		return nil
	}
}

func file_list(val map[string]string) error {
	token := val["token"]
	secret := val["secret"]
	host := val["host"]

	api := fmt.Sprintf("%s/%s", medeaServer, "api/medea/directory/list")

	qs := http.GetParamsSignBody(map[string]interface{}{
		"token": token,
		"nonce": RandomWithMD56(333),
	}, secret)

	request, err := libHttp.NewRequest("GET", fmt.Sprintf("%s?%s", api, qs), nil)
	if err != nil {
		return err
	}
	request.Header.Set("X-Forwarded-For", host)
	resp, err := libHttp.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if bodyBytes, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	} else {
		p := &http.Response{}
		err := json.Unmarshal(bodyBytes, p)
		if err != nil {
			return err
		}

		resp2, err := json.MarshalIndent(p, "", "    ")
		if err != nil {
			return err
		}

		fmt.Println("resp:\n", string(resp2))
	}

	return nil
}
