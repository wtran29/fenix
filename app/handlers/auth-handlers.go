package handlers

import (
	"app/data"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/wtran29/fenix/mailer"

	"github.com/wtran29/fenix/urlsigner"
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
		// w.Write([]byte("Invalid password!"))
		h.App.InfoLog.Println("Invalid password by user")

		h.App.Session.Put(r.Context(), "error", "Invalid credentials. Please try again.")
		http.Redirect(w, r, "/users/login", http.StatusSeeOther)
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

func (h *Handlers) Forgot(w http.ResponseWriter, r *http.Request) {
	err := h.render(w, r, "forgot", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering: ", err)
		h.App.ErrorIntServerErr(w, r)
	}
}

func (h *Handlers) PostForgot(w http.ResponseWriter, r *http.Request) {
	// parse form
	err := r.ParseForm()
	if err != nil {
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	// verify email exists
	var u *data.User
	email := r.Form.Get("email")
	u, err = u.GetByEmail(email)
	if err != nil {
		// http.Redirect(w, r, "/users/forgot-password", http.StatusSeeOther)
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	// create a link to password reset form - /users/reset-password
	link := fmt.Sprintf("%s/users/reset-password?email=%s", h.App.Server.URL, email)
	// sign the link
	sign := urlsigner.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}

	signedLink := sign.GenerateTokenFromString(link)
	h.App.InfoLog.Println("signed link is", signedLink)

	// email the message
	var data struct {
		Link string
	}

	data.Link = signedLink

	msg := mailer.Message{
		To:       u.Email,
		Subject:  "Password reset",
		Template: "password-reset",
		Data:     data,
		From:     "admin@example.com",
	}
	h.App.Mail.Jobs <- msg
	res := <-h.App.Mail.Results
	if res.Error != nil {
		fmt.Println("error processing email:", res.Error)
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	// redirect the user
	h.App.Session.Put(r.Context(), "flash", "Reset password sent. Please check your email.")
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

func (h *Handlers) ResetPasswordForm(w http.ResponseWriter, r *http.Request) {
	// get form values
	email := r.URL.Query().Get("email")
	url := r.RequestURI
	testUrl := fmt.Sprintf("%s%s", h.App.Server.URL, url)

	// validate the url
	signer := urlsigner.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}

	valid := signer.VerifyToken(testUrl)
	if !valid {
		h.App.ErrorLog.Print("invalid url")
		h.App.ErrorUnauthorized(w, r)
	}
	// validate expiry of 1 hour reset
	expired := signer.Expired(testUrl, 60)
	if expired {
		h.App.ErrorLog.Print("link expired")
		h.App.ErrorUnauthorized(w, r)
		return
	}
	// display form with encrypted email
	eEmail, _ := h.encrypt(email)
	vars := make(jet.VarMap)
	vars.Set("email", eEmail)

	err := h.render(w, r, "reset-password", vars, nil)
	if err != nil {
		h.App.ErrorLog.Print("link expired")
		return
	}
	return
}

func (h *Handlers) PostResetPassword(w http.ResponseWriter, r *http.Request) {
	// parse form
	err := r.ParseForm()
	if err != nil {
		h.App.ErrorIntServerErr(w, r)
		return
	}
	// decrypt the email
	email, err := h.decrypt(r.Form.Get("email"))
	if err != nil {
		h.App.ErrorIntServerErr(w, r)
		return
	}
	// get the user
	var u data.User
	user, err := u.GetByEmail(email)
	if err != nil {
		h.App.ErrorIntServerErr(w, r)
		return
	}
	// reset the password
	err = user.ResetPassword(user.ID, r.Form.Get("password"))
	if err != nil {
		h.App.ErrorIntServerErr(w, r)
		return
	}
	// redirect
	h.App.Session.Put(r.Context(), "flash", "Password has been reset. You can now log in.")
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}
