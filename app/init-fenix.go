package main

import (
	"app/data"
	"app/handlers"
	"app/middleware"
	"log"
	"os"

	"github.com/wtran29/fenix"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// init fenix
	fnx := &fenix.Fenix{}
	err = fnx.New(path)
	if err != nil {
		log.Fatal(err)
	}

	fnx.AppName = "myapp"

	middleware := &middleware.Middleware{
		App: fnx,
	}

	handlers := &handlers.Handlers{
		App: fnx,
	}

	app := &application{
		App:        fnx,
		Handlers:   handlers,
		Middleware: middleware,
	}

	// overwriting the default routes from Fenix with routes from Fenix and our own routes
	app.App.Routes = app.routes()

	app.Models = data.New(app.App.DB.Pool)
	handlers.Models = app.Models
	app.Middleware.Models = app.Models

	return app
}
