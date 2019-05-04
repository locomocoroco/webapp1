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
