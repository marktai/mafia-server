package server

import (
	//	"auth"
	"fmt"
	"game"
	"github.com/gorilla/mux"
	// "io/ioutil"
	//	"errors"
	"log"
	//	"math/rand"
	"net/http"
	"ws"
)

func sexgod(w http.ResponseWriter, r *http.Request) {
	WriteJson(w, genMap("ID", "fuck mark"))
}

func makeGame(w http.ResponseWriter, r *http.Request) {
	var options game.GameOptions
	options.PlayerCount = 5
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

	game, err := game.GetGame(gameID)
	if err != nil {
		WriteError(w, err, 500)
		return
	}

	WriteJson(w, genMap("Info", *game))
}

func getPlayerInfo(w http.ResponseWriter, r *http.Request) {
	WriteJson(w, genMap("Players", []int{123, 456, 789}))
}

func registerPlayer(w http.ResponseWriter, r *http.Request) {
	retMap := make(map[string]int)
	retMap["Jay"] = 123
	retMap["Mark"] = 456
	WriteJson(w, retMap)
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

	retMap := make(map[string]interface{})
	retMap["UserID"] = 123
	retMap["Target"] = 2
	retMap["Role"] = "Mafia"

	WriteJson(w, retMap)
	if err == nil {
		err = ws.BroadcastEvent(gameID, "Change", fmt.Sprintf("Sup"))
		if err != nil {
			log.Println(err)
		}

	}
}

/*
func makeGame(w http.ResponseWriter, r *http.Request) {
	player1, err := stringtoUint(r.FormValue("Player1"))
	if err != nil {
		WriteError(w, errors.New("Error parsing Player1 form value"), 400)
		return
	}

	player2, err := stringtoUint(r.FormValue("Player2"))
	if err != nil {
		player2Username := r.FormValue("Player2")
		if player2Username != "" {
			player2, err = auth.GetUserID(player2Username)
		}
		if err != nil {
			WriteError(w, errors.New("Error parsing Player2 form value"), 400)
			return
		}
	}

	starter, err := stringtoUint(r.FormValue("Starter"))
	// 0 means random
	// 1 means player1 starts
	// 2 means player2 starts
	if err != nil {
		starter = 0
	}

	if starter > 2 {
		WriteError(w, errors.New("Starter must be 0-2"), 400)
		return
	}

	if requireAuth {

		timeInt, path, messageHMACString, encoding, err := auth.ExtractAuthParamsNoUser(r)
		if err != nil {
			WriteError(w, err, 400)
			return
		}

		authed, err := auth.CheckAuthParams(player1, timeInt, path, messageHMACString, encoding)
		if err != nil || !authed {
			if err != nil {
				log.Println(err)
			}
			WriteErrorString(w, "Not Authorized Request", 401)
			return
		}
	}

	// 0 means that 50% chance of switching
	switch starter {
	case 0:
		if rand.Intn(2) == 1 {
			// log.Println("Same order")
			break
		}
		// log.Println("Reversed order")
		fallthrough
	case 2:
		player1, player2 = player2, player1
	}

	game, err := game.MakeGame(player1, player2)

	if err != nil {
		WriteError(w, err, 400)
		return
	}

	WriteJson(w, genMap("ID", game.GameID))
}

func getAllGames(w http.ResponseWriter, r *http.Request) {
	games, err := game.GetAllGames()

	if err != nil {
		WriteError(w, err, 400)
		return
	}

	WriteJson(w, genMap("Games", games))
}

func getGame(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := stringtoUint(vars["ID"])
	if err != nil {
		WriteError(w, err, 400)
		return
	}

	game, err := game.GetGame(id)

	if err != nil {
		WriteError(w, err, 400)
		return
	}

	info, err := game.InfoWithNames()

	if err != nil {
		WriteError(w, err, 500)
	}

	WriteJson(w, genMap("Game", info))
}

func getBoard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := stringtoUint(vars["ID"])
	if err != nil {
		WriteError(w, err, 400)
		return
	}

	game, err := game.GetGame(id)

	if err != nil {
		WriteError(w, err, 400)
		return
	}

	WriteJson(w, genMap("Board", game.Board))
}

func getGameString(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := stringtoUint(vars["ID"])
	if err != nil {
		WriteError(w, err, 400)
		return
	}

	game, err := game.GetGame(id)

	if err != nil {
		WriteError(w, err, 400)
		return
	}

	WriteJson(w, genMap("Board", game.Board.StringArray(true)))
}

func makeGameMove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := stringtoUint(vars["ID"])
	if err != nil {
		WriteError(w, err, 400)
		return
	}

	player, err := stringtoUint(r.FormValue("Player"))
	if err != nil {
		WriteError(w, errors.New("Error parsing Player form value"), 400)
		return
	}

	box, err := stringtoUint(r.FormValue("Box"))
	if err != nil {
		WriteError(w, errors.New("Error parsing Box form value"), 400)
		return
	}

	square, err := stringtoUint(r.FormValue("Square"))
	if err != nil {
		WriteError(w, errors.New("Error parsing Square form value"), 400)
		return
	}

	if requireAuth {

		timeInt, path, messageHMACString, encoding, err := auth.ExtractAuthParamsNoUser(r)
		if err != nil {
			WriteError(w, err, 400)
			return
		}

		authed, err := auth.CheckAuthParams(player, timeInt, path, messageHMACString, encoding)
		if err != nil || !authed {
			if err != nil {
				log.Println(err)
			}
			WriteErrorString(w, "Not Authorized Request", 401)
			return
		}
	}

	game, err := game.GetGame(id)
	if err != nil {
		WriteError(w, err, 400)
		return
	}

	err = game.MakeMove(player, box, square)
	if err != nil {
		WriteError(w, err, 400)
		return
	}

	_, err = game.Update()
	WriteOutputError(w, genMap("Output", "Successful"), err)

	if err == nil {
		err = ws.BroadcastEvent(id, "Change", fmt.Sprintf("Changed %d, %d", box, square))
		if err != nil {
			log.Println(err)
		}

	}
}

func getUserGames(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	userID, err := stringtoUint(vars["userID"])
	if err != nil {
		WriteError(w, err, 400)
		return
	}

	if requireAuth {

		timeInt, path, messageHMACString, encoding, err := auth.ExtractAuthParamsNoUser(r)
		if err != nil {
			WriteError(w, err, 400)
			return
		}

		authed, err := auth.CheckAuthParams(userID, timeInt, path, messageHMACString, encoding)
		if err != nil || !authed {
			if err != nil {
				log.Println(err)
			}
			WriteErrorString(w, "Not Authorized Request", 401)
			return
		}
	}
	games, err := game.GetUserGames(userID)

	if err != nil {
		WriteError(w, err, 400)
		return
	}

	WriteJson(w, genMap("Games", games))
}
*/
