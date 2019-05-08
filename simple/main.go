package main

import (
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"net/http"
	"webapp1/simple/controllers"
	"webapp1/simple/middleware"
	"webapp1/simple/models"
	"webapp1/simple/rand"
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

	userMw := middleware.User{
		UserService: services.User,
	}
	requreUserMw := middleware.RequireUser{
		User: userMw,
	}

	r := mux.NewRouter()
	galleriesC := controllers.NewGalleries(services.Gallery, services.Image, r)
	usersC := controllers.NewUsers(services.User)
	staticC := controllers.NewStatic()

	r.Handle("/", staticC.HomeView).Methods("GET")
	r.Handle("/contact", staticC.ContactView).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")

	isProd := false
	b, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}
	csrfMw := csrf.Protect(b, csrf.Secure(isProd))

	assetHandler := http.FileServer(http.Dir("./assets"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)

	imageHandler := http.FileServer(http.Dir("./images"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	r.Handle("/galleries", requreUserMw.ApplyFn(galleriesC.Index)).Methods("GET")
	r.Handle("/galleries/new", requreUserMw.Apply(galleriesC.New)).Methods("GET")
	r.HandleFunc("/galleries", requreUserMw.ApplyFn(galleriesC.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET").Name("show_gallery")
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requreUserMw.ApplyFn(galleriesC.Edit)).Methods("GET").Name("edit_gallery")
	r.HandleFunc("/galleries/{id:[0-9]+}/update", requreUserMw.ApplyFn(galleriesC.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", requreUserMw.ApplyFn(galleriesC.Delete)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images", requreUserMw.ApplyFn(galleriesC.ImageUpload)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", requreUserMw.ApplyFn(galleriesC.ImageDelete)).Methods("POST")
	http.ListenAndServe(":3000", csrfMw(userMw.Apply(r)))

}
