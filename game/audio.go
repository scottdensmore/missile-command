package game

import (
	"bytes"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const SampleRate = 44100

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
	audioMu         sync.Mutex

	// Master volume & mute
	masterVolume = 1.0
	isMuted      = false
)

// SetMasterVolume sets the global audio output volume (clamped between 0.0 and 1.0).
func SetMasterVolume(v float64) {
	audioMu.Lock()
	defer audioMu.Unlock()
	masterVolume = clamp(v, 0.0, 1.0)
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

// SetSoundEffectsEnabled enables or disables all game sound effects.
func SetSoundEffectsEnabled(enabled bool) {
	audioMu.Lock()
	defer audioMu.Unlock()
	isMuted = !enabled
	updateLoopingVolumesLocked()
}

// IsSoundEffectsEnabled returns true if game sound effects are currently active.
func IsSoundEffectsEnabled() bool {
	audioMu.Lock()
	defer audioMu.Unlock()
	return !isMuted
}

// AdjustVolume increases or decreases global volume by delta and returns the new level.
func AdjustVolume(delta float64) float64 {
	audioMu.Lock()
	defer audioMu.Unlock()
	masterVolume = clamp(masterVolume+delta, 0.0, 1.0)
	updateLoopingVolumesLocked()
	return masterVolume
}

func effectiveVolumeLocked(base float64) float64 {
	if isMuted {
		return 0.0
	}
	return base * masterVolume
}

func updateLoopingVolumesLocked() {
	if bomberPlayer != nil {
		bomberPlayer.SetVolume(effectiveVolumeLocked(0.16))
	}
	if satPlayer != nil {
		satPlayer.SetVolume(effectiveVolumeLocked(0.14))
	}
	if smartBombPlayer != nil {
		smartBombPlayer.SetVolume(effectiveVolumeLocked(0.18))
	}
}

// InitAudio initializes the audio context and synthesizes all POKEY waveforms.
func InitAudio() {
	audioOnce.Do(func() {
		defer func() { _ = recover() }()

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
	effectiveVol := effectiveVolumeLocked(vol)
	audioMu.Unlock()

	if audioContext == nil || buf == nil || effectiveVol <= 0.0 {
		return
	}
	defer func() { _ = recover() }()
	if p := audioContext.NewPlayerFromBytes(buf); p != nil {
		p.SetVolume(effectiveVol)
		p.Play()
	}
}

// --- Event Sound Effects ---

func PlayLaunchSound()    { playSoundSafe(launchPCM, 0.20) }
func PlayExplosionSound() { playSoundSafe(explosionPCM, 0.45) }
func PlayTallySound()     { playSoundSafe(tallyPCM, 0.18) }
func PlaySiloLowSound()   { playSoundSafe(siloLowPCM, 0.25) }
func PlayCantFireSound()  { playSoundSafe(cantFirePCM, 0.22) }
func PlayBonusCitySound() { playSoundSafe(bonusCityPCM, 0.28) }
func PlayGameOverSound()  { playSoundSafe(gameOverPCM, 0.45) }
func PlaySirenAlert()     { playSoundSafe(sirenPCM, 0.20) }

// --- Continuous Sound Effects ---

func setLoopingSound(player **audio.Player, pcm []byte, baseVol float64, active bool) {
	audioMu.Lock()
	defer audioMu.Unlock()
	if audioContext == nil {
		return
	}
	if active {
		if *player == nil && len(pcm) > 0 {
			loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))
			if p, err := audioContext.NewPlayer(loop); err == nil {
				p.SetVolume(effectiveVolumeLocked(baseVol))
				*player = p
				p.Play()
			}
		}
	} else if *player != nil {
		_ = (*player).Close()
		*player = nil
	}
}

func SetBomberSound(active bool)    { setLoopingSound(&bomberPlayer, bomberPCM, 0.16, active) }
func SetSatelliteSound(active bool) { setLoopingSound(&satPlayer, satPCM, 0.14, active) }
func SetSmartBombSound(active bool) { setLoopingSound(&smartBombPlayer, smartBombPCM, 0.18, active) }

// StopAllContinuousSounds silences all looping sounds.
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
	return &PokeyLFSR{poly4: 0x0F, poly5: 0x1F, poly17: 0x1FFFF}
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

func advancePhase(phase *float64, freq float64) float64 {
	*phase += 2.0 * math.Pi * freq / SampleRate
	if *phase > 2.0*math.Pi {
		*phase -= 2.0 * math.Pi
	}
	if math.Sin(*phase) >= 0 {
		return 1.0
	}
	return -1.0
}

// synthLaunch creates the authentic ABM missile sweep down using POKEY linear AUDF register progression ($20 -> $E0).
func synthLaunch() []byte {
	duration := 0.28
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		progress := float64(i) / float64(numSamples)
		freq := pokeyFreq(int(32.0 + progress*192.0))
		sq := advancePhase(&phase, freq)
		vol4Bit := math.Max(0, (1.0-progress*0.8)*15.0)
		write16BitLE(buf, i*2, int16(sq)*int16(vol4Bit)*550)
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
		progress := float64(i) / float64(numSamples)

		noiseClock := 240.0*math.Exp(-1.5*progress) + 55.0
		sampleHold := max(1, int(SampleRate/noiseClock))
		if i%sampleHold == 0 {
			noiseVal = lfsr.stepPoly17()*0.80 + lfsr.stepPoly5()*0.20
		}

		subFreq := 68.0 * (1.0 - progress*0.55)
		subPhase += 2.0 * math.Pi * subFreq / SampleRate
		if subPhase > 2.0*math.Pi {
			subPhase -= 2.0 * math.Pi
		}
		subBass := math.Sin(subPhase)

		var vol4Bit float64
		switch {
		case progress < 0.12:
			vol4Bit = 15.0
		case progress < 0.45:
			vol4Bit = 15.0 - (progress-0.12)*8.0
		default:
			vol4Bit = 12.0 * math.Pow(1.0-(progress-0.45)/0.55, 1.3)
		}
		vol4Bit = math.Max(0, vol4Bit)

		mixed := noiseVal*0.65 + subBass*0.35
		write16BitLE(buf, i*2, int16(mixed*(vol4Bit/15.0)*16500.0))
	}
	return buf
}

