// Package slime implements the logic for the Slime Volleyball Multiplayer server.
package slime

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var pCount = 0 // current number of players
var pCountLock sync.Mutex

// HandleNum writes the current number of players to the HTTP request.
func HandleNum(w http.ResponseWriter, r *http.Request) {
	n := pCount
	fmt.Fprintf(w, "%v", n)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandlePlayer serves a game client.
func HandlePlayer(w http.ResponseWriter, r *http.Request) {
	log.Printf(" [%v] connected\n", r.RemoteAddr)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("*[%v] upgrade failed: %v\n", r.RemoteAddr, err)
		return
	}
	defer log.Printf(" [%v] disconnected\n", r.RemoteAddr)
	defer c.Close()

	p := processHello(c)
	if p == nil {
		return
	}

	pCountLock.Lock()
	pCount++
	pCountLock.Unlock()
	log.Printf("+[%v] %v #%06x (%v now)\n", r.RemoteAddr, p.Name, p.Color, pCount)

	go reader(p, c)
	go writer(p, c)
	playMatches(p.Player)

	pCountLock.Lock()
	pCount--
	pCountLock.Unlock()
	log.Printf("-[%v] %v (%v total)\n", r.RemoteAddr, p.Name, pCount)
}

func processHello(c *websocket.Conn) *Player {
	mt, h, err := c.ReadMessage()

	if mt != websocket.BinaryMessage || err != nil || len(h) < 3 {
		return nil
	}

	name := h[3:]
	col := int(h[0])<<16 | int(h[1])<<8 | int(h[2])

	return NewPlayer(name, col)
}

func reader(p *Player, c *websocket.Conn) {
	defer p.Close()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}
		p.Recv(msg)
	}
}

func writer(p *Player, c *websocket.Conn) {
	defer p.Close()

	for msg := range p.SendBuf {
		mt := websocket.BinaryMessage
		if err := c.WriteMessage(mt, msg); err != nil {
			break
		}
	}
}

type matchReq struct {
	p      *Player
	result chan struct{}
}

var matcher = make(chan matchReq)

func playMatches(p *Player) {
	p.SendWelcome()
	for {
		m := matchReq{p, make(chan struct{})}
		select {
		case <-p.Stop:
			return
		case matcher <- m:
			// wait for game to end
			<-m.result
		case other := <-matcher:
			m.result = nil // free unused chan
			g := NewGame(p, other.p)
			g.Run()
			other.result <- struct{}{}
		}
	}
}
