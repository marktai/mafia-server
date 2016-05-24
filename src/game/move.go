package game

import (
	"database/sql"
	"db"
	"sort"
	"time"
)

type Move struct {
	GameID    uint
	TurnCount uint
	PlayerID  uint
	TargetID  uint
	Type      uint
	Time      time.Time
}

type Moves []*Move

//sorts by TurnCount and then PlayerID
func (a Moves) Len() int      { return len(a) }
func (a Moves) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Moves) Less(i, j int) bool {
	if a[i].TurnCount == a[j].TurnCount {
		return a[i].PlayerID < a[j].PlayerID
	}
	return a[i].TurnCount < a[j].TurnCount
}

// Creates a new player and uploads it to the database
// default player has no name and no role
func MakeMove(gameID, turnCount, playerID, targetID, moveType uint) (*Move, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}
	var m Move

	m.GameID = gameID
	m.TurnCount = turnCount
	m.PlayerID = playerID
	m.TargetID = targetID
	m.Type = moveType
	m.Time = time.Now().UTC()

	_, err = m.Upload()

	if err != nil {
		return nil, err
	}

	return &m, nil
}

// updates database version of the game
func (m *Move) Update() (sql.Result, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	updateGame, err := db.Db.Prepare("UPDATE moves SET targetid=?, type=?, time=? WHERE gameid=? AND playerid=? AND turncount=?")
	if err != nil {
		return nil, err
	}
	m.Time = time.Now().UTC()

	return updateGame.Exec(m.TargetID, m.Type, m.Time, m.GameID, m.PlayerID, m.TurnCount)
}

// updates database version of the game
func (m *Move) Upload() (sql.Result, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	addGame, err := db.Db.Prepare("INSERT INTO moves (gameid, turncount, playerid, targetid, type, time) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	return addGame.Exec(m.GameID, m.TurnCount, m.PlayerID, m.TargetID, m.Type, m.Time)
}

// wrapper function around GetPlayerMoves to get moves for entire game
// sorted by turn count then playerID
func GetGameMoves(gameID uint) (Moves, error) {
	return GetPlayerMoves(gameID, 0)
}

// gets all moves by a player in a specific game
// sorted by turn count
func GetPlayerMoves(gameID uint, playerID uint) (Moves, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	// don't check for matching player
	if playerID == 0 {
		rows, err = db.Db.Query("SELECT turncount, playerid, targetid, type, time FROM moves WHERE gameid=?", gameID)
	} else { // actually check for matching player
		rows, err = db.Db.Query("SELECT turncount, playerid, targetid, type, time FROM moves WHERE gameid=? AND playerid=?", gameID, playerID)
	}

	return ParseMoveRows(gameID, rows, err)

}

func GetGameTurnMoves(gameID uint, turnCount uint) (Moves, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	rows, err = db.Db.Query("SELECT turncount, playerid, targetid, type, time FROM moves WHERE gameid=? AND turncount=?", gameID, turnCount)
	return ParseMoveRows(gameID, rows, err)
}

func GetGamePlayerTurnMoves(gameID, playerID, turnCount uint) (Moves, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	var rows *sql.Rows

	rows, err = db.Db.Query("SELECT turncount, playerid, targetid, type, time FROM moves WHERE gameid=? AND playerID=? AND turncount=?", gameID, playerID, turnCount)
	return ParseMoveRows(gameID, rows, err)
}

func ParseMoveRows(gameID uint, rows *sql.Rows, err error) (Moves, error) {
	moves := make(Moves, 0)
	var timeString string

	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var move Move
		move.GameID = gameID
		if err := rows.Scan(&move.TurnCount, &move.PlayerID, &move.TargetID, &move.Type, &timeString); err != nil {
			return nil, err
		}

		move.Time, err = time.Parse(sqlForm, timeString)
		if err != nil {
			return nil, err
		}
		moves = append(moves, &move)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Sort(moves)

	return moves, nil
}
