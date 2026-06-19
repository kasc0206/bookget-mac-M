package app

import (
	"bookget/config"
	"bookget/pkg/downloader"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

type HannomNlv struct {
	dt   *DownloadTask
	dm   *downloader.DownloadManager
	body []byte
	ctx  context.Context
}

func NewHannomNlv() *HannomNlv {
	ctx, cancel := context.WithCancel(context.Background())
	return &HannomNlv{
		dt:  new(DownloadTask),
		dm:  downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
		ctx: ctx,
	}
}

func (r *HannomNlv) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (r *HannomNlv) Run(sUrl string) (msg string, err error) {
	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl
	r.dt.Jar, _ = cookiejar.New(nil)
	r.dt.BookId = r.getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	return r.download()
}

func (r *HannomNlv) getBookId(sUrl string) (bookId string) {
	var err error
	r.body, err = getBody(sUrl, r.dt.Jar)
	if err != nil {
		return ""
	}
	m := regexp.MustCompile(`var[\s+]documentOID[\s+]=[\s+]['"]([^“]+?)['"];`).FindSubmatch(r.body)
	if m != nil {
		return string(m[1])
	}
	return ""
}

func (r *HannomNlv) download() (msg string, err error) {
	log.Printf("Get %s\n", r.dt.Url)
	r.dt.SavePath = config.Conf.Directory
	canvases, err := r.getCanvases(r.dt.Url, r.dt.Jar)
	if err != nil || canvases == nil {
		fmt.Println(err)
	}
	log.Printf(" %d pages \n", len(canvases))
	r.dm.AddImageTasks(canvases, r.dt.SavePath, config.Conf.FileExt, 0, nil, r.dt.Jar, true)
	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	return "", nil
}

func (r *HannomNlv) getVolumes(_ string, _ *cookiejar.Jar) (volumes []string, err error) {
	return nil, fmt.Errorf("getVolumes not implemented for HannomNlv")
}

func (r *HannomNlv) getCanvases(sUrl string, jar *cookiejar.Jar) (canvases []string, err error) {
	matches := regexp.MustCompile(`'([^']+)':\{'w':([0-9]+),'h':([0-9]+)\}`).FindAllSubmatch(r.body, -1)
	if matches == nil {
		return nil, errors.New("No image")
	}
	apiUrl := r.dt.UrlParsed.Scheme + "://" + r.dt.UrlParsed.Host
	match := regexp.MustCompile(`imageserverPageTileImageRequest[\s+]=[\s+]['"]([^;]+)['"];`).FindSubmatch(r.body)
	if match != nil {
		apiUrl += string(match[1])
	} else {
		apiUrl += "/hannom/cgi-bin/imageserver/imageserver.pl?color=all&ext=jpg"
	}
	for _, m := range matches {
		imgUrl := apiUrl + fmt.Sprintf("&oid=%s.%s&key=&width=%s&crop=0,0,%s,%s", r.dt.BookId, m[1], m[2], m[2], m[3])
		canvases = append(canvases, imgUrl)
	}
	return canvases, err
}

func (r *HannomNlv) getBody(_ string, _ *cookiejar.Jar) ([]byte, error) {
	return nil, fmt.Errorf("getBody not implemented for HannomNlv")
}

func (r *HannomNlv) postBody(_ string, _ []byte) ([]byte, error) {
	return nil, fmt.Errorf("postBody not implemented for HannomNlv")
}
