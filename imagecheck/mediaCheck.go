package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"strconv"
)

const (
	companyID            = "123"
	secretKey            = "xxxxxxxxxxxxxx"
	endpointURL          = "https://media-safe.ilivedata.com/service/check"
	textAppID            = "80700001"
	imageAppID            = "81000001"
)

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func check(text string, imageList []string, userID string) string {

	var timestamp = time.Now().Unix()
	var hash = GetMD5Hash(companyID + ":" + strconv.Itoa(int(timestamp)) + ":" + secretKey)
	var signature = strings.ToUpper(hash)

	// Prepare parameters
	var parameters = map[string]interface{}{
		"companyID":   companyID,
		"timestamp":  timestamp,
		"signature": signature,
		"textAppID": textAppID,
		"imageAppID": imageAppID,
		"userID": userID,
		"text": text,
		"imageList": imageList,
	}
	var queryBody []byte
	queryBody, err := json.Marshal(parameters)
	if err != nil {
		// log something
		return err.Error()
	}
	//fmt.Println(string(queryBody))

	// Send request
	return request(string(queryBody))
}

func request(body string) string {
	var httpClient = http.Client{}
	request, _ := http.NewRequest("POST", endpointURL, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response, err := httpClient.Do(request)

	if err != nil {
		// log something
		return err.Error()
	}
	//fmt.Println(response.StatusCode)
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// log something
		return err.Error()
	}
	//fmt.Println(string(content))
	return string(content)
}

func main() {
	var imageList = [...]string {"https://test.com/1.jpg", "https://test/2.jpg"}
	var responseJson = check("fuck test", imageList[:], "12345678")
	fmt.Println(responseJson)
}
