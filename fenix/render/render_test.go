package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var pageData = []struct {
	name          string
	renderer      string
	template      string
	variables     interface{}
	data          interface{}
	errorExpected bool
	errorMessage  string
}{
	{"go_page", "go", "home", nil, nil, false, "error rendering go template"},
	// {"go_page_no_template", "go", "no-file", true, "no error rendering non-existent go template"},
	{"go_page_no_data", "go", "home", nil, nil, false, "error rendering go template with no data"},
	{"go_page_with_data", "go", "home", nil, &TemplateData{}, false, "error rendering go template with data"},
	{"jet_page", "jet", "home", nil, nil, false, "error rendering go template"},
	{"jet_page_no_template", "jet", "no-file", nil, nil, true, "no error rendering non-existent jet template"},
	{"jet_page_no_variables", "jet", "home", nil, nil, false, "error rendering jet template with no variables"},
	// {"jet_page_with_variables", "jet", "home", map[string]interface{}{"var1": "value1"}, nil, false, "error rendering jet template with variables"},
	{"invalid_page_template", "foo", "home", nil, nil, true, "no error rendering non-existent template engine"},
}

func TestRender_Page(t *testing.T) {

	for _, e := range pageData {
		r, err := http.NewRequest("GET", "/some-url", nil)
		if err != nil {
			t.Error(err)
		}

		w := httptest.NewRecorder()

		testRenderer.Renderer = e.renderer
		testRenderer.RootPath = "./testdata"

		err = testRenderer.Page(w, r, e.template, nil, e.data)
		if e.errorExpected {
			if err == nil {
				t.Errorf("%s: %s", e.name, e.errorMessage)
			}
		} else {
			if err != nil {
				t.Errorf("%s: %s: %s", e.name, e.errorMessage, err.Error())
			}
		}
	}

}

func TestRender_GoPage(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/url", nil)
	if err != nil {
		t.Error(err)
	}

	testRenderer.Renderer = "go"
	testRenderer.RootPath = "./testdata"

	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering page", err)
	}
}

func TestRender_JetPage(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/url", nil)
	if err != nil {
		t.Error(err)
	}

	testRenderer.Renderer = "jet"

	err = testRenderer.Page(w, r, "home", nil, nil)
	if err != nil {
		t.Error("Error rendering page", err)
	}
}
