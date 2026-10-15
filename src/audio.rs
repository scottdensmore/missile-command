use crate::math::clamp;
use std::f64::consts::PI;
use std::sync::{Mutex, OnceLock};

pub const SAMPLE_RATE: usize = 44100;

/// Discrete sound effect identifiers triggered by the game engine.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum SoundEffect {
    Launch,
    Explosion,
    Tally,
    Siren,
    SiloLow,
    CantFire,
    BonusCity,
    GameOver,
}

/// Pre-synthesized 16-bit mono PCM buffers for all 11 Atari POKEY sounds.
#[derive(Debug, Clone)]
pub struct SynthesizedAudioBank {
    pub launch_pcm: Vec<u8>,
    pub explosion_pcm: Vec<u8>,
    pub tally_pcm: Vec<u8>,
    pub siren_pcm: Vec<u8>,
    pub silo_low_pcm: Vec<u8>,
    pub cant_fire_pcm: Vec<u8>,
    pub bonus_city_pcm: Vec<u8>,
    pub game_over_pcm: Vec<u8>,
    pub bomber_pcm: Vec<u8>,
    pub sat_pcm: Vec<u8>,
    pub smart_bomb_pcm: Vec<u8>,
}

#[derive(Debug)]
struct AudioState {
    master_volume: f64,
    is_muted: bool,
    bomber_active: bool,
    satellite_active: bool,
    smart_bomb_active: bool,
    pending_events: Vec<(SoundEffect, f32)>,
}

impl Default for AudioState {
    fn default() -> Self {
        Self {
            master_volume: 1.0,
            is_muted: false,
            bomber_active: false,
            satellite_active: false,
            smart_bomb_active: false,
            pending_events: Vec::new(),
        }
    }
}

static AUDIO_BANK: OnceLock<SynthesizedAudioBank> = OnceLock::new();
static AUDIO_STATE: OnceLock<Mutex<AudioState>> = OnceLock::new();

fn state_lock() -> std::sync::MutexGuard<'static, AudioState> {
    AUDIO_STATE
        .get_or_init(|| Mutex::new(AudioState::default()))
        .lock()
        .unwrap_or_else(|e| e.into_inner())
}

/// Initializes and returns the synthesized POKEY audio bank.
pub fn init_audio() -> &'static SynthesizedAudioBank {
    let _ = AUDIO_STATE.get_or_init(|| Mutex::new(AudioState::default()));
    AUDIO_BANK.get_or_init(|| SynthesizedAudioBank {
        launch_pcm: synth_launch(),
        explosion_pcm: synth_explosion(),
        tally_pcm: synth_tally(),
        siren_pcm: synth_siren(),
        silo_low_pcm: synth_silo_low(),
        cant_fire_pcm: synth_cant_fire(),
        bonus_city_pcm: synth_bonus_city(),
        game_over_pcm: synth_game_over(),
        bomber_pcm: synth_bomber(),
        sat_pcm: synth_satellite(),
        smart_bomb_pcm: synth_smart_bomb(),
    })
}

/// Sets the global audio output volume (clamped between 0.0 and 1.0).
pub fn set_master_volume(v: f64) {
    let mut st = state_lock();
    st.master_volume = clamp(v, 0.0, 1.0);
}

/// Returns the current global audio output volume.
pub fn get_master_volume() -> f64 {
    state_lock().master_volume
}

/// Increases or decreases global volume by delta and returns the new level.
pub fn adjust_volume(delta: f64) -> f64 {
    let mut st = state_lock();
    st.master_volume = clamp(st.master_volume + delta, 0.0, 1.0);
    st.master_volume
}

/// Toggles sound output mute on or off and returns the new muted state.
pub fn toggle_mute() -> bool {
    let mut st = state_lock();
    st.is_muted = !st.is_muted;
    st.is_muted
}

/// Returns true if sound output is currently muted.
pub fn is_muted() -> bool {
    state_lock().is_muted
}

/// Enables or disables all game sound effects.
pub fn set_sound_effects_enabled(enabled: bool) {
    let mut st = state_lock();
    st.is_muted = !enabled;
}

/// Returns true if game sound effects are currently active.
pub fn is_sound_effects_enabled() -> bool {
    !state_lock().is_muted
}

