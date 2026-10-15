#![allow(clippy::field_reassign_with_default)]

use missile_command::assets::{get_macroquad_icon, get_window_icons};
use missile_command::audio::{
    adjust_volume, drain_audio_events, get_master_volume, init_audio, is_muted, pcm_to_wav,
    pokey_freq, set_bomber_sound, set_master_volume, set_satellite_sound, set_smart_bomb_sound,
    set_sound_effects_enabled, stop_all_continuous_sounds, PokeyLFSR, SoundEffect, SAMPLE_RATE,
};
use missile_command::framebuffer::Framebuffer;
use missile_command::game::{Game, GameState, BATTERY_X_POSITIONS, CITY_X_POSITIONS};
use missile_command::input::{GamepadInput, InputState};
use missile_command::math::{clamp, screen_to_sim, Point, SIM_HEIGHT, SIM_WIDTH};
use missile_command::objects::{
    get_wave_data, Explosion, ExplosionState, Flier, FlierType, SmartBomb, ARCADE_WAVE_TABLE, ICBM,
};
use missile_command::palette::{
    get_explosion_color, get_palette_for_wave, Rgba, COLOR_BLACK, COLOR_RED, COLOR_WHITE,
    COLOR_YELLOW, EXPLOSION_COLORS, WAVE_PALETTES,
};
use missile_command::pipeline::{compute_letterbox, evaluate_crt_pixel, Pipeline};
use missile_command::scores::{default_high_scores, load_high_scores, set_custom_high_score_path};
use missile_command::settings::{
    default_settings, load_settings, set_custom_settings_path, GameSettings,
};
use missile_command::sprites::{
    arcade_font_glyph, draw_ammo_pyramid, draw_arcade_text, draw_bomber, draw_bonus_city_icon,
    draw_city, draw_satellite, draw_silo_ticks, draw_smart_bomb,
};
use std::path::PathBuf;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::{Mutex, MutexGuard};

static TEST_COUNTER: AtomicUsize = AtomicUsize::new(0);
static TEST_MUTEX: Mutex<()> = Mutex::new(());

struct IsolatedTestEnv {
    _guard: MutexGuard<'static, ()>,
    dir: PathBuf,
    scores_path: PathBuf,
    settings_path: PathBuf,
}

impl IsolatedTestEnv {
    fn new(prefix: &str) -> (Self, Game) {
        let guard = TEST_MUTEX.lock().unwrap_or_else(|e| e.into_inner());
        let id = TEST_COUNTER.fetch_add(1, Ordering::SeqCst);
        let dir = std::env::temp_dir().join(format!(
            "missile_command_rust_test_{}_{}_{}",
            prefix,
            std::process::id(),
            id
        ));
        let _ = std::fs::create_dir_all(&dir);
        let scores_path = dir.join("highscores.json");
        let settings_path = dir.join("settings.json");

        set_custom_high_score_path(Some(&scores_path));
        set_custom_settings_path(Some(&settings_path));
        set_master_volume(1.0);
        set_sound_effects_enabled(true);
        stop_all_continuous_sounds();
        let _ = drain_audio_events();

        let game = Game::new();
        let _ = drain_audio_events();
        (
            Self {
                _guard: guard,
                dir,
                scores_path,
                settings_path,
            },
            game,
        )
    }
}

impl Drop for IsolatedTestEnv {
    fn drop(&mut self) {
        set_custom_high_score_path::<PathBuf>(None);
        set_custom_settings_path::<PathBuf>(None);
        stop_all_continuous_sounds();
        let _ = drain_audio_events();
        let _ = std::fs::remove_dir_all(&self.dir);
    }
}

#[test]
fn test_math_point_distance_lerp_clamp_and_screen_to_sim() {
    let a = Point::new(0.0, 0.0);
    let b = Point::new(3.0, 4.0);
    assert!((a.distance(b) - 5.0).abs() < 1e-9);

    let mid = a.lerp(b, 0.5);
    assert!((mid.x - 1.5).abs() < 1e-9);
    assert!((mid.y - 2.0).abs() < 1e-9);

    assert_eq!(a.lerp(b, -1.0), a);
    assert_eq!(a.lerp(b, 2.0), b);

    assert!((clamp(5.0, 10.0, 20.0) - 10.0).abs() < 1e-9);
    assert!((clamp(25.0, 10.0, 20.0) - 20.0).abs() < 1e-9);
    assert!((clamp(15.0, 10.0, 20.0) - 15.0).abs() < 1e-9);

    // Exact 4x scale (1024x924)
    let p = screen_to_sim(512.0, 462.0, 1024, 924);
    assert!((p.x - 128.0).abs() < 1e-6);
    assert!((p.y - 115.5).abs() < 1e-6);

    // Zero/invalid window dimensions fallback to center
    let fallback = screen_to_sim(100.0, 100.0, 0, 0);
    assert!((fallback.x - SIM_WIDTH / 2.0).abs() < 1e-6);
    assert!((fallback.y - SIM_HEIGHT / 2.0).abs() < 1e-6);
}

#[test]
fn test_get_palette_for_wave_and_silo_visibility() {
    assert_eq!(WAVE_PALETTES.len(), 10);

    let p1 = get_palette_for_wave(1);
    let p2 = get_palette_for_wave(2);
    assert_eq!(p1.multiplier, 1);
    assert_eq!(p2.multiplier, 1);

    let p3 = get_palette_for_wave(3);
    assert_eq!(p3.multiplier, 2);

    let p11 = get_palette_for_wave(11);
    assert_eq!(p11.multiplier, 6);

    // Wave 21 wraps palette colors back to Wave 1 while keeping 6x multiplier
    let p21 = get_palette_for_wave(21);
    assert_eq!(p21.multiplier, 6);
    assert_eq!(p21.ground_color, p1.ground_color);
    assert_eq!(p21.sky_color, p1.sky_color);

    // Across all 20+ waves, silo_color must never equal ground_color or sky_color
    for w in 1..=25 {
        let pal = get_palette_for_wave(w);
        assert_ne!(
            pal.silo_color, pal.ground_color,
            "Wave {} silo_color matches ground_color (invisible ammo ticks!)",
            w
        );
        assert_ne!(
            pal.silo_color, pal.sky_color,
            "Wave {} silo_color matches sky_color",
            w
        );
        assert!((1..=6).contains(&pal.multiplier));
    }

    // Explosion color cycling changes every 3 ticks
    for (i, &col) in EXPLOSION_COLORS.iter().enumerate() {
        assert_eq!(get_explosion_color((i * 3) as i32), col);
    }
}

