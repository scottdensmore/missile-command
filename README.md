<div align="center">

<img src="assets/banner.png" alt="Missile Command Arcade Cabinet" width="100%" style="border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.6);" />

# MISSILE COMMAND

### *Authentic 1980 Atari Arcade Replica in Go & Ebitengine*

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![CI Status](https://img.shields.io/badge/CI-Passing-2ea44f?style=for-the-badge&logo=githubactions)](https://github.com/scottdensmore/missile-command/actions)
[![Release](https://img.shields.io/badge/Release-v1.0.0-blue?style=for-the-badge&logo=tag)](https://github.com/scottdensmore/missile-command/releases)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS%20|%20Linux%20|%20Windows-lightgrey?style=for-the-badge)]()

<br/>

**Defend six cities from an escalating nuclear barrage.**  
A feature-complete, byte-for-byte authentic recreation of Dave Theurer's legendary 1980 Atari arcade masterpiece.

[Quick Start](#-quick-start) • [Controls](#-controls) • [Documentation](#-documentation) • [Releases](https://github.com/scottdensmore/missile-command/releases)

---

</div>

## 📖 Overview

Originally released in 1980 by Atari, Inc., **Missile Command** is one of the most intense and iconic titles of the Golden Age of Arcade Games. This project is a pure Go implementation powered by [Ebitengine](https://ebitengine.org/), designed to replicate every detail of the original arcade hardware: from the **Atari POKEY 4-channel discrete sound synthesis** to the **256×231 CRT raster display, wave color RAM tables, smart bomb evasion algorithms, and fast center silo mechanics**.

Zero external audio assets or sprite packs are required—every sound waveform and pixel graphic is synthesized and rendered dynamically in real time.

---

## 🚀 Quick Start

### 1. Download Precompiled Binaries
Ready-to-run release binaries for **macOS (Apple Silicon & Intel)**, **Linux (x86_64)**, and **Windows (x64)** are available on the **[Releases Page](https://github.com/scottdensmore/missile-command/releases/tag/v1.0.0)**.

### 2. Run from Source
Prerequisites: [Go 1.25+](https://golang.org/dl/) installed.

```bash
# Clone the repository
git clone https://github.com/scottdensmore/missile-command.git
cd missile-command

# Run directly
go run .

# Or build the binary
go build -o missile-command .
./missile-command
```

> **Linux Users**: Ensure audio and OpenGL libraries are installed:
> ```bash
> sudo apt-get install -y libgl1-mesa-dev xorg-dev libasound2-dev
> ```

---

## 🕹️ Controls

<div align="center">

| Action | Primary Input | Keyboard Alternative |
| :--- | :--- | :--- |
| **Start Game** | `Left Click` on Title Screen | `Space` / `1` / `Enter` |
| **Aim Crosshair** | `Mouse Movement` (Trackball feel) | `Arrow Keys` |
| **Fire Left Base (Alpha)** | `Left Mouse Button` | `A` or `1` |
| **Fire Center Base (Delta - Fast Launcher)** | `Middle Mouse Button` / `Space` | `S` or `2` |
| **Fire Right Base (Omega)** | `Right Mouse Button` | `D` or `3` |
| **Pause / Resume** | **`P`** | **`P`** |
| **Quit Game / Exit** | **`Q`** | **`Q`** |
| **Release Mouse Cursor** | **`Escape`** | **`Escape`** |

</div>

---

## 📚 Documentation

Explore the detailed technical documentation and game guides:

* 🔊 **[Atari POKEY Audio Synthesizer (`docs/audio.md`)](docs/audio.md)**  
  Discrete 4-channel sound synthesis, polynomial noise LFSR algorithms (poly4, poly5, poly17), event sound effects, and continuous ambient tones.

* 📊 **[Attack Wave Guide & Color Palettes (`docs/wave-guide.md`)](docs/wave-guide.md)**  
  Complete wave tables, scoring multipliers ($1\times$ to $6\times$), 10 cycling wave color RAM combinations, smart bomb evasion mechanics, and bonus city replenishment.

* 🎯 **[Controls & Strategic Firing Guide (`docs/controls.md`)](docs/controls.md)**  
  Comprehensive firing mechanics, the strategic value of the fast Delta center silo ($7.5\times$ velocity), and ammunition conservation tactics.

* 🏛️ **[Technical Architecture & Pipeline (`docs/architecture.md`)](docs/architecture.md)**  
  Software design, 256×231 native raster framebuffer pipeline, aspect-ratio letterboxing, Kage CRT scanline shaders, and game state machine.

* 📜 **[Changelog & Release Notes (`CHANGELOG.md`)](CHANGELOG.md)**  
  Full version history and release notes.

---

## 🧪 Testing & CI/CD

Run the automated test suite locally:
```bash
go test -v ./...
```

The repository includes multi-platform [GitHub Actions CI](.github/workflows/ci.yml) running on macOS, Ubuntu (with virtual framebuffer `xvfb`), and Windows, alongside automated cross-platform binary releases.

---

## 📜 Historical Note & License

*Missile Command* was created by **Dave Theurer** at Atari, Inc. in 1980. This project is dedicated to preserving the craftsmanship, audio engineering, and gameplay dynamics of the original arcade cabinet.

This project is open source and available under the [MIT License](LICENSE).
Missile Command is a registered trademark of Atari Interactive, Inc. This fan recreation is developed for educational and historical preservation purposes.
