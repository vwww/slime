package slime

import (
	"bytes"
	"sync"
)

// MoveState is the origin (position) and velocity of dynamic entities.
type MoveState struct {
	O, V Vec2
}

// InputState represents client input (which keys on keyboard are pressed).
type InputState struct {
	L, R, U bool
}

// Player represents a connected client.
type Player struct {
	// Inputs
	Name  string
	Color int
	InputState

	// Game State
	MoveState

	Stop     chan struct{}
	Stopped  bool
	stopLock sync.Mutex

	// If implemented, we would "send" nothing to server bots
	Ping int
	PlayerSender
}

type PlayerSender interface {
	SendWelcome()
	SendState(self, other, ball MoveState, selfKeys, otherKeys InputState)
	SendEnter(name string, col int)
	SendLeave()
	SendEndRound(win bool)
	SendNextRound(isFirst bool)
	SendPing()
	SendPingTimes(lPing, rPing int)
}

// NewPlayer makes a player with the given name and color.
func NewPlayer(name []byte, col int, s PlayerSender) *Player {
	p := new(Player)
	p.Name = filterName(name)
	p.Color = filterColor(col)
	p.Ping = -1
	p.Stop = make(chan struct{})
	p.PlayerSender = s
	return p
}

// Close marks the player as "stopped" and closes the Stop chan, so
// the Stop chan will no longer block.
func (p *Player) Close() {
	if !p.Stopped {
		p.stopLock.Lock()
		defer p.stopLock.Unlock()

		if !p.Stopped {
			p.Stopped = true
			close(p.Stop)
		}
	}
}

// TODO inline this function
func (p *Player) AddPing(ping int) {
	if p.Ping != -1 {
		ping = ((p.Ping * 3) + ping) / 4
	}
	p.Ping = ping
}

// filterName sanitizes a name.
// Invalid characters are removed.
// Names are truncated if they are too long.
// If a name would be blank, a valid name will be returned instead.
func filterName(name []byte) string {
	const MAX_NAME_LEN = 16

	name = bytes.Map(func(r rune) rune {
		switch {
		case (r >= '0' && r <= '9'),
			(r >= 'A' && r <= 'Z'),
			(r >= 'a' && r <= 'z'),
			r == '-',
			r == ' ':
			return r

		default:
			return -1
		}
	}, name)

	name = bytes.TrimSpace(name)
	if len(name) > MAX_NAME_LEN {
		name = name[:MAX_NAME_LEN]
	}

	// Convert to string
	s := "unnamed"
	if len(name) != 0 {
		s = string(name)
	}
	return s
}

// filterColor sanitizes a color value.
func filterColor(c int) int {
	return c & 0xFFFFFF
}
