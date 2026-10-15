package game

import (
	"math"
	"math/rand"
)

// stepSimulation runs one frame of entity spawning, movement, and collision detection.
func (g *Game) stepSimulation() {
	g.updateSpawning()
	g.updateFlier()
	g.updateABMs()
	g.updateICBMs()
	g.updateSmartBombs()
	g.updateExplosions()
	g.checkCollisions()
}

func (g *Game) survivingCitiesCount() int {
	count := 0
	for _, city := range g.Cities {
		if !city.Destroyed {
			count++
		}
	}
	return count
}

func (g *Game) restoreContinuousSounds() {
	if g.ActiveFlier != nil && g.ActiveFlier.Active {
		if g.ActiveFlier.Type == FlierBomber {
			SetBomberSound(true)
		} else {
			SetSatelliteSound(true)
		}
	}
	for _, sb := range g.SmartBombs {
		if sb.Active {
			SetSmartBombSound(true)
			break
		}
	}
}

func (g *Game) fireABM(siloIdx int) {
	// Authentic arcade limit: Max 8 simultaneous ABMs on screen
	if len(g.ABMs) >= 8 {
		return
	}

	bat := &g.Batteries[siloIdx]
	if bat.Destroyed || bat.Ammo <= 0 {
		PlayCantFireSound()
		return
	}

	bat.Ammo--
	if bat.Ammo <= 3 && !bat.LowWarningPlayed {
		PlaySiloLowSound()
		bat.LowWarningPlayed = true
	}

	// Center silo (Delta) fires at 7.5 units/frame; side silos fire at 3.2 units/frame
	launchSpeed := 3.2
	if siloIdx == 1 {
		launchSpeed = 7.5
	}

	g.ABMs = append(g.ABMs, &ABM{
		Start:     bat.Position,
		Curr:      bat.Position,
		Target:    g.Crosshair.Pos,
		SiloIndex: siloIdx,
		Speed:     launchSpeed,
		Active:    true,
	})
	PlayLaunchSound()
}

func (g *Game) updateSpawning() {
	// Limit simultaneous active attacks to 8 (Smart bombs count as 2)
	if len(g.ICBMs)+len(g.SmartBombs)*2 >= 8 {
		return
	}
	if g.ICBMsRemaining <= 0 && g.SmartBombsLeft <= 0 {
		return
	}

	// Authentic arcade altitude pacing: block new spawns while any threat is above thresholdY
	thresholdY := math.Min(51.0, 29.0+float64(g.Wave)*2.0)
	for _, icbm := range g.ICBMs {
		if icbm.Active && icbm.Curr.Y < thresholdY {
			return
		}
	}
	for _, sb := range g.SmartBombs {
		if sb.Active && sb.Curr.Y < thresholdY {
			return
		}
	}

	g.SpawnCooldown--
	if g.SpawnCooldown > 0 {
		return
	}

	wdata := GetWaveData(g.Wave)
	canSpawnSB := g.SmartBombsLeft > 0 && len(g.SmartBombs) < 3

	if canSpawnSB && (g.ICBMsRemaining <= 0 || rand.Float64() < 0.35) {
		g.SmartBombsLeft--
		spawnX := 20.0 + rand.Float64()*(SimWidth-40.0)
		g.SmartBombs = append(g.SmartBombs, &SmartBomb{
			Start:  Point{X: spawnX, Y: 0},
			Curr:   Point{X: spawnX, Y: 0},
			Target: g.pickRandomTarget(),
			Speed:  wdata.Speed,
			Active: true,
		})
		SetSmartBombSound(true)
	} else if g.ICBMsRemaining > 0 {
		g.ICBMsRemaining--
		spawnX := rand.Float64() * SimWidth
		isSplit := g.Wave >= 2 && g.ICBMsRemaining >= 2 && rand.Float64() < 0.25

		g.ICBMs = append(g.ICBMs, &ICBM{
			Start:         Point{X: spawnX, Y: 0},
			Curr:          Point{X: spawnX, Y: 0},
			Target:        g.pickRandomTarget(),
			Speed:         wdata.Speed,
			IsSplitter:    isSplit,
			SplitAltitude: 72.0 + rand.Float64()*31.0,
			Active:        true,
		})
	}

	minCD := math.Max(20, 55-float64(g.Wave)*2.5)
	maxCD := math.Max(40, 95-float64(g.Wave)*4.0)
	g.SpawnCooldown = int(minCD) + rand.Intn(int(maxCD-minCD+1))
}

func (g *Game) pickRandomTarget() Point {
	var targets []Point
	for _, city := range g.Cities {
		if !city.Destroyed {
			targets = append(targets, Point{X: city.Position.X + 8, Y: 216})
		}
	}
	for _, bat := range g.Batteries {
		if !bat.Destroyed {
			targets = append(targets, bat.Position)
		}
	}
	if len(targets) > 0 {
		return targets[rand.Intn(len(targets))]
	}
	return Point{X: 20 + rand.Float64()*(SimWidth-40), Y: 216}
}

