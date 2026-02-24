package utils

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

type TemplateRenderer struct {
	templates map[string]*template.Template
}

func NewAdminRenderer(templatesDir string) *TemplateRenderer {
	t := &TemplateRenderer{templates: make(map[string]*template.Template)}
	funcs := TemplateFuncs()

	base := filepath.Join(templatesDir, "admin", "layouts", "base.html")
	sidebar := filepath.Join(templatesDir, "admin", "partials", "sidebar.html")
	header := filepath.Join(templatesDir, "admin", "partials", "header.html")

	pages, _ := filepath.Glob(filepath.Join(templatesDir, "admin", "pages", "*", "*.html"))
	for _, page := range pages {
		name := adminTemplateName(templatesDir, page)
		t.templates[name] = template.Must(
			template.New("").Funcs(funcs).ParseFiles(base, sidebar, header, page),
		)
	}

	login := filepath.Join(templatesDir, "admin", "pages", "login.html")
	t.templates["admin/login"] = template.Must(
		template.New("").Funcs(funcs).ParseFiles(login),
	)

	return t
}

func NewWebRenderer(templatesDir string) *TemplateRenderer {
	t := &TemplateRenderer{templates: make(map[string]*template.Template)}
	funcs := TemplateFuncs()

	base := filepath.Join(templatesDir, "web", "layouts", "base.html")
	navbar := filepath.Join(templatesDir, "web", "partials", "navbar.html")
	footer := filepath.Join(templatesDir, "web", "partials", "footer.html")
	authModal := filepath.Join(templatesDir, "web", "partials", "auth_modal.html")
	productCard := filepath.Join(templatesDir, "web", "partials", "product_card.html")

	pages, _ := filepath.Glob(filepath.Join(templatesDir, "web", "pages", "*", "*.html"))
	for _, page := range pages {
		name := webTemplateName(templatesDir, page)
		t.templates[name] = template.Must(
			template.New("").Funcs(funcs).ParseFiles(base, navbar, footer, authModal, productCard, page),
		)
	}

	return t
}

func (r *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	templateName := "base"
	if name == "admin/login" {
		templateName = "login"
	}

	return tmpl.ExecuteTemplate(w, templateName, data)
}

func adminTemplateName(base, path string) string {
	rel, _ := filepath.Rel(filepath.Join(base, "admin", "pages"), path)
	rel = strings.TrimSuffix(rel, ".html")
	return "admin/" + filepath.ToSlash(rel)
}

func webTemplateName(base, path string) string {
	rel, _ := filepath.Rel(filepath.Join(base, "web", "pages"), path)
	rel = strings.TrimSuffix(rel, ".html")
	return "web/" + filepath.ToSlash(rel)
}
