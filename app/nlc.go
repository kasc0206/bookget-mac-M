package app

import (
	"bookget/config"
	"bookget/pkg/downloader"
	"bookget/pkg/gohttp"
	"bookget/pkg/util"
	"context"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ChinaNlc struct {
	dm     *downloader.DownloadManager
	ctx    context.Context
	cancel context.CancelFunc
	client *http.Client
	jar    *cookiejar.Jar

	rawUrl    string
	parsedUrl *url.URL
	savePath  string
	bookId    string
	bookTitle string
	pubTime   string

	body         []byte
	dataType     int //0=pdf,1=pic
	aid          string
	vectorBooks  []string
	volumeTitles map[string]string
}

func NewChinaNlc() *ChinaNlc {
	ctx, cancel := context.WithCancel(context.Background())
	dm := downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent)

	client, _ := NewHttpClient()
	jar, _ := cookiejar.New(nil)

	return &ChinaNlc{
		// 初始化字段
		dm:           dm,
		client:       client,
		ctx:          ctx,
		cancel:       cancel,
		jar:          jar,
		volumeTitles: make(map[string]string),
	}
}

func (r *ChinaNlc) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	r.rawUrl = sUrl
	r.parsedUrl, _ = url.Parse(sUrl)
	msg, err := r.Run()
	return map[string]interface{}{
		"type": "dzicnlib",
		"url":  sUrl,
		"msg":  msg,
	}, err
}

func (r *ChinaNlc) Run() (msg string, err error) {
	if strings.Contains(r.rawUrl, "OutOpenBook/Open") {
		r.body, _ = r.getBody(r.rawUrl)
		r.loadMetadata(r.body)
		r.bookId = r.getBookId(string(r.body))
	} else {
		r.bookId = r.getBookId(r.rawUrl)
	}
	if r.bookId == "" {
		return "requested URL was not found.", err
	}
	return r.download()
}

func (r *ChinaNlc) getBookId(sUrl string) (bookId string) {
	var (
		// 预编译正则表达式
		identifierRegex = regexp.MustCompile(`identifier\s*=\s*["']([^"']+)["']`)
		fidRegex        = regexp.MustCompile(`fid=([A-Za-z0-9]+)`)
	)

	// 尝试第一种匹配模式
	if matches := identifierRegex.FindStringSubmatch(sUrl); matches != nil {
		return matches[1]
	}

	// 尝试第二种匹配模式
	if matches := fidRegex.FindStringSubmatch(sUrl); matches != nil {
		return matches[1]
	}

	// 默认返回空字符串
	return ""
}

func (r *ChinaNlc) loadMetadata(body []byte) {
	text := string(body)
	if r.bookTitle == "" {
		r.bookTitle = extractNlcBookTitle(text)
	}
	if r.pubTime == "" {
		r.pubTime = extractNlcPublishTime(text)
	}
	for bid, title := range extractNlcVolumeTitles(text) {
		if bid != "" && title != "" {
			r.volumeTitles[bid] = title
		}
	}
}

func extractNlcBookTitle(text string) string {
	plainText := regexp.MustCompile(`(?s)<script[^>]*>.*?</script>|<style[^>]*>.*?</style>|<[^>]+>`).ReplaceAllString(text, " ")
	plainText = normalizeNlcText(plainText)
	if idx := strings.Index(plainText, "责任者："); idx > 0 {
		prefix := strings.TrimSpace(plainText[:idx])
		for _, marker := range []string{"资源详情", "卷 册", "显示更多", "分享", "当前位置"} {
			if pos := strings.LastIndex(prefix, marker); pos >= 0 {
				prefix = strings.TrimSpace(prefix[pos+len(marker):])
			}
		}
		if prefix != "" {
			return prefix
		}
	}

	if match := regexp.MustCompile(`<title>\s*([^<>]+?)\s*</title>`).FindStringSubmatch(text); len(match) > 1 {
		title := normalizeNlcText(match[1])
		if title != "" && !strings.Contains(title, "资源详情") && !strings.Contains(title, "读者云门户") {
			return title
		}
	}
	return ""
}

func extractNlcPublishTime(text string) string {
	plainText := regexp.MustCompile(`(?s)<script[^>]*>.*?</script>|<style[^>]*>.*?</style>|<[^>]+>`).ReplaceAllString(text, " ")
	plainText = normalizeNlcText(plainText)
	if idx := strings.Index(plainText, "出版时间："); idx >= 0 {
		rest := plainText[idx+len("出版时间："):]
		markers := []string{"索取号：", "文种：", "版本：", "总册数：", "分享", "收藏", "点赞", "卷 册"}
		end := len(rest)
		for _, marker := range markers {
			if pos := strings.Index(rest, marker); pos >= 0 && pos < end {
				end = pos
			}
		}
		return normalizeNlcText(rest[:end])
	}
	return ""
}

