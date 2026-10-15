use crate::framebuffer::Framebuffer;
use crate::game::{Game, GameState, BATTERY_X_POSITIONS};
use crate::math::{SIM_HEIGHT, SIM_WIDTH};
use crate::objects::{Battery, ExplosionState, FlierType};
use crate::palette::{
    get_explosion_color, Rgba, COLOR_BLACK, COLOR_CYAN, COLOR_GREEN, COLOR_RED, COLOR_WHITE,
    COLOR_YELLOW,
};
use crate::sprites::{
    draw_arcade_text, draw_bomber, draw_bonus_city_icon, draw_city, draw_satellite,
    draw_silo_ticks, draw_smart_bomb,
};

impl Game {
    /// Renders the arcade game visuals into the internal 256x231 framebuffer and steps the pipeline.
    pub fn draw(&mut self) {
        let sky_col = if self.state == GameState::Attract {
            COLOR_BLACK
        } else {
            self.palette.sky_color
        };

        let mut fb = std::mem::replace(&mut self.pipeline.frame_buffer, Framebuffer::new(0, 0));
        fb.fill(sky_col);
        self.render_playfield(&mut fb);
        self.pipeline.frame_buffer = fb;
        self.pipeline.step_frame();
    }

    /// Renders the current game state into the provided `Framebuffer`.
    pub fn render_playfield(&self, target: &mut Framebuffer) {
        self.draw_hud(target);

        match self.state {
            GameState::Attract => {
                self.draw_attract_screen(target);
                return;
            }
            GameState::Options => {
                self.draw_options_screen(target);
                return;
            }
            _ => {}
        }

        let show_crosshair = self.state == GameState::Playing || self.state == GameState::WaveStart;
        self.draw_entities(target, show_crosshair);

        match self.state {
            GameState::WaveStart => {
                draw_arcade_text(
                    target,
                    &format!("START WAVE {}", self.wave),
                    80,
                    80,
                    self.palette.text_color,
                );
                draw_arcade_text(
                    target,
                    &format!("{}X BONUS POINTS", self.palette.multiplier),
                    68,
                    100,
                    self.palette.multiplier_color,
                );
            }
            GameState::Tally => {
                draw_arcade_text(target, "BONUS POINTS", 80, 60, COLOR_YELLOW);
                draw_arcade_text(
                    target,
                    &format!("+{}", self.tally_score_earned),
                    108,
                    80,
                    COLOR_WHITE,
                );
            }
            GameState::TheEnd => {
                self.draw_the_end_screen(target);
            }
            GameState::HighScoreEntry => {
                self.draw_high_score_entry(target);
            }
            _ => {}
        }

        if self.paused {
            target.draw_filled_rect(36.0, 72.0, 184.0, 68.0, Rgba::new(0, 0, 0, 230));
            target.stroke_rect(36.0, 72.0, 184.0, 68.0, COLOR_YELLOW);

            let pause_col = if (self.tick / 20) % 2 == 0 {
                COLOR_YELLOW
            } else {
                COLOR_WHITE
            };
            draw_arcade_text(target, "GAME PAUSED", 84, 80, pause_col);
            draw_arcade_text(target, "PRESS P TO RESUME", 60, 96, COLOR_WHITE);
            draw_arcade_text(target, "PRESS O FOR OPTIONS", 52, 110, COLOR_CYAN);
            draw_arcade_text(target, "PRESS Q TO QUIT", 68, 124, self.palette.text_color);
        }
    }

