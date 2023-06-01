package main

import "github.com/wtran/fenix"

type application struct {
	App *fenix.Fenix
}

func main() {
	f := initApplication()
	f.App.ListenAndServe()
}
