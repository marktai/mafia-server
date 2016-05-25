package game

import (
	"database/sql"
	"db"
	"sort"
)

type Player struct {
	GameID   uint
	PlayerID uint
	Name     string
	role     uint // private to not show in info
	Alive    bool
}

type PlayerIDRole struct {
	PlayerID uint
	Role     uint
}
type Players []*Player

//sorts by most recent first
func (a Players) Len() int           { return len(a) }
func (a Players) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Players) Less(i, j int) bool { return a[i].PlayerID < a[j].PlayerID }

// Creates a new player and uploads it to the database
// default player has no name and no role
func MakePlayer(gameID uint) (*Player, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	var p Player
	p.GameID = gameID
	p.PlayerID, err = getUniquePlayerID()
	if err != nil {
		return nil, err
	}
	p.Name = ""
	p.role = 0
	p.Alive = true

	_, err = p.Upload()

	if err != nil {
		return nil, err
	}

	return &p, nil
}

// func UpdatePlayer(gameID uint, name string) (*Player, error) {
// 	err := db.Db.Ping()
// 	if err != nil {
// 		return nil, err
// 	}

// 	var p Player
// 	p.GameID = gameID
// 	p.PlayerID, err = getUniquePlayerID()
// 	if err != nil {
// 		return nil, err
// 	}
// 	p.Name = name
// 	p.role = 0
// 	p.Alive = true

// 	_, err = p.Update()

// 	if err != nil {
// 		return nil, err
// 	}

// 	return &p, nil

// }

// updates database version of the game
func (p *Player) Upload() (sql.Result, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	addGame, err := db.Db.Prepare("INSERT INTO players (gameid, playerid, name, role, alive) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	return addGame.Exec(p.GameID, p.PlayerID, p.Name, p.role, p.Alive)
}

// updates database version of the game
func (p *Player) Update() (sql.Result, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	updateGame, err := db.Db.Prepare("UPDATE players SET name=?, role=?, alive=? WHERE gameid=? AND playerid=?")
	if err != nil {
		return nil, err
	}

	return updateGame.Exec(p.Name, p.role, p.Alive, p.GameID, p.PlayerID)
}

// gets all players in a specific game
// sorted by playerID
func GetGamePlayers(game uint) (Players, error) {
	err := db.Db.Ping()
	if err != nil {
		return nil, err
	}

	players := make(Players, 0)

	rows, err := db.Db.Query("SELECT playerid, name, role, alive FROM players WHERE gameid=?", game)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var player Player
		player.GameID = game
		if err := rows.Scan(&player.PlayerID, &player.Name, &player.role, &player.Alive); err != nil {
			return nil, err
		}
		players = append(players, &player)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Sort(players)

	return players, nil
}

func (p *Player) PlayerIDRole() PlayerIDRole {
	return PlayerIDRole{p.PlayerID, p.role}
}
