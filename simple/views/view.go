package views

import (
	"bytes"
	"github.com/gorilla/csrf"
	"github.com/pkg/errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"webapp1/simple/context"
)

var (
	LayoutDir   string = "views/layouts/"
	TemplateDir string = "Views/"
	TemplareExt string = ".gohtml"
)

func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)
	files = append(files, layoutFiles()...)
	t, err := template.New("").Funcs(
		template.FuncMap{
			"csrfField": func() (template.HTML, error) {
				return "", errors.New("crsfField not implemented")
			},
		}).ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	return &View{
		Template: t,
		Layout:   layout,
	}
}

type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	var vd Data
	switch d := data.(type) {
	case Data:
		vd = d
	default:
		vd = Data{
			Yield: data,
		}
	}
	if alert := getAlert(r); alert != nil && vd.Alert == nil {
		vd.Alert = alert
		clearAlert(w)
	}
	vd.User = context.User(r.Context())
	var buf bytes.Buffer
	csrfField := csrf.TemplateField(r)
	tpl := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfField
		},
	})
	if err := tpl.ExecuteTemplate(&buf, v.Layout, data); err != nil {
		log.Println(err)
		http.Error(w, "oops something went wrong try again!", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
}
func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplareExt)
	if err != nil {
		panic(err)
	}
	return files
}

func addTemplatePath(files []string) {
	for i, n := range files {
		files[i] = TemplateDir + n
	}
}
func addTemplateExt(files []string) {
	for i, n := range files {
		files[i] = n + TemplareExt
	}

}
