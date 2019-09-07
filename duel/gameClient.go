package duel

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WSWriter interface {
	Write(c *websocket.Conn) error
}

// PrepareMessage tries to create a prepared message, but fallbacks
// to a byte slice if it fails.
func PrepareMessage(msg []byte) WSWriter {
	const mt = websocket.BinaryMessage
	pm, err := websocket.NewPreparedMessage(mt, msg)
	if err == nil {
		return WSPreparedWriter{pm}
	}
	return WSByteWriter(msg)
}

type WSByteWriter []byte

func (msg WSByteWriter) Write(c *websocket.Conn) error {
	const mt = websocket.BinaryMessage
	return c.WriteMessage(mt, msg)
}

type WSPreparedWriter struct{ *websocket.PreparedMessage }

func (msg WSPreparedWriter) Write(c *websocket.Conn) error {
	return c.WritePreparedMessage(msg.PreparedMessage)
}

type Client struct {
	g       *Game
	cn      int
	SendBuf <-chan WSWriter
	sendBuf chan<- WSWriter
	lock    sync.Mutex
	ping    uint16
}

// newClient makes a new Client for a specific game and client number.
func newClient(g *Game, cn int) *Client {
	sendBuf := make(chan WSWriter, 300) // enough for at least 2 seconds
	return &Client{
		g,
		cn,
		sendBuf,
		sendBuf,
		sync.Mutex{},
		0xFFFF,
	}
}

// Send enqueues an outgoing message, or
// on failure, closes the Player.
func (c *Client) Send(msg WSWriter) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.cn == -1 {
		return
	}

	select {
	case c.sendBuf <- msg:
	default:
		// send queue overflow
		c.Close()
	}
}

// SendB calls Send for a byte slice.
func (c *Client) SendB(msg []byte) {
	c.Send(WSByteWriter(msg))
}

// Close prevents future received messages from being forwarded to the Game.
// It is safe to call Close multiple times.
func (c *Client) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.cn != -1 {
		close(c.sendBuf)
		c.g.DelPlayer(c.cn)
		c.cn = -1
	}
}
