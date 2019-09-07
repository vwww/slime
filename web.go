package main

import (
	"victorz.ca/gameserv/duel"
	"victorz.ca/gameserv/slime"

	"fmt"
	"net/http"
	"os"
)

var duelGame = duel.NewGame()

func init() {
	http.HandleFunc("/s/n", slime.HandleNum)
	http.HandleFunc("/s", slime.HandlePlayer)
	http.HandleFunc("/d/n", duelGame.HandleNum)
	http.HandleFunc("/d", duelGame.HandlePlayer)
	http.HandleFunc("/", hello)
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "hello")
}

// Entry point of server program
func main() {
	// OLD background tasks
	// slime_done := slime.LaunchCron()
	// defer close(slime_done)

	go duelGame.Run()

	bind := ":8080"
	if env := os.Getenv("OPENSHIFT_GO_PORT"); env != "" {
		bind = os.Getenv("OPENSHIFT_GO_IP") + ":" + env
	} else if env := os.Getenv("PORT"); env != "" {
		bind = ":" + env
	}

	fmt.Printf("Listening on %s\n", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		panic(err)
	}
}
