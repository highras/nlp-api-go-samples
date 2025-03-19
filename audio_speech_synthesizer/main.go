package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	endpointHost = "tts.ilivedata.com"
	endpointPath = "/api/v1/speech/synthesis"
	endpointURL  = "https://tts.ilivedata.com/api/v1/speech/synthesis"

	projectID = "81900001"
	secretKey = "iV7DP7O7zLh8OicMxnV5pkkGFNTKNxkIr6Y5AKWNRcg="
)

func main() {
	text := "云上曲率的语音合成/克隆服务利用大模型，深度融合文本理解和语音生成，精准解析并诠释各类文本内容，转化为宛如真人般的自然语音。"

	speechSynthesis(text, "ZhangWei")

	voiceUrl := "https://vtai-off-sts-ap-1306922583.cos.accelerate.myqcloud.com//timbre_library/common/费尔南多·雷伊/vocal_fc53a7f4-a30a-468f-944c-89a5925e18e8.wav"
	voiceClone(text, voiceUrl)
}

func nowTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

func speechSynthesis(text, voice string) {
	body := map[string]interface{}{
		"text": text,
		"voice": map[string]string{
			"name": voice,
		},
	}

	start := time.Now()
	resp, err := request(context.TODO(), body, nowTimestamp())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("speechSynthesis end, cost:%s, resp:%s", time.Since(start), resp)
}

func voiceClone(text, audioUrl string) {
	body := map[string]interface{}{
		"text": text,
		"voice": map[string]string{
			"audio": audioUrl,
		},
	}

	start := time.Now()
	resp, err := request(context.TODO(), body, nowTimestamp())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("voiceClone end, cost:%s, resp:%s", time.Since(start), resp)
}

var httpClient = &http.Client{
	Timeout: time.Second * 30,
}

func request(ctx context.Context, body map[string]interface{}, timestamp string) (*bytes.Buffer, error) {
	// 请求体json序列化
	buf := getBuffer()
	defer putBuffer(buf)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, fmt.Errorf("json encode failed: %v", err)
	}

	// 生成请求签名
	signBuf, err := genSign(buf.Bytes(), timestamp)
	defer putBuffer(signBuf)
	if err != nil {
		return nil, fmt.Errorf("generate sign failed: %v", err)
	}

	// 构造请求
	req, err := http.NewRequestWithContext(ctx, "POST", endpointURL, buf)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}
	req.Header.Set("X-AppId", projectID)
	req.Header.Set("X-TimeStamp", timestamp)
	req.Header.Set("Authorization", signBuf.String())
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Println("close response body failed", closeErr)
		}
	}()

	// 读取响应数据
	respBuf := new(bytes.Buffer)
	if _, err = io.Copy(respBuf, resp.Body); err != nil {
		return nil, fmt.Errorf("read response body failed: %v", err)
	}
	return respBuf, nil
}

func genSign(body []byte, timestamp string) (*bytes.Buffer, error) {
	signBuf, err := buildSign(body, timestamp)
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha256.New, []byte(secretKey))
	if _, err = io.Copy(h, signBuf); err != nil {
		return nil, fmt.Errorf("encode HmacSHA256 fail: %w", err)
	}

	baseBuf := signBuf // buffer has been reset
	enc := base64.NewEncoder(base64.StdEncoding, baseBuf)
	defer func() {
		if closeErr := enc.Close(); closeErr != nil {
			log.Println("close base64 encoder failed", closeErr)
		}
	}()
	if _, err = enc.Write(h.Sum(nil)); err != nil {
		return nil, fmt.Errorf("base64 encode fail: %w", err)
	}

	return baseBuf, nil
}

func buildSign(body []byte, timestamp string) (*bytes.Buffer, error) {
	buf := getBuffer()
	buf.WriteString("POST\n")
	buf.WriteString(endpointHost)
	buf.WriteByte('\n')
	buf.WriteString(endpointPath)
	buf.WriteByte('\n')
	if err := sha256AndHex(body, buf); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	buf.WriteString("X-AppId:")
	buf.WriteString(projectID)
	buf.WriteByte('\n')
	buf.WriteString("X-TimeStamp:")
	buf.WriteString(timestamp)
	return buf, nil
}

func sha256AndHex(data []byte, w io.Writer) error {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		return fmt.Errorf("encode sha256 fail: %w", err)
	}
	if _, err := hex.NewEncoder(w).Write(h.Sum(nil)); err != nil {
		return fmt.Errorf("encode hex fail: %w", err)
	}
	return nil
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 64))
	},
}

func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}
