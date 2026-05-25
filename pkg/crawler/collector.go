package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/gocolly/colly/v2"
)

// Collector crawls siteURL and saves all captured resources under projectPath.
//
// Security mode (assets=false, the default): downloads HTML + JS + inline
// scripts + source maps + JSON manifests — everything useful for code review.
//
// Assets mode (assets=true, --assets flag): additionally downloads CSS, fonts,
// and images for a visually complete clone.
//
// depth controls how many link levels to follow via <a href>
// (1 = main page only, 0 = unlimited).
func Collector(ctx context.Context, siteURL string, projectPath string, client *http.Client, jar *cookiejar.Jar, userAgent string, depth int, assets bool) error {
	if client == nil {
		client = &http.Client{Transport: http.DefaultTransport}
	}

	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(depth),
	)

	c.WithTransport(client.Transport)
	if jar != nil {
		c.SetCookieJar(jar)
	}
	if userAgent != "" {
		c.UserAgent = userAgent
	}

	// Save the seed page as index.html.
	c.OnResponse(func(r *colly.Response) {
		if r.Request.Depth != 1 {
			return
		}
		dest := filepath.Join(projectPath, "index.html")
		if err := os.WriteFile(dest, r.Body, 0666); err != nil {
			fmt.Printf("warning: failed to save index.html: %v\n", err)
		}
	})

	// External JS (always, security mode core).
	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		fmt.Println("JS found -->", link)
		if err := Extractor(client, e.Request.AbsoluteURL(link), projectPath, assets); err != nil {
			fmt.Printf("warning: %v\n", err)
		}
	})

	// Inline scripts — saved as js/inline_N.js.
	// These often contain API keys, endpoints, and config that external JS files do not.
	var inlineCount int64
	c.OnHTML("script:not([src])", func(e *colly.HTMLElement) {
		content := strings.TrimSpace(e.Text)
		if content == "" {
			return
		}
		n := atomic.AddInt64(&inlineCount, 1)
		name := fmt.Sprintf("inline_%d.js", n)
		dest := filepath.Join(projectPath, "js", name)
		if err := os.WriteFile(dest, []byte(content), 0666); err != nil {
			fmt.Printf("warning: failed to save inline script %s: %v\n", name, err)
			return
		}
		fmt.Printf("Inline script saved --> %s\n", name)
	})

	// PWA manifest and other linked data files (always, security mode core).
	c.OnHTML("link[rel='manifest']", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if link == "" {
			return
		}
		fmt.Println("Manifest found -->", link)
		if err := Extractor(client, e.Request.AbsoluteURL(link), projectPath, assets); err != nil {
			fmt.Printf("warning: %v\n", err)
		}
	})

	// Assets mode: CSS, images, fonts, favicons.
	if assets {
		c.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
			link := e.Attr("href")
			fmt.Println("CSS found -->", link)
			if err := Extractor(client, e.Request.AbsoluteURL(link), projectPath, true); err != nil {
				fmt.Printf("warning: %v\n", err)
			}
		})

		c.OnHTML("img[src]", func(e *colly.HTMLElement) {
			link := e.Attr("src")
			if strings.HasPrefix(link, "data:") || strings.HasPrefix(link, "blob:") {
				return
			}
			fmt.Println("Img found -->", link)
			if err := Extractor(client, e.Request.AbsoluteURL(link), projectPath, true); err != nil {
				fmt.Printf("warning: %v\n", err)
			}
		})

		c.OnHTML("link[rel~='icon'], link[rel='apple-touch-icon']", func(e *colly.HTMLElement) {
			link := e.Attr("href")
			if link == "" {
				return
			}
			fmt.Println("Icon found -->", link)
			if err := Extractor(client, e.Request.AbsoluteURL(link), projectPath, true); err != nil {
				fmt.Printf("warning: %v\n", err)
			}
		})

		c.OnHTML("link[rel='preload'][as='font']", func(e *colly.HTMLElement) {
			link := e.Attr("href")
			if link == "" {
				return
			}
			fmt.Println("Font found -->", link)
			if err := Extractor(client, e.Request.AbsoluteURL(link), projectPath, true); err != nil {
				fmt.Printf("warning: %v\n", err)
			}
		})
	}

	// Follow <a href> links; MaxDepth caps recursion.
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		_ = e.Request.Visit(e.Request.AbsoluteURL(e.Attr("href")))
	})

	if err := c.Visit(siteURL); err != nil {
		return err
	}
	c.Wait()
	return nil
}

// setUpCollector configures an existing colly.Collector with auth and network
// settings.  Kept for test compatibility.
func setUpCollector(c *colly.Collector, ctx context.Context, cookieJar *cookiejar.Jar, proxyString, userAgent string) {
	if cookieJar != nil {
		c.SetCookieJar(cookieJar)
	}
	if proxyString != "" {
		c.SetProxy(proxyString)
	} else if ctx != nil {
		c.WithTransport(cancelableTransport{ctx: ctx, transport: http.DefaultTransport})
	}
	if userAgent != "" {
		c.UserAgent = userAgent
	}
}

type cancelableTransport struct {
	ctx       context.Context
	transport http.RoundTripper
}

func (t cancelableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.ctx.Err(); err != nil {
		return nil, err
	}
	return t.transport.RoundTrip(req.WithContext(t.ctx))
}