func extractNlcVolumeTitles(text string) map[string]string {
	volumeTitles := make(map[string]string)
	matches := regexp.MustCompile(`([^<>\n]+?)\s*<a[^>]+class="a1"[^>]*href="([^"]+)OutOpenBook/([^"]+)"`).FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		bid := getURLQueryValue(match[3], "bid")
		title := normalizeNlcText(match[1])
		if title == "" || strings.Contains(title, "在线阅读") {
			continue
		}
		if bid != "" && title != "" {
			volumeTitles[bid] = title
		}
	}
	return volumeTitles
}

func normalizeNlcText(value string) string {
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "&nbsp;", " ")
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	return value
}

func getURLQueryValue(rawURL string, key string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Query().Get(key)
}

func (r *ChinaNlc) buildPDFFilename(sourceURL string, fallback string) string {
	ext := path.Ext(fallback)
	if ext == "" {
		ext = ".pdf"
	}
	stem := strings.TrimSuffix(fallback, ext)
	bid := getURLQueryValue(sourceURL, "bid")
	volumeTitle := normalizeNlcText(r.volumeTitles[bid])
	if strings.Contains(volumeTitle, "在线阅读") {
		volumeTitle = ""
	}
	bookTitle := normalizeNlcText(r.bookTitle)

	base := stem
	if bookTitle != "" {
		if volumeTitle != "" {
			if strings.Contains(bookTitle, volumeTitle) {
				base = bookTitle
			} else {
				base = bookTitle + "_" + volumeTitle
			}
		} else if stem != "" {
			base = bookTitle + "_" + stem
		} else {
			base = bookTitle
		}
	} else if volumeTitle != "" {
		base = volumeTitle
	}

	pubTime := normalizeNlcText(r.pubTime)
	if pubTime != "" {
		base += "_" + pubTime
	}

	return util.SanitizeFileName(base) + ext
}

func (r *ChinaNlc) renamePDFIfNeeded(srcPath string, sourceURL string, fallback string) error {
	if !FileExist(srcPath) {
		return nil
	}
	desiredName := r.buildPDFFilename(sourceURL, fallback)
	if desiredName == "" || filepath.Base(srcPath) == desiredName {
		return nil
	}
	finalPath, err := util.RenameFileUnique(srcPath, r.savePath, desiredName)
	if err != nil {
		return err
	}
	log.Printf("Renamed file to %s\n", filepath.Base(finalPath))
	return nil
}

func (r *ChinaNlc) download() (msg string, err error) {
	//单册PDF
	if strings.Contains(r.rawUrl, "OutOpenBook/OpenObjectBook") {
		//PDF
		r.savePath = config.Conf.Directory
		v, _ := r.identifier(r.rawUrl)
		filename := v.Get("bid") + ".pdf"
		err = r.doPdfUrl(r.rawUrl, filename)
		return "", err
	}
	//单张图
	if strings.Contains(r.rawUrl, "OutOpenBook/OpenObjectPic") {
		r.savePath = config.Conf.Directory
		canvases, err := r.getCanvases()
		if err != nil || canvases == nil {
			return "", err
		}
		log.Printf("  %d pages \n", len(canvases))
		r.do(canvases)
		return "", err
	}
	//对照阅读单册
	if strings.Contains(r.rawUrl, "OpenTwoObjectBook") {
		r.savePath = config.Conf.Directory
		v, _ := r.identifier(r.rawUrl)
		filename := v.Get("bid") + ".pdf"
		pageUrl := fmt.Sprintf("%s://%s/OutOpenBook/OpenObjectBook?aid=%s&bid=%s", r.parsedUrl.Scheme, r.parsedUrl.Host,
			v.Get("aid"), v.Get("bid"))
		if err = r.doPdfUrl(pageUrl, filename); err != nil {
			return "", err
		}

		filename = v.Get("cid") + ".pdf"
		pageUrl = fmt.Sprintf("%s://%s/OutOpenBook/OpenObjectBook?aid=%s&bid=%s", r.parsedUrl.Scheme, r.parsedUrl.Host,
			v.Get("aid"), v.Get("cid"))
		err = r.doPdfUrl(pageUrl, filename)
		return "", err
	}
	//多册/多图
	err = r.downloadForPDFs()
	if err != nil {
		fmt.Println(err)
		return "getVolumes", err
	}
	//矢量多册PDF
	r.downloadForOCR()
	return "", nil
}

