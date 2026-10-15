use crate::audio::{adjust_volume, get_master_volume, stop_all_continuous_sounds, toggle_mute};
use crate::game::{Game, GameState};
use crate::math::{clamp, screen_to_sim, Point, SIM_WIDTH};
use crate::objects::get_wave_data;
use crate::palette::get_palette_for_wave;

/// State of a single connected gamepad for the current frame.
#[derive(Debug, Clone, Default)]
pub struct GamepadInput {
    pub left_stick_x: f64,
    pub left_stick_y: f64,
    pub dpad_left_held: bool,
    pub dpad_right_held: bool,
    pub dpad_up_held: bool,
    pub dpad_down_held: bool,
    pub dpad_left_just_pressed: bool,
    pub dpad_right_just_pressed: bool,
    pub dpad_up_just_pressed: bool,
    pub dpad_down_just_pressed: bool,
    pub fire_left_just_pressed: bool,   // LB / LT / South (A)
    pub fire_center_just_pressed: bool, // West (X) / North (Y)
    pub fire_right_just_pressed: bool,  // RB / RT / East (B)
    pub start_just_pressed: bool,
    pub select_just_pressed: bool,
}

/// Snapshot of player input for a single simulation frame.
#[derive(Debug, Clone)]
pub struct InputState {
    pub cursor_x: i32,
    pub cursor_y: i32,
    pub mouse_left_just_pressed: bool,
    pub mouse_middle_just_pressed: bool,
    pub mouse_right_just_pressed: bool,

    // Held keys
    pub key_left_held: bool,
    pub key_right_held: bool,
    pub key_up_held: bool,
    pub key_down_held: bool,
    pub key_alt_held: bool,

    // Just-pressed keys
    pub key_left_just_pressed: bool,
    pub key_right_just_pressed: bool,
    pub key_up_just_pressed: bool,
    pub key_down_just_pressed: bool,
    pub key_w_just_pressed: bool,
    pub key_a_just_pressed: bool,
    pub key_s_just_pressed: bool,
    pub key_d_just_pressed: bool,
    pub key_1_just_pressed: bool,
    pub key_2_just_pressed: bool,
    pub key_3_just_pressed: bool,
    pub key_space_just_pressed: bool,
    pub key_enter_just_pressed: bool,
    pub key_escape_just_pressed: bool,
    pub key_f1_just_pressed: bool,
    pub key_tab_just_pressed: bool,
    pub key_f11_just_pressed: bool,
    pub key_m_just_pressed: bool,
    pub key_minus_just_pressed: bool,
    pub key_equal_just_pressed: bool,
    pub key_o_just_pressed: bool,
    pub key_p_just_pressed: bool,
    pub key_q_just_pressed: bool,

    // Direct letter typing A..=Z (indices 0..26)
    pub typed_letters_just_pressed: [bool; 26],

    pub gamepads: Vec<GamepadInput>,
}

impl Default for InputState {
    fn default() -> Self {
        Self {
            cursor_x: 128,
            cursor_y: 115,
            mouse_left_just_pressed: false,
            mouse_middle_just_pressed: false,
            mouse_right_just_pressed: false,
            key_left_held: false,
            key_right_held: false,
            key_up_held: false,
            key_down_held: false,
            key_alt_held: false,
            key_left_just_pressed: false,
            key_right_just_pressed: false,
            key_up_just_pressed: false,
            key_down_just_pressed: false,
            key_w_just_pressed: false,
            key_a_just_pressed: false,
            key_s_just_pressed: false,
            key_d_just_pressed: false,
            key_1_just_pressed: false,
            key_2_just_pressed: false,
            key_3_just_pressed: false,
            key_space_just_pressed: false,
            key_enter_just_pressed: false,
            key_escape_just_pressed: false,
            key_f1_just_pressed: false,
            key_tab_just_pressed: false,
            key_f11_just_pressed: false,
            key_m_just_pressed: false,
            key_minus_just_pressed: false,
            key_equal_just_pressed: false,
            key_o_just_pressed: false,
            key_p_just_pressed: false,
            key_q_just_pressed: false,
            typed_letters_just_pressed: [false; 26],
            gamepads: Vec::new(),
        }
    }
}

