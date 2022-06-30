package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"medea/pkg/http"
	libHttp "net/http"
	"strings"
	"time"
)

func token_create(val map[string]string) error {
	appUid := val["uid"]
	appSecret := val["secret"]
	host := val["host"]

	api := fmt.Sprintf("%s/%s", medeaServer, "api/medea/token/create")

	nonce := RandomWithMD56(128)
	body := http.GetParamsSignBody(map[string]interface{}{
		"appUid":         appUid,
		"availableTimes": -1,
		"expiredAt":      time.Now().AddDate(0, 0, 2).Unix(),
		"ip":             host,
		"nonce":          nonce,
		"path":           "/",
		"readOnly":       false,
		"secret":         RandomWithMD56(44),
	}, appSecret)

	request, err := libHttp.NewRequest("POST", api, strings.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

		return nil
	}
}
