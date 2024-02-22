package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	clients = make(map[*websocket.Conn]*Client)
)

func handleMessages(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	client := &Client{conn}
	clients[conn] = client

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(clients, conn)
			return
		}

		for _, c := range clients {
			if c.conn != conn {
				if err := c.conn.WriteMessage(messageType, p); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}
func renderTemplate(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	
	tmpl.Execute(w, nil)
}

func main() {

	fmt.Println("Server ON...")
	http.HandleFunc("/", renderTemplate)
	http.HandleFunc("/chat", handleMessages)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
