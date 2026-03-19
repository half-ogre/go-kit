package echokit

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo/v4"
)

var reloadScript = []byte(`<script>new EventSource('/api/dev/reload').onmessage = () => location.reload()</script></body>`)

// staticEntry holds a cached static file with its fingerprinted content.
type staticEntry struct {
	content     []byte
	contentType string
	etag        string
}

// StaticFilesMiddleware serves static files with content-hash fingerprinted URLs.
// JS and CSS files are served at /path/to/file.<hash>.ext with immutable caching.
// index.html is served with no-cache + ETag so the browser always revalidates.
// In dev mode, files are re-read and re-fingerprinted on every request, and a
// live reload watcher automatically triggers browser refreshes on file changes.
type StaticFilesMiddleware struct {
	root    string
	devMode bool
	skipper func(path string) bool

	mu    sync.RWMutex
	built bool
	files map[string]*staticEntry
	spa   *staticEntry

	watcher   *fsnotify.Watcher
	notifyMu  sync.Mutex
	notify    chan struct{}
	version   int
	cancelCtx context.CancelFunc
	done      chan struct{} // closed on Close() to unblock SSE handlers
	closeOnce sync.Once
}

// StaticFilesOption configures a StaticFilesMiddleware.
type StaticFilesOption func(*StaticFilesMiddleware)

// WithSkipper sets a function that determines whether a request path should skip
// static file handling and be passed to the next handler. This is useful for
// API routes, OAuth endpoints, or any other paths handled by Echo route handlers.
func WithSkipper(skipper func(path string) bool) StaticFilesOption {
	return func(m *StaticFilesMiddleware) {
		m.skipper = skipper
	}
}

// NewStaticFilesMiddleware creates a new static files middleware.
// root is the directory containing static files (e.g. "web").
// devMode controls whether files are re-read on every request and enables live reload.
func NewStaticFilesMiddleware(root string, devMode bool, opts ...StaticFilesOption) *StaticFilesMiddleware {
	m := &StaticFilesMiddleware{
		root:    root,
		devMode: devMode,
		files:   make(map[string]*staticEntry),
		notify:  make(chan struct{}),
		done:    make(chan struct{}),
	}

	for _, opt := range opts {
		opt(m)
	}

	if devMode {
		m.startWatcher()
	}

	return m
}

// Close stops the live reload watcher, unblocks SSE handlers, and releases resources.
// Safe to call multiple times.
func (m *StaticFilesMiddleware) Close() {
	m.closeOnce.Do(func() {
		close(m.done)
		if m.cancelCtx != nil {
			m.cancelCtx()
		}
		if m.watcher != nil {
			m.watcher.Close()
		}
	})
}

// Handler returns an Echo middleware function that serves static files and,
// in dev mode, registers the live reload SSE endpoint.
func (m *StaticFilesMiddleware) Handler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			// Handle live reload SSE endpoint
			if m.devMode && path == "/api/dev/reload" {
				return m.handleSSE(c)
			}

			// Skip paths the caller wants handled by route handlers
			if m.skipper != nil && m.skipper(path) {
				return next(c)
			}

			m.mu.RLock()
			built := m.built
			m.mu.RUnlock()
			if !built {
				if err := m.build(); err != nil {
					return err
				}
			}

			m.mu.RLock()
			defer m.mu.RUnlock()

			if path == "/" {
				path = "/index.html"
			}

			e, ok := m.files[path]
			if !ok {
				// SPA fallback: serve index.html for unmatched routes
				if m.spa != nil && !strings.Contains(path, ".") {
					e = m.spa
				} else {
					return next(c)
				}
			}

			if m.devMode {
				c.Response().Header().Set("Cache-Control", "no-store")
			} else if path == "/index.html" || e == m.spa {
				c.Response().Header().Set("Cache-Control", "no-cache")
				c.Response().Header().Set("ETag", e.etag)

				if match := c.Request().Header.Get("If-None-Match"); match == e.etag {
					return c.NoContent(http.StatusNotModified)
				}
			} else {
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}

			return c.Blob(http.StatusOK, e.contentType, e.content)
		}
	}
}

