package crawler

import (
	"context"
	"net/http"
	"net/http/cookiejar"
)

// Crawl clones siteURL into projectPath.  The provided client supplies the
// transport (with TLS config and proxy already set); it is wrapped here with a
// cancelableTransport so Ctrl-C interrupts in-flight requests regardless of
// whether a proxy is in use.
func Crawl(ctx context.Context, site string, projectPath string, client *http.Client, jar *cookiejar.Jar, userAgent string, depth int, assets bool) error {
	ctxClient := *client
	ctxClient.Transport = cancelableTransport{ctx: ctx, transport: client.Transport}
	return Collector(ctx, site, projectPath, &ctxClient, jar, userAgent, depth, assets)
}
