package fenix

import "net/http"

func (f *Fenix) SessionLoad(next http.Handler) http.Handler {
	return f.Session.LoadAndSave(next)
}
