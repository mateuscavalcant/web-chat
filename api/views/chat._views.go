package views

import (
	"html/template"
	"net/http"
)


func RenderChatTemplate(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/chat.html"))

	tmpl.Execute(w, nil)
}
