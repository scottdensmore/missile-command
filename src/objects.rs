use crate::math::Point;

#[derive(Debug, Clone, Copy, Default)]
pub struct Crosshair {
    pub pos: Point,
}

#[derive(Debug, Clone)]
pub struct ICBM {
    pub start: Point,
    pub curr: Point,
    pub target: Point,
    pub speed: f64,
    pub progress: f64,
    pub is_splitter: bool,
    pub splitted: bool,
    pub split_altitude: f64,
    pub active: bool,
}

#[derive(Debug, Clone)]
pub struct SmartBomb {
    pub start: Point,
    pub curr: Point,
    pub target: Point,
    pub speed: f64,
    pub active: bool,
    pub anim_tick: i32,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum FlierType {
    Bomber,
    Satellite,
}

#[derive(Debug, Clone)]
pub struct Flier {
    pub flier_type: FlierType,
    pub x: f64,
    pub y: f64,
    pub speed: f64,
    pub drop_cooldown: i32,
    pub bombs_remaining: i32,
    pub active: bool,
    pub anim_tick: i32,
}

#[derive(Debug, Clone)]
pub struct ABM {
    pub start: Point,
    pub curr: Point,
    pub target: Point,
    pub silo_index: usize,
    pub speed: f64,
    pub progress: f64,
    pub active: bool,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ExplosionState {
    Expanding,
    Holding,
    Contracting,
    Dead,
}

#[derive(Debug, Clone)]
pub struct Explosion {
    pub center: Point,
    pub radius: f64,
    pub max_radius: f64,
    pub hold_timer: i32,
    pub state: ExplosionState,
    pub tick: i32,
}

#[derive(Debug, Clone, Copy, Default)]
pub struct City {
    pub position: Point,
    pub destroyed: bool,
}

#[derive(Debug, Clone, Copy, Default)]
pub struct Battery {
    pub index: usize,
    pub position: Point,
    pub max_ammo: i32,
    pub ammo: i32,
    pub destroyed: bool,
    pub low_warning_played: bool,
}

/// WaveData defines enemy assault parameters for a specific wave.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct WaveData {
    pub total_icbms: i32,
    pub speed: f64,
    pub smart_bombs: i32,
    pub flier_delay: i32,
}

/// ARCADE_WAVE_TABLE matches the authentic 1980 Atari arcade Missile Command progression.
/// Speeds are derived directly from the arcade ROM ICBM frame delay table: speed = 1 / (1 + delay).
/// Wave 15+ reaches 0 delay (updates every frame, 1.00 px/frame at 60 FPS).
pub const ARCADE_WAVE_TABLE: [WaveData; 19] = [
    WaveData {
        total_icbms: 12,
        speed: 0.172,
        smart_bombs: 0,
        flier_delay: 320,
    }, // Wave 1 (delay 4.8125)
    WaveData {
        total_icbms: 15,
        speed: 0.258,
        smart_bombs: 0,
        flier_delay: 280,
    }, // Wave 2 (delay 2.875)
    WaveData {
        total_icbms: 18,
        speed: 0.364,
        smart_bombs: 0,
        flier_delay: 260,
    }, // Wave 3 (delay 1.75)
    WaveData {
        total_icbms: 12,
        speed: 0.492,
        smart_bombs: 0,
        flier_delay: 240,
    }, // Wave 4 (delay 1.031)
    WaveData {
        total_icbms: 16,
        speed: 0.615,
        smart_bombs: 0,
        flier_delay: 220,
    }, // Wave 5 (delay 0.625)
    WaveData {
        total_icbms: 14,
        speed: 0.727,
        smart_bombs: 1,
        flier_delay: 200,
    }, // Wave 6 (delay 0.375)
    WaveData {
        total_icbms: 17,
        speed: 0.800,
        smart_bombs: 1,
        flier_delay: 190,
    }, // Wave 7 (delay 0.25)
    WaveData {
        total_icbms: 10,
        speed: 0.889,
        smart_bombs: 2,
        flier_delay: 180,
    }, // Wave 8 (delay 0.125)
    WaveData {
        total_icbms: 13,
        speed: 0.941,
        smart_bombs: 3,
        flier_delay: 170,
    }, // Wave 9 (delay 0.0625)
    WaveData {
        total_icbms: 16,
        speed: 0.962,
        smart_bombs: 4,
        flier_delay: 160,
    }, // Wave 10 (delay 0.04)
    WaveData {
        total_icbms: 19,
        speed: 0.980,
        smart_bombs: 4,
        flier_delay: 150,
    }, // Wave 11 (delay 0.02)
    WaveData {
        total_icbms: 12,
        speed: 0.984,
        smart_bombs: 5,
        flier_delay: 140,
    }, // Wave 12 (delay 0.016)
    WaveData {
        total_icbms: 14,
        speed: 0.992,
        smart_bombs: 5,
        flier_delay: 130,
    }, // Wave 13 (delay 0.008)
    WaveData {
        total_icbms: 16,
        speed: 0.996,
        smart_bombs: 6,
        flier_delay: 120,
    }, // Wave 14 (delay 0.004)
    WaveData {
        total_icbms: 18,
        speed: 1.000,
        smart_bombs: 6,
        flier_delay: 110,
    }, // Wave 15 (delay 0.0)
    WaveData {
        total_icbms: 14,
        speed: 1.000,
        smart_bombs: 7,
        flier_delay: 100,
    }, // Wave 16
    WaveData {
        total_icbms: 17,
        speed: 1.000,
        smart_bombs: 7,
        flier_delay: 100,
    }, // Wave 17
    WaveData {
        total_icbms: 19,
        speed: 1.000,
        smart_bombs: 7,
        flier_delay: 90,
    }, // Wave 18
    WaveData {
        total_icbms: 22,
        speed: 1.000,
        smart_bombs: 7,
        flier_delay: 80,
    }, // Wave 19+
];

/// Returns the WaveData configuration for any wave index (1-based).
pub fn get_wave_data(wave: i32) -> WaveData {
    let w = wave.max(1);
    let idx = ((w - 1) as usize).min(ARCADE_WAVE_TABLE.len() - 1);
    ARCADE_WAVE_TABLE[idx]
}
