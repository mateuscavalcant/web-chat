package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func echo(w http.ResponseWriter, r *http.Request) {
    // Upgrade the HTTP connection to a WebSocket connection
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    defer conn.Close()

    for {
        // Read message from the browser
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Println(err)
            return
        }

        // Print the message to the console
        log.Printf("Received message: %s", p)

        // Write message back to the browser
        if err := conn.WriteMessage(messageType, p); err != nil {
            log.Println(err)
            return
        }
    }
}

func renderTemplate(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	
	tmpl.Execute(w, r)
}

func main() {

	fmt.Println("Server ON...")
	http.HandleFunc("/", renderTemplate)

    http.HandleFunc("/echo", echo)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
