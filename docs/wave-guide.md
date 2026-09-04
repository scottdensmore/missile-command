# Attack Wave Guide & Color Palettes

This document details the attack wave parameters, enemy threat progression, scoring multipliers, and authentic color RAM palettes implemented in [`game/palette.go`](../game/palette.go) and [`game/objects.go`](../game/objects.go).

---

## Wave Progression Table

Enemy ordnance parameters match the original 1980 Atari arcade revision specifications:

| Wave | Sky Color | Ground Color | Multiplier | # ICBMs | Frame Delay | Speed (px/f) | # Smart Bombs | Flier Delay |
| :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **1** | Black | Yellow | **1X** | 12 | 4.8125 | **0.172** | 0 | 320 frames |
| **2** | Black | Yellow | **1X** | 15 | 2.8750 | **0.258** | 0 | 280 frames |
| **3** | Black | Blue | **2X** | 18 | 1.7500 | **0.364** | 0 | 260 frames |
| **4** | Black | Blue | **2X** | 12 | 1.0313 | **0.492** | 0 | 240 frames |
| **5** | Black | Red | **3X** | 16 | 0.6250 | **0.615** | 0 | 220 frames |
| **6** | Black | Red | **3X** | 14 | 0.3750 | **0.727** | 1 | 200 frames |
| **7** | Black | Red | **4X** | 17 | 0.2500 | **0.800** | 1 | 190 frames |
| **8** | Black | Red | **4X** | 10 | 0.1250 | **0.889** | 2 | 180 frames |
| **9** | Dark Blue | Yellow | **5X** | 13 | 0.0625 | **0.941** | 3 | 170 frames |
| **10** | Dark Blue | Yellow | **5X** | 16 | 0.0400 | **0.962** | 4 | 160 frames |
| **11** | Light Blue | Yellow | **6X** | 19 | 0.0200 | **0.980** | 4 | 150 frames |
| **12** | Light Blue | Yellow | **6X** | 12 | 0.0160 | **0.984** | 5 | 140 frames |
| **13** | Purple | Green | **6X** | 14 | 0.0080 | **0.992** | 5 | 130 frames |
| **14** | Purple | Green | **6X** | 16 | 0.0040 | **0.996** | 6 | 120 frames |
| **15** | Yellow | Green | **6X** | 18 | 0.0000 | **1.000** | 6 | 110 frames |
| **16** | Yellow | Green | **6X** | 14 | 0.0000 | **1.000** | 7 | 100 frames |
| **17** | White | Red | **6X** | 17 | 0.0000 | **1.000** | 7 | 100 frames |
| **18** | White | Red | **6X** | 19 | 0.0000 | **1.000** | 7 | 90 frames |
| **19+** | Red | Yellow | **6X** | 22 | 0.0000 | **1.000** | 7 | 80 frames |

*(Waves 21 and beyond repeat the 10 color combinations in 20-wave cycles at maximum difficulty).*

---

## Scoring & Multipliers

Points awarded for destroying incoming hostiles scale directly with the current wave multiplier:

$$\text{Points Earned} = \text{Base Points} \times \text{Wave Multiplier}$$

| Target | Base Points | 1X | 2X | 3X | 4X | 5X | 6X (Max) |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **ICBM Warhead** | 25 | 25 | 50 | 75 | 100 | 125 | 150 |
| **Bomber Aircraft** | 100 | 100 | 200 | 300 | 400 | 500 | 600 |
| **Killer Satellite** | 100 | 100 | 200 | 300 | 400 | 500 | 600 |
| **Smart Bomb** | 125 | 125 | 250 | 375 | 500 | 625 | 750 |
| **Unused ABM** *(End of Wave)* | 5 | 5 | 10 | 15 | 20 | 25 | 30 |
| **Surviving City** *(End of Wave)* | 100 | 100 | 200 | 300 | 400 | 500 | 600 |

---

## Enemy Behavior Details

### 1. ICBMs, Speed & Attack Pacing
- **Speed Formula**: Missile speed is determined by the frame delay $d$: $\text{Speed} = \frac{1}{1 + d}$ pixels/frame. Speeds cap at **1.00 px/frame** on Wave 15+.
- **Altitude-Based Pacing**: In accordance with original arcade ROM logic ($202 - 2 \times \text{wave}$ altitude check), new attacks are blocked until the highest active missile descends past the altitude threshold (top 31 px on Wave 1 to top 51 px on Wave 11+). This creates a steady trickle on early waves and structured volleys on later waves.
- **MIRV Splitters**: Starting in Wave 2, eligible ICBMs split into 2 sister warheads in the upper atmosphere ($Y \in [72, 103]$). All 3 warheads count against the wave missile quota, and splitting only occurs when missile quota remains.

### 2. Smart Bombs (Wave 6+)
- Smart bombs move at the same speed as ICBMs and actively scan for active expanding player explosions within a danger perimeter ($\text{Radius} + 18.0$).
- When an explosion is detected along its trajectory, the smart bomb calculates an evasion vector and maneuvers sideways/around the blast perimeter to survive.
- Capped at a maximum of 3 simultaneous smart bombs on screen (each counting as 2 against the 8-threat active limit).

### 3. Fliers (Bombers & Satellites)
- **Bombers**: Fly horizontally at medium altitudes ($Y \in [65, 90]$) at a constant arcade speed of **1 pixel every 3 frames** ($\approx 0.333$ px/frame), dropping bombs toward player assets.
- **Satellites**: Fly across the upper atmosphere ($Y \in [28, 46]$) at a constant arcade speed of **1 pixel every 2 frames** ($0.500$ px/frame).
- Only one flier may be active at a time with respawn cooldowns decreasing on later waves. Dropped bombs count against the wave quota.

---

## Bonus City Reserve System

- A **Bonus City** is awarded every **10,000 points**.
- Bonus cities are queued in reserve (displayed as mini city icons in the lower margin).
- At the end of a wave, if any of the 6 city positions have been destroyed, one queued bonus city is deployed with an authentic rebuilding animation and musical chime fanfare.

