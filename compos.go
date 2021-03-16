package main

import (
	"strings"
	"time"
	"strconv"
	"encoding/json"
	"fmt"
//	"net/url"
	
	"github.com/maxence-charriere/go-app/v7/pkg/app"
	"github.com/maxence-charriere/go-app/v7/pkg/errors"
)

type section int

const (
	SMenu section = iota
	SSession
	SNewGame
	SGame
	SSessions
	SPlayers
	SGames
	SDownload
)

type fullpage struct {
	app.Compo

	Section section
	
	// for downpages
	Session int
	Game int
	Previous section
}

func (f *fullpage) Render() app.UI {
	if f.Section == SDownload {
		return app.Div().Body(
			&downloadpage { Full: f },
		)
	}
	return app.Div().Body(
		app.H1().Text("Personal Boardgame Logbook"),
		app.If(f.Section == SMenu, &mainmenu{ Full: f },).
			ElseIf(f.Section == SSession, &sessionpage { Full: f, SessionID: f.Session },).
			ElseIf(f.Section == SNewGame, &newgamepage { Full: f, SessionID: f.Session },).
			ElseIf(f.Section == SSessions, &sessionspage { Full: f },).
			ElseIf(f.Section == SPlayers, &playerspage { Full: f },).
			ElseIf(f.Section == SGames, &boardspage { Full: f },).
			ElseIf(f.Section == SGame, &gamepage { Full: f, SessionID: f.Session, GameID: f.Game },),

	)
}

func (f *fullpage) newSession() error {
	Session, err := newSession()
	if err != nil {
		return errors.New("error creating new session").Wrap(err)
	}
	f.Section = SSession
	f.Session = Session.ID
	f.Previous = SMenu
	f.Update()
	
	return nil
}

func (f *fullpage) download() {
	f.Section = SDownload
	f.Update()
}

type mainmenu struct {
	app.Compo

	Full *fullpage
}

func (m *mainmenu) Render() app.UI {
	return app.Div().Body(app.Stack().Center().
		Content(
		app.Button().Text("New Session").OnClick(m.onNewSession),
		app.Button().Text("Sessions").OnClick(m.onSessions),
		app.Button().Text("Players").OnClick(m.onPlayers),
		app.Button().Text("Games").OnClick(m.onGames),
		app.Button().Text("Download").OnClick(m.onDownload),
	))
}

func (m *mainmenu) onNewSession(ctx app.Context, e app.Event) {
	if err := m.Full.newSession(); err != nil {	
		app.Log("%s", errors.New("creating new session failed").Wrap(err))
	}
}
func (m *mainmenu) onSessions(ctx app.Context, e app.Event) {
	m.Full.Section = SSessions
	m.Full.Update()
}

func (m *mainmenu) onPlayers(ctx app.Context, e app.Event) {
	m.Full.Section = SPlayers
	m.Full.Update()
}

func (m *mainmenu) onGames(ctx app.Context, e app.Event) {
	m.Full.Section = SGames
	m.Full.Update()
}

func (m *mainmenu) onDownload(ctx app.Context, e app.Event) {
	m.Full.download()
}

type sessionpage struct {
	app.Compo

	Full *fullpage
	SessionID int
	Session session
	Games []game
	Boards map[int]board
}

func (s *sessionpage) OnMount(ctx app.Context) {
	var err error
	if s.Session, err = retrieveSession(s.SessionID); err != nil {
		app.Log("%s", errors.New("error fetching session").Wrap(err))
		return
	}
	if s.Games, err = retrieveGamesInSession(s.SessionID); err != nil {
		app.Log("%s", errors.New("error fetching games for session").Wrap(err))
		return
	}
	s.Boards = map[int]board{}
	for _, game := range s.Games {
		if s.Boards[game.Board], err = retrieveBoard(game.Board); err != nil {
			app.Log("%s", errors.Newf("error fetching board game %v for session %v", game.Board, s.SessionID).Wrap(err))
			return
		}
	}
	s.Update()
}

