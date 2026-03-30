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
	content       []byte
	contentType   string
	etag          string
	fingerprinted bool
	vendored      bool
}

// StaticFilesMiddleware serves static files with content-hash fingerprinted URLs.
// JS and CSS files are always fingerprinted. Other files (images, fonts, etc.) are
// fingerprinted if they are referenced from HTML (src/href), JS (import), or CSS (url()).
// Fingerprinted files are served at /path/to/file.<hash>.ext with immutable caching.
// Unreferenced files and HTML files are served at their original paths with no-store.
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

			if (e.fingerprinted || e.vendored) && !m.devMode {
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				c.Response().Header().Set("Cache-Control", "no-store")
			}

			return c.Blob(http.StatusOK, e.contentType, e.content)
		}
	}
}

// build reads all static files, discovers references, fingerprints referenced files,
// rewrites references, and registers files for serving.
func (m *StaticFilesMiddleware) build() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.built {
		return nil
	}

	files := make(map[string]*staticEntry)

	// Phase 1: Read all files
	rawFiles := make(map[string][]byte)

	var walkFiles func(root, base string) error
	walkFiles = func(root, base string) error {
		return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if info.Name() != "." && strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}

			// Follow symlinks to directories by walking into them
			if info.Mode()&os.ModeSymlink != 0 {
				resolved, err := filepath.EvalSymlinks(path)
				if err != nil {
					return nil
				}
				resolvedInfo, err := os.Stat(resolved)
				if err != nil {
					return nil
				}
				if resolvedInfo.IsDir() {
					relPath, _ := filepath.Rel(root, path)
					return walkFiles(resolved, filepath.Join(base, relPath))
				}
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil // skip unreadable files
			}

			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			urlPath := "/" + filepath.ToSlash(filepath.Join(base, relPath))
			rawFiles[urlPath] = content

			return nil
		})
	}
	err := walkFiles(m.root, "")
	if err != nil {
		return err
	}

	// Phase 2: Discover referenced files and build fingerprint map
	// All JS and CSS files are fingerprinted (their references are always rewritable).
	// Other files are fingerprinted only if they are referenced from HTML, JS, or CSS.
	importFromRegex := regexp.MustCompile(`(from\s+['"])(\./[^'"]+|\.\.\/[^'"]+)(['"])`)
	importSideEffectRegex := regexp.MustCompile(`(import\s+['"])(\./[^'"]+|\.\.\/[^'"]+)(['"])`)
	dynamicImportRegex := regexp.MustCompile(`([^{]import\s*\(\s*['"])(\./[^'"]+|\.\.\/[^'"]+)(['"]\s*\))`)
	cssURLRegex := regexp.MustCompile(`(url\s*\(\s*['"]?)(\./[^'")]+|\.\.\/[^'")]+)(['"]?\s*\))`)
	htmlSrcRegex := regexp.MustCompile(`(?:src|href)="(/[^"]+)"`)

	shouldFingerprint := make(map[string]bool)

	// All JS and CSS files are always fingerprinted, except vendored files
	// (files in /vendor/ directories are versioned by directory name)
	for urlPath := range rawFiles {
		ext := filepath.Ext(urlPath)
		if (ext == ".js" || ext == ".css") && !strings.HasPrefix(urlPath, "/vendor/") {
			shouldFingerprint[urlPath] = true
		}
	}

	// Scan for referenced files
	extractRelativeRefs := func(re *regexp.Regexp, content []byte, dir string) {
		for _, match := range re.FindAllSubmatch(content, -1) {
			resolved := resolveStaticImportPath(dir, string(match[2]))
			if _, exists := rawFiles[resolved]; exists {
				shouldFingerprint[resolved] = true
			}
		}
	}

	for urlPath, content := range rawFiles {
		ext := filepath.Ext(urlPath)
		dir := filepath.Dir(urlPath)

		switch ext {
		case ".js":
			extractRelativeRefs(importFromRegex, content, dir)
			extractRelativeRefs(importSideEffectRegex, content, dir)
			extractRelativeRefs(dynamicImportRegex, content, dir)
		case ".css":
			extractRelativeRefs(cssURLRegex, content, dir)
		case ".html":
			// HTML uses absolute paths in src/href
			for _, match := range htmlSrcRegex.FindAllSubmatch(content, -1) {
				ref := string(match[1])
				if _, exists := rawFiles[ref]; exists {
					shouldFingerprint[ref] = true
				}
			}
			// Import map entries use absolute paths
			for _, ref := range extractImportMapRefs(content) {
				if _, exists := rawFiles[ref]; exists {
					shouldFingerprint[ref] = true
				}
			}
		}
	}

	// Phase 3: Build dependency graph, topologically sort, rewrite and fingerprint
	// bottom-up so that parent hashes reflect rewritten (fingerprinted) child paths.

	// Build dependency graph: parent -> children
	deps := make(map[string][]string)
	for urlPath, content := range rawFiles {
		ext := filepath.Ext(urlPath)
		dir := filepath.Dir(urlPath)

		addDeps := func(re *regexp.Regexp) {
			for _, match := range re.FindAllSubmatch(content, -1) {
				resolved := resolveStaticImportPath(dir, string(match[2]))
				if _, exists := rawFiles[resolved]; exists {
					deps[urlPath] = append(deps[urlPath], resolved)
				}
			}
		}

		switch ext {
		case ".js":
			addDeps(importFromRegex)
			addDeps(importSideEffectRegex)
			addDeps(dynamicImportRegex)
		case ".css":
			addDeps(cssURLRegex)
		case ".html":
			for _, match := range htmlSrcRegex.FindAllSubmatch(content, -1) {
				ref := string(match[1])
				if _, exists := rawFiles[ref]; exists {
					deps[urlPath] = append(deps[urlPath], ref)
				}
			}
			for _, ref := range extractImportMapRefs(content) {
				if _, exists := rawFiles[ref]; exists {
					deps[urlPath] = append(deps[urlPath], ref)
				}
			}
		}
	}

	// Topological sort (Kahn's algorithm)
	inDegree := make(map[string]int)
	reverseDeps := make(map[string][]string) // child -> parents
	for urlPath := range rawFiles {
		if _, ok := inDegree[urlPath]; !ok {
			inDegree[urlPath] = 0
		}
	}
	for parent, children := range deps {
		for _, child := range children {
			inDegree[parent]++
			reverseDeps[child] = append(reverseDeps[child], parent)
			_ = inDegree[child] // ensure child is in map
		}
	}

	var order []string
	var queue []string
	for path, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, path)
		}
	}
	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		order = append(order, path)
		for _, parent := range reverseDeps[path] {
			inDegree[parent]--
			if inDegree[parent] == 0 {
				queue = append(queue, parent)
			}
		}
	}
	// Detect and warn about circular dependencies — files stuck in cycles
	// will have their imports only partially rewritten (fingerprinting may be wrong).
	ordered := make(map[string]bool, len(order))
	for _, p := range order {
		ordered[p] = true
	}
	for path, deg := range inDegree {
		if deg > 0 && !ordered[path] {
			// Find which of this file's dependencies are also stuck
			var cycle []string
			for _, child := range deps[path] {
				if !ordered[child] {
					cycle = append(cycle, child)
				}
			}
			slog.Warn("static file has circular dependency — fingerprinted imports may be incomplete",
				"file", path, "unresolved_deps", cycle)
		}
	}

	// Add remaining files (cycles or disconnected) in arbitrary order
	for path := range rawFiles {
		if !ordered[path] {
			order = append(order, path)
		}
	}

	// Process files in topological order (leaves first).
	// Rewrite references using already-computed fingerprints, then hash the result.
	fingerprints := make(map[string]string)
	rewrittenContent := make(map[string][]byte)

	rewriteRelativePaths := func(re *regexp.Regexp, src []byte, dir string) []byte {
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

	for _, urlPath := range order {
		ext := filepath.Ext(urlPath)
		content := rawFiles[urlPath]
		dir := filepath.Dir(urlPath)

		// Rewrite JS imports
		if ext == ".js" {
			content = rewriteRelativePaths(importFromRegex, content, dir)
			content = rewriteRelativePaths(importSideEffectRegex, content, dir)
			content = rewriteRelativePaths(dynamicImportRegex, content, dir)
		}

		// Rewrite CSS url() references
		if ext == ".css" {
			content = rewriteRelativePaths(cssURLRegex, content, dir)
		}

		// Rewrite HTML src/href attributes and import maps
		if ext == ".html" {
			content = rewriteStaticHTML(content, fingerprints)
			if m.devMode {
				content = bytes.Replace(content, []byte("</body>"), reloadScript, 1)
			}
		}

		rewrittenContent[urlPath] = content

		// Compute fingerprint from rewritten content (not raw)
		if shouldFingerprint[urlPath] {
			hash := md5.Sum(content)
			shortHash := fmt.Sprintf("%x", hash[:6])
			base := strings.TrimSuffix(filepath.Base(urlPath), ext)
			fpDir := filepath.Dir(urlPath)
			fingerprintedPath := strings.TrimRight(fpDir, "/") + "/" + base + "." + shortHash + ext
			fingerprints[urlPath] = fingerprintedPath
		}
	}

	// Register all files for serving
	for urlPath, content := range rewrittenContent {
		ct := staticContentType(filepath.Ext(urlPath))
		hash := md5.Sum(content)
		e := &staticEntry{
			content:     content,
			contentType: ct,
			etag:        fmt.Sprintf(`"%x"`, hash),
		}

		if fp, ok := fingerprints[urlPath]; ok {
			e.fingerprinted = true
			files[fp] = e
		} else {
			if strings.HasPrefix(urlPath, "/vendor/") {
				e.vendored = true
			}
			files[urlPath] = e
		}

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

	var watchDirs func(dir string) error
	watchDirs = func(dir string) error {
		return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() != "." && strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return fsw.Add(path)
			}
			if d.Type()&os.ModeSymlink != 0 {
				resolved, err := filepath.EvalSymlinks(path)
				if err != nil {
					return nil
				}
				info, err := os.Stat(resolved)
				if err != nil {
					return nil
				}
				if info.IsDir() {
					return watchDirs(resolved)
				}
			}
			return nil
		})
	}
	err = watchDirs(m.root)
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
	// Rewrite import map entries: "module": "/path/to/file.js" → "module": "/path/to/file.hash.js"
	s = rewriteImportMap(s, fingerprints)
	return []byte(s)
}

