package slime

import (
	"sync"
)

type player struct {
	name     string
	color    int
	opponent *player

	lastReceive int64

	// messages
	oppDisc    bool
	oppJoin    bool
	oppMovePos map[string]interface{}
	oppBallPos map[string]interface{} // ball or transfer
	oppPing    int
	msgStart   int
}

var players = make(map[string]*player)
var pNum int
var pWait *player
var pLock sync.RWMutex

func (p *player) tryMatch() (ok bool) {
	if pWait == nil {
		p.opponent = nil
		pWait = p
		return false
	} else {
		p.opponent = pWait
		p.oppJoin = true
		pWait.opponent = p
		pWait.oppJoin = true
		pWait = nil
		return true
	}
}
