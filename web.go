package main

import (
	"victorz.ca/slimeserv/slime"

	"fmt"
	"net/http"
	"os"
)

func init() {
	http.HandleFunc("/slime/", slime.HandleIndex)
	http.HandleFunc("/slime/c", slime.HandleConnect)
	http.HandleFunc("/slime/d", slime.HandleData)
	http.HandleFunc("/slime/x", slime.HandleDisc)
	http.Handle("/crossdomain.xml", http.FileServer(http.Dir(".")))
}

func main() {
	slime_done := slime.LaunchCron()
	defer close(slime_done)

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
