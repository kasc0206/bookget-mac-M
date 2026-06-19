package app

import (
	"bookget/config"
	"bookget/pkg/gohttp"
	xhash "bookget/pkg/hash"
	"bookget/pkg/util"
	"bytes"
	"context"
	"fmt"
	"html"
	"io"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
)

type DownloadTask struct {
	Index     int
	Url       string
	UrlParsed *url.URL
	SavePath  string
	BookId    string
	Title     string
	VolumeId  string
	Param     map[string]interface{} //备用参数
	Jar       *cookiejar.Jar
}

type Volume struct {
	Title string
	Url   string
	Seq   int
}
type PartialVolumes struct {
	directory string
	Title     string
	volumes   []string
}

type PartialCanvases struct {
	directory string
	Title     string
	Canvases  []string
}

func getBookId(sUrl string) (bookId string) {
	if sUrl == "" {
		return ""
	}
	mh := xhash.NewMultiHasher()
	_, _ = io.Copy(mh, bytes.NewBuffer([]byte(sUrl)))
	bookId, _ = mh.SumString(xhash.QuickXorHash, false)
	return bookId
}

func getBody(sUrl string, jar *cookiejar.Jar) ([]byte, error) {
	referer := url.QueryEscape(sUrl)
	ctx := context.Background()
	cli := gohttp.NewClient(ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  jar,
		Headers: map[string]interface{}{
			"User-Agent": config.Conf.UserAgent,
			"Referer":    referer,
		},
	})
	resp, err := cli.Get(sUrl)
	if err != nil {
		return nil, err
	}
	bs, _ := resp.GetBody()
	if bs == nil {
		return nil, fmt.Errorf("ErrCode:%d, %s", resp.GetStatusCode(), resp.GetReasonPhrase())
	}
	return bs, nil
}

func postBody(sUrl string, d []byte, jar *cookiejar.Jar) ([]byte, error) {
	ctx := context.Background()
	cli := gohttp.NewClient(ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  jar,
		Headers: map[string]interface{}{
			"User-Agent":   config.Conf.UserAgent,
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body: d,
	})
	resp, err := cli.Post(sUrl)
	if err != nil {
		return nil, err
	}
	bs, _ := resp.GetBody()
	return bs, err
}

func NormalizeNamePart(value string) string {
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "&nbsp;", " ")
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	return value
}

func BuildOutputFileName(ext string, parts ...string) string {
	if ext == "" {
		ext = ".pdf"
	}
	cleaned := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = NormalizeNamePart(part)
		if part != "" {
			if _, ok := seen[part]; ok {
				continue
			}
			seen[part] = struct{}{}
			cleaned = append(cleaned, part)
		}
	}
	if len(cleaned) == 0 {
		cleaned = append(cleaned, "bookget")
	}
	return util.SanitizeFileName(strings.Join(cleaned, "_")) + ext
}

func ExtractHTMLTitle(bs []byte) string {
	text := string(bs)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']+)["']`),
		regexp.MustCompile(`(?is)<meta[^>]+name=["']title["'][^>]+content=["']([^"']+)["']`),
		regexp.MustCompile(`(?is)<title>\s*([^<>]+?)\s*</title>`),
	}
	for _, pattern := range patterns {
		if match := pattern.FindStringSubmatch(text); len(match) > 1 {
			title := NormalizeNamePart(match[1])
			if title != "" {
				return title
			}
		}
	}
	return ""
}

func postJSON(sUrl string, d interface{}, jar *cookiejar.Jar) ([]byte, error) {
	ctx := context.Background()
	cli := gohttp.NewClient(ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  jar,
		Headers: map[string]interface{}{
			"User-Agent":   config.Conf.UserAgent,
			"Content-Type": "application/json",
		},
		JSON: d,
	})
	resp, err := cli.Post(sUrl)
	if err != nil {
		return nil, err
	}
	bs, _ := resp.GetBody()
	return bs, err
}

func FileExist(path string) bool {
	fi, err := os.Stat(path)
	if err == nil && fi.Size() > 0 {
		return true
	}
	return false
}

func CreateDirectory(volumeId string) string {
	dirPath := config.Conf.Directory
	if volumeId != "" {
		dirPath = path.Join(config.Conf.Directory, "vol."+volumeId)
	}
	_ = os.MkdirAll(dirPath, os.ModePerm)
	return dirPath
}

func WaitNewCookie() {
	if FileExist(config.Conf.CookieFile) {
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	fmt.Println("请使用 bookget-gui 浏览器，打开图书网址，完成「真人验证 / 登录用户」，然后 「刷新」 网页.")
	go func() {
		defer wg.Done()
		for i := 0; i < 3600*8; i++ {
			if FileExist(config.Conf.CookieFile) {
				break
			}
			util.PrintSleepTime(config.Conf.Sleep)
		}
	}()
	wg.Wait()
}

func WaitNewCookieWithMsg(uri string) {
	_ = os.Remove(config.Conf.CookieFile)
	var wg sync.WaitGroup
	wg.Add(1)
	fmt.Println("请使用 bookget-gui 浏览器打开下面 URL，完成「真人验证 / 登录用户」，然后 「刷新」 网页.")
	fmt.Println(uri)
	go func() {
		defer wg.Done()
		for i := 0; i < 3600*8; i++ {
			if FileExist(config.Conf.CookieFile) {
				break
			}
			util.PrintSleepTime(config.Conf.Sleep)
		}
	}()
	wg.Wait()
}

func IsChinaIP(jar *cookiejar.Jar) bool {
	ctx := context.Background()
	cli := gohttp.NewClient(ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  jar,
		Headers: map[string]interface{}{
			"User-Agent": config.Conf.UserAgent,
			"Referer":    "http://ip-api.com/",
		},
	})
	resp, err := cli.Get("http://ip-api.com/json/?lang=zh-CN")
	if err != nil {
		return false
	}
	bs, _ := resp.GetBody()
	text := string(bs)
	return strings.Contains(text, "\"countryCode\":\"CN\"")
}
