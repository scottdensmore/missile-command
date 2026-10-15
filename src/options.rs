use crate::audio::{play_tally_sound, set_sound_effects_enabled};
use crate::game::{Game, GameState};
use crate::input::InputState;
use crate::math::screen_to_sim;

impl Game {
    pub fn open_options(&mut self, from_state: GameState) {
        self.previous_state = Some(from_state);
        self.option_selected_idx = 0;
        self.state = GameState::Options;
    }

    pub fn close_options(&mut self) {
        let _ = self.settings.save();
        if self.previous_state == Some(GameState::Playing) {
            for bat in &mut self.batteries {
                bat.max_ammo = self.settings.missiles_per_silo;
                if bat.ammo > self.settings.missiles_per_silo {
                    bat.ammo = self.settings.missiles_per_silo;
                }
            }
        }

        match self.previous_state {
            Some(GameState::Playing) => {
                self.paused = true;
                self.state = GameState::Playing;
            }
            Some(prev) if prev != GameState::Attract => {
                self.state = prev;
            }
            _ => {
                self.enter_attract_mode();
            }
        }
    }

    pub fn update_options(&mut self, input: &InputState) {
        let up = input.key_up_just_pressed
            || input.key_w_just_pressed
            || input.gamepads.iter().any(|gp| gp.dpad_up_just_pressed);
        let down = input.key_down_just_pressed
            || input.key_s_just_pressed
            || input.gamepads.iter().any(|gp| gp.dpad_down_just_pressed);

        if up {
            self.option_selected_idx = (self.option_selected_idx + 3) % 4;
            play_tally_sound();
        }
        if down {
            self.option_selected_idx = (self.option_selected_idx + 1) % 4;
            play_tally_sound();
        }

        let left = input.key_left_just_pressed
            || input.key_a_just_pressed
            || input.gamepads.iter().any(|gp| gp.dpad_left_just_pressed);
        let right = input.key_right_just_pressed
            || input.key_d_just_pressed
            || input.gamepads.iter().any(|gp| gp.dpad_right_just_pressed);

        if left {
            self.adjust_option(-1);
        }
        if right {
            self.adjust_option(1);
        }

        let activate = input.key_enter_just_pressed
            || input.key_space_just_pressed
            || input.gamepads.iter().any(|gp| gp.fire_left_just_pressed);
        if activate {
            self.activate_option();
        }

        if input.key_escape_just_pressed || input.key_q_just_pressed {
            self.close_options();
            return;
        }

        if input.mouse_left_just_pressed {
            let sim_pos = screen_to_sim(
                input.cursor_x as f64,
                input.cursor_y as f64,
                self.width,
                self.height,
            );
            let start_y = 64;
            let row_height = 24;
            for i in 0..4 {
                let y = start_y + (i as i32) * row_height;
                if sim_pos.y >= (y - 4) as f64
                    && sim_pos.y <= (y + 16) as f64
                    && (16.0..=240.0).contains(&sim_pos.x)
                {
                    self.option_selected_idx = i;
                    match i {
                        0 => self.toggle_sound_option(),
                        1 => {
                            if sim_pos.x < 192.0 {
                                self.adjust_missiles_option(-10);
                            } else {
                                self.adjust_missiles_option(10);
                            }
                        }
                        2 => self.toggle_crt_option(),
                        3 => self.close_options(),
                        _ => {}
                    }
                    return;
                }
            }
        }
    }

    pub fn adjust_option(&mut self, dir: i32) {
        match self.option_selected_idx {
            0 => self.toggle_sound_option(),
            1 => self.adjust_missiles_option(dir * 10),
            2 => self.toggle_crt_option(),
            _ => {}
        }
    }

    pub fn activate_option(&mut self) {
        match self.option_selected_idx {
            0 => self.toggle_sound_option(),
            1 => {
                let mut new_val = self.settings.missiles_per_silo + 10;
                if new_val > 100 {
                    new_val = 10;
                }
                self.settings.missiles_per_silo = new_val;
                let _ = self.settings.save();
                play_tally_sound();
            }
            2 => self.toggle_crt_option(),
            3 => self.close_options(),
            _ => {}
        }
    }

    pub fn toggle_sound_option(&mut self) {
        self.settings.sound_effects_enabled = !self.settings.sound_effects_enabled;
        set_sound_effects_enabled(self.settings.sound_effects_enabled);
        let _ = self.settings.save();
        if self.settings.sound_effects_enabled {
            play_tally_sound();
        }
    }

    pub fn adjust_missiles_option(&mut self, delta: i32) {
        let new_val = (self.settings.missiles_per_silo + delta).clamp(10, 100);
        if new_val != self.settings.missiles_per_silo {
            self.settings.missiles_per_silo = new_val;
            let _ = self.settings.save();
            play_tally_sound();
        }
    }

    pub fn toggle_crt_option(&mut self) {
        self.pipeline.use_crt = !self.pipeline.use_crt;
        self.settings.use_crt = self.pipeline.use_crt;
        let _ = self.settings.save();
        play_tally_sound();
    }
}
