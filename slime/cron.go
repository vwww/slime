package slime

import (
	"time"
)

func pruneOldPlayers() {
	const FOUR_SECONDS = int64(4 * 1000 * 1000 * 1000)
	pruneTime := time.Now().UnixNano() - FOUR_SECONDS

	removed := make(map[*player]struct{})

	pLock.Lock()
	defer pLock.Unlock()
	for sid, p := range players {
		if p.lastReceive < pruneTime {
			removed[p] = struct{}{}
			delete(players, sid)
			pNum--
			logline("!%v timeout (%v total)", sid, pNum)

			if p == pWait {
				pWait = nil
			}
		}
	}

	for p := range removed {
		op := p.opponent
		if op == nil {
			continue
		}
		if _, bad := removed[op]; !bad {
			op.oppDisc = true
			op.tryMatch()
		}
	}
}

func cron(done <-chan struct{}) {
	checkPlayers := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-checkPlayers.C:
			pruneOldPlayers()
		case <-done:
			return
		}
	}
}

func LaunchCron() chan<- struct{} {
	done := make(chan struct{})
	go cron(done)
	return done
}
