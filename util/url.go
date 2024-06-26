package util

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
)

// 根据URL返回带端口的域名
func HandleHost(tmpUrl string) (host string) {
	tmpUrl2, err := url.Parse(tmpUrl)
	if err != nil {
		return
	}
	if tmpUrl2.Host == "" {
		return
	}
	return fmt.Sprintf("%s://%s", tmpUrl2.Scheme, tmpUrl2.Host)
}

func HandleHostname(tmpUrl string) (host string) {
	tmpUrl2, err := url.Parse(tmpUrl)
	if err != nil {
		return
	}
	return tmpUrl2.Hostname()
}

// 是否是http协议的路径
func IsHttpUrl(tmpUrl string) bool {
	return strings.HasPrefix(tmpUrl, "http://") || strings.HasPrefix(tmpUrl, "https://")
}

// 获取重定向内容
func HandleRedirectUrl(requestUrl string) (redirectUrl string) {
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirectUrl = req.URL.String()
			return nil
		},
	}
	_, err := httpClient.Head(requestUrl)
	if err != nil {
		log.Println("[http.Client.Do.Error]", err)
		return requestUrl
	}

	return redirectUrl
}

// 把url转为请求 /api/video/cors 接口的形式，方便后续获取重定向内容
func HandleUrlToCORS(tmpUrl string) string {
	return fmt.Sprintf(
		"%s/api/video/cors?src=%s",
		strings.TrimRight(ApiConfig.Server, "/"),
		url.QueryEscape(tmpUrl),
	)
}

// 合并URL
func ChangeUrlPath(tmpUrl, tmpPath string) string {
	if tmpPath == "" {
		return tmpUrl
	}
	// 如果是 / 开头，直接域名+路径
	if strings.HasPrefix(tmpPath, "/") {
		return fmt.Sprintf("%s/%s", HandleHost(tmpUrl), strings.TrimLeft(tmpPath, "/"))
	}
	parsedUrl, err := url.Parse(tmpUrl)
	if err != nil {
		return tmpPath
	}
	// 防止空域导致地址不对
	if parsedUrl.Host == "" {
		return tmpPath
	}
	return fmt.Sprintf(
		"%s://%s/%s/%s",
		parsedUrl.Scheme,
		parsedUrl.Host,
		strings.TrimLeft(path.Dir(parsedUrl.Path), "/"),
		tmpPath,
	)
}

func CheckVideoUrl(url string) bool {

	log.Println("[checkUrl]", url)
	var httpW = HttpWrapper{}
	headers, err := httpW.Head(url)
	if err != nil {
		log.Println("[CheckVideoUrl.Error]", err.Error())
		return false
	}

	defer func() {
		log.Println("[CheckVideoUrl.headers]", ToJSON(headers, false))
	}()

	v, ok := headers["Content-Type"]
	if ok {
		for _, s := range v {
			// 有可能返回m3u8文件还是Accept-Ranges=bytes的情况
			if s == "application/vnd.apple.mpegurl" {
				return false
			}
		}
		for _, s := range v {
			if s == "video/mp4" {
				return true
			}
		}
	}

	// 检测文件大小，太大了也认为是视频文件
	v, ok = headers["Content-Length"]
	if ok {
		for _, s := range v {
			u, _ := strconv.ParseUint(s, 10, 64)
			// 5*1024*1024 5MB
			if u > 5*1024*1024 {
				return true
			}
		}
	}

	// 如果返回的数据是支持范围请求，则说明可能是个大文件
	v, ok = headers["Accept-Ranges"]
	if ok && slices.Contains(v, "bytes") {
		return true
	}

	return false
}

func FillUrlHost(tmpUrl string, host string) string {
	if IsHttpUrl(tmpUrl) {
		return tmpUrl
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(host, "/"), strings.TrimLeft(tmpUrl, "/"))
}
