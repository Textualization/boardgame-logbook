package main

import (
	"fmt"
	"time"
	
	"github.com/maxence-charriere/go-app/v7/pkg/app"
	"github.com/maxence-charriere/go-app/v7/pkg/errors"
)

type player struct {
	ID int
	Text string
	Hidden bool
}

type board struct {
	ID int
	Text string
	Hidden bool
}

type score struct {
	Player int
	Game int
	Score float32
}

type game struct {
	ID int
	Board int
	Session int
}

type session struct {
	ID int
	Date int64
}

func getCount(key string) (int, error) {
	count := 0
	if err := app.LocalStorage.Get(key + "-count", &count); err != nil {
		return 0, errors.Newf("error fetching %v count", key).Wrap(err)
	}
	return count, nil
}

func getSessionCount() (int, error) { return getCount("session") }
func getPlayerCount() (int, error) { return getCount("player") }
func getBoardCount() (int, error) { return getCount("board") }
func getGameCount() (int, error) { return getCount("game") }

func incCount(key string) (int, error) {
	count, err := getCount(key)
	if err != nil {
		return count, err
	}
	if err = app.LocalStorage.Set(key + "-count", count + 1); err != nil {
		return 0, errors.Newf("error increasing %v count", key).Wrap(err)
	}
	return count, nil
}

func incSessionCount() (int, error) { return incCount("session") }
func incPlayerCount()  (int, error) { return incCount("player") }
func incBoardCount()   (int, error) { return incCount("board") }
func incGameCount()    (int, error) { return incCount("game") }

func newSession() (session, error) {
	currentTime := time.Now().Unix()
	
	sessionID, err := incSessionCount()
	if err != nil {
		return session{}, errors.New("error storing session count").Wrap(err)
	}
	Session := session {
		ID: sessionID,
		Date: currentTime,
	}
	return Session, Session.store()
}

func (s session) store() error {
	if err := app.LocalStorage.Set(fmt.Sprintf("session-%v", s.ID), s); err != nil {
		return errors.New("error storing session").Wrap(err)
	}
	return nil
}

func retrieveSession(ID int) (session, error) {
	Session := session{}
	if err := app.LocalStorage.Get(fmt.Sprintf("session-%v", ID), &Session); err != nil {
		return session{}, errors.Newf("error fetching session %v", ID).Wrap(err)
	}
	return Session, nil
}

func retrieveGame(ID int) (game, error) {
	Game := game{}
	if err := app.LocalStorage.Get(fmt.Sprintf("game-%v", ID), &Game); err != nil {
		return game{}, errors.Newf("error fetching game %v", ID).Wrap(err)
	}
	return Game, nil
}

func retrievePlayer(ID int) (player, error) {
	Player := player{}
	if err := app.LocalStorage.Get(fmt.Sprintf("player-%v", ID), &Player); err != nil {
		return player{}, errors.Newf("error fetching player %v", ID).Wrap(err)
	}
	return Player, nil
}

func retrieveBoard(ID int) (board, error) {
	Board := board{}
	if err := app.LocalStorage.Get(fmt.Sprintf("board-%v", ID), &Board); err != nil {
		return Board, errors.Newf("error fetching board %v",ID).Wrap(err)
	}
	return Board, nil
}

func retrieveGamesInSession(ID int) ([]game, error) {
	GameIDs := make([]int, 0)
	if err := app.LocalStorage.Get(fmt.Sprintf("session-%v-games", ID), &GameIDs); err != nil {
		return nil, errors.New("error fetching session games").Wrap(err)
	}
	Games := make([]game, len(GameIDs))

	for idx, id := range GameIDs {
		if err := app.LocalStorage.Get(fmt.Sprintf("game-%v", id), &Games[idx]); err != nil {
			return Games, errors.Newf("error fetching game %v for session %v", id, ID).Wrap(err)
		}
	}
	return Games, nil
}