#[test]
fn test_get_wave_data_progression() {
    assert_eq!(ARCADE_WAVE_TABLE.len(), 19);

    let w1 = get_wave_data(1);
    assert_eq!(w1.total_icbms, 12);
    assert!((w1.speed - 0.172).abs() < 1e-6);
    assert_eq!(w1.smart_bombs, 0);
    assert_eq!(w1.flier_delay, 320);

    let w6 = get_wave_data(6);
    assert_eq!(w6.smart_bombs, 1);

    let w15 = get_wave_data(15);
    assert!((w15.speed - 1.000).abs() < 1e-6);

    let w19 = get_wave_data(19);
    assert_eq!(w19.total_icbms, 22);
    assert!((w19.speed - 1.000).abs() < 1e-6);
    assert_eq!(w19.smart_bombs, 7);

    // Waves > 19 clamp to Wave 19 entry; wave <= 0 clamps to Wave 1
    assert_eq!(get_wave_data(25), w19);
    assert_eq!(get_wave_data(0), w1);
}

#[test]
fn test_center_silo_speed_and_abm_cap_and_ammo_warnings() {
    let (_env, mut game) = IsolatedTestEnv::new("abm");
    game.start_game();
    let _ = drain_audio_events();
    game.crosshair.pos = Point::new(128.0, 100.0);

    // Left battery (Alpha, idx 0): speed 3.2
    game.fire_abm(0);
    assert_eq!(game.abms.len(), 1);
    assert!((game.abms[0].speed - 3.2).abs() < 1e-6);
    assert_eq!(game.batteries[0].ammo, 9);

    // Center battery (Delta, idx 1): speed 7.5
    game.fire_abm(1);
    assert_eq!(game.abms.len(), 2);
    assert!((game.abms[1].speed - 7.5).abs() < 1e-6);
    assert_eq!(game.batteries[1].ammo, 9);

    // Right battery (Omega, idx 2): speed 3.2
    game.fire_abm(2);
    assert_eq!(game.abms.len(), 3);
    assert!((game.abms[2].speed - 3.2).abs() < 1e-6);
    assert_eq!(game.batteries[2].ammo, 9);

    // Fill up to 8 active ABMs
    for _ in 0..5 {
        game.fire_abm(0);
    }
    assert_eq!(game.abms.len(), 8);

    // 9th ABM is blocked by the 8-ABM cap
    let ammo_before = game.batteries[1].ammo;
    game.fire_abm(1);
    assert_eq!(game.abms.len(), 8);
    assert_eq!(game.batteries[1].ammo, ammo_before);

    // Low ammo warning triggers once when ammo drops to <= 3
    game.abms.clear();
    game.batteries[2].ammo = 4;
    game.batteries[2].low_warning_played = false;
    let _ = drain_audio_events();
    game.fire_abm(2);
    assert_eq!(game.batteries[2].ammo, 3);
    assert!(game.batteries[2].low_warning_played);
    let (events, _) = drain_audio_events();
    assert!(events.iter().any(|(e, _)| *e == SoundEffect::Launch));
    assert!(events.iter().any(|(e, _)| *e == SoundEffect::SiloLow));

    // Empty silo plays CantFire buzzer
    game.batteries[2].ammo = 0;
    game.fire_abm(2);
    let (events_empty, _) = drain_audio_events();
    assert!(events_empty
        .iter()
        .any(|(e, _)| *e == SoundEffect::CantFire));
}

#[test]
fn test_abm_arrival_and_explosion_lifecycle() {
    let (_env, mut game) = IsolatedTestEnv::new("abm_exp");
    game.start_game();
    game.crosshair.pos = Point::new(128.0, 200.0);
    game.fire_abm(1);
    assert_eq!(game.abms.len(), 1);

    for _ in 0..10 {
        game.update_abms();
    }
    assert!(game.abms.is_empty());
    assert_eq!(game.explosions.len(), 1);
    assert_eq!(game.explosions[0].state, ExplosionState::Expanding);

    // Step explosion through Expanding -> Holding -> Contracting -> Dead
    let mut saw_holding = false;
    let mut saw_contracting = false;
    for _ in 0..200 {
        game.update_explosions();
        if let Some(exp) = game.explosions.first() {
            if exp.state == ExplosionState::Holding {
                saw_holding = true;
            } else if exp.state == ExplosionState::Contracting {
                saw_contracting = true;
            }
        } else {
            break;
        }
    }
    assert!(saw_holding);
    assert!(saw_contracting);
    assert!(game.explosions.is_empty());
}

#[test]
fn test_smart_bomb_evasion_ai_and_looping_sound() {
    let (_env, mut game) = IsolatedTestEnv::new("smart_evade");
    game.start_game();

    game.smart_bombs.push(SmartBomb {
        start: Point::new(120.0, 0.0),
        curr: Point::new(120.0, 100.0),
        target: Point::new(120.0, 216.0),
        speed: 0.8,
        active: true,
        anim_tick: 0,
    });
    set_smart_bomb_sound(true);

    // Place an expanding explosion just below and to the left of the smart bomb
    game.explosions.push(Explosion {
        center: Point::new(112.0, 110.0),
        radius: 10.0,
        max_radius: 22.0,
        hold_timer: 18,
        state: ExplosionState::Expanding,
        tick: 1,
    });

    let x_before = game.smart_bombs[0].curr.x;
    game.update_smart_bombs();
    let x_after = game.smart_bombs[0].curr.x;

    // Smart bomb must evade to the right (away from x=112)
    assert!(
        x_after > x_before,
        "Expected smart bomb to evade right ({} > {})",
        x_after,
        x_before
    );

    // When smart bombs clear, looping sound turns off
    game.smart_bombs.clear();
    game.update_smart_bombs();
    let (_, looping) = drain_audio_events();
    assert!(!looping.smart_bomb_active);
}