impl InputState {
    /// Accumulates per-render-frame input into a pending simulation tick snapshot
    /// so `just_pressed` triggers are never lost on >60Hz (e.g. 120Hz ProMotion) displays.
    pub fn merge_frame(&mut self, frame: &InputState) {
        self.cursor_x = frame.cursor_x;
        self.cursor_y = frame.cursor_y;

        self.mouse_left_just_pressed |= frame.mouse_left_just_pressed;
        self.mouse_middle_just_pressed |= frame.mouse_middle_just_pressed;
        self.mouse_right_just_pressed |= frame.mouse_right_just_pressed;

        self.key_left_held = frame.key_left_held;
        self.key_right_held = frame.key_right_held;
        self.key_up_held = frame.key_up_held;
        self.key_down_held = frame.key_down_held;
        self.key_alt_held = frame.key_alt_held;

        self.key_left_just_pressed |= frame.key_left_just_pressed;
        self.key_right_just_pressed |= frame.key_right_just_pressed;
        self.key_up_just_pressed |= frame.key_up_just_pressed;
        self.key_down_just_pressed |= frame.key_down_just_pressed;
        self.key_w_just_pressed |= frame.key_w_just_pressed;
        self.key_a_just_pressed |= frame.key_a_just_pressed;
        self.key_s_just_pressed |= frame.key_s_just_pressed;
        self.key_d_just_pressed |= frame.key_d_just_pressed;
        self.key_1_just_pressed |= frame.key_1_just_pressed;
        self.key_2_just_pressed |= frame.key_2_just_pressed;
        self.key_3_just_pressed |= frame.key_3_just_pressed;
        self.key_space_just_pressed |= frame.key_space_just_pressed;
        self.key_enter_just_pressed |= frame.key_enter_just_pressed;
        self.key_escape_just_pressed |= frame.key_escape_just_pressed;
        self.key_f1_just_pressed |= frame.key_f1_just_pressed;
        self.key_tab_just_pressed |= frame.key_tab_just_pressed;
        self.key_f11_just_pressed |= frame.key_f11_just_pressed;
        self.key_m_just_pressed |= frame.key_m_just_pressed;
        self.key_minus_just_pressed |= frame.key_minus_just_pressed;
        self.key_equal_just_pressed |= frame.key_equal_just_pressed;
        self.key_o_just_pressed |= frame.key_o_just_pressed;
        self.key_p_just_pressed |= frame.key_p_just_pressed;
        self.key_q_just_pressed |= frame.key_q_just_pressed;

        for (dst, &src) in self
            .typed_letters_just_pressed
            .iter_mut()
            .zip(frame.typed_letters_just_pressed.iter())
        {
            *dst |= src;
        }

        if self.gamepads.len() != frame.gamepads.len() {
            self.gamepads = frame.gamepads.clone();
        } else {
            for (dst, src) in self.gamepads.iter_mut().zip(frame.gamepads.iter()) {
                dst.left_stick_x = src.left_stick_x;
                dst.left_stick_y = src.left_stick_y;
                dst.dpad_left_held = src.dpad_left_held;
                dst.dpad_right_held = src.dpad_right_held;
                dst.dpad_up_held = src.dpad_up_held;
                dst.dpad_down_held = src.dpad_down_held;
                dst.dpad_left_just_pressed |= src.dpad_left_just_pressed;
                dst.dpad_right_just_pressed |= src.dpad_right_just_pressed;
                dst.dpad_up_just_pressed |= src.dpad_up_just_pressed;
                dst.dpad_down_just_pressed |= src.dpad_down_just_pressed;
                dst.fire_left_just_pressed |= src.fire_left_just_pressed;
                dst.fire_center_just_pressed |= src.fire_center_just_pressed;
                dst.fire_right_just_pressed |= src.fire_right_just_pressed;
                dst.start_just_pressed |= src.start_just_pressed;
                dst.select_just_pressed |= src.select_just_pressed;
            }
        }
    }

