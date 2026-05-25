> **Fork of [goclone-dev/goclone](https://github.com/goclone-dev/goclone)** (unmaintained upstream) — rewritten and extended with AI assistance (Claude Sonnet).

<p align="center">
  <img alt="goclone" src="docs/media/logo.png">
</p>
<p align="center">
Clone websites to a local directory. Captures HTML, JavaScript, inline scripts, source maps, and manifests — with a security-first default mode that skips visual assets and focuses on what matters for code review and recon.
</p>
<p align="center">
  <a href="https://github.com/m-1tZ/goclone/actions/workflows/master-workflow.yml"><img src="https://github.com/m-1tZ/goclone/actions/workflows/master-workflow.yml/badge.svg"></a>
  <a href="https://github.com/m-1tZ/goclone/blob/master/LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg"></a>
</p>

---

## What changed from upstream

The original project was unmaintained and had several issues that made it unreliable in practice:

- **No cookies/UA/proxy on the main HTML fetch** — the page was downloaded with a bare `http.Get`, ignoring all auth options. Now a single shared `http.Client` is used for every request.
- **Global TLS mutation** — `http.DefaultTransport` was permanently modified with `InsecureSkipVerify`. Now scoped to the client.
- **Double page fetch** — the seed URL was fetched twice (once for HTML, once by colly). Replaced with `OnResponse`.
- **Context cancellation broken with `--proxy_string`** — `cancelableTransport` was only applied when no proxy was set. Now always applied.
- **Basename collisions** — `filepath.Base` was used for asset filenames, so `/css/main.css` and `/admin/css/main.css` would overwrite each other. Now the full URL path is flattened (`admin_css_main.css`).
- **`go install` broken** — module path in `go.mod` pointed at the original org. Fixed to `github.com/m-1tZ/goclone`.

## New features

- **Security-first mode (default)** — only downloads HTML, JS (`.js`, `.mjs`), source maps, and JSON/XML manifests. No CSS, fonts, or images unless you ask for them.
- **`--assets` / `-a`** — full visual clone mode: also pulls CSS, fonts (WOFF/WOFF2/TTF/EOT/OTF), and images (PNG/JPG/GIF/SVG/WebP/ICO).
- **Inline script extraction** — `<script>` blocks without a `src` attribute are saved as `js/inline_1.js`, `js/inline_2.js`, etc. These often contain API endpoints, session config, and feature flags that external bundles do not.
- **Source map auto-fetch** — after each JS file is saved, the last 512 bytes are scanned for `//# sourceMappingURL=`. If found (and not a data URL), the map is fetched and saved automatically, exposing pre-minification source.
- **PWA manifest fetch** — `<link rel="manifest">` is followed and the JSON saved to `data/`.
- **`--depth` / `-d`** — controls how many `<a href>` link levels to follow (default `1` = main page only, `0` = unlimited).
- **Redirect following fixed** — cookies, proxy, and UA are now consistently applied to all requests, so authenticated redirects work correctly.
- **Extended extension support** — `.mjs`, `.webp`, `.ico`, `.woff`, `.woff2`, `.ttf`, `.eot`, `.otf`, `.json`, `.xml`, `.map`.

## Installation

```bash
# Go 1.21+
go install github.com/m-1tZ/goclone/cmd/goclone@latest
```

Or build from source:

```bash
git clone https://github.com/m-1tZ/goclone.git
cd goclone
go build -o goclone ./cmd/goclone
```

## Usage

```
goclone <url> [flags]

Flags:
  -a, --assets                Also download CSS, fonts, and images (full visual clone)
  -C, --cookie strings        Pre-set cookies (format: "name=value")
  -d, --depth int             Maximum crawl depth (1 = main page assets only, 0 = unlimited) (default 1)
  -h, --help                  help for goclone
  -o, --open                  Open result in default browser when done
  -p, --proxy_string string   Proxy URL (http or socks5)
  -s, --serve                 Serve the cloned files with a local HTTP server
  -P, --servePort int         Port for --serve (default 5000)
  -u, --user_agent string     Custom User Agent string
```

## Examples

```bash
# Security mode (default): HTML + JS + inline scripts + source maps
goclone https://example.com

# Full visual clone: also pulls CSS, fonts, images
goclone --assets https://example.com

# With cookies (authenticated session)
goclone -C "session=abc123; csrf=xyz" https://app.example.com

# Through Burp
goclone -p http://127.0.0.1:8080 https://example.com

# Clone and immediately serve locally
goclone --serve --open https://example.com
```

## Output structure

```
<hostname>/
├── index.html
├── js/
│   ├── signin_v2_main.abc123.js    # external scripts (path-flattened)
│   ├── signin_v2_main.abc123.map   # source map (if found)
│   └── inline_1.js                 # inline <script> blocks
├── data/
│   └── manifest.json
├── css/                            # only with --assets
├── imgs/                           # only with --assets
└── fonts/                          # only with --assets
```

## License

MIT — see [LICENSE](LICENSE).
Original authors: [goclone-dev](https://github.com/goclone-dev/goclone).
