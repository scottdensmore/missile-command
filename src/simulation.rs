use crate::audio::{
    play_cant_fire_sound, play_explosion_sound, play_game_over_sound, play_launch_sound,
    play_silo_low_sound, set_bomber_sound, set_satellite_sound, set_smart_bomb_sound,
    stop_all_continuous_sounds,
};
use crate::game::{Game, GameState};
use crate::math::{clamp, Point, SIM_WIDTH};
use crate::objects::{
    get_wave_data, Explosion, ExplosionState, Flier, FlierType, SmartBomb, ABM, ICBM,
};
use rand::RngExt;

impl Game {
    /// Runs one frame of entity spawning, movement, and collision detection.
    pub fn step_simulation(&mut self) {
        self.update_spawning();
        self.update_flier();
        self.update_abms();
        self.update_icbms();
        self.update_smart_bombs();
        self.update_explosions();
        self.check_collisions();
    }

    pub fn surviving_cities_count(&self) -> usize {
        self.cities.iter().filter(|c| !c.destroyed).count()
    }

    pub fn restore_continuous_sounds(&self) {
        if let Some(ref flier) = self.active_flier {
            if flier.active {
                if flier.flier_type == FlierType::Bomber {
                    set_bomber_sound(true);
                } else {
                    set_satellite_sound(true);
                }
            }
        }
        if self.smart_bombs.iter().any(|sb| sb.active) {
            set_smart_bomb_sound(true);
        }
    }

    pub fn fire_abm(&mut self, silo_idx: usize) {
        // Authentic arcade limit: Max 8 simultaneous ABMs on screen
        if self.abms.len() >= 8 || silo_idx >= self.batteries.len() {
            return;
        }

        let bat = &mut self.batteries[silo_idx];
        if bat.destroyed || bat.ammo <= 0 {
            play_cant_fire_sound();
            return;
        }

        bat.ammo -= 1;
        if bat.ammo <= 3 && !bat.low_warning_played {
            play_silo_low_sound();
            bat.low_warning_played = true;
        }

        // Center silo (Delta) fires at 7.5 units/frame; side silos fire at 3.2 units/frame
        let launch_speed = if silo_idx == 1 { 7.5 } else { 3.2 };
        let start = bat.position;

        self.abms.push(ABM {
            start,
            curr: start,
            target: self.crosshair.pos,
            silo_index: silo_idx,
            speed: launch_speed,
            progress: 0.0,
            active: true,
        });
        play_launch_sound();
    }

    pub fn update_spawning(&mut self) {
        // Limit simultaneous active attacks to 8 (Smart bombs count as 2)
        if self.icbms.len() + self.smart_bombs.len() * 2 >= 8 {
            return;
        }
        if self.icbms_remaining <= 0 && self.smart_bombs_left <= 0 {
            return;
        }

        // Authentic arcade altitude pacing: block new spawns while any threat is above threshold_y
        let threshold_y = (29.0 + (self.wave as f64) * 2.0).min(51.0);
        if self
            .icbms
            .iter()
            .any(|icbm| icbm.active && icbm.curr.y < threshold_y)
        {
            return;
        }
        if self
            .smart_bombs
            .iter()
            .any(|sb| sb.active && sb.curr.y < threshold_y)
        {
            return;
        }

        self.spawn_cooldown -= 1;
        if self.spawn_cooldown > 0 {
            return;
        }

        let wdata = get_wave_data(self.wave);
        let can_spawn_sb = self.smart_bombs_left > 0 && self.smart_bombs.len() < 3;
        let mut rng = rand::rng();

        if can_spawn_sb && (self.icbms_remaining <= 0 || rng.random::<f64>() < 0.35) {
            self.smart_bombs_left -= 1;
            let spawn_x = 20.0 + rng.random::<f64>() * (SIM_WIDTH - 40.0);
            let target = self.pick_random_target();
            self.smart_bombs.push(SmartBomb {
                start: Point::new(spawn_x, 0.0),
                curr: Point::new(spawn_x, 0.0),
                target,
                speed: wdata.speed,
                active: true,
                anim_tick: 0,
            });
            set_smart_bomb_sound(true);
        } else if self.icbms_remaining > 0 {
            self.icbms_remaining -= 1;
            let spawn_x = rng.random::<f64>() * SIM_WIDTH;
            let target = self.pick_random_target();
            let is_split =
                self.wave >= 2 && self.icbms_remaining >= 2 && rng.random::<f64>() < 0.25;
            let split_alt = 72.0 + rng.random::<f64>() * 31.0;

            self.icbms.push(ICBM {
                start: Point::new(spawn_x, 0.0),
                curr: Point::new(spawn_x, 0.0),
                target,
                speed: wdata.speed,
                progress: 0.0,
                is_splitter: is_split,
                splitted: false,
                split_altitude: split_alt,
                active: true,
            });
        }

        let min_cd = (55.0 - (self.wave as f64) * 2.5).max(20.0) as i32;
        let max_cd = (95.0 - (self.wave as f64) * 4.0).max(40.0) as i32;
        self.spawn_cooldown = min_cd + rng.random_range(0..=(max_cd - min_cd).max(0));
    }

