package data

import (
	"database/sql"
	db2 "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
	"os"
)

var db *sql.DB
var upper db2.Session

type Models struct {
	// any models inserted here (and in the New function)
	// are easily accessible throughout the entire app

}

func New(dbPool *sql.DB) Models {
	db = dbPool

	if os.Getenv("DATABASE_TYPE") == "mysql" || os.Getenv("DATABASE_TYPE") == "mariadb" {
		// TODO
		upper, _ = mysql.New(dbPool)
	} else {
		upper, _ = postgresql.New(dbPool)
	}

	return Models{}
}
