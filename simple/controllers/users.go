package controllers

import (
	"log"
	"net/http"
	"time"
	"webapp1/simple/context"
	"webapp1/simple/email"
	"webapp1/simple/models"
	"webapp1/simple/rand"
	"webapp1/simple/views"
)

func NewUsers(us models.UserService, emailer *email.Client) *Users {
	return &Users{
		NewView:   views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us:        us,
		emailer:   emailer,
	}
}

type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        models.UserService
	emailer   *email.Client
}

func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	var form NewForm
	parseURLParams(r, &form)
	u.NewView.Render(w, r, form)
}

type NewForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form NewForm
	vd.Yield = &form
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}
	user := models.Users{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}
	if err := u.us.Create(&user); err != nil {
		log.Println(err)
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
		return
	}
	u.emailer.Welcome(user.Name, user.Email) //err check?
	err := u.signIn(w, &user)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	alert := views.Alert{
		Level:   views.AlertSuccess,
		Message: "Welcome to the Monkey Show",
	}
	views.RedirectAlert(w, r, "/galleries", http.StatusFound, alert)
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form LoginForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}
	user, err := u.us.Auth(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.AlertError("Invalid email provided")
		default:
			vd.SetAlert(err)

		}
		u.LoginView.Render(w, r, vd)
		return
	}
	err = u.signIn(w, user)
	if err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	user := context.User(r.Context())
	token, _ := rand.RememberToken()
	user.Remember = token
	u.us.Update(user)
	http.Redirect(w, r, "/", http.StatusFound)
}
func (u *Users) signIn(w http.ResponseWriter, user *models.Users) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = u.us.Update(user)
	}

	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return nil
}
