package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)
func home (w http.ResponseWriter,r *http.Request,_ httprouter.Params) {
	w.Header().Set("Content-Type","text/html")
	fmt.Fprint(w, "<h1>Welcome to my awesome! site</h1>")
}
func contact (w http.ResponseWriter, r *http.Request,_ httprouter.Params) {
	w.Header().Set("Content-Type","text/html")
	fmt.Fprint(w, "to get in touch <a href=\"mailto:lol@loc.com\">NilsW</a>")
}
func notFound (w http.ResponseWriter, r *http.Request,) {
	w.Header().Set("Content-Type","text/html")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>Whoops shit hit the fan</h1>")

}

func main() {
	r:=httprouter.New()
	r.NotFound=http.HandlerFunc(notFound)
	r.GET("/",home)
	r.GET("/contact", contact)
	http.ListenAndServe(":3000", r)

}
