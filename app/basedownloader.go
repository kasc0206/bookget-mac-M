package app

import (
	"bookget/config"
	"bookget/pkg/downloader"
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
	"regexp"
	"strings"
	"sync"
)

// HTTPError 表示 HTTP 请求错误（含状态码）
type HTTPError struct {
	StatusCode int
	URL        string
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s (%s)", e.StatusCode, e.Message, e.URL)
}

// BaseDownloader 封装下载器的公共字段和方法
// 各站点下载器可通过嵌入此类型减少重复代码
type BaseDownloader struct {
	// 公共字段
	Dm   *downloader.DownloadManager // 下载管理器
	Ctx  context.Context
	Cancel context.CancelFunc

	// 下载任务元数据
	BookId   string
	Url      string
	UrlParsed *url.URL
	SavePath string
	VolumeId string
	Jar      *cookiejar.Jar

	// 下载结果
	Canvases []string
	Body     []byte
}

// NewBaseDownloader 创建 BaseDownloader 实例
func NewBaseDownloader() *BaseDownloader {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseDownloader{
		Ctx:    ctx,
		Cancel: cancel,
		Dm:     downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

// GetBookId 从 URL 生成唯一 ID
func (b *BaseDownloader) GetBookId(sUrl string) string {
	if sUrl == "" {
		return ""
	}
	mh := xhash.NewMultiHasher()
	_, _ = io.Copy(mh, bytes.NewBuffer([]byte(sUrl)))
	bookId, _ := mh.SumString(xhash.QuickXorHash, false)
	return bookId
}

// GetBody 发送 GET 请求
func (b *BaseDownloader) GetBody(sUrl string) ([]byte, error) {
	referer := url.QueryEscape(sUrl)
	cli := gohttp.NewClient(b.Ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  b.Jar,
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
		return nil, &HTTPError{
			StatusCode: resp.GetStatusCode(),
			URL:        sUrl,
			Message:    resp.GetReasonPhrase(),
		}
	}
	return bs, nil
}

// PostBody 发送 POST 表单请求
func (b *BaseDownloader) PostBody(sUrl string, data []byte) ([]byte, error) {
	cli := gohttp.NewClient(b.Ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  b.Jar,
		Headers: map[string]interface{}{
			"User-Agent":   config.Conf.UserAgent,
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body: data,
	})
	resp, err := cli.Post(sUrl)
	if err != nil {
		return nil, err
	}
	bs, _ := resp.GetBody()
	return bs, err
}

// PostJSON 发送 POST JSON 请求
func (b *BaseDownloader) PostJSON(sUrl string, data interface{}) ([]byte, error) {
	cli := gohttp.NewClient(b.Ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  b.Jar,
		Headers: map[string]interface{}{
			"User-Agent":   config.Conf.UserAgent,
			"Content-Type": "application/json",
		},
		JSON: data,
	})
	resp, err := cli.Post(sUrl)
	if err != nil {
		return nil, err
	}
	bs, _ := resp.GetBody()
	return bs, err
}

// AddImageTask 添加单张图片下载任务
func (b *BaseDownloader) AddImageTask(imgUrl, filename string) {
	b.Dm.AddFromLegacy(imgUrl, "GET", nil, nil, b.SavePath, filename, 1, b.Jar, true)
}

// AddImageTasks 批量添加图片下载任务
func (b *BaseDownloader) AddImageTasks(imgUrls []string, startIndex int) int {
	headers := map[string]string{"User-Agent": config.Conf.UserAgent}
	return b.Dm.AddImageTasks(imgUrls, b.SavePath, config.Conf.FileExt, startIndex, headers, b.Jar, true)
}

// StartDownload 开始下载并等待完成
func (b *BaseDownloader) StartDownload() {
	if b.Dm != nil && len(b.Dm.Tasks()) > 0 {
		b.Dm.Start()
	}
}

// CreateVolumeDir 创建分卷目录
func (b *BaseDownloader) CreateVolumeDir(volumeId string) string {
	if volumeId == "" {
		b.SavePath = config.Conf.Directory
		return b.SavePath
	}
	dirPath, _ := downloader.CreateVolumeDirectory(config.Conf.Directory, volumeId)
	b.SavePath = dirPath
	return dirPath
}

// NormalizeNamePart 规范化标题文本
func (b *BaseDownloader) NormalizeNamePart(value string) string {
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "&nbsp;", " ")
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	return value
}

// ExtractHTMLTitle 从 HTML 提取标题
func (b *BaseDownloader) ExtractHTMLTitle(bs []byte) string {
	text := string(bs)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']+)["']`),
		regexp.MustCompile(`(?is)<meta[^>]+name=["']title["'][^>]+content=["']([^"']+)["']`),
		regexp.MustCompile(`(?is)<title>\s*([^<>]+?)\s*</title>`),
	}
	for _, pattern := range patterns {
		if match := pattern.FindStringSubmatch(text); len(match) > 1 {
			title := b.NormalizeNamePart(match[1])
			if title != "" {
				return title
			}
		}
	}
	return ""
}

// BuildOutputFileName 构建输出文件名
func (b *BaseDownloader) BuildOutputFileName(ext string, parts ...string) string {
	if ext == "" {
		ext = ".pdf"
	}
	cleaned := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = b.NormalizeNamePart(part)
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

// FileExist 检查文件是否已存在
func (b *BaseDownloader) FileExist(filePath string) bool {
	fi, err := os.Stat(filePath)
	return err == nil && fi.Size() > 0
}

// WaitNewCookie 等待用户完成验证
func (b *BaseDownloader) WaitNewCookie() {
	if b.FileExist(config.Conf.CookieFile) {
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	fmt.Println("请使用 bookget-gui 浏览器，打开图书网址，完成「真人验证 / 登录用户」，然后 「刷新」 网页.")
	go func() {
		defer wg.Done()
		for i := 0; i < 3600*8; i++ {
			if b.FileExist(config.Conf.CookieFile) {
				break
			}
			util.PrintSleepTime(config.Conf.Sleep)
		}
	}()
	wg.Wait()
}

// WaitNewCookieWithMsg 等待用户完成验证（带自定义消息）
func (b *BaseDownloader) WaitNewCookieWithMsg(uri string) {
	if b.FileExist(config.Conf.CookieFile) {
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	fmt.Printf("请使用 bookget-gui 浏览器，打开 %s ，完成登录/验证。\n", uri)
	go func() {
		defer wg.Done()
		for i := 0; i < 3600*8; i++ {
			if b.FileExist(config.Conf.CookieFile) {
				break
			}
			util.PrintSleepTime(config.Conf.Sleep)
		}
	}()
	wg.Wait()
}

// ResolveURL 解析 URL
func (b *BaseDownloader) ResolveURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	b.Url = rawURL
	b.UrlParsed = u
	b.BookId = b.GetBookId(rawURL)
	return u, nil
}

// CreateJar 创建新的 CookieJar
func (b *BaseDownloader) CreateJar() {
	b.Jar, _ = cookiejar.New(nil)
}


