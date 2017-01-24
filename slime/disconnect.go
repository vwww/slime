package slime

import (
	"net/http"
)

func HandleDisc(res http.ResponseWriter, req *http.Request) {
	sid := req.FormValue("d")
	if ok, pCount := tryDisconnect(sid); ok {
		res.Write([]byte{'.'})
		logline("-%v %v (%v total)", sid, req.RemoteAddr, pCount)
	} else {
		res.Write([]byte{'!'})
	}
}

func tryDisconnect(sid string) (ok bool, pCount int) {
	pLock.Lock()
	defer pLock.Unlock()

	p, ok := players[sid]
	if !ok {
		// Player not found
		return
	}

	delete(players, sid)
	pNum--

	if pWait == p {
		pWait = nil
	} else {
		p.opponent.oppDisc = true
		p.opponent.tryMatch()
	}

	return true, pNum
}
