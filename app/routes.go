package main

import (
	"fmt"
	"log"
	"myapp/data"
	"net/http"
	"strconv"

	"github.com/wtran29/fenix/fenix"
	"github.com/wtran29/fenix/fenix/cmd/filesystems/miniofilesystem"
	"github.com/wtran29/fenix/fenix/mailer"
	"github.com/wtran29/fenix/fenix/testFolder"

	"github.com/go-chi/chi/v5"
)

func (a *application) routes() *chi.Mux {
	// middleware must come before any routes
	a.use(a.Middleware.CheckRemember)

	// add routes here
	a.get("/", a.Handlers.Home)
	a.get("/go-page", a.Handlers.GoPage)
	a.get("/jet-page", a.Handlers.JetPage)
	a.get("/sessions", a.Handlers.SessionTest)
	a.get("/tester", a.Handlers.Clicker)

	a.get("/users/login", a.Handlers.UserLogin)
	a.post("/users/login", a.Handlers.PostUserLogin)
	a.get("/users/logout", a.Handlers.Logout)
	a.get("/users/forgot-password", a.Handlers.Forgot)
	a.post("/users/forgot-password", a.Handlers.PostForgot)
	a.get("/users/reset-password", a.Handlers.ResetPasswordForm)
	a.post("/users/reset-password", a.Handlers.PostResetPassword)

	a.get("/auth/{provider}", a.Handlers.SocialLogin)
	a.get("/auth/{provider}/callback", a.Handlers.SocialMediaCallback)

	a.App.Routes.Get("/form", a.Handlers.Form)
	a.App.Routes.Post("/form", a.Handlers.PostForm)

	a.get("/json", a.Handlers.JSON)
	a.get("/xml", a.Handlers.XML)
	a.get("/download-file", a.Handlers.DownloadFile)

	a.get("/crypto", a.Handlers.TestCrypto)

	a.get("/cache-test", a.Handlers.CachePage)
	a.post("/api/save-in-cache", a.Handlers.SaveInCache)
	a.post("/api/get-from-cache", a.Handlers.GetFromCache)
	a.post("/api/delete-from-cache", a.Handlers.DeleteFromCache)
	a.post("/api/empty-cache", a.Handlers.EmptyCache)

	// upload to S3
	a.get("/upload", a.Handlers.FenixUpload)
	a.post("/upload", a.Handlers.PostFenixUpload)

	// upload to preferred fs - local disk or S3 bucket
	a.get("/list-fs", a.Handlers.ListFS)
	a.get("/files/upload", a.Handlers.UploadToFS)
	a.post("/files/upload", a.Handlers.PostUploadToFS)
	a.get("/delete-from-fs", a.Handlers.DeleteFromFS)

	a.get("/test-route", testFolder.TestHandler)
	a.get("/test-minio", func(w http.ResponseWriter, r *http.Request) {
		log.Println("route to test-minio")
		fs, ok := a.App.FileSystems["MINIO"].(miniofilesystem.Minio)
		if !ok {
			log.Println("File system 'MINIO' not found or has an unexpected type")
			return
		}
		log.Println(fs)

		files, err := fs.List("")
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(files)

		for _, file := range files {
			log.Println(file.Key)
		}
	})

	a.get("/test-mail", func(w http.ResponseWriter, r *http.Request) {
		msg := mailer.Message{
			From:        "wtran4hire@gmail.com",
			To:          "wtran29@gmail.com",
			Subject:     "Test Subject - sent using an api",
			Template:    "test",
			Attachments: nil,
			Data:        nil,
		}

		a.App.Mail.Jobs <- msg
		res := <-a.App.Mail.Results
		if res.Error != nil {
			a.App.ErrorLog.Println(res.Error)
		}

		// err := a.App.Mail.SendSMTPMessage(msg)
		// if err != nil {
		// 	a.App.ErrorLog.Println(err)
		// 	return
		// }

		fmt.Fprint(w, "Sent mail!")
	})

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

	// routes from fenix
	a.App.Routes.Mount("/fenix", fenix.Routes())
	a.App.Routes.Mount("/api", a.ApiRoutes())

	return a.App.Routes
}
