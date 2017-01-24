package slime

import (
	"net/http"
	"strconv"
)

// index.php
func HandleIndex(res http.ResponseWriter, req *http.Request) {
	// Read ping time
	ping, _ := strconv.Atoi(req.FormValue("t"))

	// don't get lock
	// (but use local to ensure they are the same in the response)
	pCount := pNum

	writeJSON(res, map[string]interface{}{
		"slime-server":   "0.0 beta",
		"desc":           "Victor's Golang Slime-Serv",
		"players_active": pCount,
		"players_in":     pCount,
		"players_max":    1337420,
		"ping":           ping,
		"connect":        "c",
	}, []byte{}, []byte("!Error: encoding JSON"))
}
