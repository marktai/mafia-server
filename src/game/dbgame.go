package game

import (
	"db"
	"fmt"
	"strings"
	"sync"
)

var mutexMap = make(map[string]*sync.Mutex)

// checks if a id already exists in the database
func checkGameIDConflict(id uint) (bool, error) {
	return checkIDConflict(id, "game")
}

func getUniqueGameID() (uint, error) {
	return getUniqueID("game")
}

// checks if a id already exists in the database
func checkPlayerIDConflict(id uint) (bool, error) {
	return checkIDConflict(id, "player")
}

func getUniquePlayerID() (uint, error) {
	return getUniqueID("player")
}

// gets a unique id for a new game
func getUniqueID(idType string) (uint, error) {

	var key string
	if strings.Contains(strings.ToLower(idType), "game") {
		key = "games"
	} else if strings.Contains(strings.ToLower(idType), "player") {
		key = "players"
	}

	mutex, ok := mutexMap[key]
	if !ok {
		mutexMap[key] = &sync.Mutex{}
		mutex, _ = mutexMap[key]
	}

	var count uint
	var scale uint
	var addConst uint

	var newID uint

	conflict := true

	// this is to prevent race conditions between acquiring ID's and updating the count
	mutex.Lock()
	defer mutex.Unlock()

	err := db.Db.QueryRow(fmt.Sprintf("SELECT count, scale, addConst FROM count WHERE type='%s'", key)).Scan(&count, &scale, &addConst)
	if err != nil {
		return 0, err
	}

	for conflict || newID == 0 {
		count += 1
		newID = (count*scale + addConst) % 65536
		conflict, err = checkIDConflict(newID, idType)
		if err != nil {
			return 0, err
		}
	}

	updateCount, err := db.Db.Prepare(fmt.Sprintf("UPDATE count SET count=? WHERE type='%s'", key))
	if err != nil {
		return newID, err
	}

	_, err = updateCount.Exec(count)
	if err != nil {
		return newID, err
	}

	return newID, nil
}

func checkIDConflict(id uint, idType string) (bool, error) {
	collision := 1
	var key string
	var table string
	if strings.Contains(strings.ToLower(idType), "game") {
		table = "games"
		key = "gameid"
	} else if strings.Contains(strings.ToLower(idType), "player") {
		table = "players"
		key = "playerid"
	}
	err := db.Db.QueryRow(fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s=?)", table, key), id).Scan(&collision)
	return collision != 0, err
}
