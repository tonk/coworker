use std::path::PathBuf;

use serde::{Deserialize, Serialize};

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
// Startup diagnostic log
//
// Each run truncates the file first, then appends HH:MM:SS.mmm timestamps.
// On Windows: %APPDATA%\com.warmdesk.desktop\warmdesk-startup.log
// Read it after launch to pinpoint where the 33-second startup delay occurs.
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

// ---------------------------------------------------------------------------
// Windows proxy configuration — prevents WPAD-triggered tokio hangs
// ---------------------------------------------------------------------------

/// Set HTTP(S)_PROXY env vars so reqwest never calls WinHTTP WPAD detection.
///
/// reqwest checks HTTP_PROXY / HTTPS_PROXY first; only when they are absent
/// does it call WinHttpGetProxyForUrl with WINHTTP_AUTOPROXY_AUTO_DETECT,
/// which can block the tokio thread pool for 30-70 s.  By always providing a
/// value we short-circuit that code path entirely.
#[cfg(target_os = "windows")]
fn configure_reqwest_proxy() {
    // Respect any proxy env vars the operator already set explicitly.
    let http_already_set = std::env::var_os("HTTP_PROXY").is_some();
    let https_already_set = std::env::var_os("HTTPS_PROXY").is_some();
    if http_already_set && https_already_set {
        startup_log("proxy: HTTP_PROXY+HTTPS_PROXY already set — leaving unchanged");
        return;
    }

    // Read the user's manual proxy from the Windows Internet Settings registry
    // key.  This is a synchronous registry read (<1 ms) with no WPAD involved.
    let manual_proxy = read_windows_manual_proxy();

    // SAFETY: single-threaded here, before the Tauri runtime starts.
    unsafe {
        if let Some(ref url) = manual_proxy {
            // Forward the manually configured proxy so requests still go through it.
            if !http_already_set {
                std::env::set_var("HTTP_PROXY", url);
            }
            if !https_already_set {
                std::env::set_var("HTTPS_PROXY", url);
            }
            startup_log(&format!("proxy: manual proxy → HTTP(S)_PROXY={}", url));
        } else {
            // No manual proxy (disabled or WPAD).  Point reqwest at an
            // unreachable dummy so it never calls WinHTTP WPAD detection,
            // then set NO_PROXY=* so every request bypasses the dummy and
            // connects directly.
            if !http_already_set {
                std::env::set_var("HTTP_PROXY", "http://0.0.0.0:0");
            }
            if !https_already_set {
                std::env::set_var("HTTPS_PROXY", "http://0.0.0.0:0");
            }
            if std::env::var_os("NO_PROXY").is_none() {
                std::env::set_var("NO_PROXY", "*");
            }
            if std::env::var_os("no_proxy").is_none() {
                std::env::set_var("no_proxy", "*");
            }
            startup_log("proxy: no manual proxy → dummy HTTP(S)_PROXY + NO_PROXY=* (direct)");
        }
    }
}

/// Return the manually configured Windows proxy as `"http://host:port"`, or
/// `None` when no manual proxy is active (proxy is disabled or uses WPAD).
#[cfg(target_os = "windows")]
fn read_windows_manual_proxy() -> Option<String> {
    use winreg::{enums::HKEY_CURRENT_USER, RegKey};

    let key = RegKey::predef(HKEY_CURRENT_USER)
        .open_subkey("Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings")
        .ok()?;

    let enabled: u32 = key.get_value("ProxyEnable").unwrap_or(0);
    if enabled == 0 {
        return None;
    }

    let server: String = key.get_value("ProxyServer").unwrap_or_default();
    if server.is_empty() {
        return None;
    }

    // ProxyServer can be a plain "host:port" or a per-protocol string like
    // "http=host:8080;https=host:8080;ftp=host:8080".
    if server.contains('=') {
        for part in server.split(';') {
            if let Some(addr) = part
                .strip_prefix("http=")
                .or_else(|| part.strip_prefix("https="))
            {
                if !addr.is_empty() {
                    return Some(with_http_scheme(addr));
                }
            }
        }
        None
    } else {
        Some(with_http_scheme(&server))
    }
}

