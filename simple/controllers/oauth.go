package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"time"
	llctx "webapp1/simple/context"
	"webapp1/simple/models"
)

func NewOauths(os models.OAuthService, configs map[string]*oauth2.Config) *OAuths {
	return &OAuths{
		os:      os,
		configs: configs,
	}
}

type OAuths struct {
	os      models.OAuthService
	configs map[string]*oauth2.Config
}

func (o *OAuths) Connect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service := vars["service"]
	oAuthConfig, ok := o.configs[service]
	if !ok {
		http.Error(w, "Invalid OAuth2 service", http.StatusBadRequest)
		return
	}
	state := csrf.Token(r)
	cookie := http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Expires:  time.Now().Add(5 * time.Minute),
	}
	http.SetCookie(w, &cookie)

	url := oAuthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}
func (o *OAuths) Callback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service := vars["service"]
	oAuthConfig, ok := o.configs[service]
	if !ok {
		http.Error(w, "Invalid OAuth2 service", http.StatusBadRequest)
		return
	}
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
	token, err := oAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user := llctx.User(r.Context())
	existin, err := o.os.Find(user.ID, service)
	if err == models.ErrNotFound {

	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		o.os.Delete(existin.ID)
	}
	oauthU := models.OAuth{
		UserID:  user.ID,
		Token:   *token,
		Service: service,
	}
	err = o.os.Create(&oauthU)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%+v", token)
	fmt.Fprint(w, r.FormValue("code"), r.FormValue("state"))
}
func (o *OAuths) DropboxTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service := vars["service"]
	oAuthConfig, ok := o.configs[service]
	if !ok {
		http.Error(w, "Invalid OAuth2 service", http.StatusBadRequest)
		return
	}
	r.ParseForm()
	path := r.FormValue("path")
	user := llctx.User(r.Context())
	userOAuth, err := o.os.Find(user.ID, service)
	if err != nil {
		panic(err)
	}
	token := userOAuth.Token

	data := struct {
		Path string `json:"path"`
	}{
		Path: path,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	client := oAuthConfig.Client(context.Background(), &token)
	req, err := http.NewRequest(http.MethodPost, "https://api.dopboxapi.com/2/files/list_folder",
		bytes.NewReader(dataBytes))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)

}
