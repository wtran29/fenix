package main

import (
	"app/data"
	"app/handlers"
	"app/middleware"

	"github.com/wtran29/fenix"
)

type application struct {
	App        *fenix.Fenix
	Handlers   *handlers.Handlers
	Models     data.Models
	Middleware *middleware.Middleware
}

func main() {
	f := initApplication()
	f.App.ListenAndServe()
}
