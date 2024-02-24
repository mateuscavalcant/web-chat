package utils

import (
	"crypto/rand"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte(generateRandomKey(32))) // Chave de 32 bytes

func generateRandomKey(length int) []byte {
    key := make([]byte, length)
    _, err := rand.Read(key)
    if err != nil {
        log.Fatal("Erro ao gerar a chave aleatória:", err)
    }
    return key
}

// GetSession retrieves the session from the request.
func GetSession(r *http.Request) *sessions.Session {
    session, err := store.Get(r, "session")
    if err != nil {
        log.Println("Erro ao recuperar a sessão:", err)
    }
    return session
}

// AllSessions retrieves all session data.
func AllSessions(r *http.Request) (interface{}, interface{}) {
    session := GetSession(r)
    id := session.Values["id"]
    email := session.Values["email"]
    return id, email
}