func (g *Game) updateFlier() {
	if g.ActiveFlier == nil {
		if !g.FlierSpawnedThisWave && g.ICBMsRemaining > 2 {
			g.FlierCooldown--
			if g.FlierCooldown <= 0 {
				g.FlierSpawnedThisWave = true
				ft := FlierBomber
				alt := 65.0 + rand.Float64()*25.0
				speed := 1.0 / 3.0
				if rand.Float64() < 0.45 {
					ft = FlierSatellite
					alt = 28.0 + rand.Float64()*18.0
					speed = 0.50
				}

				startX := -16.0
				if rand.Float64() >= 0.5 {
					startX = SimWidth + 16.0
					speed = -speed
				}

				g.ActiveFlier = &Flier{
					Type:           ft,
					X:              startX,
					Y:              alt,
					Speed:          speed,
					DropCooldown:   40 + rand.Intn(50),
					BombsRemaining: 2,
					Active:         true,
				}

				if ft == FlierBomber {
					SetBomberSound(true)
				} else {
					SetSatelliteSound(true)
				}
			}
		}
		return
	}

	flier := g.ActiveFlier
	flier.AnimTick++
	flier.X += flier.Speed

	if (flier.Speed > 0 && flier.X > SimWidth+20) || (flier.Speed < 0 && flier.X < -20) {
		flier.Active = false
		g.ActiveFlier = nil
		SetBomberSound(false)
		SetSatelliteSound(false)
		return
	}

	if flier.BombsRemaining > 0 && flier.X >= 15 && flier.X <= SimWidth-15 {
		flier.DropCooldown--
		if flier.DropCooldown <= 0 {
			flier.BombsRemaining--
			flier.DropCooldown = 60 + rand.Intn(40)
			if g.ICBMsRemaining > 0 {
				g.ICBMsRemaining--
			}

			wdata := GetWaveData(g.Wave)
			g.ICBMs = append(g.ICBMs, &ICBM{
				Start:  Point{X: flier.X, Y: flier.Y},
				Curr:   Point{X: flier.X, Y: flier.Y},
				Target: g.pickRandomTarget(),
				Speed:  wdata.Speed,
				Active: true,
			})
		}
	}
}

func (g *Game) updateABMs() {
	active := g.ABMs[:0]
	for _, abm := range g.ABMs {
		if !abm.Active {
			continue
		}
		dist := abm.Start.Distance(abm.Target)
		if dist <= 0 {
			abm.Active = false
			continue
		}

		abm.Progress += abm.Speed / dist
		if abm.Progress >= 1.0 {
			abm.Active = false
			g.spawnExplosion(abm.Target, 22.0, 18)
		} else {
			abm.Curr = abm.Start.Lerp(abm.Target, abm.Progress)
			active = append(active, abm)
		}
	}
	g.ABMs = active
}

func (g *Game) updateICBMs() {
	var active []*ICBM
	for _, icbm := range g.ICBMs {
		if !icbm.Active {
			continue
		}

		// MIRV Split Logic: sister warheads count toward wave missile quota
		if icbm.IsSplitter && !icbm.Splitted && icbm.Curr.Y >= icbm.SplitAltitude {
			icbm.Splitted = true
			sistersToSpawn := min(2, max(0, g.ICBMsRemaining))
			g.ICBMsRemaining -= sistersToSpawn

			for s := 0; s < sistersToSpawn; s++ {
				active = append(active, &ICBM{
					Start:  icbm.Curr,
					Curr:   icbm.Curr,
					Target: g.pickRandomTarget(),
					Speed:  icbm.Speed,
					Active: true,
				})
			}
		}

		dist := icbm.Start.Distance(icbm.Target)
		if dist <= 0 {
			icbm.Active = false
			continue
		}

		icbm.Progress += icbm.Speed / dist
		if icbm.Progress >= 1.0 || icbm.Curr.Y >= 216 {
			icbm.Active = false
			g.detonateThreatOnGround(icbm.Target)
		} else {
			icbm.Curr = icbm.Start.Lerp(icbm.Target, icbm.Progress)
			active = append(active, icbm)
		}
	}
	g.ICBMs = active
}

func (g *Game) updateSmartBombs() {
	var active []*SmartBomb
	for _, sb := range g.SmartBombs {
		if !sb.Active {
			continue
		}
		sb.AnimTick++

		// Smart Bomb Evasion AI: scan for nearby expanding explosions
		evading := false
		for _, exp := range g.Explosions {
			if exp.State == StateDead {
				continue
			}
			if sb.Curr.Distance(exp.Center) < exp.Radius+18.0 {
				evading = true
				dx := sb.Curr.X - exp.Center.X
				dy := sb.Curr.Y - exp.Center.Y
				if math.Abs(dx) < 0.1 {
					if rand.Float64() < 0.5 {
						dx = 1.0
					} else {
						dx = -1.0
					}
				}
				if length := math.Hypot(dx, dy); length > 0 {
					sb.Curr.X += (dx / length) * sb.Speed * 1.25
					sb.Curr.Y += sb.Speed * 0.40
				}
				break
			}
		}

		if !evading {
			dist := sb.Curr.Distance(sb.Target)
			if dist <= sb.Speed {
				sb.Active = false
				g.detonateThreatOnGround(sb.Target)
				continue
			}
			sb.Curr.X += ((sb.Target.X - sb.Curr.X) / dist) * sb.Speed
			sb.Curr.Y += ((sb.Target.Y - sb.Curr.Y) / dist) * sb.Speed
		}

		sb.Curr.X = clamp(sb.Curr.X, 4, SimWidth-4)
		if sb.Curr.Y >= 216 {
			sb.Active = false
			g.detonateThreatOnGround(sb.Curr)
		} else {
			active = append(active, sb)
		}
	}
	g.SmartBombs = active

	if len(g.SmartBombs) == 0 {
		SetSmartBombSound(false)
	}
}