func retrieveScoresInGame(ID int) ([]score, error) {
	ScoreMap := make(map[int]float32)
	if err := app.LocalStorage.Get(fmt.Sprintf("game-%v-scores", ID), &ScoreMap); err != nil {
		return nil, errors.New("error fetching game scores").Wrap(err)
	}
	Scores := make([]score, 0)

	app.Log(fmt.Sprintf("scores in game %v: %v", ID, len(ScoreMap)))
	
	for Player, Score := range ScoreMap {
		app.Log(fmt.Sprintf("Player: %v", Player))
		app.Log(fmt.Sprintf("Score: %v", Score))		
		Scores = append(Scores, score{
			Player: Player,
			Game: ID,
			Score: Score,
		})
	}
	return Scores, nil
}

func retrieveScoresInGameMap(ID int) (map[int]float32, error) {
	ScoreMap := make(map[int]float32)
	if err := app.LocalStorage.Get(fmt.Sprintf("game-%v-scores", ID), &ScoreMap); err != nil {
		return nil, errors.New("error fetching game scores").Wrap(err)
	}
	return ScoreMap, nil
}

func retrieveAllSessions() ([]session, error) {
	sessions, err := getSessionCount()
	if err != nil {
		return nil, errors.New("error fetching session count").Wrap(err)
	}
	AllSessions := make([]session, sessions)
	for idx, _ := range AllSessions {
		if AllSessions[idx], err = retrieveSession(idx); err != nil {
			return AllSessions, errors.Newf("error fetching session %v", idx).Wrap(err)
		}
	}
	return AllSessions, nil
}

func retrieveAllBoards() ([]board, error) {
	boards, err := getBoardCount()
	if err != nil {
		return nil, errors.New("error fetching board count").Wrap(err)
	}
	AllBoards := make([]board, boards)
	for idx, _ := range AllBoards {
		if AllBoards[idx], err = retrieveBoard(idx); err != nil {
			return AllBoards, errors.Newf("error fetching board %v", idx).Wrap(err)
		}
	}
	return AllBoards, nil
}

func retrieveAllPlayers() ([]player, error) {
	players, err := getPlayerCount()
	if err != nil {
		return nil, errors.New("error fetching player count").Wrap(err)
	}
	AllPlayers := make([]player, players)
	for idx, _ := range AllPlayers {
		if AllPlayers[idx], err = retrievePlayer(idx); err != nil {
			return AllPlayers, errors.Newf("error fetching player %v", idx).Wrap(err)
		}
	}
	return AllPlayers, nil
}

func newBoard(text string) (board, error) {
	ID, err := incBoardCount()
	if err != nil {
		return board{}, err
	}
	Board := board{
		ID: ID,
		Text: text,
	}
	return Board, Board.store()
}

func (b board) store() error {
	if err := app.LocalStorage.Set(fmt.Sprintf("board-%v", b.ID), b); err != nil {
		return errors.New("error storing board").Wrap(err)
	}
	return nil
}

func newPlayer(text string) (player, error) {
	ID, err := incPlayerCount()
	if err != nil {
		return player{}, err
	}
	Player := player{
		ID: ID,
		Text: text,
	}
	return Player, Player.store()
}

func (p player) store() error {
	if err := app.LocalStorage.Set(fmt.Sprintf("player-%v", p.ID), p); err != nil {
		return errors.New("error storing player").Wrap(err)
	}
	return nil
}

func newGame(Board int, Session int, Scores map[int]float32) (game, error) {
	ID, err := incGameCount()
	if err != nil {
		return game{}, err
	}
	Game := game{
		ID: ID,
		Board: Board,
		Session: Session,
	}
	err = Game.store()
	if err != nil {
		return Game, err
	}
	gameIDs := make([]int, 0)
	key := fmt.Sprintf("session-%v-games", Session)
	if err := app.LocalStorage.Get(key, &gameIDs); err != nil {
		return Game, errors.New("error fetching session games").Wrap(err)
	}
	gameIDs = append(gameIDs, Game.ID)
	if err := app.LocalStorage.Set(key, gameIDs); err != nil {
		return Game, errors.New("error storing session games").Wrap(err)
	}

	key = fmt.Sprintf("game-%v-scores", Game.ID)
	if err := app.LocalStorage.Set(key, Scores); err != nil {
		return Game, errors.New("error storing game scores").Wrap(err)
	}
	
	return Game, err
}

func (g game) store() error {
	if err := app.LocalStorage.Set(fmt.Sprintf("game-%v", g.ID), g); err != nil {
		return errors.New("error storing game").Wrap(err)
	}
	return nil
}