#[test]
fn test_mirv_splitter_warheads_and_quota() {
    let (_env, mut game) = IsolatedTestEnv::new("mirv");
    game.start_game();
    game.wave = 2;
    game.icbms_remaining = 5;

    game.icbms.push(ICBM {
        start: Point::new(100.0, 0.0),
        curr: Point::new(100.0, 85.0),
        target: Point::new(100.0, 216.0),
        speed: 0.3,
        progress: 0.4,
        is_splitter: true,
        splitted: false,
        split_altitude: 80.0,
        active: true,
    });

    game.update_icbms();

    // Parent + 2 sister warheads = 3 active ICBMs, and 2 deducted from icbms_remaining
    assert_eq!(game.icbms.len(), 3);
    assert!(game.icbms.iter().any(|m| m.splitted));
    assert_eq!(game.icbms_remaining, 3);
}

#[test]
fn test_fliers_bomber_and_satellite_movement_and_bomb_drop() {
    let (_env, mut game) = IsolatedTestEnv::new("fliers");
    game.start_game();
    game.icbms_remaining = 5;

    // Spawn a flier via cooldown expiry
    game.flier_spawned_this_wave = false;
    game.flier_cooldown = 1;
    game.update_flier();
    assert!(game.active_flier.is_some());
    assert!(game.flier_spawned_this_wave);

    // Position flier inside drop zone with 1 frame left on drop_cooldown
    if let Some(ref mut flier) = game.active_flier {
        flier.flier_type = FlierType::Bomber;
        flier.x = 100.0;
        flier.y = 70.0;
        flier.speed = 1.0 / 3.0;
        flier.drop_cooldown = 1;
        flier.bombs_remaining = 2;
    }

    let icbms_before = game.icbms.len();
    game.update_flier();
    assert_eq!(game.icbms.len(), icbms_before + 1);
    assert_eq!(game.icbms_remaining, 4);
    assert!((game.active_flier.as_ref().unwrap().x - (100.0 + 1.0 / 3.0)).abs() < 1e-6);

    // Move flier past right edge of screen -> despawns and stops looping sound
    if let Some(ref mut flier) = game.active_flier {
        flier.x = SIM_WIDTH + 25.0;
        flier.speed = 0.5;
    }
    game.update_flier();
    assert!(game.active_flier.is_none());
    let (_, looping) = drain_audio_events();
    assert!(!looping.bomber_active);
    assert!(!looping.satellite_active);
}

#[test]
fn test_altitude_attack_pacing_and_threat_cap() {
    let (_env, mut game) = IsolatedTestEnv::new("pacing");
    game.start_game();
    game.spawn_cooldown = 1;

    // First spawn succeeds
    game.update_spawning();
    assert_eq!(game.icbms.len() + game.smart_bombs.len(), 1);

    // While the spawned ICBM is at y = 0.0 (< threshold_y = 31.0 for wave 1), spawning is blocked
    game.spawn_cooldown = 1;
    game.update_spawning();
    assert_eq!(game.icbms.len() + game.smart_bombs.len(), 1);

    // Once the ICBM descends below threshold_y (e.g. y = 40.0), next spawn proceeds
    game.icbms[0].curr.y = 40.0;
    game.spawn_cooldown = 1;
    game.update_spawning();
    assert_eq!(game.icbms.len() + game.smart_bombs.len(), 2);
}

#[test]
fn test_chain_reaction_collisions_and_scoring() {
    let (_env, mut game) = IsolatedTestEnv::new("collisions");
    game.start_game();
    game.start_wave(3); // Wave 3 = 2x multiplier
    game.state = GameState::Playing;
    assert_eq!(game.palette.multiplier, 2);

    // Active explosion at (100, 100) with radius 20
    game.explosions.push(Explosion {
        center: Point::new(100.0, 100.0),
        radius: 20.0,
        max_radius: 22.0,
        hold_timer: 18,
        state: ExplosionState::Expanding,
        tick: 5,
    });

    // Place an ICBM, SmartBomb, and Flier within the blast radius
    game.icbms.push(ICBM {
        start: Point::new(100.0, 0.0),
        curr: Point::new(105.0, 100.0),
        target: Point::new(105.0, 216.0),
        speed: 0.3,
        progress: 0.5,
        is_splitter: false,
        splitted: false,
        split_altitude: 0.0,
        active: true,
    });
    game.smart_bombs.push(SmartBomb {
        start: Point::new(95.0, 0.0),
        curr: Point::new(95.0, 100.0),
        target: Point::new(95.0, 216.0),
        speed: 0.3,
        active: true,
        anim_tick: 0,
    });
    game.active_flier = Some(Flier {
        flier_type: FlierType::Satellite,
        x: 92.0,
        y: 96.0,
        speed: 0.5,
        drop_cooldown: 50,
        bombs_remaining: 1,
        active: true,
        anim_tick: 0,
    });

    game.check_collisions();

    assert!(!game.icbms[0].active);
    assert!(!game.smart_bombs[0].active);
    assert!(game.active_flier.is_none());

    // Score: (25 + 125 + 100) * 2 = 500
    assert_eq!(game.score, 500);
    // 3 secondary chain-reaction explosions added (total 4)
    assert_eq!(game.explosions.len(), 4);
}

#[test]
fn test_ground_impact_destroys_city_and_battery() {
    let (_env, mut game) = IsolatedTestEnv::new("impact");
    game.start_game();

    let city0_target = Point::new(CITY_X_POSITIONS[0] + 8.0, 216.0);
    game.detonate_threat_on_ground(city0_target);
    assert!(game.cities[0].destroyed);
    assert_eq!(game.pipeline.flash_timer, 8);
    assert_eq!(game.explosions.len(), 1);

    let bat1_target = Point::new(BATTERY_X_POSITIONS[1], 214.0);
    game.detonate_threat_on_ground(bat1_target);
    assert!(game.batteries[1].destroyed);
    assert_eq!(game.batteries[1].ammo, 0);
    assert_eq!(game.pipeline.flash_timer, 6);
}

