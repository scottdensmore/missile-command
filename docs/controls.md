# Game Controls & Tactics Guide

This document outlines the controls, input handling options, and strategic gameplay tips for *Missile Command*.

---

## Control Layout

<div align="center">

| Action | Primary Input | Keyboard Equivalent | Gamepad |
| :--- | :--- | :--- | :--- |
| **Start Game** | `Left Click` on Title Screen | `Space` / `1` / `Enter` | `Start` / `A` / `Cross` |
| **Aim Crosshair** | `Mouse Movement` | `Arrow Keys` (Up/Down/Left/Right) | `Left Stick` / `D-Pad` |
| **Fire Left Base (Alpha)** | `Left Mouse Button` | `A` or `1` | `LB` / `LT` / `A` (Bottom action) |
| **Fire Center Base (Delta - Fast)** | `Middle Mouse Button` / `Space` | `S` or `2` | `X` / `Y` (Left/Top action) |
| **Fire Right Base (Omega)** | `Right Mouse Button` | `D` or `3` | `RB` / `RT` / `B` (Right action) |
| **Toggle CRT Scanlines** | **`F1` / `Tab`** | **`F1` / `Tab`** | — |
| **Toggle Fullscreen** | **`F11` / `Alt+Enter`** | **`F11` / `Alt+Enter`** | — |
| **Mute / Unmute Audio** | **`M`** | **`M`** | — |
| **Adjust Audio Volume** | **`-` / `+`** | **`-` / `+`** (or `Numpad -/+`) | — |
| **Pause / Resume** | **`P`** | **`P`** | `Start` / `Options` |
| **Quit Game / Exit** | **`Q`** | **`Q`** | — |
| **Release Mouse Cursor** | **`Escape`** | **`Escape`** | — |

</div>

---

## Firing Mechanics & Silo Characteristics

The arcade game features three missile bases (Alpha, Delta, Omega) with distinct properties:

```
[Base 1: Alpha (Left)]          [Base 2: Delta (Center)]          [Base 3: Omega (Right)]
    10 Missiles                     10 Missiles                       10 Missiles
   Speed: 3.2 units/frame          Speed: 7.5 units/frame (FAST!)    Speed: 3.2 units/frame
```

### 1. The Critical Role of the Center Base (Delta)
- Counter-missiles fired from the center base travel at **7.5 units/frame** ($>2.3\times$ faster than side bases).
- **Strategy**: Reserve center base missiles for urgent emergency intercepts (e.g. low-altitude missiles about to hit a city or evasive smart bombs).

### 2. Side Bases (Alpha & Omega)
- Side base missiles travel at **3.2 units/frame**.
- **Strategy**: Use side bases to lead high-altitude incoming ICBMs early in their descent trajectory.

### 3. Ammo Management & Indicators
- Each base holds a maximum of **10 missiles** arranged in a pyramid (4 bottom, 3, 2, 1 top).
- When a base has **$\le 3$ missiles** remaining:
  - An urgent dual-tone alarm blip plays.
  - `LOW` flashes in yellow above the base.
- When a base is empty:
  - `OUT` flashes in red above the base.
  - Clicking an empty base produces a low "Can't Fire" buzzer sound.
- If an enemy ICBM hits a base, it loses all remaining ammunition for the rest of that wave, but is automatically rebuilt and reloaded at the start of the next wave if any cities survive.

---

## Strategy & Tips

1. **Aim Ahead (Leading Targets)**:
   Explosions take time to expand to their full radius. Fire ahead of descending ICBMs rather than directly at them.
2. **Chain Reactions**:
   A single well-placed explosion can detonate multiple incoming warheads in a cluster. Each destroyed warhead triggers a secondary explosion.
3. **Smart Bomb Interception**:
   Smart bombs dodge expanding blast perimeters. Intercept them by timing detonations right on top of them or trapping them between two simultaneous explosions.
4. **End-of-Wave Bonus Points**:
   Unused missiles award **$5 \times \text{Multiplier}$** and saved cities award **$100 \times \text{Multiplier}$**. Conserving ammo on early waves builds high scores quickly.

