package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"webapp1/simple/controllers"
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
	staticC := controllers.NewStatic()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	services, err := models.NewServices(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()

	usersC := controllers.NewUsers(services.User)
	r := mux.NewRouter()
	r.Handle("/", staticC.HomeView).Methods("GET")
	r.Handle("/contact", staticC.ContactView).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersC.Cookietest).Methods("GET")
	http.ListenAndServe(":3000", r)

}
