package app

import (
	"bookget/config"
	"bookget/model/wzlib"
	"bookget/pkg/downloader"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

type Wzlib struct {
	dt       *DownloadTask
	dm       *downloader.DownloadManager
	pdfNames map[string]string
}

func NewWzlib() *Wzlib {
	ctx, cancel := context.WithCancel(context.Background())
	return &Wzlib{
		dt:       new(DownloadTask),
		dm:       downloader.NewDownloadManager(ctx, cancel, config.Conf.MaxConcurrent),
		pdfNames: make(map[string]string),
	}
}

func (r *Wzlib) GetRouterInit(sUrl string) (map[string]interface{}, error) {
	msg, err := r.Run(sUrl)
	return map[string]interface{}{
		"url": sUrl,
		"msg": msg,
	}, err
}

func (p *Wzlib) Run(sUrl string) (msg string, err error) {
	p.dt.UrlParsed, err = url.Parse(sUrl)
	p.dt.Url = sUrl
	p.dt.BookId = p.getBookId(p.dt.Url)
	if p.dt.BookId == "" {
		return "requested URL was not found.", err
	}
	p.dt.Jar, _ = cookiejar.New(nil)
	return p.download()
}

func (p *Wzlib) getBookId(sUrl string) (bookId string) {
	m := regexp.MustCompile(`(?i)id=([A-z0-9_-]+)`).FindStringSubmatch(sUrl)
	//m := regexp.MustCompile(`\?id=([A-z\d]+)`).FindStringSubmatch(sUrl)
	if m != nil {
		bookId = m[1]
	}
	return bookId
}

func (p *Wzlib) download() (msg string, err error) {
	log.Printf("Get %s\n", p.dt.Url)
	p.dt.SavePath = config.Conf.Directory

	//旧版：瓯越记忆
	if p.dt.UrlParsed.Host == "oyjy.wzlib.cn" {
		canvases, err := p.OyjyGetCanvases(p.dt.BookId)
		if err != nil || canvases == nil {
			fmt.Println(err)
		}
		return p.do(canvases)
	}
	//新版温州图书馆
	canvases, err := p.getCanvases(p.dt.Url, p.dt.Jar)
	if err != nil || canvases == nil {
		fmt.Println(err)
	}
	return p.do(canvases)
}

func (p *Wzlib) do(dUrls []string) (msg string, err error) {
	if dUrls == nil {
		return "", nil
	}
	size := len(dUrls)
	log.Printf(" %d PDFs.\n", size)
	for i, uri := range dUrls {
		if !config.PageRange(i, size) {
			continue
		}
		if uri == "" {
			continue
		}
		log.Printf("Get %d/%d, URL: %s\n", i+1, size, uri)
		sortId := fmt.Sprintf("%04d", i+1)
		filename := p.pdfNames[uri]
		if filename == "" {
			filename = BuildOutputFileName(".pdf", p.dt.Title, sortId)
		}
		p.dm.AddFromLegacy(uri, "GET", nil, nil, p.dt.SavePath, filename, config.Conf.Threads, p.dt.Jar, true)
	}
	if len(p.dm.Tasks()) > 0 {
		p.dm.Start()
	}
	return "", nil
}

func (p *Wzlib) getVolumes(_ string, _ *cookiejar.Jar) (volumes []string, err error) {
	return nil, fmt.Errorf("getVolumes not implemented for Wzlib")
}

func (p *Wzlib) getCanvases(sUrl string, jar *cookiejar.Jar) (canvases []string, err error) {
	apiUrl := fmt.Sprintf("https://%s/search/juhe_detail/%s/true?Flag=s", p.dt.UrlParsed.Host, p.dt.BookId)
	bs, err := getBody(apiUrl, jar)
	if err != nil {
		return
	}

	var resT = new(wzlib.Digital)
	if err = json.Unmarshal(bs, &resT); err != nil {
		log.Printf("json.Unmarshal failed: %s\n", err)
		return
	}
	p.dt.Title = NormalizeNamePart(resT.Title)
	for _, ret := range resT.DigitalResourceData {
		m := regexp.MustCompile(`file=(\S+)`).FindStringSubmatch(ret.Url)
		if m == nil {
			continue
		}
		pdfUrl := "https://db.wzlib.cn" + m[1]
		p.pdfNames[pdfUrl] = BuildOutputFileName(".pdf", p.dt.Title, ret.Title)
		canvases = append(canvases, pdfUrl)
	}
	return canvases, nil
}

func (p *Wzlib) OyjyGetCanvases(bookId string) (canvases []string, err error) {
	//一册
	uri := fmt.Sprintf("https://oyjy.wzlib.cn/api/search/v1/resource/%s", bookId)
	bs, err := getBody(uri, p.dt.Jar)
	if err == nil {
		var result wzlib.ResultPdf
		if err = json.Unmarshal(bs, &result); err == nil {
			p.dt.Title = NormalizeNamePart(result.Data.DcTitle)
			m := regexp.MustCompile(`file=(\S+)`).FindStringSubmatch(result.Data.WzlPdfUrl)
			if m != nil {
				pdfUrl := "https://db.wzlib.cn" + m[1]
				p.pdfNames[pdfUrl] = BuildOutputFileName(".pdf", result.Data.DcTitle, result.Data.RelateName)
				canvases = append(canvases, pdfUrl)
				return canvases, err
			}
		}
	}

	//多册
	relatedUri := fmt.Sprintf("https://oyjy.wzlib.cn/api/search/v1/resource_related/%s", bookId)
	bs, err = getBody(relatedUri, p.dt.Jar)
	if err != nil {
		return
	}
	var result wzlib.Result
	if err = json.Unmarshal(bs, &result); err != nil {
		return
	}
	if len(result) > 0 {
		p.dt.Title = NormalizeNamePart(result[0].Title)
	}
	for _, v := range result[0].Items {
		if v.WzlPdfUrl == "" {
			continue
		}
		pdfUrl := "https://db.wzlib.cn" + v.WzlPdfUrl
		p.pdfNames[pdfUrl] = BuildOutputFileName(".pdf", result[0].Title, v.DcTitle)
		canvases = append(canvases, pdfUrl)
	}
	return canvases, err
}
