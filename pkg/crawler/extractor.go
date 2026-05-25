package crawler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/m-1tZ/goclone/pkg/parser"
)

// coreExtensions are always fetched regardless of --assets: JS and data files
// that carry executable logic or structured data relevant to security review.
var coreExtensions = map[string]string{
	".js":   "js",
	".mjs":  "js",
	".json": "data",
	".xml":  "data",
	".map":  "js", // source maps
}

// assetExtensions are only fetched when --assets is set (visual clone mode).
var assetExtensions = map[string]string{
	".css":   "css",
	".jpg":   "imgs",
	".jpeg":  "imgs",
	".gif":   "imgs",
	".png":   "imgs",
	".svg":   "imgs",
	".webp":  "imgs",
	".ico":   "imgs",
	".woff":  "fonts",
	".woff2": "fonts",
	".ttf":   "fonts",
	".eot":   "fonts",
	".otf":   "fonts",
}

// Extractor downloads a single asset URL into the correct subdirectory of
// projectPath.  In security mode (assets=false) only coreExtensions are
// fetched; in assets mode all extension maps are consulted.
//
// For JS files, it additionally checks for a sourceMappingURL directive and
// fetches the referenced source map automatically.
func Extractor(client *http.Client, link string, projectPath string, assets bool) error {
	fmt.Println("Extracting --> ", link)

	ext := parser.URLExtension(link)
	dirPath, ok := coreExtensions[ext]
	if !ok {
		if !assets {
			return nil
		}
		dirPath, ok = assetExtensions[ext]
		if !ok {
			return nil
		}
	}

	resp, err := client.Get(link)
	if err != nil {
		return fmt.Errorf("failed to GET %s: %w", link, err)
	}
	defer resp.Body.Close()

	localName := parser.URLToLocalFilename(link)

	// JS files are buffered so we can scan for sourceMappingURL after saving.
	if ext == ".js" || ext == ".mjs" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", link, err)
		}
		if err := writeFileToPath(projectPath, localName, dirPath, bytes.NewReader(body)); err != nil {
			return err
		}
		fetchSourceMap(client, link, body, projectPath, assets)
		return nil
	}

	return writeFileToPath(projectPath, localName, dirPath, resp.Body)
}

// fetchSourceMap scans jsBody for a sourceMappingURL directive and fetches
// the referenced map file.  Data-URL-embedded maps are skipped (already inline).
func fetchSourceMap(client *http.Client, jsURL string, jsBody []byte, projectPath string, assets bool) {
	mapURL := sourceMapURL(jsBody)
	if mapURL == "" || strings.HasPrefix(mapURL, "data:") {
		return
	}

	base, err := url.Parse(jsURL)
	if err != nil {
		return
	}
	ref, err := url.Parse(mapURL)
	if err != nil {
		return
	}
	abs := base.ResolveReference(ref).String()

	fmt.Println("Source map found -->", abs)
	if err := Extractor(client, abs, projectPath, assets); err != nil {
		fmt.Printf("warning: source map %s: %v\n", abs, err)
	}
}

// sourceMapURL extracts the URL from a //# sourceMappingURL= or
// //@ sourceMappingURL= directive.  Only the last 512 bytes are scanned
// because the directive must appear at the end of the file.
func sourceMapURL(body []byte) string {
	tail := body
	if len(tail) > 512 {
		tail = tail[len(tail)-512:]
	}
	s := string(tail)
	for _, prefix := range []string{"//# sourceMappingURL=", "//@ sourceMappingURL="} {
		if idx := strings.LastIndex(s, prefix); idx >= 0 {
			val := s[idx+len(prefix):]
			if nl := strings.IndexAny(val, "\r\n"); nl >= 0 {
				val = val[:nl]
			}
			return strings.TrimSpace(val)
		}
	}
	return ""
}

func writeFileToPath(projectPath, filename, fileDir string, body io.Reader) error {
	fullPath := filepath.Join(projectPath, fileDir, filename)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0777); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", fullPath, err)
	}

	f, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", fullPath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, body); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}
	return nil
}
