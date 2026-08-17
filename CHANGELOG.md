# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2026-08-17

### Added
- **Atari POKEY 4-Channel Audio Synthesizer**:
  - Full real-time discrete polynomial noise simulation (poly4, poly5, poly17 LFSRs) and square-wave frequency dividers.
  - Authentic ABM launch whistle, deep crunchy explosion rumble, tally chirps, wave-start siren klaxon, silo-low alarm, can't fire buzzer, bonus city fanfare, and apocalyptic game over detonation.
  - Continuous enemy sounds: bomber engine thrum, satellite telemetry beeps, and smart bomb warbles with channel priority management.
- **Arcade Rendering & CRT Shader Pipeline**:
  - Native 256×231 pixel framebuffer with integer scaling, aspect-ratio letterboxing, and optional CRT scanlines/phosphor bloom shader.
  - 10 authentic cycling wave color palettes (cycling every 20 waves).
  - Rapid multi-color palette cycling for expanding, holding, and contracting explosions.
  - Pixel-perfect 8×8 arcade font, intact and destroyed city graphics, 3 base terrain mounds, 10-missile ammo pyramids, `LOW` and `OUT` flashing status indicators, bomber aircraft, satellite, spinning smart bomb diamond star, and bonus city reserve icons.
- **Authentic Gameplay Mechanics & Enemy AI**:
  - **Fast Center Silo (Delta)**: Launches counter-missiles at $7.5\times$ speed for urgent defense.
  - **Smart Bombs**: Introduced in Wave 6+ featuring real-time explosion avoidance steering AI.
  - **Fliers (Bombers & Satellites)**: Cross the sky horizontally dropping targeted ordnance.
  - **MIRV Splitters**: Multi-warhead ICBMs that split at mid-altitudes into multiple targets.
  - **Scoring & Multipliers**: Progressive $1\times$ to $6\times$ scoring multipliers on kills and end-of-wave tallies.
  - **Bonus Cities**: Earned at every 10,000 points, queued in reserve and rebuilding destroyed city slots at wave end.
  - **"THE END" Climax**: Total city annihilation triggers full-screen flashing nuclear fireball, giant typography, and apocalyptic rumble.
  - **High Scores & Attract Mode**: Title screen demo, top 10 arcade leaderboard, interactive 3-letter initials entry, and persistent storage in `highscores.json`.
- **Keyboard & Mouse Controls**:
  - Mouse aiming with crosshair clamping.
  - 3-button firing (Left Mouse / `A`, Middle Mouse / `Space` / `S`, Right Mouse / `D`).
  - Pause (`P`) and Quit (`Q`) controls.
- **CI/CD & GitHub Workflows**:
  - Multi-platform GitHub Actions CI test suite across macOS, Linux, and Windows.
  - Automated GitHub Releases workflow publishing cross-platform precompiled binaries.
