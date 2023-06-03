package fenix

import "net/http"

func (f *Fenix) SessionLoad(next http.Handler) http.Handler {
	f.InfoLog.Println("SessionLoad called")
	return f.Session.LoadAndSave(next)
}