func (s *sessionpage) Render() app.UI {
	theTime := time.Unix(s.Session.Date, 0)
	return app.Div().Body(
		app.H2().Text("Session for "  +  theTime.Format("2006-01-02")),
		app.Button().Text("New Game").OnClick(s.onNewGame),
		app.Button().Text("Close Session").OnClick(s.onCloseSession),
		app.Ol().Body(
			app.Range(s.Games).Slice(func(i int) app.UI {
				return app.Li().Body(
					app.Button().
						Text(s.Boards[s.Games[i].Board].Text).
						DataSet("game", i).
						OnClick(s.onGame),
				)},
			),
		),
	)
}

func (s *sessionpage) onCloseSession(ctx app.Context, e app.Event) {
	s.Full.Section = s.Full.Previous
	if s.Full.Previous == SSession {
		s.Full.Section = SMenu
	}
	s.Full.Update()
}

func (s *sessionpage) onNewGame(ctx app.Context, e app.Event) {
	s.Full.Section = SNewGame
	s.Full.Update()
}

func (s *sessionpage) onGame(ctx app.Context, e app.Event) {
	i, err := strconv.Atoi(ctx.JSSrc.Get("dataset").Get("game").String())
	if err != nil {
		app.Log("%s", "Unknown game for onGame")
	}
	s.Full.Game = i
	s.Full.Section = SGame
	s.Full.Previous = SSession
	s.Full.Update()
}

type newgamepage struct {
	app.Compo

	Full *fullpage
	SessionID int
	AllBoards []board
	AllPlayers []player
	HasBoard bool
	Board int
	BoardInput string
	PlayerInput string
	Players []int
	Scores map[int]float32
}

func (n *newgamepage) OnMount(ctx app.Context) {
	var err error
	n.HasBoard = false
	if n.AllBoards, err = retrieveAllBoards(); err != nil {
		app.Log("%s", errors.New("error fetching all boards").Wrap(err))
		return
	}
	if n.AllPlayers, err = retrieveAllPlayers(); err != nil {
		app.Log("%s", errors.New("error fetching all players").Wrap(err))
		return
	}
	n.Scores = make(map[int]float32)
	n.Update()
}

func  (n *newgamepage) Render() app.UI {
	gameOf := ""
	if n.HasBoard {
		gameOf = n.AllBoards[n.Board].Text
	}
	return app.Div().Body(
		app.If(n.HasBoard,
			app.H2().Text("New Game of " + gameOf),
		).Else(
			app.H2().Text("New Game"),
			app.H3().Text("Choose a boardgame (or type a name to create it)"),
			app.Div().Body(
				app.Input().
					OnInput(n.onBoardChange),//.OnKeyPress(n.onBoardChangeKey),
				app.Button().Text("New").OnClick(n.onNewBoard).Disabled(len(n.BoardInput) == 0),
				app.Ul().Body(
					app.Range(n.AllBoards).Slice(func(i int) app.UI {
						Board := n.AllBoards[i]
						text := n.AllBoards[i].Text
						if !Board.Hidden && (len(n.BoardInput) == 0 || strings.Index(text, n.BoardInput) >= 0) {
							return app.Li().Body(
								app.Button().Text(text).
									DataSet("board", i).OnClick(n.onSetBoard))
						}
						return app.Text("")
					})),
			),
		),
		app.H3().Text("Players:"),
		app.Div().Body(
			app.Range(n.Players).Slice(func(i int) app.UI {
				return app.Stack().Content(
					app.Button().Text("-").DataSet("player", i).OnClick(n.onDelPlayer),
					app.Text(n.AllPlayers[n.Players[i]].Text),
					app.Text(". Score:"),
					app.Input().DataSet("player", i).OnInput(n.onSetScore))
			}),
			app.Button().Text("RECORD GAME").OnClick(n.onSave),
			app.Button().Text("Cancel").OnClick(n.onCancel),
		),
		app.H3().Text("Add players:"),
		app.Div().Body(
			app.Input().OnInput(n.onPlayerChange),
			app.Button().Text("New").OnClick(n.onNewPlayer).Disabled(len(n.PlayerInput) == 0),
			app.Ul().Body(
				app.Range(n.AllPlayers).Slice(func(i int) app.UI {
					Player := n.AllPlayers[i]
					text := n.AllPlayers[i].Text
					if !Player.Hidden && (len(n.PlayerInput) == 0 || strings.Index(text, n.PlayerInput) >= 0) {
						return app.Li().Body(
							app.Button().Text("+ " + text).
								DataSet("player", i).OnClick(n.onAddPlayer))
					}
					return app.Text("")
				})),
		),
	)
}

