package game

import (
	"database/sql"
	"db"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"
	"ws"
)

//golang constant thingy
//reference time is "Mon Jan 2 15:04:05 -0700 MST 2006"
const sqlForm = "2006-01-02 15:04:05"

// Game is the class that represents all of the mafia game data

// Stored as a 32 bit int
type GameOptions struct {
	PlayerCount        uint // 6 bits
	MafiaCount         uint // 4 bits
	DoctorCount        uint // 2 bits
	SherriffCount      uint // 2 bits
	DayTimeIntervals   uint // 8 bits
	NightTimeIntervals uint // 8 bits
}

// time intervals are mulitples of 15 seconds

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

func (o *GameOptions) VillagerCount() uint {
	return o.PlayerCount - o.MafiaCount - o.DoctorCount - o.SherriffCount
}

// Metadata about the game
type Game struct {
	GameID      uint
	Stage       int
	Started     time.Time
	Modified    time.Time
	StageFinish time.Time
	TurnCount   uint
	Players     Players
	Moves       Moves
	Options     GameOptions
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
	g.Stage = -1
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
// Only updates game and not players or moves
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
func (g *Game) MakeGameMove(playerID uint, targetID uint, moveType uint) (map[string]interface{}, error) {

	p, err := g.FindPlayerWithID(playerID)
	if err != nil {
		return nil, err
	}

	if !p.Alive {
		return nil, errors.New("Dead players cannot move")
	}

	if targetID != 0 { // no target for 0
		_, err = g.FindPlayerWithID(targetID)
		if err != nil {
			return nil, err
		}
	}

	if g.Stage == 1 {
		if p.Role != moveType {
			return nil, errors.New("Invalid move type, wrong role")
		}
	} else if g.Stage == 2 {
		if moveType != 0 {
			return nil, errors.New("Invalid move type, not vote")
		}
	}

	createdMoves, err := GetGamePlayerTurnMoves(g.GameID, playerID, g.TurnCount)
	if err != nil {
		return nil, err
	} else if len(createdMoves) == 0 {
		// if the player has not yet made a move
		move, err := MakeMove(g.GameID, g.TurnCount, playerID, targetID, moveType)
		if err != nil {
			return nil, err
		}
		g.Moves = append(Moves{move}, g.Moves...) //prepend
	} else {
		// if the player already made a move
		if p.Role == 4 { // sherriff
			return nil, errors.New("Sherriff cannot change his move")
		}
		for _, move := range g.Moves {
			if move.PlayerID == playerID {

				move.TargetID = targetID
				move.Type = moveType

				_, err := move.Update()
				if err != nil {
					return nil, err
				}

				break
			}
		}
	}

	retMap := make(map[string]interface{})

	// if a role has an immediate action
	if targetID != 0 {
		if moveType == 4 && g.Stage == 1 {
			mafiaCheck, err := g.ProcessSherriffMove(targetID)
			if err != nil {
				return nil, err
			}
			retMap["Sherriff"] = mafiaCheck
		}
	}

	moves, err := GetGameTurnMoves(g.GameID, g.TurnCount)
	if err != nil {
		return nil, err
	}

	playerMoveMap := make(map[uint]struct{})
	allPlayersMoved := true

	for _, move := range moves {
		playerMoveMap[move.PlayerID] = struct{}{}
	}

	for _, player := range g.Players {
		if _, ok := playerMoveMap[player.PlayerID]; !ok {

			// dead players cant move
			if !player.Alive {
				continue
			}

			allPlayersMoved = false
			break
		}
	}

	if allPlayersMoved {
		g.ProgressStage()
	}

	g.Modified = time.Now().UTC()
	_, err = g.Update()
	if err != nil {
		return nil, err
	}

	return retMap, nil
}

func (g *Game) ProcessSherriffMove(targetID uint) (bool, error) {
	p, err := g.FindPlayerWithID(targetID)
	if err != nil {
		return false, err
	}

	return p.Role == 2, nil
}

func (g *Game) GetCurrentMoves() (Moves, error) {
	return GetGameTurnMoves(g.GameID, g.TurnCount)
}

func (g *Game) RegisterPlayer(name string) error {
	unnamedCount := 0
	var err error
	registeredPlayerCheck := false
	var emptyPlayer *Player

	for _, player := range g.Players {
		if player.Name == "" {
			if !registeredPlayerCheck {
				registeredPlayerCheck = true
				emptyPlayer = player
				continue
			}
			unnamedCount += 1
		}

		if player.Name == name {
			return errors.New("Cannot have two players with the same name")
		}
	}

	emptyPlayer.Name = name
	emptyPlayer.Role, err = g.GenerateRole()
	if err != nil {
		return err
	}
	_, err = emptyPlayer.Update()
	if err != nil {
		return err
	}

	// If all players have registered, start game
	if unnamedCount == 0 {
		g.ProgressStage()
	}

	g.Modified = time.Now().UTC()
	_, err = g.Update()
	if err != nil {
		return err
	}
	return nil

}

func (g *Game) GenerateRole() (uint, error) {

	roleCounts := make([]uint, 5)
	// 0 is how many players are generated
	// 1 is villager
	// 2 is mafia
	// 3 is doctor
	// 4 is sherriff

	for _, player := range g.Players {
		if player.Role == 0 {
			continue
		} else {
			roleCounts[0] += 1
			roleCounts[player.Role] += 1
		}
	}

	roleCounts[0] = g.Options.PlayerCount - roleCounts[0]
	roleCounts[1] = g.Options.VillagerCount() - roleCounts[1]
	roleCounts[2] = g.Options.MafiaCount - roleCounts[2]
	roleCounts[3] = g.Options.DoctorCount - roleCounts[3]
	roleCounts[4] = g.Options.SherriffCount - roleCounts[4]

	// 0 is how many players are left to be assigned roles
	// 1 is villagers left to be assigned roles
	// 2 is mafias left to be assigned roles
	// 3 is doctors left to be assigned roles
	// 4 is sherriffs left to be assigned roles

	if roleCounts[0] <= 0 {
		return 0, errors.New("No new roles to be generated")
	}

	randRole := uint(rand.Intn(int(roleCounts[0])))
	for i, roleCount := range roleCounts {
		if i == 0 {
			continue // ignore the first index
		}
		if randRole < roleCount {
			return uint(i), nil
		}
		randRole -= roleCount
	}

	return 1, errors.New("Shouldn't really get here tbh")
}

// kills necessary people and moves stage to next
func (g *Game) ProgressStage() error {
	var ret int
	var err error
	if g.Stage == -1 { // start of game
		g.Stage = 1
		g.StageFinish = time.Now().Add(time.Duration(g.Options.NightTimeIntervals) * 15 * time.Second)
	} else if g.Stage == 1 { // night
		ret, err = g.processNight()
		log.Println("Night returned", ret)
		if err != nil {
			return err
		}
	} else if g.Stage == 2 { // day
		ret, err = g.processDay()
		log.Println("Day returned", ret)
		if err != nil {
			return err
		}
	}

	g.TurnCount += 1

	if g.CheckFinish() {
		err = ws.BroadcastEvent(g.GameID, "Victory", g.Stage)
		if err != nil {
			log.Println(err)
		}
	} else {

		err = ws.BroadcastEvent(g.GameID, "Turn", g.TurnCount)
		if err != nil {
			log.Println(err)
		}
	}

	_, err = g.Update()
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) processNight() (int, error) {
	var returnCode int = 0

	moves, err := GetGameTurnMoves(g.GameID, g.TurnCount)
	if err != nil {
		return returnCode, err
	}
	doctorVoteCounts := make(map[uint]uint)
	mafiaVoteCounts := make(map[uint]uint)
	for _, move := range moves {
		player, err := g.FindPlayerWithID(move.PlayerID)
		if err != nil {
			return returnCode, err
		}
		if player.Role != move.Type { // bad move by player
			continue
		}
		if move.Type == 2 {
			// mafia
			if _, ok := mafiaVoteCounts[move.TargetID]; !ok {
				mafiaVoteCounts[move.TargetID] = 0
			}
			mafiaVoteCounts[move.TargetID] += 1
		}

		if move.Type == 3 {
			// doctor
			if _, ok := doctorVoteCounts[move.TargetID]; !ok {
				doctorVoteCounts[move.TargetID] = 0
			}
			doctorVoteCounts[move.TargetID] += 1
		}
	}

	var mafiaMajorityTarget uint
	var mafiaMajorityCount uint

	for target, count := range mafiaVoteCounts {
		if count > mafiaMajorityCount {
			mafiaMajorityCount = count
			mafiaMajorityTarget = target
		}
	}

	var doctorMajorityTarget uint
	var doctorMajorityCount uint

	for target, count := range doctorVoteCounts {
		if count > doctorMajorityCount {
			doctorMajorityCount = count
			doctorMajorityTarget = target
		}
	}

	if mafiaMajorityTarget != doctorMajorityTarget {
		p, err := g.FindPlayerWithID(mafiaMajorityTarget)
		if err != nil {
			return returnCode, err
		}
		p.Alive = false
		_, err = p.Update()
		if err != nil {
			return returnCode, err
		}

		returnCode = 1 // 1 for successful mafia kill
	} else {

		returnCode = 2 // 2 for doctor save
	}

	g.Stage = 2
	g.StageFinish = time.Now().Add(time.Duration(g.Options.DayTimeIntervals) * 15 * time.Second)

	return returnCode, nil
}

func (g *Game) processDay() (int, error) {
	var returnCode int = 0

	moves, err := GetGameTurnMoves(g.GameID, g.TurnCount)
	if err != nil {
		return returnCode, err
	}
	lynchVoteCounts := make(map[uint]uint)
	for _, move := range moves {
		if move.Type == 0 {
			// lynch
			if _, ok := lynchVoteCounts[move.TargetID]; !ok {
				lynchVoteCounts[move.TargetID] = 0
			}
			lynchVoteCounts[move.TargetID] += 1
		}

	}

	var lynchMajorityTarget uint
	var lynchMajorityCount uint
	lynchMajorityAchieved := true

	log.Println(lynchVoteCounts)

	for target, count := range lynchVoteCounts {
		if count > lynchMajorityCount {
			lynchMajorityCount = count
			lynchMajorityTarget = target
			lynchMajorityAchieved = true
		}
		if count == lynchMajorityCount {
			lynchMajorityAchieved = false
		}
	}

	log.Println(lynchMajorityTarget)
	if !lynchMajorityAchieved {
		returnCode = 2 // 2 for no majority
	} else if lynchMajorityTarget != 0 {
		p, err := g.FindPlayerWithID(lynchMajorityTarget)
		if err != nil {
			return returnCode, err
		}
		p.Alive = false
		_, err = p.Update()
		if err != nil {
			return returnCode, err
		}
		returnCode = 1 // 1 for successful  kill
	} else {
		returnCode = 3 // 3 for no kill
	}

	g.Stage = 1
	g.StageFinish = time.Now().Add(time.Duration(g.Options.NightTimeIntervals) * 15 * time.Second)

	return returnCode, nil
}

func (g *Game) CheckFinish() bool {
	mafiaWin := true
	townWin := true
	for _, player := range g.Players {

		if player.Role == 2 {
			// if any mafia alive, town has not won yet
			townWin = false
		} else {
			// if anyone but mafia alive, mafia has not won yet
			mafiaWin = false
		}
	}
	if townWin {
		g.Stage = 11
	}

	if mafiaWin {
		g.Stage = 12
	}

	return townWin || mafiaWin
}

func (g *Game) FindPlayerWithID(playerID uint) (*Player, error) {
	for _, player := range g.Players {
		if player.PlayerID == playerID {
			return player, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("PlayerID %d not found in GameID %d", playerID, g.GameID))
}

func (g *Game) PlayerMap() map[string]uint {
	playerMap := make(map[string]uint)
	for _, player := range g.Players {
		if player == nil || player.Name == "" {
			continue
		}
		playerMap[player.Name] = player.PlayerID
	}
	return playerMap
}
