package app

import (
	"bookget/config"
	"bookget/pkg/downloader"
	"bookget/pkg/gohttp"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

type Gslib struct {
	dt        *DownloadTask
	dm        *downloader.DownloadManager
	ServerUrl string
}

func NewGslib() *Gslib {
	ctx, cancel := context.WithCancel(context.Background())
	return &Gslib{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *Gslib) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (r *Gslib) Run(sUrl string) (msg string, err error) {
	// 检测用户是否提供了书目页面 URL（而非具体卷册页面）
	if strings.Contains(sUrl, "ancientObjectBookCatalog") {
		fmt.Printf("\033[31m[错误] 您提供的是书名目录页面的网址，无法直接下载。\n")
		fmt.Printf("请点击进入具体的卷册（如卷一、卷二），然后使用该卷册页面的网址来下载。\n")
		fmt.Printf("正确的网址格式应包含 ancientObjectBookView，而非 ancientObjectBookCatalog。\033[0m\n")
		return "请使用具体卷册的网址，而非书名目录页面的网址", fmt.Errorf("invalid URL: ancientObjectBookCatalog detected")
	}

	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl
	r.dt.BookId = r.getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	r.dt.Jar, _ = cookiejar.New(nil)
	r.ServerUrl = r.dt.UrlParsed.Host
	return r.download()
}

func (r *Gslib) getBookId(sUrl string) (bookId string) {
	// Example URL: https://zszy.gslib.com.cn/ancientObjectBookView/2004309788581335041
	m := regexp.MustCompile(`/([0-9]+)$`).FindStringSubmatch(sUrl)
	if m != nil {
		bookId = m[1]
	} else {
		bookId = getBookId(sUrl)
	}
	return bookId
}

func (r *Gslib) download() (msg string, err error) {
	log.Printf("Get %s\n", r.dt.Url)

	volumes, err := r.getVolumes(r.dt.Url, r.dt.Jar)
	if err != nil {
		return "getVolumes", err
	}

	sizeVol := len(volumes)
	for i, vol := range volumes {
		if !config.VolumeRange(i) {
			continue
		}

		title := r.dt.BookId
		if r.dt.Title != "" {
			title = r.dt.Title
		}

		if sizeVol == 1 {
			r.dt.SavePath = CreateDirectory(title)
		} else {
			vid := title + "/vol." + fmt.Sprintf("%04d", i+1)
			r.dt.SavePath = CreateDirectory(vid)
		}

		canvases, err := r.getCanvases(vol, r.dt.Jar)
		if err != nil || canvases == nil {
			log.Printf("getCanvases failed for volume %d: %v\n", i+1, err)
			continue
		}
		log.Printf(" %d/%d volume, %d pages \n", i+1, sizeVol, len(canvases))
		r.dm.AddImageTasks(canvases, r.dt.SavePath, config.Conf.FileExt, 0, nil, r.dt.Jar, true)
	}
	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	return "", nil
}



func (r *Gslib) getVolumes(sUrl string, _ *cookiejar.Jar) (volumes []string, err error) {
	// Currently assumes the initial URL is the only volume
	volumes = append(volumes, sUrl)
	return volumes, nil
}

func (r *Gslib) getCanvases(apiUrl string, jar *cookiejar.Jar) (canvases []string, err error) {
	bs, err := r.getBody(apiUrl, jar)
	if err != nil {
		return nil, err
	}
	html := string(bs)

	objectBookId := r.extractVar(html, `let objectBookId = "([^"]+)"`)
	bookId := r.extractVar(html, `let bookId = "([^"]+)"`)
	catalogId := r.extractVar(html, `let catalogId = "([^"]+)"`)
	r.dt.Title = r.extractVar(html, `let bookTitle = "([^"]+)"`)

	if r.dt.Title != "" {
		r.dt.Title = r.unicodeUnescape(r.dt.Title)
	}

	pagesMatch := regexp.MustCompile(`var pages = \[(.*?)\];`).FindStringSubmatch(html)
	if pagesMatch == nil || objectBookId == "" || bookId == "" || catalogId == "" {
		return nil, fmt.Errorf("failed to extract critical metadata from HTML")
	}

	pagesStr := pagesMatch[1]
	pageParts := strings.Split(pagesStr, ",")
	for _, p := range pageParts {
		p = strings.Trim(p, ` "`)
		if p == "" {
			continue
		}
		// Pattern: https://zszy.gslib.com.cn/api/v1/library/ancient/bookObject/{objectBookId}/{bookId}/{catalogId}/{filename}
		imgUrl := fmt.Sprintf("https://%s/api/v1/library/ancient/bookObject/%s/%s/%s/%s", r.ServerUrl, objectBookId, bookId, catalogId, p)
		canvases = append(canvases, imgUrl)
	}

	return canvases, nil
}

func (r *Gslib) unicodeUnescape(s string) string {
	quoted := `"` + s + `"`
	var unquoted string
	err := json.Unmarshal([]byte(quoted), &unquoted)
	if err != nil {
		return s
	}
	return unquoted
}

func (r *Gslib) extractVar(html, pattern string) string {
	m := regexp.MustCompile(pattern).FindStringSubmatch(html)
	if m != nil {
		return m[1]
	}
	return ""
}

func (r *Gslib) getBody(sUrl string, jar *cookiejar.Jar) ([]byte, error) {
	referer := sUrl
	ctx := context.Background()
	cli := gohttp.NewClient(ctx, gohttp.Options{
		CookieJar: jar,
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
	return bs, nil
}


