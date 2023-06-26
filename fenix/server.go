package fenix

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func (f *Fenix) ListenAndServe() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		ErrorLog:     f.ErrorLog,
		Handler:      f.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	if f.DB.Pool != nil {
		defer f.DB.Pool.Close()
	}

	if redisPool != nil {
		defer redisPool.Close()

	}

	if badgerConn != nil {
		defer badgerConn.Close()
	}

	go f.listenRPC()

	f.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	return srv.ListenAndServe()

}
