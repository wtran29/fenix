package handlers

import "net/http"

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

	h.App.Session.Put(r.Context(), "userID", user.ID)

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	h.App.Session.RenewToken(r.Context())
	h.App.Session.Remove(r.Context(), "userID")
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}