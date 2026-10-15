pub const SIM_WIDTH: f64 = 256.0;
pub const SIM_HEIGHT: f64 = 231.0;

/// Point represents a position in the 2D simulation plane (256x231).
#[derive(Debug, Clone, Copy, PartialEq, Default)]
pub struct Point {
    pub x: f64,
    pub y: f64,
}

impl Point {
    pub const fn new(x: f64, y: f64) -> Self {
        Self { x, y }
    }

    /// Returns the Euclidean distance between two points.
    pub fn distance(self, other: Point) -> f64 {
        (self.x - other.x).hypot(self.y - other.y)
    }

    /// Returns the authentic 1980 Atari 6502 ROM octagonal path step count (`GENVEL`):
    /// `max(|dx|, |dy|) + (3/8) * min(|dx|, |dy|)`, capped at 255.0 steps.
    pub fn arcade_distance(self, other: Point) -> f64 {
        let dx = (self.x - other.x).abs();
        let dy = (self.y - other.y).abs();
        let max_d = dx.max(dy);
        let min_d = dx.min(dy);
        (max_d + 0.375 * min_d).min(255.0)
    }

    /// Performs linear interpolation between two points for t in [0.0, 1.0].
    pub fn lerp(self, other: Point, t: f64) -> Point {
        let t = clamp(t, 0.0, 1.0);
        Point {
            x: self.x + (other.x - self.x) * t,
            y: self.y + (other.y - self.y) * t,
        }
    }
}

/// Converts window pixel coordinates to 256x231 simulation coordinates,
/// accounting for aspect-ratio letterboxing.
pub fn screen_to_sim(screen_x: f64, screen_y: f64, window_w: i32, window_h: i32) -> Point {
    let scale_x = window_w as f64 / SIM_WIDTH;
    let scale_y = window_h as f64 / SIM_HEIGHT;
    let scale = scale_x.min(scale_y);
    if scale <= 0.0 {
        return Point::new(SIM_WIDTH / 2.0, SIM_HEIGHT / 2.0);
    }

    let offset_x = (window_w as f64 - SIM_WIDTH * scale) / 2.0;
    let offset_y = (window_h as f64 - SIM_HEIGHT * scale) / 2.0;

    Point {
        x: (screen_x - offset_x) / scale,
        y: (screen_y - offset_y) / scale,
    }
}

/// Clamps a value between `min_val` and `max_val`.
pub fn clamp(v: f64, min_val: f64, max_val: f64) -> f64 {
    v.max(min_val).min(max_val)
}