func (n *newgamepage) onBoardChange(ctx app.Context, e app.Event) {
	n.BoardInput = ctx.JSSrc.Get("value").String()
	n.Update()
}

func (n *newgamepage) onPlayerChange(ctx app.Context, e app.Event) {
	n.PlayerInput = ctx.JSSrc.Get("value").String()
	n.Update()
}

func (n *newgamepage) onNewBoard(ctx app.Context, e app.Event) {
	Board, err := newBoard(n.BoardInput)
	if err != nil {
		app.Log("%s", errors.New("error creating new board").Wrap(err))
		return
	}
	n.Board = Board.ID
	n.AllBoards = append(n.AllBoards, Board)
	n.HasBoard = true
	n.BoardInput = ""
	n.Update()
}

func (n *newgamepage) onSetBoard(ctx app.Context, e app.Event) {
	i, err := strconv.Atoi(ctx.JSSrc.Get("dataset").Get("board").String())
	if err != nil {
		app.Log("%s", "Unknown board for onSetBoard")
	}
	n.Board = i
	n.HasBoard = true
	n.Update()
}

func (n *newgamepage) onSetScore(ctx app.Context, e app.Event) {
	i, err := strconv.Atoi(ctx.JSSrc.Get("dataset").Get("player").String())
	if err != nil {
		app.Log("%s", "Unknown player for onSetScore")
	}
	score, err := strconv.ParseFloat(ctx.JSSrc.Get("value").String(), 32)
	if err != nil {
		score = 0
	}
	n.Scores[i] = float32(score)
	n.Update()
}

func (n *newgamepage) onCancel(ctx app.Context, e app.Event) {
	n.Full.Section = SSession
	n.Full.Update()	
}

func (n *newgamepage) onSave(ctx app.Context, e app.Event) {
	_, err := newGame(n.Board, n.SessionID, n.Scores)
	if err != nil {
		app.Log("%s", errors.New("error creating new game").Wrap(err))
		return
	}
	n.Full.Section = SSession
	n.Full.Update()
}

func (n *newgamepage) onNewPlayer(ctx app.Context, e app.Event) {
	Player, err := newPlayer(n.PlayerInput)
	if err != nil {
		app.Log("%s", errors.New("error creating new player").Wrap(err))
		return
	}
	n.Players = append(n.Players, Player.ID)
	n.AllPlayers = append(n.AllPlayers, Player)
	n.PlayerInput = ""
	n.Update()
}

func (n *newgamepage) onAddPlayer(ctx app.Context, e app.Event) {
	id, err := strconv.Atoi(ctx.JSSrc.Get("dataset").Get("player").String())
	if err != nil {
		app.Log("%s", "Unknown player for onAddPlayer")
	}
	
	n.Players = append(n.Players, id)
	n.Update()
}

func (n *newgamepage) onDelPlayer(ctx app.Context, e app.Event) {
	id, err := strconv.Atoi(ctx.JSSrc.Get("dataset").Get("player").String())
	if err != nil {
		app.Log("%s", "Unknown player for onDelPlayer")
	}
	found := -1
	for pos, other := range n.Players {
		if other == id {
			found = pos
		}
	}
	if found >= 0 {
		n.Players = append(n.Players[:found], n.Players[found+1:]...)
		n.Update()
	}
}


