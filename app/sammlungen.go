package app

import (
	"bookget/config"
	"bookget/pkg/downloader"
	"context"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

type Sammlungen struct {
	dt *DownloadTask
	dm *downloader.DownloadManager
}

func NewSammlungen() *Sammlungen {
	ctx, cancel := context.WithCancel(context.Background())
	return &Sammlungen{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *Sammlungen) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (r *Sammlungen) Run(sUrl string) (msg string, err error) {

	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl

	r.dt.BookId = r.getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	r.dt.Jar, _ = cookiejar.New(nil)
	return r.download()
}

func (r *Sammlungen) getBookId(sUrl string) (bookId string) {
	m := regexp.MustCompile(`/view/([A-z\d]+)`).FindStringSubmatch(sUrl)
	if m != nil {
		bookId = m[1]
	}
	return bookId
}

func (r *Sammlungen) download() (msg string, err error) {
	log.Printf("Get %s\n", r.dt.Url)
	manifestUrl := fmt.Sprintf("https://api.digitale-sammlungen.de/iiif/presentation/v2/%s/manifest", r.dt.BookId)
	var iiif IIIF
	return iiif.InitWithId(r.dt.Index, manifestUrl, r.dt.BookId)
}
