use crate::audio::{
    init_audio, play_bonus_city_sound, play_siren_alert, play_tally_sound,
    set_sound_effects_enabled, stop_all_continuous_sounds,
};
use crate::input::InputState;
use crate::math::{Point, SIM_HEIGHT, SIM_WIDTH};
use crate::objects::{
    get_wave_data, Battery, City, Crosshair, Explosion, Flier, SmartBomb, ABM, ICBM,
};
use crate::palette::{get_palette_for_wave, Palette};
use crate::pipeline::Pipeline;
use crate::scores::{load_high_scores, HighScores};
use crate::settings::{load_settings, GameSettings};

/// High-level arcade game state machine.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum GameState {
    #[default]
    Attract,
    WaveStart,
    Playing,
    Tally,
    BonusRebuild,
    TheEnd,
    HighScoreEntry,
    Options,
}

pub const CITY_X_POSITIONS: [f64; 6] = [40.0, 65.0, 90.0, 150.0, 175.0, 200.0];
pub const BATTERY_X_POSITIONS: [f64; 3] = [18.0, 128.0, 238.0];

/// Game implements the full 1980 Atari Missile Command arcade state machine and simulation.
#[derive(Debug, Clone)]
pub struct Game {
    pub width: i32,
    pub height: i32,

    // Rendering pipeline & High Scores
    pub pipeline: Pipeline,
    pub high_scores: HighScores,

    // State Machine
    pub state: GameState,
    pub state_tick: i32,

    // Crosshair
    pub crosshair: Crosshair,

    // Game Entities
    pub cities: [City; 6],
    pub batteries: [Battery; 3],
    pub icbms: Vec<ICBM>,
    pub smart_bombs: Vec<SmartBomb>,
    pub active_flier: Option<Flier>,
    pub abms: Vec<ABM>,
    pub explosions: Vec<Explosion>,

    // Wave & Scoring
    pub wave: i32,
    pub score: i32,
    pub bonus_cities_stored: i32,
    pub next_bonus_threshold: i32,
    pub palette: Palette,

    // Wave Spawning Management
    pub icbms_remaining: i32,
    pub smart_bombs_left: i32,
    pub spawn_cooldown: i32,
    pub flier_cooldown: i32,
    pub flier_spawned_this_wave: bool,
    pub wave_end_grace_timer: i32,
    pub cities_lost_this_wave: usize,

    // End-of-Wave Tally State
    pub tally_phase: i32, // 0: Ammo tally, 1: City tally, 2: Complete
    pub tally_timer: i32,
    pub tally_battery_idx: usize,
    pub tally_city_idx: usize,
    pub tally_score_earned: i32,

    // Rebuild bonus city state
    pub rebuild_city_idx: usize,
    pub rebuild_timer: i32,

    // "THE END" state
    pub the_end_timer: i32,

    // High score initials entry
    pub initials: [char; 3],
    pub initial_slot: usize,
    pub char_select_idx: usize,
    pub letters: Vec<char>,

    // Attract Mode & AI Demo
    pub attract_demo_mode: bool,
    pub attract_timer: i32,
    pub ai_demo_shoot_tick: i32,

    // On-screen HUD Notification
    pub notification_text: String,
    pub notification_timer: i32,

    // Mouse, Fullscreen, Pause and Input
    pub prev_mouse_x: i32,
    pub prev_mouse_y: i32,
    pub mouse_captured: bool,
    pub fullscreen: bool,
    pub paused: bool,
    pub should_quit: bool,
    pub tick: i32,

    // Options & Settings
    pub settings: GameSettings,
    pub option_selected_idx: usize,
    pub previous_state: Option<GameState>,
}

impl Default for Game {
    fn default() -> Self {
        Self::new()
    }
}

