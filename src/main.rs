use macroquad::audio::{
    load_sound_from_bytes, play_sound, set_sound_volume, stop_sound, PlaySoundParams, Sound,
};
use macroquad::prelude::*;
use missile_command::assets::get_macroquad_icon;
use missile_command::audio::{drain_audio_events, init_audio, pcm_to_wav, SoundEffect};
use missile_command::game::Game;
use missile_command::input::InputState;
use missile_command::math::{SIM_HEIGHT, SIM_WIDTH};
use missile_command::pipeline::{compute_letterbox, CRT_FRAGMENT_SHADER, CRT_VERTEX_SHADER};

fn window_conf() -> Conf {
    Conf {
        window_title: "Missile Command (1980 Atari Arcade Replica)".to_owned(),
        window_width: 1024,
        window_height: 924,
        window_resizable: true,
        high_dpi: false,
        icon: get_macroquad_icon(),
        ..Default::default()
    }
}

struct RuntimeAudio {
    launch: Option<Sound>,
    explosion: Option<Sound>,
    tally: Option<Sound>,
    siren: Option<Sound>,
    silo_low: Option<Sound>,
    cant_fire: Option<Sound>,
    bonus_city: Option<Sound>,
    game_over: Option<Sound>,
    bomber: Option<Sound>,
    satellite: Option<Sound>,
    smart_bomb: Option<Sound>,
    bomber_playing: bool,
    satellite_playing: bool,
    smart_bomb_playing: bool,
}

impl RuntimeAudio {
    async fn init() -> Self {
        let bank = init_audio();
        Self {
            launch: load_sound_from_bytes(&pcm_to_wav(&bank.launch_pcm))
                .await
                .ok(),
            explosion: load_sound_from_bytes(&pcm_to_wav(&bank.explosion_pcm))
                .await
                .ok(),
            tally: load_sound_from_bytes(&pcm_to_wav(&bank.tally_pcm))
                .await
                .ok(),
            siren: load_sound_from_bytes(&pcm_to_wav(&bank.siren_pcm))
                .await
                .ok(),
            silo_low: load_sound_from_bytes(&pcm_to_wav(&bank.silo_low_pcm))
                .await
                .ok(),
            cant_fire: load_sound_from_bytes(&pcm_to_wav(&bank.cant_fire_pcm))
                .await
                .ok(),
            bonus_city: load_sound_from_bytes(&pcm_to_wav(&bank.bonus_city_pcm))
                .await
                .ok(),
            game_over: load_sound_from_bytes(&pcm_to_wav(&bank.game_over_pcm))
                .await
                .ok(),
            bomber: load_sound_from_bytes(&pcm_to_wav(&bank.bomber_pcm))
                .await
                .ok(),
            satellite: load_sound_from_bytes(&pcm_to_wav(&bank.sat_pcm)).await.ok(),
            smart_bomb: load_sound_from_bytes(&pcm_to_wav(&bank.smart_bomb_pcm))
                .await
                .ok(),
            bomber_playing: false,
            satellite_playing: false,
            smart_bomb_playing: false,
        }
    }

    fn sync(&mut self) {
        let (events, looping) = drain_audio_events();
        for (effect, volume) in events {
            let snd = match effect {
                SoundEffect::Launch => self.launch.as_ref(),
                SoundEffect::Explosion => self.explosion.as_ref(),
                SoundEffect::Tally => self.tally.as_ref(),
                SoundEffect::Siren => self.siren.as_ref(),
                SoundEffect::SiloLow => self.silo_low.as_ref(),
                SoundEffect::CantFire => self.cant_fire.as_ref(),
                SoundEffect::BonusCity => self.bonus_city.as_ref(),
                SoundEffect::GameOver => self.game_over.as_ref(),
            };
            if let Some(s) = snd {
                play_sound(
                    s,
                    PlaySoundParams {
                        looped: false,
                        volume,
                    },
                );
            }
        }

        Self::sync_loop(
            self.bomber.as_ref(),
            &mut self.bomber_playing,
            looping.bomber_active,
            0.16 * looping.volume_scale,
        );
        Self::sync_loop(
            self.satellite.as_ref(),
            &mut self.satellite_playing,
            looping.satellite_active,
            0.14 * looping.volume_scale,
        );
        Self::sync_loop(
            self.smart_bomb.as_ref(),
            &mut self.smart_bomb_playing,
            looping.smart_bomb_active,
            0.18 * looping.volume_scale,
        );
    }

