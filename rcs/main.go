package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	projectID             = "YOUR_PROJECT_ID_GOES_HERE"
	secrectKey            = "YOUR_SECRET_KEY_GOES_HERE"
	endpointURL           = "https://profanity.ilivedata.com/api/v2/profanity"
	endpointHost          = "profanity.ilivedata.com"
	endpointPath          = "/api/v2/profanity"
	iso8601DateFormatNoMS = "2006-01-02T15:04:05Z"
)

// Profanity Invoke profanity v2 API
func Profanity(sentence string, classify int, userID string, userName string) string {
	// UTC Time
	var now = time.Now().UTC().Format(iso8601DateFormatNoMS)
	// Prepare parameters
	var params = map[string]string{
		"q":         sentence,
		"classify":  strconv.Itoa(classify),
		"userId":    userID,
		"userName":  userName,
		"timeStamp": now,
		"appId":     projectID,
	}

	// Compute signature
	var signature = signAndBase64Encode(stringToSign(params), secrectKey)
	fmt.Println(signature)
	// Send request
	return request(params, signature)
}

func signAndBase64Encode(data string, secrectKey string) string {
	var mac = hmac.New(sha256.New, []byte(secrectKey))
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func stringToSign(data map[string]string) string {
	queryKeys := make([]string, 0, len(data))
	for key := range data {
		queryKeys = append(queryKeys, key)
	}
	sort.Strings(queryKeys)

	query := make([]string, 0, len(data))
	for _, key := range queryKeys {
		k := strings.Replace(url.QueryEscape(key), "+", "%20", -1)
		v := strings.Replace(url.QueryEscape(data[key]), "+", "%20", -1)
		query = append(query, k+"="+v)
	}

	var stringToSign = []string{
		"POST",
		endpointHost,
		endpointPath,
		strings.Join(query, "&"),
	}
	fmt.Println(strings.Join(stringToSign, "\n"))
	return strings.Join(stringToSign, "\n")
}

func request(params map[string]string, signature string) string {
	var body = url.Values{}
	for k, v := range params {
		body.Add(k, v)
	}
	var httpClient = http.Client{}
	request, _ := http.NewRequest("POST", endpointURL, strings.NewReader(body.Encode()))
	request.Header.Set("Authorization", signature)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset='utf-8'")
	request.Header.Set("User-Agent", "Golang_HTTP_Client/1.0")

	response, err := httpClient.Do(request)

	if err != nil {
		// log something
		return err.Error()
	}
	fmt.Println(response.StatusCode)
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// log something
		return err.Error()
	}
	fmt.Println(string(content))
	return string(content)
}

func main() {
	Profanity("加抠13812123434", 1, "12345678", "张三")
	Profanity("武汉肺炎", 0, "12345678", "张三")
}
