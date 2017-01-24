package slime

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// connect.php
func HandleConnect(res http.ResponseWriter, req *http.Request) {
	name := filterName(req.FormValue("n"))
	color := filterColor(req.FormValue("c"))

	var secretBytes [16]byte
	n, _ := rand.Read(secretBytes[:])
	secret := hex.EncodeToString(secretBytes[:n])

	if ok, oppMsg, pCount := tryConnect(secret, name, color); ok {
		logMsgMatched := ""
		if oppMsg != nil {
			logMsgMatched = "matched, "
		}
		logline("+%v %v (%v%v now)", secret, req.RemoteAddr, logMsgMatched, pCount)

		writeJSON(res, map[string]interface{}{
			"secret":   secret,
			"n":        name,
			"c":        color,
			"opponent": oppMsg,
			"resources": map[string]string{
				"d":    "d",
				"disc": "x",
			},
		}, []byte{'*'}, []byte("Error: encoding JSON"))
	} else {
		res.Write([]byte("Error: ID collision"))
	}
}

func tryConnect(secret, name string, color int) (ok bool, oppMsg interface{}, pCount int) {
	pLock.Lock()
	defer pLock.Unlock()

	_, bad := players[secret]
	if bad {
		// ID collision
		return
	}

	// Add player
	p := &player{
		name:        name,
		color:       color,
		lastReceive: time.Now().UnixNano(),
	}
	players[secret] = p
	pNum++

	if p.tryMatch() {
		// report join immediately
		p.oppJoin = false
		oppMsg = map[string]interface{}{
			"n": p.opponent.name,
			"c": p.opponent.color,
		}
	}
	return true, oppMsg, pNum
}

func filterName(name string) string {
	name = strings.Map(func(r rune) rune {
		switch {
		case (r >= '0' && r <= '9'),
			(r >= 'A' && r <= 'Z'),
			(r >= 'a' && r <= 'z'),
			r == '-', r == ' ':
			return r

		default:
			return -1
		}
	}, name)
	name = strings.TrimSpace(name)
	if len(name) < 1 {
		name = "unnamed"
	} else if len(name) > 16 {
		name = name[:16]
	}
	return name
}

func filterColor(color string) (c int) {
	c, _ = strconv.Atoi(color)
	if c < 0 {
		c = 0
	} else if c > 0xFFFFFF {
		c = 0xFFFFFF
	}
	return
}
