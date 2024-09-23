package main

import (
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Add claims to the data
	claims := getClaims(c)
	if data != nil {
		data.(map[string]interface{})["claims"] = claims
	} else {
		data = map[string]interface{}{"claims": claims}
	}

	return t.templates.ExecuteTemplate(w, name, data)
}
