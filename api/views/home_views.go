package views

import(
	"html/template"
	"net/http"
)

func RenderHomeTemplate(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/home.html"))

	tmpl.Execute(w, nil)
}