#[test]
fn test_bonus_cities_earning_and_rebuilding() {
    let (_env, mut game) = IsolatedTestEnv::new("bonus_city");
    game.start_game();
    game.state = GameState::Playing;
    let _ = drain_audio_events();

    // Score reaches 10,000 during playing update -> awards 1 bonus city
    game.score = 10500;
    game.update_playing(&InputState::default());
    assert_eq!(game.bonus_cities_stored, 1);
    assert_eq!(game.next_bonus_threshold, 20000);
    let (events, _) = drain_audio_events();
    assert!(events.iter().any(|(e, _)| *e == SoundEffect::BonusCity));

    // Destroy 1 city and store 2 bonus cities
    game.cities[2].destroyed = true;
    game.bonus_cities_stored = 2;
    game.state = GameState::BonusRebuild;
    game.rebuild_timer = 1;

    for _ in 0..100 {
        game.update_bonus_rebuild();
        if game.state == GameState::WaveStart {
            break;
        }
    }

    assert!(!game.cities[2].destroyed);
    assert_eq!(game.surviving_cities_count(), 6);
    // 1 bonus city used to rebuild cities[2], 1 remains in reserve
    assert_eq!(game.bonus_cities_stored, 1);
    assert_eq!(game.wave, 2);
}

#[test]
fn test_wave_completion_and_end_of_wave_tally() {
    let (_env, mut game) = IsolatedTestEnv::new("tally");
    game.start_game();
    game.state = GameState::Playing;
    game.icbms_remaining = 0;
    game.smart_bombs_left = 0;
    game.icbms.clear();
    game.smart_bombs.clear();
    game.abms.clear();
    game.explosions.clear();
    game.active_flier = None;

    // 3 missiles remaining (3 * 5 = 15 pts) + 2 cities surviving (2 * 100 = 200 pts) = 215 pts
    game.batteries[0].ammo = 2;
    game.batteries[1].ammo = 1;
    game.batteries[2].ammo = 0;
    for i in 2..6 {
        game.cities[i].destroyed = true;
    }

    game.check_wave_completion();
    assert_eq!(game.state, GameState::Tally);

    for _ in 0..400 {
        game.update();
        if game.state == GameState::WaveStart && game.wave == 2 {
            break;
        }
    }

    assert_eq!(game.state, GameState::WaveStart);
    assert_eq!(game.wave, 2);
    assert_eq!(game.score, 215);
    assert_eq!(game.tally_score_earned, 215);
}

#[test]
fn test_high_capacity_silo_fast_tally() {
    let (_env, mut game) = IsolatedTestEnv::new("fast_tally");
    game.settings.missiles_per_silo = 30;
    game.start_game();
    game.state = GameState::Tally;
    game.tally_phase = 0;
    game.tally_timer = 1;
    game.batteries[0].ammo = 30;
    game.batteries[1].ammo = 0;
    game.batteries[2].ammo = 0;
    for c in &mut game.cities {
        c.destroyed = true;
    }

    for _ in 0..200 {
        game.update_tally();
        if game.tally_phase == 2 {
            break;
        }
    }

    // 30 missiles * 5 pts * 1x = 150 pts
    assert_eq!(game.batteries[0].ammo, 0);
    assert_eq!(game.score, 150);
}

#[test]
fn test_all_cities_destroyed_transitions_to_the_end_and_high_score_entry() {
    let (_env, mut game) = IsolatedTestEnv::new("the_end");
    game.start_game();
    game.state = GameState::Playing;
    game.score = 85000; // Qualifies for #1 high score
    for c in &mut game.cities {
        c.destroyed = true;
    }

    game.check_wave_completion();
    assert_eq!(game.state, GameState::TheEnd);
    assert_eq!(game.pipeline.flash_timer, 25);

    for _ in 0..200 {
        game.update_the_end();
        if game.state == GameState::HighScoreEntry {
            break;
        }
    }
    assert_eq!(game.state, GameState::HighScoreEntry);

    // Slot 0: type 'R' directly
    let mut in_r = InputState::default();
    in_r.typed_letters_just_pressed[(b'R' - b'A') as usize] = true;
    game.update_high_score_entry(&in_r);
    assert_eq!(game.initials[0], 'R');
    assert_eq!(game.initial_slot, 1);

    // Slot 1: press Down arrow to change 'A' -> 'B', then Enter to confirm
    let mut in_down = InputState::default();
    in_down.key_down_just_pressed = true;
    game.update_high_score_entry(&in_down);
    assert_eq!(game.initials[1], 'B');

    let mut in_enter = InputState::default();
    in_enter.key_enter_just_pressed = true;
    game.update_high_score_entry(&in_enter);
    assert_eq!(game.initial_slot, 2);

    // Slot 2: type 'Z' to complete entry
    let mut in_z = InputState::default();
    in_z.typed_letters_just_pressed[(b'Z' - b'A') as usize] = true;
    game.update_high_score_entry(&in_z);

    assert_eq!(game.state, GameState::Attract);
    assert_eq!(game.high_scores.entries[0].initials, "RBZ");
    assert_eq!(game.high_scores.entries[0].score, 85000);
}

#[test]
fn test_high_scores_persistence_and_sorting() {
    let (env, _game) = IsolatedTestEnv::new("scores_io");
    assert!(env.scores_path.exists());

    let defaults = default_high_scores();
    assert_eq!(defaults.entries.len(), 10);
    assert_eq!(defaults.entries[0].initials, "DFT");
    assert_eq!(defaults.entries[0].score, 75000);
    assert_eq!(defaults.entries[9].initials, "ACE");
    assert_eq!(defaults.entries[9].score, 5000);

    let mut hs = load_high_scores();
    assert!(!hs.is_high_score(0));
    assert!(!hs.is_high_score(1000));
    assert!(hs.is_high_score(6000));
    assert!(hs.is_high_score(99999));

    hs.add_score("TOP", 99999);
    assert_eq!(hs.entries.len(), 10);
    assert_eq!(hs.get_top_score(), 99999);
    assert_eq!(hs.entries[0].initials, "TOP");

    let reloaded = load_high_scores();
    assert_eq!(reloaded.entries[0].initials, "TOP");
    assert_eq!(reloaded.entries[0].score, 99999);
}

