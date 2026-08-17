package game

import (
	"bytes"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	SampleRate = 44100
)

var (
	audioContext *audio.Context
	audioOnce    sync.Once

	// Pre-synthesized PCM buffers
	launchPCM    []byte
	explosionPCM []byte
	tallyPCM     []byte
	sirenPCM     []byte
	siloLowPCM   []byte
	cantFirePCM  []byte
	bonusCityPCM []byte
	gameOverPCM  []byte
	bomberPCM    []byte
	satPCM       []byte
	smartBombPCM []byte

	// Active looping players
	bomberPlayer    *audio.Player
	satPlayer       *audio.Player
	smartBombPlayer *audio.Player
	sirenPlayer     *audio.Player
	audioMu         sync.Mutex
)

// InitAudio initializes the audio context and synthesizes all POKEY waveforms.
func InitAudio() {
	audioOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				// Audio driver unavailable
			}
		}()

		audioContext = audio.NewContext(SampleRate)

		launchPCM = synthLaunch()
		explosionPCM = synthExplosion()
		tallyPCM = synthTally()
		sirenPCM = synthSiren()
		siloLowPCM = synthSiloLow()
		cantFirePCM = synthCantFire()
		bonusCityPCM = synthBonusCity()
		gameOverPCM = synthGameOver()
		bomberPCM = synthBomber()
		satPCM = synthSatellite()
		smartBombPCM = synthSmartBomb()
	})
}

func playSoundSafe(buf []byte, vol float64) {
	if audioContext == nil || buf == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	p := audioContext.NewPlayerFromBytes(buf)
	if p != nil {
		p.SetVolume(vol)
		p.Play()
	}
}

// --- Event Sound Effects ---

// PlayLaunchSound plays the classic high-to-low ABM rocket launch whistle.
func PlayLaunchSound() {
	playSoundSafe(launchPCM, 0.20)
}

// PlayExplosionSound plays the signature deep, crunchy POKEY noise explosion rumble.
func PlayExplosionSound() {
	playSoundSafe(explosionPCM, 0.38)
}

// PlayTallySound plays the rapid electronic blip used during end-of-wave counting.
func PlayTallySound() {
	playSoundSafe(tallyPCM, 0.18)
}

// PlaySiloLowSound plays the urgent dual-tone warning alarm when a silo has <= 3 missiles.
func PlaySiloLowSound() {
	playSoundSafe(siloLowPCM, 0.25)
}

// PlayCantFireSound plays the low buzzy warning click when trying to fire an empty or ruined silo.
func PlayCantFireSound() {
	playSoundSafe(cantFirePCM, 0.22)
}

// PlayBonusCitySound plays the joyful randomized electronic fanfare when earning a bonus city.
func PlayBonusCitySound() {
	playSoundSafe(bonusCityPCM, 0.28)
}

// PlayGameOverSound plays the massive apocalyptic nuclear detonation rumble for "THE END".
func PlayGameOverSound() {
	playSoundSafe(gameOverPCM, 0.45)
}

// PlaySirenAlert plays the 1.2-second wave start klaxon alert siren.
func PlaySirenAlert() {
	playSoundSafe(sirenPCM, 0.20)
}

// --- Continuous Sound Effects ---

// SetBomberSound turns the bomber engine rumble on or off.
func SetBomberSound(active bool) {
	audioMu.Lock()
	defer audioMu.Unlock()
	if audioContext == nil {
		return
	}
	if active {
		if bomberPlayer == nil {
			r := bytes.NewReader(bomberPCM)
			loop := audio.NewInfiniteLoop(r, int64(len(bomberPCM)))
			p, err := audioContext.NewPlayer(loop)
			if err == nil {
				p.SetVolume(0.16)
				bomberPlayer = p
				bomberPlayer.Play()
			}
		}
	} else {
		if bomberPlayer != nil {
			_ = bomberPlayer.Close()
			bomberPlayer = nil
		}
	}
}

// SetSatelliteSound turns the satellite high telemetry beep on or off.
func SetSatelliteSound(active bool) {
	audioMu.Lock()
	defer audioMu.Unlock()
	if audioContext == nil {
		return
	}
	if active {
		if satPlayer == nil {
			r := bytes.NewReader(satPCM)
			loop := audio.NewInfiniteLoop(r, int64(len(satPCM)))
			p, err := audioContext.NewPlayer(loop)
			if err == nil {
				p.SetVolume(0.14)
				satPlayer = p
				satPlayer.Play()
			}
		}
	} else {
		if satPlayer != nil {
			_ = satPlayer.Close()
			satPlayer = nil
		}
	}
}

