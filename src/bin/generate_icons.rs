use image::imageops::FilterType;
use image::{DynamicImage, ImageFormat};
use std::collections::HashMap;
use std::fs;
use std::io::Cursor;
use std::path::Path;
use std::process::Command;

fn encode_png(img: &DynamicImage) -> Result<Vec<u8>, String> {
    let mut buf = Vec::new();
    img.write_to(&mut Cursor::new(&mut buf), ImageFormat::Png)
        .map_err(|e| e.to_string())?;
    Ok(buf)
}

fn save_png(img: &DynamicImage, path: impl AsRef<Path>) -> Result<(), String> {
    let path = path.as_ref();
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent).map_err(|e| e.to_string())?;
    }
    img.save_with_format(path, ImageFormat::Png)
        .map_err(|e| e.to_string())
}

fn pack_ico(png_images: &[(u32, Vec<u8>)], output_path: impl AsRef<Path>) -> Result<(), String> {
    let count = png_images.len() as u16;
    let mut out = Vec::new();

    // ICO Header (6 bytes)
    out.extend_from_slice(&0u16.to_le_bytes());
    out.extend_from_slice(&1u16.to_le_bytes());
    out.extend_from_slice(&count.to_le_bytes());

    let mut offset = 6u32 + 16u32 * (count as u32);
    for (sz, data) in png_images {
        let dim = if *sz >= 256 { 0u8 } else { *sz as u8 };
        let data_len = data.len() as u32;

        out.push(dim); // width
        out.push(dim); // height
        out.push(0); // color count
        out.push(0); // reserved
        out.extend_from_slice(&1u16.to_le_bytes()); // color planes
        out.extend_from_slice(&32u16.to_le_bytes()); // bits per pixel
        out.extend_from_slice(&data_len.to_le_bytes());
        out.extend_from_slice(&offset.to_le_bytes());

        offset += data_len;
    }

    for (_, data) in png_images {
        out.extend_from_slice(data);
    }

    if let Some(parent) = output_path.as_ref().parent() {
        fs::create_dir_all(parent).map_err(|e| e.to_string())?;
    }
    fs::write(output_path, out).map_err(|e| e.to_string())
}

fn main() {
    let master_path = "assets/icon-1024.png";
    let master_img = match image::open(master_path) {
        Ok(img) => img,
        Err(e) => {
            eprintln!("Error opening {}: {}", master_path, e);
            std::process::exit(1);
        }
    };

    println!(
        "Generating multi-platform icon assets from {}...",
        master_path
    );

    let standard_sizes: [u32; 8] = [16, 32, 48, 64, 128, 256, 512, 1024];
    let mut scaled_map: HashMap<u32, DynamicImage> = HashMap::new();

    for &sz in &standard_sizes {
        let img = if sz == 1024 {
            master_img.clone()
        } else {
            master_img.resize_exact(sz, sz, FilterType::Lanczos3)
        };
        let out_png = format!("assets/icons/icon-{}.png", sz);
        if let Err(e) = save_png(&img, &out_png) {
            eprintln!("Failed to save {}: {}", out_png, e);
        } else {
            println!("  Generated {}", out_png);
        }
        scaled_map.insert(sz, img);
    }

    // macOS .icns via iconutil
    let iconset_dir = Path::new("assets/icon.iconset");
    if fs::create_dir_all(iconset_dir).is_ok() {
        let apple_sizes = [
            ("icon_16x16.png", 16u32),
            ("icon_16x16@2x.png", 32),
            ("icon_32x32.png", 32),
            ("icon_32x32@2x.png", 64),
            ("icon_128x128.png", 128),
            ("icon_128x128@2x.png", 256),
            ("icon_256x256.png", 256),
            ("icon_256x256@2x.png", 512),
            ("icon_512x512.png", 512),
            ("icon_512x512@2x.png", 1024),
        ];

        for (filename, sz) in apple_sizes {
            if let Some(img) = scaled_map.get(&sz) {
                let _ = save_png(img, iconset_dir.join(filename));
            }
        }

        if let Ok(output) = Command::new("iconutil")
            .args([
                "-c",
                "icns",
                "assets/icon.iconset",
                "-o",
                "assets/icon.icns",
            ])
            .output()
        {
            if output.status.success() {
                println!("  Generated assets/icon.icns");
            }
        }
        let _ = fs::remove_dir_all(iconset_dir);
    }

    // Windows Multi-Resolution Icon (.ico)
    let ico_sizes: [u32; 6] = [16, 32, 48, 64, 128, 256];
    let mut ico_payloads = Vec::new();
    for &sz in &ico_sizes {
        if let Some(img) = scaled_map.get(&sz) {
            if let Ok(bytes) = encode_png(img) {
                ico_payloads.push((sz, bytes));
            }
        }
    }
    if let Err(e) = pack_ico(&ico_payloads, "assets/icon.ico") {
        eprintln!("Failed to generate assets/icon.ico: {}", e);
    } else {
        println!("  Generated assets/icon.ico");
    }

    // Web Favicons & PWA Icons
    let mut fav_payloads = Vec::new();
    for &sz in &[16u32, 32, 48] {
        if let Some(img) = scaled_map.get(&sz) {
            if let Ok(bytes) = encode_png(img) {
                fav_payloads.push((sz, bytes));
            }
        }
    }
    let _ = pack_ico(&fav_payloads, "web/favicon.ico");
    if let Some(img32) = scaled_map.get(&32) {
        let _ = save_png(img32, "web/favicon-light.png");
        let _ = save_png(img32, "web/favicon-dark.png");
    }
    let touch_img = master_img.resize_exact(180, 180, FilterType::Lanczos3);
    let _ = save_png(&touch_img, "web/apple-touch-icon.png");
    let pwa_192 = master_img.resize_exact(192, 192, FilterType::Lanczos3);
    let _ = save_png(&pwa_192, "web/icon-192.png");
    if let Some(img512) = scaled_map.get(&512) {
        let _ = save_png(img512, "web/icon-512.png");
    }

    println!("All icon assets generated successfully!");
}
