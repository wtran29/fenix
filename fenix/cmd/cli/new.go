package main

import (
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

func doNew(appName string) {
	appName = strings.ToLower(appName)

	// convert app name to single word

	if strings.Contains(appName, "/") {
		spltAftr := strings.SplitAfter(appName, "/")
		appName = spltAftr[(len(spltAftr) - 1)]
	}

	log.Println("App name is", appName)

	// git clone the skeleton application
	color.Green("\tCloning repository...")
	_, err := git.PlainClone("./"+appName, false, &git.CloneOptions{
		URL:      "git@github.com/wtran29/fenix-app.git",
		Progress: os.Stdout,
		Depth:    1,
	})

	if err != nil {
		exitGracefully(err)
	}

	// remove .git directory

	// create general .env file

	// create a makefile

	// update go.mod file

	// update existing .go files with correct name/imports

	// run go mod tidy in the project directory
}
