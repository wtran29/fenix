package render

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet/v6"
)

type Render struct {
	Renderer   string // for rendering engine
	RootPath   string // for views
	Secure     bool   // is http mode
	Port       string
	ServerName string
	JetViews   *jet.Set
}

type TemplateData struct {
	IsAuthenticated bool
	IntMap          map[string]int
	StringMap       map[string]string
	FloatMap        map[string]float64
	Data            map[string]interface{}
	CSRFToken       string
	Port            string
	ServerName      string
	Secure          bool
}

func (f *Render) Page(w http.ResponseWriter, r *http.Request, view string, args, data interface{}) error {
	switch strings.ToLower(f.Renderer) {
	case "go":
		return f.GoPage(w, r, view, data)
	case "jet":

	}
	return nil
}

func (f *Render) GoPage(w http.ResponseWriter, r *http.Request, view string, data interface{}) error {
	tpl := template.Must(template.ParseFiles(fmt.Sprintf("%s/views/%s.page.tmpl", f.RootPath, view)))

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}
	err := tpl.Execute(w, &td)
	if err != nil {
		return err
	}

	return nil
}
