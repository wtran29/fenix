package main

import (
	"myapp/data"
	"myapp/handlers"
	"myapp/middleware"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/wtran29/fenix/fenix"
)

type application struct {
	App        *fenix.Fenix
	Handlers   *handlers.Handlers
	Models     data.Models
	Middleware *middleware.Middleware
	wg         sync.WaitGroup
}

func main() {
	f := initApplication()
	go f.listenForShutdown()
	err := f.App.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		f.App.ErrorLog.Println(err)
	}
}

func (a *application) shutdown() {
	// put any clean up tasks here

	// block until the waitgroup is empty
	a.wg.Wait()
}

func (a *application) listenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit

	a.App.InfoLog.Println("Received signal", s.String())
	a.shutdown()

	os.Exit(0)
}
