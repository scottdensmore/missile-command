use crate::framebuffer::Framebuffer;
use crate::math::{SIM_HEIGHT, SIM_WIDTH};
use crate::palette::Rgba;

pub const CRT_VERTEX_SHADER: &str = r#"#version 100
attribute vec3 position;
attribute vec2 texcoord;
attribute vec4 color0;

varying lowp vec2 uv;
varying lowp vec4 color;

uniform mat4 Model;
uniform mat4 Projection;

void main() {
    gl_Position = Projection * Model * vec4(position, 1);
    color = color0 / 255.0;
    uv = texcoord;
}
"#;

pub const CRT_FRAGMENT_SHADER: &str = r#"#version 100
precision mediump float;

varying lowp vec2 uv;
varying lowp vec4 color;

uniform sampler2D Texture;
uniform vec2 Resolution;
uniform float FlashIntensity;

void main() {
    vec4 orig = texture2D(Texture, uv) * color;
    if (FlashIntensity > 0.0) {
        orig = mix(orig, vec4(1.0, 1.0, 1.0, 1.0), FlashIntensity);
    }

    float scanline = sin(uv.y * Resolution.y * 3.14159) * 0.08;
    vec3 col = orig.rgb - vec3(scanline);
    col = col * 1.05;

    gl_FragColor = vec4(clamp(col, 0.0, 1.0), orig.a);
}
"#;

/// Letterbox viewport geometry for presenting the 256x231 buffer inside a window.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct LetterboxRect {
    pub scale: f64,
    pub offset_x: f64,
    pub offset_y: f64,
    pub render_w: f64,
    pub render_h: f64,
}

/// Computes the aspect-ratio preserving letterbox rectangle for a given window size.
pub fn compute_letterbox(window_w: i32, window_h: i32) -> LetterboxRect {
    let scale_x = window_w as f64 / SIM_WIDTH;
    let scale_y = window_h as f64 / SIM_HEIGHT;
    let scale = scale_x.min(scale_y).max(0.0);
    let render_w = SIM_WIDTH * scale;
    let render_h = SIM_HEIGHT * scale;
    let offset_x = (window_w as f64 - render_w) / 2.0;
    let offset_y = (window_h as f64 - render_h) / 2.0;
    LetterboxRect {
        scale,
        offset_x,
        offset_y,
        render_w,
        render_h,
    }
}

/// Evaluates the CRT scanline, phosphor saturation boost, and flash modulation for a single pixel.
pub fn evaluate_crt_pixel(orig: Rgba, uv_y: f64, use_crt: bool, flash_intensity: f64) -> Rgba {
    let mut r = orig.r as f64 / 255.0;
    let mut g = orig.g as f64 / 255.0;
    let mut b = orig.b as f64 / 255.0;

    let flash = flash_intensity.clamp(0.0, 1.0);
    if flash > 0.0 {
        r = r * (1.0 - flash) + flash;
        g = g * (1.0 - flash) + flash;
        b = b * (1.0 - flash) + flash;
    }

    if use_crt {
        let scanline = (uv_y * SIM_HEIGHT * std::f64::consts::PI).sin() * 0.08;
        r = (r - scanline) * 1.05;
        g = (g - scanline) * 1.05;
        b = (b - scanline) * 1.05;
    }

    Rgba::new(
        (r.clamp(0.0, 1.0) * 255.0).round() as u8,
        (g.clamp(0.0, 1.0) * 255.0).round() as u8,
        (b.clamp(0.0, 1.0) * 255.0).round() as u8,
        orig.a,
    )
}

/// Pipeline manages the 256x231 arcade framebuffer, aspect-ratio scaling, and CRT post-processing.
#[derive(Debug, Clone)]
pub struct Pipeline {
    pub frame_buffer: Framebuffer,
    pub use_crt: bool,
    pub flash_timer: i32,
    pub last_flash_intensity: f32,
}

impl Default for Pipeline {
    fn default() -> Self {
        Self::new()
    }
}

impl Pipeline {
    pub fn new() -> Self {
        Self {
            frame_buffer: Framebuffer::new(SIM_WIDTH as usize, SIM_HEIGHT as usize),
            use_crt: true,
            flash_timer: 0,
            last_flash_intensity: 0.0,
        }
    }

    /// Steps the flash timer for the current frame and returns the active flash intensity (0.0..=2.5).
    pub fn step_frame(&mut self) -> f32 {
        let flash_val = if self.flash_timer > 0 {
            let v = self.flash_timer as f32 / 10.0;
            self.flash_timer -= 1;
            v
        } else {
            0.0
        };
        self.last_flash_intensity = flash_val;
        flash_val
    }

    /// Triggers a brief bright white screen flash (e.g. city destruction or game over).
    pub fn trigger_flash(&mut self, frames: i32) {
        self.flash_timer = frames;
    }
}