    fn sync_loop(sound: Option<&Sound>, is_playing: &mut bool, should_play: bool, volume: f32) {
        let Some(s) = sound else { return };
        if should_play {
            if !*is_playing {
                play_sound(
                    s,
                    PlaySoundParams {
                        looped: true,
                        volume,
                    },
                );
                *is_playing = true;
            } else {
                set_sound_volume(s, volume);
            }
        } else if *is_playing {
            stop_sound(s);
            *is_playing = false;
        }
    }
}

fn collect_input() -> InputState {
    let (mx, my) = mouse_position();
    let mut typed_letters = [false; 26];
    let letter_keys = [
        KeyCode::A,
        KeyCode::B,
        KeyCode::C,
        KeyCode::D,
        KeyCode::E,
        KeyCode::F,
        KeyCode::G,
        KeyCode::H,
        KeyCode::I,
        KeyCode::J,
        KeyCode::K,
        KeyCode::L,
        KeyCode::M,
        KeyCode::N,
        KeyCode::O,
        KeyCode::P,
        KeyCode::Q,
        KeyCode::R,
        KeyCode::S,
        KeyCode::T,
        KeyCode::U,
        KeyCode::V,
        KeyCode::W,
        KeyCode::X,
        KeyCode::Y,
        KeyCode::Z,
    ];
    for (i, &k) in letter_keys.iter().enumerate() {
        typed_letters[i] = is_key_pressed(k);
    }

    InputState {
        cursor_x: mx.round() as i32,
        cursor_y: my.round() as i32,
        mouse_left_just_pressed: is_mouse_button_pressed(MouseButton::Left),
        mouse_middle_just_pressed: is_mouse_button_pressed(MouseButton::Middle),
        mouse_right_just_pressed: is_mouse_button_pressed(MouseButton::Right),

        key_left_held: is_key_down(KeyCode::Left),
        key_right_held: is_key_down(KeyCode::Right),
        key_up_held: is_key_down(KeyCode::Up),
        key_down_held: is_key_down(KeyCode::Down),
        key_alt_held: is_key_down(KeyCode::LeftAlt) || is_key_down(KeyCode::RightAlt),

        key_left_just_pressed: is_key_pressed(KeyCode::Left),
        key_right_just_pressed: is_key_pressed(KeyCode::Right),
        key_up_just_pressed: is_key_pressed(KeyCode::Up),
        key_down_just_pressed: is_key_pressed(KeyCode::Down),
        key_w_just_pressed: is_key_pressed(KeyCode::W),
        key_a_just_pressed: is_key_pressed(KeyCode::A),
        key_s_just_pressed: is_key_pressed(KeyCode::S),
        key_d_just_pressed: is_key_pressed(KeyCode::D),
        key_1_just_pressed: is_key_pressed(KeyCode::Key1) || is_key_pressed(KeyCode::Kp1),
        key_2_just_pressed: is_key_pressed(KeyCode::Key2) || is_key_pressed(KeyCode::Kp2),
        key_3_just_pressed: is_key_pressed(KeyCode::Key3) || is_key_pressed(KeyCode::Kp3),
        key_space_just_pressed: is_key_pressed(KeyCode::Space),
        key_enter_just_pressed: is_key_pressed(KeyCode::Enter) || is_key_pressed(KeyCode::KpEnter),
        key_escape_just_pressed: is_key_pressed(KeyCode::Escape),
        key_f1_just_pressed: is_key_pressed(KeyCode::F1),
        key_tab_just_pressed: is_key_pressed(KeyCode::Tab),
        key_f11_just_pressed: is_key_pressed(KeyCode::F11),
        key_m_just_pressed: is_key_pressed(KeyCode::M),
        key_minus_just_pressed: is_key_pressed(KeyCode::Minus)
            || is_key_pressed(KeyCode::KpSubtract),
        key_equal_just_pressed: is_key_pressed(KeyCode::Equal) || is_key_pressed(KeyCode::KpAdd),
        key_o_just_pressed: is_key_pressed(KeyCode::O),
        key_p_just_pressed: is_key_pressed(KeyCode::P),
        key_q_just_pressed: is_key_pressed(KeyCode::Q),

        typed_letters_just_pressed: typed_letters,
        gamepads: Vec::new(),
    }
}

