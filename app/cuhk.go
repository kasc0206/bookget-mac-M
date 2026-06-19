package app

import (
	"bookget/config"
	"bookget/model/cuhk"
	"bookget/pkg/chromedphelper"
	"bookget/pkg/downloader"
	"bookget/pkg/progressbar"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type Cuhk struct {
	dm     *downloader.DownloadManager
	ctx    context.Context
	cancel context.CancelFunc
	client *http.Client

	responseBody []byte
	urlsFile     string
	bufBuilder   strings.Builder
	bufBody      string
	canvases     []string
	cookies      []*http.Cookie

	rawUrl    string
	parsedUrl *url.URL
	savePath  string
	bookId    string
}

func NewCuhk() *Cuhk {
	ctx, cancel := context.WithCancel(context.Background())
	dm := downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent)

	client, _ := NewHttpClient()
	return &Cuhk{
		dm:     dm,
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (r *Cuhk) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	lastPos := strings.Index(sUrl, "#")
	if lastPos > 0 {
		r.rawUrl = strings.Replace(sUrl[:lastPos], "hk/sc/", "hk/en/", -1)
	} else {
		r.rawUrl = strings.Replace(sUrl, "hk/sc/", "hk/en/", -1)
	}
	r.parsedUrl, _ = url.Parse(r.rawUrl)
	msg, err := r.Run()
	return map[string]interface{}{
		"url": r.rawUrl,
		"msg": msg,
	}, err
}

func (r *Cuhk) getBookId() string {
	const (
		IdPattern  = `(?i)/item/([^#/]+)`
		IdPattern2 = `(?i)/object/([^#/]+)`
	)

	var (
		IdRe  = regexp.MustCompile(IdPattern)
		IdRe2 = regexp.MustCompile(IdPattern2)
	)

	if matches := IdRe.FindStringSubmatch(r.rawUrl); len(matches) > 1 {
		return matches[1]
	}

	if matches := IdRe2.FindStringSubmatch(r.rawUrl); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func (r *Cuhk) Run() (msg string, err error) {
	r.bookId = r.getBookId()
	if r.bookId == "" {
		return "[err=getBookId]", err
	}
	r.savePath = config.Conf.Directory

	// 尝试获取页面内容（cookie文件 → chromedp → 引导用户）
	if err = r.resolvePage(); err != nil {
		return "", err
	}

	canvases, err := r.parseCanvases(r.responseBody)
	if err != nil || canvases == nil {
		return "", err
	}
	r.canvases = canvases

	r.savePath = config.Conf.Directory
	r.urlsFile = path.Join(r.savePath, "urls.txt")
	_ = os.WriteFile(r.urlsFile, []byte(r.bufBuilder.String()), os.ModePerm)
	fmt.Printf("已生成图片URLs文件[%s]\n", r.urlsFile)

	r.downloadImages(canvases)
	return "", nil
}

// resolvePage 尝试获取 CUHK 页面内容（绕过 AWS WAF）
func (r *Cuhk) resolvePage() error {
	// 方案1: cookie 文件
	if config.Conf.CookieFile != "" {
		fmt.Println("尝试使用 cookie 文件访问...")
		jar, _ := cookiejar.New(nil)
		bs, err := r.fetchPage(r.rawUrl, jar)
		if err == nil {
			r.responseBody = bs
			return nil
		}
		fmt.Printf("cookie 文件请求失败: %v\n", err)
	}

	// 方案2: chromedp 绕过 WAF
	fmt.Println("正在通过 headless Chrome 绕过 AWS WAF 挑战...")
	html, jar, err := chromedphelper.SolveWAFChallenge(r.rawUrl, 90*time.Second)
	if err == nil && len(html) > 1000 && !strings.Contains(html, "Human Verification") {
		fmt.Println("AWS WAF 挑战通过！")
		r.responseBody = []byte(html)
		parsedURL, _ := url.Parse(r.rawUrl)
		r.cookies = jar.Cookies(parsedURL)
		return nil
	}
	if err != nil {
		fmt.Printf("headless Chrome 挑战失败: %v\n", err)
	} else {
		fmt.Println("headless Chrome 未能完成 WAF 挑战")
	}

	// 方案3: 引导用户手动操作
	fmt.Println("\n⚠️  CUHK 香港中文大学使用 AWS WAF 保护，需要浏览器完成验证。")
	fmt.Println("   请运行 bookget-gui 或在浏览器中手动打开以下链接完成验证：")
	fmt.Println("   " + r.rawUrl)
	fmt.Println("   然后在页面中点击右键 → 另存为 → 保存 HTML 文件")
	fmt.Println("   再使用 --cookie 参数传入 cookie.txt（从浏览器导出）")
	return fmt.Errorf("需要浏览器完成 AWS WAF 验证")
}

// fetchPage 使用标准 HTTP 客户端获取页面内容
func (r *Cuhk) fetchPage(pageURL string, jar *cookiejar.Jar) ([]byte, error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", config.Conf.UserAgent)

	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// parseCanvases 从页面 HTML 中解析 IIIF 图片 URL
func (r *Cuhk) parseCanvases(body []byte) ([]string, error) {
	var resp cuhk.ResponsePage
	matches := regexp.MustCompile(`"pages":([^]]+)]`).FindSubmatch(body)
	if matches == nil {
		return nil, errors.New("[err=parseCanvases: pages not found]")
	}
	data := []byte("{\"pages\":" + string(matches[1]) + "]}")
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Printf("json.Unmarshal failed: %s\n", err)
		return nil, err
	}
	var canvases []string
	for _, page := range resp.ImagePage {
		imgUrl := fmt.Sprintf("https://%s/iiif/2/%s/%s", r.parsedUrl.Host, page.Identifier, config.Conf.Format)
		r.bufBuilder.WriteString(imgUrl)
		r.bufBuilder.WriteString("\n")
		canvases = append(canvases, imgUrl)
	}
	return canvases, nil
}

// downloadImages 下载 IIIF 图片，使用 chromedp 获取的 cookies
func (r *Cuhk) downloadImages(canvases []string) {
	fmt.Println()
	sizeVol := len(canvases)
	bar := progressbar.Default(int64(sizeVol), "downloading")

	// 创建 cookie jar 并注入 cookies
	jar, _ := cookiejar.New(nil)
	if parsedURL, err := url.Parse(r.rawUrl); err == nil && len(r.cookies) > 0 {
		jar.SetCookies(parsedURL, r.cookies)
	}

	for i, uri := range canvases {
		if uri == "" || !config.PageRange(i, sizeVol) {
			bar.Add(1)
			continue
		}
		sortId := fmt.Sprintf("%04d", i+1)
		filename := sortId + config.Conf.FileExt
		targetFilePath := path.Join(r.savePath, filename)
		if FileExist(targetFilePath) {
			bar.Add(1)
			continue
		}

		err := r.downloadIIIFImage(uri, targetFilePath, jar)
		if err != nil {
			log.Printf("下载失败: %s (%v)\n", filename, err)
		}
		bar.Add(1)
	}
	fmt.Println()
}

// downloadIIIFImage 下载单个 IIIF 图片
func (r *Cuhk) downloadIIIFImage(imgURL, targetPath string, jar *cookiejar.Jar) error {
	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	req, err := http.NewRequest("GET", imgURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", config.Conf.UserAgent)
	req.Header.Set("Referer", r.rawUrl)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	outFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}
