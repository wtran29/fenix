package fenix

import (
	"log"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/pop"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// popConnect connects to the buffalo/pop library to leverage multiple db migrations
func (f *Fenix) popConnect() (*pop.Connection, error) {
	tx, err := pop.Connect("development")
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreatePopMigration is a wrapper for pop that writes contents for the given migration
func (f *Fenix) CreatePopMigration(up, down []byte, migrationName, migrationType string) error {
	var migrationPath = f.RootPath + "/migrations"
	err := pop.MigrationCreate(migrationPath, migrationName, migrationType, up, down)
	if err != nil {
		return err
	}
	return nil
}

// RunPopMigrations is wrapper for pop to migrate up one or more migrations
func (f *Fenix) RunPopMigrations(tx *pop.Connection) error {
	var migrationPath = f.RootPath + "/migrations"

	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}

	err = fm.Up()
	if err != nil {
		return err
	}
	return nil
}

// PopMigrationDown is a wrapper that migrates down to roll back to previous db
func (f *Fenix) PopMigrationDown(tx *pop.Connection, steps ...int) error {
	var migrationPath = f.RootPath + "/migrations"

	step := 1
	if len(steps) > 0 {
		step = steps[0]
	}
	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}

	err = fm.Down(step)
	if err != nil {
		return err
	}
	return nil
}

// is a wrapper that resets the migration, runs down the migration in reverse order and migrations up
func (f *Fenix) PopMigrateReset(tx *pop.Connection) error {
	var migrationPath = f.RootPath + "/migrations"

	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}

	err = fm.Reset()
	if err != nil {
		return err
	}
	return nil

}

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
