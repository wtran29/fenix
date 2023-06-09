package main

import (
	"app/data"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	// middleware must come before any routes

	// add routes here
	a.get("/", a.Handlers.Home)
	a.get("/go-page", a.Handlers.GoPage)
	a.get("/jet-page", a.Handlers.JetPage)
	a.get("/sessions", a.Handlers.SessionTest)

	a.get("/users/login", a.Handlers.UserLogin)
	a.post("/users/login", a.Handlers.PostUserLogin)
	a.get("/users/logout", a.Handlers.Logout)

	a.App.Routes.Get("/form", a.Handlers.Form)
	a.App.Routes.Post("/form", a.Handlers.PostForm)

	a.get("/json", a.Handlers.JSON)
	a.get("/xml", a.Handlers.XML)
	a.get("/download-file", a.Handlers.DownloadFile)

	a.get("/crypto", a.Handlers.TestCrypto)

	a.App.Routes.Get("/create-user", func(w http.ResponseWriter, r *http.Request) {
		u := data.User{
			FirstName: "Will",
			LastName:  "Tran",
			Email:     "me@here.com",
			Active:    1,
			Password:  "password",
		}

		id, err := a.Models.Users.Insert(u)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprintf(w, "%d: %s", id, u.FirstName)
		return
	})

	a.App.Routes.Get("/get-all-users", func(w http.ResponseWriter, r *http.Request) {
		users, err := a.Models.Users.GetAll()
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}
		for _, x := range users {
			fmt.Fprint(w, x.LastName)
		}
	})

	a.App.Routes.Get("/get-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		u, err := a.Models.Users.Get(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}
		fmt.Fprintf(w, "%s %s %s", u.FirstName, u.LastName, u.Email)
		return
	})

	a.App.Routes.Get("/update-user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		u, err := a.Models.Users.Get(id)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		u.LastName, err = a.App.RandomString(10)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		validator := a.App.Validator(nil)
		u.LastName = ""

		u.Validate(validator)

		if !validator.Valid() {
			fmt.Fprint(w, "failed validation")
			return
		}

		err = u.Update(*u)
		if err != nil {
			a.App.ErrorLog.Println(err)
			return
		}

		fmt.Fprintf(w, "updated last name to %s", u.LastName)
		return
	})

	// a.App.Routes.Get("/test-database", func(w http.ResponseWriter, r *http.Request) {
	// 	query := "select id, first_name from users where id = 1"
	// 	row := a.App.DB.Pool.QueryRowContext(r.Context(), query)

	// 	var id int
	// 	var name string

	// 	err := row.Scan(&id, &name)
	// 	if err != nil {
	// 		a.App.ErrorLog.Println(err)
	// 		return
	// 	}

	// 	fmt.Fprintf(w, "%d %s", id, name)
	// })

	// static routes
	fileServer := http.FileServer(http.Dir("./public"))
	a.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))

	return a.App.Routes
}
