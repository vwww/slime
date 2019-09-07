package duel

import (
	"encoding/binary"
	"time"

	"github.com/gorilla/websocket"
)

// processHello processes the first incoming message.
func processHello(c *websocket.Conn) (name []byte, col uint8) {
	mt, h, err := c.ReadMessage()

	if mt != websocket.BinaryMessage || err != nil || len(h) < 1 {
		return
	}

	name = h[1:]
	col = h[0]
	return
}

// Recv processes incoming messages after the hello message.
func Recv(c *Client, msg []byte) {
	c.g.pLock.Lock()
	defer c.g.pLock.Unlock()

	if c.cn == -1 {
		return
	}

	if len(msg) == 8 {
		// handle pongs
		t := binary.BigEndian.Uint64(msg)
		n := uint64(time.Now().UnixNano())
		if n >= t {
			newPing := uint((n - t) / 1000000)
			if c.ping != 0xFFFF {
				newPing = ((uint(c.ping) * 3) + newPing) / 4
			}
			c.ping = uint16(newPing)
		}
	} else {
		p := &c.g.players[c.cn]
		if len(msg) == 4 {
			// movement
			p.D.X = float64(binary.BigEndian.Uint16(msg)) * (MAX_W / 0xFFFF)
			p.D.Y = float64(binary.BigEndian.Uint16(msg[2:])) * (MAX_H / 0xFFFF)
		} else if len(msg) == 1 {
			// spawn
			wantSpawn := msg[0] != 0
			if wantSpawn {
				// TODO add to queue
			} else {
				// TODO remove from queue
			}
		}
	}
}

func MsgWelcome(cn int) []byte {
	return []byte{0, byte(cn)}
}

func msgEnter(code, cn int, col uint8, k, d, c, s uint, name string) []byte {
	b := make([]byte, 19+len(name))
	b[0] = byte(code)
	b[1] = byte(cn)
	binary.BigEndian.PutUint32(b[2:], uint32(k))
	binary.BigEndian.PutUint32(b[6:], uint32(d))
	binary.BigEndian.PutUint32(b[10:], uint32(c))
	binary.BigEndian.PutUint32(b[14:], uint32(s))
	b[18] = col
	copy(b[19:], name)
	return b
}

func MsgEnter(cn int, col uint8, k, d, c, s uint, name string) []byte {
	return msgEnter(1, cn, col, k, d, c, s, name)
}

func MsgEnterBot(cn int, col uint8, k, d, c, s uint, name string) []byte {
	return msgEnter(2, cn, col, k, d, c, s, name)
}

func MsgLeave(cn int) []byte {
	return []byte{3, byte(cn)}
}

func buildWorldState(g *Game) []byte {
	// Pre-allocate size
	n := uint(0)
	for i := range g.players {
		p := &g.players[i]
		if p.IsAlive {
			n++
		}
	}

	// Build message
	msg := make([]byte, n*13+1)
	msg[0] = 4

	n = 1
	for i := range g.players {
		p := &g.players[i]
		if p.IsAlive {
			b := msg[n:]
			b[0] = byte(i)
			binary.BigEndian.PutUint16(b[1:], uint16(p.O.X*(0xFFFF/MAX_W)))
			binary.BigEndian.PutUint16(b[3:], uint16(p.O.Y*(0xFFFF/MAX_H)))
			binary.BigEndian.PutUint16(b[5:], uint16(p.D.X*(0xFFFF/MAX_W)))
			binary.BigEndian.PutUint16(b[7:], uint16(p.D.Y*(0xFFFF/MAX_H)))
			binary.BigEndian.PutUint32(b[9:], uint32(p.M))
			n += 13
		}
	}

	return msg
}

func MsgDeath(killer, victim int) []byte {
	return []byte{5, byte(killer), byte(victim)}
}

func MsgPingTime(cn int, ping uint16) []byte {
	b := [3]byte{6}
	binary.BigEndian.PutUint16(b[1:], ping)
	return b[:]
}

func MsgPing() []byte {
	t := uint64(time.Now().UnixNano())
	b := [9]byte{7}
	binary.BigEndian.PutUint64(b[1:], t)
	return b[:]
}