#[test]
fn test_attract_mode_and_ai_demo_player() {
    let (_env, mut game) = IsolatedTestEnv::new("attract_ai");
    assert_eq!(game.state, GameState::Attract);
    assert!(!game.attract_demo_mode);

    // Fast-forward attract_timer to trigger AI demo mode
    game.attract_timer = 1;
    game.update_attract(&InputState::default());
    assert!(game.attract_demo_mode);

    // Run 150 frames of AI demo mode; verify AI moves crosshair and engages incoming threats
    let initial_pos = game.crosshair.pos;
    let mut fired_or_moved = false;
    for _ in 0..150 {
        game.update_attract(&InputState::default());
        if !game.abms.is_empty() || game.crosshair.pos != initial_pos {
            fired_or_moved = true;
        }
    }
    assert!(
        fired_or_moved,
        "AI demo player should move crosshair and fire at demo threats"
    );

    // Pressing '1' starts a real game from attract demo mode
    let mut start_in = InputState::default();
    start_in.key_1_just_pressed = true;
    game.update_attract(&start_in);
    assert_eq!(game.state, GameState::WaveStart);
    assert_eq!(game.wave, 1);
}

#[test]
fn test_options_menu_navigation_and_settings_persistence() {
    let (env, mut game) = IsolatedTestEnv::new("options");
    assert_eq!(game.settings, default_settings());

    // Open Options from Attract via 'O' key
    let mut open_in = InputState::default();
    open_in.key_o_just_pressed = true;
    game.update_with_input(&open_in);
    assert_eq!(game.state, GameState::Options);
    assert_eq!(game.option_selected_idx, 0);

    // Item 0: Sound Effects -> toggle off with Enter
    let mut enter_in = InputState::default();
    enter_in.key_enter_just_pressed = true;
    game.update_with_input(&enter_in);
    assert!(!game.settings.sound_effects_enabled);
    assert!(is_muted());

    // Move Down to Item 1: Missiles/Silo -> press Right to increase 10 -> 20
    let mut down_in = InputState::default();
    down_in.key_down_just_pressed = true;
    game.update_with_input(&down_in);
    assert_eq!(game.option_selected_idx, 1);

    let mut right_in = InputState::default();
    right_in.key_right_just_pressed = true;
    game.update_with_input(&right_in);
    assert_eq!(game.settings.missiles_per_silo, 20);

    // Move Down to Item 2: CRT Scanlines -> toggle off with Left
    game.update_with_input(&down_in);
    assert_eq!(game.option_selected_idx, 2);
    let mut left_in = InputState::default();
    left_in.key_left_just_pressed = true;
    game.update_with_input(&left_in);
    assert!(!game.pipeline.use_crt);
    assert!(!game.settings.use_crt);

    // Move Down to Item 3: Back -> press Enter to close Options
    game.update_with_input(&down_in);
    assert_eq!(game.option_selected_idx, 3);
    game.update_with_input(&enter_in);
    assert_eq!(game.state, GameState::Attract);

    // Verify settings persisted to disk
    let loaded = load_settings();
    assert!(!loaded.sound_effects_enabled);
    assert_eq!(loaded.missiles_per_silo, 20);
    assert!(!loaded.use_crt);

    // Verify normalization of out-of-range values and corrupt JSON fallback
    let mut raw = GameSettings {
        sound_effects_enabled: true,
        missiles_per_silo: -50,
        use_crt: true,
    };
    raw.normalize();
    assert_eq!(raw.missiles_per_silo, 10);

    raw.missiles_per_silo = 26;
    raw.normalize();
    assert_eq!(raw.missiles_per_silo, 30);

    raw.missiles_per_silo = 999;
    raw.normalize();
    assert_eq!(raw.missiles_per_silo, 100);

    let _ = std::fs::write(&env.settings_path, "{corrupt json");
    let recovered = load_settings();
    assert_eq!(recovered, default_settings());
}

#[test]
fn test_pause_and_global_hotkeys_and_controls() {
    let (_env, mut game) = IsolatedTestEnv::new("hotkeys");
    game.start_game();
    game.state = GameState::Playing;

    // Activate bomber and satellite loops
    set_satellite_sound(true);
    game.active_flier = Some(Flier {
        flier_type: FlierType::Bomber,
        x: 50.0,
        y: 70.0,
        speed: 0.33,
        drop_cooldown: 50,
        bombs_remaining: 1,
        active: true,
        anim_tick: 0,
    });
    set_bomber_sound(true);

    // Press P to pause -> stops continuous sounds
    let mut p_in = InputState::default();
    p_in.key_p_just_pressed = true;
    game.update_with_input(&p_in);
    assert!(game.paused);
    let (_, looping_paused) = drain_audio_events();
    assert!(!looping_paused.bomber_active);

    // Press P again to resume -> restores continuous bomber sound
    game.update_with_input(&p_in);
    assert!(!game.paused);
    let (_, looping_resumed) = drain_audio_events();
    assert!(looping_resumed.bomber_active);

    // Toggle CRT via F1
    let crt_before = game.pipeline.use_crt;
    let mut f1_in = InputState::default();
    f1_in.key_f1_just_pressed = true;
    game.update_with_input(&f1_in);
    assert_eq!(game.pipeline.use_crt, !crt_before);

    // Toggle Fullscreen via F11
    let mut f11_in = InputState::default();
    f11_in.key_f11_just_pressed = true;
    game.update_with_input(&f11_in);
    assert!(game.fullscreen);

    // Volume & Mute hotkeys
    let mut minus_in = InputState::default();
    minus_in.key_minus_just_pressed = true;
    game.update_with_input(&minus_in);
    assert!((get_master_volume() - 0.9).abs() < 1e-6);

    let mut equal_in = InputState::default();
    equal_in.key_equal_just_pressed = true;
    game.update_with_input(&equal_in);
    assert!((get_master_volume() - 1.0).abs() < 1e-6);

    let mut m_in = InputState::default();
    m_in.key_m_just_pressed = true;
    game.update_with_input(&m_in);
    assert!(is_muted());

    // Gamepad movement & firing
    let mut gp_in = InputState::default();
    gp_in.gamepads.push(GamepadInput {
        left_stick_x: 1.0,
        left_stick_y: -0.5,
        fire_center_just_pressed: true,
        ..Default::default()
    });
    game.update_with_input(&gp_in);
    assert_eq!(game.abms.len(), 1);
    assert_eq!(game.abms[0].silo_index, 1);

    // Pressing Q during Playing returns to Attract; pressing Q in Attract sets should_quit
    let mut q_in = InputState::default();
    q_in.key_q_just_pressed = true;
    game.update_with_input(&q_in);
    assert_eq!(game.state, GameState::Attract);
    assert!(!game.should_quit);

    game.update_with_input(&q_in);
    assert!(game.should_quit);
}

