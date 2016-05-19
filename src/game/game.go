package game

import (
	"database/sql"
	"db"
	"errors"
	"fmt"
	"time"
)

//golang constant thingy
//reference time is "Mon Jan 2 15:04:05 -0700 MST 2006"
const sqlForm = "2006-01-02 15:04:05"

// Game is the class that represents all of the mafia game data

// Stored as a 32 bit int
type GameOptions struct {
	PlayerCount   uint // 6 bits
	MafiaCount    uint // 4 bits
	DoctorCount   uint // 2 bits
	SherriffCount uint // 2 bits
}

// bit sizes of every field
var GameOptionSizes = GameOptions{
	PlayerCount:   6,
	MafiaCount:    4,
	DoctorCount:   2,
	SherriffCount: 2,
}

func (o *GameOptions) Verify() error {
	var max uint
	max = 1 << GameOptionSizes.PlayerCount
	if o.PlayerCount >= max {
		return errors.New(fmt.Sprintf("PlayerCount is too large. Max is %d", max-1))
	}

	max = 1 << GameOptionSizes.MafiaCount
	if o.MafiaCount >= max {
		return errors.New(fmt.Sprintf("MafiaCount is too large. Max is %d", max-1))
	}

	max = 1 << GameOptionSizes.DoctorCount
	if o.DoctorCount >= max {
		return errors.New(fmt.Sprintf("DoctorCount is too large. Max is %d", max-1))
	}

	max = 1 << GameOptionSizes.SherriffCount
	if o.SherriffCount >= max {
		return errors.New(fmt.Sprintf("SherriffCount is too large. Max is %d", max-1))
	}

	return nil
}

func (o *GameOptions) Encode() (uint, error) {
	err := o.Verify()
	if err != nil {
		return 0, err
	}

	var total uint = 0

	total <<= GameOptionSizes.PlayerCount
	total += o.PlayerCount

	total <<= GameOptionSizes.MafiaCount
	total += o.MafiaCount

	total <<= GameOptionSizes.DoctorCount
	total += o.DoctorCount

	total <<= GameOptionSizes.SherriffCount
	total += o.SherriffCount

	return total, nil
}

func GetLastNBits(a, n uint) uint {
	var mask uint = 0
	var i uint = 0
	for ; i < n; i++ {
		mask |= 1 << i
	}
	return a & mask
}

func DecodeGameOptions(encoded uint) (*GameOptions, error) {
	var retOptions GameOptions

	retOptions.SherriffCount = GetLastNBits(encoded, GameOptionSizes.SherriffCount)
	encoded >>= GameOptionSizes.SherriffCount

	retOptions.DoctorCount = GetLastNBits(encoded, GameOptionSizes.DoctorCount)
	encoded >>= GameOptionSizes.DoctorCount

	retOptions.MafiaCount = GetLastNBits(encoded, GameOptionSizes.MafiaCount)
	encoded >>= GameOptionSizes.MafiaCount

	retOptions.PlayerCount = GetLastNBits(encoded, GameOptionSizes.PlayerCount)
	encoded >>= GameOptionSizes.PlayerCount

	if encoded != 0 {
		return nil, errors.New("Encoded GameOption has too many bits")
	}

	return &retOptions, nil
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
	var encodedOptions uint

	//TODO: handle NULLS
	err = db.Db.QueryRow("SELECT stage, started, modified, turnCount, options FROM games WHERE gameid=?", gameID).Scan(&game.Stage, &started, &modified, &game.TurnCount, &encodedOptions)
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

	options, err := DecodeGameOptions(encodedOptions)
	if err != nil {
		return nil, err
	}

	game.Options = *options

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
	encodedOptions, err := g.Options.Encode()
	if err != nil {
		return nil, err
	}

	err = db.Db.Ping()
	if err != nil {
		return nil, err
	}

	addGame, err := db.Db.Prepare("INSERT INTO games (gameid, stage, started, modified, turncount, options) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	return addGame.Exec(g.GameID, g.Stage, g.Started, g.Modified, g.TurnCount, encodedOptions)
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

	result, err := updateGame.Exec(g.Stage, g.Started, g.Modified, g.TurnCount, g.GameID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Validates and then makes a move
func (g *Game) MakeGameMove(playerID uint, targetID uint) error {
	var role uint
	for _, player := range g.Players {
		if player.PlayerID == playerID {
			role = player.Role
			break
		}
	}

	move, err := MakeMove(g.GameID, g.TurnCount, playerID, targetID, role)
	if err != nil {
		return err
	}

	// so we dont invalidate the game object
	g.Moves = append(Moves{move}, g.Moves...) //prepend
	g.Modified = time.Now().UTC()
	_, err = g.Upload()
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) GetCurrentMoves() (Moves, error) {
	return GetGameTurnMoves(g.GameID, g.TurnCount)
}
