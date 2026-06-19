package app

import (
	"bookget/config"
	"bookget/pkg/downloader"
	"bookget/pkg/util"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"path"
	"regexp"
	"strconv"
)

type ZhuCheng struct {
	dt *DownloadTask
	dm *downloader.DownloadManager
}

func NewZhuCheng() *ZhuCheng {
	ctx, cancel := context.WithCancel(context.Background())
	return &ZhuCheng{
		dt: new(DownloadTask),
		dm: downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
	}
}

func (r *ZhuCheng) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (r *ZhuCheng) Run(sUrl string) (msg string, err error) {
	r.dt.UrlParsed, err = url.Parse(sUrl)
	r.dt.Url = sUrl
	r.dt.BookId = r.getBookId(r.dt.Url)
	if r.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	r.dt.Jar, _ = cookiejar.New(nil)
	return r.download()
}

func (r *ZhuCheng) getBookId(sUrl string) (bookId string) {
	if m := regexp.MustCompile(`&id=(\d+)`).FindStringSubmatch(sUrl); m != nil {
		bookId = m[1]
	}
	return bookId
}

func (r *ZhuCheng) download() (msg string, err error) {
	log.Printf("Get %s\n", r.dt.Url)
	respVolume, err := r.getVolumes(r.dt.BookId, r.dt.Jar)
	if err != nil {
		fmt.Println(err)
		return "getVolumes", err
	}
	r.dt.SavePath = config.Conf.Directory
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
		r.do(canvases)
	}
	return msg, err
}

func (r *ZhuCheng) do(imgUrls []string) (msg string, err error) {
	if imgUrls == nil {
		return
	}
	size := len(imgUrls)
	for i, uri := range imgUrls {
		if uri == "" || !config.PageRange(i, size) {
			continue
		}
		ext := util.FileExt(uri)
		sortId := fmt.Sprintf("%04d", i+1)
		filename := sortId + ext
		if FileExist(path.Join(r.dt.SavePath, filename)) {
			continue
		}
		imgUrl := uri
		log.Printf("Get %d/%d  %s\n", i+1, size, imgUrl)
		headers := map[string]string{
			"User-Agent": config.Conf.UserAgent,
		}
		r.dm.AddFromLegacy(imgUrl, "GET", headers, nil, r.dt.SavePath, filename, 1, r.dt.Jar, true)
	}
	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	return "", nil
}

func (r *ZhuCheng) getVolumes(bookId string, jar *cookiejar.Jar) (volumes []string, err error) {
	hostUrl := r.dt.UrlParsed.Scheme + "://" + r.dt.UrlParsed.Host
	apiUrl := hostUrl + "/index.php?ac=catalog&id=" + bookId
	bs, err := getBody(apiUrl, jar)
	if err != nil {
		return
	}

	//取册数
	matches := regexp.MustCompile(`href="./reader.php([^"]+?)"`).FindAllStringSubmatch(string(bs), -1)
	if matches == nil {
		return
	}
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		ids = append(ids, match[1])
	}
	volumes = make([]string, 0, len(ids))
	for _, v := range ids {
		sUrl := hostUrl + "/reader.php" + v
		volumes = append(volumes, sUrl)
	}
	return volumes, nil
}

func (r *ZhuCheng) getCanvases(sUrl string, jar *cookiejar.Jar) (canvases []string, err error) {
	bs, err := getBody(sUrl, jar)
	if err != nil {
		return
	}
	bid, err := r.getBid(bs)
	if err != nil {
		return nil, err
	}
	cid, err := r.getCID(bs)
	if err != nil {
		return nil, err
	}
	ext, err := r.getImgType(bs)
	if err != nil {
		return nil, err
	}
	pageSize, err := r.getPageNumber(bs)
	if err != nil {
		return nil, err
	}
	hostUrl := r.dt.UrlParsed.Scheme + "://" + r.dt.UrlParsed.Host + "/images/book/" + bid + "/" + cid + "/"
	for i := 1; i <= pageSize; i++ {
		imgUrl := hostUrl + fmt.Sprintf("%d", i) + ext
		canvases = append(canvases, imgUrl)
	}
	return canvases, err
}

func (r *ZhuCheng) getBody(_ string, _ *cookiejar.Jar) ([]byte, error) {
	return nil, fmt.Errorf("getBody not implemented for ZhuCheng")
}

func (r *ZhuCheng) postBody(_ string, _ []byte) ([]byte, error) {
	return nil, fmt.Errorf("postBody not implemented for ZhuCheng")
}

func (r *ZhuCheng) getBid(bs []byte) (string, error) {
	match := regexp.MustCompile(`var\s+BID\s+=\s+'([A-z0-9]+)'`).FindSubmatch(bs)
	if match != nil {
		return string(match[1]), nil
	}
	return "", errors.New("not found bid")
}

func (r *ZhuCheng) getCID(bs []byte) (string, error) {
	match := regexp.MustCompile(`var\s+CID\s+=\s+'([A-z0-9]+)'`).FindSubmatch(bs)
	if match != nil {
		return string(match[1]), nil
	}
	return "", errors.New("not found cid")
}
func (r *ZhuCheng) getImgType(bs []byte) (string, error) {
	match := regexp.MustCompile(`var\s+imgtype\s+=\s+'([A-z.]+)'`).FindSubmatch(bs)
	if match != nil {
		return string(match[1]), nil
	}
	return "", errors.New("not found ImgType")
}

func (r *ZhuCheng) getPageNumber(bs []byte) (int, error) {
	match := regexp.MustCompile(`var\s+PAGES\s+=\s+([0-9]+)`).FindSubmatch(bs)
	if match != nil {
		size, _ := strconv.Atoi(string(match[1]))
		return size, nil
	}
	return 0, errors.New("not found PAGES")
}
