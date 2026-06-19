package app

import (
	"bookget/config"
	"bookget/pkg/downloader"
	"bookget/pkg/gohttp"
	"bookget/pkg/util"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
)

type Luoyang struct {
	dt *DownloadTask
	dm *downloader.DownloadManager
}

func NewLuoyang() *Luoyang {
	ctx, cancel := context.WithCancel(context.Background())
	return &Luoyang{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *Luoyang) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (p *Luoyang) Run(sUrl string) (msg string, err error) {
	p.dt.UrlParsed, err = url.Parse(sUrl)
	p.dt.Url = sUrl
	p.dt.BookId = p.getBookId(p.dt.Url)
	if p.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	p.dt.Jar, _ = cookiejar.New(nil)
	return p.download()
}

func (p *Luoyang) getBookId(sUrl string) (bookId string) {
	if m := regexp.MustCompile(`&id=(\d+)`).FindStringSubmatch(sUrl); m != nil {
		bookId = m[1]
	}
	return bookId
}

func (p *Luoyang) download() (msg string, err error) {
	log.Printf("Get %s\n", p.dt.Url)
	respVolume, err := p.getVolumes(p.dt.Url, p.dt.Jar)
	if err != nil {
		fmt.Println(err)
		return "getVolumes", err
	}
	p.dt.SavePath = config.Conf.Directory
	for i, vol := range respVolume {
		if !config.VolumeRange(i) {
			continue
		}
		log.Printf(" %d/%d volume, %s \n", i+1, len(respVolume), vol)
		sortId := fmt.Sprintf("%04d", i+1)
		filename := BuildOutputFileName(path.Ext(vol), p.dt.Title, sortId)
		dest := filepath.Join(p.dt.SavePath, filename)
		p.do(dest, vol)
		util.PrintSleepTime(config.Conf.Sleep)
	}
	return msg, err
}

func (p *Luoyang) do(dest, pdfUrl string) (msg string, err error) {
	headers := map[string]string{
		"User-Agent": config.Conf.UserAgent,
	}
	p.dm.AddFromLegacy(pdfUrl, "GET", headers, nil, filepath.Dir(dest), filepath.Base(dest), config.Conf.Threads, p.dt.Jar, false)
	if len(p.dm.Tasks()) > 0 {
		p.dm.Start()
	}
	return "", nil
}

func (p *Luoyang) getVolumes(sUrl string, jar *cookiejar.Jar) (volumes []string, err error) {
	bs, err := p.getBody(sUrl, jar)
	if err != nil {
		return
	}
	if p.dt.Title == "" {
		p.dt.Title = ExtractHTMLTitle(bs)
	}
	//取册数
	matches := regexp.MustCompile(`href=["']viewer.php\?pdf=(.+?)\.pdf&`).FindAllStringSubmatch(string(bs), -1)
	if matches == nil {
		return
	}
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		ids = append(ids, match[1])
	}
	hostUrl := util.GetHostUrl(sUrl)
	volumes = make([]string, 0, len(ids))
	for _, v := range ids {
		s := fmt.Sprintf("%s%s.pdf", hostUrl, v)
		volumes = append(volumes, s)
	}
	return volumes, nil
}

func (p *Luoyang) getCanvases(_ string, _ *cookiejar.Jar) (canvases []string, err error) {
	return nil, fmt.Errorf("getCanvases not implemented for Luoyang")
}

func (p *Luoyang) getBody(sUrl string, jar *cookiejar.Jar) ([]byte, error) {
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
		return nil, &HTTPError{
			StatusCode: resp.GetStatusCode(),
			URL:        sUrl,
			Message:    resp.GetReasonPhrase(),
		}
	}
	return bs, nil
}

func (p *Luoyang) postBody(sUrl string, d []byte) ([]byte, error) {
	return nil, errors.New("postBody not implemented for Luoyang")
}
