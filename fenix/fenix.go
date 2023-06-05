package fenix

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/wtran29/fenix/render"
	"github.com/wtran29/fenix/session"
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
	Session  *scs.SessionManager
	DB       Database
	JetViews *jet.Set
	config   config
}

type config struct {
	port        string
	renderer    string // template engine used
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
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

	// connect to the database
	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := f.OpenDB(os.Getenv("DATABASE_TYPE"), f.BuildDSN())
		if err != nil {
			errorLog.Println(err)
			os.Exit(1)
		}
		f.DB = Database{
			DataType: os.Getenv("DATABASE_TYPE"),
			Pool:     db,
		}
	}

	f.InfoLog = infoLog
	f.ErrorLog = errorLog
	f.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	f.Version = version
	f.RootPath = rootPath
	f.Routes = f.routes().(*chi.Mux)

	f.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      f.BuildDSN(),
		},
	}

	// create session

	sess := session.Session{
		CookieLifetime: f.config.cookie.lifetime,
		CookiePersist:  f.config.cookie.persist,
		CookieName:     f.config.cookie.name,
		SessionType:    f.config.sessionType,
		CookieDomain:   f.config.cookie.domain,
	}

	f.Session = sess.InitSession()

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		jet.InDevelopmentMode(),
	)

	f.JetViews = views

	f.createRenderer()

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
		Handler:      f.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	defer f.DB.Pool.Close()

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

func (f *Fenix) createRenderer() {
	renderer := render.Render{
		Renderer: f.config.renderer,
		RootPath: f.RootPath,
		Port:     f.config.port,
		JetViews: f.JetViews,
		Session:  f.Session,
	}

	f.Render = &renderer
}

func (f *Fenix) BuildDSN() string {
	var dsn string

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"))

		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}

	default:

	}
	return dsn
}
