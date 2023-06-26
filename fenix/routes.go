package fenix

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (f *Fenix) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	if f.Debug {
		mux.Use(middleware.Logger)
	}
	mux.Use(middleware.Recoverer)
	mux.Use(f.SessionLoad)
	mux.Use(f.NoSurf)
	mux.Use(f.CheckForMaintenanceMode)

	return mux
}

// Routes are fenix specific routes that are mounted in the routes file
func Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/test-fenix", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("it works!"))
	})
	return r
}
