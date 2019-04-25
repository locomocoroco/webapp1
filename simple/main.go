package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"webapp1/views/layouts"
)

var (
	homeView    *view.View
	contactView *view.View
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := homeView.Template.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}
func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := contactView.Template.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}
func main() {
	homeView = view.NewView("../views/home.gohtml")
	contactView = view.NewView("../views/contact.gohtml")

	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)
	http.ListenAndServe(":3000", r)

}