#[test]
fn test_framebuffer_sprites_and_full_render_all_states() {
    let mut fb = Framebuffer::new(SIM_WIDTH as usize, SIM_HEIGHT as usize);
    assert_eq!(fb.width, 256);
    assert_eq!(fb.height, 231);

    fb.clear();
    assert_eq!(fb.get_pixel(10, 10), Rgba::new(0, 0, 0, 0));

    fb.fill(COLOR_BLACK);
    assert_eq!(fb.get_pixel(10, 10), COLOR_BLACK);

    // Alpha blending in set_pixel
    fb.set_pixel(10, 10, Rgba::new(200, 100, 50, 128));
    let blended = fb.get_pixel(10, 10);
    assert!(blended.r > 90 && blended.r < 110);

    // Out-of-bounds safety
    fb.set_pixel(-5, -5, COLOR_RED);
    fb.set_pixel(999, 999, COLOR_RED);
    assert_eq!(fb.get_pixel(-5, -5), Rgba::default());

    // Verify all font glyphs exist for standard characters
    for ch in "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ :-+.!?/*©<>[]".chars() {
        assert!(
            arcade_font_glyph(ch).is_some(),
            "Missing 8x8 glyph for {:?}",
            ch
        );
    }

    draw_arcade_text(&mut fb, "MISSILE COMMAND © 1980", 8, 8, COLOR_WHITE);
    draw_city(&mut fb, 40, 216, false, COLOR_WHITE);
    draw_city(&mut fb, 65, 216, true, COLOR_WHITE);
    draw_bomber(&mut fb, 100, 70, true, COLOR_RED);
    draw_bomber(&mut fb, 120, 70, false, COLOR_RED);
    draw_satellite(&mut fb, 150, 50, COLOR_YELLOW);
    draw_smart_bomb(&mut fb, 180, 90, 0, COLOR_YELLOW);
    draw_smart_bomb(&mut fb, 190, 90, 1, COLOR_YELLOW);
    draw_bonus_city_icon(&mut fb, 10, 224, COLOR_WHITE);
    draw_ammo_pyramid(&mut fb, 18, 220, 10, COLOR_WHITE);
    draw_silo_ticks(&mut fb, 18, 222, 10, 10, COLOR_WHITE, COLOR_YELLOW);
    draw_silo_ticks(&mut fb, 128, 222, 25, 30, COLOR_WHITE, COLOR_YELLOW);

    // Render all game states via game.draw() and verify non-black pixel output
    let (_env, mut game) = IsolatedTestEnv::new("render_states");
    for state in [
        GameState::Attract,
        GameState::WaveStart,
        GameState::Playing,
        GameState::Tally,
        GameState::BonusRebuild,
        GameState::TheEnd,
        GameState::HighScoreEntry,
        GameState::Options,
    ] {
        game.state = state;
        game.draw();
        let lit_pixels = game
            .pipeline
            .frame_buffer
            .pixels
            .as_chunks::<4>()
            .0
            .iter()
            .filter(|px| px[0] > 0 || px[1] > 0 || px[2] > 0)
            .count();
        assert!(
            lit_pixels > 100,
            "State {:?} rendered too few lit pixels ({})",
            state,
            lit_pixels
        );
    }

    // Also verify paused overlay rendering and attract demo rendering
    game.state = GameState::Playing;
    game.paused = true;
    game.draw();
    game.paused = false;

    game.state = GameState::Attract;
    game.attract_demo_mode = true;
    game.draw();
}

#[test]
fn test_crt_pipeline_and_letterbox_math() {
    let lb = compute_letterbox(1024, 924);
    assert!((lb.scale - 4.0).abs() < 1e-6);
    assert!((lb.offset_x - 0.0).abs() < 1e-6);
    assert!((lb.offset_y - 0.0).abs() < 1e-6);

    let base = Rgba::new(180, 100, 60, 255);
    let crt_px = evaluate_crt_pixel(base, 0.25, true, 0.0);
    let flash_px = evaluate_crt_pixel(base, 0.25, false, 1.0);
    assert_eq!(flash_px, Rgba::new(255, 255, 255, 255));
    assert_eq!(crt_px.a, 255);

    let mut pipe = Pipeline::new();
    pipe.trigger_flash(10);
    let f0 = pipe.step_frame();
    assert!((f0 - 1.0).abs() < 1e-6);
    for _ in 0..10 {
        pipe.step_frame();
    }
    assert_eq!(pipe.last_flash_intensity, 0.0);
}