type downloadpage struct {
	app.Compo

	Full *fullpage
	Ready bool
	Data string
}

func (d *downloadpage) OnMount(ctx app.Context) {
	go d.prepareData()
}

func (d *downloadpage) Render() app.UI {
	if !d.Ready {
		return app.Text("Preparing your download...")
	}
	return app.Pre().Text(d.Data)
	// problem with the router and pushstate
	//return app.A().Text("download").Download(true).Href(d.Data)
}

func (d *downloadpage) prepareData() {
	data := make(map[string] interface{})
	var err error
	data["sessions"], err = retrieveAllSessions()
	if err != nil {
		app.Log("%s", errors.New("error preparing data").Wrap(err))
	}
	data["boards"], err = retrieveAllBoards()
	if err != nil {
		app.Log("%s", errors.New("error preparing data").Wrap(err))
	}
	data["players"], err = retrieveAllPlayers()
	if err != nil {
		app.Log("%s", errors.New("error preparing data").Wrap(err))
	}
	//DataBytes, err := json.Marshal(data)
	DataBytes, err := json.MarshalIndent(data, "", "  ")	
	if err != nil {
		app.Log("%s", errors.New("error preparing data").Wrap(err))
	}
	// not working yet
	//Data := "data:text/plain;charset=utf-8," + url.QueryEscape(string(DataBytes))
	Data := string(DataBytes)
	app.Dispatch(func() { // Ensures update is on UI goroutine.
		d.Data = Data
		d.Ready = true
		d.Update()
	})
}


type gamepage struct {
	app.Compo

	Full *fullpage
	SessionID int
	Session session
	GameID int
	Game game
	Scores []score
	Board board
	Players map[int]player
}

func (g *gamepage) OnMount(ctx app.Context) {
	var err error
	g.Session, err = retrieveSession(g.SessionID)
	if err != nil {
		app.Log("%s", errors.New("error retrieving session").Wrap(err))
		return
	}
	g.Game, err = retrieveGame(g.GameID)
	if err != nil {
		app.Log("%s", errors.New("error retrieving game").Wrap(err))
		return
	}
	g.Scores, err = retrieveScoresInGame(g.Game.ID)
	if err != nil {
		app.Log("%s", errors.New("error retrieving scores").Wrap(err))
		return
	}
	g.Board, err = retrieveBoard(g.Game.Board)
	if err != nil {
		app.Log("%s", errors.New("error retrieving board").Wrap(err))
		return
	}
	g.Players = make(map[int]player, len(g.Scores))
	for _, Score := range g.Scores {
		g.Players[Score.Player], err = retrievePlayer(Score.Player)
		if err != nil {
			app.Log("%s", errors.New("error retrieving player").Wrap(err))
			return
		}
	}
	g.Update()
}

func  (g *gamepage) Render() app.UI {
	theTime := time.Unix(g.Session.Date, 0)

	return app.Div().Body(
		app.H2().Text("Session for "  +  theTime.Format("2006-01-02")),
		app.H3().Text("Game of " + g.Board.Text),
		app.Text("Players:"),
		app.Ul().Body(
			app.Range(g.Scores).Slice(func(i int) app.UI {
				Player := g.Players[g.Scores[i].Player]
				Score := g.Scores[i].Score
				return app.Li().Body(
					app.Text(fmt.Sprintf("%v: %v", Player.Text, Score)),
				)
			})),
		app.Button().Text("close").OnClick(g.onClose),
	)
}

func (g *gamepage) onClose(ctx app.Context, e app.Event) {
	g.Full.Section = g.Full.Previous
	g.Full.Update()	
}


type sessionspage struct {
	app.Compo

	Full *fullpage
	Sessions []session
}

func (s *sessionspage) OnMount(ctx app.Context) {
	var err error
	s.Sessions, err = retrieveAllSessions()
	if err != nil {
		app.Log("%s", errors.New("error retrieving sessions").Wrap(err))
		return
	}
	s.Update()
}

