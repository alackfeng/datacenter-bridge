package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// HttpClient -
type HttpClient struct {
	*http.Client
}

// NewHttpClient -
func NewHttpClient() *HttpClient {
	return NewHttpClientWithTimeout(30 * time.Second)
}

// NewHttpClientWithTimeout -
func NewHttpClientWithTimeout(timeout time.Duration) *HttpClient {
	return &HttpClient{
		Client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

// Get - url.
func (c *HttpClient) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if _, ok := headers["Content-Type"]; !ok {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Post - json.
func (c *HttpClient) Post(ctx context.Context, url string, body interface{}, headers map[string]string) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	if _, ok := headers["Content-Type"]; !ok {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

const baseUrl = "http://10.16.3.149:9898"

var ipcId string
var username string
var password string
var appId int

// go run -v examples/app/mock_http_app.go
// go build -o mock_http_app examples/app/mock_http_app.go
// CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o mock_http_app.exe examples/app/mock_http_app.go

func main() {
	flag.IntVar(&appId, "appId", 2, "app id")
	flag.StringVar(&username, "username", "yexuhui@houwei.com", "app username")
	flag.StringVar(&password, "password", "aa123456", "app password")
	flag.StringVar(&ipcId, "ipcId", "test500000", "ipc device id")
	flag.Parse()

	client := NewHttpClient()
	req := struct {
		AppId    int    `json:"app_id"`
		Account  string `json:"account"`
		Password string `json:"password"`
	}{
		AppId:    appId,
		Account:  username,
		Password: password,
	}
	url3 := fmt.Sprintf("%s%s", baseUrl, "/api/v1/user/login")
	res, err := client.Post(context.Background(), url3, &req, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("res: ", string(res))

	var connResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Token   string `json:"token"`
			WsToken string `json:"ws_token"`
			WsFrom  string `json:"ws_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(res, &connResp); err != nil {
		panic(err)
	}
	value := url.Values{}
	value.Add("App", connResp.Data.WsToken)
	fmt.Println(">>>ws conn token base64: ", string(connResp.Data.WsToken))
	fmt.Println(">>>ws conn token base64 url encode: ", string(value.Encode()))
	fmt.Println(">>>ws conn token base64 url encode: ", url.QueryEscape(connResp.Data.WsToken))
	fmt.Println(">>>ws conn token device id: ", string(connResp.Data.WsFrom))

	req2 := struct {
		IpcId string `json:"ipc_id"`
	}{
		IpcId: ipcId,
	}
	url2 := fmt.Sprintf("%s%s", baseUrl, "/api/v1/userDevice/getDeviceToken")
	res2, err := client.Post(context.Background(), url2, &req2, map[string]string{
		"Token": connResp.Data.Token,
	})
	if err != nil {
		panic(err)
	}
	var loginResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(res2, &loginResp); err != nil {
		panic(err)
	}
	fmt.Println(">>>ws login token base64: ", string(loginResp.Data.Token))
	fmt.Println(">>>ws login token base64 url encode: ", url.QueryEscape(loginResp.Data.Token))

}