    pub fn pick_random_target(&self) -> Point {
        // Authentic 1980 Atari ROM targeting (`GUICBM` at $5791 / A35820.1D):
        // Never allow more than 3 cities to be lost in a single wave.
        // Count how many surviving cities currently have an active missile or smart bomb targeting them.
        let mut targeted_living_cities = [false; 6];
        for (i, city) in self.cities.iter().enumerate() {
            if !city.destroyed {
                let cx = city.position.x + 8.0;
                let is_targeted = self
                    .icbms
                    .iter()
                    .any(|m| m.active && (m.target.x - cx).abs() < 4.0)
                    || self
                        .smart_bombs
                        .iter()
                        .any(|sb| sb.active && (sb.target.x - cx).abs() < 4.0);
                targeted_living_cities[i] = is_targeted;
            }
        }

        let threatened_or_lost =
            self.cities_lost_this_wave + targeted_living_cities.iter().filter(|&&t| t).count();

        let mut targets = Vec::with_capacity(9);
        if threatened_or_lost < 3 {
            // Prefer untargeted living cities, living batteries, or already-targeted living cities
            for (i, city) in self.cities.iter().enumerate() {
                if !city.destroyed && !targeted_living_cities[i] {
                    targets.push(Point::new(city.position.x + 8.0, 216.0));
                }
            }
            for bat in &self.batteries {
                if !bat.destroyed {
                    targets.push(bat.position);
                }
            }
            if targets.is_empty() {
                for city in &self.cities {
                    if !city.destroyed {
                        targets.push(Point::new(city.position.x + 8.0, 216.0));
                    }
                }
            }
        } else {
            // 3 cities already lost/threatened this wave: protect remaining untargeted living cities!
            // Target surviving batteries, already-targeted living cities, or destroyed city ruins (`NOTLCI`).
            for bat in &self.batteries {
                if !bat.destroyed {
                    targets.push(bat.position);
                }
            }
            for (i, city) in self.cities.iter().enumerate() {
                if city.destroyed || targeted_living_cities[i] {
                    targets.push(Point::new(city.position.x + 8.0, 216.0));
                }
            }
        }

        let mut rng = rand::rng();
        if !targets.is_empty() {
            targets[rng.random_range(0..targets.len())]
        } else {
            Point::new(20.0 + rng.random::<f64>() * (SIM_WIDTH - 40.0), 216.0)
        }
    }

