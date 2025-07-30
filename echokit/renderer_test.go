package echokit

import (
	"bytes"
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

func TestNewRenderer(t *testing.T) {
	layoutModelFunc := func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
		return data, nil
	}
	theTemplatePath := "/templates"

	renderer := NewRenderer(theTemplatePath, layoutModelFunc)

	assert.Equal(t, theTemplatePath, renderer.templateFilesPath)
	assert.NotNil(t, renderer.layoutModelFunc)
	assert.NotNil(t, renderer.templates)
	assert.Empty(t, renderer.templates)
}

func TestHasLayoutFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "renderer_test_*")
	assert.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	t.Run("directory_with_layout_file", func(t *testing.T) {
		withLayoutDir := filepath.Join(tmpDir, "with_layout")
		err := os.MkdirAll(withLayoutDir, 0755)
		assert.NoError(t, err)
		layoutFile := filepath.Join(withLayoutDir, "_layout.html")
		err = os.WriteFile(layoutFile, []byte("<html>{{ template \"content\" . }}</html>"), 0644)
		assert.NoError(t, err)

		result := hasLayoutFile(withLayoutDir)

		assert.True(t, result)
	})

	t.Run("directory_without_layout_file", func(t *testing.T) {
		withoutLayoutDir := filepath.Join(tmpDir, "without_layout")
		err := os.MkdirAll(withoutLayoutDir, 0755)
		assert.NoError(t, err)

		result := hasLayoutFile(withoutLayoutDir)

		assert.False(t, result)
	})

	t.Run("non-existent_directory", func(t *testing.T) {
		nonExistentPath := "/non/existent/path"

		result := hasLayoutFile(nonExistentPath)

		assert.False(t, result)
	})
}

func TestFindLayoutAndPartials(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "renderer_test_*")
	assert.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	templateDir := filepath.Join(tmpDir, "templates")
	err = os.MkdirAll(templateDir, 0755)
	assert.NoError(t, err)

	rootLayout := filepath.Join(templateDir, "_layout.html")
	err = os.WriteFile(rootLayout, []byte("<html>{{ template \"content\" . }}</html>"), 0644)
	assert.NoError(t, err)

	rootPartial := filepath.Join(templateDir, "_header.html")
	err = os.WriteFile(rootPartial, []byte("<header>Header</header>"), 0644)
	assert.NoError(t, err)

	subDir := filepath.Join(templateDir, "pages")
	err = os.MkdirAll(subDir, 0755)
	assert.NoError(t, err)

	subPartial := filepath.Join(subDir, "_sidebar.html")
	err = os.WriteFile(subPartial, []byte("<sidebar>Sidebar</sidebar>"), 0644)
	assert.NoError(t, err)

	e := echo.New()
	e.Use(middleware.Logger())
	req := e.NewContext(nil, nil)

	t.Run("root_directory_with_layout", func(t *testing.T) {
		layout, partials, err := findLayoutAndPartials(req, templateDir, templateDir)

		assert.NoError(t, err)
		assert.Equal(t, rootLayout, layout)
		assert.Len(t, partials, 1)
		assert.Contains(t, partials, filepath.Join(templateDir, "_header.html"))
	})

	t.Run("subdirectory_inherits_parent_layout", func(t *testing.T) {
		layout, partials, err := findLayoutAndPartials(req, templateDir, subDir)

		assert.NoError(t, err)
		assert.Equal(t, rootLayout, layout)
		assert.Len(t, partials, 2)
		assert.Contains(t, partials, filepath.Join(templateDir, "_sidebar.html"))
		assert.Contains(t, partials, filepath.Join(templateDir, "_header.html"))
	})

	t.Run("invalid_path_outside_template_directory", func(t *testing.T) {
		layout, partials, err := findLayoutAndPartials(req, templateDir, "/invalid/path")

		assert.Error(t, err)
		assert.Empty(t, layout)
		assert.Nil(t, partials)
	})
}

func TestRenderer_Render(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "renderer_test_*")
	assert.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	layoutContent := `{{ define "layout" }}<html><body>{{ template "content" . }}</body></html>{{ end }}`
	templateContent := `{{ define "content" }}<h1>{{ .Title }}</h1><p>{{ .Message }}</p>{{ end }}`

	layoutFile := filepath.Join(tmpDir, "_layout.html")
	err = os.WriteFile(layoutFile, []byte(layoutContent), 0644)
	assert.NoError(t, err)

	templateFile := filepath.Join(tmpDir, "test.html")
	err = os.WriteFile(templateFile, []byte(templateContent), 0644)
	assert.NoError(t, err)

	e := echo.New()
	e.Use(middleware.Logger())
	req := e.NewContext(nil, nil)

	t.Run("successful_render", func(t *testing.T) {
		theData := map[string]string{"Title": "Test Title", "Message": "Test Message"}
		layoutModelFunc := func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
			return data, nil
		}
		renderer := NewRenderer(tmpDir, layoutModelFunc)

		var buf bytes.Buffer
		err := renderer.Render(&buf, "test", theData, req)

		assert.NoError(t, err)
		result := strings.TrimSpace(buf.String())
		assert.Equal(t, "<html><body><h1>Test Title</h1><p>Test Message</p></body></html>", result)
	})

	t.Run("layout_model_func_error", func(t *testing.T) {
		theData := map[string]string{"Title": "Test Title", "Message": "Test Message"}
		layoutModelFunc := func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
			return nil, errors.New("layout model error")
		}
		renderer := NewRenderer(tmpDir, layoutModelFunc)

		var buf bytes.Buffer
		err := renderer.Render(&buf, "test", theData, req)

		assert.Error(t, err)
	})

	t.Run("non-existent_template", func(t *testing.T) {
		theData := map[string]string{"Title": "Test Title", "Message": "Test Message"}
		layoutModelFunc := func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
			return data, nil
		}
		renderer := NewRenderer(tmpDir, layoutModelFunc)

		var buf bytes.Buffer
		err := renderer.Render(&buf, "nonexistent", theData, req)

		assert.Error(t, err)
	})
}

