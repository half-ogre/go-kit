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
)

func TestNewRenderer(t *testing.T) {
	layoutModelFunc := func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
		return data, nil
	}

	renderer := NewRenderer("/templates", layoutModelFunc)

	if renderer.templateFilesPath != "/templates" {
		t.Errorf("NewRenderer() templateFilesPath = %q, want %q", renderer.templateFilesPath, "/templates")
	}
	if renderer.layoutModelFunc == nil {
		t.Error("NewRenderer() layoutModelFunc is nil")
	}
	if renderer.templates == nil {
		t.Error("NewRenderer() templates map is nil")
	}
	if len(renderer.templates) != 0 {
		t.Errorf("NewRenderer() templates map length = %d, want 0", len(renderer.templates))
	}
}

func TestHasLayoutFile(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "renderer_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a directory with layout file
	withLayoutDir := filepath.Join(tmpDir, "with_layout")
	err = os.MkdirAll(withLayoutDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	layoutFile := filepath.Join(withLayoutDir, "_layout.html")
	err = os.WriteFile(layoutFile, []byte("<html>{{ template \"content\" . }}</html>"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a directory without layout file
	withoutLayoutDir := filepath.Join(tmpDir, "without_layout")
	err = os.MkdirAll(withoutLayoutDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "directory with layout file",
			path:     withLayoutDir,
			expected: true,
		},
		{
			name:     "directory without layout file",
			path:     withoutLayoutDir,
			expected: false,
		},
		{
			name:     "non-existent directory",
			path:     "/non/existent/path",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasLayoutFile(tt.path)
			if result != tt.expected {
				t.Errorf("hasLayoutFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestFindLayoutAndPartials(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "renderer_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create template structure
	templateDir := filepath.Join(tmpDir, "templates")
	err = os.MkdirAll(templateDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create root layout
	rootLayout := filepath.Join(templateDir, "_layout.html")
	err = os.WriteFile(rootLayout, []byte("<html>{{ template \"content\" . }}</html>"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create root partial
	rootPartial := filepath.Join(templateDir, "_header.html")
	err = os.WriteFile(rootPartial, []byte("<header>Header</header>"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create subdirectory with partial
	subDir := filepath.Join(templateDir, "pages")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	subPartial := filepath.Join(subDir, "_sidebar.html")
	err = os.WriteFile(subPartial, []byte("<sidebar>Sidebar</sidebar>"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create Echo context
	e := echo.New()
	e.Use(middleware.Logger())
	req := e.NewContext(nil, nil)

	tests := []struct {
		name              string
		templateFilesPath string
		dir               string
		expectedLayout    string
		expectedPartials  []string
		expectError       bool
	}{
		{
			name:              "root directory with layout",
			templateFilesPath: templateDir,
			dir:               templateDir,
			expectedLayout:    rootLayout,
			expectedPartials:  []string{filepath.Join(templateDir, "_header.html")},
			expectError:       false,
		},
		{
			name:              "subdirectory inherits parent layout",
			templateFilesPath: templateDir,
			dir:               subDir,
			expectedLayout:    rootLayout,
			expectedPartials:  []string{filepath.Join(templateDir, "_sidebar.html"), filepath.Join(templateDir, "_header.html")},
			expectError:       false,
		},
		{
			name:              "invalid path outside template directory",
			templateFilesPath: templateDir,
			dir:               "/invalid/path",
			expectedLayout:    "",
			expectedPartials:  nil,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout, partials, err := findLayoutAndPartials(req, tt.templateFilesPath, tt.dir)

			if tt.expectError {
				if err == nil {
					t.Error("findLayoutAndPartials() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("findLayoutAndPartials() unexpected error: %v", err)
				return
			}

			if layout != tt.expectedLayout {
				t.Errorf("findLayoutAndPartials() layout = %q, want %q", layout, tt.expectedLayout)
			}

			if len(partials) != len(tt.expectedPartials) {
				t.Errorf("findLayoutAndPartials() partials length = %d, want %d", len(partials), len(tt.expectedPartials))
			}

			for i, partial := range partials {
				if i < len(tt.expectedPartials) && partial != tt.expectedPartials[i] {
					t.Errorf("findLayoutAndPartials() partials[%d] = %q, want %q", i, partial, tt.expectedPartials[i])
				}
			}
		})
	}
}

func TestRenderer_Render(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "renderer_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create template files
	layoutContent := `{{ define "layout" }}<html><body>{{ template "content" . }}</body></html>{{ end }}`
	templateContent := `{{ define "content" }}<h1>{{ .Title }}</h1><p>{{ .Message }}</p>{{ end }}`

	layoutFile := filepath.Join(tmpDir, "_layout.html")
	err = os.WriteFile(layoutFile, []byte(layoutContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	templateFile := filepath.Join(tmpDir, "test.html")
	err = os.WriteFile(templateFile, []byte(templateContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		path            string
		data            interface{}
		layoutModelFunc LayoutModelFunc
		expectError     bool
		expectedContent string
	}{
		{
			name: "successful render",
			path: "test",
			data: map[string]string{"Title": "Test Title", "Message": "Test Message"},
			layoutModelFunc: func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
				return data, nil
			},
			expectError:     false,
			expectedContent: "<html><body><h1>Test Title</h1><p>Test Message</p></body></html>",
		},
		{
			name: "layout model func error",
			path: "test",
			data: map[string]string{"Title": "Test Title", "Message": "Test Message"},
			layoutModelFunc: func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
				return nil, errors.New("layout model error")
			},
			expectError:     true,
			expectedContent: "",
		},
		{
			name: "non-existent template",
			path: "nonexistent",
			data: map[string]string{"Title": "Test Title", "Message": "Test Message"},
			layoutModelFunc: func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
				return data, nil
			},
			expectError:     true,
			expectedContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Echo context
			e := echo.New()
			e.Use(middleware.Logger())
			req := e.NewContext(nil, nil)

			// Create renderer
			renderer := NewRenderer(tmpDir, tt.layoutModelFunc)

			// Render template
			var buf bytes.Buffer
			err := renderer.Render(&buf, tt.path, tt.data, req)

			if tt.expectError {
				if err == nil {
					t.Error("Render() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Render() unexpected error: %v", err)
				return
			}

			result := strings.TrimSpace(buf.String())
			if result != tt.expectedContent {
				t.Errorf("Render() content = %q, want %q", result, tt.expectedContent)
			}
		})
	}
}

func TestRenderer_RenderCaching(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "renderer_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create template files
	layoutContent := `{{ define "layout" }}<html><body>{{ template "content" . }}</body></html>{{ end }}`
	templateContent := `{{ define "content" }}<h1>{{ .Title }}</h1>{{ end }}`

	layoutFile := filepath.Join(tmpDir, "_layout.html")
	err = os.WriteFile(layoutFile, []byte(layoutContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	templateFile := filepath.Join(tmpDir, "cached.html")
	err = os.WriteFile(templateFile, []byte(templateContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	layoutModelFunc := func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
		return data, nil
	}

	// Test with debug mode off (caching enabled)
	t.Run("caching enabled in production mode", func(t *testing.T) {
		e := echo.New()
		e.Debug = false // Production mode
		e.Use(middleware.Logger())
		req := e.NewContext(nil, nil)

		renderer := NewRenderer(tmpDir, layoutModelFunc)
		data := map[string]string{"Title": "Cached Test"}

		// First render - should cache template
		var buf1 bytes.Buffer
		err := renderer.Render(&buf1, "cached", data, req)
		if err != nil {
			t.Fatalf("First render error: %v", err)
		}

		// Check that template is cached
		if _, exists := renderer.templates["cached"]; !exists {
			t.Error("Template should be cached in production mode")
		}

		// Second render - should use cached template
		var buf2 bytes.Buffer
		err = renderer.Render(&buf2, "cached", data, req)
		if err != nil {
			t.Fatalf("Second render error: %v", err)
		}

		if buf1.String() != buf2.String() {
			t.Error("Cached render should produce same output")
		}
	})

	// Test with debug mode on (caching disabled)
	t.Run("caching disabled in debug mode", func(t *testing.T) {
		e := echo.New()
		e.Debug = true // Debug mode
		e.Use(middleware.Logger())
		req := e.NewContext(nil, nil)

		renderer := NewRenderer(tmpDir, layoutModelFunc)
		data := map[string]string{"Title": "Debug Test"}

		// Render in debug mode
		var buf bytes.Buffer
		err := renderer.Render(&buf, "cached", data, req)
		if err != nil {
			t.Fatalf("Debug render error: %v", err)
		}

		// Check that template is not cached in debug mode
		if _, exists := renderer.templates["cached"]; exists {
			t.Error("Template should not be cached in debug mode")
		}
	})
}

func TestRenderer_RenderWithPartials(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "renderer_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create template files with partials
	layoutContent := `{{ define "layout" }}<html><body>{{ template "_header" }}{{ template "content" . }}{{ template "_footer" }}</body></html>{{ end }}`
	templateContent := `{{ define "content" }}<h1>{{ .Title }}</h1>{{ end }}`
	headerContent := `{{ define "_header" }}<header>Site Header</header>{{ end }}`
	footerContent := `{{ define "_footer" }}<footer>Site Footer</footer>{{ end }}`

	layoutFile := filepath.Join(tmpDir, "_layout.html")
	err = os.WriteFile(layoutFile, []byte(layoutContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	templateFile := filepath.Join(tmpDir, "withpartials.html")
	err = os.WriteFile(templateFile, []byte(templateContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	headerFile := filepath.Join(tmpDir, "_header.html")
	err = os.WriteFile(headerFile, []byte(headerContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	footerFile := filepath.Join(tmpDir, "_footer.html")
	err = os.WriteFile(footerFile, []byte(footerContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create Echo context
	e := echo.New()
	e.Use(middleware.Logger())
	req := e.NewContext(nil, nil)

	// Create renderer
	layoutModelFunc := func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error) {
		return data, nil
	}
	renderer := NewRenderer(tmpDir, layoutModelFunc)

	// Render template with partials
	data := map[string]string{"Title": "Partials Test"}
	var buf bytes.Buffer
	err = renderer.Render(&buf, "withpartials", data, req)
	if err != nil {
		t.Fatalf("Render with partials error: %v", err)
	}

	expected := "<html><body><header>Site Header</header><h1>Partials Test</h1><footer>Site Footer</footer></body></html>"
	result := strings.TrimSpace(buf.String())
	if result != expected {
		t.Errorf("Render with partials = %q, want %q", result, expected)
	}
}
