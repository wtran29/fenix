package main

import (
	"log"
	"os"

	"github.com/wtran/fenix"
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

	fnx.AppName = "testapp"

	fnx.InfoLog.Println("DEBUG is set to", fnx.Debug)

	app := &application{
		App: fnx,
	}

	return app
}
