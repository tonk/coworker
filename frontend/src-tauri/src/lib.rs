use std::io::Cursor;
use std::path::PathBuf;
use std::sync::atomic::{AtomicBool, Ordering};

static CLOSE_TO_TRAY_ENABLED: AtomicBool = AtomicBool::new(true);

use serde::{Deserialize, Serialize};
use tauri::image::Image;
use tauri::menu::{MenuBuilder, MenuItemBuilder};
use tauri::tray::TrayIconBuilder;
use tauri::Manager;

fn png_to_image(bytes: &[u8]) -> Result<Image<'static>, String> {
    let decoder = png::Decoder::new(Cursor::new(bytes));
    let mut reader = decoder.read_info().map_err(|e| e.to_string())?;
    let buf_size = reader.output_buffer_size().unwrap_or(0);
    let mut buf = vec![0u8; buf_size];
    let info = reader.next_frame(&mut buf).map_err(|e| e.to_string())?;
    let rgba = match info.color_type {
        png::ColorType::Rgba => buf[..info.buffer_size()].to_vec(),
        png::ColorType::Rgb => {
            let mut rgba_buf = Vec::with_capacity(info.buffer_size() / 3 * 4);
            for chunk in buf[..info.buffer_size()].chunks(3) {
                rgba_buf.extend_from_slice(chunk);
                rgba_buf.push(255);
            }
            rgba_buf
        }
        _ => return Err("unsupported PNG color type".to_string()),
    };
    Ok(Image::new_owned(rgba, info.width, info.height))
}

// ---------------------------------------------------------------------------
// Profile types
// ---------------------------------------------------------------------------

#[derive(Serialize, Deserialize, Clone, Debug)]
struct Profile {
    name: String,
    label: String,
}

#[derive(Serialize, Deserialize)]
struct ProfilesConfig {
    default: String,
    profiles: Vec<Profile>,
}

#[derive(Serialize, Clone)]
struct ProfileInfo {
    name: String,
    label: String,
    is_default: bool,
}

// ---------------------------------------------------------------------------
// Platform config / data directories
// (computed before the Tauri app handle is available, e.g. for --list-profiles)
// ---------------------------------------------------------------------------

const APP_ID: &str = "com.warmdesk.desktop";

#[cfg(any(target_os = "linux", target_os = "macos"))]
fn platform_home() -> PathBuf {
    std::env::var("HOME")
        .or_else(|_| std::env::var("USERPROFILE"))
        .map(PathBuf::from)
        .unwrap_or_else(|_| PathBuf::from("/"))
}

fn warmdesk_config_dir() -> PathBuf {
    #[cfg(target_os = "linux")]
    {
        let base = std::env::var("XDG_CONFIG_HOME")
            .map(PathBuf::from)
            .unwrap_or_else(|_| platform_home().join(".config"));
        base.join(APP_ID)
    }
    #[cfg(target_os = "macos")]
    {
        platform_home().join("Library/Application Support").join(APP_ID)
    }
    #[cfg(target_os = "windows")]
    {
        PathBuf::from(std::env::var("APPDATA").unwrap_or_else(|_| ".".to_string())).join(APP_ID)
    }
    #[cfg(not(any(target_os = "linux", target_os = "macos", target_os = "windows")))]
    PathBuf::from(".").join(APP_ID)
}

fn warmdesk_data_dir() -> PathBuf {
    #[cfg(target_os = "linux")]
    {
        let base = std::env::var("XDG_DATA_HOME")
            .map(PathBuf::from)
            .unwrap_or_else(|_| platform_home().join(".local/share"));
        base.join(APP_ID)
    }
    #[cfg(target_os = "macos")]
    {
        platform_home().join("Library/Application Support").join(APP_ID)
    }
    #[cfg(target_os = "windows")]
    {
        PathBuf::from(std::env::var("APPDATA").unwrap_or_else(|_| ".".to_string())).join(APP_ID)
    }
    #[cfg(not(any(target_os = "linux", target_os = "macos", target_os = "windows")))]
    PathBuf::from(".").join(APP_ID)
}

// ---------------------------------------------------------------------------
// Profile file I/O
// ---------------------------------------------------------------------------

fn profiles_path() -> PathBuf {
    warmdesk_config_dir().join("profiles.json")
}

fn load_profiles() -> ProfilesConfig {
    if let Ok(data) = std::fs::read_to_string(profiles_path()) {
        if let Ok(p) = serde_json::from_str::<ProfilesConfig>(&data) {
            if !p.profiles.is_empty() {
                return p;
            }
        }
    }
    // First run: synthesise a single "default" profile.
    ProfilesConfig {
        default: "default".to_string(),
        profiles: vec![Profile {
            name: "default".to_string(),
            label: "Default".to_string(),
        }],
    }
}

fn save_profiles(cfg: &ProfilesConfig) -> std::io::Result<()> {
    std::fs::create_dir_all(warmdesk_config_dir())?;
    let data = serde_json::to_string_pretty(cfg)
        .map_err(|e| std::io::Error::new(std::io::ErrorKind::Other, e))?;
    std::fs::write(profiles_path(), data)
}

