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

	moves := make(Moves, 0)
	var timeString string

	var rows *sql.Rows

	// don't check for matching player
	if playerID == 0 {
		rows, err = db.Db.Query("SELECT turncount, playerid, targetid, type, time FROM moves WHERE gameid=?", gameID)
	} else { // actually check for matching player
		rows, err = db.Db.Query("SELECT turncount, playerid, targetid, type, time FROM moves WHERE gameid=? AND playerid=?", gameID, playerID)
	}

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
