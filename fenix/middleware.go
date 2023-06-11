package fenix

import (
	"net/http"
	"strconv"

	"github.com/justinas/nosurf"
)

func (f *Fenix) SessionLoad(next http.Handler) http.Handler {
	f.InfoLog.Println("SessionLoad called")
	return f.Session.LoadAndSave(next)
}

func (f *Fenix) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	secure, _ := strconv.ParseBool(f.config.cookie.secure)

	csrfHandler.ExemptGlob("/api/*")

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Domain:   f.config.cookie.domain,
	})

	return csrfHandler
}