    pub fn update_flier(&mut self) {
        let mut rng = rand::rng();
        if self.active_flier.is_none() {
            if !self.flier_spawned_this_wave && self.icbms_remaining > 2 {
                self.flier_cooldown -= 1;
                if self.flier_cooldown <= 0 {
                    self.flier_spawned_this_wave = true;
                    let (ft, alt, mut speed) = if rng.random::<f64>() < 0.45 {
                        (
                            FlierType::Satellite,
                            28.0 + rng.random::<f64>() * 18.0,
                            0.50,
                        )
                    } else {
                        (
                            FlierType::Bomber,
                            65.0 + rng.random::<f64>() * 25.0,
                            1.0 / 3.0,
                        )
                    };

                    let start_x = if rng.random::<f64>() < 0.5 {
                        -16.0
                    } else {
                        speed = -speed;
                        SIM_WIDTH + 16.0
                    };

                    self.active_flier = Some(Flier {
                        flier_type: ft,
                        x: start_x,
                        y: alt,
                        speed,
                        drop_cooldown: 40 + rng.random_range(0..50),
                        bombs_remaining: 2,
                        active: true,
                        anim_tick: 0,
                    });

                    if ft == FlierType::Bomber {
                        set_bomber_sound(true);
                    } else {
                        set_satellite_sound(true);
                    }
                }
            }
            return;
        }

        let target = self.pick_random_target();
        let wdata = get_wave_data(self.wave);
        let can_drop =
            self.icbms_remaining > 0 && (self.icbms.len() + self.smart_bombs.len() * 2) < 8;
        let mut spawned_bomb = None;
        let mut left_screen = false;

        if let Some(ref mut flier) = self.active_flier {
            flier.anim_tick += 1;
            flier.x += flier.speed;

            if (flier.speed > 0.0 && flier.x > SIM_WIDTH + 20.0)
                || (flier.speed < 0.0 && flier.x < -20.0)
            {
                flier.active = false;
                left_screen = true;
            } else if flier.bombs_remaining > 0 && flier.x >= 15.0 && flier.x <= SIM_WIDTH - 15.0 {
                flier.drop_cooldown -= 1;
                if flier.drop_cooldown <= 0 && can_drop {
                    flier.bombs_remaining -= 1;
                    flier.drop_cooldown = 60 + rng.random_range(0..40);
                    spawned_bomb = Some(Point::new(flier.x, flier.y));
                }
            }
        }

        if left_screen {
            self.active_flier = None;
            set_bomber_sound(false);
            set_satellite_sound(false);
            return;
        }

        if let Some(pos) = spawned_bomb {
            if self.icbms_remaining > 0 {
                self.icbms_remaining -= 1;
            }
            self.icbms.push(ICBM {
                start: pos,
                curr: pos,
                target,
                speed: wdata.speed,
                progress: 0.0,
                is_splitter: false,
                splitted: false,
                split_altitude: 0.0,
                active: true,
            });
        }
    }

    pub fn update_abms(&mut self) {
        let mut active = Vec::with_capacity(self.abms.len());
        let mut detonations = Vec::new();

        for mut abm in self.abms.drain(..) {
            if !abm.active {
                continue;
            }
            let dist = abm.start.distance(abm.target);
            if dist <= 0.0 {
                continue;
            }

            abm.progress += abm.speed / dist;
            if abm.progress >= 1.0 {
                detonations.push(abm.target);
            } else {
                abm.curr = abm.start.lerp(abm.target, abm.progress);
                active.push(abm);
            }
        }

        self.abms = active;
        for target in detonations {
            self.spawn_explosion(target, 22.0, 18);
        }
    }

    pub fn update_icbms(&mut self) {
        let mut active = Vec::with_capacity(self.icbms.len() + 2);
        let mut ground_hits = Vec::new();
        let old_icbms = std::mem::take(&mut self.icbms);

        // Authentic 1980 Atari ROM MIRV gating (`UPICBM` / `ICNORM` / `MIRVER` in A35820.1D):
        // - No MIRVs if >= 12 explosions on screen (`CPY I,12.`)
        // - No MIRVs if any threat is above the wave altitude threshold (`HIICBV`)
        // - At most 1 ICBM splits per frame (`MIRVIX`), respecting the 8-threat screen limit (`MXICON`)
        let threshold_y = (29.0 + (self.wave as f64) * 2.0).min(51.0);
        let any_above_gate = old_icbms.iter().any(|m| m.active && m.curr.y < threshold_y)
            || self
                .smart_bombs
                .iter()
                .any(|sb| sb.active && sb.curr.y < threshold_y);
        let live_icbm_count = old_icbms.iter().filter(|m| m.active).count();
        let mut threat_slots_used = live_icbm_count + self.smart_bombs.len() * 2;
        let mut mirved_this_frame = false;

        for mut icbm in old_icbms {
            if !icbm.active {
                continue;
            }

            // MIRV Split Logic (`MIRVLO`..`MIRVHI`): sister warheads count toward wave missile quota
            if icbm.is_splitter
                && !icbm.splitted
                && !mirved_this_frame
                && !any_above_gate
                && self.explosions.len() < 12
                && icbm.curr.y >= icbm.split_altitude
                && icbm.curr.y <= 115.0
                && threat_slots_used < 8
            {
                icbm.splitted = true;
                mirved_this_frame = true;
                let max_by_slots = (8 - threat_slots_used).min(2) as i32;
                let sisters_to_spawn = self.icbms_remaining.clamp(0, max_by_slots);
                self.icbms_remaining -= sisters_to_spawn;
                threat_slots_used += sisters_to_spawn as usize;

                for _ in 0..sisters_to_spawn {
                    let sister_target = self.pick_random_target();
                    active.push(ICBM {
                        start: icbm.curr,
                        curr: icbm.curr,
                        target: sister_target,
                        speed: icbm.speed,
                        progress: 0.0,
                        is_splitter: false,
                        splitted: false,
                        split_altitude: 0.0,
                        active: true,
                    });
                }
            }

            // Use authentic 1980 Atari ROM octagonal step count (`GENVEL`: max(|dx|,|dy|) + 3/8 min(|dx|,|dy|))
            let dist = icbm.start.arcade_distance(icbm.target);
            if dist <= 0.0 {
                continue;
            }

            icbm.progress += icbm.speed / dist;
            if icbm.progress >= 1.0 || icbm.curr.y >= 216.0 {
                ground_hits.push(icbm.target);
            } else {
                icbm.curr = icbm.start.lerp(icbm.target, icbm.progress);
                active.push(icbm);
            }
        }

        self.icbms = active;
        for hit in ground_hits {
            self.detonate_threat_on_ground(hit);
        }
    }

