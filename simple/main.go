package main

import (
	"fmt"
	"net/http"
)

func handkeFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if r.URL.Path == "/" {
		fmt.Fprint(w, "<h1>Welcome to my awesome! site</h1>")
	} else if r.URL.Path == "/contact" {
		fmt.Fprint(w, "to get in touch <ahref=\"mailto:lol@loc.com\">myemail</a>")
	}
}

func main() {
	http.HandleFunc("/", handkeFunc)
	http.ListenAndServe(":3000", nil)

}
