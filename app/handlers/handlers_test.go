package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHome(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL + "/")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("for home page, expected status 200 but got %d", resp.StatusCode)
	}
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bodyText), "IMMORTAL") {
		fnx.TakeScreenShot(ts.URL+"/", "Hometest", 1500, 1000)
		t.Error("did not find result")
	}
}

func TestClicker(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	page := fnx.FetchPage(ts.URL + "/tester")
	outputElement := fnx.SelectElementByID(page, "output")
	button := fnx.SelectElementByID(page, "clicker")

	testHTML, err := outputElement.HTML()

	if err != nil {
		fmt.Println(err)
	}
	if strings.Contains(testHTML, "Clicked the button") {
		t.Errorf("found text that should not be there")
	}

	button.MustClick()
	fmt.Println("clicked button")
	testHTML, _ = outputElement.HTML()
	if !strings.Contains(testHTML, "Clicked the button") {
		t.Errorf("did not find the text that should be there")
	}
}

func TestHome2(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	fnx.Session.Put(ctx, "test_key", "hello world")
	h := http.HandlerFunc(testHandlers.Home)

	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Errorf("returned wrong response code; got %d but expect 200", rr.Code)
	}

	if fnx.Session.GetString(ctx, "test_key") != "hello world" {
		t.Error("did not get correct value from session")
	}
}
