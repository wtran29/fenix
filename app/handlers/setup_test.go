package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/wtran29/fenix/fenix"
	"github.com/wtran29/fenix/fenix/mailer"
	"github.com/wtran29/fenix/fenix/render"
)

var fnx fenix.Fenix
var testSession *scs.SessionManager
var testHandlers Handlers

func TestMain(m *testing.M) {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	testSession = scs.New()
	testSession.Lifetime = 24 * time.Hour
	testSession.Cookie.Persist = true
	testSession.Cookie.SameSite = http.SameSiteLaxMode
	testSession.Cookie.Secure = false

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader("../views"),
		jet.InDevelopmentMode(),
	)

	testRenderer := render.Render{
		Renderer: "jet",
		RootPath: "../",
		Port:     "4000",
		JetViews: views,
		Session:  testSession,
	}
	testKey, _ := fnx.RandomString(32)
	fnx = fenix.Fenix{
		AppName:       "myapp",
		Debug:         true,
		Version:       "1.0.0",
		ErrorLog:      errorLog,
		InfoLog:       infoLog,
		RootPath:      "../",
		Routes:        nil,
		Render:        &testRenderer,
		Session:       testSession,
		DB:            fenix.Database{},
		JetViews:      views,
		EncryptionKey: testKey,
		Cache:         nil,
		Scheduler:     nil,
		Mail:          mailer.Mail{},
		Server:        fenix.Server{},
	}

	testHandlers.App = &fnx

	os.Exit(m.Run())
}

func getRoutes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(fnx.SessionLoad)
	mux.Get("/", testHandlers.Home)
	mux.Get("/tester", testHandlers.Clicker)
	fileServer := http.FileServer(http.Dir("./../public"))
	mux.Handle("/public/*", http.StripPrefix("/public", fileServer))
	return mux
}

func getCtx(r *http.Request) context.Context {
	ctx, err := testSession.Load(r.Context(), r.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}