func TestRenderer_RenderCaching(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "renderer_test_*")
	assert.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	layoutContent := `{{ define "layout" }}<html><body>{{ template "content" . }}</body></html>{{ end }}`
	templateContent := `{{ define "content" }}<h1>{{ .Title }}</h1>{{ end }}`

	layoutFile := filepath.Join(tmpDir, "_layout.html")
	err = os.WriteFile(layoutFile, []byte(layoutContent), 0644)
	assert.NoError(t, err)

	templateFile := filepath.Join(tmpDir, "cached.html")
	err = os.WriteFile(templateFile, []byte(templateContent), 0644)
	assert.NoError(t, err)

	layoutModelFunc := func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
		return data, nil
	}

	t.Run("caching_enabled_in_production_mode", func(t *testing.T) {
		e := echo.New()
		e.Debug = false
		e.Use(middleware.Logger())
		req := e.NewContext(nil, nil)
		renderer := NewRenderer(tmpDir, layoutModelFunc)
		theData := map[string]string{"Title": "Cached Test"}

		var buf1 bytes.Buffer
		err := renderer.Render(&buf1, "cached", theData, req)
		assert.NoError(t, err)

		_, exists := renderer.templates["cached"]
		assert.True(t, exists, "Template should be cached in production mode")

		var buf2 bytes.Buffer
		err = renderer.Render(&buf2, "cached", theData, req)
		assert.NoError(t, err)
		assert.Equal(t, buf1.String(), buf2.String())
	})

	t.Run("caching_disabled_in_debug_mode", func(t *testing.T) {
		e := echo.New()
		e.Debug = true
		e.Use(middleware.Logger())
		req := e.NewContext(nil, nil)
		renderer := NewRenderer(tmpDir, layoutModelFunc)
		theData := map[string]string{"Title": "Debug Test"}

		var buf bytes.Buffer
		err := renderer.Render(&buf, "cached", theData, req)
		assert.NoError(t, err)

		_, exists := renderer.templates["cached"]
		assert.False(t, exists, "Template should not be cached in debug mode")
	})
}

func TestRenderer_RenderWithPartials(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "renderer_test_*")
	assert.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	layoutContent := `{{ define "layout" }}<html><body>{{ template "_header" }}{{ template "content" . }}{{ template "_footer" }}</body></html>{{ end }}`
	templateContent := `{{ define "content" }}<h1>{{ .Title }}</h1>{{ end }}`
	headerContent := `{{ define "_header" }}<header>Site Header</header>{{ end }}`
	footerContent := `{{ define "_footer" }}<footer>Site Footer</footer>{{ end }}`

	layoutFile := filepath.Join(tmpDir, "_layout.html")
	err = os.WriteFile(layoutFile, []byte(layoutContent), 0644)
	assert.NoError(t, err)

	templateFile := filepath.Join(tmpDir, "withpartials.html")
	err = os.WriteFile(templateFile, []byte(templateContent), 0644)
	assert.NoError(t, err)

	headerFile := filepath.Join(tmpDir, "_header.html")
	err = os.WriteFile(headerFile, []byte(headerContent), 0644)
	assert.NoError(t, err)

	footerFile := filepath.Join(tmpDir, "_footer.html")
	err = os.WriteFile(footerFile, []byte(footerContent), 0644)
	assert.NoError(t, err)

	e := echo.New()
	e.Use(middleware.Logger())
	req := e.NewContext(nil, nil)

	layoutModelFunc := func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
		return data, nil
	}
	renderer := NewRenderer(tmpDir, layoutModelFunc)
	theData := map[string]string{"Title": "Partials Test"}

	var buf bytes.Buffer
	err = renderer.Render(&buf, "withpartials", theData, req)

	assert.NoError(t, err)
	result := strings.TrimSpace(buf.String())
	assert.Equal(t, "<html><body><header>Site Header</header><h1>Partials Test</h1><footer>Site Footer</footer></body></html>", result)
}
