package fenix

import (
	"fmt"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

type Fenix struct {
	AppName string
	Debug   bool
	Version string
}

func (f *Fenix) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middleware"},
	}

	err := f.Init(pathConfig)
	if err != nil {
		return err
	}

	err = f.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	// read .env
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		fmt.Println("Error loading .env file")
		return err
	}

	return nil
}

func (f *Fenix) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		// create folder if it does not exist
		err := f.CreateDirIfNotExist(root + "/" + path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Fenix) checkDotEnv(path string) error {
	err := f.CreateDirIfNotExist(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}
	return nil
}
