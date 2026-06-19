package app

import (
	"bookget/config"
	"bookget/model/usthk"
	"bookget/pkg/downloader"
	"bookget/pkg/gohttp"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

type Usthk struct {
	dt *DownloadTask
	dm *downloader.DownloadManager
}

func NewUsthk() *Usthk {
	ctx, cancel := context.WithCancel(context.Background())
	return &Usthk{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *Usthk) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (r *Usthk) Run(sUrl string) (msg string, err error) {
	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl
	r.dt.BookId = r.getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	r.dt.Jar, _ = cookiejar.New(nil)
	return r.download()
}

func (r *Usthk) getBookId(sUrl string) (bookId string) {
	m := regexp.MustCompile(`bib/([A-z0-9_-]+)`).FindStringSubmatch(sUrl)
	if m != nil {
		bookId = m[1]
	}
	return bookId
}

func (r *Usthk) download() (msg string, err error) {
	log.Printf("Get %s\n", r.dt.Url)

	respVolume, err := r.getVolumes(r.dt.Url)
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

		canvases, err := r.getCanvases(vol)
		if err != nil || canvases == nil {
			fmt.Println(err)
			continue
		}
		fmt.Println()
		log.Printf(" %d/%d volume, %d pages \n", i+1, sizeVol, len(canvases))
		r.dm.AddImageTasks(canvases, r.dt.SavePath, config.Conf.FileExt, 0, nil, r.dt.Jar, true)
	}
	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	return "", nil
}

func (r *Usthk) getVolumes(sUrl string) (volumes []string, err error) {
	volumes = append(volumes, sUrl)
	return volumes, nil
}

func (r *Usthk) getCanvases(sUrl string) ([]string, error) {
	bs, err := r.getBody(sUrl)
	if err != nil {
		return nil, err
	}
	//view_book('6/o/b1129168/ebook'
	matches := regexp.MustCompile(`view_book\(["'](\S+)["']`).FindAllStringSubmatch(string(bs), -1)
	if matches == nil {
		return nil, fmt.Errorf("Canvas not found")
	}

	canvases := make([]string, 0, len(matches))
	for _, m := range matches {
		sPath := m[1]
		apiUrl := fmt.Sprintf("https://%s/bookreader/getfilelist.php?path=%s", r.dt.UrlParsed.Host, sPath)
		bs, err = r.getBody(apiUrl)
		if err != nil {
			break
		}
		respFiles := new(usthk.Response)
		if err = json.Unmarshal(bs, respFiles); err != nil {
			log.Printf("json.Unmarshal failed: %s\n", err)
			break
		}
		//imgUrls := make([]string, 0, len(result.FileList))
		for _, v := range respFiles.FileList {
			imgUrl := fmt.Sprintf("https://%s/obj/%s/%s", r.dt.UrlParsed.Host, sPath, v)
			canvases = append(canvases, imgUrl)
		}
	}
	return canvases, nil
}

func (r *Usthk) getBody(sUrl string) ([]byte, error) {
	ctx := context.Background()
	cli := gohttp.NewClient(ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  r.dt.Jar,
		Headers: map[string]interface{}{
			"User-Agent": config.Conf.UserAgent,
		},
	})
	resp, err := cli.Get(sUrl)
	if err != nil {
		return nil, err
	}
	bs, _ := resp.GetBody()
	if resp.GetStatusCode() != 200 || bs == nil {
		return nil, fmt.Errorf("ErrCode:%d, %s", resp.GetStatusCode(), resp.GetReasonPhrase())
	}
	return bs, nil
}