// synthTally creates the crisp 1520Hz square wave tally blip for bonus point counting (32ms).
func synthTally() []byte {
	duration := 0.032
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	freq := pokeyFreq(20)
	phase := 0.0
	for i := 0; i < numSamples; i++ {
		progress := float64(i) / float64(numSamples)
		sq := advancePhase(&phase, freq)
		vol4Bit := int16((1.0 - progress*0.3) * 14.0)
		write16BitLE(buf, i*2, int16(sq)*vol4Bit*500)
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

		audf := 61
		if int(t*18.0)%2 == 0 {
			audf = 40
		}
		sq := advancePhase(&phase, pokeyFreq(audf))

		vol := 12.0
		if progress > 0.88 {
			vol = 12.0 * (1.0 - (progress-0.88)/0.12)
		}
		write16BitLE(buf, i*2, int16(sq*vol*500))
	}
	return buf
}

// synthSiloLow creates the fast dual-tone alarm blip (AUDF=$1E ~1032Hz, AUDF=$3C ~516Hz).
func synthSiloLow() []byte {
	duration := 0.08
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		audf := 60
		if i < numSamples/2 {
			audf = 30
		}
		sq := advancePhase(&phase, pokeyFreq(audf))
		write16BitLE(buf, i*2, int16(sq)*14*550)
	}
	return buf
}

