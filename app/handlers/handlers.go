package handlers

import (
	"fmt"
	"io"
	"log"
	"myapp/data"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/wtran29/fenix/fenix"
	"github.com/wtran29/fenix/fenix/cmd/filesystems"
	"github.com/wtran29/fenix/fenix/cmd/filesystems/miniofilesystem"
	"github.com/wtran29/fenix/fenix/cmd/filesystems/sftpfilesystem"
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

func (h *Handlers) ListFS(w http.ResponseWriter, r *http.Request) {
	var fs filesystems.FS
	var list []filesystems.Listing

	fsType := ""
	if r.URL.Query().Get("fs-type") != "" {
		fsType = r.URL.Query().Get("fs-type")
	}

	curPath := "/"
	if r.URL.Query().Get("curPath") != "" {
		curPath = r.URL.Query().Get("curPath")
		curPath, _ = url.QueryUnescape(curPath)
	}

	if fsType != "" {
		switch fsType {
		case "MINIO":
			f := h.App.FileSystems["MINIO"].(miniofilesystem.Minio)
			fs = &f
			fsType = "MINIO"

		case "SFTP":
			f := h.App.FileSystems["SFTP"].(sftpfilesystem.SFTP)
			fs = &f
			fsType = "SFTP"
		}

		l, err := fs.List(curPath)
		if err != nil {
			h.App.ErrorLog.Println(err)
			return
		}

		list = l
	}

	vars := make(jet.VarMap)
	vars.Set("list", list)
	vars.Set("fs_type", fsType)
	vars.Set("curPath", curPath)
	err := h.render(w, r, "list-fs", vars, nil)
	if err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) UploadToFS(w http.ResponseWriter, r *http.Request) {
	fsType := r.URL.Query().Get("type")

	vars := make(jet.VarMap)
	vars.Set("fs_type", fsType)

	err := h.render(w, r, "upload", vars, nil)
	if err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) PostUploadToFS(w http.ResponseWriter, r *http.Request) {
	filename, err := getFileToUpload(r, "formFile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uploadType := r.Form.Get("upload-type")
	switch uploadType {
	case "MINIO":
		fs := h.App.FileSystems["MINIO"].(miniofilesystem.Minio)
		err = fs.Put(filename, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.App.Session.Put(r.Context(), "flash", "File uploaded!")
	http.Redirect(w, r, "/files/upload?type="+uploadType, http.StatusSeeOther)
}

func getFileToUpload(r *http.Request, fieldname string) (string, error) {
	_ = r.ParseMultipartForm(10 << 20)
	// if err != nil {
	// 	fmt.Println("could not parse multipart form:", err)
	// }
	file, header, err := r.FormFile(fieldname)
	if err != nil {
		return "", err
	}

	defer file.Close()

	dst, err := os.Create(fmt.Sprintf("./tmp/%s", header.Filename))
	if err != nil {
		return "", err
	}

	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	log.Printf("./tmp/%s", header.Filename)
	return fmt.Sprintf("./tmp/%s", header.Filename), nil
}

func (h *Handlers) DeleteFromFS(w http.ResponseWriter, r *http.Request) {
	var fs filesystems.FS
	fsType := r.URL.Query().Get("fs-type")
	item := r.URL.Query().Get("file")

	switch fsType {
	case "MINIO":
		f := h.App.FileSystems["MINIO"].(miniofilesystem.Minio)
		fs = &f

	}

	deleted := fs.Delete([]string{item})
	if deleted {
		h.App.Session.Put(r.Context(), "flash", fmt.Sprintf("%s was deleted", item))
		http.Redirect(w, r, "/list-fs?fs-type="+fsType, http.StatusSeeOther)
	}
}