func (r *ChinaNlc) do(imgUrls []string) (msg string, err error) {
	if imgUrls == nil {
		return
	}
	fmt.Println()
	referer := url.QueryEscape(r.rawUrl)
	size := len(imgUrls)
	headers := map[string]string{
		"User-Agent": config.Conf.UserAgent,
		"Referer":    referer,
	}
	for i, uri := range imgUrls {
		if uri == "" || !config.PageRange(i, size) {
			continue
		}
		sortId := fmt.Sprintf("%04d", i+1)
		filename := sortId + config.Conf.FileExt
		if FileExist(path.Join(r.savePath, filename)) {
			continue
		}
		imgUrl := uri
		log.Printf("Get %d/%d, URL: %s\n", i+1, size, imgUrl)
		r.dm.AddFromLegacy(imgUrl, "GET", headers, nil, r.savePath, filename, 1, r.jar, true)
	}
	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	fmt.Println()
	return "", nil
}

func (r *ChinaNlc) downloadForPDFs() error {
	respVolume, err := r.getVolumes()
	if err != nil {
		return err
	}
	size := len(respVolume)
	for i, vol := range respVolume {
		if !config.VolumeRange(i) {
			continue
		}
		vid := fmt.Sprintf("%04d", i+1)
		//图片
		if strings.Contains(vol, "OpenObjectPic") {
			r.dataType = 1
			r.savePath = CreateDirectory(vid)
			canvases, err := r.getCanvases()
			if err != nil || canvases == nil {
				fmt.Println(err)
				continue
			}
			log.Printf(" %d/%d volume, %d pages \n", i+1, size, len(canvases))
			r.do(canvases)
		} else {
			//PDF
			r.savePath = config.Conf.Directory
			log.Printf("Get %d/%d volume, URL: %s\n", i+1, size, vol)
			filename := vid + ".pdf"
			r.doPdfUrl(vol, filename)
		}
	}
	return nil
}

func (r *ChinaNlc) downloadForOCR() {
	if r.vectorBooks == nil {
		return
	}
	for i, vol := range r.vectorBooks {
		if !config.VolumeRange(i) {
			continue
		}
		vid := fmt.Sprintf("%04d", i+1)
		r.savePath = CreateDirectory("ocr")
		log.Printf("Get %d/%d volume, URL: %s\n", i+1, len(r.vectorBooks), vol)
		filename := vid + ".pdf"
		r.doPdfUrl(vol, filename)
	}
}

func (r *ChinaNlc) getVolumes() (volumes []string, err error) {
	r.body, err = r.getBody(r.rawUrl)
	if err != nil {
		return nil, err
	}
	r.loadMetadata(r.body)
	text := util.SubText(string(r.body), "<div id=\"multiple\"", "id=\"catalogDiv\">")
	//取册数
	aUrls := regexp.MustCompile(`<a[^>]+class="a1"[^>].+href="([^"]+)OutOpenBook/([^"]+)"`).FindAllStringSubmatch(text, -1)
	for _, uri := range aUrls {
		pageUrl := fmt.Sprintf("%s://%s%sOutOpenBook/%s", r.parsedUrl.Scheme, r.parsedUrl.Host, uri[1], uri[2])
		volumes = append(volumes, pageUrl)
	}
	//
	aid := ""
	if volumes != nil {
		match := regexp.MustCompile(`aid=([^&]+)`).FindStringSubmatch(volumes[0])
		if match != nil {
			aid = match[1]
		}
	}

	//对照阅读
	twoUrls := regexp.MustCompile(`openTwoBookNew\('([^"']+)','([^"']+)'`).FindAllStringSubmatch(text, -1)
	if twoUrls != nil && aid != "" {
		r.vectorBooks = make([]string, 0, len(twoUrls))
		for _, uri := range twoUrls {
			pageUrl := fmt.Sprintf("%s://%s/OutOpenBook/OpenObjectBook?aid=%s&bid=%s", r.parsedUrl.Scheme, r.parsedUrl.Host, aid, uri[2])
			r.vectorBooks = append(r.vectorBooks, pageUrl)
		}
	}
	return volumes, err
}

