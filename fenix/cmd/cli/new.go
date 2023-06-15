package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

var appURL string

func doNew(appName string) {
	appName = strings.ToLower(appName)
	appURL = appName

	// convert app name to single word

	if strings.Contains(appName, "/") {
		spltAftr := strings.SplitAfter(appName, "/")
		appName = spltAftr[(len(spltAftr) - 1)]
	}

	log.Println("App name is", appName)

	// git clone the skeleton application
	color.Green("\tCloning repository...")
	_, err := git.PlainClone("./"+appName, false, &git.CloneOptions{
		URL:      "https://github.com/wtran29/fenix-app.git",
		Progress: os.Stdout,
		Depth:    1,
	})

	if err != nil {
		exitGracefully(err)
	}

	// remove .git directory
	err = os.RemoveAll(fmt.Sprintf("./%s/.git", appName))
	if err != nil {
		exitGracefully(err)
	}

	// create general .env file

	color.Yellow("\tCreating .env file...")
	data, err := templateFS.ReadFile("templates/env.txt")
	if err != nil {
		exitGracefully(err)
	}

	env := string(data)
	env = strings.ReplaceAll(env, "${APP_NAME}", appName)
	key, _ := fnx.RandomString(32)
	env = strings.ReplaceAll(env, "${KEY}", key)

	err = copyDataToFile([]byte(env), fmt.Sprintf("./%s/.env", appName))
	if err != nil {
		exitGracefully(err)
	}
	// create a makefile
	if runtime.GOOS == "windows" {
		source, err := os.Open(fmt.Sprintf("./%s/Makefile.windows", appName))
		if err != nil {
			exitGracefully(err)
		}
		defer source.Close()

		destination, err := os.Create(fmt.Sprintf("./%s/Makefile", appName))
		if err != nil {
			exitGracefully(err)
		}
		defer destination.Close()

		_, err = io.Copy(destination, source)
		if err != nil {
			exitGracefully(err)
		}
	} else {
		source, err := os.Open(fmt.Sprintf("./%s/Makefile.mac", appName))
		if err != nil {
			exitGracefully(err)
		}
		defer source.Close()

		destination, err := os.Create(fmt.Sprintf("./%s/Makefile", appName))
		if err != nil {
			exitGracefully(err)
		}
		defer destination.Close()

		_, err = io.Copy(destination, source)
		if err != nil {
			exitGracefully(err)
		}
	}

	_ = os.Remove("./" + appName + "/Makefile.mac")
	_ = os.Remove("./" + appName + "/Makefile.windows")

	// update go.mod file
	color.Yellow("\tCreating go.mod file...")
	_ = os.Remove("./" + appName + "/go.mod")

	data, err = templateFS.ReadFile("templates/go.mod.txt")
	if err != nil {
		exitGracefully(err)
	}

	mod := string(data)
	mod = strings.ReplaceAll(mod, "${APP_NAME}", appURL)

	err = copyDataToFile([]byte(mod), "./"+appName+"/go.mod")
	if err != nil {
		exitGracefully(err)
	}
	// update existing .go files with correct name/imports
	color.Yellow("\tUpdating source files...")
	updateSource()

	// run go mod tidy in the project directory
}
