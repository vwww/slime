package duel

import (
	"math/rand"
)

// Arena constants
const (
	// Width
	MAX_W = 1600.0
	// Height
	MAX_H = 900.0
)

// Player constants
const (
	// Movement speed (per second)
	PL_SPEED = 200.0

	// Starting size
	PL_RAD_START  = 20
	PL_MASS_START = PL_RAD_START * PL_RAD_START

	// Minimum size
	PL_RAD_MIN  = 15
	PL_MASS_MIN = PL_RAD_MIN * PL_RAD_MIN

	// Maximum size
	PL_RAD_MAX  = 900
	PL_MASS_MAX = PL_RAD_MAX * PL_RAD_MAX

	// Decay (1/2^-x per frame)
	PL_MASS_DECAY_SHIFT = 10
)

/*
	func clamp(f *float64, min, max float64) bool {
		if *f < min {
			*f = min
			return true
		} else if *f > max {
			*f = max
			return true
		}
		return false
	}

	func clampAbs(f *float64, magnitude float64) bool {
		return clamp(f, -magnitude, +magnitude)
	}
*/

func botThinkPlayer(g *Game, p *Player) {
	if p.BotDivider == 0 {
		p.BotDivider = 2 * PHYS_FPS
	} else {
		p.BotDivider--
		return
	}
	best := p.D
	bestDist2 := 1e200
	for i := range g.players {
		pp := &g.players[i]
		if !pp.IsAlive || p == pp || pp.M > p.M {
			continue
		}
		dist2 := p.O.Sub(pp.O).LengthSquared()
		if bestDist2 > dist2 {
			best = pp.O
			bestDist2 = dist2
		}
	}
	p.D = best
}

func movePlayer(p *Player) {
	diff := p.D.Sub(p.O)
	moveDist := PL_SPEED / PHYS_FPS
	if moveDist*moveDist < diff.LengthSquared() {
		diff = diff.Normalize().Mul(moveDist)
	}

	p.O = p.O.Add(diff)
}

func decayPlayer(p *Player) {
	decayMass := p.M >> PL_MASS_DECAY_SHIFT
	newMass := p.M - decayMass
	if newMass < PL_MASS_MIN {
		newMass = PL_MASS_MIN
	}
	p.setMass(newMass)
}

func collide(a, b *Player) bool {
	diff := a.O.Sub(b.O)
	dist := a.R + b.R
	return diff.X <= dist &&
		diff.Y <= dist &&
		diff.LengthSquared() <= dist*dist
}

func (g *Game) checkCollision(a, b *Player, aCn, bCn int) {
	if !collide(a, b) {
		return
	}

	// Calculate probability that Player A wins
	p := 0.1 + float64(a.M)/float64(a.M+b.M)*0.8
	aIsBot := a.Client == nil
	bIsBot := b.Client == nil
	if aIsBot != bIsBot {
		p *= 0.1
		if bIsBot {
			p += 0.9
		}
	}
	if rand.Float64() >= p {
		// Player B wins
		a, b = b, a
		aCn, bCn = bCn, aCn
	}

	// 75% of mass is transferable
	newMass := a.M + b.M // - (b.M >> 1)
	// check against maximum and for overflow
	if newMass > PL_MASS_MAX || newMass < a.M {
		newMass = PL_MASS_MAX
	}
	a.setMass(newMass)

	a.Kills++
	a.Combo++
	b.Deaths++
	b.Combo = 0
	b.IsAlive = false
	g.Broadcast(MsgDeath(aCn, bCn))
}

func (g *Game) spawnPlayer(p *Player) {
BRUTE_FORCE_SPAWN_POS:
	for i := 0; i < 256; i++ {
		p.O.X = rand.Float64() * MAX_W
		p.O.Y = rand.Float64() * MAX_H

		for j := range g.players {
			pp := &g.players[j]
			if collide(p, pp) {
				continue BRUTE_FORCE_SPAWN_POS
			}
		}
		break
	}

	p.D.X = p.O.X
	p.D.Y = p.O.Y
	p.M = PL_MASS_START
	p.R = PL_RAD_START
	p.IsAlive = true
	p.BotDivider = 0
}

// PhysicsFrame applies physics by moving all objects for a time increment of PHYS_TIME.
func (g *Game) PhysicsFrame() {
	for i := range g.players {
		p := &g.players[i]
		if !p.IsValid {
			continue
		} else if p.IsAlive {
			if p.Client == nil {
				botThinkPlayer(g, p)
			}
			movePlayer(p)
			decayPlayer(p)

			// check only against higher players,
			// to avoid double-checking
			for j := i + 1; j < len(g.players); j++ {
				b := &g.players[j]
				if !b.IsAlive {
					continue
				}
				g.checkCollision(p, b, i, j)
				// Stop if the player died
				if !p.IsAlive {
					break
				}
			}
		} else {
			// TODO use spawn queue
			// force spawn
			g.spawnPlayer(p)
		}
	}
}