#[macroquad::main(window_conf)]
async fn main() {
    let mut game = Game::new();
    let mut runtime_audio = RuntimeAudio::init().await;

    let crt_material = load_material(
        ShaderSource::Glsl {
            vertex: CRT_VERTEX_SHADER,
            fragment: CRT_FRAGMENT_SHADER,
        },
        MaterialParams {
            uniforms: vec![
                UniformDesc::new("Resolution", UniformType::Float2),
                UniformDesc::new("FlashIntensity", UniformType::Float1),
            ],
            ..Default::default()
        },
    )
    .ok();

    let mut fb_image = Image {
        bytes: game.pipeline.frame_buffer.pixels.clone(),
        width: SIM_WIDTH as u16,
        height: SIM_HEIGHT as u16,
    };
    let texture = Texture2D::from_image(&fb_image);
    texture.set_filter(FilterMode::Nearest);

    let mut prev_captured = false;
    let mut prev_fullscreen = false;
    let mut accumulator = 0.0f64;
    let target_dt = 1.0 / 60.0;
    let mut pending_input = InputState::default();

    loop {
        let win_w = screen_width().round() as i32;
        let win_h = screen_height().round() as i32;
        game.layout(win_w, win_h);

        let dt = (get_frame_time() as f64).min(0.1);
        accumulator += dt;

        // Accumulate per-render-frame input so 120Hz+ displays step strictly at 60Hz
        // without dropping edge-triggered key/mouse presses between simulation ticks.
        let frame_input = collect_input();
        pending_input.merge_frame(&frame_input);

        while accumulator >= target_dt && !game.should_quit {
            game.update_with_input(&pending_input);
            pending_input.clear_just_pressed();
            accumulator -= target_dt;
        }

        if game.should_quit {
            break;
        }

        if game.mouse_captured != prev_captured {
            show_mouse(!game.mouse_captured);
            prev_captured = game.mouse_captured;
        }
        if game.fullscreen != prev_fullscreen {
            set_fullscreen(game.fullscreen);
            prev_fullscreen = game.fullscreen;
        }

        runtime_audio.sync();
        game.draw();

        fb_image
            .bytes
            .copy_from_slice(&game.pipeline.frame_buffer.pixels);
        texture.update(&fb_image);

        clear_background(BLACK);
        let lb = compute_letterbox(win_w, win_h);
        let flash = game.pipeline.last_flash_intensity;

        if game.pipeline.use_crt {
            if let Some(ref mat) = crt_material {
                mat.set_uniform("Resolution", (SIM_WIDTH as f32, SIM_HEIGHT as f32));
                mat.set_uniform("FlashIntensity", flash);
                gl_use_material(mat);
            }
        }

        let draw_color = if !game.pipeline.use_crt && flash > 0.0 {
            Color::new(1.0 + flash, 1.0 + flash, 1.0 + flash, 1.0)
        } else {
            WHITE
        };

        draw_texture_ex(
            &texture,
            lb.offset_x as f32,
            lb.offset_y as f32,
            draw_color,
            DrawTextureParams {
                dest_size: Some(vec2(lb.render_w as f32, lb.render_h as f32)),
                ..Default::default()
            },
        );

        if game.pipeline.use_crt && crt_material.is_some() {
            gl_use_default_material();
        }

        next_frame().await;
    }
}