// synthCantFire generates the low 150Hz 5-bit/4-bit distortion buzzer for empty/ruined base clicks.
func synthCantFire() []byte {
	duration := 0.11
	numSamples := int(SampleRate * duration)
	buf := make([]byte, numSamples*2)

	lfsr := newPokeyLFSR()
	sampleHold := max(1, int(SampleRate/pokeyFreq(100)))
	phase := 0.0
	freq := pokeyFreq(140)
	noiseVal := 0.0

	for i := 0; i < numSamples; i++ {
		progress := float64(i) / float64(numSamples)
		if i%sampleHold == 0 {
			noiseVal = lfsr.stepPoly5() * lfsr.stepPoly4()
		}
		sq := advancePhase(&phase, freq)
		vol4Bit := (1.0 - progress*0.6) * 14.0
		write16BitLE(buf, i*2, int16((sq*0.6+noiseVal*0.4)*vol4Bit*600.0))
	}
	return buf
}

// synthBonusCity generates the iconic 6-tone ascending arcade fanfare arpeggio.
func synthBonusCity() []byte {
	tones := []int{45, 35, 29, 23, 22, 17} // F5, A5, C6, E6, F6, A6
	samplesPerTone := int(SampleRate * 0.08)
	buf := make([]byte, samplesPerTone*len(tones)*2)

	phase := 0.0
	for toneIdx, audf := range tones {
		freq := pokeyFreq(audf)
		startSample := toneIdx * samplesPerTone
		for s := 0; s < samplesPerTone; s++ {
			progress := float64(s) / float64(samplesPerTone)
			sq := advancePhase(&phase, freq)
			vol4Bit := (1.0 - progress*0.4) * 14.0
			write16BitLE(buf, (startSample+s)*2, int16(sq*vol4Bit*500))
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
		progress := float64(i) / float64(numSamples)
		clockRate := 64000.0 / (128.0 + progress*127.0 + 1.0)
		sampleHold := max(1, int(SampleRate/clockRate))
		if i%sampleHold == 0 {
			noiseVal = lfsr.stepPoly17()
		}
		envelope := math.Exp(-1.4 * progress)
		write16BitLE(buf, i*2, int16(noiseVal*envelope*15000.0))
	}
	return buf
}

// synthBomber creates a seamless 0.4s looping low engine motor drone (58Hz POKEY poly4 distortion).
func synthBomber() []byte {
	numSamples := int(SampleRate * 0.40)
	buf := make([]byte, numSamples*2)

	lfsr := newPokeyLFSR()
	sampleHold := SampleRate / 220
	phase := 0.0
	noiseVal := 0.0

	for i := 0; i < numSamples; i++ {
		if i%sampleHold == 0 {
			noiseVal = lfsr.stepPoly4()
		}
		sq := advancePhase(&phase, 58.0)
		write16BitLE(buf, i*2, int16((sq*0.65+noiseVal*0.35)*4500.0))
	}
	return buf
}

// synthSatellite creates a seamless 0.4s looping high-orbit radar telemetry chirp (1684Hz pulsed beacon).
func synthSatellite() []byte {
	numSamples := int(SampleRate * 0.40)
	buf := make([]byte, numSamples*2)

	phase := 0.0
	freq := pokeyFreq(18)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		sq := advancePhase(&phase, freq)
		var sampleVal int16
		if (t < 0.05) || (t >= 0.10 && t < 0.15) {
			sampleVal = int16(sq) * 5000
		}
		write16BitLE(buf, i*2, sampleVal)
	}
	return buf
}

// synthSmartBomb creates a seamless 0.4s looping 240Hz rhythmic warble with 5-bit poly buzz.
func synthSmartBomb() []byte {
	numSamples := int(SampleRate * 0.40)
	buf := make([]byte, numSamples*2)

	lfsr := newPokeyLFSR()
	sampleHold := SampleRate / 480
	phase := 0.0
	noiseVal := 0.0

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		freq := 240.0 + 60.0*math.Sin(2.0*math.Pi*7.5*t)
		if i%sampleHold == 0 {
			noiseVal = lfsr.stepPoly5()
		}
		sq := advancePhase(&phase, freq)
		write16BitLE(buf, i*2, int16((sq*0.7+noiseVal*0.3)*4800.0))
	}
	return buf
}
