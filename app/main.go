package main

import (
	"app/handlers"

	"github.com/wtran29/fenix"
)

type application struct {
	App      *fenix.Fenix
	Handlers *handlers.Handlers
}

func main() {
	f := initApplication()
	f.App.ListenAndServe()
}
