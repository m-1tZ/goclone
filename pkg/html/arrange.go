package html

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/m-1tZ/goclone/pkg/parser"
)

// LinkRestructure rewrites asset paths in index.html to point at the locally
// saved copies.  baseURL resolves relative hrefs/srcs.  assets controls
// whether CSS/image/font paths are rewritten (only meaningful when --assets
// was also passed to the crawler).
func LinkRestructure(projectDir string, baseURL string, assets bool) error {
	indexfile := filepath.Join(projectDir, "index.html")
	input, err := os.ReadFile(indexfile)
	if err != nil {
		return err
	}

	base, _ := url.Parse(baseURL)

	// resolve turns a potentially-relative href/src into an absolute URL so
	// URLToLocalFilename always operates on a consistent full path.
	resolve := func(href string) string {
		if base == nil {
			return href
		}
		ref, err := url.Parse(href)
		if err != nil {
			return href
		}
		return base.ResolveReference(ref).String()
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(input))
	if err != nil {
		return err
	}

	// Always rewrite: external JS.
	doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
		if data, exists := s.Attr("src"); exists {
			s.SetAttr("src", "js/"+parser.URLToLocalFilename(resolve(data)))
		}
	})

	// Always rewrite: manifest link.
	doc.Find("link[rel='manifest']").Each(func(i int, s *goquery.Selection) {
		if data, exists := s.Attr("href"); exists {
			s.SetAttr("href", "data/"+parser.URLToLocalFilename(resolve(data)))
		}
	})

	// Assets mode only: CSS, images, fonts, favicons.
	if assets {
		doc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
			if data, exists := s.Attr("href"); exists {
				s.SetAttr("href", "css/"+parser.URLToLocalFilename(resolve(data)))
			}
		})

		doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
			if data, exists := s.Attr("src"); exists && !strings.HasPrefix(data, "data:") {
				s.SetAttr("src", "imgs/"+parser.URLToLocalFilename(resolve(data)))
			}
		})

		doc.Find("link[rel~='icon'], link[rel='apple-touch-icon']").Each(func(i int, s *goquery.Selection) {
			if data, exists := s.Attr("href"); exists {
				s.SetAttr("href", "imgs/"+parser.URLToLocalFilename(resolve(data)))
			}
		})

		doc.Find("link[rel='preload'][as='font']").Each(func(i int, s *goquery.Selection) {
			if data, exists := s.Attr("href"); exists {
				s.SetAttr("href", "fonts/"+parser.URLToLocalFilename(resolve(data)))
			}
		})
	}

	html, err := doc.Html()
	if err != nil {
		return err
	}

	return os.WriteFile(indexfile, []byte(html), 0666)
}
