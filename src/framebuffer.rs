use crate::palette::Rgba;

/// Software 2D RGBA framebuffer representing the 256x231 arcade CRT raster.
#[derive(Debug, Clone)]
pub struct Framebuffer {
    pub width: usize,
    pub height: usize,
    pub pixels: Vec<u8>, // width * height * 4 RGBA bytes
}

impl Framebuffer {
    pub fn new(width: usize, height: usize) -> Self {
        Self {
            width,
            height,
            pixels: vec![0; width * height * 4],
        }
    }

    pub fn clear(&mut self) {
        self.pixels.fill(0);
    }

    pub fn fill(&mut self, col: Rgba) {
        for chunk in self.pixels.as_chunks_mut::<4>().0 {
            chunk[0] = col.r;
            chunk[1] = col.g;
            chunk[2] = col.b;
            chunk[3] = col.a;
        }
    }

    #[inline]
    pub fn set_pixel(&mut self, x: i32, y: i32, col: Rgba) {
        if x < 0 || y < 0 || x as usize >= self.width || y as usize >= self.height {
            return;
        }
        let idx = ((y as usize) * self.width + (x as usize)) * 4;
        if col.a == 255 {
            self.pixels[idx] = col.r;
            self.pixels[idx + 1] = col.g;
            self.pixels[idx + 2] = col.b;
            self.pixels[idx + 3] = 255;
        } else if col.a > 0 {
            let alpha = col.a as u16;
            let inv = 255 - alpha;
            self.pixels[idx] = ((col.r as u16 * alpha + self.pixels[idx] as u16 * inv) / 255) as u8;
            self.pixels[idx + 1] =
                ((col.g as u16 * alpha + self.pixels[idx + 1] as u16 * inv) / 255) as u8;
            self.pixels[idx + 2] =
                ((col.b as u16 * alpha + self.pixels[idx + 2] as u16 * inv) / 255) as u8;
            self.pixels[idx + 3] = 255;
        }
    }

    #[inline]
    pub fn get_pixel(&self, x: i32, y: i32) -> Rgba {
        if x < 0 || y < 0 || x as usize >= self.width || y as usize >= self.height {
            return Rgba::default();
        }
        let idx = ((y as usize) * self.width + (x as usize)) * 4;
        Rgba::new(
            self.pixels[idx],
            self.pixels[idx + 1],
            self.pixels[idx + 2],
            self.pixels[idx + 3],
        )
    }

    pub fn draw_filled_rect(&mut self, x: f32, y: f32, w: f32, h: f32, col: Rgba) {
        let x0 = x.round() as i32;
        let y0 = y.round() as i32;
        let x1 = (x + w).round() as i32;
        let y1 = (y + h).round() as i32;
        for py in y0..y1 {
            for px in x0..x1 {
                self.set_pixel(px, py, col);
            }
        }
    }

    pub fn stroke_rect(&mut self, x: f32, y: f32, w: f32, h: f32, col: Rgba) {
        let x0 = x.round() as i32;
        let y0 = y.round() as i32;
        let x1 = (x + w).round() as i32 - 1;
        let y1 = (y + h).round() as i32 - 1;
        for px in x0..=x1 {
            self.set_pixel(px, y0, col);
            self.set_pixel(px, y1, col);
        }
        for py in y0..=y1 {
            self.set_pixel(x0, py, col);
            self.set_pixel(x1, py, col);
        }
    }

    /// Draws a 1px raster line using Bresenham's line algorithm.
    pub fn stroke_line(&mut self, x0: f32, y0: f32, x1: f32, y1: f32, col: Rgba) {
        let mut x0 = x0.round() as i32;
        let mut y0 = y0.round() as i32;
        let x1 = x1.round() as i32;
        let y1 = y1.round() as i32;

        let dx = (x1 - x0).abs();
        let sx = if x0 < x1 { 1 } else { -1 };
        let dy = -(y1 - y0).abs();
        let sy = if y0 < y1 { 1 } else { -1 };
        let mut err = dx + dy;

        loop {
            self.set_pixel(x0, y0, col);
            if x0 == x1 && y0 == y1 {
                break;
            }
            let e2 = 2 * err;
            if e2 >= dy {
                err += dy;
                x0 += sx;
            }
            if e2 <= dx {
                err += dx;
                y0 += sy;
            }
        }
    }

    /// Draws a filled circle centered at `(cx, cy)` with radius `r`.
    pub fn draw_filled_circle(&mut self, cx: f32, cy: f32, r: f32, col: Rgba) {
        if r <= 0.0 {
            return;
        }
        let min_x = (cx - r).floor() as i32;
        let max_x = (cx + r).ceil() as i32;
        let min_y = (cy - r).floor() as i32;
        let max_y = (cy + r).ceil() as i32;
        let r_sq = r * r;

        for py in min_y..=max_y {
            for px in min_x..=max_x {
                let dx = (px as f32 + 0.5) - cx;
                let dy = (py as f32 + 0.5) - cy;
                if dx * dx + dy * dy <= r_sq {
                    self.set_pixel(px, py, col);
                }
            }
        }
    }
}