fn validate_profile_name(name: &str) -> Result<(), String> {
    if name.is_empty() {
        return Err("Profile name cannot be empty".to_string());
    }
    if !name
        .chars()
        .all(|c| c.is_alphanumeric() || c == '-' || c == '_')
    {
        return Err(
            "Profile name may only contain letters, digits, hyphens, and underscores".to_string(),
        );
    }
    Ok(())
}

// ---------------------------------------------------------------------------
// Argument helpers
// ---------------------------------------------------------------------------

/// Parse `--flag value` or `--flag=value` from args.
fn parse_flag_value(args: &[String], flag: &str) -> Option<String> {
    let prefix = format!("{}=", flag);
    let mut iter = args.iter();
    while let Some(arg) = iter.next() {
        if let Some(val) = arg.strip_prefix(&prefix) {
            return Some(val.to_string());
        }
        if arg == flag {
            return iter.next().cloned();
        }
    }
    None
}

// ---------------------------------------------------------------------------
// JS injection helpers
// ---------------------------------------------------------------------------

fn profile_window_title(profile: &Profile) -> String {
    if profile.name == "default" {
        "WarmDesk".to_string()
    } else {
        format!("WarmDesk \u{2014} {}", profile.label)
    }
}

fn build_init_js(server_url: Option<&str>, profile: &Profile) -> String {
    let mut parts: Vec<String> = Vec::new();
    if let Some(url) = server_url {
        if let Ok(js) = serde_json::to_string(url) {
            parts.push(format!(
                "window.__WARMDESK_RUNTIME_SERVER_URL__={0};\
                 sessionStorage.setItem('warmdesk_runtime_server_url',{0});",
                js
            ));
        }
    }
    if let (Ok(name), Ok(label)) = (
        serde_json::to_string(&profile.name),
        serde_json::to_string(&profile.label),
    ) {
        parts.push(format!(
            "window.__WARMDESK_PROFILE_NAME__={};window.__WARMDESK_PROFILE_LABEL__={};",
            name, label
        ));
    }
    parts.join("")
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

#[cfg_attr(mobile, tauri::mobile_entry_point)]
// ---------------------------------------------------------------------------
// DIAGNOSTIC: startup timing log
//
// Truncated at the start of every run, then appended to with HH:MM:SS.mmm
// timestamps at key points in the native startup sequence. Read it after
// reproducing a slow/blank-screen launch to see which stage the delay is in.
// Windows: %APPDATA%\com.warmdesk.desktop\warmdesk-startup.log
// Remove once the Windows startup-delay investigation concludes.
// ---------------------------------------------------------------------------

fn startup_log(msg: &str) {
    use std::io::Write as _;
    let dir = warmdesk_data_dir();
    let _ = std::fs::create_dir_all(&dir);
    if let Ok(mut f) = std::fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open(dir.join("warmdesk-startup.log"))
    {
        let now = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default();
        let h = (now.as_secs() % 86400) / 3600;
        let m = (now.as_secs() % 3600) / 60;
        let s = now.as_secs() % 60;
        let ms = now.subsec_millis();
        let _ = writeln!(f, "{h:02}:{m:02}:{s:02}.{ms:03} | {msg}");
    }
}

pub fn run() {
    let _ = std::fs::write(warmdesk_data_dir().join("warmdesk-startup.log"), "");
    startup_log("run() started — Rust entry point reached");

    let args: Vec<String> = std::env::args().collect();

    if args.iter().any(|a| a == "--help" || a == "-h") {
        print_help();
        std::process::exit(0);
    }

    if args.iter().any(|a| a == "--version" || a == "-V") {
        println!("WarmDesk {}", env!("CARGO_PKG_VERSION"));
        std::process::exit(0);
    }

    if args.iter().any(|a| a == "--list-profiles") {
        let cfg = load_profiles();
        println!("WarmDesk profiles  ({})\n", profiles_path().display());
        for p in &cfg.profiles {
            let marker = if p.name == cfg.default { '*' } else { ' ' };
            println!("  {} {:<24} {}", marker, p.name, p.label);
        }
        println!("\n  * = default  |  use --profile <name> to select");
        std::process::exit(0);
    }

    if let Some(name) = parse_flag_value(&args, "--create-profile") {
        if let Err(e) = validate_profile_name(&name) {
            eprintln!("error: {}", e);
            std::process::exit(1);
        }
        let label = parse_flag_value(&args, "--label").unwrap_or_else(|| name.clone());
        let mut cfg = load_profiles();
        if cfg.profiles.iter().any(|p| p.name == name) {
            eprintln!("error: profile '{}' already exists", name);
            std::process::exit(1);
        }
        cfg.profiles.push(Profile { name: name.clone(), label: label.clone() });
        if let Err(e) = save_profiles(&cfg) {
            eprintln!("error: could not save profiles: {}", e);
            std::process::exit(1);
        }
        println!("Created profile '{}' ({})", name, label);
        std::process::exit(0);
    }

    if let Some(name) = parse_flag_value(&args, "--set-default") {
        let mut cfg = load_profiles();
        if !cfg.profiles.iter().any(|p| p.name == name) {
            eprintln!("error: profile '{}' not found", name);
            std::process::exit(1);
        }
        cfg.default = name.clone();
        if let Err(e) = save_profiles(&cfg) {
            eprintln!("error: could not save profiles: {}", e);
            std::process::exit(1);
        }
        println!("Default profile set to '{}'", name);
        std::process::exit(0);
    }

    if let Some(name) = parse_flag_value(&args, "--delete-profile") {
        let mut cfg = load_profiles();
        if cfg.profiles.len() <= 1 {
            eprintln!("error: cannot delete the last remaining profile");
            std::process::exit(1);
        }
        if cfg.default == name {
            eprintln!("error: cannot delete the default profile; use --set-default <name> first");
            std::process::exit(1);
        }
        if !cfg.profiles.iter().any(|p| p.name == name) {
            eprintln!("error: profile '{}' not found", name);
            std::process::exit(1);
        }
        cfg.profiles.retain(|p| p.name != name);
        if let Err(e) = save_profiles(&cfg) {
            eprintln!("error: could not save profiles: {}", e);
            std::process::exit(1);
        }
        println!("Deleted profile '{}'", name);
        println!("Note: profile data directory was not removed; clean it up manually if desired.");
        std::process::exit(0);
    }

    let maximized = args.iter().any(|a| a == "--maximized");
    let runtime_server_url_override = parse_server_url_override(&args);
    let runtime_server_url_for_page_load = runtime_server_url_override.clone();

    // ------------------------------------------------------------------
    // Resolve active profile
    // ------------------------------------------------------------------
    let profiles_cfg = load_profiles();
    let requested_name = parse_flag_value(&args, "--profile")
        .unwrap_or_else(|| profiles_cfg.default.clone());
    let active_profile = match profiles_cfg
        .profiles
        .iter()
        .find(|p| p.name == requested_name)
    {
        Some(p) => p.clone(),
        None => {
            eprintln!(
                "error: profile '{}' not found.\n\
                 Run with --list-profiles to see available profiles.",
                requested_name
            );
            std::process::exit(1);
        }
    };

    let profile_data_dir = warmdesk_data_dir()
        .join("profiles")
        .join(&active_profile.name);
    if let Err(e) = std::fs::create_dir_all(&profile_data_dir) {
        eprintln!(
            "warning: could not create profile data directory {}: {}",
            profile_data_dir.display(),
            e
        );
    }

    let active_profile_for_page_load = active_profile.clone();
    let active_profile_for_setup = active_profile.clone();
    let profile_data_dir_for_setup = profile_data_dir;

    // ------------------------------------------------------------------
    // Linux environment tweaks (unchanged from original)
    // ------------------------------------------------------------------
    // On Linux, WebKitGTK's DMA-BUF renderer silently fails on many GPU
    // configurations (integrated GPUs, NVIDIA, VMs, some Wayland compositors),
    // producing a completely blank window.  Disabling it forces the fallback
    // compositing path, which works reliably across all configurations.
    //
    // GDK_RENDERING=image forces GDK into full software rendering, which also
    // prevents a COLRv1 font crash in webkit2gtk/Skia (Fedora 43,
    // webkit2gtk 2.50.x): assertion failure in colrv1_configure_skpaint when
    // rendering colour emoji.  The env var checks let users override if needed.
    //
    // SAFETY: single-threaded at this point, before the Tauri runtime starts.
    #[cfg(target_os = "linux")]
    {
        if std::env::var("WEBKIT_DISABLE_DMABUF_RENDERER").is_err() {
            unsafe { std::env::set_var("WEBKIT_DISABLE_DMABUF_RENDERER", "1") };
        }
        if std::env::var("GDK_RENDERING").is_err() {
            unsafe { std::env::set_var("GDK_RENDERING", "image") };
        }
        // When running as an AppImage, Tauri's AppRun does not export
        // GST_PLUGIN_PATH, so GStreamer falls back to system plugin paths
        // (e.g. /usr/lib64 on Fedora).  Those paths contain GStreamer 1.26
        // plugins that are ABI-incompatible with the ubuntu-24.04 GStreamer
        // core bundled in the AppImage, causing crashes or "element not found"
        // errors (appsink, v4l2src, …).
        //
        // APPDIR is set by the AppImage runtime.  linuxdeploy-plugin-gstreamer
        // copies the build-host plugins to $APPDIR/usr/lib/gstreamer-1.0/ and
        // the plugin scanner to …/gstreamer-1.0/gst-plugin-scanner.  We point
        // GStreamer there so it only sees the bundled, ABI-compatible plugins.
        // The exists() guard is a no-op outside AppImages (APPDIR not set) or
        // when the plugin was not bundled.
        if std::env::var("GST_PLUGIN_PATH").is_err() {
            if let Ok(appdir) = std::env::var("APPDIR") {
                let plugin_dir =
                    std::path::PathBuf::from(&appdir).join("usr/lib/gstreamer-1.0");
                // Only activate when plugins are actually present; an empty
                // directory still passes exists() and would set GST_PLUGIN_PATH
                // to an empty path, causing GStreamer to find nothing.
                let has_plugins = plugin_dir
                    .read_dir()
                    .map(|mut d| {
                        d.any(|e| {
                            e.map(|e| e.file_name().to_string_lossy().ends_with(".so"))
                                .unwrap_or(false)
                        })
                    })
                    .unwrap_or(false);
                if has_plugins {
                    unsafe { std::env::set_var("GST_PLUGIN_PATH", &plugin_dir) };
                    // Override the compiled-in Ubuntu system plugin path so GStreamer
                    // does not fall back to the (nonexistent on Fedora) Ubuntu path.
                    if std::env::var("GST_PLUGIN_SYSTEM_PATH_1_0").is_err() {
                        unsafe {
                            std::env::set_var("GST_PLUGIN_SYSTEM_PATH_1_0", &plugin_dir)
                        };
                    }
                    let scanner = plugin_dir.join("gstreamer-1.0/gst-plugin-scanner");
                    if scanner.exists() && std::env::var("GST_PLUGIN_SCANNER").is_err() {
                        unsafe { std::env::set_var("GST_PLUGIN_SCANNER", &scanner) };
                    }
                    // Use a private registry so Fedora's system registry cache
                    // (built from GStreamer 1.26 plugins) does not hide our bundled
                    // Ubuntu 1.24 plugins.  The file is shared across AppImage runs
                    // so GStreamer only rescans when plugins actually change.
                    if std::env::var("GST_REGISTRY").is_err() {
                        unsafe {
                            std::env::set_var(
                                "GST_REGISTRY",
                                "/tmp/warmdesk-gst-registry.bin",
                            )
                        };
                    }
                }
            }
        }
    }

    // On Windows, reqwest (used by tauri-plugin-http) performs Windows Proxy
    // Auto-Detection (WPAD) the first time it builds an HTTP client, which
    // blocks the Tokio thread pool handling Tauri's IPC channel for 30-70+
    // seconds on networks where the WPAD endpoint is unreachable (i.e. most
    // networks without a corporate proxy) — this is what makes the app look
    // hung with a blank window on startup. Setting NO_PROXY=* makes reqwest
    // skip proxy detection entirely. Must be set before any HTTP client is
    // built, so this happens before the Tauri runtime (and its plugins) start.
    // SAFETY: single-threaded at this point, before the Tauri runtime starts.
    #[cfg(target_os = "windows")]
    {
        if std::env::var("NO_PROXY").is_err() {
            unsafe { std::env::set_var("NO_PROXY", "*") };
        }
        if std::env::var("no_proxy").is_err() {
            unsafe { std::env::set_var("no_proxy", "*") };
        }
    }

    // ------------------------------------------------------------------
    // Build and run Tauri
    // ------------------------------------------------------------------
    tauri::Builder::default()
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_fs::init())
        .plugin(tauri_plugin_http::init())
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_notification::init())
        .manage(RuntimeSettings {
            runtime_server_url: runtime_server_url_override.clone(),
            profile_name: active_profile.name.clone(),
            profile_label: active_profile.label.clone(),
        })
        .invoke_handler(tauri::generate_handler![
            runtime_server_url,
            client_log,
            installation_method,
            fetch_binary_b64,
            current_profile,
            list_profiles,
            create_profile,
            rename_profile,
            set_default_profile,
            delete_profile,
            set_tray_unread,
        ])
        .on_page_load(move |window, _payload| {
            startup_log("on_page_load fired — WebView2 has loaded the HTML page");
            let js = build_init_js(
                runtime_server_url_for_page_load.as_deref(),
                &active_profile_for_page_load,
            );
            if !js.is_empty() {
                let _ = window.eval(&js);
            }
        })
        .setup(move |app| {
            let title = profile_window_title(&active_profile_for_setup);
            #[cfg_attr(not(target_os = "windows"), allow(unused_mut))]
            let mut win_builder = tauri::WebviewWindowBuilder::new(
                app,
                "main",
                tauri::WebviewUrl::App("index.html".into()),
            )
            .title(&title)
            .inner_size(1280.0, 800.0)
            .min_inner_size(900.0, 600.0)
            .data_directory(profile_data_dir_for_setup);
            // WebView2's own content process does its own WPAD proxy
            // auto-detection independent of reqwest's — bypass it here too
            // (see the NO_PROXY comment above `tauri::Builder::default()`).
            //
            // additional_browser_args() REPLACES wry's default args rather than
            // appending to them, so the default set (which works around a stray
            // WebView2 mini-menu and a SmartScreen issue — see wry's webview2/mod.rs)
            // must be included explicitly here or it's silently lost.
            #[cfg(target_os = "windows")]
            {
                // DIAGNOSTIC: --disable-gpu tests whether the WebView2 renderer
                // crash reported as "Out of Memory" is actually a GPU/graphics
                // driver crash (Chromium's crash page mislabels many renderer
                // crashes this way). Remove once the startup-delay/crash
                // investigation concludes either way.
                win_builder = win_builder.additional_browser_args(
                    "--disable-features=msWebOOUI,msPdfOOUI,msSmartScreenProtection --no-proxy-server --disable-gpu",
                );
            }
            startup_log("WebviewWindowBuilder::build() — calling now");
            let win = win_builder.build()?;
            startup_log("WebviewWindowBuilder::build() — returned (window visible)");

            // Inject before the first on_page_load fires (best-effort).
            let js = build_init_js(
                runtime_server_url_override.as_deref(),
                &active_profile_for_setup,
            );
            if !js.is_empty() {
                let _ = win.eval(&js);
            }

            if maximized {
                win.maximize()?;
            }

            // Close-to-tray: hide instead of quit when enabled, so the tray stays active.
            let win_hide = win.clone();
            win.on_window_event(move |event| {
                if let tauri::WindowEvent::CloseRequested { api, .. } = event {
                    if !CLOSE_TO_TRAY_ENABLED.load(Ordering::Relaxed) {
                        return;
                    }
                    api.prevent_close();
                    let _ = win_hide.hide();
                }
            });

            // Disable WebKit hardware acceleration on Linux to work around
            // a COLRv1 font rendering crash in webkit2gtk/Skia (Fedora 43,
            // webkit2gtk 2.50.x). Forces software compositing.
            #[cfg(target_os = "linux")]
            win.with_webview(|webview| {
                use webkit2gtk::{
                    HardwareAccelerationPolicy, PermissionRequestExt, SettingsExt, WebViewExt,
                };
                if let Some(settings) = WebViewExt::settings(&webview.inner()) {
                    settings.set_hardware_acceleration_policy(
                        HardwareAccelerationPolicy::Never,
                    );
                    // webkit2gtk disables getUserMedia by default (false).
                    // Must be enabled explicitly or navigator.mediaDevices
                    // returns no devices at all.
                    settings.set_enable_media_stream(true);
                }
                // webkit2gtk denies all getUserMedia requests by default.
                // Allow them so the device-selection dropdown and call
                // previews can access the microphone and camera.
                webview.inner().connect_permission_request(|_view, request| {
                    request.allow();
                    true
                });
            })?;

            // On Windows, WebView2's autofill/credential service sends a
            // synchronous IPC message to its browser process on every
            // keystroke in any field it classifies as a password field.
            // ICoreWebView2Settings4 (WebView2 SDK 1.0.992+) exposes
            // IsPasswordAutosaveEnabled and IsGeneralAutofillEnabled —
            // disabling both eliminates that round-trip and removes the
            // typing lag on the login screen.
            #[cfg(target_os = "windows")]
            {
                startup_log("with_webview() — calling now (waiting for WebView2 controller)");
                win.with_webview(|wv| {
                    startup_log("with_webview() — callback entered (controller ready)");
                    use webview2_com::Microsoft::Web::WebView2::Win32::ICoreWebView2Settings4;
                    use windows::core::Interface;
                    unsafe {
                        let Ok(core) = wv.controller().CoreWebView2() else { return };
                        let Ok(settings) = core.Settings() else { return };
                        if let Ok(s4) = settings.cast::<ICoreWebView2Settings4>() {
                            let _ = s4.SetIsGeneralAutofillEnabled(false);
                            let _ = s4.SetIsPasswordAutosaveEnabled(false);
                        }
                    }
                    startup_log("with_webview() — callback done");
                })?;
                startup_log("with_webview() — returned");
            }
            startup_log("setup() reaching end — Rust side complete");

            // ── System tray icon ────────────────────────────────────────────
            let tray_icon = png_to_image(include_bytes!("../icons/tray-icon.png"))
                .expect("failed to load tray icon");
            let show_item = MenuItemBuilder::with_id("show", "WarmDesk")
                .build(app)?;
            let quit_item = MenuItemBuilder::with_id("quit", "Quit")
                .build(app)?;
            let tray_menu = MenuBuilder::new(app)
                .item(&show_item)
                .separator()
                .item(&quit_item)
                .build()?;
            TrayIconBuilder::new()
                .icon(tray_icon)
                .title("WarmDesk")
                .tooltip("WarmDesk")
                .menu(&tray_menu)
                .on_menu_event(|app, event| {
                    match event.id.as_ref() {
                        "show" => {
                            if let Some(w) = app.get_webview_window("main") {
                                let visible = w.is_visible().unwrap_or(true);
                                if visible {
                                    let _ = w.hide();
                                } else {
                                    let _ = w.show();
                                    let _ = w.set_focus();
                                }
                            }
                        }
                        "quit" => {
                            app.exit(0);
                        }
                        _ => {}
                    }
                })
                .build(app)?;

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running WarmDesk");
}

// ---------------------------------------------------------------------------
// Tauri state
// ---------------------------------------------------------------------------

#[derive(Clone)]
struct RuntimeSettings {
    runtime_server_url: Option<String>,
    profile_name: String,
    profile_label: String,
}

// ---------------------------------------------------------------------------
// Tauri commands
// ---------------------------------------------------------------------------

#[tauri::command]
fn runtime_server_url(state: tauri::State<'_, RuntimeSettings>) -> Option<String> {
    state.runtime_server_url.clone()
}

// DIAGNOSTIC: lets main.js append its own timing laps to warmdesk-startup.log
// so JS-side and native-side timestamps land in one unified timeline. Remove
// once the Windows startup-delay investigation concludes.
#[tauri::command]
fn client_log(msg: String) {
    startup_log(&format!("[js] {msg}"));
}

#[tauri::command]
fn current_profile(state: tauri::State<'_, RuntimeSettings>) -> ProfileInfo {
    let cfg = load_profiles();
    ProfileInfo {
        name: state.profile_name.clone(),
        label: state.profile_label.clone(),
        is_default: state.profile_name == cfg.default,
    }
}

#[tauri::command]
fn list_profiles() -> Vec<ProfileInfo> {
    let cfg = load_profiles();
    cfg.profiles
        .iter()
        .map(|p| ProfileInfo {
            name: p.name.clone(),
            label: p.label.clone(),
            is_default: p.name == cfg.default,
        })
        .collect()
}

#[tauri::command]
fn create_profile(name: String, label: String) -> Result<(), String> {
    validate_profile_name(&name)?;
    let mut cfg = load_profiles();
    if cfg.profiles.iter().any(|p| p.name == name) {
        return Err(format!("Profile '{}' already exists", name));
    }
    cfg.profiles.push(Profile { name, label });
    save_profiles(&cfg).map_err(|e| e.to_string())
}

#[tauri::command]
fn rename_profile(name: String, new_label: String) -> Result<(), String> {
    let mut cfg = load_profiles();
    let p = cfg
        .profiles
        .iter_mut()
        .find(|p| p.name == name)
        .ok_or_else(|| format!("Profile '{}' not found", name))?;
    p.label = new_label;
    save_profiles(&cfg).map_err(|e| e.to_string())
}

#[tauri::command]
fn set_default_profile(name: String) -> Result<(), String> {
    let mut cfg = load_profiles();
    if !cfg.profiles.iter().any(|p| p.name == name) {
        return Err(format!("Profile '{}' not found", name));
    }
    cfg.default = name;
    save_profiles(&cfg).map_err(|e| e.to_string())
}

#[tauri::command]
fn delete_profile(name: String) -> Result<(), String> {
    let mut cfg = load_profiles();
    if cfg.profiles.len() <= 1 {
        return Err("Cannot delete the last remaining profile".to_string());
    }
    if cfg.default == name {
        return Err(
            "Cannot delete the default profile; set another profile as default first".to_string(),
        );
    }
    if !cfg.profiles.iter().any(|p| p.name == name) {
        return Err(format!("Profile '{}' not found", name));
    }
    cfg.profiles.retain(|p| p.name != name);
    // The profile's data directory is intentionally left on disk to avoid
    // accidental data loss; the user can clean it up manually if desired.
    save_profiles(&cfg).map_err(|e| e.to_string())
}

// ---------------------------------------------------------------------------
// Tray icon commands
// ---------------------------------------------------------------------------

#[tauri::command]
fn set_tray_unread(
    app: tauri::AppHandle,
    count: u32,
    enabled: bool,
    is_timetracking: bool,
    close_to_tray: bool,
) -> Result<(), String> {
    CLOSE_TO_TRAY_ENABLED.store(close_to_tray, Ordering::Relaxed);

    let tray = app.tray_by_id("main").ok_or("tray not found")?;
    let (normal_icon, badge_icon) = if is_timetracking {
        (
            include_bytes!("../icons/timetracking-tray-icon.png") as &[u8],
            include_bytes!("../icons/timetracking-tray-icon-badge.png") as &[u8],
        )
    } else {
        (
            include_bytes!("../icons/tray-icon.png") as &[u8],
            include_bytes!("../icons/tray-icon-badge.png") as &[u8],
        )
    };

    let title = if is_timetracking {
        "WarmDesk — Time Tracking"
    } else {
        "WarmDesk"
    };
    tray.set_title(Some(title)).map_err(|e| e.to_string())?;

    if !enabled {
        let icon = png_to_image(normal_icon)?;
        tray.set_icon(Some(icon)).map_err(|e| e.to_string())?;
        tray.set_tooltip(Some(title)).map_err(|e| e.to_string())?;
        return Ok(());
    }

    let icon = if count > 0 {
        png_to_image(badge_icon)?
    } else {
        png_to_image(normal_icon)?
    };
    tray.set_icon(Some(icon)).map_err(|e| e.to_string())?;

    let tooltip = if count > 0 {
        format!("{} — {} unread", title, count)
    } else {
        title.to_string()
    };
    tray.set_tooltip(Some(&tooltip)).map_err(|e| e.to_string())?;

    Ok(())
}

// ---------------------------------------------------------------------------
// Installation detection
// ---------------------------------------------------------------------------

/// Returns how the desktop client was installed.
/// - `"appimage"`  — running as an AppImage
/// - `"deb"`       — installed via .deb package (dpkg-based system)
/// - `"rpm"`       — installed via .rpm package (rpm-based system)
/// - `"portable"`  — portable archive or source build
/// - `"dmg"`       — macOS DMG
/// - `"windows"`   — Windows installer / portable
#[tauri::command]
fn installation_method() -> String {
    #[cfg(target_os = "linux")]
    {
        if std::env::var("APPDIR").is_ok() {
            return "appimage".to_string();
        }
        // Resolve symlinks so dpkg/rpm get the real on-disk path.
        let exe = match std::env::current_exe()
            .ok()
            .and_then(|p| std::fs::canonicalize(p).ok())
        {
            Some(p) => p,
            None => return "portable".to_string(),
        };
        let exe_str = exe.to_string_lossy();
        match linux_os_family().as_str() {
            "debian" => {
                // dpkg --search <path>: exit 0 when the path is owned by a package.
                let owned = std::process::Command::new("dpkg")
                    .args(["--search", exe_str.as_ref()])
                    .output()
                    .map(|o| o.status.success())
                    .unwrap_or(false);
                if owned {
                    "deb".to_string()
                } else {
                    "portable".to_string()
                }
            }
            "redhat" => {
                // rpm -qf <path>: exit 0 when the path is owned by a package.
                let owned = std::process::Command::new("rpm")
                    .args(["-qf", exe_str.as_ref()])
                    .output()
                    .map(|o| o.status.success())
                    .unwrap_or(false);
                if owned {
                    "rpm".to_string()
                } else {
                    "portable".to_string()
                }
            }
            _ => "portable".to_string(),
        }
    }
    #[cfg(target_os = "macos")]
    {
        "dmg".to_string()
    }
    #[cfg(target_os = "windows")]
    {
        "windows".to_string()
    }
    #[cfg(not(any(target_os = "linux", target_os = "macos", target_os = "windows")))]
    {
        "unknown".to_string()
    }
}

/// Parse /etc/os-release and return the distro family: `"debian"`, `"redhat"`, or `"unknown"`.
///
/// Explicit ID= lists are checked first (source: Ansible OS_FAMILY_MAP) so that distros
/// without a useful ID_LIKE= (e.g. Amazon Linux 2) are still classified correctly.
/// The id_like keyword fallback handles future or unlisted distros that follow conventions.
#[cfg(target_os = "linux")]
fn linux_os_family() -> String {
    let content = match std::fs::read_to_string("/etc/os-release") {
        Ok(c) => c,
        Err(_) => return "unknown".to_string(),
    };
    let mut id = String::new();
    let mut id_like = String::new();
    for line in content.lines() {
        if let Some(v) = line.strip_prefix("ID=") {
            id = v.trim_matches('"').to_lowercase();
        } else if let Some(v) = line.strip_prefix("ID_LIKE=") {
            id_like = v.trim_matches('"').to_lowercase();
        }
    }

    // RedHat family: RHEL, Fedora, and all known rebuilds/derivatives.
    const REDHAT_IDS: &[&str] = &[
        "rhel",
        "fedora",
        "centos",
        "scientific",
        "slc",
        "ascendos",
        "cloudlinux",
        "psbm",
        "ol",
        "ovs",
        "amzn",
        "virtuozzo",
        "xenenterprise",
        "alinux",
        "euleros",
        "hce",
        "openeuler",
        "almalinux",
        "rocky",
        "tencentos",
        "eurolinux",
        "kylin",
        "miraclelinux",
    ];
    // Debian family: Debian, Ubuntu, and all known derivatives.
    const DEBIAN_IDS: &[&str] = &[
        "debian",
        "ubuntu",
        "raspbian",
        "neon",
        "linuxmint",
        "devuan",
        "kali",
        "parrot",
        "pop",
        "pardus",
        "deepin",
        "osmc",
        "univention",
        "cumulus-linux",
    ];

    if REDHAT_IDS.contains(&id.as_str()) {
        return "redhat".to_string();
    }
    if DEBIAN_IDS.contains(&id.as_str()) {
        return "debian".to_string();
    }

    // Fallback: id_like keyword matching for unlisted/future distros.
    if id_like.contains("debian") || id_like.contains("ubuntu") {
        "debian".to_string()
    } else if id_like.contains("rhel")
        || id_like.contains("fedora")
        || id_like.contains("centos")
    {
        "redhat".to_string()
    } else {
        "unknown".to_string()
    }
}

// ---------------------------------------------------------------------------
// URL / arg helpers
// ---------------------------------------------------------------------------

fn parse_server_url_override(args: &[String]) -> Option<String> {
    // Supported forms:
    //   --url=http://localhost:8080
    //   --url http://localhost:8080
    let mut i = 0usize;
    while i < args.len() {
        let arg = &args[i];
        if let Some(raw) = arg.strip_prefix("--url=") {
            return normalize_server_url(raw);
        }
        if arg == "--url" {
            if let Some(next) = args.get(i + 1) {
                return normalize_server_url(next);
            }
            return None;
        }
        i += 1;
    }
    None
}

fn normalize_server_url(raw: &str) -> Option<String> {
    let trimmed = raw.trim().trim_end_matches('/').to_string();
    if trimmed.starts_with("http://") || trimmed.starts_with("https://") {
        Some(trimmed)
    } else {
        None
    }
}

// ---------------------------------------------------------------------------
// Binary download (bypasses WebKit's broken Response.arrayBuffer() on Linux)
// ---------------------------------------------------------------------------

#[tauri::command]
async fn fetch_binary_b64(url: String, headers: Vec<(String, String)>) -> Result<String, String> {
    let mut req = reqwest::Client::new().get(&url);
    for (k, v) in &headers {
        req = req.header(k.as_str(), v.as_str());
    }
    let bytes = req
        .send()
        .await
        .map_err(|e| e.to_string())?
        .bytes()
        .await
        .map_err(|e| e.to_string())?;
    Ok(b64_encode(&bytes))
}

fn b64_encode(data: &[u8]) -> String {
    const A: &[u8; 64] =
        b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    let mut s = String::with_capacity((data.len() + 2) / 3 * 4);
    for c in data.chunks(3) {
        let n = c.len();
        let v = (c[0] as u32) << 16
            | if n > 1 { (c[1] as u32) << 8 } else { 0 }
            | if n > 2 { c[2] as u32 } else { 0 };
        s.push(A[((v >> 18) & 63) as usize] as char);
        s.push(A[((v >> 12) & 63) as usize] as char);
        s.push(if n > 1 {
            A[((v >> 6) & 63) as usize] as char
        } else {
            '='
        });
        s.push(if n > 2 {
            A[(v & 63) as usize] as char
        } else {
            '='
        });
    }
    s
}

// ---------------------------------------------------------------------------
// Help text
// ---------------------------------------------------------------------------

fn print_help() {
    let ver = env!("CARGO_PKG_VERSION");
    println!("WarmDesk {ver}");
    println!();
    println!("Usage:");
    println!("  warmdesk [OPTIONS]");
    println!();
    println!("Options:");
    println!("  -h, --help                   Show this help and exit");
    println!("  -V, --version                Show version and exit");
    println!("      --maximized              Start with the main window maximized");
    println!("      --url <URL>              Override server URL for this launch only");
    println!("      --url=<URL>              Same as above");
    println!("      --profile <NAME>         Launch with the named profile");
    println!("      --profile=<NAME>         Same as above");
    println!("      --list-profiles          List available profiles and exit");
    println!("      --create-profile <NAME>  Create a new profile and exit");
    println!("        --label <LABEL>        Human-readable label for --create-profile");
    println!("      --set-default <NAME>     Set the default profile and exit");
    println!("      --delete-profile <NAME>  Remove a profile and exit");
    println!();
    println!("Profiles:");
    println!("  Each profile has its own isolated localStorage, login session, and");
    println!("  settings.  Profiles are defined in:");
    println!("    Linux:   ~/.config/com.warmdesk.desktop/profiles.json");
    println!("    macOS:   ~/Library/Application Support/com.warmdesk.desktop/profiles.json");
    println!("    Windows: %APPDATA%\\com.warmdesk.desktop\\profiles.json");
    println!();
    println!("  Profile data is stored under:");
    println!("    Linux:   ~/.local/share/com.warmdesk.desktop/profiles/<name>/");
    println!("    macOS:   ~/Library/Application Support/com.warmdesk.desktop/profiles/<name>/");
    println!("    Windows: %APPDATA%\\com.warmdesk.desktop\\profiles\\<name>\\");
    println!();
    println!("  profiles.json format:");
    println!("    {{");
    println!("      \"default\": \"work\",");
    println!("      \"profiles\": [");
    println!("        {{ \"name\": \"work\",       \"label\": \"Work\" }},");
    println!("        {{ \"name\": \"customer-a\", \"label\": \"Customer A\" }}");
    println!("      ]");
    println!("    }}");
    println!();
    println!("Notes:");
    println!("  - URL override is runtime-only and is not saved to settings.");
    println!("  - URL must start with http:// or https://");
    println!("  - Profile names may only contain letters, digits, hyphens, and underscores.");
}
