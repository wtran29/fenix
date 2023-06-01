package fenix

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/wtran/fenix/render"
)

const version = "1.0.0"

type Fenix struct {
	AppName  string
	Debug    bool
	Version  string
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	RootPath string
	Routes   *chi.Mux
	Render   *render.Render
	config   config
}

type config struct {
	port     string
	renderer string // template engine used

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

	// create loggers
	infoLog, errorLog := f.startLoggers()
	f.InfoLog = infoLog
	f.ErrorLog = errorLog
	f.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	f.Version = version
	f.RootPath = rootPath
	f.Routes = f.routes().(*chi.Mux)

	f.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
	}

	f.Render = f.createRenderer(f)

	return nil
}

// initializes the necessary folders for Fenix
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

// Startup the web server
func (f *Fenix) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		ErrorLog:     f.ErrorLog,
		Handler:      f.routes(),
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	f.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	err := srv.ListenAndServe()
	f.ErrorLog.Fatal(err)
}

func (f *Fenix) checkDotEnv(path string) error {
	err := f.CreateDirIfNotExist(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}
	return nil
}

func (f *Fenix) startLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}

func (f *Fenix) createRenderer(fnx *Fenix) *render.Render {
	renderer := render.Render{
		Renderer: fnx.config.renderer,
		RootPath: fnx.RootPath,
		Port:     fnx.config.port,
	}

	return &renderer
}
