package main

import (
	"myapp/data"
	"myapp/handlers"
	"myapp/middleware"

	"github.com/wtran29/fenix/fenix"
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
