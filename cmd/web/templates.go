package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"snippetbox.mike9708.net/internal/models"
	"snippetbox.mike9708.net/ui"
)

type TemplateData struct {
	Snippet *models.Snippet
	Snippets []*models.Snippet
	CurrentYear int 
	Form any
	Flash string
	IsAuth bool
	CSRFToken string
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{} 
	
	pages, err :=  fs.Glob(ui.Files, "html/pages/*.tmpl",)
	if err != nil {
		return nil, err
	}
	

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string {
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil	
}

func humanDate(date time.Time) string{
	if date.IsZero() {
		return ""
	}
	return date.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}
