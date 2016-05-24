package server

import (
	"encoding/json"
	"game"
	"github.com/gorilla/mux"
	// "log"
	"net/http"
	// "ws"
)

func sexgod(w http.ResponseWriter, r *http.Request) {
	WriteJson(w, genMap("ID", "fuck mark"))
}

func makeGame(w http.ResponseWriter, r *http.Request) {
	var parsedJson map[string]uint
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&parsedJson)
	if err != nil {
		WriteErrorString(w, err.Error()+" in parsing POST body (JSON)", 400)
		return
	}

	playerCount, ok := parsedJson["PlayerCount"]
	if !ok {
		WriteErrorString(w, "PlayerCount (uint) not in POST body (JSON)", 400)
		return
	}

	mafiaCount, ok := parsedJson["MafiaCount"]
	if !ok {
		WriteErrorString(w, "MafiaCount (uint) not in POST body (JSON)", 400)
		return
	}

	doctorCount, ok := parsedJson["DoctorCount"]
	if !ok {
		// WriteErrorString(w, "DoctorCount (uint) not in POST body (JSON)", 400)
		// return
		doctorCount = 0
	}

	sherriffCount, ok := parsedJson["SherriffCount"]
	if !ok {
		// WriteErrorString(w, "SherriffCount (uint) not in POST body (JSON)", 400)
		// return
		sherriffCount = 0
	}

	options := game.GameOptions{
		PlayerCount:   playerCount,
		MafiaCount:    mafiaCount,
		DoctorCount:   doctorCount,
		SherriffCount: sherriffCount,
	}

	err = options.Verify()
	if err != nil {
		WriteError(w, err, 400)
		return
	}

	newGame, err := game.MakeGame(options)
	if err != nil {
		WriteError(w, err, 500)
		return
	}

	WriteJson(w, genMap("GameID", newGame.GameID))
}

func getGameInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID, err := stringtoUint(vars["GameID"])
	if err != nil {
		WriteErrorString(w, "Error parsing Game ID", 400)
		return
	}

	g, err := game.GetGame(gameID)
	if err != nil {
		WriteError(w, err, 500)
		return
	}

	WriteJson(w, genMap("Info", *g))
}

func getPlayerInfo(w http.ResponseWriter, r *http.Request) {
	WriteJson(w, genMap("Players", []int{123, 456, 789}))
}

func registerPlayer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID, err := stringtoUint(vars["GameID"])
	if err != nil {
		WriteErrorString(w, "Error parsing Game ID", 400)
		return
	}

	var parsedJson map[string][]string
	decoder := json.NewDecoder(r.Body)

	err = decoder.Decode(&parsedJson)
	if err != nil {
		WriteErrorString(w, err.Error()+" in parsing POST body (JSON)", 400)
		return
	}

	if _, ok := parsedJson["PlayerNames"]; !ok {
		WriteErrorString(w, "\"PlayerNames\" not in POST body", 400)
		return
	}

	g, err := game.GetGame(gameID)
	if err != nil {
		WriteError(w, err, 500)
		return
	}

	for _, playerName := range parsedJson["PlayerNames"] {
		err = g.RegisterPlayer(playerName)
		if err != nil {
			WriteError(w, err, 500)
			return
		}
	}

	WriteJson(w, g.PlayerMap())
}

func getRoles(w http.ResponseWriter, r *http.Request) {
	WriteJson(w, genMap("Role", "Mafia"))
}

func makeMove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID, err := stringtoUint(vars["GameID"])
	if err != nil {
		WriteErrorString(w, "Error parsing Game ID", 400)
		return
	}

	playerID, err := stringtoUint(r.FormValue("PlayerID"))
	if err != nil {
		WriteErrorString(w, "Error parsing PlayerID (Query)", 400)
		return
	}

	targetID, err := stringtoUint(r.FormValue("TargetID"))
	if err != nil {
		WriteErrorString(w, "Error parsing TargetID (Query)", 400)
		return
	}

	role, err := stringtoUint(r.FormValue("MoveType"))
	if err != nil {
		WriteErrorString(w, "Error parsing MoveType (Query)", 400)
		return
	}

	g, err := game.GetGame(gameID)
	if err != nil {
		WriteError(w, err, 500)
		return
	}

	retMap, err := g.MakeGameMove(playerID, targetID, role)
	if err != nil {
		WriteError(w, err, 500)
		return
	}

	WriteJson(w, genMap("Result", retMap))

	// if err == nil {
	// 	err = ws.BroadcastEvent(gameID, "Move", playerID)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// }
}

func progressStage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID, err := stringtoUint(vars["GameID"])
	if err != nil {
		WriteErrorString(w, "Error parsing Game ID", 400)
		return
	}

	g, err := game.GetGame(gameID)
	if err != nil {
		WriteError(w, err, 500)
		return
	}

	err = g.ProgressStage()
	if err != nil {
		WriteError(w, err, 500)
		return
	}

	w.WriteHeader(200)
}