func rewriteImportMap(html string, fingerprints map[string]string) string {
	const startTag = `<script type="importmap">`
	const endTag = `</script>`

	startIdx := strings.Index(html, startTag)
	if startIdx < 0 {
		return html
	}
	startIdx += len(startTag)
	endIdx := strings.Index(html[startIdx:], endTag)
	if endIdx < 0 {
		return html
	}
	endIdx += startIdx

	mapContent := html[startIdx:endIdx]
	for original, fingerprinted := range fingerprints {
		mapContent = strings.ReplaceAll(mapContent, `"`+original+`"`, `"`+fingerprinted+`"`)
	}

	return html[:startIdx] + mapContent + html[endIdx:]
}

// extractImportMapRefs extracts absolute paths referenced in an HTML import map.
func extractImportMapRefs(content []byte) []string {
	importMapValueRegex := regexp.MustCompile(`":\s*"(/[^"]+)"`)
	const startTag = `<script type="importmap">`
	const endTag = `</script>`

	s := string(content)
	startIdx := strings.Index(s, startTag)
	if startIdx < 0 {
		return nil
	}
	startIdx += len(startTag)
	endIdx := strings.Index(s[startIdx:], endTag)
	if endIdx < 0 {
		return nil
	}
	mapContent := s[startIdx : startIdx+endIdx]

	var refs []string
	for _, match := range importMapValueRegex.FindAllStringSubmatch(mapContent, -1) {
		refs = append(refs, match[1])
	}
	return refs
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
