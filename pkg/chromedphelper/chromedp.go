// Package chromedphelper 提供基于 chromedp 的 AWS WAF Challenge 绕过功能。
// 适用于 CUHK（香港中文大学）等使用 AWS WAF Challenge 保护的站点。
package chromedphelper

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// findChromePath 从 Playwright 缓存和系统路径中查找 Chromium/Chrome 可执行文件。
func findChromePath() string {
	// 可能的 Chrome/Chromium 路径
	candidates := []string{}

	// macOS: Playwright 缓存的 headless shell
	home, _ := os.UserHomeDir()
	playwrightCache := filepath.Join(home, "Library", "Caches", "ms-playwright")
	if entries, err := os.ReadDir(playwrightCache); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			// headless shell 优先（更轻量）
			headlessDir := filepath.Join(playwrightCache, name, "chrome-headless-shell-mac-arm64")
			if info, err := os.Stat(headlessDir); err == nil && info.IsDir() {
				if entries2, err := os.ReadDir(headlessDir); err == nil {
					for _, e2 := range entries2 {
						if e2.Name() == "chrome-headless-shell" {
							candidates = append(candidates, filepath.Join(headlessDir, "chrome-headless-shell"))
						}
					}
				}
			}
			// 完整 Chrome
			chromeMacDir := filepath.Join(playwrightCache, name, "chrome-mac-arm64")
			chromeApp := filepath.Join(chromeMacDir, "Google Chrome for Testing.app", "Contents", "MacOS", "Google Chrome for Testing")
			if _, err := os.Stat(chromeApp); err == nil {
				candidates = append(candidates, chromeApp)
			}
		}
	}

	// macOS: 系统路径
	candidates = append(candidates,
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
		"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
	)

	// Linux 路径
	candidates = append(candidates,
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
	)

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// SolveWAFChallenge 使用 headless Chrome 绕过 AWS WAF Challenge,
// 返回页面的完整 HTML 和获取到的 CookieJar。
func SolveWAFChallenge(pageURL string, timeout time.Duration) (html string, jar *cookiejar.Jar, err error) {
	chromePath := findChromePath()
	if chromePath == "" {
		return "", nil, fmt.Errorf("未找到 Chrome/Chromium，请安装 Google Chrome 或 Chromium")
	}

	log.Printf("使用 Chrome: %s", chromePath)

	// 创建临时用户数据目录
	tmpDir, err := os.MkdirTemp("", "bookget-chrome-*")
	if err != nil {
		return "", nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 配置 chromedp
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.UserDataDir(tmpDir),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("window-size", "1920,1080"),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	// 创建浏览器上下文
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置超时
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Printf("正在打开页面: %s", pageURL)

	var pageHTML string
	var cookies []*http.Cookie

	// 执行导航并等待页面加载
	err = chromedp.Run(ctx,
		chromedp.Navigate(pageURL),
		// 等待网络空闲（WAF 挑战完成后页面才会加载）
		chromedp.WaitReady("body"),
		// 额外等待确保动态内容加载完成
		chromedp.Sleep(3*time.Second),
		// 获取完整 HTML
		chromedp.OuterHTML("html", &pageHTML),
		// 获取所有 cookies
		chromedp.ActionFunc(func(ctx context.Context) error {
			cdpCookies, err := network.GetCookies().Do(ctx)
			if err != nil {
				return err
			}
			for _, c := range cdpCookies {
				cookies = append(cookies, &http.Cookie{
					Name:   c.Name,
					Value:  c.Value,
					Domain: c.Domain,
					Path:   c.Path,
				})
			}
			return nil
		}),
	)

	if err != nil {
		return "", nil, fmt.Errorf("chromedp 执行失败: %v", err)
	}

	// 创建 CookieJar
	jar, err = cookiejar.New(nil)
	if err != nil {
		return "", nil, fmt.Errorf("创建 CookieJar 失败: %v", err)
	}

	if u, err := url.Parse(pageURL); err == nil {
		jar.SetCookies(u, cookies)
	}

	log.Printf("WAF 挑战通过，获取到 %d 个 cookie", len(cookies))
	return pageHTML, jar, nil
}
