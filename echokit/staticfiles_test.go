package echokit

import (
	"context"
	"net/http"
	"strings"
	"net/http/httptest"
	"os"
	"path/filepath"
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

	t.Run("returns_304_for_matching_etag_on_index_html", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>hello</body></html>"), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())

		// First request to get ETag
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		etag := rec.Header().Get("ETag")
		assert.NotEmpty(t, etag)

		// Second request with If-None-Match
		req = httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("If-None-Match", etag)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotModified, rec.Code)
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

	t.Run("serves_non_js_css_html_files", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html><body>hello</body></html>"), 0644)
		os.WriteFile(filepath.Join(dir, "logo.png"), []byte("fake-png-data"), 0644)
		m := NewStaticFilesMiddleware(dir, false)
		defer m.Close()

		e := echo.New()
		e.Use(m.Handler())
		req := httptest.NewRequest(http.MethodGet, "/logo.png", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "fake-png-data", rec.Body.String())
		assert.Equal(t, "image/png", rec.Header().Get("Content-Type"))
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
