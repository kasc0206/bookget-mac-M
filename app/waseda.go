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
	"path"
	"regexp"
	"sort"
)

type Waseda struct {
	dt *DownloadTask
	dm *downloader.DownloadManager
}

func NewWaseda() *Waseda {
	ctx, cancel := context.WithCancel(context.Background())
	return &Waseda{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *Waseda) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}
func (r Waseda) Run(sUrl string) (msg string, err error) {

	r.dt.UrlParsed, err = url.Parse(sUrl)
	if err != nil {
		return "", err
	}
	r.dt.Url = sUrl

	r.dt.BookId = getBookId(r.dt.Url)
	r.dt.Jar, _ = cookiejar.New(nil)
	return r.download()
}

func (r Waseda) download() (msg string, err error) {
	respVolume, err := r.getVolumes(r.dt.Url, r.dt.Jar)
	if err != nil {
		fmt.Println(err)
		return "getVolumes", err
	}
	if config.Conf.FileExt == ".pdf" {
		for i, vol := range respVolume {
			if !config.VolumeRange(i) {
				continue
			}
			sortId := fmt.Sprintf("%04d", i+1)
			log.Printf(" %d/%d volume, URL:%s \n", i+1, len(respVolume), vol)
			filename := sortId + config.Conf.FileExt
			r.dm.AddFromLegacy(vol, "GET", nil, nil, config.Conf.Directory, filename, 1, r.dt.Jar, true)
		}
	} else {
		for i, vol := range respVolume {
			if !config.VolumeRange(i) {
				continue
			}
			var savePath string
			if len(respVolume) == 1 {
				savePath = config.Conf.Directory
			} else {
				vid := fmt.Sprintf("%04d", i+1)
				savePath, _ = downloader.CreateVolumeDirectory(config.Conf.Directory, vid)
			}
			canvases, err := r.getCanvases(vol, r.dt.Jar)
			if err != nil || canvases == nil {
				fmt.Println(err)
				continue
			}
			log.Printf(" %d/%d volume, %d pages \n", i+1, len(respVolume), len(canvases))
			r.dm.AddImageTasks(canvases, savePath, config.Conf.FileExt, 0, nil, r.dt.Jar, true)
		}
	}

	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	return "", nil
}

func (r Waseda) getVolumes(sUrl string, jar *cookiejar.Jar) (volumes []string, err error) {
	bs, err := r.getBody(sUrl, jar)
	if err != nil {
		return
	}
	text := string(bs)
	//取册数
	matches := regexp.MustCompile(`href=["'](.+?)\.html["']`).FindAllStringSubmatch(text, -1)
	if matches == nil {
		return
	}
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		ids = append(ids, match[1])
	}
	sort.Sort(util.SortByStr(ids))
	volumes = make([]string, 0, len(ids))
	for _, v := range ids {
		var htmlUrl string
		if config.Conf.FileExt == ".pdf" {
			htmlUrl = sUrl + v + ".pdf"
		} else {
			htmlUrl = sUrl + v + ".html"
		}
		volumes = append(volumes, htmlUrl)
	}
	return volumes, nil
}

func (r Waseda) getCanvases(sUrl string, jar *cookiejar.Jar) (canvases []string, err error) {
	bs, err := r.getBody(sUrl, jar)
	if err != nil {
		return
	}
	text := string(bs)
	//取册数
	matches := regexp.MustCompile(`(?i)href=["'](.+?)\.jpg["']\s+target="_blank">\d+</A>`).FindAllStringSubmatch(text, -1)
	if matches == nil {
		return
	}
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		ids = append(ids, match[1])
	}
	sort.Sort(util.SortByStr(ids))
	canvases = make([]string, 0, len(ids))
	dir, _ := path.Split(sUrl)
	for _, v := range ids {
		imgUrl := dir + v + ".jpg"
		canvases = append(canvases, imgUrl)
	}
	return canvases, nil
}

func (r Waseda) getBody(apiUrl string, jar *cookiejar.Jar) ([]byte, error) {
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
	if bs == nil {
		return nil, fmt.Errorf("ErrCode:%d, %s", resp.GetStatusCode(), resp.GetReasonPhrase())
	}
	return bs, nil
}
