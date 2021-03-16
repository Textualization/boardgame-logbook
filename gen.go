// +build !wasm

package main

import (
	"log"

        "github.com/maxence-charriere/go-app/v7/pkg/app"
)

func main() {
	err := app.GenerateStaticWebsite("docs/", &app.Handler{
		Title:        "Personal Boardgame Logbook",
		Name:        "Boardgame Logbook",
		Description: "Your personal logbook for board game sessions.",
		Icon: app.Icon{
			Default: "/web/logo192.png",
		},
		Keywords: []string{
			"board games",
			"logbook",
			"private",
			"app",
			"pwa",
		},
		Resources:   app.GitHubPages("boardgame-logbook"),
    })

    if err != nil {
        log.Fatal(err)
    }
}
