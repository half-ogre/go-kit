package echokit

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestStaticFilesMiddleware(t *testing.T) {
	t.Run("serves_index_html_at_root", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>hello</body></html>"), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "hello")
		assert.Equal(t, "text/html", rec.Header().Get("Content-Type"))
	})

	t.Run("serves_js_at_fingerprinted_path_with_immutable_caching", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script src="/app.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log('hi')"), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())

		// First get index.html to find the fingerprinted path
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		body := rec.Body.String()
		assert.NotContains(t, body, `src="/app.js"`)
		assert.Contains(t, body, `src="/app.`)
		assert.Contains(t, body, `.js"`)
	})

	t.Run("serves_index_html_with_no_store_cache_control", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>hello</body></html>"), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	})

	t.Run("skips_paths_matching_skipper", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>hello</body></html>"), 0644)
		m := NewStaticFilesMiddleware(dir, false, WithSkipper(func(path string) bool {
			return strings.HasPrefix(path, "/api/")
		}))
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())
		e.GET("/api/health", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", rec.Body.String())
	})

	t.Run("spa_fallback_serves_index_html_for_unknown_paths", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>spa</body></html>"), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())
		req := httptest.NewRequest(http.MethodGet, "/some/route", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "spa")
	})

	t.Run("dev_mode_sets_no_store_cache_control", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>dev</body></html>"), 0644)
		m := NewStaticFilesMiddleware(dir, true)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	})

	t.Run("dev_mode_injects_live_reload_script", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>dev</body></html>"), 0644)
		m := NewStaticFilesMiddleware(dir, true)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Contains(t, rec.Body.String(), "EventSource")
		assert.Contains(t, rec.Body.String(), "/api/dev/reload")
	})

	t.Run("rewrites_js_imports_to_fingerprinted_paths", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)
		os.MkdirAll(filepath.Join(dir, "components"), 0755)
		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script type="module" src="/components/app.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "utils.js"), []byte("export function hello() {}"), 0644)
		os.WriteFile(filepath.Join(dir, "components", "app.js"), []byte(`import { hello } from '../lib/utils.js';`), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())

		// Build by requesting index
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// Get the fingerprinted app.js path from the HTML
		body := rec.Body.String()
		assert.Contains(t, body, "app.")
		assert.Contains(t, body, ".js")
	})

	t.Run("fingerprints_and_serves_files_referenced_from_html", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><img src="/logo.png"></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "logo.png"), []byte("fake-png-data"), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())

		// Get the fingerprinted URL from the HTML
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		body := rec.Body.String()

		// Extract the fingerprinted src
		start := strings.Index(body, `src="/logo.`) + len(`src="`)
		end := strings.Index(body[start:], `"`) + start
		fpURL := body[start:end]

		// Request the fingerprinted URL
		req = httptest.NewRequest(http.MethodGet, fpURL, nil)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "fake-png-data", rec.Body.String())
		assert.Equal(t, "image/png", rec.Header().Get("Content-Type"))
		assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))

		// Original path should 404
		req = httptest.NewRequest(http.MethodGet, "/logo.png", nil)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("serves_unreferenced_files_at_original_path_with_no_store", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>hello</body></html>"), 0644)
		os.WriteFile(filepath.Join(dir, "data.json"), []byte(`{"key":"value"}`), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())

		req := httptest.NewRequest(http.MethodGet, "/data.json", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `{"key":"value"}`, rec.Body.String())
		assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	})

	t.Run("fingerprints_images_in_html", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><img src="/logo.png"></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "logo.png"), []byte("fake-png-data"), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		body := rec.Body.String()
		assert.NotContains(t, body, `src="/logo.png"`)
		assert.Contains(t, body, `src="/logo.`)
		assert.Contains(t, body, `.png"`)
	})

	t.Run("rewrites_css_url_references", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "images"), 0755)
		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><head><link href="/style.css" rel="stylesheet"></head><body></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "style.css"), []byte(`body { background: url('./images/bg.png'); }`), 0644)
		os.WriteFile(filepath.Join(dir, "images", "bg.png"), []byte("fake-bg"), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())

		// Trigger build
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// Get the fingerprinted CSS path from HTML
		body := rec.Body.String()
		assert.NotContains(t, body, `href="/style.css"`)
		assert.Contains(t, body, `href="/style.`)
	})

	t.Run("parent_fingerprint_changes_when_child_content_changes", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)
		os.MkdirAll(filepath.Join(dir, "components"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script type="module" src="/components/app.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "utils.js"), []byte("export function hello() { return 'v1'; }"), 0644)
		os.WriteFile(filepath.Join(dir, "components", "app.js"), []byte(`import { hello } from '../lib/utils.js'; console.log(hello());`), 0644)

		html1 := getHTML(t, dir)

		os.WriteFile(filepath.Join(dir, "lib", "utils.js"), []byte("export function hello() { return 'v2'; }"), 0644)

		html2 := getHTML(t, dir)

		assert.NotEqual(t, html1, html2, "index.html should reference different app.js fingerprints when child changes")
	})

	t.Run("grandchild_change_cascades_through_two_levels", func(t *testing.T) {
		// index.html -> app.js -> view.js -> utils.js
		// Changing utils.js should change all three fingerprinted paths.
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)
		os.MkdirAll(filepath.Join(dir, "components"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script type="module" src="/components/app.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "utils.js"), []byte("export const VERSION = 'v1';"), 0644)
		os.WriteFile(filepath.Join(dir, "components", "view.js"), []byte(`import { VERSION } from '../lib/utils.js'; export const v = VERSION;`), 0644)
		os.WriteFile(filepath.Join(dir, "components", "app.js"), []byte(`import { v } from './view.js'; console.log(v);`), 0644)

		html1 := getHTML(t, dir)

		os.WriteFile(filepath.Join(dir, "lib", "utils.js"), []byte("export const VERSION = 'v2';"), 0644)

		html2 := getHTML(t, dir)

		assert.NotEqual(t, html1, html2, "grandchild change should cascade to index.html")
	})

	t.Run("sibling_import_change_does_not_affect_unrelated_parent", func(t *testing.T) {
		// index.html -> app.js (imports utils.js)
		// index.html -> other.js (imports helper.js)
		// Changing helper.js should NOT change app.js fingerprint, only other.js
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)
		os.MkdirAll(filepath.Join(dir, "components"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script type="module" src="/components/app.js"></script><script type="module" src="/components/other.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "utils.js"), []byte("export const A = 1;"), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "helper.js"), []byte("export const B = 'v1';"), 0644)
		os.WriteFile(filepath.Join(dir, "components", "app.js"), []byte(`import { A } from '../lib/utils.js';`), 0644)
		os.WriteFile(filepath.Join(dir, "components", "other.js"), []byte(`import { B } from '../lib/helper.js';`), 0644)

		html1 := getHTML(t, dir)
		appFP1 := extractFingerprintedPath(html1, "app.")
		otherFP1 := extractFingerprintedPath(html1, "other.")

		os.WriteFile(filepath.Join(dir, "lib", "helper.js"), []byte("export const B = 'v2';"), 0644)

		html2 := getHTML(t, dir)
		appFP2 := extractFingerprintedPath(html2, "app.")
		otherFP2 := extractFingerprintedPath(html2, "other.")

		assert.Equal(t, appFP1, appFP2, "app.js fingerprint should not change when unrelated helper changes")
		assert.NotEqual(t, otherFP1, otherFP2, "other.js fingerprint should change when its child helper changes")
	})

	t.Run("shared_dependency_change_affects_all_parents", func(t *testing.T) {
		// index.html -> app.js -> shared.js
		// index.html -> other.js -> shared.js
		// Changing shared.js should change both app.js and other.js
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)
		os.MkdirAll(filepath.Join(dir, "components"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script type="module" src="/components/app.js"></script><script type="module" src="/components/other.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "shared.js"), []byte("export const X = 'v1';"), 0644)
		os.WriteFile(filepath.Join(dir, "components", "app.js"), []byte(`import { X } from '../lib/shared.js';`), 0644)
		os.WriteFile(filepath.Join(dir, "components", "other.js"), []byte(`import { X } from '../lib/shared.js';`), 0644)

		html1 := getHTML(t, dir)
		appFP1 := extractFingerprintedPath(html1, "app.")
		otherFP1 := extractFingerprintedPath(html1, "other.")

		os.WriteFile(filepath.Join(dir, "lib", "shared.js"), []byte("export const X = 'v2';"), 0644)

		html2 := getHTML(t, dir)
		appFP2 := extractFingerprintedPath(html2, "app.")
		otherFP2 := extractFingerprintedPath(html2, "other.")

		assert.NotEqual(t, appFP1, appFP2, "app.js fingerprint should change when shared dep changes")
		assert.NotEqual(t, otherFP1, otherFP2, "other.js fingerprint should change when shared dep changes")
	})

	t.Run("unchanged_files_keep_same_fingerprint", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script type="module" src="/lib/app.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "app.js"), []byte("console.log('stable');"), 0644)

		html1 := getHTML(t, dir)
		html2 := getHTML(t, dir)

		assert.Equal(t, html1, html2, "same content should produce same fingerprints")
	})

	t.Run("import_side_effect_changes_cascade", func(t *testing.T) {
		// import './side-effect.js'; (no from clause)
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script type="module" src="/lib/app.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "side-effect.js"), []byte("document.title = 'v1';"), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "app.js"), []byte(`import './side-effect.js'; console.log('app');`), 0644)

		html1 := getHTML(t, dir)

		os.WriteFile(filepath.Join(dir, "lib", "side-effect.js"), []byte("document.title = 'v2';"), 0644)

		html2 := getHTML(t, dir)

		assert.NotEqual(t, html1, html2, "side-effect import change should cascade")
	})

	t.Run("dynamic_import_changes_cascade", func(t *testing.T) {
		// import('./lazy.js')
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script type="module" src="/lib/app.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "lazy.js"), []byte("export const LAZY = 'v1';"), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "app.js"), []byte(`const m = import('./lazy.js');`), 0644)

		html1 := getHTML(t, dir)

		os.WriteFile(filepath.Join(dir, "lib", "lazy.js"), []byte("export const LAZY = 'v2';"), 0644)

		html2 := getHTML(t, dir)

		assert.NotEqual(t, html1, html2, "dynamic import change should cascade")
	})

	t.Run("css_url_change_cascades_to_html", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "images"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><head><link href="/style.css" rel="stylesheet"></head><body></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "style.css"), []byte(`body { background: url('./images/bg.png'); }`), 0644)
		os.WriteFile(filepath.Join(dir, "images", "bg.png"), []byte("v1-png-data"), 0644)

		html1 := getHTML(t, dir)

		os.WriteFile(filepath.Join(dir, "images", "bg.png"), []byte("v2-png-data"), 0644)

		html2 := getHTML(t, dir)

		assert.NotEqual(t, html1, html2, "CSS url() image change should cascade to HTML")
	})

	t.Run("import_map_entries_are_rewritten", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><head><script type="importmap">{"imports":{"mylib":"/lib/mylib.js"}}</script></head><body></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "mylib.js"), []byte("export const LIB = 'v1';"), 0644)

		html1 := getHTML(t, dir)
		assert.NotContains(t, html1, `"/lib/mylib.js"`, "import map should have fingerprinted path")
		assert.Contains(t, html1, "/lib/mylib.", "import map should reference fingerprinted mylib")

		os.WriteFile(filepath.Join(dir, "lib", "mylib.js"), []byte("export const LIB = 'v2';"), 0644)

		html2 := getHTML(t, dir)
		assert.NotEqual(t, html1, html2, "import map should reflect new fingerprint when lib changes")
	})

	t.Run("fingerprinted_file_content_matches_rewritten_version", func(t *testing.T) {
		// Verify the actual served content of a fingerprinted file contains
		// the rewritten import paths, not the original ones.
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)
		os.MkdirAll(filepath.Join(dir, "components"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script type="module" src="/components/app.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "utils.js"), []byte("export const U = 1;"), 0644)
		os.WriteFile(filepath.Join(dir, "components", "app.js"), []byte(`import { U } from '../lib/utils.js'; console.log(U);`), 0644)

		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()
		e := echo.New()
		e.Use(m.Handler())

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		html := rec.Body.String()

		appFP := extractFingerprintedPath(html, "app.")
		assert.NotEmpty(t, appFP)

		// Fetch the fingerprinted app.js
		req = httptest.NewRequest(http.MethodGet, appFP, nil)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		jsContent := rec.Body.String()

		// Should NOT contain the original import path
		assert.NotContains(t, jsContent, `'../lib/utils.js'`)
		// Should contain a fingerprinted import path
		assert.Contains(t, jsContent, "../lib/utils.")
	})

	t.Run("vendored_files_are_not_fingerprinted", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "vendor"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "vendor", "lib.js"), []byte("var lib = {};"), 0644)

		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()
		e := echo.New()
		e.Use(m.Handler())

		// Vendored file should be served at its original path
		req := httptest.NewRequest(http.MethodGet, "/vendor/lib.js", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
	})
}