#[cfg(target_os = "windows")]
fn with_http_scheme(addr: &str) -> String {
    if addr.starts_with("http://") || addr.starts_with("https://") {
        addr.to_string()
    } else {
        format!("http://{}", addr)
    }
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
pub fn run() {
    let args: Vec<String> = std::env::args().collect();

    // Overwrite the previous run's log so only the current run is visible.
    let _ = std::fs::write(warmdesk_data_dir().join("warmdesk-startup.log"), "");
    startup_log("run() started — Rust entry point reached");

    // Prevent reqwest (used by tauri-plugin-http and fetch_binary_b64) from
    // blocking the tokio thread pool with WinHTTP WPAD proxy auto-detection.
    //
    // Root cause: reqwest::Client::new() on Windows calls WinHttpGetProxyForUrl
    // with WINHTTP_AUTOPROXY_AUTO_DETECT when no HTTP_PROXY / HTTPS_PROXY env
    // var is set, triggering a WPAD DNS/HTTP lookup that can time out after
    // 30-70 s.  While that lookup is in progress, all tokio worker threads are
    // blocked, so every Tauri invoke() call from JavaScript is queued and the
    // app appears frozen.
    //
    // Fix: set HTTP_PROXY / HTTPS_PROXY *before* any reqwest client is built so
    // reqwest reads from the env and never falls through to WinHTTP detection.
    //   • If the user has a manual proxy configured in Windows Internet Settings
    //     we forward it to HTTP(S)_PROXY → proxy connectivity is preserved.
    //   • Otherwise we set a dummy unreachable value and NO_PROXY=* so every
    //     request bypasses the dummy and connects directly — no WPAD, no block.
    //
    // SAFETY: single-threaded here, before the Tauri / tokio runtime starts.
    #[cfg(target_os = "windows")]
    configure_reqwest_proxy();

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
            installation_method,
            fetch_binary_b64,
            current_profile,
            list_profiles,
            create_profile,
            rename_profile,
            set_default_profile,
            delete_profile,
            js_boot_log,
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
            let win_builder = tauri::WebviewWindowBuilder::new(
                app,
                "main",
                tauri::WebviewUrl::App("index.html".into()),
            )
            .title(&title)
            .inner_size(1280.0, 800.0)
            .min_inner_size(900.0, 600.0)
            .data_directory(profile_data_dir_for_setup);

            // On Windows, WebView2's browser process phones home to Microsoft on every
            // startup — SmartScreen data, component updates, metrics, telemetry — and
            // will not load the local page until that background phase completes (~30 s
            // on networks where those endpoints are slow or unreachable).  WarmDesk is
            // a self-hosted tool with no use for any of these Microsoft services.
            //
            // The flags below disable every known call-home mechanism at the browser-
            // process level.  They are passed before the process starts, so they take
            // effect before any background task can be scheduled:
            //
            //   --disable-background-networking          SmartScreen/SafeBrowsing syncs,
            //                                            network-time pings, component
            //                                            update checks, extension updates
            //   --disable-component-update               CRLSet, Pepper, and other
            //                                            on-demand component downloads
            //   --disable-sync                           Microsoft account / profile sync
            //   --no-pings                               <a ping> hyperlink beacons
            //   --no-first-run                           first-run registration flow
            //   --disable-domain-reliability             domain reliability uploads
            //   --disable-client-side-phishing-detection ML phishing model fetches
            //   --disable-features=msSmartScreen         SmartScreen URL reputation
            //   --disable-features=Translate             translation service
            //   --disable-features=AutofillServerCommunication  autofill server sync
            //   --disable-features=MediaRouter           Cast/media routing to external devices
            //   --disable-features=ReportingObserver     Reporting Observer API uploads
            //   --metrics-recording-only                 record UMA histograms locally
            //                                            but never upload them
            #[cfg(target_os = "windows")]
            let win_builder = win_builder.additional_browser_args(
                "--disable-background-networking \
                 --disable-component-update \
                 --disable-sync \
                 --no-pings \
                 --no-first-run \
                 --disable-domain-reliability \
                 --disable-client-side-phishing-detection \
                 --disable-features=msSmartScreen,Translate,AutofillServerCommunication,\
MediaRouter,ReportingObserver \
                 --metrics-recording-only",
            );

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

            // Belt-and-suspenders WebView2 settings applied after the controller
            // is available.  The browser-process flags above handle the same
            // concerns at the process level; these Settings interfaces act as a
            // second layer in case a future WebView2 runtime ignores a flag.
            #[cfg(target_os = "windows")]
            {
                startup_log("with_webview() — calling now (waiting for WebView2 controller)");
                win.with_webview(|wv| {
                    startup_log("with_webview() — callback entered (controller ready)");
                    use webview2_com::Microsoft::Web::WebView2::Win32::{
                        ICoreWebView2Settings4,
                        ICoreWebView2Settings8,
                    };
                    use windows::core::Interface;
                    unsafe {
                        let Ok(core) = wv.controller().CoreWebView2() else { return };
                        let Ok(settings) = core.Settings() else { return };
                        // Autofill / credential IPC — eliminates per-keystroke round-trips
                        // on password fields (typing lag on the login screen).
                        if let Ok(s4) = settings.cast::<ICoreWebView2Settings4>() {
                            let _ = s4.SetIsGeneralAutofillEnabled(false);
                            let _ = s4.SetIsPasswordAutosaveEnabled(false);
                        }
                        // SmartScreen URL reputation checks (WebView2 SDK 1.0.1823+).
                        if let Ok(s8) = settings.cast::<ICoreWebView2Settings8>() {
                            let _ = s8.SetIsReputationCheckingRequired(false);
                        }
                    }
                    startup_log("with_webview() — callback done");
                })?;
                startup_log("with_webview() — returned");
            }

            startup_log("setup() done — Rust side complete, Tauri event loop starting");
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

/// Called from JavaScript at each boot checkpoint so JS timing appears in
/// warmdesk-startup.log alongside the Rust-side timestamps.
#[tauri::command]
fn js_boot_log(msg: String) {
    startup_log(&format!("JS | {msg}"));
}

#[tauri::command]
fn runtime_server_url(state: tauri::State<'_, RuntimeSettings>) -> Option<String> {
    state.runtime_server_url.clone()
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