    /// Clears all edge-triggered (`just_pressed`) flags after a 60Hz simulation step consumes them.
    pub fn clear_just_pressed(&mut self) {
        self.mouse_left_just_pressed = false;
        self.mouse_middle_just_pressed = false;
        self.mouse_right_just_pressed = false;
        self.key_left_just_pressed = false;
        self.key_right_just_pressed = false;
        self.key_up_just_pressed = false;
        self.key_down_just_pressed = false;
        self.key_w_just_pressed = false;
        self.key_a_just_pressed = false;
        self.key_s_just_pressed = false;
        self.key_d_just_pressed = false;
        self.key_1_just_pressed = false;
        self.key_2_just_pressed = false;
        self.key_3_just_pressed = false;
        self.key_space_just_pressed = false;
        self.key_enter_just_pressed = false;
        self.key_escape_just_pressed = false;
        self.key_f1_just_pressed = false;
        self.key_tab_just_pressed = false;
        self.key_f11_just_pressed = false;
        self.key_m_just_pressed = false;
        self.key_minus_just_pressed = false;
        self.key_equal_just_pressed = false;
        self.key_o_just_pressed = false;
        self.key_p_just_pressed = false;
        self.key_q_just_pressed = false;
        self.typed_letters_just_pressed = [false; 26];
        for gp in &mut self.gamepads {
            gp.dpad_left_just_pressed = false;
            gp.dpad_right_just_pressed = false;
            gp.dpad_up_just_pressed = false;
            gp.dpad_down_just_pressed = false;
            gp.fire_left_just_pressed = false;
            gp.fire_center_just_pressed = false;
            gp.fire_right_just_pressed = false;
            gp.start_just_pressed = false;
            gp.select_just_pressed = false;
        }
    }
}

impl Game {
    /// Processes window/audio/options/pause hotkeys and returns true if `update` should return early.
    pub fn handle_global_input(&mut self, input: &InputState) -> bool {
        // Toggle Mouse Capture on Click
        if input.mouse_left_just_pressed && !self.mouse_captured {
            self.mouse_captured = true;
            self.prev_mouse_x = input.cursor_x;
            self.prev_mouse_y = input.cursor_y;
        }
        if input.key_escape_just_pressed && self.mouse_captured {
            self.mouse_captured = false;
        }

        // CRT Filter Toggle (F1 / Tab)
        if input.key_f1_just_pressed || input.key_tab_just_pressed {
            self.pipeline.use_crt = !self.pipeline.use_crt;
            self.settings.use_crt = self.pipeline.use_crt;
            let _ = self.settings.save();
            if self.pipeline.use_crt {
                self.show_notification("CRT SHADER: ON");
            } else {
                self.show_notification("CRT SHADER: OFF");
            }
        }

        // Fullscreen Toggle (F11 / Alt+Enter)
        if input.key_f11_just_pressed || (input.key_enter_just_pressed && input.key_alt_held) {
            self.fullscreen = !self.fullscreen;
            if self.fullscreen {
                self.show_notification("FULLSCREEN");
            } else {
                self.show_notification("WINDOWED");
            }
        }

        // Sound Mute Toggle (M Key)
        if input.key_m_just_pressed {
            let muted = toggle_mute();
            self.settings.sound_effects_enabled = !muted;
            let _ = self.settings.save();
            if muted {
                self.show_notification("SOUND: MUTED");
            } else {
                self.show_notification(format!("SOUND: {}%", (get_master_volume() * 100.0) as i32));
            }
        }

        // Volume Adjust (- / +)
        if input.key_minus_just_pressed {
            let vol = adjust_volume(-0.1);
            self.show_notification(format!("VOL: {}%", (vol * 100.0).round() as i32));
        }
        if input.key_equal_just_pressed {
            let vol = adjust_volume(0.1);
            self.show_notification(format!("VOL: {}%", (vol * 100.0).round() as i32));
        }

        if self.notification_timer > 0 {
            self.notification_timer -= 1;
        }

        // Options Toggle (O Key or Gamepad Select/Back)
        let is_options_pressed =
            input.key_o_just_pressed || input.gamepads.iter().any(|gp| gp.select_just_pressed);

        if is_options_pressed {
            match self.state {
                GameState::Options => {
                    self.close_options();
                    return true;
                }
                GameState::Attract => {
                    self.open_options(GameState::Attract);
                    return true;
                }
                GameState::Playing => {
                    if !self.paused {
                        self.paused = true;
                        stop_all_continuous_sounds();
                    }
                    self.open_options(GameState::Playing);
                    return true;
                }
                _ if self.paused => {
                    self.open_options(GameState::Playing);
                    return true;
                }
                _ => {}
            }
        }

        if self.state == GameState::Options {
            self.update_options(input);
            return true;
        }

        // Quit / Exit Handling (Q Key)
        if input.key_q_just_pressed {
            if self.state == GameState::Attract {
                self.should_quit = true;
                return true;
            }
            self.enter_attract_mode();
            return true;
        }

        // Pause / Resume Toggle (P Key or Gamepad Start in game)
        if self.state != GameState::Attract && self.state != GameState::HighScoreEntry {
            let is_pause_pressed =
                input.key_p_just_pressed || input.gamepads.iter().any(|gp| gp.start_just_pressed);
            if is_pause_pressed {
                self.paused = !self.paused;
                if self.paused {
                    stop_all_continuous_sounds();
                } else {
                    self.restore_continuous_sounds();
                }
            }
        }

        false
    }

