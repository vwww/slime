// Package duel implements the logic for the Duel server.
package duel

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// HandleNum responds to the HTTP request by writing the current number of players.
func (g *Game) HandleNum(w http.ResponseWriter, r *http.Request) {
	n := g.pCount
	fmt.Fprintf(w, "%v", n)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandlePlayer serves a game client.
func (g *Game) HandlePlayer(w http.ResponseWriter, r *http.Request) {
	log.Printf(" [%v] connected\n", r.RemoteAddr)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("*[%v] upgrade failed: %v\n", r.RemoteAddr, err)
		return
	}
	defer log.Printf(" [%v] disconnected\n", r.RemoteAddr)
	defer c.Close()

	cl := g.AddPlayer(processHello(c))
	if cl == nil {
		return
	}

	name := g.players[cl.cn].Name

	n := g.pCount
	log.Printf("+[%v] %v (%v now)\n", r.RemoteAddr, name, n)

	go reader(cl, c)
	writer(cl, c)

	n = g.pCount
	log.Printf("-[%v] %v (%v total)\n", r.RemoteAddr, name, n)
}

func reader(c *Client, conn *websocket.Conn) {
	defer c.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		Recv(c, msg)
	}
}

func writer(c *Client, conn *websocket.Conn) {
	defer c.Close()

	for msg := range c.SendBuf {
		err := msg.Write(conn)
		if err != nil {
			break
		}
	}
}
