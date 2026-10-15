use crate::framebuffer::Framebuffer;
use crate::palette::Rgba;

/// Bitmap8x8 represents an 8x8 single-bit pixel matrix.
pub type Bitmap8x8 = [u8; 8];

/// Returns the 1980 Atari arcade 8x8 glyph bitmap for a character.
pub const fn arcade_font_glyph(ch: char) -> Option<Bitmap8x8> {
    match ch {
        '0' => Some([0x3C, 0x66, 0x6E, 0x76, 0x66, 0x66, 0x3C, 0x00]),
        '1' => Some([0x18, 0x38, 0x18, 0x18, 0x18, 0x18, 0x7E, 0x00]),
        '2' => Some([0x3C, 0x66, 0x06, 0x0C, 0x18, 0x30, 0x7E, 0x00]),
        '3' => Some([0x3C, 0x66, 0x06, 0x1C, 0x06, 0x66, 0x3C, 0x00]),
        '4' => Some([0x0C, 0x1C, 0x34, 0x64, 0x7E, 0x04, 0x04, 0x00]),
        '5' => Some([0x7E, 0x60, 0x7C, 0x06, 0x06, 0x66, 0x3C, 0x00]),
        '6' => Some([0x1C, 0x30, 0x60, 0x7C, 0x66, 0x66, 0x3C, 0x00]),
        '7' => Some([0x7E, 0x06, 0x0C, 0x18, 0x18, 0x18, 0x18, 0x00]),
        '8' => Some([0x3C, 0x66, 0x66, 0x3C, 0x66, 0x66, 0x3C, 0x00]),
        '9' => Some([0x3C, 0x66, 0x66, 0x3E, 0x06, 0x0C, 0x38, 0x00]),

        'A' => Some([0x18, 0x3C, 0x66, 0x66, 0x7E, 0x66, 0x66, 0x00]),
        'B' => Some([0x7C, 0x66, 0x66, 0x7C, 0x66, 0x66, 0x7C, 0x00]),
        'C' => Some([0x3C, 0x66, 0x60, 0x60, 0x60, 0x66, 0x3C, 0x00]),
        'D' => Some([0x78, 0x6C, 0x66, 0x66, 0x66, 0x6C, 0x78, 0x00]),
        'E' => Some([0x7E, 0x60, 0x60, 0x7C, 0x60, 0x60, 0x7E, 0x00]),
        'F' => Some([0x7E, 0x60, 0x60, 0x7C, 0x60, 0x60, 0x60, 0x00]),
        'G' => Some([0x3C, 0x66, 0x60, 0x6E, 0x66, 0x66, 0x3C, 0x00]),
        'H' => Some([0x66, 0x66, 0x66, 0x7E, 0x66, 0x66, 0x66, 0x00]),
        'I' => Some([0x3C, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00]),
        'J' => Some([0x0E, 0x06, 0x06, 0x06, 0x06, 0x66, 0x3C, 0x00]),
        'K' => Some([0x66, 0x6C, 0x78, 0x70, 0x78, 0x6C, 0x66, 0x00]),
        'L' => Some([0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x7E, 0x00]),
        'M' => Some([0x63, 0x77, 0x7F, 0x6B, 0x63, 0x63, 0x63, 0x00]),
        'N' => Some([0x66, 0x76, 0x7E, 0x7E, 0x6E, 0x66, 0x66, 0x00]),
        'O' => Some([0x3C, 0x66, 0x66, 0x66, 0x66, 0x66, 0x3C, 0x00]),
        'P' => Some([0x7C, 0x66, 0x66, 0x7C, 0x60, 0x60, 0x60, 0x00]),
        'Q' => Some([0x3C, 0x66, 0x66, 0x66, 0x6A, 0x64, 0x3A, 0x00]),
        'R' => Some([0x7C, 0x66, 0x66, 0x7C, 0x78, 0x6C, 0x66, 0x00]),
        'S' => Some([0x3C, 0x66, 0x60, 0x3C, 0x06, 0x66, 0x3C, 0x00]),
        'T' => Some([0x7E, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x00]),
        'U' => Some([0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x3C, 0x00]),
        'V' => Some([0x66, 0x66, 0x66, 0x66, 0x66, 0x3C, 0x18, 0x00]),
        'W' => Some([0x63, 0x63, 0x63, 0x6B, 0x7F, 0x77, 0x63, 0x00]),
        'X' => Some([0x66, 0x66, 0x3C, 0x18, 0x3C, 0x66, 0x66, 0x00]),
        'Y' => Some([0x66, 0x66, 0x66, 0x3C, 0x18, 0x18, 0x18, 0x00]),
        'Z' => Some([0x7E, 0x06, 0x0C, 0x18, 0x30, 0x60, 0x7E, 0x00]),

        ' ' => Some([0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00]),
        ':' => Some([0x00, 0x18, 0x18, 0x00, 0x18, 0x18, 0x00, 0x00]),
        '-' => Some([0x00, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x00, 0x00]),
        '+' => Some([0x00, 0x18, 0x18, 0x7E, 0x18, 0x18, 0x00, 0x00]),
        '.' => Some([0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x00]),
        '!' => Some([0x18, 0x18, 0x18, 0x18, 0x00, 0x18, 0x18, 0x00]),
        '?' => Some([0x3C, 0x66, 0x06, 0x1C, 0x18, 0x00, 0x18, 0x00]),
        '/' => Some([0x06, 0x0C, 0x18, 0x30, 0x60, 0x40, 0x00, 0x00]),
        '*' => Some([0x00, 0x66, 0x3C, 0xFF, 0x3C, 0x66, 0x00, 0x00]),
        '©' => Some([0x3C, 0x42, 0x99, 0xA1, 0x99, 0x42, 0x3C, 0x00]),
        '<' => Some([0x0C, 0x18, 0x30, 0x60, 0x30, 0x18, 0x0C, 0x00]),
        '>' => Some([0x60, 0x30, 0x18, 0x0C, 0x18, 0x30, 0x60, 0x00]),
        '[' => Some([0x3C, 0x30, 0x30, 0x30, 0x30, 0x30, 0x3C, 0x00]),
        ']' => Some([0x3C, 0x0C, 0x0C, 0x0C, 0x0C, 0x0C, 0x3C, 0x00]),
        _ => None,
    }
}

