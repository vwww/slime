package slime

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type clientJSON struct {
	Sid   string                   `json:"s"`
	Ping  int                      `json:"p"`
	Input []map[string]interface{} `json:"m"`
}

func HandleData(res http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.Write([]byte("!Error reading"))
		return
	}

	var in clientJSON
	err = json.Unmarshal(body, &in)
	if err != nil || in.Input == nil {
		res.Write([]byte("!Malformed input"))
		return
	}

	pLock.RLock()
	p, ok := players[in.Sid]
	pLock.RUnlock()

	if !ok {
		res.Write([]byte("!Invalid session"))
		return
	}

	p.lastReceive = time.Now().UnixNano()
	for _, message := range in.Input {
		p.Msg2Server(message)
	}

	resMsg := append(p.Msg2Client(),
		map[string]interface{}{
			"t": "pp",
			"p": in.Ping,
		})

	b, _ := json.Marshal(resMsg)
	if b == nil || err != nil {
		res.Write([]byte("!Internal JSON error"))
		return
	}
	res.Write(b)
}

func (p *player) Msg2Server(message map[string]interface{}) (fail bool) {
	op := p.opponent
	if op == nil {
		return true
	}

	switch message["t"] {
	case "t": // transfer control
		y, _ := message["y"].(float64)
		w, _ := message["w"].(float64)
		z, _ := message["z"].(float64)
		op.oppBallPos = map[string]interface{}{
			"t": "t",
			"y": y, "w": w, "z": z,
		}

	case "b": // ball position
		// FIXME check for who has control?
		x, _ := message["x"].(float64)
		y, _ := message["y"].(float64)
		w, _ := message["w"].(float64)
		z, _ := message["z"].(float64)
		op.oppBallPos = map[string]interface{}{
			"t": "b",
			"x": x, "y": y, "w": w, "z": z,
		}

	case "l": // loses
		// FIXME record loss?
		p.msgStart = 1
		op.msgStart = 2

	case "p": // player position
		x, _ := message["x"].(float64)
		y, _ := message["y"].(float64)
		w, _ := message["w"].(float64)
		z, _ := message["z"].(float64)
		op.oppMovePos = map[string]interface{}{
			"t": "p",
			"x": x, "y": y, "w": w, "z": z,
		}

	case "pp": // player ping
		pingNum, _ := message["p"].(float64)
		ping := int(pingNum)
		if ping < 1 {
			ping = 1
		} else if ping > 9999 {
			ping = 9999
		}
		op.oppPing = ping

	default:
		return true
	}
	return false
}

func (p *player) Msg2Client() (resMsg []map[string]interface{}) {

	if p.oppDisc {
		resMsg = append(resMsg, map[string]interface{}{
			"t": "d",
		})
		p.oppDisc = false
	}

	if p.oppJoin {
		op := p.opponent
		if op != nil {
			resMsg = append(resMsg, map[string]interface{}{
				"t": "j",
				"n": op.name,
				"c": op.color,
			})
		}
		p.oppJoin = false
	}

	if p.oppMovePos != nil {
		resMsg = append(resMsg, p.oppMovePos)
		p.oppMovePos = nil
	}

	if p.oppBallPos != nil {
		resMsg = append(resMsg, p.oppBallPos)
		p.oppBallPos = nil
	}

	if p.oppPing != 0 {
		resMsg = append(resMsg, map[string]interface{}{
			"t": "op",
			"p": p.oppPing,
		})
		p.oppPing = 0
	}

	if p.msgStart != 0 {
		resMsg = append(resMsg, map[string]interface{}{
			"t": "s",
			"s": p.msgStart != 1,
		})
		p.msgStart = 0
	}

	return
}
