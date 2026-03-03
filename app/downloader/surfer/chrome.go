//go:build !cover

package surfer

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"github.com/andeya/gust/result"
	"github.com/chromedp/chromedp"
)

// Chrome is a Chromium-based headless browser downloader that keeps a
// single long-lived browser process. Each request opens a new tab that
// first navigates to the target site's homepage (establishing session
// cookies and a valid Referer) before loading the actual URL. This
// two-step approach reliably bypasses JS-based security verification
// pages (e.g. Baidu CAPTCHA) that block direct URL access.
type Chrome struct {
	mu            sync.Mutex
	CookieJar     *cookiejar.Jar
	allocCtx      context.Context
	allocCancel   context.CancelFunc
	browserCtx    context.Context    // root tab – keeps the browser alive
	browserCancel context.CancelFunc // closing this shuts down the browser
	started       bool
}

func NewChrome(jar ...*cookiejar.Jar) Surfer {
	c := &Chrome{}
	if len(jar) != 0 {
		c.CookieJar = jar[0]
	} else {
		c.CookieJar, _ = cookiejar.New(nil)
	}
	return c
}

// ensureBrowser lazily starts the shared Chrome process. Must be called
// while c.mu is held.
func (c *Chrome) ensureBrowser(ua string) {
	if c.started {
		return
	}
	opts := chromeAllocatorOpts(ua)
	c.allocCtx, c.allocCancel = chromedp.NewExecAllocator(context.Background(), opts...)
	c.browserCtx, c.browserCancel = chromedp.NewContext(c.allocCtx)
	c.started = true
}

// chromeAllocatorOpts returns chromedp allocator options with
// anti-detection tweaks applied.
func chromeAllocatorOpts(ua string) []chromedp.ExecAllocatorOption {
	var opts []chromedp.ExecAllocatorOption
	for _, o := range chromedp.DefaultExecAllocatorOptions {
		opts = append(opts, o)
	}
	opts = append(opts,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("enable-automation", false),
		chromedp.WindowSize(1920, 1080),
	)
	if ua != "" {
		opts = append(opts, chromedp.UserAgent(ua))
	}
	return opts
}

// hideWebdriver removes the navigator.webdriver flag so that anti-bot
// scripts cannot detect headless automation.
func hideWebdriver() chromedp.Action {
	var res interface{}
	return chromedp.Evaluate(`Object.defineProperty(navigator, 'webdriver', {get: () => undefined})`, &res)
}

func (c *Chrome) Download(req Request) (r result.Result[*http.Response]) {
	defer r.Catch()

	param := NewParam(req).Unwrap()

	c.mu.Lock()
	c.ensureBrowser(param.header.Get("User-Agent"))
	c.mu.Unlock()

	timeout := req.GetConnTimeout()
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	// Open a new tab inside the shared browser; cookies are shared
	// across tabs within the same browser context.
	tabCtx, tabCancel := chromedp.NewContext(c.browserCtx)
	defer tabCancel()

	tabCtx, timeoutCancel := context.WithTimeout(tabCtx, timeout)
	defer timeoutCancel()

	retries := req.GetTryTimes()
	if retries <= 0 {
		retries = 1
	}

	var body string
	var err error
	for i := 0; i < retries; i++ {
		if i != 0 {
			time.Sleep(req.GetRetryPause())
		}

		body, err = tryDownload(tabCtx, req.GetURL())
		if err != nil {
			log.Printf("[W] Chrome attempt %d/%d for %s: %v", i+1, retries, req.GetURL(), err)
			continue
		}
		break
	}

	resp := &http.Response{
		Request: &http.Request{},
		Header:  make(http.Header),
	}
	resp.Request.Method = strings.ToUpper(req.GetMethod())
	resp.Request.Header = param.header
	resp.Request.URL = param.url
	resp.Request.Host = param.url.Host

	if err != nil {
		resp.StatusCode = http.StatusBadGateway
		resp.Status = err.Error()
		resp.Body = io.NopCloser(strings.NewReader(""))
	} else {
		resp.StatusCode = http.StatusOK
		resp.Status = http.StatusText(http.StatusOK)
		resp.Body = io.NopCloser(strings.NewReader(body))
	}

	return result.Ok(resp)
}

// tryDownload navigates to the target URL and returns the HTML.
//
// Every request follows a "homepage-first" pattern within the same tab:
//  1. Navigate to the site homepage — this establishes session cookies,
//     runs any JS fingerprinting, and sets Referer for the next hop.
//  2. Navigate to the actual target URL — the site sees a natural
//     browsing flow (homepage → subpage) rather than a bot hitting a
//     deep link directly.
//
// If verification is still detected after this two-step flow, the
// function returns an error so the framework can retry later.
func tryDownload(ctx context.Context, targetURL string) (string, error) {
		homepage := ExtractHomepage(targetURL)

	// Step 1: visit the homepage first to look like a real user.
	if homepage != "" && homepage != targetURL {
		if err := chromedp.Run(ctx,
			hideWebdriver(),
			chromedp.Navigate(homepage),
			chromedp.WaitReady("body"),
			chromedp.Sleep(1*time.Second),
		); err != nil {
			return "", err
		}
	}

	// Step 2: navigate to the actual target URL.
	if err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(3*time.Second),
	); err != nil {
		return "", err
	}

	// Check if we hit a verification page.
	if isVerificationPage(ctx) {
		// Wait a bit — some verification pages auto-redirect after
		// JS execution completes.
		waitUntilNotVerification(ctx, 10*time.Second)

		if isVerificationPage(ctx) {
			return "", fmt.Errorf("blocked by security verification at %s", targetURL)
		}
	}

	var body string
	if err := chromedp.Run(ctx, chromedp.OuterHTML("html", &body)); err != nil {
		return "", err
	}
	return body, nil
}

// waitUntilNotVerification polls the page title, returning as soon as
// the page is no longer a verification page.
func waitUntilNotVerification(ctx context.Context, maxWait time.Duration) {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		if !isVerificationPage(ctx) {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// isVerificationPage checks the current page title for known security
// verification indicators.
func isVerificationPage(ctx context.Context) bool {
	var title string
	if err := chromedp.Run(ctx, chromedp.Title(&title)); err != nil {
		return false
	}
	return strings.Contains(title, "安全验证") ||
		strings.Contains(title, "verify") ||
		strings.Contains(title, "security check")
}