// SetSmartBombSound turns the smart bomb warble on or off.
func SetSmartBombSound(active bool) {
	audioMu.Lock()
	defer audioMu.Unlock()
	if audioContext == nil {
		return
	}
	if active {
		if smartBombPlayer == nil {
			r := bytes.NewReader(smartBombPCM)
			loop := audio.NewInfiniteLoop(r, int64(len(smartBombPCM)))
			p, err := audioContext.NewPlayer(loop)
			if err == nil {
				p.SetVolume(0.18)
				smartBombPlayer = p
				smartBombPlayer.Play()
			}
		}
	} else {
		if smartBombPlayer != nil {
			_ = smartBombPlayer.Close()
			smartBombPlayer = nil
		}
	}
}

// StopAllContinuousSounds silences all looping sounds (e.g. at wave end or game over).
func StopAllContinuousSounds() {
	SetBomberSound(false)
	SetSatelliteSound(false)
	SetSmartBombSound(false)
}

// --- POKEY NOISE & SOUND SYNTHESIS ALGORITHMS ---

// LFSR noise generator simulation for Atari POKEY 4-bit, 5-bit, and 17-bit registers.
type PokeyLFSR struct {
	poly4  uint8
	poly5  uint8
	poly17 uint32
}

func newPokeyLFSR() *PokeyLFSR {
	return &PokeyLFSR{
		poly4:  0x0F,
		poly5:  0x1F,
		poly17: 0x1FFFF,
	}
}

func (l *PokeyLFSR) stepPoly17() float64 {
	// 17-bit polynomial: bit 16 XOR bit 13
	bit := ((l.poly17 >> 16) ^ (l.poly17 >> 13)) & 1
	l.poly17 = ((l.poly17 << 1) | bit) & 0x1FFFF
	if (l.poly17 & 1) != 0 {
		return 1.0
	}
	return -1.0
}

func (l *PokeyLFSR) stepPoly5Poly4() float64 {
	// Combined 5-bit and 4-bit polynomial for gritty metallic crunch
	b5 := ((l.poly5 >> 4) ^ (l.poly5 >> 2)) & 1
	l.poly5 = ((l.poly5 << 1) | b5) & 0x1F

	b4 := ((l.poly4 >> 3) ^ (l.poly4 >> 2)) & 1
	l.poly4 = ((l.poly4 << 1) | b4) & 0x0F

	val := float64((l.poly5&1)^(l.poly4&1))*2.0 - 1.0
	return val
}

func write16BitLE(buf []byte, idx int, val int16) {
	buf[idx] = byte(val)
	buf[idx+1] = byte(val >> 8)
}

