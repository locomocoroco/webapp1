package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"webapp1/simple/controllers"
	"webapp1/simple/middleware"
	"webapp1/simple/models"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "dbpass"
	dbname   = "simpleapes_dev"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	services, err := models.NewServices(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()

	requreUserMw := middleware.RequireUser{
		UserService: services.User,
	}

	r := mux.NewRouter()
	galleriesC := controllers.NewGalleries(services.Gallery, r)
	usersC := controllers.NewUsers(services.User)
	staticC := controllers.NewStatic()

	r.Handle("/", staticC.HomeView).Methods("GET")
	r.Handle("/contact", staticC.ContactView).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersC.Cookietest).Methods("GET")

	r.Handle("/galleries/new", requreUserMw.Apply(galleriesC.New)).Methods("GET")
	r.HandleFunc("/galleries", requreUserMw.ApplyFn(galleriesC.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET").Name("show_gallery")
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requreUserMw.ApplyFn(galleriesC.Edit)).Methods("GET")
	r.HandleFunc("/galleries/{id:[0-9]+}/update", requreUserMw.ApplyFn(galleriesC.Update)).Methods("POST")
	http.ListenAndServe(":3000", r)

}
