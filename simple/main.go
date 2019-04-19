package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)
func home (w http.ResponseWriter,r *http.Request) {
	w.Header().Set("Content-Type","text/html")
	fmt.Fprint(w, "<h1>Welcome to my awesome! site</h1>")
}
func contact (w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type","text/html")
	fmt.Fprint(w, "to get in touch <a href=\"mailto:lol@loc.com\">NilsW</a>")
}
func faq (w http.ResponseWriter,r *http.Request) {
	w.Header().Set("Content-Type","text/html")
	fmt.Fprint(w, "<h1>Here are some of the most asked Questions.</h1>")
	fmt.Fprint(w, "This that and this and that")
}

func notFound (w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type","text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>Whoops something went wrong here.</h1>")
	fmt.Fprint(w, "get in touch if error persists <a href=\"mailto:lol@loc.com\">lol@loc.com</a>")
}
func main() {
	r:=mux.NewRouter()
	r.NotFoundHandler=http.HandlerFunc(notFound)
	r.HandleFunc("/",home)
	r.HandleFunc("/contact", contact)
	r.HandleFunc("/faq",faq)
	http.ListenAndServe(":3000", r)


}