impl Game {
    /// Initializes the game engine and loads scores/audio/settings.
    pub fn new() -> Self {
        init_audio();
        let settings = load_settings();
        set_sound_effects_enabled(settings.sound_effects_enabled);

        let mut pipeline = Pipeline::new();
        pipeline.use_crt = settings.use_crt;

        Self {
            width: 1024,
            height: 924,
            pipeline,
            high_scores: load_high_scores(),
            state: GameState::Attract,
            state_tick: 0,
            crosshair: Crosshair {
                pos: Point::new(SIM_WIDTH / 2.0, SIM_HEIGHT / 2.0),
            },
            cities: [City::default(); 6],
            batteries: [Battery::default(); 3],
            icbms: Vec::new(),
            smart_bombs: Vec::new(),
            active_flier: None,
            abms: Vec::new(),
            explosions: Vec::new(),
            wave: 1,
            score: 0,
            bonus_cities_stored: 0,
            next_bonus_threshold: 10000,
            palette: get_palette_for_wave(1),
            icbms_remaining: 0,
            smart_bombs_left: 0,
            spawn_cooldown: 0,
            flier_cooldown: 0,
            flier_spawned_this_wave: false,
            wave_end_grace_timer: 0,
            cities_lost_this_wave: 0,
            tally_phase: 0,
            tally_timer: 0,
            tally_battery_idx: 0,
            tally_city_idx: 0,
            tally_score_earned: 0,
            rebuild_city_idx: 0,
            rebuild_timer: 0,
            the_end_timer: 0,
            initials: ['A', 'A', 'A'],
            initial_slot: 0,
            char_select_idx: 0,
            letters: "ABCDEFGHIJKLMNOPQRSTUVWXYZ.!- 0123456789".chars().collect(),
            attract_demo_mode: false,
            attract_timer: 600, // 10 seconds Great Scores display before AI demo
            ai_demo_shoot_tick: 0,
            notification_text: String::new(),
            notification_timer: 0,
            prev_mouse_x: 0,
            prev_mouse_y: 0,
            mouse_captured: false,
            fullscreen: false,
            paused: false,
            should_quit: false,
            tick: 0,
            settings,
            option_selected_idx: 0,
            previous_state: None,
        }
    }

    pub fn enter_attract_mode(&mut self) {
        self.paused = false;
        stop_all_continuous_sounds();
        self.state = GameState::Attract;
        self.attract_demo_mode = false;
        self.attract_timer = 600;
    }

    pub fn show_notification(&mut self, msg: impl Into<String>) {
        self.notification_text = msg.into();
        self.notification_timer = 90;
    }

    /// Starts a fresh 1-player game.
    pub fn start_game(&mut self) {
        self.wave = 1;
        self.score = 0;
        self.bonus_cities_stored = 0;
        self.next_bonus_threshold = 10000;
        self.palette = get_palette_for_wave(1);

        self.reset_cities();
        self.reset_silos();
        self.start_wave(1);
    }

    pub fn reset_cities(&mut self) {
        for (i, &x) in CITY_X_POSITIONS.iter().enumerate() {
            self.cities[i] = City {
                position: Point::new(x, 216.0),
                destroyed: false,
            };
        }
    }

    pub fn reset_silos(&mut self) {
        let max_ammo = if self.settings.missiles_per_silo > 0 {
            self.settings.missiles_per_silo
        } else {
            10
        };
        for (i, &x) in BATTERY_X_POSITIONS.iter().enumerate() {
            self.batteries[i] = Battery {
                index: i,
                position: Point::new(x, 214.0),
                max_ammo,
                ammo: max_ammo,
                destroyed: false,
                low_warning_played: false,
            };
        }
    }

    pub fn clear_active_entities(&mut self) {
        self.icbms.clear();
        self.smart_bombs.clear();
        self.active_flier = None;
        self.abms.clear();
        self.explosions.clear();
        stop_all_continuous_sounds();
    }

    pub fn start_wave(&mut self, wave: i32) {
        self.wave = wave;
        self.palette = get_palette_for_wave(wave);
        let wdata = get_wave_data(wave);

        self.icbms_remaining = wdata.total_icbms;
        self.smart_bombs_left = wdata.smart_bombs;
        self.spawn_cooldown = 45;
        self.flier_cooldown = wdata.flier_delay;
        self.flier_spawned_this_wave = false;
        self.wave_end_grace_timer = 0;
        self.cities_lost_this_wave = 0;

        self.clear_active_entities();
        self.reset_silos();

        self.state = GameState::WaveStart;
        self.state_tick = 90; // 1.5 seconds wave start banner
        play_siren_alert();
    }

    /// Steps the game state machine with default (empty) input.
    pub fn update(&mut self) {
        let input = InputState {
            cursor_x: self.crosshair.pos.x.round() as i32,
            cursor_y: self.crosshair.pos.y.round() as i32,
            ..Default::default()
        };
        self.update_with_input(&input);
    }

    /// Handles game logic across all arcade states for a given frame's `InputState`.
    pub fn update_with_input(&mut self, input: &InputState) {
        self.tick += 1;

        if self.handle_global_input(input) {
            return;
        }

        if self.paused {
            return;
        }

        match self.state {
            GameState::Attract => self.update_attract(input),
            GameState::WaveStart => self.update_wave_start(),
            GameState::Playing => self.update_playing(input),
            GameState::Tally => self.update_tally(),
            GameState::BonusRebuild => self.update_bonus_rebuild(),
            GameState::TheEnd => self.update_the_end(),
            GameState::HighScoreEntry => self.update_high_score_entry(input),
            GameState::Options => self.update_options(input),
        }
    }

