package app

import (
	"bookget/config"
	"bookget/model/korea"
	"bookget/pkg/downloader"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

type Korea struct {
	dt *DownloadTask
	dm *downloader.DownloadManager
}

func NewKorea() *Korea {
	ctx, cancel := context.WithCancel(context.Background())
	return &Korea{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *Korea) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (r *Korea) Run(sUrl string) (msg string, err error) {
	r.dt = new(DownloadTask)
	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl
	r.dt.BookId = r.getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	r.dt.Jar, _ = cookiejar.New(nil)
	return r.download()
}

func (r *Korea) getBookId(sUrl string) (bookId string) {
	m := regexp.MustCompile(`uci=([^&]+)`).FindStringSubmatch(sUrl)
	if m != nil {
		return m[1]
	}
	return ""
}

func (r *Korea) download() (msg string, err error) {
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
		if vol.Canvases == nil {
			continue
		}
		log.Printf(" %d/%d volume, %d pages \n", i+1, sizeVol, len(vol.Canvases))
		r.dm.AddImageTasks(vol.Canvases, r.dt.SavePath, config.Conf.FileExt, 0, nil, r.dt.Jar, true)
	}
	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	return "", nil
}

func (r *Korea) getVolumes(sUrl string, jar *cookiejar.Jar) (volumes []korea.PartialCanvases, err error) {
	bs, err := getBody(sUrl, jar)
	if err != nil {
		return nil, err
	}
	matches := regexp.MustCompile(`var[\s+]bookInfos[\s+]=[\s+]([^;]+);`).FindSubmatch(bs)
	if matches == nil {
		return
	}
	resp := make([]korea.Response, 0, 100)
	if err = json.Unmarshal(matches[1], &resp); err != nil {
		return nil, err
	}
	ossHost := fmt.Sprintf("%s://%s/data/des/%s/IMG/", r.dt.UrlParsed.Scheme, r.dt.UrlParsed.Host, r.dt.BookId)
	for _, match := range resp {
		vol := korea.PartialCanvases{
			Directory: "",
			Title:     "",
			Canvases:  make([]string, 0, len(match.ImgInfos)),
		}
		for _, m := range match.ImgInfos {
			imgUrl := ossHost + m.BookPath + "/" + m.Fname
			vol.Canvases = append(vol.Canvases, imgUrl)
		}
		volumes = append(volumes, vol)
	}
	return
}
