package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"net/http"
	"time"
	llctx "webapp1/simple/context"
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

	dbxOAuth := &oauth2.Config{
		ClientID:     cfg.Dropbox.id,
		ClientSecret: cfg.Dropbox.sercret,
		Endpoint: oauth2.Endpoint{
			TokenURL: cfg.Dropbox.tokenUrl,
			AuthURL:  cfg.Dropbox.authUrl,
		},
		RedirectURL: "http://localhost:3000/dbx/callback",
	}
	dbxRe := func(w http.ResponseWriter, r *http.Request) {
		state := csrf.Token(r)
		cookie := http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			HttpOnly: true,
			Expires:  time.Now().Add(5 * time.Minute),
		}
		http.SetCookie(w, &cookie)
		url := dbxOAuth.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusFound)
	}
	r.HandleFunc("/oauth/dbx/connect", requreUserMw.ApplyFn(dbxRe))
	dbxCallb := func(w http.ResponseWriter, r *http.Request) {

		r.ParseForm()
		state := r.FormValue("state")
		cookie, err := r.Cookie("oauth_state")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if cookie == nil || cookie.Value != state {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		cookie.Value = ""
		cookie.Expires = time.Now()

		code := r.FormValue("code")
		token, err := dbxOAuth.Exchange(context.TODO(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		user := llctx.User(r.Context())
		existin, err := services.OAuth.Find(user.ID, models.OAtuhDropbox)
		if err == models.ErrNotFound {

		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			services.OAuth.Delete(existin.ID)
		}
		oauthU := models.OAuth{
			UserID:  user.ID,
			Token:   *token,
			Service: models.OAtuhDropbox,
		}
		err = services.OAuth.Create(&oauthU)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "%+v", token)
		fmt.Fprint(w, r.FormValue("code"), r.FormValue("state"))
	}
	r.HandleFunc("/oauth/dbx/callback", requreUserMw.ApplyFn(dbxCallb))

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
