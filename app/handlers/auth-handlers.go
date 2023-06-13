package handlers

import (
	"app/data"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (h *Handlers) UserLogin(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.Page(w, r, "login", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println(err)
		return
	}
}

func (h *Handlers) PostUserLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := h.Models.Users.GetByEmail(email)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	pwMatch, err := user.IsPasswordMatch(password)
	if err != nil {
		w.Write([]byte("Error validating password"))
		return
	}

	if !pwMatch {
		w.Write([]byte("Invalid password!"))
		return
	}

	// check remember me?
	if r.Form.Get("remember") == "remember" {
		randStr, _ := h.randomString(12)

		sha := sha256.New()
		_, err := sha.Write([]byte(randStr))
		if err != nil {
			h.App.ErrorStatus(w, http.StatusBadRequest)
			return
		}

		hash := base64.URLEncoding.EncodeToString(sha.Sum(nil))
		rToken := data.RememberToken{}
		err = rToken.InsertToken(user.ID, hash)
		if err != nil {
			h.App.ErrorStatus(w, http.StatusBadRequest)
			return
		}

		// set cookie - default 30 days
		expiry := time.Now().Add(30 * 24 * time.Hour)
		cookie := http.Cookie{
			Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
			Value:    fmt.Sprintf("%d|%s", user.ID, hash),
			Path:     "/",
			Expires:  expiry,
			HttpOnly: true,
			Domain:   h.App.Session.Cookie.Domain,
			MaxAge:   2592000,
			Secure:   h.App.Session.Cookie.Secure,
			SameSite: h.App.Session.Cookie.SameSite,
		}
		http.SetCookie(w, &cookie)
		// save hash in session
		h.App.Session.Put(r.Context(), "userID", user.ID)
	}

	h.App.Session.Put(r.Context(), "userID", user.ID)

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// delete remember token if it exists
	if h.App.Session.Exists(r.Context(), "remember_token") {
		rToken := data.RememberToken{}
		err := rToken.Delete(h.App.Session.GetString(r.Context(), "remember_token"))
		if err != nil {
			log.Printf("Failed to delete remember token: %s", err)
		}
	}

	// delete cookie
	cookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   h.App.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   h.App.Session.Cookie.Secure,
		SameSite: h.App.Session.Cookie.SameSite,
	}
	http.SetCookie(w, &cookie)

	h.App.Session.RenewToken(r.Context())
	h.App.Session.Remove(r.Context(), "userID")
	h.App.Session.Remove(r.Context(), "remember_token")
	h.App.Session.Destroy(r.Context())
	h.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}
