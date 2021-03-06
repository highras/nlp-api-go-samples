package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	projectID             = "YOUR_PROJECT_ID_GOES_HERE"
	secrectKey            = "YOUR_SECRET_KEY_GOES_HERE"
	endpointURL           = "https://asr.ilivedata.com/api/v1/speech/recognize/submit"
	endpointHost          = "asr.ilivedata.com"
	endpointPath          = "/api/v1/speech/recognize/submit"
	iso8601DateFormatNoMS = "2006-01-02T15:04:05Z"
)

// Recognize Invoke speech recognize v1 API
func Recognize(audioURL string, languageCode string, userID string, speakerDiarization bool) string {
	// UTC Time
	var now = time.Now().UTC().Format(iso8601DateFormatNoMS)
	// Prepare parameters
	var parameters = map[string]interface{}{
		"languageCode": languageCode,
		"config": map[string]interface{}{
			"codec":           "PCM",
			"sampleRateHertz": 16000,
		},
		"diarizationConfig": map[string]interface{}{
			"enableSpeakerDiarization": speakerDiarization,
		},
		"uri":    audioURL,
		"userId": userID,
	}
	var queryBody []byte
	queryBody, err := json.Marshal(parameters)
	if err != nil {
		// log something
		return err.Error()
	}
	//fmt.Println(string(queryBody))

	var preparedString = []string{
		"POST",
		endpointHost,
		endpointPath,
		sha256AndHexEncode(string(queryBody)),
		"X-AppId:" + projectID,
		"X-TimeStamp:" + now,
	}
	var stringToSign = strings.Join(preparedString, "\n")
	fmt.Println(stringToSign)
	// Compute signature
	var signature = signAndBase64Encode(stringToSign, secrectKey)
	fmt.Println(signature)
	// Send request
	return request(string(queryBody), signature, now)
}

func signAndBase64Encode(data string, secrectKey string) string {
	var mac = hmac.New(sha256.New, []byte(secrectKey))
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func sha256AndHexEncode(data string) string {
	var sha256Hash = sha256.New()
	sha256Hash.Write([]byte(data))
	return hex.EncodeToString(sha256Hash.Sum(nil))
}

func request(body string, signature string, timeStamp string) string {
	var httpClient = http.Client{}
	request, _ := http.NewRequest("POST", endpointURL, strings.NewReader(body))
	request.Header.Set("X-AppId", projectID)
	request.Header.Set("X-TimeStamp", timeStamp)
	request.Header.Set("Authorization", signature)
	request.Header.Set("Content-Type", "application/json")
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
	var audioURL = "https://rcs-us-west-2.s3.us-west-2.amazonaws.com/test.wav"
	Recognize(audioURL, "zh-CN", "12345678", true)
}
