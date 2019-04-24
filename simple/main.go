package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
)

var homeTemplate *template.Template

func home (w http.ResponseWriter,r *http.Request) {
	w.Header().Set("Content-Type","text/html")

	if err:=homeTemplate.Execute(w,nil); err!=nil {
		panic(err)
	}
}
func contact (w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type","text/html")
	fmt.Fprint(w, "to get in touch <a href=\"mailto:lol@loc.com\">NilsW</a>")
}
func main() {
	var err error
	homeTemplate, err=template.ParseFiles("../views/home.gohtml")
	if err !=nil {
		panic(err)
	}
	r:=mux.NewRouter()
	r.HandleFunc("/",home)
	r.HandleFunc("/contact", contact)
	http.ListenAndServe(":3000", r)

}
