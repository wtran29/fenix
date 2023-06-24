package main

import (
	"errors"
	"os"

	"strings"

	"github.com/fatih/color"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

func doMake(arg2, arg3, arg4 string) error {
	switch arg2 {
	case "key":
		rnd, err := fnx.RandomString(32)
		if err != nil {
			exitGracefully(err, "error generating key")
		}
		color.Yellow("32 character encryption key: %s", rnd)

	case "migration":
		checkForDB()

		// dbType := fnx.DB.DataType
		if arg3 == "" {
			exitGracefully(errors.New("you must give the migration a name"))
		}

		// default to migration type of fizz
		migrationType := "fizz"
		var up, down string

		// using fizz or sql?
		if arg4 == "fizz" || arg4 == "" {
			upBytes, err := templateFS.ReadFile("templates/migrations/migration_up.fizz")
			if err != nil {
				exitGracefully(err, "could not read templates/migrations/migration_up.fizz file")
			}
			downBytes, err := templateFS.ReadFile("templates/migrations/migration_down.fizz")
			if err != nil {
				exitGracefully(err, "could not read templates/migrations/migration_down.fizz file")
			}
			up = string(upBytes)
			down = string(downBytes)
		} else {
			migrationType = "sql"
		}

		// create migrations for fizz or sql
		err := fnx.CreatePopMigration([]byte(up), []byte(down), arg3, migrationType)
		if err != nil {
			exitGracefully(err)
		}

		// fileName := fmt.Sprintf("%d_%s", time.Now().UnixMicro(), arg3)

		// upFile := fnx.RootPath + "/migrations/" + fileName + "." + dbType + ".up.sql"
		// downFile := fnx.RootPath + "/migrations/" + fileName + "." + dbType + ".down.sql"

		// err := copyFileFromTemplate("templates/migrations/migration."+dbType+".up.sql", upFile)
		// if err != nil {
		// 	exitGracefully(err)
		// }

		// err = copyFileFromTemplate("templates/migrations/migration."+dbType+".down.sql", downFile)
		// if err != nil {
		// 	exitGracefully(err)
		// }

	case "auth":
		err := doAuth()
		if err != nil {
			exitGracefully(err)
		}

	case "handler":
		if arg3 == "" {
			exitGracefully(errors.New("you must give the handler a name"))
		}

		fileName := fnx.RootPath + "/handlers/" + strings.ToLower(arg3) + ".go"
		if fileExists(fileName) {
			exitGracefully(errors.New(fileName + " already exists!"))
		}

		data, err := templateFS.ReadFile("templates/handlers/handler.go.txt")
		if err != nil {
			exitGracefully(err)
		}

		handler := string(data)
		handler = strings.ReplaceAll(handler, "$HANDLERNAME$", strcase.ToCamel(arg3))

		err = os.WriteFile(fileName, []byte(handler), 0644)
		if err != nil {
			exitGracefully(err)
		}

	case "model":
		if arg3 == "" {
			exitGracefully(errors.New("you must give the model a name"))
		}

		data, err := templateFS.ReadFile("templates/data/model.go.txt")
		if err != nil {
			exitGracefully(err)
		}

		model := string(data)

		plur := pluralize.NewClient()

		var modelName = arg3
		var tableName = arg3

		if plur.IsPlural(arg3) {
			modelName = plur.Singular(arg3)
			tableName = strings.ToLower(tableName)
		} else {
			tableName = strings.ToLower(plur.Plural(arg3))
		}

		fileName := fnx.RootPath + "/data/" + strings.ToLower(modelName) + ".go"
		if fileExists(fileName) {
			exitGracefully(errors.New(fileName + " already exists!"))
		}

		model = strings.ReplaceAll(model, "$MODELNAME$", strcase.ToCamel(modelName))
		model = strings.ReplaceAll(model, "$TABLENAME$", tableName)

		err = copyDataToFile([]byte(model), fileName)
		if err != nil {
			exitGracefully(err)
		}
	case "session":
		err := doSessionTable()
		if err != nil {
			exitGracefully(err)
		}
	case "mail":
		if arg3 == "" {
			exitGracefully(errors.New("must include mail template name"))
		}
		htmlMail := fnx.RootPath + "/mail/" + strings.ToLower(arg3) + ".html.tmpl"
		plainMail := fnx.RootPath + "/mail/" + strings.ToLower(arg3) + ".plain.tmpl"

		err := copyFileFromTemplate("templates/mailer/mail.html.tmpl", htmlMail)
		if err != nil {
			exitGracefully(err)
		}
		err = copyFileFromTemplate("templates/mailer/mail.plain.tmpl", plainMail)
		if err != nil {
			exitGracefully(err)
		}
	}
	return nil
}
