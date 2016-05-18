package game

import (
	"database/sql"
	"db"
	"errors"
	"time"
)

//golang constant thingy
//reference time is "Mon Jan 2 15:04:05 -0700 MST 2006"
const sqlForm = "2006-01-02 15:04:05"

// Game is the class that represents all of the mafia game data

type GameOptions struct {
	PlayerCount uint
}

// Metadata about the game
type Game struct {
	GameID    uint
	Stage     uint
	Started   time.Time
	Modified  time.Time
	TurnCount uint
	Players   Players
	Moves     Moves
	Options   GameOptions
}

// Creates a new game and uploads it to the database
func MakeGame(options GameOptions) (*Game, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	var g Game
	g.GameID, err = getUniqueGameID()
	if err != nil {
		return nil, err
	}
	g.Stage = 0
	g.Started = time.Now().UTC()
	g.Modified = time.Now().UTC()
	g.TurnCount = 0
	g.Players = make(Players, options.PlayerCount)
	for i, _ := range g.Players {
		g.Players[i], err = MakePlayer(g.GameID)
		if err != nil {
			return nil, err
		}
	}

	g.Moves = make(Moves, 0)
	g.Options = options

	_, err = g.Upload()

	if err != nil {
		return nil, err
	}

	return &g, nil
}

// Gets a Game frome the database
func GetGame(gameID uint) (*Game, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	var game Game
	game.GameID = gameID

	var started, modified string

	//TODO: handle NULLS
	err = db.Db.QueryRow("SELECT stage, started, modified, turnCount FROM games WHERE gameid=?", gameID).Scan(&game.Stage, &started, &modified, &game.TurnCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Game not found")
		}
		return nil, err
	}

	game.Started, err = time.Parse(sqlForm, started)
	if err != nil {
		return nil, err
	}
	game.Modified, err = time.Parse(sqlForm, modified)
	if err != nil {
		return nil, err
	}

	game.Players, err = GetGamePlayers(gameID)
	if err != nil {
		return nil, err
	}

	game.Moves, err = GetGameMoves(gameID)
	if err != nil {
		return nil, err
	}

	return &game, nil
}

// updates database version of the game
func (g *Game) Upload() (sql.Result, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	addGame, err := db.Db.Prepare("INSERT INTO games (gameid, stage, started, modified, turncount) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	return addGame.Exec(g.GameID, g.Stage, g.Started, g.Modified, g.TurnCount)
}

// updates database version of the game
func (g *Game) Update() (sql.Result, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	updateGame, err := db.Db.Prepare("UPDATE games SET stage=?, started=?, modified=?, turnCount=? WHERE gameid=?")
	if err != nil {
		return nil, err
	}

	return updateGame.Exec(g.Stage, g.Started, g.Modified, g.TurnCount, g.GameID)
}

/*
// Validates and then makes a move
func (g *Game) MakeMove(player, box, square uint) error {
	if g.Board.Box().CheckOwned() != 0 {
		return errors.New("Game already finished")
	}

	playerTurn := g.Turn / 10 % 2
	if player != g.Players[playerTurn] {
		return errors.New("Not player's turn")
	}

	moveBox := g.Turn % 10

	if moveBox != 9 && box != moveBox {
		return errors.New("Not correct box")
	}

	if box > 8 {
		return errors.New("Box out of range")
	}

	if g.Board[box].Owned != 0 {
		return errors.New("Box already taken")
	}

	if square > 8 {
		return errors.New("Square out of range")
	}

	err := g.Board[box].MakeMove(playerTurn+1, square)
	if err != nil {
		return err
	}

	g.MoveHistory.AddMove(9*box + square)
	g.Modified = time.Now().UTC()

	g.Turn = (1 - playerTurn) * 10
	if g.Board[square].Owned != 0 {
		g.Turn += 9
	} else {
		g.Turn += square
	}

	g.CheckVictor()

	return nil
}
*/