    pub fn update_smart_bombs(&mut self) {
        let mut active = Vec::with_capacity(self.smart_bombs.len());
        let mut ground_hits = Vec::new();
        let old_sbs = std::mem::take(&mut self.smart_bombs);
        let mut rng = rand::rng();

        for mut sb in old_sbs {
            if !sb.active {
                continue;
            }
            sb.anim_tick += 1;

            // Smart Bomb Evasion AI: scan for nearby expanding explosions
            let mut evading = false;
            for exp in &self.explosions {
                if exp.state == ExplosionState::Dead {
                    continue;
                }
                if sb.curr.distance(exp.center) < exp.radius + 18.0 {
                    evading = true;
                    let mut dx = sb.curr.x - exp.center.x;
                    let dy = sb.curr.y - exp.center.y;
                    if dx.abs() < 0.1 {
                        dx = if rng.random::<f64>() < 0.5 { 1.0 } else { -1.0 };
                    }
                    let length = dx.hypot(dy);
                    if length > 0.0 {
                        sb.curr.x += (dx / length) * sb.speed * 1.25;
                        sb.curr.y += sb.speed * 0.40;
                    }
                    break;
                }
            }

            if !evading {
                let dist = sb.curr.arcade_distance(sb.target);
                if dist <= sb.speed {
                    ground_hits.push(sb.target);
                    continue;
                }
                sb.curr.x += ((sb.target.x - sb.curr.x) / dist) * sb.speed;
                sb.curr.y += ((sb.target.y - sb.curr.y) / dist) * sb.speed;
            }

            sb.curr.x = clamp(sb.curr.x, 4.0, SIM_WIDTH - 4.0);
            if sb.curr.y >= 216.0 {
                ground_hits.push(sb.curr);
            } else {
                active.push(sb);
            }
        }

        self.smart_bombs = active;
        for hit in ground_hits {
            self.detonate_threat_on_ground(hit);
        }

        if self.smart_bombs.is_empty() {
            set_smart_bomb_sound(false);
        }
    }

    pub fn spawn_explosion(&mut self, center: Point, max_radius: f64, hold_timer: i32) {
        self.explosions.push(Explosion {
            center,
            radius: 0.0,
            max_radius,
            hold_timer,
            state: ExplosionState::Expanding,
            tick: 0,
        });
        play_explosion_sound();
    }

    pub fn detonate_threat_on_ground(&mut self, pos: Point) {
        for city in &mut self.cities {
            if !city.destroyed {
                let cx = city.position.x + 8.0;
                if (pos.x - cx).abs() < 12.0 && (pos.y - 216.0).abs() < 8.0 {
                    city.destroyed = true;
                    self.cities_lost_this_wave += 1;
                    self.pipeline.trigger_flash(8);
                    break;
                }
            }
        }

        for bat in &mut self.batteries {
            if !bat.destroyed {
                let bx = bat.position.x;
                if (pos.x - bx).abs() < 14.0 && (pos.y - 214.0).abs() < 8.0 {
                    bat.destroyed = true;
                    bat.ammo = 0;
                    self.pipeline.trigger_flash(6);
                    break;
                }
            }
        }

        self.spawn_explosion(pos, 18.0, 14);
    }

