package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"webapp1/simple/controllers"
)

func main() {
	staticC := controllers.NewStatic()

	usersC := controllers.NewUsers()
	r := mux.NewRouter()
	r.Handle("/", staticC.HomeView).Methods("GET")
	r.Handle("/contact", staticC.ContactView).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("PUT")
	http.ListenAndServe(":3000", r)

}