    /// Renders terrain, cities, silos, ordnance, explosions, and the crosshair.
    /// Shared between active gameplay and the attract mode AI demo.
    pub fn draw_entities(&self, target: &mut Framebuffer, show_crosshair: bool) {
        self.draw_terrain(target);

        for city in &self.cities {
            draw_city(
                target,
                city.position.x as i32,
                city.position.y as i32,
                city.destroyed,
                self.palette.city_color,
            );
        }

        for bat in &self.batteries {
            self.draw_silo(target, bat);
        }

        let reserve_count = self.bonus_cities_stored.clamp(0, 12);
        for i in 0..reserve_count {
            draw_bonus_city_icon(
                target,
                4 + i * 10,
                (SIM_HEIGHT as i32) - 7,
                self.palette.city_color,
            );
        }

        for icbm in &self.icbms {
            if !icbm.active {
                continue;
            }
            target.stroke_line(
                icbm.start.x as f32,
                icbm.start.y as f32,
                icbm.curr.x as f32,
                icbm.curr.y as f32,
                self.palette.icbm_color,
            );
            target.set_pixel(icbm.curr.x as i32, icbm.curr.y as i32, COLOR_WHITE);
        }

        for sb in &self.smart_bombs {
            if !sb.active {
                continue;
            }
            target.stroke_line(
                sb.start.x as f32,
                sb.start.y as f32,
                sb.curr.x as f32,
                sb.curr.y as f32,
                COLOR_RED,
            );
            draw_smart_bomb(
                target,
                (sb.curr.x as i32) - 4,
                (sb.curr.y as i32) - 4,
                sb.anim_tick / 6,
                COLOR_YELLOW,
            );
        }

        if let Some(ref flier) = self.active_flier {
            if flier.active {
                if flier.flier_type == FlierType::Bomber {
                    draw_bomber(
                        target,
                        flier.x as i32,
                        flier.y as i32,
                        flier.speed > 0.0,
                        self.palette.icbm_color,
                    );
                } else {
                    draw_satellite(
                        target,
                        flier.x as i32,
                        flier.y as i32,
                        self.palette.icbm_color,
                    );
                }
            }
        }

        for abm in &self.abms {
            if !abm.active {
                continue;
            }
            target.stroke_line(
                abm.start.x as f32,
                abm.start.y as f32,
                abm.curr.x as f32,
                abm.curr.y as f32,
                self.palette.abm_color,
            );
            draw_arcade_text(
                target,
                "X",
                (abm.target.x as i32) - 4,
                (abm.target.y as i32) - 4,
                COLOR_WHITE,
            );
        }

        for exp in &self.explosions {
            if exp.state == ExplosionState::Dead {
                continue;
            }
            let col = get_explosion_color(exp.tick);
            target.draw_filled_circle(
                exp.center.x as f32,
                exp.center.y as f32,
                exp.radius as f32,
                col,
            );
            if exp.radius > 5.0 {
                target.draw_filled_circle(
                    exp.center.x as f32,
                    exp.center.y as f32,
                    (exp.radius * 0.4) as f32,
                    COLOR_WHITE,
                );
            }
        }

        if show_crosshair {
            draw_arcade_text(
                target,
                "X",
                (self.crosshair.pos.x as i32) - 4,
                (self.crosshair.pos.y as i32) - 4,
                COLOR_WHITE,
            );
        }
    }

    pub fn draw_hud(&self, target: &mut Framebuffer) {
        draw_arcade_text(target, &self.score.to_string(), 24, 8, COLOR_WHITE);
        draw_arcade_text(
            target,
            &self.high_scores.get_top_score().to_string(),
            116,
            8,
            COLOR_YELLOW,
        );

        if self.state == GameState::Playing {
            draw_arcade_text(
                target,
                &format!("{}X", self.palette.multiplier),
                220,
                8,
                self.palette.multiplier_color,
            );
        }

        if self.notification_timer > 0 {
            let x = 128 - (self.notification_text.len() as i32) * 4;
            draw_arcade_text(target, &self.notification_text, x, 22, COLOR_YELLOW);
        }
    }

    pub fn draw_terrain(&self, target: &mut Framebuffer) {
        let ground_col = self.palette.ground_color;
        target.draw_filled_rect(0.0, 224.0, SIM_WIDTH as f32, 7.0, ground_col);

        for &sx in &BATTERY_X_POSITIONS {
            let sx = sx as f32;
            target.draw_filled_rect(sx - 13.0, 219.0, 26.0, 5.0, ground_col);
            target.draw_filled_rect(sx - 10.0, 215.0, 20.0, 5.0, ground_col);
            target.draw_filled_rect(sx - 7.0, 211.0, 14.0, 5.0, ground_col);
            target.draw_filled_rect(sx - 4.0, 207.0, 8.0, 5.0, ground_col);
        }
    }

    pub fn draw_silo(&self, target: &mut Framebuffer, bat: &Battery) {
        let bx = bat.position.x as i32;
        let by = bat.position.y as i32;
        if bat.destroyed {
            draw_arcade_text(target, "...", bx - 12, by - 4, COLOR_RED);
            return;
        }

        draw_silo_ticks(
            target,
            bx,
            222,
            bat.ammo,
            bat.max_ammo,
            self.palette.silo_color,
            self.palette.ground_color,
        );

        if (self.tick / 15) % 2 == 0 {
            if bat.ammo == 0 {
                draw_arcade_text(target, "OUT", bx - 12, 196, COLOR_RED);
            } else if bat.ammo <= 3 {
                draw_arcade_text(target, "LOW", bx - 12, 196, COLOR_YELLOW);
            }
        }
    }

