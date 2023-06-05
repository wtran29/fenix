package main

import (
	"app/data"
	"app/handlers"

	"github.com/wtran29/fenix"
)

type application struct {
	App      *fenix.Fenix
	Handlers *handlers.Handlers
	Models   data.Models
}

func main() {
	f := initApplication()
	f.App.ListenAndServe()
}
