package handlers

import (
	"app/data"
	"fmt"
	"net/http"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/wtran29/fenix"
)

type Handlers struct {
	App    *fenix.Fenix
	Models data.Models
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	defer h.App.LoadTime(time.Now())
	err := h.render(w, r, "home", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) GoPage(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.GoPage(w, r, "home", nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) JetPage(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.JetPage(w, r, "jet-template", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

// SessionTest is a handler that demos session data
func (h *Handlers) SessionTest(w http.ResponseWriter, r *http.Request) {
	myData := "bar"

	h.App.Session.Put(r.Context(), "foo", myData)

	val := h.App.Session.GetString(r.Context(), "foo")

	vars := make(jet.VarMap)

	vars.Set("foo", val)

	err := h.App.Render.JetPage(w, r, "sessions", vars, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

// JSON is a handler to demo writing JSON
func (h *Handlers) JSON(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ID      int64    `json:"id"`
		Name    string   `json:"name"`
		Hobbies []string `json:"hobbies"`
	}

	payload.ID = 10
	payload.Name = "Jack Jones"
	payload.Hobbies = []string{"karate", "tennis", "programming"}

	err := h.App.WriteJSON(w, http.StatusOK, payload)
	if err != nil {
		h.App.ErrorLog.Println(err)
	}
}

// XML is a handler to demo writing XML
func (h *Handlers) XML(w http.ResponseWriter, r *http.Request) {
	type Payload struct {
		ID      int64    `xml:"id"`
		Name    string   `xml:"name"`
		Hobbies []string `xml:"hobbies>hobby"`
	}

	var payload Payload
	payload.ID = 10
	payload.Name = "John Smith"
	payload.Hobbies = []string{"kung fu", "basketball", "mukbang"}

	err := h.App.WriteXML(w, http.StatusOK, payload)
	if err != nil {
		h.App.ErrorLog.Println(err)
	}
}

// DownloadFile is a handler that demos downloading a file
func (h *Handlers) DownloadFile(w http.ResponseWriter, r *http.Request) {
	h.App.DownloadFile(w, r, "./public/images", "fenix.png")
}

func (h *Handlers) TestCrypto(w http.ResponseWriter, r *http.Request) {
	plaintext := "hello world"
	fmt.Fprint(w, "Unencrypted: "+plaintext+"\n")
	encrypted, err := h.encrypt(plaintext)
	if err != nil {
		h.App.ErrorLog.Println(err)
		h.App.ErrorIntServerErr(w, r)
		return
	}

	fmt.Fprint(w, "Encrypted: "+encrypted+"\n")
	decrypted, err := h.decrypt(encrypted)
	if err != nil {
		h.App.ErrorLog.Println(err)
		h.App.ErrorIntServerErr(w, r)
		return
	}
	fmt.Fprint(w, "Decrypted: "+decrypted+"\n")
}