#[test]
fn test_pokey_audio_synthesis_and_wav_generation() {
    let mut lfsr = PokeyLFSR::new();
    let mut pos4 = 0;
    let mut pos5 = 0;
    let mut pos17 = 0;
    for _ in 0..120 {
        if lfsr.step_poly4() > 0.0 {
            pos4 += 1;
        }
        if lfsr.step_poly5() > 0.0 {
            pos5 += 1;
        }
        if lfsr.step_poly17() > 0.0 {
            pos17 += 1;
        }
    }
    assert!((20..100).contains(&pos4));
    assert!((20..100).contains(&pos5));
    assert!((20..100).contains(&pos17));
    assert!((pokey_freq(31) - 1000.0).abs() < 1e-6);

    let bank = init_audio();
    let buffers = [
        ("launch", &bank.launch_pcm),
        ("explosion", &bank.explosion_pcm),
        ("tally", &bank.tally_pcm),
        ("siren", &bank.siren_pcm),
        ("silo_low", &bank.silo_low_pcm),
        ("cant_fire", &bank.cant_fire_pcm),
        ("bonus_city", &bank.bonus_city_pcm),
        ("game_over", &bank.game_over_pcm),
        ("bomber", &bank.bomber_pcm),
        ("satellite", &bank.sat_pcm),
        ("smart_bomb", &bank.smart_bomb_pcm),
    ];

    for (name, pcm) in buffers {
        assert!(!pcm.is_empty(), "{} PCM buffer is empty", name);
        assert_eq!(
            pcm.len() % 2,
            0,
            "{} PCM buffer must be 16-bit aligned",
            name
        );

        let mut peak: i32 = 0;
        for &chunk in pcm.as_chunks::<2>().0 {
            let sample = i16::from_le_bytes(chunk) as i32;
            peak = peak.max(sample.abs());
        }
        assert!(
            peak > 2000,
            "{} PCM buffer peak amplitude too low: {}",
            name,
            peak
        );

        let wav = pcm_to_wav(pcm);
        assert_eq!(&wav[0..4], b"RIFF");
        assert_eq!(&wav[8..12], b"WAVE");
        assert_eq!(&wav[12..16], b"fmt ");
        assert_eq!(&wav[36..40], b"data");
        let sr = u32::from_le_bytes([wav[24], wav[25], wav[26], wav[27]]) as usize;
        assert_eq!(sr, SAMPLE_RATE);
        assert_eq!(wav.len(), 44 + pcm.len());
    }

    // Verify adjust_volume clamping
    let _guard = TEST_MUTEX.lock().unwrap_or_else(|e| e.into_inner());
    set_master_volume(0.5);
    assert!((adjust_volume(0.25) - 0.75).abs() < 1e-6);
    assert!((adjust_volume(5.0) - 1.0).abs() < 1e-6);
    assert!((adjust_volume(-5.0) - 0.0).abs() < 1e-6);
    set_master_volume(1.0);
}

#[test]
fn test_embedded_window_icons() {
    let icons = get_window_icons();
    let expected_sizes = [16, 32, 48, 64, 128];
    assert_eq!(icons.len(), expected_sizes.len());

    for (icon, &expected) in icons.iter().zip(expected_sizes.iter()) {
        assert_eq!(icon.width, expected);
        assert_eq!(icon.height, expected);
        assert_eq!(icon.rgba.len(), (expected * expected * 4) as usize);
        let opaque = icon
            .rgba
            .as_chunks::<4>()
            .0
            .iter()
            .filter(|px| px[3] > 0)
            .count();
        assert!(
            opaque > 0,
            "Icon {}x{} has no opaque pixels",
            expected,
            expected
        );
    }

    let mq_icon = get_macroquad_icon().expect("macroquad icon should decode");
    assert_eq!(mq_icon.small.len(), 16 * 16 * 4);
    assert_eq!(mq_icon.medium.len(), 32 * 32 * 4);
    assert_eq!(mq_icon.big.len(), 64 * 64 * 4);
}

#[test]
fn test_city_placement_and_silo_separation() {
    let (_env, mut game) = IsolatedTestEnv::new("city_layout");
    game.start_game();

    let silo_mounds = [
        (18.0 - 13.0, 18.0 + 13.0),
        (128.0 - 13.0, 128.0 + 13.0),
        (238.0 - 13.0, 238.0 + 13.0),
    ];
    let city_width = 16.0;

    // Verify no city overlaps with any silo mound
    for (i, city) in game.cities.iter().enumerate() {
        let city_left = city.position.x;
        let city_right = city.position.x + city_width;
        for (s_idx, &(s_left, s_right)) in silo_mounds.iter().enumerate() {
            assert!(
                !(city_left < s_right && city_right > s_left),
                "City {} [{}, {}] overlaps with silo {} [{}, {}]",
                i,
                city_left,
                city_right,
                s_idx,
                s_left,
                s_right
            );
        }
    }

    // Verify exact X coordinates and symmetry across screen center (128.0)
    let expected_xs = [40.0, 65.0, 90.0, 150.0, 175.0, 200.0];
    for (i, &exp_x) in expected_xs.iter().enumerate() {
        assert!((game.cities[i].position.x - exp_x).abs() < 1e-6);
    }
    for i in 0..3 {
        let left_center = game.cities[i].position.x + city_width / 2.0;
        let right_center = game.cities[5 - i].position.x + city_width / 2.0;
        let left_dist = 128.0 - left_center;
        let right_dist = right_center - 128.0;
        assert!((left_dist - right_dist).abs() < 1e-6);
    }
}

#[test]
fn test_authentic_arcade_speed_progression() {
    let mut prev_speed = 0.0;
    for wave in 1..=25 {
        let w = get_wave_data(wave);
        assert!(
            w.speed <= 1.0,
            "Wave {} speed {} exceeds 1.0",
            wave,
            w.speed
        );
        assert!(
            w.speed >= prev_speed,
            "Wave {} speed {} decreased from {}",
            wave,
            w.speed,
            prev_speed
        );
        prev_speed = w.speed;
    }
}

#[test]
fn test_smart_bomb_simultaneous_cap() {
    let (_env, mut game) = IsolatedTestEnv::new("sb_cap");
    game.start_game();
    game.state = GameState::Playing;
    game.wave = 10;
    game.icbms_remaining = 0;
    game.smart_bombs_left = 5;
    game.spawn_cooldown = 1;

    // Pre-populate 3 smart bombs below threshold_y
    for &x in &[50.0, 100.0, 150.0] {
        game.smart_bombs.push(SmartBomb {
            start: Point::new(x, 0.0),
            curr: Point::new(x, 70.0),
            target: Point::new(x, 216.0),
            speed: 0.9,
            active: true,
            anim_tick: 0,
        });
    }

    game.update_spawning();
    assert_eq!(game.smart_bombs.len(), 3);
    assert_eq!(game.smart_bombs_left, 5);
}

