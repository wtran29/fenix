package data

import (
	"database/sql"
	"fmt"
	"os"

	db2 "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
)

var db *sql.DB
var upper db2.Session

type Models struct {
	// any models inserted here (and in the New function)
	// are easily accessible throughout the entire app
	Users  User
	Tokens Token
}

func New(dbPool *sql.DB) Models {
	db = dbPool

	switch os.Getenv("DATABASE_TYPE") {
	case "mysql", "mariadb":
		upper, _ = mysql.New(dbPool)
	case "postgres", "postgresql":
		upper, _ = postgresql.New(dbPool)
	default:
		// do nothing
	}

	return Models{
		Users:  User{},
		Tokens: Token{},
	}
}

func getInsertID(id db2.ID) int {
	idType := fmt.Sprintf("%T", id)
	if idType == "int64" {
		return int(id.(int64))
	}
	return id.(int)
}