fn queue_sound(effect: SoundEffect, base_vol: f64) {
    let mut st = state_lock();
    if st.is_muted || st.master_volume <= 0.0 {
        return;
    }
    let effective = (base_vol * st.master_volume) as f32;
    if effective > 0.0 {
        st.pending_events.push((effect, effective));
    }
}

pub fn play_launch_sound() {
    queue_sound(SoundEffect::Launch, 0.20);
}
pub fn play_explosion_sound() {
    queue_sound(SoundEffect::Explosion, 0.45);
}
pub fn play_tally_sound() {
    queue_sound(SoundEffect::Tally, 0.18);
}
pub fn play_silo_low_sound() {
    queue_sound(SoundEffect::SiloLow, 0.25);
}
pub fn play_cant_fire_sound() {
    queue_sound(SoundEffect::CantFire, 0.22);
}
pub fn play_bonus_city_sound() {
    queue_sound(SoundEffect::BonusCity, 0.28);
}
pub fn play_game_over_sound() {
    queue_sound(SoundEffect::GameOver, 0.45);
}
pub fn play_siren_alert() {
    queue_sound(SoundEffect::Siren, 0.20);
}

pub fn set_bomber_sound(active: bool) {
    state_lock().bomber_active = active;
}

pub fn set_satellite_sound(active: bool) {
    state_lock().satellite_active = active;
}

pub fn set_smart_bomb_sound(active: bool) {
    state_lock().smart_bomb_active = active;
}

pub fn stop_all_continuous_sounds() {
    let mut st = state_lock();
    st.bomber_active = false;
    st.satellite_active = false;
    st.smart_bomb_active = false;
}

/// Snapshot of active looping channels and effective volume scale for the audio backend.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct LoopingSoundStatus {
    pub bomber_active: bool,
    pub satellite_active: bool,
    pub smart_bomb_active: bool,
    pub volume_scale: f32,
}

/// Drains queued one-shot sound events and returns current looping channel status.
pub fn drain_audio_events() -> (Vec<(SoundEffect, f32)>, LoopingSoundStatus) {
    let mut st = state_lock();
    let events = std::mem::take(&mut st.pending_events);
    let volume_scale = if st.is_muted {
        0.0
    } else {
        st.master_volume as f32
    };
    let status = LoopingSoundStatus {
        bomber_active: st.bomber_active,
        satellite_active: st.satellite_active,
        smart_bomb_active: st.smart_bomb_active,
        volume_scale,
    };
    (events, status)
}

/// Wraps raw 16-bit little-endian mono PCM at 44,100 Hz in a standard 44-byte RIFF/WAVE header.
pub fn pcm_to_wav(pcm: &[u8]) -> Vec<u8> {
    let data_len = pcm.len() as u32;
    let file_size = 36 + data_len;
    let sample_rate = SAMPLE_RATE as u32;
    let byte_rate = sample_rate * 2; // 1 channel * 16-bit (2 bytes)

    let mut wav = Vec::with_capacity(44 + pcm.len());
    wav.extend_from_slice(b"RIFF");
    wav.extend_from_slice(&file_size.to_le_bytes());
    wav.extend_from_slice(b"WAVE");
    wav.extend_from_slice(b"fmt ");
    wav.extend_from_slice(&16u32.to_le_bytes()); // Subchunk1Size (16 for PCM)
    wav.extend_from_slice(&1u16.to_le_bytes()); // AudioFormat (1 = PCM)
    wav.extend_from_slice(&1u16.to_le_bytes()); // NumChannels (1 = mono)
    wav.extend_from_slice(&sample_rate.to_le_bytes());
    wav.extend_from_slice(&byte_rate.to_le_bytes());
    wav.extend_from_slice(&2u16.to_le_bytes()); // BlockAlign
    wav.extend_from_slice(&16u16.to_le_bytes()); // BitsPerSample
    wav.extend_from_slice(b"data");
    wav.extend_from_slice(&data_len.to_le_bytes());
    wav.extend_from_slice(pcm);
    wav
}

// --- POKEY HARDWARE EMULATION & SYNTHESIS ---

/// Emulates the exact Atari POKEY polynomial shift registers (poly4, poly5, poly17).
#[derive(Debug, Clone)]
pub struct PokeyLFSR {
    poly4: u8,
    poly5: u8,
    poly17: u32,
}

