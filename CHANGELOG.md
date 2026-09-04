# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2026-09-04

### Added
- **Windows Setup Wizard Installer**:
  - Inno Setup 6 configuration in `scripts/setup-windows.iss` compiling `Missile-Command-Setup-windows-amd64.exe`.
  - Non-elevated user-mode installation support (`PrivilegesRequired=lowest`) enabling standard users to install without UAC prompts, with optional all-users dialog elevation.
  - Start Menu group, uninstaller registration in Windows Settings & Control Panel, and optional desktop shortcut.
  - Local PowerShell build script for Windows developers in `scripts/build-windows-setup.ps1`.
  - Automated compilation integrated into GitHub Actions Windows release matrix.

---

## [1.1.0] - 2026-08-20

### Added
- **Attract Mode Simulated AI Defense**:
  - Live automated AI demo gameplay cycling with the Great Scores screen during attract mode.
  - Automated intercept tracking, predictive firing from optimal silos, and authentic arcade `"DEMO MODE"` overlay.
- **Arcade Controls, Hotkeys & Notifications**:
  - Full Gamepad & Controller support (analog stick aiming, D-Pad, shoulder triggers, and action button firing).
  - Quick hotkeys: `F1` / `Tab` CRT scanline shader toggle, `F11` / `Alt+Enter` fullscreen toggle, `M` sound mute toggle, and `-` / `+` global volume adjustments.
  - Non-intrusive on-screen HUD notification banners for all hotkey settings.
- **WebAssembly (WASM) Web Port**:
  - Pure WebAssembly build support (`GOOS=js GOARCH=wasm`) with automated compilation script in `scripts/build-wasm.sh`.
  - Retro arcade-styled web runner and loader page in `web/index.html` with responsive aspect-ratio scaling and audio auto-resume.
- **Hygiene & High Score Storage Isolation**:
  - User high score persistence in `os.UserConfigDir()` (`~/.config/missile-command/highscores.json`), preventing repository file pollution during local play.
  - Isolated test fixtures for high score storage.
  - Pruned deprecated legacy stroke font.
  - Symmetrical valley city placement with clearance and spacing unit tests.
- **Professional Multi-Platform Application Icon & macOS Packaging**:
  - Apple macOS Human Interface Guidelines (HIG) continuous-curvature squircle grid with brushed titanium bezel, specular rim highlight, and dual-layer ambient/directional elevation shadow.
  - Light & Dark mode contrast optimization for light desktop backgrounds and dark mode Docks.
  - Native Apple Icon Image container (`assets/icon.icns`) with 10 Retina and standard densities.
  - Multi-resolution Windows Icon (`assets/icon.ico`) spanning 16×16 to 256×256.
  - Linux Freedesktop icon suite in `assets/icons/icon-*.png`.
  - Theme-aware WebAssembly web favicons (`web/favicon.ico`, `web/favicon-light.png`, `web/favicon-dark.png`, `web/apple-touch-icon.png`).
  - Runtime desktop window and taskbar/dock icon integration via embedded `assets.GetWindowIcons()` and `ebiten.SetWindowIcon()`.
  - Standalone double-clickable macOS application bundle packaging script in `scripts/bundle-macos.sh`.

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