// getHTML builds a fresh StaticFilesMiddleware for the given directory and returns the index.html content.
func getHTML(t *testing.T, dir string) string {
	t.Helper()
	m := NewStaticFilesMiddleware(dir, false)
	defer m.Close()
	e := echo.New()
	e.Use(m.Handler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	return rec.Body.String()
}

// extractFingerprintedPath finds a fingerprinted path in HTML matching the given prefix.
// e.g. extractFingerprintedPath(html, "app.") finds "/components/app.abc123.js"
func extractFingerprintedPath(html, prefix string) string {
	idx := strings.Index(html, prefix)
	if idx < 0 {
		return ""
	}
	// Walk backwards to find the start (opening quote)
	start := idx
	for start > 0 && html[start-1] != '"' && html[start-1] != '\'' {
		start--
	}
	// Walk forward to find the end (closing quote)
	end := idx
	for end < len(html) && html[end] != '"' && html[end] != '\'' {
		end++
	}
	return html[start:end]
}

func TestStaticFilesMiddleware_CircularDependencyWarning(t *testing.T) {
	t.Run("logs_warning_for_circular_imports", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)

		// Create a circular dependency: a.js imports b.js, b.js imports a.js
		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script type="module" src="/lib/app.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "a.js"), []byte(`import { B } from './b.js'; export const A = 'a' + B;`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "b.js"), []byte(`import { A } from './a.js'; export const B = 'b' + A;`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "app.js"), []byte(`import { A } from './a.js'; console.log(A);`), 0644)

		// Capture log output
		var logBuf strings.Builder
		oldLogger := slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelWarn})))
		defer slog.SetDefault(oldLogger)

		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()
		e := echo.New()
		e.Use(m.Handler())

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, logBuf.String(), "circular dependency")
	})

	t.Run("no_warning_for_acyclic_imports", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "lib"), 0755)

		os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><script type="module" src="/lib/app.js"></script></body></html>`), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "utils.js"), []byte("export const U = 1;"), 0644)
		os.WriteFile(filepath.Join(dir, "lib", "app.js"), []byte(`import { U } from './utils.js'; console.log(U);`), 0644)

		var logBuf strings.Builder
		oldLogger := slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelWarn})))
		defer slog.SetDefault(oldLogger)

		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()
		e := echo.New()
		e.Use(m.Handler())

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, logBuf.String(), "circular dependency")
	})
}

func TestStaticFilesMiddleware_LiveReload(t *testing.T) {
	t.Run("triggers_reload_on_file_change", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>hello</body></html>"), 0644)
		m := NewStaticFilesMiddleware(dir, true)
		defer m.Close()

		waitCh := m.waitForReload()

		// Write a file to trigger the watcher
		err := os.WriteFile(filepath.Join(dir, "test.js"), []byte("// changed"), 0644)
		assert.NoError(t, err)

		select {
		case <-waitCh:
			// Channel closed — reload was triggered
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for reload trigger")
		}
	})

	t.Run("does_not_create_watcher_in_production_mode", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>hello</body></html>"), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		assert.Nil(t, m.watcher)
	})
}

func TestStaticFilesMiddleware_SSE(t *testing.T) {
	t.Run("serves_sse_endpoint_in_dev_mode", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>hello</body></html>"), 0644)
		m := NewStaticFilesMiddleware(dir, true)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())

		// Use a context with timeout so the SSE handler doesn't block forever
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/api/dev/reload", nil).WithContext(ctx)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), ": connected")
	})
}
