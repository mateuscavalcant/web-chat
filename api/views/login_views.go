package views

import(
	"html/template"
	"net/http"
)

func RenderLoginTemplate(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login.html"))

	tmpl.Execute(w, nil)
}