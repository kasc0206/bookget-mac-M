package app

import (
	"bookget/config"
	"bookget/pkg/chttp"
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// NewHttpClient 创建基于配置的 HTTP 客户端（统一 TLS 设置）
func NewHttpClient() (*http.Client, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.Conf.SkipVerify,
		},
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Timeout:   config.Conf.Timeout * time.Second,
		Jar:       jar,
		Transport: tr,
	}, nil
}

func BuildRequestHeader() map[string]string {
	httpHeaders := map[string]string{"User-Agent": config.Conf.UserAgent}
	cookies, _ := chttp.ReadCookiesFromFile(config.Conf.CookieFile)
	if cookies != "" {
		httpHeaders["Cookie"] = cookies
	}

	headers, err := chttp.ReadHeadersFromFile(config.Conf.HeaderFile)
	if err == nil {
		for key, value := range headers {
			httpHeaders[key] = value
		}
	}
	return httpHeaders
}