    pub fn check_any_start_input(&self, input: &InputState) -> bool {
        if input.mouse_left_just_pressed
            || input.key_space_just_pressed
            || input.key_1_just_pressed
            || input.key_enter_just_pressed
        {
            return true;
        }
        input.gamepads.iter().any(|gp| {
            gp.start_just_pressed
                || gp.fire_left_just_pressed
                || gp.fire_center_just_pressed
                || gp.fire_right_just_pressed
        })
    }

    pub fn handle_crosshair_input(&mut self, input: &InputState) {
        if self.mouse_captured {
            let dx = (input.cursor_x - self.prev_mouse_x) as f64;
            let dy = (input.cursor_y - self.prev_mouse_y) as f64;
            self.crosshair.pos.x += dx * 0.45;
            self.crosshair.pos.y += dy * 0.45;
            self.prev_mouse_x = input.cursor_x;
            self.prev_mouse_y = input.cursor_y;
        } else {
            self.crosshair.pos = screen_to_sim(
                input.cursor_x as f64,
                input.cursor_y as f64,
                self.width,
                self.height,
            );
        }

        let speed = 2.5;
        if input.key_left_held {
            self.crosshair.pos.x -= speed;
        }
        if input.key_right_held {
            self.crosshair.pos.x += speed;
        }
        if input.key_up_held {
            self.crosshair.pos.y -= speed;
        }
        if input.key_down_held {
            self.crosshair.pos.y += speed;
        }

        for gp in &input.gamepads {
            if gp.left_stick_x.abs() > 0.18 {
                self.crosshair.pos.x += gp.left_stick_x * 3.5;
            }
            if gp.left_stick_y.abs() > 0.18 {
                self.crosshair.pos.y += gp.left_stick_y * 3.5;
            }
            if gp.dpad_left_held {
                self.crosshair.pos.x -= speed;
            }
            if gp.dpad_right_held {
                self.crosshair.pos.x += speed;
            }
            if gp.dpad_up_held {
                self.crosshair.pos.y -= speed;
            }
            if gp.dpad_down_held {
                self.crosshair.pos.y += speed;
            }
        }

        self.crosshair.pos.x = clamp(self.crosshair.pos.x, 4.0, SIM_WIDTH - 4.0);
        self.crosshair.pos.y = clamp(self.crosshair.pos.y, 8.0, 206.0);
    }

    pub fn handle_firing_input(&mut self, input: &InputState) {
        let mut fired_idx: Option<usize> = None;

        if input.key_a_just_pressed || input.key_1_just_pressed {
            fired_idx = Some(0);
        } else if input.key_s_just_pressed
            || input.key_2_just_pressed
            || input.key_space_just_pressed
        {
            fired_idx = Some(1);
        } else if input.key_d_just_pressed || input.key_3_just_pressed {
            fired_idx = Some(2);
        }

        if self.mouse_captured {
            if input.mouse_left_just_pressed {
                fired_idx = Some(0);
            } else if input.mouse_middle_just_pressed {
                fired_idx = Some(1);
            } else if input.mouse_right_just_pressed {
                fired_idx = Some(2);
            }
        }

        for gp in &input.gamepads {
            if gp.fire_left_just_pressed {
                fired_idx = Some(0);
            } else if gp.fire_center_just_pressed {
                fired_idx = Some(1);
            } else if gp.fire_right_just_pressed {
                fired_idx = Some(2);
            }
        }

        if let Some(idx) = fired_idx {
            self.fire_abm(idx);
        }
    }

    // --- ATTRACT MODE & AI DEMO ---

