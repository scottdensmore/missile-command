# Atari POKEY Audio Synthesizer

This document details the audio architecture and sound synthesis routines implemented in [`game/audio.go`](../game/audio.go).

---

## Overview

The original 1980 arcade version of *Missile Command* generated its soundscape using an Atari **POKEY** (Potentiometer / Keyboard) integrated circuit (part number **CO-12294**). The chip provides **4 independent audio channels**, each with:
- 8-bit frequency divider registers (`AUDF0`–`AUDF3`).
- 8-bit audio control registers (`AUDC0`–`AUDC3`) selecting distortion modes (polynomial noise generators) and 4-bit output volume.
- Master clock dividers (64 kHz / 15 kHz) and filter options (`AUDCTL`).

This project simulates the POKEY hardware at a standard **44,100 Hz** sample rate in pure Go with zero external audio assets.

---

## Linear Feedback Shift Register (LFSR) Noise Algorithms

The signature gritty, metallic, and rumbling textures of Atari arcade sound effects are produced by three polynomial noise shift registers:

```mermaid
graph LR
    subgraph POKEY_LFSRs [POKEY Shift Registers]
        P4[4-bit LFSR: Period 15<br/>Grungy 8-bit Metallic Tone]
        P5[5-bit LFSR: Period 31<br/>Low Pitched Buzz]
        P17[17-bit LFSR: Period 131,071<br/>Deep Pink/White Noise Rumble]
    end
    P4 --> Mixer
    P5 --> Mixer
    P17 --> Mixer
    Mixer --> AudioOutput[44.1kHz PCM Stream]
```

### 1. 17-Bit Polynomial (Deep Explosion Noise)
- **Polynomial**: $x^{17} + x^{14} + 1$ (Period: 131,071 states)
- Used for deep explosion rumbles, sub-bass detonations, and the apocalyptic "THE END" nuclear climax.

### 2. 5-Bit & 4-Bit Polynomials (Gritty Distortion)
- **5-bit Polynomial**: $x^5 + x^3 + 1$ (Period: 31 states)
- **4-bit Polynomial**: $x^4 + x^3 + 1$ (Period: 15 states)
- Combined to create the engine thrum of bomber aircraft and the buzzy alert warnings.

---

## Event Sound Effects

| Effect | Characteristics & Synthesis Method | Duration |
| :--- | :--- | :---: |
| **ABM Rocket Launch** | Linear POKEY AUDF frequency sweep from `$20` (~970 Hz) down to `$E0` (~142 Hz) with 4-bit quantized DAC decay. | 0.28s |
| **Explosion Rumble** | Heavy sub-bass foundation (68Hz down to 32Hz) layered with 17-bit LFSR polynomial noise down-sweep (240Hz to 55Hz). | 1.05s |
| **Bonus Tally Chirp** | Crisp 1520 Hz (`AUDF=$14`) square wave pulse per remaining missile and surviving city counted. | 32ms |
| **Wave Start Klaxon** | 9 Hz alternating dual-tone siren between 780 Hz (`AUDF=$28`) and 520 Hz (`AUDF=$3D`). | 1.25s |
| **Silo Low Alarm** | Fast dual-tone warning (1032 Hz followed by 516 Hz) when a silo reaches $\le 3$ missiles. | 80ms |
| **Can't Fire Buzzer** | Low 226 Hz square wave combined with 5-bit/4-bit poly distortion when clicking an empty or ruined base. | 110ms |
| **Bonus City Fanfare** | Ascending 6-tone arcade arpeggio sequence ($F_5, A_5, C_6, E_6, F_6, A_6$) for bonus city awards and rebuilding. | 0.50s |
| **"THE END" Nuclear Blast** | Multi-channel deep sub-bass earthquake rumble sweeping down to 25 Hz with high-gain 17-bit poly noise. | 2.80s |

---

## Continuous Ambient Sounds

The original arcade hardware dedicated channels for continuous audio when enemy aircraft or smart weapons entered the playfield:

1. **Bomber Aircraft Drone**: 58 Hz low motor fundamental combined with POKEY 4-bit polynomial distortion.
2. **Satellite Telemetry**: 1684 Hz (`AUDF=$18`) dual-pulse radar beacon every 400ms.
3. **Smart Bomb Warble**: 7.5 Hz frequency-modulated warble oscillating between 180 Hz and 300 Hz with 5-bit poly buzz.

### Channel Priority & Interruption Rules
High-priority audio alerts (such as the *Silo Low* warning or *Can't Fire* buzzer) take precedence over continuous ambient tones, momentarily pausing them and resuming smoothly upon completion. When the game is paused via `P`, all continuous channels are silenced immediately.

