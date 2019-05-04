package views

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
	Yield interface{}
}

func (d *Data) SetAlert(err error) {
	if pErr, ok := err.(PublicError); ok {
		d.Alert = &Alert{
			Level:   AlertError,
			Message: pErr.Puclic(),
		}
	} else {
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