func  (s *sessionspage) Render() app.UI {
	totalLen := len(s.Sessions)
	return app.Div().Body(
		app.H2().Text("Sessions"),
		app.Ul().Body(
			app.Range(s.Sessions).Slice(func(i int) app.UI {
				theTime := time.Unix(s.Sessions[totalLen - i - 1].Date, 0)
				return app.Li().Body(
					app.Button().Text("Session for "  +  theTime.Format("2006-01-02")).
						DataSet("session", totalLen - i - 1).
						OnClick(s.onSession))
			})),
		app.Button().Text("close").OnClick(s.onClose),
	)
}


func (s *sessionspage) onSession(ctx app.Context, e app.Event) {
	i, err := strconv.Atoi(ctx.JSSrc.Get("dataset").Get("session").String())
	if err != nil {
		app.Log("%s", "Unknown session for onSession")
	}
	s.Full.Section = SSession
	s.Full.Session = i
	s.Full.Previous = SSessions
	s.Full.Update()
}

func (s *sessionspage) onClose(ctx app.Context, e app.Event) {
	s.Full.Section = SMenu
	s.Full.Update()	
}

type playerspage struct {
	app.Compo

	Full *fullpage
	Players []player
}

func (p *playerspage) OnMount(ctx app.Context) {
	var err error
	p.Players, err = retrieveAllPlayers()
	if err != nil {
		app.Log("%s", errors.New("error retrieving players").Wrap(err))
		return
	}
	p.Update()
}

func  (p *playerspage) Render() app.UI {
	return app.Div().Body(
		app.H2().Text("Players"),
		app.Ul().Body(
			app.Range(p.Players).Slice(func(i int) app.UI {
				Player := p.Players[i]
				show := "hide"
				if Player.Hidden {
					show = "show"
				}
				return app.Li().Body(
					app.Text(Player.Text),
					app.Button().Text(show).
						DataSet("player", i).
						OnClick(p.onToggle),
				)
			})),
		app.Button().Text("close").OnClick(p.onClose),
	)
}


func (p *playerspage) onToggle(ctx app.Context, e app.Event) {
	i, err := strconv.Atoi(ctx.JSSrc.Get("dataset").Get("player").String())
	if err != nil {
		app.Log("%s", "Unknown player for onToggle")
	}
	p.Players[i].Hidden = !p.Players[i].Hidden
	p.Players[i].store()
	p.Update()
}

func (p *playerspage) onClose(ctx app.Context, e app.Event) {
	p.Full.Section = SMenu
	p.Full.Update()	
}

type boardspage struct {
	app.Compo

	Full *fullpage
	Boards []board
}

func (b *boardspage) OnMount(ctx app.Context) {
	var err error
	b.Boards, err = retrieveAllBoards()
	if err != nil {
		app.Log("%s", errors.New("error retrieving boards").Wrap(err))
		return
	}
	b.Update()
}

func  (b *boardspage) Render() app.UI {
	return app.Div().Body(
		app.H2().Text("Games"),
		app.Ul().Body(
			app.Range(b.Boards).Slice(func(i int) app.UI {
				Board := b.Boards[i]
				show := "hide"
				if Board.Hidden {
					show = "show"
				}
				return app.Li().Body(
					app.Text(Board.Text),
					app.Button().Text(show).
						DataSet("board", i).
						OnClick(b.onToggle),
				)
			})),
		app.Button().Text("close").OnClick(b.onClose),
	)
}


func (b *boardspage) onToggle(ctx app.Context, e app.Event) {
	i, err := strconv.Atoi(ctx.JSSrc.Get("dataset").Get("board").String())
	if err != nil {
		app.Log("%s", "Unknown board for onToggle")
	}
	b.Boards[i].Hidden = !b.Boards[i].Hidden
	b.Boards[i].store()
	b.Update()
}

func (b *boardspage) onClose(ctx app.Context, e app.Event) {
	b.Full.Section = SMenu
	b.Full.Update()	
}

