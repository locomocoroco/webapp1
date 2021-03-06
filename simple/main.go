package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"net/http"
	"webapp1/simple/controllers"
	"webapp1/simple/email"
	"webapp1/simple/middleware"
	"webapp1/simple/models"
	"webapp1/simple/rand"
)

func main() {
	boolPtr := flag.Bool("prod", false, "set true in prod."+
		"provide a .config file.")

	cfg := LoadConfig(*boolPtr)
	dbCfg := cfg.Database
	services, err := models.NewServices(models.WithGorm(dbCfg.Dialect(), dbCfg.ConnectionInfo()),
		models.WithLog(!cfg.IsProd()),
		models.WithUser(cfg.Pepper, cfg.HMACKey),
		models.WithGallery(),
		models.WithImage(),
		models.WithOAuth(),
	)
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

	mgCf := cfg.Mailgun
	emailer := email.NewClient(email.WithSender("", ""))
	email.WithMailgun(mgCf.Domain, mgCf.APIKey)

	r := mux.NewRouter()
	galleriesC := controllers.NewGalleries(services.Gallery, services.Image, r)
	usersC := controllers.NewUsers(services.User, emailer)
	staticC := controllers.NewStatic()

	configs := make(map[string]*oauth2.Config)
	configs[models.OAtuhDropbox] = &oauth2.Config{
		ClientID:     cfg.Dropbox.id,
		ClientSecret: cfg.Dropbox.sercret,
		Endpoint: oauth2.Endpoint{
			TokenURL: cfg.Dropbox.tokenUrl,
			AuthURL:  cfg.Dropbox.authUrl,
		},
		RedirectURL: "http://localhost:3000/dbx/callback",
	}
	oAuthsC := controllers.NewOauths(services.OAuth, configs)

	r.HandleFunc("/oauth/{service:[A-Za-z0-9]+}/connect", requreUserMw.ApplyFn(oAuthsC.Connect))
	r.HandleFunc("/oauth/{service:[A-Za-z0-9]+}/callback", requreUserMw.ApplyFn(oAuthsC.Callback))
	r.HandleFunc("/oauth/{service:[A-Za-z0-9]+}/test", requreUserMw.ApplyFn(oAuthsC.DropboxTest))

	r.Handle("/", staticC.HomeView).Methods("GET")
	r.Handle("/contact", staticC.ContactView).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/logout", requreUserMw.ApplyFn(usersC.Logout)).Methods("POST")

	r.Handle("/forgot", usersC.ForgotPwView).Methods("GET")
	r.HandleFunc("/forgot", usersC.InitiateReset).Methods("POST")
	r.HandleFunc("/reset", usersC.ResetPw).Methods("GET")
	r.HandleFunc("/reset", usersC.CompleteReset).Methods("POST")

	b, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}
	csrfMw := csrf.Protect(b, csrf.Secure(cfg.IsProd()))

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
	http.ListenAndServe(fmt.Sprintf("%v", cfg.Port), csrfMw(userMw.Apply(r)))

}
