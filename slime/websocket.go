package slime

import (
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

var pCount = 0
var pCountLock sync.Mutex

func HandleNum(w http.ResponseWriter, r *http.Request) {
	n := pCount // no lock for reading int
	fmt.Fprintf(w, "%v", n)
}

var Handler = websocket.Handler(HandlePlayer)

func HandlePlayer(ws *websocket.Conn) {
	defer ws.Close()

	r := ws.Request()
	if r == nil {
		return
	}

	logline(" [%v] connected", r.RemoteAddr)
	defer logline(" [%v] disconnected", r.RemoteAddr)

	p := processHello(ws)
	if p == nil {
		return
	}

	pCountLock.Lock()
	pCount++
	pCountLock.Unlock()
	logline("+[%v] %v #%06x (%v now)", r.RemoteAddr, p.Name, p.Color, pCount)

	go reader(p, ws)
	go writer(p, ws)
	playMatches(p.Player)

	pCountLock.Lock()
	pCount--
	pCountLock.Unlock()
	logline("-[%v] %v (%v total)", r.RemoteAddr, p.Name, pCount)
}

func processHello(ws *websocket.Conn) *RemotePlayer {
	var h []byte
	err := websocket.Message.Receive(ws, &h)

	if err != nil || len(h) < 3 {
		return nil
	}

	n := h[3:]
	c := int(h[0])<<16 | int(h[1])<<8 | int(h[2])

	return NewRemotePlayer(n, c)
}

func reader(r *RemotePlayer, ws *websocket.Conn) {
	defer r.Close()

	for {
		var msg []byte
		if err := websocket.Message.Receive(ws, &msg); err != nil {
			break
		}
		r.Recv(msg)
	}
}

func writer(r *RemotePlayer, ws *websocket.Conn) {
	defer r.Close()

	for msg := range r.SendBuf {
		if err := websocket.Message.Send(ws, msg); err != nil {
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
			<-m.result
		case other := <-matcher:
			m.result = nil // unused chan
			g := NewGame(p, other.p)
			g.Run()
			other.result <- struct{}{}
		}
	}
}