impl Default for PokeyLFSR {
    fn default() -> Self {
        Self::new()
    }
}

impl PokeyLFSR {
    pub const fn new() -> Self {
        Self {
            poly4: 0x0F,
            poly5: 0x1F,
            poly17: 0x1FFFF,
        }
    }

    pub fn step_poly17(&mut self) -> f64 {
        let bit = ((self.poly17 >> 16) ^ (self.poly17 >> 13)) & 1;
        self.poly17 = ((self.poly17 << 1) | bit) & 0x1FFFF;
        if (self.poly17 & 1) != 0 {
            1.0
        } else {
            -1.0
        }
    }

    pub fn step_poly5(&mut self) -> f64 {
        let b5 = ((self.poly5 >> 4) ^ (self.poly5 >> 2)) & 1;
        self.poly5 = ((self.poly5 << 1) | b5) & 0x1F;
        if (self.poly5 & 1) != 0 {
            1.0
        } else {
            -1.0
        }
    }

    pub fn step_poly4(&mut self) -> f64 {
        let b4 = ((self.poly4 >> 3) ^ (self.poly4 >> 2)) & 1;
        self.poly4 = ((self.poly4 << 1) | b4) & 0x0F;
        if (self.poly4 & 1) != 0 {
            1.0
        } else {
            -1.0
        }
    }
}

#[inline]
fn write_16bit_le(buf: &mut [u8], idx: usize, val: i16) {
    let bytes = val.to_le_bytes();
    buf[idx] = bytes[0];
    buf[idx + 1] = bytes[1];
}

/// Converts POKEY 64kHz frequency divider register AUDF to frequency in Hz.
pub fn pokey_freq(audf: i32) -> f64 {
    let a = audf.max(0);
    64000.0 / (2.0 * (a + 1) as f64)
}

fn advance_phase(phase: &mut f64, freq: f64) -> f64 {
    *phase += 2.0 * PI * freq / (SAMPLE_RATE as f64);
    if *phase > 2.0 * PI {
        *phase -= 2.0 * PI;
    }
    if phase.sin() >= 0.0 {
        1.0
    } else {
        -1.0
    }
}

/// Creates the authentic ABM missile sweep down using POKEY linear AUDF register progression ($20 -> $E0).
pub fn synth_launch() -> Vec<u8> {
    let duration = 0.28;
    let num_samples = (SAMPLE_RATE as f64 * duration) as usize;
    let mut buf = vec![0u8; num_samples * 2];

    let mut phase = 0.0;
    for i in 0..num_samples {
        let progress = i as f64 / num_samples as f64;
        let freq = pokey_freq((32.0 + progress * 192.0) as i32);
        let sq = advance_phase(&mut phase, freq);
        let vol_4bit = ((1.0 - progress * 0.8) * 15.0).max(0.0) as i16;
        let sample = (sq as i16) * vol_4bit * 550;
        write_16bit_le(&mut buf, i * 2, sample);
    }
    buf
}

/// Generates the authentic POKEY 17-bit polynomial noise explosion with heavy sub-bass and deep rumble.
pub fn synth_explosion() -> Vec<u8> {
    let duration = 1.05;
    let num_samples = (SAMPLE_RATE as f64 * duration) as usize;
    let mut buf = vec![0u8; num_samples * 2];

    let mut lfsr = PokeyLFSR::new();
    let mut noise_val = 0.0;
    let mut sub_phase = 0.0;

    for i in 0..num_samples {
        let progress = i as f64 / num_samples as f64;

        let noise_clock = 240.0 * (-1.5 * progress).exp() + 55.0;
        let sample_hold = ((SAMPLE_RATE as f64 / noise_clock) as usize).max(1);
        if i % sample_hold == 0 {
            noise_val = lfsr.step_poly17() * 0.80 + lfsr.step_poly5() * 0.20;
        }

        let sub_freq = 68.0 * (1.0 - progress * 0.55);
        sub_phase += 2.0 * PI * sub_freq / (SAMPLE_RATE as f64);
        if sub_phase > 2.0 * PI {
            sub_phase -= 2.0 * PI;
        }
        let sub_bass = sub_phase.sin();

        let vol_4bit = if progress < 0.12 {
            15.0
        } else if progress < 0.45 {
            15.0 - (progress - 0.12) * 8.0
        } else {
            12.0 * (1.0 - (progress - 0.45) / 0.55).powf(1.3)
        }
        .max(0.0);

        let mixed = noise_val * 0.65 + sub_bass * 0.35;
        let sample = (mixed * (vol_4bit / 15.0) * 16500.0) as i16;
        write_16bit_le(&mut buf, i * 2, sample);
    }
    buf
}

