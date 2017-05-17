package slime

import (
	"bytes"
	"sync"
)

type MoveState struct {
	O, V Vec2
}

type InputState struct {
	L, R, U bool
}

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

func NewPlayer(name []byte, col int, s PlayerSender) *Player {
	p := new(Player)
	p.Name = filterName(name)
	p.Color = filterColor(col)
	p.Ping = -1
	p.Stop = make(chan struct{})
	p.PlayerSender = s
	return p
}

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

func (p *Player) AddPing(ping int) {
	if p.Ping != -1 {
		ping = ((p.Ping * 3) + ping) / 4
	}
	p.Ping = ping
}

func filterName(name []byte) string {
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
	if len(name) > 16 {
		name = name[:16]
	}

	// Convert to string
	s := "unnamed"
	if len(name) != 0 {
		s = string(name)
	}
	return s
}

func filterColor(c int) int {
	return c & 0xFFFFFF
}
