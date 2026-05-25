package crawler

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/goclone-dev/goclone/pkg/file"
	"github.com/goclone-dev/goclone/testutils"
	"github.com/gocolly/colly/v2"
)

var TsUrl string

func collectAndGetFileContent(tsUrl, projectDirectory, relativeRoute string, assets bool) string {
	client := &http.Client{Transport: http.DefaultTransport}
	Collector(context.Background(), tsUrl, projectDirectory, client, nil, "", 1, assets)
	route := projectDirectory + relativeRoute
	fileContent := file.GetFileContent(route)
	return fileContent
}

var collectorTests = map[string]func(*testing.T){
	"indexDownload": func(t *testing.T) {
		projectDirectory := file.CreateProject("test")
		collectorContent := collectAndGetFileContent(TsUrl+"/hello", projectDirectory, "/index.html", false)

		if collectorContent != testutils.CrawlerHelloContent {
			t.Fatalf("Expect \"%s\", but got: %s", testutils.CrawlerHelloContent, collectorContent)
		}
		os.RemoveAll(projectDirectory)
	},
	"cssDownload": func(t *testing.T) {
		projectDirectory := file.CreateProject("test")
		cssFileContent := collectAndGetFileContent(TsUrl, projectDirectory, "/css/index.css", true)
		if cssFileContent != testutils.CrawlerCssContent {
			t.Fatalf("Expect \"%s\", but got: %s", testutils.CrawlerCssContent, cssFileContent)
		}
		os.RemoveAll(projectDirectory)
	},
	"jsDownload": func(t *testing.T) {
		projectDirectory := file.CreateProject("test")
		jsFileContent := collectAndGetFileContent(TsUrl, projectDirectory, "/js/index.js", false)

		if jsFileContent != testutils.CrawlerJsContent {
			t.Fatalf("Expect \"%s\", but got: %s", testutils.CrawlerJsContent, jsFileContent)
		}
		os.RemoveAll(projectDirectory)
	},
	"imgDownload": func(t *testing.T) {
		projectDirectory := file.CreateProject("test")
		imgFileContent := collectAndGetFileContent(TsUrl, projectDirectory, "/imgs/image.png", true)
		if imgFileContent != testutils.CrawlerImgContent {
			t.Fatalf("Expect \"%s\", but got: %s", testutils.CrawlerImgContent, imgFileContent)
		}
		os.RemoveAll(projectDirectory)
	},
}

func TestSetUpCollector(t *testing.T) {
	testutils.SilenceStdoutInTests()
	c := colly.NewCollector(colly.Async(true))
	userAgent := "Firefox"
	setUpCollector(c, nil, nil, "http://127.0.0.1:9999", "Firefox")

	if c.UserAgent != userAgent {
		t.Fatalf("Expect %s, but got: %s", userAgent, c.UserAgent)
	}
}
func TestCollectorTests(t *testing.T) {
	testutils.SilenceStdoutInTests()
	ts := testutils.NewCrawlerTestServer()
	defer ts.Close()
	TsUrl = ts.URL
	for testName, testFuntion := range collectorTests {
		t.Run(testName, testFuntion)
	}
}