/// Creates the crisp 1520Hz square wave tally blip for bonus point counting (32ms).
pub fn synth_tally() -> Vec<u8> {
    let duration = 0.032;
    let num_samples = (SAMPLE_RATE as f64 * duration) as usize;
    let mut buf = vec![0u8; num_samples * 2];

    let freq = pokey_freq(20);
    let mut phase = 0.0;
    for i in 0..num_samples {
        let progress = i as f64 / num_samples as f64;
        let sq = advance_phase(&mut phase, freq);
        let vol_4bit = ((1.0 - progress * 0.3) * 14.0) as i16;
        write_16bit_le(&mut buf, i * 2, (sq as i16) * vol_4bit * 500);
    }
    buf
}

/// Generates the authentic wave-start siren alternating between 780Hz ($28) and 520Hz ($3D) at 9Hz.
pub fn synth_siren() -> Vec<u8> {
    let duration = 1.25;
    let num_samples = (SAMPLE_RATE as f64 * duration) as usize;
    let mut buf = vec![0u8; num_samples * 2];

    let mut phase = 0.0;
    for i in 0..num_samples {
        let t = i as f64 / SAMPLE_RATE as f64;
        let progress = t / duration;

        let audf = if ((t * 18.0) as i32) % 2 == 0 { 40 } else { 61 };
        let sq = advance_phase(&mut phase, pokey_freq(audf));

        let vol = if progress > 0.88 {
            12.0 * (1.0 - (progress - 0.88) / 0.12)
        } else {
            12.0
        };
        write_16bit_le(&mut buf, i * 2, (sq * vol * 500.0) as i16);
    }
    buf
}

/// Creates the fast dual-tone alarm blip (AUDF=$1E ~1032Hz, AUDF=$3C ~516Hz).
pub fn synth_silo_low() -> Vec<u8> {
    let duration = 0.08;
    let num_samples = (SAMPLE_RATE as f64 * duration) as usize;
    let mut buf = vec![0u8; num_samples * 2];

    let mut phase = 0.0;
    for i in 0..num_samples {
        let audf = if i < num_samples / 2 { 30 } else { 60 };
        let sq = advance_phase(&mut phase, pokey_freq(audf));
        write_16bit_le(&mut buf, i * 2, (sq as i16) * 14 * 550);
    }
    buf
}

/// Generates the low 150Hz 5-bit/4-bit distortion buzzer for empty/ruined base clicks.
pub fn synth_cant_fire() -> Vec<u8> {
    let duration = 0.11;
    let num_samples = (SAMPLE_RATE as f64 * duration) as usize;
    let mut buf = vec![0u8; num_samples * 2];

    let mut lfsr = PokeyLFSR::new();
    let sample_hold = ((SAMPLE_RATE as f64 / pokey_freq(100)) as usize).max(1);
    let mut phase = 0.0;
    let freq = pokey_freq(140);
    let mut noise_val = 0.0;

    for i in 0..num_samples {
        let progress = i as f64 / num_samples as f64;
        if i % sample_hold == 0 {
            noise_val = lfsr.step_poly5() * lfsr.step_poly4();
        }
        let sq = advance_phase(&mut phase, freq);
        let vol_4bit = (1.0 - progress * 0.6) * 14.0;
        write_16bit_le(
            &mut buf,
            i * 2,
            ((sq * 0.6 + noise_val * 0.4) * vol_4bit * 600.0) as i16,
        );
    }
    buf
}

