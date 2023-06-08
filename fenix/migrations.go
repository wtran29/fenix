package fenix

import (
	"log"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (f *Fenix) MigrateUp(dsn string) error {
	rootPath := filepath.ToSlash(f.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err = m.Up(); err != nil {
		log.Println("Error running migration up: ", err)
		return err
	}
	return nil
}

func (f *Fenix) MigrateDownAll(dsn string) error {
	rootPath := filepath.ToSlash(f.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err = m.Down(); err != nil {
		log.Println("Error running migration down all: ", err)
		return err
	}
	return nil
}

func (f *Fenix) Steps(n int, dsn string) error {
	rootPath := filepath.ToSlash(f.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(n); err != nil {
		return err
	}
	return nil
}

func (f *Fenix) MigrateForce(dsn string) error {
	rootPath := filepath.ToSlash(f.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(-1); err != nil {
		if err != nil {
			return err
		}
	}
	return nil
}
