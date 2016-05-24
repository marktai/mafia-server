package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"time"
	"ws"
)

var requireAuth bool

func Log(handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func Run(port int, disableAuth bool) {
	//start := time.Now()

	rand.Seed(time.Now().UTC().UnixNano())
	r := mux.NewRouter()
	requireAuth = !disableAuth

	// user requests
	// r.HandleFunc("/login", Log(login)).Methods("POST")
	// r.HandleFunc("/verifySecret", Log(verifySecret)).Methods("POST")
	// r.HandleFunc("/users", Log(makeUser)).Methods("POST")

	// unauthorized requests
	// r.HandleFunc("/games", getAllGames).Methods("GET")
	//	r.HandleFunc("/games/{ID}", getGame).Methods("GET") // only for backwards compatibility
	//	r.HandleFunc("/games/{ID}/info", getGame).Methods("GET")
	//	r.HandleFunc("/games/{ID}/board", getBoard).Methods("GET")
	//	r.HandleFunc("/games/{ID}/string", getGameString).Methods("GET")
	//	r.HandleFunc("/hello_world", sexgod).Methods("GET")
	r.HandleFunc("/games", Log(makeGame)).Methods("POST")
	r.HandleFunc("/games/{GameID:[0-9]+}/info", Log(getGameInfo)).Methods("GET")
	r.HandleFunc("/games/{GameID:[0-9]+}/roles/{UserID:[0-9]+}", Log(getRoles)).Methods("GET")
	r.HandleFunc("/games/{GameID:[0-9]+}/ws", Log(ws.ServeWs)).Methods("GET")
	r.HandleFunc("/games/{GameID:[0-9]+}/move", Log(makeMove)).Methods("POST") // only for backwards compatibility
	r.HandleFunc("/games/{GameID:[0-9]+}/deviceRegister", Log(registerPlayer)).Methods("POST")
	r.HandleFunc("/games/{GameID:[0-9]+}/progressStage", Log(progressStage)).Methods("POST")

	//	r.HandleFunc("/games/{ID}/move", Log(makeGameMove)).Methods("POST")
	//	r.HandleFunc("/users/{userID}/games", getUserGames).Methods("GET")

	for {
		log.Printf("Running at 0.0.0.0:%d\n", port)
		log.Println(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), r))
		time.Sleep(1 * time.Second)
	}
}
