package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

func setup(arg1, arg2 string) {
	if arg1 != "new" && arg1 != "version" && arg1 != "help" {
		err := godotenv.Load()
		if err != nil {
			exitGracefully(err)
		}

		path, err := os.Getwd()
		if err != nil {
			exitGracefully(err)
		}

		fnx.RootPath = path
		fnx.DB.DataType = os.Getenv("DATABASE_TYPE")
	}

}

func getDSN() string {
	dbType := fnx.DB.DataType

	if dbType == "pgx" {
		dbType = "postgres"
	}

	if dbType == "postgres" {
		var dsn string
		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_PASS"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"))
		} else {
			dsn = fmt.Sprintf("postgres://%ss@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"))
		}
		return dsn
	}
	return "mysql://" + fnx.BuildDSN()

}

func showHelp() {
	color.Yellow(`Available commands:

	help			- Show available commands
	version			- Print the application version
	migrate			- Runs all pending up migrations
	migrate down		- Reverse the most recent migration
	migrate reset		- Revert all migrations and then run all up migrations 
	make migrations <name>	- Create new up and down migrations files in the migrations folder
	make auth		- Create and runs migrations for auth tables, and create models and middleware
	make handler <name>	- Create a stub handler in the handlers directory
	make model <name>	- Create a new model in the data directory
	make session		- Create a table in the database as session store
	make mail <name>	- Create two starter email templates in the mail directory

	`)
}

func updateSourceFiles(path string, fi os.FileInfo, err error) error {
	// check for an error before doing anything else
	if err != nil {
		return err
	}

	// check if current file is directory
	if fi.IsDir() {
		return nil
	}

	// check go files
	match, err := filepath.Match("*.go", fi.Name())
	if err != nil {
		return err
	}

	if match {
		// read contents
		read, err := os.ReadFile(path)
		if err != nil {
			exitGracefully(err)
		}

		newContents := strings.Replace(string(read), "myapp", appURL, -1)

		// write changed file
		err = os.WriteFile(path, []byte(newContents), 0)
		if err != nil {
			exitGracefully(err)
		}
	}
	return nil
}

func updateSource() {
	// walk entire project folder, including subfolders
	err := filepath.Walk(".", updateSourceFiles)
	if err != nil {
		exitGracefully(err)
	}
}