func (g *Game) spawnExplosion(center Point, maxRadius float64, holdTimer int) {
	g.Explosions = append(g.Explosions, &Explosion{
		Center:    center,
		MaxRadius: maxRadius,
		HoldTimer: holdTimer,
		State:     StateExpanding,
	})
	PlayExplosionSound()
}

func (g *Game) detonateThreatOnGround(pos Point) {
	for i := range g.Cities {
		if !g.Cities[i].Destroyed {
			cx := g.Cities[i].Position.X + 8
			if math.Abs(pos.X-cx) < 12.0 && math.Abs(pos.Y-216) < 8.0 {
				g.Cities[i].Destroyed = true
				g.pipeline.TriggerFlash(8)
				break
			}
		}
	}

	for i := range g.Batteries {
		if !g.Batteries[i].Destroyed {
			bx := g.Batteries[i].Position.X
			if math.Abs(pos.X-bx) < 14.0 && math.Abs(pos.Y-214) < 8.0 {
				g.Batteries[i].Destroyed = true
				g.Batteries[i].Ammo = 0
				g.pipeline.TriggerFlash(6)
				break
			}
		}
	}

	g.spawnExplosion(pos, 18.0, 14)
}

func (g *Game) updateExplosions() {
	active := g.Explosions[:0]
	for _, exp := range g.Explosions {
		if exp.State == StateDead {
			continue
		}
		exp.Tick++

		switch exp.State {
		case StateExpanding:
			exp.Radius += 0.65
			if exp.Radius >= exp.MaxRadius {
				exp.State = StateHolding
			}
		case StateHolding:
			exp.HoldTimer--
			if exp.HoldTimer <= 0 {
				exp.State = StateContracting
			}
		case StateContracting:
			exp.Radius -= 0.55
			if exp.Radius <= 0 {
				exp.State = StateDead
			}
		}

		if exp.State != StateDead {
			active = append(active, exp)
		}
	}
	g.Explosions = active
}

func (g *Game) isHitByExplosion(pos Point, extraRadius float64) bool {
	for _, exp := range g.Explosions {
		if exp.State != StateDead && pos.Distance(exp.Center) <= exp.Radius+extraRadius {
			return true
		}
	}
	return false
}

func (g *Game) checkCollisions() {
	for _, icbm := range g.ICBMs {
		if icbm.Active && g.isHitByExplosion(icbm.Curr, 0) {
			icbm.Active = false
			g.Score += 25 * g.Palette.Multiplier
			g.spawnExplosion(icbm.Curr, 18.0, 14)
		}
	}

	for _, sb := range g.SmartBombs {
		if sb.Active && g.isHitByExplosion(sb.Curr, 0) {
			sb.Active = false
			g.Score += 125 * g.Palette.Multiplier
			g.spawnExplosion(sb.Curr, 18.0, 14)
		}
	}

	if g.ActiveFlier != nil && g.ActiveFlier.Active {
		fPos := Point{X: g.ActiveFlier.X + 8, Y: g.ActiveFlier.Y + 4}
		if g.isHitByExplosion(fPos, 6.0) {
			g.ActiveFlier.Active = false
			g.ActiveFlier = nil
			SetBomberSound(false)
			SetSatelliteSound(false)
			g.Score += 100 * g.Palette.Multiplier
			g.spawnExplosion(fPos, 22.0, 16)
		}
	}
}

func (g *Game) checkWaveCompletion() {
	if g.survivingCitiesCount() == 0 {
		StopAllContinuousSounds()
		g.State = StateTheEnd
		g.TheEndTimer = 190
		g.pipeline.TriggerFlash(25)
		PlayGameOverSound()
		return
	}

	noThreatsRemaining := g.ICBMsRemaining <= 0 && g.SmartBombsLeft <= 0 &&
		len(g.ICBMs) == 0 && len(g.SmartBombs) == 0 && g.ActiveFlier == nil

	if noThreatsRemaining {
		g.WaveEndGraceTimer++
		if (len(g.ABMs) == 0 && len(g.Explosions) == 0) || g.WaveEndGraceTimer >= 45 {
			StopAllContinuousSounds()
			g.State = StateTally
			g.TallyPhase = 0
			g.TallyTimer = 30
			g.TallyBatteryIdx = 0
			g.TallyCityIdx = 0
			g.TallyScoreEarned = 0
			g.WaveEndGraceTimer = 0
		}
	} else {
		g.WaveEndGraceTimer = 0
	}
}