// build reads all static files, computes fingerprinted URLs, and rewrites imports.
func (m *StaticFilesMiddleware) build() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.built {
		return nil
	}

	files := make(map[string]*staticEntry)

	// Phase 1: Read all files and compute hashes of raw content
	rawFiles := make(map[string][]byte)
	fingerprints := make(map[string]string)

	err := filepath.Walk(m.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(m.root, path)
		if err != nil {
			return err
		}
		urlPath := "/" + filepath.ToSlash(relPath)
		ext := filepath.Ext(path)

		rawFiles[urlPath] = content

		// Fingerprint everything except index.html (the SPA entry point)
		if urlPath != "/index.html" {
			hash := md5.Sum(content)
			shortHash := fmt.Sprintf("%x", hash[:6])
			base := strings.TrimSuffix(filepath.Base(path), ext)
			dir := filepath.Dir(urlPath)
			fingerprintedPath := strings.TrimRight(dir, "/") + "/" + base + "." + shortHash + ext
			fingerprints[urlPath] = fingerprintedPath
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Phase 2: Rewrite references and register files
	importFromRegex := regexp.MustCompile(`(from\s+['"])(\./[^'"]+|\.\.\/[^'"]+)(['"])`)
	importSideEffectRegex := regexp.MustCompile(`(import\s+['"])(\./[^'"]+|\.\.\/[^'"]+)(['"])`)
	dynamicImportRegex := regexp.MustCompile(`(import\s*\(\s*['"])(\./[^'"]+|\.\.\/[^'"]+)(['"]\s*\))`)
	cssURLRegex := regexp.MustCompile(`(url\s*\(\s*['"]?)(\./[^'")]+|\.\.\/[^'")]+)(['"]?\s*\))`)

	for urlPath, raw := range rawFiles {
		ext := filepath.Ext(urlPath)
		content := raw
		dir := filepath.Dir(urlPath)

		rewriteRelativePaths := func(re *regexp.Regexp, src []byte) []byte {
			return re.ReplaceAllFunc(src, func(match []byte) []byte {
				parts := re.FindSubmatch(match)
				importPath := string(parts[2])
				resolved := resolveStaticImportPath(dir, importPath)
				if fp, ok := fingerprints[resolved]; ok {
					relFP := relativeStaticPath(dir, fp)
					return []byte(string(parts[1]) + relFP + string(parts[3]))
				}
				return match
			})
		}

		// Rewrite JS imports
		if ext == ".js" {
			content = rewriteRelativePaths(importFromRegex, content)
			content = rewriteRelativePaths(importSideEffectRegex, content)
			content = rewriteRelativePaths(dynamicImportRegex, content)
		}

		// Rewrite CSS url() references
		if ext == ".css" {
			content = rewriteRelativePaths(cssURLRegex, content)
		}

		// Rewrite HTML src/href attributes
		if ext == ".html" {
			content = rewriteStaticHTML(content, fingerprints)
			if m.devMode {
				content = bytes.Replace(content, []byte("</body>"), reloadScript, 1)
			}
		}

		ct := staticContentType(ext)
		hash := md5.Sum(content)
		e := &staticEntry{
			content:     content,
			contentType: ct,
			etag:        fmt.Sprintf(`"%x"`, hash),
		}

		// Register at fingerprinted path (everything except index.html)
		if fp, ok := fingerprints[urlPath]; ok {
			files[fp] = e
		}

		// Also register at original path (needed for import maps, absolute references, and index.html)
		files[urlPath] = e

		if urlPath == "/index.html" {
			m.spa = e
		}
	}

	m.files = files
	m.built = true
	return nil
}

// startWatcher creates an fsnotify watcher for live reload in dev mode.
func (m *StaticFilesMiddleware) startWatcher() {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Failed to create live reload watcher", "error", err)
		return
	}

	err = filepath.WalkDir(m.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return fsw.Add(path)
		}
		return nil
	})
	if err != nil {
		slog.Error("Failed to walk directory for live reload", "error", err)
		fsw.Close()
		return
	}

	m.watcher = fsw

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelCtx = cancel

	go m.runWatcher(ctx)
	slog.Info("Live reload watcher started", "directory", m.root)
}

// runWatcher processes file system events and broadcasts reload signals.
func (m *StaticFilesMiddleware) runWatcher(ctx context.Context) {
	var debounce *time.Timer

	for {
		select {
		case <-ctx.Done():
			if debounce != nil {
				debounce.Stop()
			}
			return

		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Create) {
				filepath.WalkDir(event.Name, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return nil
					}
					if d.IsDir() {
						m.watcher.Add(path)
					}
					return nil
				})
			}

			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(100*time.Millisecond, func() {
				m.broadcastReload()
			})

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("File watcher error", "error", err)
		}
	}
}

func (m *StaticFilesMiddleware) broadcastReload() {
	// Mark as needing rebuild so next request picks up changes
	m.mu.Lock()
	m.built = false
	m.mu.Unlock()

	m.notifyMu.Lock()
	defer m.notifyMu.Unlock()
	m.version++
	close(m.notify)
	m.notify = make(chan struct{})
	slog.Debug("Live reload triggered", "version", m.version)
}

func (m *StaticFilesMiddleware) waitForReload() <-chan struct{} {
	m.notifyMu.Lock()
	defer m.notifyMu.Unlock()
	return m.notify
}

// handleSSE streams SSE reload events to the browser.
func (m *StaticFilesMiddleware) handleSSE(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "streaming not supported")
	}

	c.Response().Write([]byte(": connected\n\n"))
	flusher.Flush()

	ctx := c.Request().Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-m.done:
			return nil
		case <-m.waitForReload():
			c.Response().Write([]byte("data: reload\n\n"))
			flusher.Flush()
		}
	}
}

func rewriteStaticHTML(content []byte, fingerprints map[string]string) []byte {
	s := string(content)
	for original, fingerprinted := range fingerprints {
		s = strings.ReplaceAll(s, `src="`+original+`"`, `src="`+fingerprinted+`"`)
		s = strings.ReplaceAll(s, `href="`+original+`"`, `href="`+fingerprinted+`"`)
	}
	return []byte(s)
}

func resolveStaticImportPath(dir, importPath string) string {
	resolved := filepath.Join(dir, importPath)
	return filepath.ToSlash(filepath.Clean(resolved))
}

func relativeStaticPath(dir, target string) string {
	rel, err := filepath.Rel(dir, target)
	if err != nil {
		return target
	}
	result := filepath.ToSlash(rel)
	if !strings.HasPrefix(result, ".") {
		result = "./" + result
	}
	return result
}

func staticContentType(ext string) string {
	// Go's mime package doesn't know about .js → application/javascript
	switch ext {
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".html":
		return "text/html"
	default:
		ct := mime.TypeByExtension(ext)
		if ct == "" {
			return "application/octet-stream"
		}
		return ct
	}
}