    pub fn draw_attract_screen(&self, target: &mut Framebuffer) {
        if self.attract_demo_mode {
            self.draw_entities(target, true);
            draw_arcade_text(target, "GAME OVER", 92, 40, COLOR_RED);
            draw_arcade_text(target, "DEFEND CITIES", 76, 55, COLOR_YELLOW);
            draw_arcade_text(target, "DEMO MODE", 92, 70, COLOR_CYAN);
            if (self.tick / 30) % 2 == 0 {
                draw_arcade_text(target, "PRESS SPACE OR CLICK", 48, 90, COLOR_WHITE);
            }
            return;
        }

        draw_arcade_text(target, "MISSILE COMMAND", 68, 30, COLOR_YELLOW);
        draw_arcade_text(target, "DEFEND CITIES", 76, 45, COLOR_RED);

        draw_arcade_text(target, "GREAT SCORES", 80, 68, COLOR_CYAN);
        for (i, entry) in self.high_scores.entries.iter().take(8).enumerate() {
            let line = format!("{:2}  {:5}  {:3}", i + 1, entry.score, entry.initials);
            draw_arcade_text(target, &line, 64, 84 + (i as i32) * 11, COLOR_WHITE);
        }

        draw_arcade_text(target, "© 1980 ATARI INC", 64, 176, COLOR_YELLOW);
        draw_arcade_text(target, "O:OPTIONS  P:PAUSE  Q:QUIT", 24, 190, COLOR_CYAN);
        if (self.tick / 30) % 2 == 0 {
            draw_arcade_text(target, "PRESS SPACE OR CLICK", 48, 208, COLOR_WHITE);
        }
    }

    pub fn draw_options_screen(&self, target: &mut Framebuffer) {
        draw_arcade_text(target, "GAME OPTIONS", 80, 28, COLOR_YELLOW);

        let back_label = if self.previous_state == Some(GameState::Playing) {
            "BACK TO GAME"
        } else {
            "BACK TO TITLE"
        };

        let (sound_val, sound_col) = if self.settings.sound_effects_enabled {
            ("[ ON  ]", COLOR_GREEN)
        } else {
            ("[ OFF ]", COLOR_RED)
        };

        let missiles_str = format!("< {:3} >", self.settings.missiles_per_silo);

        let (crt_val, crt_col) = if self.pipeline.use_crt {
            ("[ ON  ]", COLOR_GREEN)
        } else {
            ("[ OFF ]", COLOR_RED)
        };

        let rows: [(&str, &str, Rgba); 4] = [
            ("SOUND EFFECTS", sound_val, sound_col),
            ("MISSILES/SILO", &missiles_str, COLOR_YELLOW),
            ("CRT SCANLINES", crt_val, crt_col),
            (back_label, "", COLOR_WHITE),
        ];

        let start_y = 64;
        let row_height = 24;
        for (i, &(label, val, val_col)) in rows.iter().enumerate() {
            let y = start_y + (i as i32) * row_height;
            let lbl_col = if i == self.option_selected_idx {
                let cursor_col = if (self.tick / 15) % 2 == 0 {
                    COLOR_YELLOW
                } else {
                    COLOR_WHITE
                };
                draw_arcade_text(target, ">", 16, y, cursor_col);
                COLOR_YELLOW
            } else {
                COLOR_CYAN
            };

            draw_arcade_text(target, label, 28, y, lbl_col);
            if !val.is_empty() {
                draw_arcade_text(target, val, 164, y, val_col);
            }
        }

        draw_arcade_text(target, "UP/DOWN : SELECT", 64, 178, COLOR_CYAN);
        draw_arcade_text(target, "LEFT/RIGHT/CLICK : CHANGE", 28, 192, COLOR_WHITE);
        draw_arcade_text(target, "ESC/ENTER : BACK", 64, 206, COLOR_YELLOW);
    }

    pub fn draw_the_end_screen(&self, target: &mut Framebuffer) {
        let blast_radius = (((200 - self.the_end_timer) as f64) * 2.0).min(SIM_WIDTH) as f32;
        target.draw_filled_circle(
            (SIM_WIDTH / 2.0) as f32,
            (SIM_HEIGHT / 2.0 + 20.0) as f32,
            blast_radius,
            COLOR_RED,
        );

        draw_arcade_text(target, "THE END", 100, 100, COLOR_YELLOW);
        draw_arcade_text(target, "GAME OVER", 92, 120, COLOR_RED);
    }

    pub fn draw_high_score_entry(&self, target: &mut Framebuffer) {
        draw_arcade_text(target, "GREAT SCORE!", 80, 40, COLOR_YELLOW);
        draw_arcade_text(
            target,
            &format!("SCORE: {}", self.score),
            76,
            60,
            COLOR_WHITE,
        );
        draw_arcade_text(target, "ENTER YOUR INITIALS", 52, 90, COLOR_CYAN);

        let initials_str = format!(
            "{} {} {}",
            self.initials[0], self.initials[1], self.initials[2]
        );
        draw_arcade_text(target, &initials_str, 108, 120, COLOR_YELLOW);
        draw_arcade_text(
            target,
            "-",
            108 + (self.initial_slot as i32) * 16,
            130,
            COLOR_WHITE,
        );

        draw_arcade_text(target, "UP/DOWN TO CHANGE", 60, 160, COLOR_WHITE);
        draw_arcade_text(target, "CLICK/SPACE TO SELECT", 44, 175, COLOR_WHITE);
    }
}
