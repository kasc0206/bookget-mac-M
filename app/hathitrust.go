package app

import (
	"bookget/config"
	"bookget/pkg/downloader"
	"bookget/pkg/gohttp"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
)

type Hathitrust struct {
	dt *DownloadTask
	dm *downloader.DownloadManager
}

func NewHathitrust() *Hathitrust {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hathitrust{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *Hathitrust) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (r Hathitrust) Run(sUrl string) (msg string, err error) {
	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl
	r.dt.BookId = r.getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	r.dt.Jar, _ = cookiejar.New(nil)
	return r.download()
}

func (r Hathitrust) getBookId(sUrl string) (bookId string) {
	m := regexp.MustCompile(`id=([^&]+)`).FindStringSubmatch(sUrl)
	if m != nil {
		bookId = m[1]
	}
	return bookId
}

func (r Hathitrust) download() (msg string, err error) {
	log.Printf("Get %s\n", r.dt.Url)
	canvases, err := r.getCanvases(r.dt.Url, r.dt.Jar)
	if err != nil {
		fmt.Println(err.Error())
		return "requested URL was not found.", err
	}
	r.dt.SavePath = config.Conf.Directory
	msg, err = r.do(canvases)
	return msg, err
}

func (r Hathitrust) do(imgUrls []string) (msg string, err error) {
	if imgUrls == nil {
		return "", nil
	}
	referer := url.QueryEscape(r.dt.Url)
	headers := map[string]string{
		"User-Agent": config.Conf.UserAgent,
		"Referer":    referer,
	}
	r.dm.AddImageTasks(imgUrls, r.dt.SavePath, config.Conf.FileExt, 0, headers, r.dt.Jar, true)
	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	return "", nil
}

func (r Hathitrust) getVolumes(_ string, _ *cookiejar.Jar) (volumes []string, err error) {
	return nil, fmt.Errorf("getVolumes not implemented for Hathitrust")
}

func (r Hathitrust) getCanvases(_ string, _ *cookiejar.Jar) (canvases []string, err error) {
	bs, err := r.getBody(r.dt.Url, r.dt.Jar)
	if err != nil || bs == nil {
		return nil, err
	}
	//
	if !bytes.Contains(bs, []byte("HT.params.allowSinglePageDownload = true;")) {
		return nil, errors.New("This item is not available online —  Limited - search only")
	}
	// HT.params.totalSeq = 1220;
	matches := regexp.MustCompile(`HT.params.totalSeq = ([0-9]+);`).FindStringSubmatch(string(bs))
	if matches == nil {
		return
	}
	size, _ := strconv.Atoi(matches[1])

	canvases = make([]string, 0, size)
	format := "jpeg"
	switch config.Conf.FileExt {
	case ".png":
		format = "png"
	case ".tif":
		format = "tiff"
	}
	for i := 0; i < size; i++ {
		imgurl := fmt.Sprintf("https://babel.hathitrust.org/cgi/imgsrv/image?id=%s&attachment=1&size=ppi%%3A300&format=image/%s&seq=%d", r.dt.BookId, format, i+1)
		canvases = append(canvases, imgurl)
	}
	return canvases, err
}

func (r Hathitrust) getBody(apiUrl string, jar *cookiejar.Jar) ([]byte, error) {
	ctx := context.Background()
	cli := gohttp.NewClient(ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  jar,
		Headers: map[string]interface{}{
			"User-Agent": config.Conf.UserAgent,
		},
	})
	resp, err := cli.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	bs, _ := resp.GetBody()
	if bs == nil {
		err = errors.New(resp.GetReasonPhrase())
		return nil, err
	}
	return bs, nil
}