#[test]
fn test_crosshair_and_input_clamping() {
    let (_env, mut game) = IsolatedTestEnv::new("clamp");
    game.start_game();
    game.state = GameState::Playing;

    let mut min_in = InputState::default();
    min_in.cursor_x = -5000;
    min_in.cursor_y = -5000;
    game.handle_crosshair_input(&min_in);
    assert!(game.crosshair.pos.x >= 4.0);
    assert!(game.crosshair.pos.y >= 8.0);

    let mut max_in = InputState::default();
    max_in.cursor_x = 5000;
    max_in.cursor_y = 5000;
    game.handle_crosshair_input(&max_in);
    assert!(game.crosshair.pos.x <= SIM_WIDTH - 4.0);
    assert!(game.crosshair.pos.y <= 206.0);
}

#[test]
fn test_missiles_per_silo_configuration_and_silo_drawing() {
    let (_env, mut game) = IsolatedTestEnv::new("silo_cfg");
    for &cap in &[10, 20, 50, 100] {
        game.settings.missiles_per_silo = cap;
        game.reset_silos();
        for bat in &game.batteries {
            assert_eq!(bat.max_ammo, cap);
            assert_eq!(bat.ammo, cap);
        }
    }

    let mut fb = Framebuffer::new(256, 231);
    for ammo in (0..=100).step_by(5) {
        let mut bat = game.batteries[1];
        bat.max_ammo = 100;
        bat.ammo = ammo;
        game.draw_silo(&mut fb, &bat);
    }
}

#[test]
fn test_arcade_octagonal_distance_and_120hz_input_accumulation() {
    // 1980 Atari ROM GENVEL octagonal distance: max(|dx|,|dy|) + 3/8 * min(|dx|,|dy|), capped at 255.0
    let start = Point::new(50.0, 14.0);
    let target = Point::new(130.0, 214.0); // |dx| = 80, |dy| = 200 -> 200 + 30 = 230
    assert!((start.arcade_distance(target) - 230.0).abs() < 1e-6);

    // Extreme corner-to-corner distance is capped at 255.0 just like 8-bit J in 6502 ROM
    let corner_a = Point::new(0.0, 0.0);
    let corner_b = Point::new(255.0, 216.0);
    assert!((corner_a.arcade_distance(corner_b) - 255.0).abs() < 1e-6);

    // Verify 120Hz sub-frame input merging preserves just_pressed until consumed by 60Hz tick
    let mut pending = InputState::default();
    let mut frame1 = InputState::default();
    frame1.mouse_left_just_pressed = true;
    frame1.key_space_just_pressed = true;
    frame1.cursor_x = 140;
    frame1.cursor_y = 90;
    pending.merge_frame(&frame1);

    // Second 120Hz frame before 60Hz step has no buttons pressed, but cursor moved
    let mut frame2 = InputState::default();
    frame2.cursor_x = 145;
    frame2.cursor_y = 92;
    frame2.key_left_held = true;
    pending.merge_frame(&frame2);

    assert!(pending.mouse_left_just_pressed);
    assert!(pending.key_space_just_pressed);
    assert!(pending.key_left_held);
    assert_eq!(pending.cursor_x, 145);
    assert_eq!(pending.cursor_y, 92);

    // Consuming the 60Hz tick clears edge triggers while keeping held state and cursor
    pending.clear_just_pressed();
    assert!(!pending.mouse_left_just_pressed);
    assert!(!pending.key_space_just_pressed);
    assert!(pending.key_left_held);
    assert_eq!(pending.cursor_x, 145);
}

#[test]
fn test_three_city_per_wave_protection_and_mirv_gating() {
    let (_env, mut game) = IsolatedTestEnv::new("three_city_rule");
    game.start_game();
    game.state = GameState::Playing;

    // Destroy 2 cities in this wave and have 1 active ICBM targeting city 2
    game.detonate_threat_on_ground(Point::new(CITY_X_POSITIONS[0] + 8.0, 216.0));
    game.detonate_threat_on_ground(Point::new(CITY_X_POSITIONS[1] + 8.0, 216.0));
    assert_eq!(game.cities_lost_this_wave, 2);

    let city2_target = Point::new(CITY_X_POSITIONS[2] + 8.0, 216.0);
    game.icbms.push(ICBM {
        start: Point::new(100.0, 0.0),
        curr: Point::new(100.0, 60.0),
        target: city2_target,
        speed: 0.3,
        progress: 0.3,
        is_splitter: false,
        splitted: false,
        split_altitude: 0.0,
        active: true,
    });

    // With 2 cities lost + 1 living city targeted = 3, untargeted living cities 3, 4, 5 must be spared!
    for _ in 0..50 {
        let picked = game.pick_random_target();
        for &protected_idx in &[3usize, 4, 5] {
            let protected_x = CITY_X_POSITIONS[protected_idx] + 8.0;
            assert!(
                (picked.x - protected_x).abs() > 1.0,
                "Untargeted living city {} at x={} should be protected by 3-city rule, but got target {:?}",
                protected_idx,
                protected_x,
                picked
            );
        }
    }

    // Verify MIRV split is blocked while another ICBM is above the top altitude gate (y < threshold_y)
    game.icbms.clear();
    game.explosions.clear();
    game.wave = 3;
    game.icbms_remaining = 5;
    // High ICBM at y = 10.0 (< threshold_y = 35.0)
    game.icbms.push(ICBM {
        start: Point::new(40.0, 0.0),
        curr: Point::new(40.0, 10.0),
        target: Point::new(40.0, 216.0),
        speed: 0.364,
        progress: 0.05,
        is_splitter: false,
        splitted: false,
        split_altitude: 0.0,
        active: true,
    });
    // Splitter ICBM in MIRV zone at y = 85.0
    game.icbms.push(ICBM {
        start: Point::new(120.0, 0.0),
        curr: Point::new(120.0, 85.0),
        target: Point::new(120.0, 216.0),
        speed: 0.364,
        progress: 0.4,
        is_splitter: true,
        splitted: false,
        split_altitude: 80.0,
        active: true,
    });

    game.update_icbms();
    // Should NOT split while the first ICBM is still above the altitude gate
    assert_eq!(game.icbms.len(), 2);
}
