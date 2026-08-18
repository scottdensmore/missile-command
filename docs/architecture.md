# Technical Architecture & Pipeline

This document describes the software architecture, rendering pipeline, and state machine of the *Missile Command* replica.

---

## High-Level Architecture

```mermaid
graph TD
    subgraph Engine [Ebitengine Game Loop]
        Input[Player Input: Mouse / Keyboard] --> Update[g.Update]
        Update --> StateMachine{Game State Machine}
        StateMachine -->|StateAttract| Attract[Title & High Scores Demo]
        StateMachine -->|StateWaveStart| WaveStart[Wave Announcement & Siren]
        StateMachine -->|StatePlaying| Sim[Physics, Spawning, Evasion & Collisions]
        StateMachine -->|StateTally| Tally[End-of-Wave Accounting]
        StateMachine -->|StateBonusRebuild| Rebuild[Bonus City Deployment]
        StateMachine -->|StateTheEnd| TheEnd[Nuclear Climax Sequence]
        StateMachine -->|StateHighScoreEntry| Initials[3-Letter Leaderboard Entry]
    end

    subgraph Simulation [Simulation & Audio]
        Sim --> Entities[ICBMs, SmartBombs, Fliers, ABMs, Explosions]
        Sim --> POKEY[4-Channel Real-time Audio Synthesizer]
    end

    subgraph Rendering [Arcade Raster Pipeline]
        Sim --> Framebuffer[256x231 Native Raster Buffer]
        Framebuffer --> Scaler[Integer Aspect-Ratio Letterboxer]
        Scaler --> CRTShader[Kage CRT Scanline & Phosphor Bloom Shader]
        CRTShader --> Screen[Window Presentation]
    end
```

---

## Native 256×231 Raster Framebuffer Pipeline

The original arcade cabinet displayed graphics on a 256×231 cathode ray tube with dynamic color lookup RAM.

### 1. Framebuffer Isolation
All game primitives (lines, filled polygons, 8×8 bitmap text, and pixel sprites) are drawn directly onto an internal `256×231` `*ebiten.Image`.

### 2. Aspect-Ratio Letterboxer (`game/pipeline.go`)
When running on modern displays with arbitrary resolutions:
1. Calculates the maximum integer/float scale factor:
   $$\text{scale} = \min\left(\frac{W_{\text{window}}}{256}, \frac{H_{\text{window}}}{231}\right)$$
2. Centers the 4:3 playfield with letterbox/pillarbox margins.
3. Maps mouse cursor input from window pixels back to the $256 \times 231$ simulation coordinate space using [`ScreenToSimWithLetterbox`](../game/math.go).

### 3. Kage CRT Scanline Shader (`game/shader.go`)
The shader performs:
- Sinusoidal scanline modulation synchronized to the 231-line vertical raster:
  $$\text{scanline} = \sin(y \cdot 231 \cdot \pi) \cdot 0.08$$
- Color saturation boost mimicking aged CRT phosphor bleed.
- Dynamic screen flash intensity for nuclear detonations and city destruction.

---

## State Machine & Transitions

```mermaid
stateDiagram-v2
    [*] --> StateAttract: Application Launch
    StateAttract --> StateWaveStart: Space / Click / 1 Key
    StateWaveStart --> StatePlaying: Siren Finishes (1.5s)
    StatePlaying --> StateTally: All Threats Neutralized
    StatePlaying --> StateTheEnd: All 6 Cities Destroyed
    StateTally --> StateBonusRebuild: Ammo & Cities Counted
    StateBonusRebuild --> StateWaveStart: Next Wave (Wave++)
    StateTheEnd --> StateHighScoreEntry: Score Qualifies for Top 10
    StateTheEnd --> StateAttract: Score Does Not Qualify
    StateHighScoreEntry --> StateAttract: 3 Initials Saved
```

---

## Directory Structure

| File | Purpose |
| :--- | :--- |
| [`game/audio.go`](../game/audio.go) | Real-time POKEY 4-channel sound synthesizer and LFSR noise generators. |
| [`game/game.go`](../game/game.go) | Central game loop, state transitions, entity updates, and HUD rendering. |
| [`game/objects.go`](../game/objects.go) | Data structures for ICBMs, SmartBombs, Fliers, ABMs, Explosions, Cities, and Silos. |
| [`game/palette.go`](../game/palette.go) | 10 wave color RAM combinations and dynamic explosion color-cycling tables. |
| [`game/sprites.go`](../game/sprites.go) | 8×8 arcade bitmap font, city sprites, aircraft, satellites, and ammo pyramids. |
| [`game/pipeline.go`](../game/pipeline.go) | Aspect-ratio letterbox scaler and CRT shader compositing. |
| [`game/scores.go`](../game/scores.go) | Top 10 leaderboard management, JSON persistence, and initials entry. |
| [`game/shader.go`](../game/shader.go) | Kage GLSL-style CRT scanline and phosphor bloom fragment shader. |
| [`game/math.go`](../game/math.go) | Simulation geometry, vector distances, and coordinate transformation math. |