/// Generates the iconic 6-tone ascending arcade fanfare arpeggio.
pub fn synth_bonus_city() -> Vec<u8> {
    let tones = [45, 35, 29, 23, 22, 17]; // F5, A5, C6, E6, F6, A6
    let samples_per_tone = (SAMPLE_RATE as f64 * 0.08) as usize;
    let mut buf = vec![0u8; samples_per_tone * tones.len() * 2];

    let mut phase = 0.0;
    for (tone_idx, &audf) in tones.iter().enumerate() {
        let freq = pokey_freq(audf);
        let start_sample = tone_idx * samples_per_tone;
        for s in 0..samples_per_tone {
            let progress = s as f64 / samples_per_tone as f64;
            let sq = advance_phase(&mut phase, freq);
            let vol_4bit = (1.0 - progress * 0.4) * 14.0;
            write_16bit_le(
                &mut buf,
                (start_sample + s) * 2,
                (sq * vol_4bit * 500.0) as i16,
            );
        }
    }
    buf
}

/// Creates the massive apocalyptic 2.8s nuclear detonation rumble for "THE END".
pub fn synth_game_over() -> Vec<u8> {
    let duration = 2.8;
    let num_samples = (SAMPLE_RATE as f64 * duration) as usize;
    let mut buf = vec![0u8; num_samples * 2];

    let mut lfsr = PokeyLFSR::new();
    let mut noise_val = 0.0;
    for i in 0..num_samples {
        let progress = i as f64 / num_samples as f64;
        let clock_rate = 64000.0 / (128.0 + progress * 127.0 + 1.0);
        let sample_hold = ((SAMPLE_RATE as f64 / clock_rate) as usize).max(1);
        if i % sample_hold == 0 {
            noise_val = lfsr.step_poly17();
        }
        let envelope = (-1.4 * progress).exp();
        write_16bit_le(&mut buf, i * 2, (noise_val * envelope * 15000.0) as i16);
    }
    buf
}

/// Creates a seamless 0.4s looping low engine motor drone (58Hz POKEY poly4 distortion).
pub fn synth_bomber() -> Vec<u8> {
    let num_samples = (SAMPLE_RATE as f64 * 0.40) as usize;
    let mut buf = vec![0u8; num_samples * 2];

    let mut lfsr = PokeyLFSR::new();
    let sample_hold = SAMPLE_RATE / 220;
    let mut phase = 0.0;
    let mut noise_val = 0.0;

    for i in 0..num_samples {
        if i % sample_hold == 0 {
            noise_val = lfsr.step_poly4();
        }
        let sq = advance_phase(&mut phase, 58.0);
        write_16bit_le(
            &mut buf,
            i * 2,
            ((sq * 0.65 + noise_val * 0.35) * 4500.0) as i16,
        );
    }
    buf
}

/// Creates a seamless 0.4s looping high-orbit radar telemetry chirp (1684Hz pulsed beacon).
pub fn synth_satellite() -> Vec<u8> {
    let num_samples = (SAMPLE_RATE as f64 * 0.40) as usize;
    let mut buf = vec![0u8; num_samples * 2];

    let mut phase = 0.0;
    let freq = pokey_freq(18);
    for i in 0..num_samples {
        let t = i as f64 / SAMPLE_RATE as f64;
        let sq = advance_phase(&mut phase, freq);
        let sample = if t < 0.05 || (0.10..0.15).contains(&t) {
            (sq as i16) * 5000
        } else {
            0
        };
        write_16bit_le(&mut buf, i * 2, sample);
    }
    buf
}

/// Creates a seamless 0.4s looping 240Hz rhythmic warble with 5-bit poly buzz.
pub fn synth_smart_bomb() -> Vec<u8> {
    let num_samples = (SAMPLE_RATE as f64 * 0.40) as usize;
    let mut buf = vec![0u8; num_samples * 2];

    let mut lfsr = PokeyLFSR::new();
    let sample_hold = SAMPLE_RATE / 480;
    let mut phase = 0.0;
    let mut noise_val = 0.0;

    for i in 0..num_samples {
        let t = i as f64 / SAMPLE_RATE as f64;
        let freq = 240.0 + 60.0 * (2.0 * PI * 7.5 * t).sin();
        if i % sample_hold == 0 {
            noise_val = lfsr.step_poly5();
        }
        let sq = advance_phase(&mut phase, freq);
        write_16bit_le(
            &mut buf,
            i * 2,
            ((sq * 0.7 + noise_val * 0.3) * 4800.0) as i16,
        );
    }
    buf
}
