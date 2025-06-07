package echokit

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/half-ogre/go-kit/kit"
	"github.com/labstack/echo/v4"
)

type LayoutModelFunc func(c echo.Context, path string, tmpl *template.Template, data interface{}) (interface{}, error)

type Renderer struct {
	layoutModelFunc   LayoutModelFunc
	templates         map[string]*template.Template
	templateFilesPath string
}

func NewRenderer(templateFilesPath string, layoutModelFunc LayoutModelFunc) *Renderer {
	return &Renderer{
		layoutModelFunc:   layoutModelFunc,
		templates:         map[string]*template.Template{},
		templateFilesPath: templateFilesPath,
	}
}

func (r *Renderer) Render(w io.Writer, path string, data interface{}, c echo.Context) error {
	tmpl, exists := r.templates[path]
	c.Logger().Debugf("template %s exists in cache: %t", path, exists)
	if !exists {
		templateFile := fmt.Sprintf("%s/%s.html", r.templateFilesPath, path)
		c.Logger().Debugf("template file: %s", templateFile)

		fileInfo, err := os.Stat(templateFile)
		if err != nil && errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("template path %s not found", templateFile)
		} else if fileInfo.IsDir() {
			return fmt.Errorf("template path %s is a directory", templateFile)
		}

		layout, partials, err := findLayoutAndPartials(c, r.templateFilesPath, filepath.Dir(templateFile))
		if err != nil {
			return kit.WrapError(err, "error finding layout and partials")
		}

		templates := append([]string{fmt.Sprintf("%s/%s.html", r.templateFilesPath, path)}, partials...)
		if layout != "" {
			templates = append([]string{layout}, templates...)
		}

		tmpl, err = template.ParseFiles(templates...)
		if err != nil {
			return kit.WrapError(err, "error parsing template files")
		}

		if !c.Echo().Debug {
			r.templates[path] = tmpl
		}
	}

	layoutModel, err := r.layoutModelFunc(c, path, tmpl, data)
	if err != nil {
		return kit.WrapError(err, "error getting layout model")
	}

	return tmpl.ExecuteTemplate(w, "layout", &layoutModel)
}

func findLayoutAndPartials(c echo.Context, templateFilesPath string, dir string) (layout string, partials []string, err error) {
	c.Logger().Debugf("dir: %s", dir)

	if templateFilesPath != dir && !strings.Contains(dir, templateFilesPath) {
		return "", nil, fmt.Errorf("path %s is not a subdirectory of %s", dir, templateFilesPath)
	}

	foundPartials := []string{}

	files, err := os.ReadDir(dir)
	if err != nil {
		return "", nil, kit.WrapError(err, "error reading files for path %s", dir)
	}

	for _, f := range files {
		c.Logger().Debugf("found template file %s in path %s", f.Name(), dir)

		if !f.IsDir() && strings.HasPrefix(f.Name(), "_") && strings.HasSuffix(f.Name(), ".html") && f.Name() != "_layout.html" {
			foundPartials = append(foundPartials, fmt.Sprintf("%s/%s", templateFilesPath, f.Name()))
		}
	}

	if templateFilesPath != dir {
		parentDir := filepath.Dir(dir)
		c.Logger().Debugf("parent dir: %s", parentDir)

		parentLayout, parentPartials, err := findLayoutAndPartials(c, templateFilesPath, parentDir)
		if err != nil {
			return "", nil, err
		}

		if parentLayout != "" {
			return parentLayout, append(foundPartials, parentPartials...), nil
		} else if hasLayoutFile(dir) {
			return fmt.Sprintf("%s/_layout.html", dir), foundPartials, nil
		} else {
			return "", foundPartials, nil
		}
	} else {
		if hasLayoutFile(dir) {
			return fmt.Sprintf("%s/_layout.html", dir), foundPartials, nil
		} else {
			return "", foundPartials, nil
		}
	}
}

func hasLayoutFile(path string) bool {
	fileInfo, err := os.Stat(fmt.Sprintf("%s/_layout.html", path))
	if err != nil {
		return false
	} else if fileInfo.IsDir() {
		return false
	} else {
		return true
	}
}
