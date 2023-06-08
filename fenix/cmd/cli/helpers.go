package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

func setup() {
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
	make model <name>	- Create a new model in the models directory

	`)
}
