package http

import (
	"fmt"
	"sort"
	"strings"

	janitor "github.com/json-iterator/go"
)

var GetParamsSignBody = getParamsSignBody

var GetParamsSignature = getParamsSignature

func assertTokenRespStructure(data interface{}) bool {
	keys := []string{"availableTimes", "token", "ip", "readOnly", "expiredAt", "path", "secret"}
	mData := data.(map[string]interface{})
	for _, k := range keys {
		if _, ok := mData[k]; !ok {
			return false
		}
	}
	return true
}

func parseResponse(body string) (*Response, error) {
	var response = &Response{}
	json := janitor.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal([]byte(body), response)
	return response, err
}

func getParamsSignBody(p map[string]interface{}, secret string) string {
	sign, buf := getParamsSignatureWithBuf(p, secret)
	buf.WriteString(fmt.Sprintf("&sign=%s", sign))
	return buf.String()
}

func getParamsSignatureWithBuf(p map[string]interface{}, secret string) (string, *strings.Builder) {
	var (
		buf   = new(strings.Builder)
		index = 0
		keys  = make([]string, len(p))
	)
	for k := range p {
		keys[index] = k
		index++
	}
	sort.Strings(keys)
	for i, k := range keys {
		buf.WriteString(k)
		buf.WriteRune('=')
		buf.WriteString(fmt.Sprintf("%v", p[k]))
		if i != len(keys)-1 {
			buf.WriteRune('&')
		}
	}
	return SignStrWithSecret(buf.String(), secret), buf
}

func getParamsSignature(p map[string]interface{}, secret string) string {
	sign, _ := getParamsSignatureWithBuf(p, secret)
	return sign
}
