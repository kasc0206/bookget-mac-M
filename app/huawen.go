package app

import (
	"bookget/config"
	"bookget/pkg/downloader"
	"bookget/pkg/gohttp"
	"bookget/pkg/util"
	"context"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

type Huawen struct {
	dt *DownloadTask
	dm *downloader.DownloadManager
}

func NewHuawen() *Huawen {
	ctx, cancel := context.WithCancel(context.Background())
	return &Huawen{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *Huawen) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (r *Huawen) Run(sUrl string) (msg string, err error) {
	if !strings.Contains(sUrl, "/reader") && strings.Contains(sUrl, "/zh-tw/book/") {
		sUrl += "/reader"
	}

	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl

	r.dt.BookId = getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	r.dt.Jar, _ = cookiejar.New(nil)
	return r.download()
}

func (r *Huawen) getBookId(_ string) (bookId string) {
	return ""
}

func (r *Huawen) download() (msg string, err error) {
	log.Printf("Get %s\n", r.dt.Url)

	respVolume, err := r.getVolumes(r.dt.Url, r.dt.Jar)
	if err != nil {
		fmt.Println(err)
		return "getVolumes", err
	}
	savePath := config.Conf.Directory
	for i, vol := range respVolume {
		if !config.VolumeRange(i) {
			continue
		}
		log.Printf(" %d/%d PDFs \n", i+1, len(respVolume))
		u, _ := url.Parse(vol)
		headers := map[string]string{
			"Referer": "https://" + r.dt.UrlParsed.Host + "/pdfjs/web/viewer.html?file=" + u.Path,
		}
		filename := util.FileName(vol)
		r.dm.AddFromLegacy(vol, "GET", headers, nil, savePath, filename, 1, r.dt.Jar, true)
	}
	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	return "", nil
}



func (r *Huawen) getVolumes(sUrl string, jar *cookiejar.Jar) (volumes []string, err error) {
	bs, err := r.getBody(sUrl, jar)
	if err != nil {
		return
	}
	matches := regexp.MustCompile(`(?i)viewer.html\?file=([^"]+)"`).FindAllSubmatch(bs, -1)
	if matches == nil {
		return
	}
	for _, match := range matches {
		sPath := strings.TrimSpace(string(match[1]))
		if pos := strings.Index(sPath, "&"); pos > 0 {
			sPath = sPath[:pos]
		}
		pdfUrl := "https://" + r.dt.UrlParsed.Host + sPath
		volumes = append(volumes, pdfUrl)
	}
	return volumes, nil
}

func (r *Huawen) getCanvases(_ string, _ *cookiejar.Jar) (canvases []string, err error) {
	return nil, fmt.Errorf("getCanvases not implemented for Huawen")
}

func (r *Huawen) getBody(apiUrl string, jar *cookiejar.Jar) ([]byte, error) {
	referer := url.QueryEscape(apiUrl)
	ctx := context.Background()
	cli := gohttp.NewClient(ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  jar,
		Headers: map[string]interface{}{
			"User-Agent": config.Conf.UserAgent,
			"Referer":    referer,
		},
	})
	resp, err := cli.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	bs, _ := resp.GetBody()
	if resp.GetStatusCode() != 200 || bs == nil {
		return nil, &HTTPError{
			StatusCode: resp.GetStatusCode(),
			URL:        apiUrl,
			Message:    resp.GetReasonPhrase(),
		}
	}
	return bs, nil
}
