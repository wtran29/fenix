package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"myapp/data"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/wtran29/fenix/fenix/mailer"

	"github.com/wtran29/fenix/fenix/urlsigner"
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

	h.socialLogout(w, r)

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
		h.App.ErrorLog.Print("user clicked on expired link")
		// w.Write([]byte("This link has expired. Please resubmit the reset password form."))
		// h.App.ErrorUnauthorized(w, r)
		h.App.Session.Put(r.Context(), "error", "This link has expired. Please resubmit the form below.")
		http.Redirect(w, r, "/users/forgot-password", http.StatusSeeOther)
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

// Used to call social auth handlers that need verification
func (h *Handlers) InitSocialAuth() {
	scope := []string{"user"}
	gScope := []string{"email", "profile"}

	goth.UseProviders(
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), os.Getenv("GITHUB_CALLBACK"), scope...),
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), os.Getenv("GOOGLE_CALLBACK"), gScope...),
	)

	key := os.Getenv("KEY")
	maxAge := 86400 * 30

	st := sessions.NewCookieStore([]byte(key))
	st.MaxAge(maxAge)
	st.Options.Path = "/"
	st.Options.HttpOnly = true
	st.Options.Secure = false

	gothic.Store = st

}

func (h *Handlers) SocialLogin(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	h.App.Session.Put(r.Context(), "social_provider", provider)
	h.InitSocialAuth()

	if _, err := gothic.CompleteUserAuth(w, r); err == nil {
		// user already logged in
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		// attempt oauth login
		gothic.BeginAuthHandler(w, r)
	}
}

func (h *Handlers) SocialMediaCallback(w http.ResponseWriter, r *http.Request) {
	h.InitSocialAuth()
	oAuthUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		h.App.Session.Put(r.Context(), "error", err.Error())
		http.Redirect(w, r, "/users/login", http.StatusSeeOther)
		return
	}

	// look up user using email address
	var u data.User
	var testUser *data.User

	testUser, err = u.GetByEmail(oAuthUser.Email)
	if err != nil {
		log.Println(err)
		provider := h.App.Session.Get(r.Context(), "social_provider").(string)
		// user does not exist, so we create a new user
		var newUser data.User
		if provider == "github" {
			split := strings.Split(oAuthUser.Name, " ")
			newUser.FirstName = split[0]
			if len(split) > 1 {
				newUser.LastName = split[1]
			}
		} else if provider == "google" {
			prompt := r.URL.Query().Get("prompt")
			if prompt == "select_account" {
				http.Redirect(w, r, "/auth/google?provider=google", http.StatusSeeOther)
				return
			}
			newUser.FirstName = oAuthUser.FirstName
			newUser.LastName = oAuthUser.LastName
		}
		newUser.Active = 1
		newUser.Email = oAuthUser.Email
		newUser.Password, _ = h.randomString(20)
		newUser.CreatedAt = time.Now()
		newUser.UpdatedAt = time.Now()
		_, err := newUser.Insert(newUser)
		if err != nil {
			h.App.Session.Put(r.Context(), "error", err.Error())
			http.Redirect(w, r, "/users/login", http.StatusSeeOther)
			return
		}

		testUser, _ = u.GetByEmail(oAuthUser.Email)

	}
	h.App.Session.Put(r.Context(), "userID", testUser.ID)
	h.App.Session.Put(r.Context(), "social_token", oAuthUser.AccessToken)
	h.App.Session.Put(r.Context(), "social_email", oAuthUser.Email)

	h.App.Session.Put(r.Context(), "flash", "You have been sucessfully logged in.")
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (h *Handlers) socialLogout(w http.ResponseWriter, r *http.Request) {
	provider, ok := h.App.Session.Get(r.Context(), "social_provider").(string)
	if !ok {
		return
	}

	// call the appropriate api for the provider and revoke auth token
	// each provider has different logic for this

	switch provider {
	case "github":
		clientID := os.Getenv("GITHUB_KEY")
		clientSecret := os.Getenv("GITHUB_SECRET")
		token := h.App.Session.Get(r.Context(), "social_token").(string)

		var payload struct {
			AccessToken string `json:"access_token"`
		}

		payload.AccessToken = token

		jsonReq, err := json.Marshal(payload)
		if err != nil {
			h.App.ErrorLog.Println(err)
			return
		}
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://%s:%s@api.github.com/applications/%s/grant", clientID, clientSecret, clientID), bytes.NewBuffer(jsonReq))
		if err != nil {
			h.App.ErrorLog.Println(err)
			return
		}

		client := &http.Client{}
		_, err = client.Do(req)
		if err != nil {
			h.App.ErrorLog.Println("Error logging out of Github:", err)
			return
		}
	case "google":
		token := h.App.Session.Get(r.Context(), "social_token").(string)
		_, err := http.PostForm(fmt.Sprintf("https://accounts.google.com/o/oauth2/revoke?%s", token), nil)
		if err != nil {
			h.App.ErrorLog.Println("Error logging out of Google:", err)
			return
		}
	}
}