    pub fn update_wave_start(&mut self) {
        self.state_tick -= 1;
        if self.state_tick <= 0 {
            self.state = GameState::Playing;
        }
    }

    pub fn update_playing(&mut self, input: &InputState) {
        self.handle_crosshair_input(input);
        self.handle_firing_input(input);
        self.step_simulation();

        if self.score >= self.next_bonus_threshold {
            self.bonus_cities_stored += 1;
            self.next_bonus_threshold += 10000;
            play_bonus_city_sound();
        }

        self.check_wave_completion();
    }

    pub fn update_tally(&mut self) {
        self.tally_timer -= 1;
        if self.tally_timer > 0 {
            return;
        }
        self.tally_timer = 8;

        match self.tally_phase {
            0 => {
                // Tally Unused ABM Ammo (+5 * Multiplier each)
                while self.tally_battery_idx < self.batteries.len() {
                    let bat = &mut self.batteries[self.tally_battery_idx];
                    if bat.ammo > 0 {
                        let mut decrement = if bat.ammo > 20 {
                            5
                        } else if bat.ammo > 10 {
                            2
                        } else {
                            1
                        };
                        decrement = decrement.min(bat.ammo);
                        bat.ammo -= decrement;
                        let pts = decrement * 5 * self.palette.multiplier;
                        self.score += pts;
                        self.tally_score_earned += pts;
                        play_tally_sound();
                        if bat.max_ammo > 10 {
                            self.tally_timer = 3;
                        }
                        return;
                    }
                    self.tally_battery_idx += 1;
                }
                self.tally_phase = 1;
            }
            1 => {
                // Tally Surviving Cities (+100 * Multiplier each)
                while self.tally_city_idx < self.cities.len() {
                    let idx = self.tally_city_idx;
                    self.tally_city_idx += 1;
                    if !self.cities[idx].destroyed {
                        let pts = 100 * self.palette.multiplier;
                        self.score += pts;
                        self.tally_score_earned += pts;
                        play_tally_sound();
                        return;
                    }
                }
                self.tally_phase = 2;
                self.tally_timer = 55;
            }
            _ => {
                self.state = GameState::BonusRebuild;
                self.rebuild_city_idx = 0;
                self.rebuild_timer = 20;
            }
        }
    }

    pub fn update_bonus_rebuild(&mut self) {
        self.rebuild_timer -= 1;
        if self.rebuild_timer > 0 {
            return;
        }

        if self.bonus_cities_stored > 0 {
            for city in &mut self.cities {
                if city.destroyed {
                    city.destroyed = false;
                    self.bonus_cities_stored -= 1;
                    play_bonus_city_sound();
                    self.rebuild_timer = 45;
                    return;
                }
            }
        }

        self.start_wave(self.wave + 1);
    }

    pub fn update_the_end(&mut self) {
        self.the_end_timer -= 1;
        if self.the_end_timer <= 0 {
            if self.high_scores.is_high_score(self.score) {
                self.state = GameState::HighScoreEntry;
                self.initials = ['A', 'A', 'A'];
                self.initial_slot = 0;
                self.char_select_idx = 0;
            } else {
                self.state = GameState::Attract;
            }
        }
    }

    pub fn update_high_score_entry(&mut self, input: &InputState) {
        if input.key_up_just_pressed {
            self.char_select_idx =
                (self.char_select_idx + self.letters.len() - 1) % self.letters.len();
            self.initials[self.initial_slot] = self.letters[self.char_select_idx];
        }
        if input.key_down_just_pressed {
            self.char_select_idx = (self.char_select_idx + 1) % self.letters.len();
            self.initials[self.initial_slot] = self.letters[self.char_select_idx];
        }

        for (idx, &pressed) in input.typed_letters_just_pressed.iter().enumerate() {
            if pressed {
                let ch = (b'A' + idx as u8) as char;
                self.initials[self.initial_slot] = ch;
                self.advance_initial_slot();
                return;
            }
        }

        if input.key_enter_just_pressed
            || input.key_space_just_pressed
            || input.mouse_left_just_pressed
        {
            self.advance_initial_slot();
        }
    }

    pub fn advance_initial_slot(&mut self) {
        play_tally_sound();
        self.initial_slot += 1;
        if self.initial_slot >= 3 {
            let initial_str: String = self.initials.iter().collect();
            self.high_scores.add_score(&initial_str, self.score);
            self.state = GameState::Attract;
        } else {
            self.char_select_idx = 0;
            self.initials[self.initial_slot] = self.letters[0];
        }
    }

    /// Updates the window dimensions used for aspect-ratio letterboxing.
    pub fn layout(&mut self, outside_width: i32, outside_height: i32) -> (i32, i32) {
        self.width = outside_width;
        self.height = outside_height;
        (outside_width, outside_height)
    }
}
