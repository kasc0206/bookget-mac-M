package app

import (
	"bookget/config"
	"bookget/model/onbdigital"
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

type OnbDigital struct {
	dt *DownloadTask
	dm *downloader.DownloadManager
}

func NewOnbDigital() *OnbDigital {
	ctx, cancel := context.WithCancel(context.Background())
	return &OnbDigital{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *OnbDigital) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (r *OnbDigital) Run(sUrl string) (msg string, err error) {
	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl
	r.dt.BookId = r.getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	r.dt.Jar, _ = cookiejar.New(nil)
	return r.download()
}

func (r *OnbDigital) getBookId(sUrl string) (bookId string) {
	if m := regexp.MustCompile(`doc=([^&]+)`).FindStringSubmatch(sUrl); m != nil {
		bookId = m[1]
	}
	return bookId
}

func (r *OnbDigital) download() (msg string, err error) {
	log.Printf("Get %s\n", r.dt.Url)
	respVolume, err := r.getVolumes(r.dt.Url, r.dt.Jar)
	if err != nil {
		fmt.Println(err)
		return "getVolumes", err
	}
	r.dt.SavePath = config.Conf.Directory
	for i, vol := range respVolume {
		if !config.VolumeRange(i) {
			continue
		}
		canvases, err := r.getCanvases(vol, r.dt.Jar)
		if err != nil || canvases == nil {
			fmt.Println(err)
			continue
		}
		log.Printf(" %d/%d volume, %d pages \n", i+1, len(respVolume), len(canvases))
		r.dm.AddImageTasks(canvases, r.dt.SavePath, config.Conf.FileExt, 0, nil, r.dt.Jar, true)
	}
	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	return msg, err
}

func (r *OnbDigital) getVolumes(sUrl string, jar *cookiejar.Jar) (volumes []string, err error) {
	//刷新cookie
	_, err = r.getBody(sUrl, jar)
	if err != nil {
		return
	}
	volumes = append(volumes, sUrl)
	return volumes, nil
}

func (r *OnbDigital) getCanvases(sUrl string, jar *cookiejar.Jar) (canvases []string, err error) {
	apiUrl := "https://" + r.dt.UrlParsed.Host + "/OnbViewer/service/viewer/imageData?doc=" + r.dt.BookId + "&from=1&to=3000"
	bs, err := r.getBody(apiUrl, jar)
	if err != nil {
		return
	}
	var result = new(onbdigital.Response)
	if err = json.Unmarshal(bs, result); err != nil {
		log.Printf("json.Unmarshal failed: %s\n", err)
		return
	}
	serverUrl := "https://" + r.dt.UrlParsed.Host + "/OnbViewer/image?"
	for _, m := range result.ImageData {
		imgUrl := serverUrl + m.QueryArgs + "&w=2400&q=70"
		canvases = append(canvases, imgUrl)
	}
	return canvases, err
}

func (r *OnbDigital) getBody(sUrl string, jar *cookiejar.Jar) ([]byte, error) {
	ctx := context.Background()
	cli := gohttp.NewClient(ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  jar,
		Headers: map[string]interface{}{
			"User-Agent": config.Conf.UserAgent,
		},
	})
	resp, err := cli.Get(sUrl)
	if err != nil {
		return nil, err
	}
	bs, _ := resp.GetBody()
	if bs == nil {
		return nil, fmt.Errorf("ErrCode:%d, %s", resp.GetStatusCode(), resp.GetReasonPhrase())
	}
	return bs, nil
}

func (r *OnbDigital) postBody(_ string, _ []byte) ([]byte, error) {
	return nil, fmt.Errorf("postBody not implemented for OnbDigital")
}