    pub fn update_explosions(&mut self) {
        // Authentic 1980 Atari ROM explosion timing (`PREXPL` in A35820.1D):
        // Each explosion updates across 27 radius steps every 5th frame (~135 frames = 2.25s total).
        for exp in &mut self.explosions {
            if exp.state == ExplosionState::Dead {
                continue;
            }
            exp.tick += 1;

            match exp.state {
                ExplosionState::Expanding => {
                    exp.radius += 0.38;
                    if exp.radius >= exp.max_radius {
                        exp.state = ExplosionState::Holding;
                    }
                }
                ExplosionState::Holding => {
                    exp.hold_timer -= 1;
                    if exp.hold_timer <= 0 {
                        exp.state = ExplosionState::Contracting;
                    }
                }
                ExplosionState::Contracting => {
                    exp.radius -= 0.36;
                    if exp.radius <= 0.0 {
                        exp.state = ExplosionState::Dead;
                    }
                }
                ExplosionState::Dead => {}
            }
        }
        self.explosions
            .retain(|exp| exp.state != ExplosionState::Dead);
    }

    fn is_hit_by_explosion(explosions: &[Explosion], pos: Point, extra_radius: f64) -> bool {
        explosions.iter().any(|exp| {
            exp.state != ExplosionState::Dead
                && pos.distance(exp.center) <= exp.radius + extra_radius
        })
    }

    pub fn check_collisions(&mut self) {
        let mut chain_explosions = Vec::new();
        let mult = self.palette.multiplier;

        for icbm in &mut self.icbms {
            if icbm.active && Self::is_hit_by_explosion(&self.explosions, icbm.curr, 0.0) {
                icbm.active = false;
                self.score += 25 * mult;
                chain_explosions.push((icbm.curr, 18.0, 14));
            }
        }

        for sb in &mut self.smart_bombs {
            if sb.active && Self::is_hit_by_explosion(&self.explosions, sb.curr, 0.0) {
                sb.active = false;
                self.score += 125 * mult;
                chain_explosions.push((sb.curr, 18.0, 14));
            }
        }

        let mut flier_hit = None;
        if let Some(ref flier) = self.active_flier {
            if flier.active {
                let f_pos = Point::new(flier.x + 8.0, flier.y + 4.0);
                if Self::is_hit_by_explosion(&self.explosions, f_pos, 6.0) {
                    flier_hit = Some(f_pos);
                }
            }
        }
        if let Some(f_pos) = flier_hit {
            self.active_flier = None;
            set_bomber_sound(false);
            set_satellite_sound(false);
            self.score += 100 * mult;
            chain_explosions.push((f_pos, 22.0, 16));
        }

        for (center, max_r, hold) in chain_explosions {
            self.spawn_explosion(center, max_r, hold);
        }
    }

    pub fn check_wave_completion(&mut self) {
        if self.surviving_cities_count() == 0 {
            stop_all_continuous_sounds();
            self.state = GameState::TheEnd;
            self.the_end_timer = 190;
            self.pipeline.trigger_flash(25);
            play_game_over_sound();
            return;
        }

        let no_threats_remaining = self.icbms_remaining <= 0
            && self.smart_bombs_left <= 0
            && self.icbms.is_empty()
            && self.smart_bombs.is_empty()
            && self.active_flier.is_none();

        if no_threats_remaining {
            self.wave_end_grace_timer += 1;
            if (self.abms.is_empty() && self.explosions.is_empty())
                || self.wave_end_grace_timer >= 45
            {
                stop_all_continuous_sounds();
                self.state = GameState::Tally;
                self.tally_phase = 0;
                self.tally_timer = 30;
                self.tally_battery_idx = 0;
                self.tally_city_idx = 0;
                self.tally_score_earned = 0;
                self.wave_end_grace_timer = 0;
            }
        } else {
            self.wave_end_grace_timer = 0;
        }
    }
}
