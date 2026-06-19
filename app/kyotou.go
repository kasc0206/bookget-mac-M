package app

import (
	"bookget/config"
	"bookget/pkg/downloader"
	"bookget/pkg/gohttp"
	"context"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Kyotou struct {
	dt *DownloadTask
	dm *downloader.DownloadManager
}

func NewKyotou() *Kyotou {
	ctx, cancel := context.WithCancel(context.Background())
	return &Kyotou{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *Kyotou) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (r *Kyotou) Run(sUrl string) (msg string, err error) {
	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl
	r.dt.Jar, _ = cookiejar.New(nil)
	r.dt.BookId = r.getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	return r.download()
}

func (r *Kyotou) getBookId(sUrl string) (bookId string) {
	if strings.Contains(sUrl, "menu") {
		return getBookId(sUrl)
	}
	return ""
}

func (r *Kyotou) download() (msg string, err error) {
	log.Printf("Get %s\n", r.dt.Url)

	respVolume, err := r.getVolumes(r.dt.Url, r.dt.Jar)
	if err != nil {
		fmt.Println(err)
		return "getVolumes", err
	}
	sizeVol := len(respVolume)
	for i, vol := range respVolume {
		if !config.VolumeRange(i) {
			continue
		}
		if sizeVol == 1 {
			r.dt.SavePath = config.Conf.Directory
		} else {
			vid := fmt.Sprintf("%04d", i+1)
			r.dt.SavePath = CreateDirectory(vid)
		}
		canvases, err := r.getCanvases(vol, r.dt.Jar)
		if err != nil || canvases == nil {
			fmt.Println(err)
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



func (r *Kyotou) getVolumes(sUrl string, jar *cookiejar.Jar) (volumes []string, err error) {
	bs, err := r.getBody(sUrl, jar)
	if err != nil {
		return
	}
	//取册数
	matches := regexp.MustCompile(`href=["']?(.+?)\.html["']?`).FindAllSubmatch(bs, -1)
	if matches == nil {
		return
	}
	pos := strings.LastIndex(sUrl, "/")
	hostUrl := sUrl[:pos]
	volumes = make([]string, 0, len(matches))
	for _, v := range matches {
		text := string(v[1])
		if strings.Contains(text, "top") {
			continue
		}
		linkUrl := fmt.Sprintf("%s/%s.html", hostUrl, text)
		volumes = append(volumes, linkUrl)
	}
	return volumes, err
}

func (r *Kyotou) getCanvases(sUrl string, jar *cookiejar.Jar) (canvases []string, err error) {
	bs, err := r.getBody(sUrl, jar)
	if err != nil {
		return
	}
	startPos, ok := r.getVolStartPos(bs)
	if !ok {
		return
	}
	maxPage, ok := r.getVolMaxPage(bs)
	if !ok {
		return
	}
	bookNumber, ok := r.getBookNumber(bs)
	if !ok {
		return
	}
	pos := strings.LastIndex(sUrl, "/")
	pos1 := strings.LastIndex(sUrl[:pos], "/")
	hostUrl := sUrl[:pos1]
	maxPos := startPos + maxPage
	for i := 1; i < maxPos; i++ {
		sortId := fmt.Sprintf("%04d", i)
		imgUrl := fmt.Sprintf("%s/L/%s%s.jpg", hostUrl, bookNumber, sortId)
		canvases = append(canvases, imgUrl)
	}
	return canvases, err
}

func (r *Kyotou) getBody(sUrl string, jar *cookiejar.Jar) ([]byte, error) {
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
	if resp.GetStatusCode() != 200 || bs == nil {
		return nil, &HTTPError{
			StatusCode: resp.GetStatusCode(),
			URL:        sUrl,
			Message:    resp.GetReasonPhrase(),
		}
	}
	return bs, nil
}

func (r *Kyotou) postBody(_ string, _ []byte) ([]byte, error) {
	return nil, fmt.Errorf("postBody not implemented for Kyotou")
}

func (r *Kyotou) getBookNumber(bs []byte) (bookNumber string, ok bool) {
	//当前开始位置
	match := regexp.MustCompile(`var[\s]+bookNum[\s]+=["'\s]*([A-z0-9]+)["'\s]*;`).FindStringSubmatch(string(bs))
	if match == nil {
		return "", false
	}
	return match[1], true
}

func (r *Kyotou) getVolStartPos(bs []byte) (startPos int, ok bool) {
	//当前开始位置
	match := regexp.MustCompile(`var[\s]+volStartPos[\s]*=[\s]*([0-9]+)[\s]*;`).FindStringSubmatch(string(bs))
	if match == nil {
		return 0, false
	}
	startPos, _ = strconv.Atoi(match[1])
	return startPos, true
}

func (r *Kyotou) getVolCurPage(bs []byte) (curPage int, ok bool) {
	//当前开始位置
	match := regexp.MustCompile(`var[\s]+curPage[\s]*=[\s]*([0-9]+)[\s]*;`).FindStringSubmatch(string(bs))
	if match == nil {
		return 0, false
	}
	curPage, _ = strconv.Atoi(match[1])
	return curPage, true
}

func (r *Kyotou) getVolMaxPage(bs []byte) (maxPage int, ok bool) {
	//当前开始位置
	match := regexp.MustCompile(`var[\s]+volMaxPage[\s]*=[\s]*([0-9]+)[\s]*;`).FindStringSubmatch(string(bs))
	if match == nil {
		return 0, false
	}
	maxPage, _ = strconv.Atoi(match[1])
	return maxPage, true
}
