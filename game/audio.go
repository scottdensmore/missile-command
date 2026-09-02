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

	// Master volume & mute
	masterVolume float64 = 1.0
	isMuted      bool    = false
)

// SetMasterVolume sets the global audio output volume (clamped between 0.0 and 1.0).
func SetMasterVolume(v float64) {
	audioMu.Lock()
	defer audioMu.Unlock()
	if v < 0.0 {
		v = 0.0
	} else if v > 1.0 {
		v = 1.0
	}
	masterVolume = v
	updateLoopingVolumesLocked()
}

// GetMasterVolume returns the current global audio output volume.
func GetMasterVolume() float64 {
	audioMu.Lock()
	defer audioMu.Unlock()
	return masterVolume
}

// ToggleMute toggles sound output mute on or off.
func ToggleMute() bool {
	audioMu.Lock()
	defer audioMu.Unlock()
	isMuted = !isMuted
	updateLoopingVolumesLocked()
	return isMuted
}

// IsMuted returns true if sound output is currently muted.
func IsMuted() bool {
	audioMu.Lock()
	defer audioMu.Unlock()
	return isMuted
}

// AdjustVolume increases or decreases global volume by delta and returns the new level.
func AdjustVolume(delta float64) float64 {
	audioMu.Lock()
	defer audioMu.Unlock()
	masterVolume += delta
	if masterVolume < 0.0 {
		masterVolume = 0.0
	} else if masterVolume > 1.0 {
		masterVolume = 1.0
	}
	updateLoopingVolumesLocked()
	return masterVolume
}

func updateLoopingVolumesLocked() {
	volScale := masterVolume
	if isMuted {
		volScale = 0.0
	}
	if bomberPlayer != nil {
		bomberPlayer.SetVolume(0.16 * volScale)
	}
	if satPlayer != nil {
		satPlayer.SetVolume(0.14 * volScale)
	}
	if smartBombPlayer != nil {
		smartBombPlayer.SetVolume(0.18 * volScale)
	}
}

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
	audioMu.Lock()
	effectiveVol := vol * masterVolume
	if isMuted {
		effectiveVol = 0.0
	}
	audioMu.Unlock()

	if audioContext == nil || buf == nil || effectiveVol <= 0.0 {
		return
	}
	defer func() {
		_ = recover()
	}()
	p := audioContext.NewPlayerFromBytes(buf)
	if p != nil {
		p.SetVolume(effectiveVol)
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
	playSoundSafe(explosionPCM, 0.45)
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

// --- POKEY HARDWARE EMULATION & SYNTHESIS ---

// PokeyLFSR emulates the exact Atari POKEY polynomial shift registers (poly4, poly5, poly17).
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
	bit := ((l.poly17 >> 16) ^ (l.poly17 >> 13)) & 1
	l.poly17 = ((l.poly17 << 1) | bit) & 0x1FFFF
	if (l.poly17 & 1) != 0 {
		return 1.0
	}
	return -1.0
}

func (l *PokeyLFSR) stepPoly5() float64 {
	b5 := ((l.poly5 >> 4) ^ (l.poly5 >> 2)) & 1
	l.poly5 = ((l.poly5 << 1) | b5) & 0x1F
	if (l.poly5 & 1) != 0 {
		return 1.0
	}
	return -1.0
}

func (l *PokeyLFSR) stepPoly4() float64 {
	b4 := ((l.poly4 >> 3) ^ (l.poly4 >> 2)) & 1
	l.poly4 = ((l.poly4 << 1) | b4) & 0x0F
	if (l.poly4 & 1) != 0 {
		return 1.0
	}
	return -1.0
}

func write16BitLE(buf []byte, idx int, val int16) {
	buf[idx] = byte(val)
	buf[idx+1] = byte(val >> 8)
}

// pokeyFreq converts POKEY 64kHz frequency divider register AUDF to frequency in Hz.
func pokeyFreq(audf int) float64 {
	if audf < 0 {
		audf = 0
	}
	return 64000.0 / (2.0 * float64(audf+1))
}

