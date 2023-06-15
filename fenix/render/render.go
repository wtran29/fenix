package render

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
)

type Render struct {
	Renderer   string // for rendering engine
	RootPath   string // for views
	Secure     bool   // is http mode
	Port       string
	ServerName string
	JetViews   *jet.Set
	Session    *scs.SessionManager
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
	Error           string // for handling errors on jet views
	Flash           string // for handling flash messages on jet views
}

func (f *Render) defaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Secure = f.Secure
	td.ServerName = f.ServerName
	td.CSRFToken = nosurf.Token(r)
	td.Port = f.Port
	if f.Session.Exists(r.Context(), "userID") {
		td.IsAuthenticated = true
	}

	td.Error = f.Session.PopString(r.Context(), "error")
	td.Flash = f.Session.PopString(r.Context(), "flash")
	return td
}

func (f *Render) Page(w http.ResponseWriter, r *http.Request, view string, variables, data interface{}) error {
	switch strings.ToLower(f.Renderer) {
	case "go":
		return f.GoPage(w, r, view, data)
	case "jet":
		return f.JetPage(w, r, view, variables, data)
	default:
	}
	return errors.New("no rendering engine specified")
}

// GoPage renders the standard Go template
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

// JetPage renders the template using the Jet template engine
func (f *Render) JetPage(w http.ResponseWriter, r *http.Request, tplName string, variables, data interface{}) error {
	var vars jet.VarMap

	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}

	td = f.defaultData(td, r)

	tpl, err := f.JetViews.GetTemplate(fmt.Sprintf("%s.jet", tplName))
	if err != nil {
		log.Println(err)
		return err
	}

	if err = tpl.Execute(w, vars, td); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
