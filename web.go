package main

import (
	"victorz.ca/slimeserv/slime"

	"fmt"
	"net/http"
	"os"
)

func init() {
	http.HandleFunc("/s/n", slime.HandleNum)
	http.HandleFunc("/s", slime.HandlePlayer)
	http.HandleFunc("/", hello)
}

func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "hello")
}

func main() {
	// OLD background tasks
	// slime_done := slime.LaunchCron()
	// defer close(slime_done)

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