// synthLaunch creates the authentic ABM missile sweep down using POKEY linear AUDF register progression ($20 -> $DC).
func synthLaunch() []byte {
	duration := 0.28
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		// POKEY AUDF register linearly ramps from $20 (32, ~970 Hz) up to $E0 (224, ~142 Hz)
		audf := 32.0 + progress*192.0
		freq := pokeyFreq(int(audf))

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		// POKEY Pure Tone Square Wave (4-bit quantized volume 15 -> 0)
		vol4Bit := int16((1.0 - progress*0.8) * 15.0)
		if vol4Bit < 0 {
			vol4Bit = 0
		}
		var sampleVal int16 = vol4Bit * 550
		if math.Sin(phase) < 0 {
			sampleVal = -sampleVal
		}

		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthExplosion generates the authentic POKEY 17-bit polynomial noise explosion with heavy sub-bass and deep rumble.
func synthExplosion() []byte {
	duration := 1.05
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	lfsr := newPokeyLFSR()
	noiseVal := 0.0
	subPhase := 0.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		// Deep POKEY noise clocking: sweeps from 240 Hz down to 55 Hz for heavy sub-bass rumble
		noiseClock := 240.0*math.Exp(-1.5*progress) + 55.0
		sampleHold := int(SampleRate / noiseClock)
		if sampleHold < 1 {
			sampleHold = 1
		}

		if i%sampleHold == 0 {
			// 17-bit deep rumble (80%) + 5-bit distortion crackle (20%)
			noiseVal = lfsr.stepPoly17()*0.80 + lfsr.stepPoly5()*0.20
		}

		// Sub-bass fundamental wave sweeping down from 68 Hz to 32 Hz
		subFreq := 68.0 * (1.0 - progress*0.55)
		subPhase += 2.0 * math.Pi * subFreq / SampleRate
		if subPhase > 2.0*math.Pi {
			subPhase -= 2.0 * math.Pi
		}
		subBass := math.Sin(subPhase)

		// 4-bit quantized volume envelope with sustained body and low-end tail
		var vol4Bit float64
		if progress < 0.12 {
			// Punchy initial attack
			vol4Bit = 15.0
		} else if progress < 0.45 {
			// Heavy sustained rumble body
			vol4Bit = 15.0 - (progress-0.12)*8.0
		} else {
			// Deep bass decay
			vol4Bit = 12.0 * math.Pow(1.0-(progress-0.45)/0.55, 1.3)
		}
		if vol4Bit < 0 {
			vol4Bit = 0
		}

		// Blend: 65% heavy 17-bit LFSR rumble + 35% sub-bass fundamental
		mixed := noiseVal*0.65 + subBass*0.35
		sampleVal := int16(mixed * (vol4Bit / 15.0) * 16500.0)
		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthTally creates the crisp 1520Hz square wave tally blip for bonus point counting (32ms).
func synthTally() []byte {
	duration := 0.032
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	// AUDF = $14 (~1523 Hz)
	freq := pokeyFreq(20)
	phase := 0.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		vol4Bit := int16((1.0 - progress*0.3) * 14.0)
		var sampleVal int16 = vol4Bit * 500
		if math.Sin(phase) < 0 {
			sampleVal = -sampleVal
		}

		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthSiren generates the authentic wave-start siren alternating between 780Hz ($28) and 520Hz ($3D) at 9Hz.
func synthSiren() []byte {
	duration := 1.25
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		// 9Hz alternating dual tone klaxon
		step := int(t * 18.0)
		var audf int
		if step%2 == 0 {
			audf = 40 // ~780 Hz
		} else {
			audf = 61 // ~516 Hz
		}
		freq := pokeyFreq(audf)

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		vol := 12.0
		if progress > 0.88 {
			vol = 12.0 * (1.0 - (progress-0.88)/0.12)
		}
		var sampleVal int16 = int16(vol * 500)
		if math.Sin(phase) < 0 {
			sampleVal = -sampleVal
		}

		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthSiloLow creates the fast dual-tone alarm blip (AUDF=$1E ~1032Hz, AUDF=$3C ~516Hz) (70ms).
func synthSiloLow() []byte {
	duration := 0.08
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate

		var audf int
		if t < duration/2.0 {
			audf = 30 // ~1032 Hz
		} else {
			audf = 60 // ~516 Hz
		}
		freq := pokeyFreq(audf)

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		var sampleVal int16 = 14 * 550
		if math.Sin(phase) < 0 {
			sampleVal = -sampleVal
		}

		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthCantFire generates the low 150Hz 5-bit/4-bit distortion buzzer for empty/ruined base clicks (110ms).
func synthCantFire() []byte {
	duration := 0.11
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	lfsr := newPokeyLFSR()
	clockRate := pokeyFreq(100) // ~316 Hz clocking
	sampleHold := int(SampleRate / clockRate)
	if sampleHold < 1 {
		sampleHold = 1
	}

	phase := 0.0
	freq := pokeyFreq(140) // ~226 Hz square wave
	noiseVal := 0.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		if i%sampleHold == 0 {
			noiseVal = lfsr.stepPoly5() * lfsr.stepPoly4()
		}

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		sq := 1.0
		if math.Sin(phase) < 0 {
			sq = -1.0
		}

		vol4Bit := (1.0 - progress*0.6) * 14.0
		sampleVal := int16((sq*0.6 + noiseVal*0.4) * vol4Bit * 600.0)
		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthBonusCity generates the iconic 6-tone ascending arcade fanfare arpeggio (0.50s).
func synthBonusCity() []byte {
	// Original Atari arcade pitch sequence: F5, A5, C6, E6, F6, A6
	tones := []int{45, 35, 29, 23, 22, 17} // AUDF dividers for precise POKEY pitches
	toneDur := 0.08
	numSamples := int(SampleRate * toneDur * float64(len(tones)))
	buf := make([]byte, numSamples*2)

	phase := 0.0
	for toneIdx, audf := range tones {
		freq := pokeyFreq(audf)
		samplesPerTone := int(SampleRate * toneDur)
		startSample := toneIdx * samplesPerTone

		for s := 0; s < samplesPerTone; s++ {
			sampleIdx := startSample + s
			progress := float64(s) / float64(samplesPerTone)

			phase += 2.0 * math.Pi * freq / SampleRate
			if phase > 2.0*math.Pi {
				phase -= 2.0 * math.Pi
			}

			vol4Bit := (1.0 - progress*0.4) * 14.0
			var sampleVal int16 = int16(vol4Bit * 500)
			if math.Sin(phase) < 0 {
				sampleVal = -sampleVal
			}

			write16BitLE(buf, sampleIdx*2, sampleVal)
		}
	}
	return buf
}

// synthGameOver creates the massive apocalyptic 2.8s nuclear detonation rumble for "THE END".
func synthGameOver() []byte {
	duration := 2.8
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	lfsr := newPokeyLFSR()
	noiseVal := 0.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		progress := t / duration

		// Low AUDF clocking from $80 down to $FF for deep sub-bass earthquake rumble
		audf := 128.0 + progress*127.0
		clockRate := 64000.0 / (float64(audf) + 1.0)
		sampleHold := int(SampleRate / clockRate)
		if sampleHold < 1 {
			sampleHold = 1
		}

		if i%sampleHold == 0 {
			noiseVal = lfsr.stepPoly17()
		}

		envelope := math.Exp(-1.4 * progress)
		sampleVal := int16(noiseVal * envelope * 15000.0)
		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthBomber creates a seamless 0.4s looping low engine motor drone (55Hz POKEY poly4 distortion).
func synthBomber() []byte {
	duration := 0.40
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	lfsr := newPokeyLFSR()
	clockRate := 220.0
	sampleHold := int(SampleRate / clockRate)

	phase := 0.0
	freq := 58.0 // ~58 Hz engine fundamental
	noiseVal := 0.0

	for i := 0; i < numSamples; i++ {
		if i%sampleHold == 0 {
			noiseVal = lfsr.stepPoly4()
		}

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		sq := 1.0
		if math.Sin(phase) < 0 {
			sq = -1.0
		}

		sampleVal := int16((sq*0.65 + noiseVal*0.35) * 4500.0)
		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthSatellite creates a seamless 0.4s looping high-orbit radar telemetry chirp (1680Hz pulsed beacon).
func synthSatellite() []byte {
	duration := 0.40
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	freq := pokeyFreq(18) // ~1684 Hz

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate

		// 2 rapid beeps every 400ms: beep at 0-50ms and 100-150ms, silent 150-400ms
		active := (t < 0.05) || (t >= 0.10 && t < 0.15)

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		var sampleVal int16 = 0
		if active {
			if math.Sin(phase) > 0 {
				sampleVal = 5000
			} else {
				sampleVal = -5000
			}
		}

		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthSmartBomb creates a seamless 0.4s looping 240Hz rhythmic warble with 5-bit poly buzz.
func synthSmartBomb() []byte {
	duration := 0.40
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	lfsr := newPokeyLFSR()
	clockRate := 480.0
	sampleHold := int(SampleRate / clockRate)

	phase := 0.0
	noiseVal := 0.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate

		// 7.5 Hz frequency modulation between 180Hz and 300Hz
		mod := math.Sin(2.0 * math.Pi * 7.5 * t)
		freq := 240.0 + 60.0*mod

		if i%sampleHold == 0 {
			noiseVal = lfsr.stepPoly5()
		}

		phase += 2.0 * math.Pi * freq / SampleRate
		if phase > 2.0*math.Pi {
			phase -= 2.0 * math.Pi
		}

		sq := 1.0
		if math.Sin(phase) < 0 {
			sq = -1.0
		}

		sampleVal := int16((sq*0.7 + noiseVal*0.3) * 4800.0)
		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}
