package slime

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var pCount = 0
var pCountLock sync.Mutex

func HandleNum(w http.ResponseWriter, r *http.Request) {
	n := pCount // no lock for reading int
	fmt.Fprintf(w, "%v", n)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandlePlayer(w http.ResponseWriter, r *http.Request) {
	logline(" [%v] connected", r.RemoteAddr)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logline("*[%v] upgrade failed: %v", r.RemoteAddr, err)
		return
	}
	defer logline(" [%v] disconnected", r.RemoteAddr)
	defer c.Close()

	p := processHello(c)
	if p == nil {
		return
	}

	pCountLock.Lock()
	pCount++
	pCountLock.Unlock()
	logline("+[%v] %v #%06x (%v now)", r.RemoteAddr, p.Name, p.Color, pCount)

	go reader(p, c)
	go writer(p, c)
	playMatches(p.Player)

	pCountLock.Lock()
	pCount--
	pCountLock.Unlock()
	logline("-[%v] %v (%v total)", r.RemoteAddr, p.Name, pCount)
}

func processHello(c *websocket.Conn) *RemotePlayer {
	mt, h, err := c.ReadMessage()

	if mt != websocket.BinaryMessage || err != nil || len(h) < 3 {
		return nil
	}

	name := h[3:]
	col := int(h[0])<<16 | int(h[1])<<8 | int(h[2])

	return NewRemotePlayer(name, col)
}

func reader(r *RemotePlayer, c *websocket.Conn) {
	defer r.Close()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}
		if len(msg) == 8 {
			// intercept pongs
			t := binary.BigEndian.Uint64(msg)
			n := uint64(time.Now().UnixNano())
			if n >= t {
				newPing := int((n - t) / 1000000)
				r.AddPing(newPing)
			}
			continue
		}
		r.Recv(msg)
	}
}

func writer(r *RemotePlayer, c *websocket.Conn) {
	defer r.Close()

	for msg := range r.SendBuf {
		mt := websocket.BinaryMessage
		if msg == nil {
			// mt = websocket.PingMessage
			msg = make([]byte, 9)
			msg[0] = 9
			t := uint64(time.Now().UnixNano())
			binary.BigEndian.PutUint64(msg[1:], t)
		}
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
			<-m.result
		case other := <-matcher:
			m.result = nil // unused chan
			g := NewGame(p, other.p)
			g.Run()
			other.result <- struct{}{}
		}
	}
}
