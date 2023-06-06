package data

import (
	"fmt"
	"os"
	"testing"

	db2 "github.com/upper/db/v4"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNew(t *testing.T) {
	fakeDB, _, _ := sqlmock.New()
	defer fakeDB.Close()

	_ = os.Setenv("DATABASE_TYPE", "postgres")
	m := New(fakeDB)
	if fmt.Sprintf("%T", m) != "data.Models" {
		t.Error("wrong type", fmt.Sprintf("%T", m))
	}
}

func TestGetInsertID(t *testing.T) {
	var id db2.ID
	id = int64(1)

	insertID := getInsertID(id)
	if fmt.Sprintf("%T", insertID) != "int" {
		t.Error("wrong type returned")
	}

	id = 1
	insertID = getInsertID(id)
	if fmt.Sprintf("%T", insertID) != "int" {
		t.Error("wrong type returned")
	}
}