func (r *ChinaNlc) doPdfUrl(sUrl, filename string) error {
	desiredName := r.buildPDFFilename(sUrl, filename)
	desiredDest := filepath.Join(r.savePath, desiredName)
	if desiredName != "" && FileExist(desiredDest) {
		return nil
	}

	dest := path.Join(r.savePath, filename)
	if FileExist(dest) {
		if err := r.renamePDFIfNeeded(dest, sUrl, filename); err != nil {
			return err
		}
		return nil
	}
	v, err := r.identifier(sUrl)
	if err != nil {
		return err
	}
	tokenKey, timeKey, timeFlag := r.getToken(sUrl)

	//http://read.nlc.cn/menhu/OutOpenBook/getReaderNew
	//http://read.nlc.cn/menhu/OutOpenBook/getReaderRangeNew
	pdfUrl := fmt.Sprintf("%s://%s/menhu/OutOpenBook/getReaderNew?aid=%s&bid=%s&kime=%s&fime=%s",
		r.parsedUrl.Scheme, r.parsedUrl.Host, v.Get("aid"), v.Get("bid"), timeKey, timeFlag)

	headers := map[string]string{
		"User-Agent": config.Conf.UserAgent,
		"Referer":    "http://read.nlc.cn/static/webpdf/lib/WebPDFJRWorker.js",
		"Range":      "bytes=0-1",
		"myreader":   tokenKey,
	}
	r.dm.AddFromLegacy(pdfUrl, "GET", headers, nil, r.savePath, filename, 1, r.jar, false)
	if len(r.dm.Tasks()) > 0 {
		r.dm.Start()
	}
	if err := r.renamePDFIfNeeded(dest, sUrl, filename); err != nil {
		return err
	}
	util.PrintSleepTime(config.Conf.Sleep)
	fmt.Println()
	return nil
}

func (r *ChinaNlc) getCanvases() (canvases []string, err error) {
	v, err := r.identifier(r.rawUrl)
	if err != nil {
		return nil, err
	}
	bid, _ := strconv.ParseFloat(v.Get("bid"), 32)
	iBid := int(bid)
	//图片类型检测
	var pageUrl string
	switch v.Get("aid") {
	case "495", "952", "467", "1080":
		pageUrl = fmt.Sprintf("%s://%s/allSearch/openBookPic?id=%d&l_id=%s&indexName=data_%s",
			r.parsedUrl.Scheme, r.parsedUrl.Host, iBid, v.Get("lid"), v.Get("aid"))
	case "022":
		//中国记忆库图片 不用登录可以查看
		pageUrl = fmt.Sprintf("%s://%s/allSearch/openPic_noUser?id=%d&identifier=%s&indexName=data_%s",
			r.parsedUrl.Scheme, r.parsedUrl.Host, iBid, v.Get("did"), v.Get("aid"))
	default:
		pageUrl = fmt.Sprintf("%s://%s/allSearch/openPic?id=%d&identifier=%s&indexName=data_%s",
			r.parsedUrl.Scheme, r.parsedUrl.Host, iBid, v.Get("did"), v.Get("aid"))
	}
	//
	bs, err := r.getBody(pageUrl)
	if err != nil {
		return
	}
	matches := regexp.MustCompile(`<img\s+src="(http|https)://(read|mylib).nlc.cn([^"]+)"`).FindAllSubmatch(bs, -1)
	for _, m := range matches {
		imgUrl := r.parsedUrl.Scheme + "://" + r.parsedUrl.Host + string(m[3])
		canvases = append(canvases, imgUrl)
	}
	return canvases, nil
}

func (r *ChinaNlc) identifier(sUrl string) (v url.Values, err error) {
	u, err := url.Parse(sUrl)
	if err != nil {
		return
	}
	m, _ := url.ParseQuery(u.RawQuery)
	if m["aid"] == nil || m["bid"] == nil {
		return nil, errors.New("error aid/bid")
	}
	return m, nil
}

func (r *ChinaNlc) getBody(apiUrl string) ([]byte, error) {
	referer := url.QueryEscape(apiUrl)
	cli := gohttp.NewClient(r.ctx, gohttp.Options{
		CookieFile: config.Conf.CookieFile,
		CookieJar:  r.jar,
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
	if resp.GetStatusCode() != 200 || bs == nil {
		return nil, fmt.Errorf("ErrCode:%d, %s", resp.GetStatusCode(), resp.GetReasonPhrase())
	}
	return bs, nil
}

func (r *ChinaNlc) getToken(uri string) (tokenKey, timeKey, timeFlag string) {
	body, err := r.getBody(uri)
	if err != nil {
		log.Printf("Server unavailable: %s", err.Error())
		return
	}
	//<iframe id="myframe" name="myframe" src="" width="100%" height="100%" scrolling="no" frameborder="0" tokenKey="4ADAD4B379874C10864990817734A2BA" timeKey="1648363906519" timeFlag="1648363906519" sflag=""></iframe>
	params := regexp.MustCompile(`(tokenKey|timeKey|timeFlag)="([a-zA-Z0-9]+)"`).FindAllStringSubmatch(string(body), -1)
	//tokenKey := ""
	//timeKey := ""
	//timeFlag := ""
	for _, v := range params {
		switch v[1] {
		case "tokenKey":
			tokenKey = v[2]
		case "timeKey":
			timeKey = v[2]
		case "timeFlag":
			timeFlag = v[2]
		}
	}
	return
}