// synthLaunch creates the authentic ABM missile sweep down from 950Hz to 160Hz.
func synthLaunch() []byte {
	duration := 0.28
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		freq := 950.0*math.Exp(-4.0*progress) + 160.0

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		var sampleVal int16 = 6000
		if math.Sin(phase) < 0 {
			sampleVal = -6000
		}

		if progress > 0.8 {
			envelope := (1.0 - progress) / 0.2
			sampleVal = int16(float64(sampleVal) * envelope)
		}

		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthExplosion generates the authentic multi-stage deep crunchy POKEY explosion rumble.
func synthExplosion() []byte {
	duration := 0.85
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	lfsr := newPokeyLFSR()
	sampleHold := 14 // Low sampling frequency for heavy bass rumble
	noiseVal := 0.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		if i%sampleHold == 0 {
			noiseVal = lfsr.stepPoly17()*0.7 + lfsr.stepPoly5Poly4()*0.3
		}

		// Nonlinear envelope: sharp attack, sustained body, exponential decay
		envelope := math.Exp(-3.2 * progress)
		sampleVal := int16(noiseVal * envelope * 14000.0)
		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthTally creates the crisp 1200Hz tally blip for end-of-wave accounting (35ms).
func synthTally() []byte {
	duration := 0.035
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	freq := 1200.0
	phase := 0.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		var sampleVal int16 = 7000
		if math.Sin(phase) < 0 {
			sampleVal = -7000
		}

		envelope := 1.0 - progress
		sampleVal = int16(float64(sampleVal) * envelope)
		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthSiren generates the classic wave-start alarm klaxon oscillating 480Hz-720Hz at 5Hz LFO (1.2s).
func synthSiren() []byte {
	duration := 1.2
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		lfo := math.Sin(2.0 * math.Pi * 5.0 * t)
		freq := 600.0 + 120.0*lfo

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		var sampleVal int16 = 4500
		if math.Sin(phase) < 0 {
			sampleVal = -4500
		}

		if progress > 0.85 {
			envelope := (1.0 - progress) / 0.15
			sampleVal = int16(float64(sampleVal) * envelope)
		}

		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthSiloLow creates the fast dual-tone 880Hz/440Hz alarm blip for low ammo warning (80ms).
func synthSiloLow() []byte {
	duration := 0.09
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		freq := 880.0
		if t > duration/2.0 {
			freq = 440.0
		}

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		var sampleVal int16 = 6500
		if math.Sin(phase) < 0 {
			sampleVal = -6500
		}

		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthCantFire generates the low 110Hz warning buzz when clicking an empty/ruined base (120ms).
func synthCantFire() []byte {
	duration := 0.12
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	lfsr := newPokeyLFSR()
	phase := 0.0
	freq := 110.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		sq := 1.0
		if math.Sin(phase) < 0 {
			sq = -1.0
		}

		noise := lfsr.stepPoly5Poly4()
		sampleVal := int16((sq*0.7 + noise*0.3) * (1.0 - progress*0.5) * 8000.0)
		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthBonusCity generates the ascending 6-tone chime sequence played when earning a bonus city (0.6s).
func synthBonusCity() []byte {
	pitches := []float64{523.25, 659.25, 783.99, 1046.50, 1318.51, 1567.98} // C5, E5, G5, C6, E6, G6
	toneDur := 0.09
	numSamples := int(SampleRate * toneDur * float64(len(pitches)))
	buf := make([]byte, numSamples*2)

	phase := 0.0
	for toneIdx, freq := range pitches {
		samplesPerTone := int(SampleRate * toneDur)
		startSample := toneIdx * samplesPerTone
		for s := 0; s < samplesPerTone; s++ {
			sampleIdx := startSample + s
			progress := float64(s) / float64(samplesPerTone)

			phase += 2.0 * math.Pi * freq / SampleRate
			if phase > 2.0*math.Pi {
				phase -= 2.0 * math.Pi
			}

			var sampleVal int16 = 6000
			if math.Sin(phase) < 0 {
				sampleVal = -6000
			}

			envelope := 1.0 - progress*0.6
			sampleVal = int16(float64(sampleVal) * envelope)
			write16BitLE(buf, sampleIdx*2, sampleVal)
		}
	}
	return buf
}

// synthGameOver creates the massive apocalyptic 2.5s nuclear detonation rumble for "THE END".
func synthGameOver() []byte {
	duration := 2.5
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	lfsr := newPokeyLFSR()
	sampleHold := 18
	noiseVal := 0.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		if i%sampleHold == 0 {
			noiseVal = lfsr.stepPoly17()
		}

		// Low rumble sweep
		envelope := math.Exp(-1.8 * progress)
		sampleVal := int16(noiseVal * envelope * 15000.0)
		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthBomber creates a 0.5s seamless looping low engine hum (60Hz poly noise).
func synthBomber() []byte {
	duration := 0.5
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	freq := 62.0

	for i := 0; i < numSamples; i++ {
		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		var sampleVal int16 = 3000
		if math.Sin(phase) < 0 {
			sampleVal = -3000
		}
		// Add subtle harmonic
		if math.Sin(phase*2.0) < 0 {
			sampleVal += 1200
		} else {
			sampleVal -= 1200
		}

		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthSatellite creates a 0.5s seamless looping high telemetry pulse (1400Hz).
func synthSatellite() []byte {
	duration := 0.5
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	freq := 1350.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		// 8Hz on/off modulation
		mod := math.Sin(2.0 * math.Pi * 8.0 * t)

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		var sampleVal int16 = 0
		if mod > 0 {
			if math.Sin(phase) > 0 {
				sampleVal = 3200
			} else {
				sampleVal = -3200
			}
		}

		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthSmartBomb creates a 0.5s seamless looping 220Hz rhythmic pulsing buzz.
func synthSmartBomb() []byte {
	duration := 0.5
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	freq := 220.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		mod := math.Sin(2.0 * math.Pi * 12.0 * t)

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		var sampleVal int16 = 0
		if mod > 0.2 {
			if math.Sin(phase) > 0 {
				sampleVal = 4000
			} else {
				sampleVal = -4000
			}
		}

		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}
