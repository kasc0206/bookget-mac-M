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
	"strconv"
	"time"
)

type Berkeley struct {
	dt *DownloadTask
	dm *downloader.DownloadManager
}

func NewBerkeley() *Berkeley {
	ctx, cancel := context.WithCancel(context.Background())
	return &Berkeley{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *Berkeley) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

type BerkeleyResponse struct {
	Name        string `json:"name"`
	Url         string `json:"url"`
	Size        int    `json:"size"`
	Description string `json:"description"`
}

func (r *Berkeley) Run(sUrl string) (msg string, err error) {
	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl
	r.dt.BookId = r.getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	r.dt.Jar, _ = cookiejar.New(nil)
	return r.download()
}

func (r *Berkeley) getBookId(sUrl string) (bookId string) {
	m := regexp.MustCompile(`(?i)record/([A-z0-9_-]+)`).FindStringSubmatch(sUrl)
	if m != nil {
		bookId = m[1]
	}
	return bookId
}

func (r *Berkeley) download() (msg string, err error) {
	log.Printf("Get %s\n", r.dt.Url)

	r.dt.SavePath = config.Conf.Directory
	canvases, err := r.getCanvases(r.dt.Url, r.dt.Jar)
	if err != nil || canvases == nil {
		return "requested URL was not found.", err
	}
	log.Printf(" %d files \n", len(canvases))
	r.dm.AddImageTasks(canvases, r.dt.SavePath, config.Conf.FileExt, 0, nil, r.dt.Jar, true)
	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	return "", nil
}



func (r *Berkeley) getVolumes(_ string, _ *cookiejar.Jar) (volumes []string, err error) {
	return nil, fmt.Errorf("getVolumes not implemented for Berkeley")
}

func (r *Berkeley) getCanvases(sUrl string, jar *cookiejar.Jar) (canvases []string, err error) {

	apiUrl := "https://" + r.dt.UrlParsed.Host + "/api/v1/file?recid=" + r.dt.BookId +
		"&file_types=%5B%5D&hidden_types=%5B%22pdf%3Bpdfa%22%2C%22hocr%22%5D&ln=en&hr=1&_=" + strconv.FormatInt(time.Now().Unix(), 10)
	bs, err := r.getBody(apiUrl, jar)
	if err != nil {
		return
	}

	var resT = make([]BerkeleyResponse, 0, 64)
	if err = json.Unmarshal(bs, &resT); err != nil {
		log.Printf("json.Unmarshal failed: %s\n", err)
		return
	}
	for _, ret := range resT {
		canvases = append(canvases, ret.Url)
	}
	return
}

func (r *Berkeley) getBody(apiUrl string, jar *cookiejar.Jar) ([]byte, error) {
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