/// Intact 16x8 arcade city silhouette.
pub const CITY_SPRITE: [[u8; 16]; 8] = [
    [0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0],
    [0, 0, 1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 0, 1, 0, 0],
    [1, 1, 1, 1, 0, 0, 1, 1, 0, 1, 1, 1, 0, 1, 1, 0],
    [1, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 1, 0, 1, 1, 0],
    [1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1],
    [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1],
    [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1],
    [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1],
];

/// Rubble/crater when a city is destroyed.
pub const DESTROYED_CITY_SPRITE: [[u8; 16]; 8] = [
    [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    [0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0],
    [0, 1, 1, 0, 1, 0, 1, 1, 0, 1, 0, 1, 1, 0, 1, 0],
    [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1],
    [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1],
];

/// Authentic arcade bomber airplane (16x6).
pub const BOMBER_SPRITE: [[u8; 16]; 6] = [
    [0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0],
    [0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0],
    [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0],
    [0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0],
    [0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0],
    [0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0],
];

/// Authentic arcade satellite (14x8).
pub const SATELLITE_SPRITE: [[u8; 14]; 8] = [
    [0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0],
    [1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1],
    [1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0, 1, 0, 1],
    [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1],
    [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1],
    [1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0, 1, 0, 1],
    [1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1],
    [0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0],
];

/// 2 rotating frames of the smart bomb (8x8 diamond/star).
pub const SMART_BOMB_FRAMES: [[[u8; 8]; 8]; 2] = [
    [
        [0, 0, 0, 1, 1, 0, 0, 0],
        [0, 0, 1, 1, 1, 1, 0, 0],
        [0, 1, 1, 0, 0, 1, 1, 0],
        [1, 1, 0, 1, 1, 0, 1, 1],
        [1, 1, 0, 1, 1, 0, 1, 1],
        [0, 1, 1, 0, 0, 1, 1, 0],
        [0, 0, 1, 1, 1, 1, 0, 0],
        [0, 0, 0, 1, 1, 0, 0, 0],
    ],
    [
        [0, 1, 0, 0, 0, 0, 1, 0],
        [1, 1, 1, 0, 0, 1, 1, 1],
        [0, 1, 1, 1, 1, 1, 1, 0],
        [0, 0, 1, 1, 1, 1, 0, 0],
        [0, 0, 1, 1, 1, 1, 0, 0],
        [0, 1, 1, 1, 1, 1, 1, 0],
        [1, 1, 1, 0, 0, 1, 1, 1],
        [0, 1, 0, 0, 0, 0, 1, 0],
    ],
];

/// Mini city icon drawn in the reserve area (8x5).
pub const BONUS_CITY_ICON: [[u8; 8]; 5] = [
    [0, 1, 0, 0, 1, 0, 0, 0],
    [1, 1, 0, 1, 1, 0, 1, 0],
    [1, 1, 1, 1, 1, 0, 1, 1],
    [1, 1, 1, 1, 1, 1, 1, 1],
    [1, 1, 1, 1, 1, 1, 1, 1],
];

fn draw_grid<const W: usize, const H: usize>(
    target: &mut Framebuffer,
    grid: &[[u8; W]; H],
    x: i32,
    y: i32,
    flip_x: bool,
    col: Rgba,
) {
    for (r, row) in grid.iter().enumerate() {
        for c in 0..W {
            let src_col = if flip_x { W - 1 - c } else { c };
            if row[src_col] != 0 {
                target.set_pixel(x + c as i32, y + r as i32, col);
            }
        }
    }
}

/// Renders string text on the target bitmap using the 8x8 font.
pub fn draw_arcade_text(target: &mut Framebuffer, text: &str, x: i32, y: i32, col: Rgba) {
    for (i, ch) in text.chars().enumerate() {
        let bmp = arcade_font_glyph(ch).unwrap_or_else(|| arcade_font_glyph('?').unwrap());
        draw_bitmap_8x8(target, &bmp, x + (i as i32) * 8, y, col);
    }
}

fn draw_bitmap_8x8(target: &mut Framebuffer, bmp: &Bitmap8x8, x: i32, y: i32, col: Rgba) {
    for (row_idx, &b) in bmp.iter().enumerate() {
        for col_idx in 0..8 {
            if (b & (0x80 >> col_idx)) != 0 {
                target.set_pixel(x + col_idx, y + row_idx as i32, col);
            }
        }
    }
}

/// Renders a city sprite (intact or destroyed) at (x, y).
pub fn draw_city(target: &mut Framebuffer, x: i32, y: i32, destroyed: bool, col: Rgba) {
    if destroyed {
        draw_grid(target, &DESTROYED_CITY_SPRITE, x, y, false, col);
    } else {
        draw_grid(target, &CITY_SPRITE, x, y, false, col);
    }
}

/// Renders the bomber plane at (x, y), flipped horizontally if moving left.
pub fn draw_bomber(target: &mut Framebuffer, x: i32, y: i32, moving_right: bool, col: Rgba) {
    draw_grid(target, &BOMBER_SPRITE, x, y, !moving_right, col);
}

/// Renders the satellite at (x, y).
pub fn draw_satellite(target: &mut Framebuffer, x: i32, y: i32, col: Rgba) {
    draw_grid(target, &SATELLITE_SPRITE, x, y, false, col);
}

/// Renders the animated spinning smart bomb at (x, y).
pub fn draw_smart_bomb(target: &mut Framebuffer, x: i32, y: i32, frame: i32, col: Rgba) {
    let f = (frame.unsigned_abs() as usize) % 2;
    draw_grid(target, &SMART_BOMB_FRAMES[f], x, y, false, col);
}

/// Renders a mini bonus city indicator icon in the lower margin.
pub fn draw_bonus_city_icon(target: &mut Framebuffer, x: i32, y: i32, col: Rgba) {
    draw_grid(target, &BONUS_CITY_ICON, x, y, false, col);
}

/// Draws the 10-missile pyramid (4, 3, 2, 1) for a battery.
pub fn draw_ammo_pyramid(
    target: &mut Framebuffer,
    center_x: i32,
    base_y: i32,
    ammo: i32,
    col: Rgba,
) {
    let positions: [(i32, i32); 10] = [
        (center_x - 6, base_y),
        (center_x - 2, base_y),
        (center_x + 2, base_y),
        (center_x + 6, base_y),
        (center_x - 4, base_y - 4),
        (center_x, base_y - 4),
        (center_x + 4, base_y - 4),
        (center_x - 2, base_y - 8),
        (center_x + 2, base_y - 8),
        (center_x, base_y - 12),
    ];

    let count = ammo.clamp(0, 10) as usize;
    for &(mx, my) in &positions[..count] {
        for ox in 0..2 {
            for oy in 0..3 {
                target.set_pixel(mx + ox, my + oy, col);
            }
        }
    }
}

/// Draws the missile ammo in the silo:
/// When ammo <= 10, it renders the classic 10-missile pyramid.
/// When ammo > 10, it renders the numerical count inside the silo mound.
pub fn draw_silo_ticks(
    target: &mut Framebuffer,
    bx: i32,
    _base_y: i32,
    ammo: i32,
    _max_ammo: i32,
    col: Rgba,
    ground_col: Rgba,
) {
    if ammo <= 0 {
        return;
    }
    if ammo <= 10 {
        draw_ammo_pyramid(target, bx, 220, ammo, col);
        return;
    }

    let ammo_str = ammo.to_string();
    target.draw_filled_rect((bx - 13) as f32, 212.0, 26.0, 12.0, ground_col);
    draw_arcade_text(
        target,
        &ammo_str,
        bx - (ammo_str.len() as i32) * 4,
        214,
        col,
    );
}
