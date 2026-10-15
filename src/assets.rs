use image::GenericImageView;
use macroquad::miniquad::conf::Icon;

// Embedded multi-resolution PNG icons for runtime desktop window/taskbar/dock icons.
pub const ICON_16_PNG: &[u8] = include_bytes!("../assets/icons/icon-16.png");
pub const ICON_32_PNG: &[u8] = include_bytes!("../assets/icons/icon-32.png");
pub const ICON_48_PNG: &[u8] = include_bytes!("../assets/icons/icon-48.png");
pub const ICON_64_PNG: &[u8] = include_bytes!("../assets/icons/icon-64.png");
pub const ICON_128_PNG: &[u8] = include_bytes!("../assets/icons/icon-128.png");

/// Decoded RGBA icon image.
#[derive(Debug, Clone)]
pub struct DecodedIcon {
    pub width: u32,
    pub height: u32,
    pub rgba: Vec<u8>,
}

/// Decodes and returns all 5 embedded multi-resolution window icons (16, 32, 48, 64, 128).
pub fn get_window_icons() -> Vec<DecodedIcon> {
    let raws = [
        ICON_16_PNG,
        ICON_32_PNG,
        ICON_48_PNG,
        ICON_64_PNG,
        ICON_128_PNG,
    ];
    let mut icons = Vec::with_capacity(raws.len());
    for raw in raws {
        if let Ok(img) = image::load_from_memory(raw) {
            let (width, height) = img.dimensions();
            let rgba = img.to_rgba8().into_raw();
            icons.push(DecodedIcon {
                width,
                height,
                rgba,
            });
        }
    }
    icons
}

/// Converts the embedded 16x16, 32x32, and 64x64 PNGs into a `macroquad` window Icon.
pub fn get_macroquad_icon() -> Option<Icon> {
    let img16 = image::load_from_memory(ICON_16_PNG).ok()?.to_rgba8();
    let img32 = image::load_from_memory(ICON_32_PNG).ok()?.to_rgba8();
    let img64 = image::load_from_memory(ICON_64_PNG).ok()?.to_rgba8();

    let small: [u8; 16 * 16 * 4] = img16.into_raw().try_into().ok()?;
    let medium: [u8; 32 * 32 * 4] = img32.into_raw().try_into().ok()?;
    let big: [u8; 64 * 64 * 4] = img64.into_raw().try_into().ok()?;

    Some(Icon { small, medium, big })
}
