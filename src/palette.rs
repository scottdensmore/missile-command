/// RGBA color with 8-bit channels.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub struct Rgba {
    pub r: u8,
    pub g: u8,
    pub b: u8,
    pub a: u8,
}

impl Rgba {
    pub const fn new(r: u8, g: u8, b: u8, a: u8) -> Self {
        Self { r, g, b, a }
    }

    pub const fn to_array(self) -> [u8; 4] {
        [self.r, self.g, self.b, self.a]
    }
}

// Shared arcade UI & HUD colors.
pub const COLOR_BLACK: Rgba = Rgba::new(0, 0, 0, 255);
pub const COLOR_WHITE: Rgba = Rgba::new(240, 240, 240, 255);
pub const COLOR_RED: Rgba = Rgba::new(240, 32, 32, 255);
pub const COLOR_YELLOW: Rgba = Rgba::new(240, 220, 32, 255);
pub const COLOR_CYAN: Rgba = Rgba::new(40, 220, 240, 255);
pub const COLOR_GREEN: Rgba = Rgba::new(40, 240, 60, 255);

/// Palette defines the colors used for rendering a specific wave.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Palette {
    pub sky_color: Rgba,
    pub ground_color: Rgba,
    pub city_color: Rgba,
    pub silo_color: Rgba,
    pub icbm_color: Rgba,
    pub abm_color: Rgba,
    pub text_color: Rgba,
    pub multiplier_color: Rgba,
    pub multiplier: i32,
}

