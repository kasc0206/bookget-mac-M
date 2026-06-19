package chttp

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)



func ReadHttpCookiesFromFile(cookieFile string) ([]http.Cookie, error) {
	fp, err := os.Open(cookieFile)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	bsHeader, err := io.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	mHeader := strings.Split(string(bsHeader), "\n")
	cookies := make([]http.Cookie, 0, len(mHeader)+1)
	for _, line := range mHeader {
		if strings.HasPrefix(line, "#") {
			continue
		}
		text := regexp.MustCompile(`\\"`).ReplaceAllString(line, "\"")
		row := strings.Split(text, "\t")
		if len(row) < 8 {
			continue
		}
		name := strings.ReplaceAll(row[5], "\"", "")
		value := strings.ReplaceAll(row[6], "\"", "")
		//expires := strings.ReplaceAll(row[4], "#HttpOnly_", "")
		cookies = append(cookies, http.Cookie{Name: name, Value: value})
	}
	return cookies, nil
}

func ReadCookiesFromFile(cfile string) (cookies string, err error) {
	bs, err := os.ReadFile(cfile)
	if err != nil {
		return "", err
	}
	mCookie := strings.Split(string(bs), "\n")

	for _, line := range mCookie {
		if strings.HasPrefix(line, "#") {
			continue
		}
		text := regexp.MustCompile(`\\"`).ReplaceAllString(line, "\"")
		row := strings.Split(text, "\t")
		if len(row) < 8 {
			continue
		}
		name := strings.ReplaceAll(row[5], "\"", "")
		value := strings.ReplaceAll(row[6], "\"", "")
		s := name + "=" + value + "; "
		cookies += s
	}
	return cookies, nil
}
