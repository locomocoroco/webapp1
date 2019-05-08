package views

import (
	"log"
	"net/http"
	"time"
	"webapp1/simple/models"
)

const (
	AlertError   = "danger"
	AlertWarning = "warning"
	AlertInfo    = "info"
	AlertSuccess = "success"

	ErrGeneric = "Oops that went wrong, try again if you feel like it."
)

type Alert struct {
	Level   string
	Message string
}
type Data struct {
	Alert *Alert
	User  *models.Users
	Yield interface{}
}

func (d *Data) SetAlert(err error) {
	if pErr, ok := err.(PublicError); ok {
		d.Alert = &Alert{
			Level:   AlertError,
			Message: pErr.Puclic(),
		}
	} else {
		log.Println(err)
		d.Alert = &Alert{
			Level:   AlertError,
			Message: ErrGeneric,
		}
	}
}
func (d *Data) AlertError(msg string) {
	d.Alert = &Alert{
		Level:   AlertError,
		Message: msg,
	}
}

type PublicError interface {
	error
	Puclic() string
}

func persistAlert(w http.ResponseWriter, alert Alert) {
	expiresAt := time.Now().Add(1 * time.Minute)
	lvl := http.Cookie{
		Name:     "alert_level",
		Value:    alert.Level,
		Expires:  expiresAt,
		HttpOnly: true,
	}
	msg := http.Cookie{
		Name:     "alert_message",
		Value:    alert.Message,
		Expires:  expiresAt,
		HttpOnly: true,
	}
	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}
func clearAlert(w http.ResponseWriter) {
	lvl := http.Cookie{
		Name:     "alert_level",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	msg := http.Cookie{
		Name:     "alert_message",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, &lvl)
	http.SetCookie(w, &msg)
}
func getAlert(r *http.Request) *Alert {
	lvl, err := r.Cookie("alert_level")
	if err != nil {
		return nil
	}
	msg, err := r.Cookie("alert_message")
	if err != nil {
		return nil
	}
	return &Alert{
		Level:   lvl.Value,
		Message: msg.Value,
	}
}
func RedirectAlert(w http.ResponseWriter, r *http.Request, urlString string, code int, alert Alert) {
	persistAlert(w, alert)
	http.Redirect(w, r, urlString, code)
}