/// WAVE_PALETTES contains the 10 authentic Atari 1980 arcade wave color combinations.
/// They cycle every 20 waves (2 waves per palette).
pub const WAVE_PALETTES: [Palette; 10] = [
    // Waves 1-2: 1X (Black Sky, Yellow Ground, Blue Cities/Ammo)
    Palette {
        sky_color: Rgba::new(0, 0, 0, 255),
        ground_color: Rgba::new(228, 196, 36, 255),
        city_color: Rgba::new(50, 120, 245, 255),
        silo_color: Rgba::new(50, 120, 245, 255),
        icbm_color: Rgba::new(240, 32, 32, 255),
        abm_color: Rgba::new(50, 200, 255, 255),
        text_color: Rgba::new(240, 240, 240, 255),
        multiplier_color: Rgba::new(240, 240, 240, 255),
        multiplier: 1,
    },
    // Waves 3-4: 2X (Black Sky, Blue Ground, Cyan Cities/Ammo)
    Palette {
        sky_color: Rgba::new(0, 0, 0, 255),
        ground_color: Rgba::new(36, 100, 240, 255),
        city_color: Rgba::new(36, 224, 180, 255),
        silo_color: Rgba::new(36, 224, 180, 255),
        icbm_color: Rgba::new(240, 220, 32, 255),
        abm_color: Rgba::new(240, 240, 240, 255),
        text_color: Rgba::new(36, 224, 180, 255),
        multiplier_color: Rgba::new(36, 224, 180, 255),
        multiplier: 2,
    },
    // Waves 5-6: 3X (Black Sky, Red Ground, Yellow Cities/Ammo)
    Palette {
        sky_color: Rgba::new(0, 0, 0, 255),
        ground_color: Rgba::new(224, 36, 36, 255),
        city_color: Rgba::new(240, 220, 32, 255),
        silo_color: Rgba::new(240, 220, 32, 255),
        icbm_color: Rgba::new(36, 224, 224, 255),
        abm_color: Rgba::new(240, 220, 32, 255),
        text_color: Rgba::new(240, 220, 32, 255),
        multiplier_color: Rgba::new(240, 220, 32, 255),
        multiplier: 3,
    },
    // Waves 7-8: 4X (Black Sky, Red Ground, Cyan Cities/Ammo)
    Palette {
        sky_color: Rgba::new(0, 0, 0, 255),
        ground_color: Rgba::new(224, 36, 36, 255),
        city_color: Rgba::new(36, 224, 224, 255),
        silo_color: Rgba::new(36, 224, 224, 255),
        icbm_color: Rgba::new(240, 240, 240, 255),
        abm_color: Rgba::new(36, 224, 224, 255),
        text_color: Rgba::new(36, 224, 224, 255),
        multiplier_color: Rgba::new(36, 224, 224, 255),
        multiplier: 4,
    },
    // Waves 9-10: 5X (Dark Blue Sky, Yellow Ground, Red Cities/Ammo)
    Palette {
        sky_color: Rgba::new(16, 24, 140, 255),
        ground_color: Rgba::new(228, 196, 36, 255),
        city_color: Rgba::new(224, 36, 36, 255),
        silo_color: Rgba::new(224, 36, 36, 255),
        icbm_color: Rgba::new(240, 40, 40, 255),
        abm_color: Rgba::new(240, 240, 240, 255),
        text_color: Rgba::new(228, 196, 36, 255),
        multiplier_color: Rgba::new(228, 196, 36, 255),
        multiplier: 5,
    },
    // Waves 11-12: 6X (Light Blue Sky, Yellow Ground, Dark Blue Cities/Ammo)
    Palette {
        sky_color: Rgba::new(60, 120, 224, 255),
        ground_color: Rgba::new(228, 196, 36, 255),
        city_color: Rgba::new(16, 24, 140, 255),
        silo_color: Rgba::new(16, 24, 140, 255),
        icbm_color: Rgba::new(20, 20, 20, 255),
        abm_color: Rgba::new(240, 240, 240, 255),
        text_color: Rgba::new(240, 240, 240, 255),
        multiplier_color: Rgba::new(240, 240, 240, 255),
        multiplier: 6,
    },
    // Waves 13-14: 6X (Purple Sky, Green Ground, Yellow Cities/Ammo)
    Palette {
        sky_color: Rgba::new(140, 24, 140, 255),
        ground_color: Rgba::new(36, 180, 48, 255),
        city_color: Rgba::new(240, 220, 32, 255),
        silo_color: Rgba::new(240, 220, 32, 255),
        icbm_color: Rgba::new(240, 220, 32, 255),
        abm_color: Rgba::new(36, 224, 224, 255),
        text_color: Rgba::new(240, 220, 32, 255),
        multiplier_color: Rgba::new(240, 220, 32, 255),
        multiplier: 6,
    },
    // Waves 15-16: 6X (Yellow Sky, Green Ground, Red Cities/Ammo)
    Palette {
        sky_color: Rgba::new(220, 190, 24, 255),
        ground_color: Rgba::new(36, 180, 48, 255),
        city_color: Rgba::new(224, 36, 36, 255),
        silo_color: Rgba::new(224, 36, 36, 255),
        icbm_color: Rgba::new(36, 36, 200, 255),
        abm_color: Rgba::new(224, 36, 36, 255),
        text_color: Rgba::new(36, 36, 36, 255),
        multiplier_color: Rgba::new(36, 36, 36, 255),
        multiplier: 6,
    },
    // Waves 17-18: 6X (White Sky, Red Ground, Blue Cities/Ammo)
    Palette {
        sky_color: Rgba::new(230, 230, 230, 255),
        ground_color: Rgba::new(224, 36, 36, 255),
        city_color: Rgba::new(24, 48, 200, 255),
        silo_color: Rgba::new(24, 48, 200, 255),
        icbm_color: Rgba::new(20, 20, 20, 255),
        abm_color: Rgba::new(24, 48, 200, 255),
        text_color: Rgba::new(24, 48, 200, 255),
        multiplier_color: Rgba::new(24, 48, 200, 255),
        multiplier: 6,
    },
    // Waves 19-20: 6X (Red Sky, Yellow Ground, Black Cities/Ammo)
    Palette {
        sky_color: Rgba::new(210, 32, 32, 255),
        ground_color: Rgba::new(228, 196, 36, 255),
        city_color: Rgba::new(20, 20, 20, 255),
        silo_color: Rgba::new(20, 20, 20, 255),
        icbm_color: Rgba::new(20, 20, 20, 255),
        abm_color: Rgba::new(228, 196, 36, 255),
        text_color: Rgba::new(228, 196, 36, 255),
        multiplier_color: Rgba::new(228, 196, 36, 255),
        multiplier: 6,
    },
];

/// Returns the authentic arcade palette for any wave index (1-based).
pub fn get_palette_for_wave(wave: i32) -> Palette {
    let w = wave.max(1);
    let idx = (((w - 1) / 2) as usize) % WAVE_PALETTES.len();
    let mut p = WAVE_PALETTES[idx];

    // Scoring multiplier progression: increases by 1 every 2 waves, capped at 6x
    let mult = (1 + (w - 1) / 2).min(6);
    p.multiplier = mult;
    p
}

/// EXPLOSION_COLORS holds the authentic arcade palette-cycling rainbow sequence.
pub const EXPLOSION_COLORS: [Rgba; 8] = [
    Rgba::new(255, 255, 255, 255), // White
    Rgba::new(255, 240, 40, 255),  // Yellow
    Rgba::new(40, 240, 255, 255),  // Cyan
    Rgba::new(255, 60, 220, 255),  // Magenta
    Rgba::new(255, 130, 20, 255),  // Orange
    Rgba::new(50, 240, 60, 255),   // Green
    Rgba::new(240, 40, 40, 255),   // Red
    Rgba::new(80, 120, 255, 255),  // Blue
];

/// Returns a pulsing color from the color-cycling explosion table based on frame tick.
pub fn get_explosion_color(tick: i32) -> Rgba {
    let idx = ((tick.max(0) / 3) as usize) % EXPLOSION_COLORS.len();
    EXPLOSION_COLORS[idx]
}