    pub fn update_attract(&mut self, input: &InputState) {
        if input.mouse_left_just_pressed {
            let sim_pos = screen_to_sim(
                input.cursor_x as f64,
                input.cursor_y as f64,
                self.width,
                self.height,
            );
            if (184.0..=198.0).contains(&sim_pos.y) && (20.0..=100.0).contains(&sim_pos.x) {
                self.open_options(GameState::Attract);
                return;
            }
        }

        if self.check_any_start_input(input) {
            self.start_game();
            return;
        }

        if self.attract_demo_mode {
            self.update_attract_demo();
        } else {
            self.attract_timer -= 1;
            if self.attract_timer <= 0 {
                self.start_attract_demo();
            }
        }
    }

    pub fn start_attract_demo(&mut self) {
        self.attract_demo_mode = true;
        self.attract_timer = 900; // 15 seconds demo
        self.ai_demo_shoot_tick = 0;

        self.wave = 1;
        self.score = 0;
        self.palette = get_palette_for_wave(1);
        let wdata = get_wave_data(1);
        self.icbms_remaining = wdata.total_icbms;
        self.smart_bombs_left = wdata.smart_bombs;
        self.spawn_cooldown = 25;
        self.flier_cooldown = 120;
        self.flier_spawned_this_wave = false;
        self.wave_end_grace_timer = 0;

        self.clear_active_entities();
        self.reset_cities();
        self.reset_silos();
    }

    pub fn update_attract_demo(&mut self) {
        self.attract_timer -= 1;
        if self.attract_timer <= 0 {
            self.stop_attract_demo();
            return;
        }

        self.step_simulation();
        self.update_ai_demo_player();

        if self.surviving_cities_count() == 0 {
            self.stop_attract_demo();
        }
    }

    pub fn stop_attract_demo(&mut self) {
        stop_all_continuous_sounds();
        self.attract_demo_mode = false;
        self.attract_timer = 600;
    }

    pub fn update_ai_demo_player(&mut self) {
        let mut best_target: Option<Point> = None;
        let mut lowest_y = -1.0;

        for icbm in &self.icbms {
            if icbm.active && icbm.curr.y > 25.0 && icbm.curr.y < 175.0 && icbm.curr.y > lowest_y {
                best_target = Some(Point::new(
                    icbm.curr.x + (icbm.target.x - icbm.curr.x) * 0.25,
                    icbm.curr.y + 28.0,
                ));
                lowest_y = icbm.curr.y;
            }
        }

        for sb in &self.smart_bombs {
            if sb.active && sb.curr.y > 25.0 && sb.curr.y < 175.0 && sb.curr.y > lowest_y {
                best_target = Some(Point::new(sb.curr.x, sb.curr.y + 22.0));
                lowest_y = sb.curr.y;
            }
        }

        if best_target.is_none() {
            if let Some(ref flier) = self.active_flier {
                if flier.active {
                    best_target = Some(Point::new(flier.x + 8.0 + flier.speed * 20.0, flier.y));
                }
            }
        }

        if let Some(target) = best_target {
            self.crosshair.pos.x = clamp(
                self.crosshair.pos.x + (target.x - self.crosshair.pos.x) * 0.18,
                4.0,
                SIM_WIDTH - 4.0,
            );
            self.crosshair.pos.y = clamp(
                self.crosshair.pos.y + (target.y - self.crosshair.pos.y) * 0.18,
                8.0,
                206.0,
            );

            self.ai_demo_shoot_tick += 1;
            if self.ai_demo_shoot_tick >= 35 {
                self.ai_demo_shoot_tick = 0;
                let target_x = self.crosshair.pos.x;
                let mut silo_idx = 1;
                if target_x < 90.0 && !self.batteries[0].destroyed && self.batteries[0].ammo > 0 {
                    silo_idx = 0;
                } else if target_x > 166.0
                    && !self.batteries[2].destroyed
                    && self.batteries[2].ammo > 0
                {
                    silo_idx = 2;
                } else if self.batteries[1].destroyed || self.batteries[1].ammo == 0 {
                    if !self.batteries[0].destroyed && self.batteries[0].ammo > 0 {
                        silo_idx = 0;
                    } else if !self.batteries[2].destroyed && self.batteries[2].ammo > 0 {
                        silo_idx = 2;
                    }
                }
                self.fire_abm(silo_idx);
            }
        }
    }
}
