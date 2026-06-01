#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let args: Vec<String> = std::env::args().collect();

    if args.iter().any(|a| a == "--help" || a == "-h") {
        print_help();
        std::process::exit(0);
    }

    if args.iter().any(|a| a == "--version" || a == "-V") {
        println!("WarmDesk {}", env!("CARGO_PKG_VERSION"));
        std::process::exit(0);
    }

    let maximized = args.iter().any(|a| a == "--maximized");
    let runtime_server_url_override = parse_server_url_override(&args);
    let runtime_server_url_for_page_load = runtime_server_url_override.clone();

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
                let has_plugins = plugin_dir.read_dir()
                    .map(|mut d| d.any(|e| e.map(|e| {
                        e.file_name().to_string_lossy().ends_with(".so")
                    }).unwrap_or(false)))
                    .unwrap_or(false);
                if has_plugins {
                    unsafe { std::env::set_var("GST_PLUGIN_PATH", &plugin_dir) };
                    // Override the compiled-in Ubuntu system plugin path so GStreamer
                    // does not fall back to the (nonexistent on Fedora) Ubuntu path.
                    if std::env::var("GST_PLUGIN_SYSTEM_PATH_1_0").is_err() {
                        unsafe { std::env::set_var("GST_PLUGIN_SYSTEM_PATH_1_0", &plugin_dir) };
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
                        unsafe { std::env::set_var("GST_REGISTRY",
                            "/tmp/warmdesk-gst-registry.bin") };
                    }
                }
            }
        }
    }

    tauri::Builder::default()
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_fs::init())
        .plugin(tauri_plugin_http::init())
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_notification::init())
        .manage(RuntimeSettings {
            runtime_server_url: runtime_server_url_override.clone(),
        })
        .invoke_handler(tauri::generate_handler![runtime_server_url, installation_method, fetch_binary_b64])
        .on_page_load(move |window, _payload| {
            if let Some(url) = &runtime_server_url_for_page_load {
                // Ensure the override is present in every page context.
                if let Ok(js_url) = serde_json::to_string(url) {
                    let _ = window.eval(&format!(
                        "window.__WARMDESK_RUNTIME_SERVER_URL__ = {}; sessionStorage.setItem('warmdesk_runtime_server_url', {});",
                        js_url, js_url
                    ));
                }
            }
        })
        .setup(move |app| {
            if let Some(win) = tauri::Manager::get_webview_window(app, "main") {
                if let Some(url) = &runtime_server_url_override {
                    // Inject the runtime-only URL override before the frontend
                    // performs its first route guard/API base resolution.
                    if let Ok(js_url) = serde_json::to_string(url) {
                        let _ = win.eval(&format!(
                            "window.__WARMDESK_RUNTIME_SERVER_URL__ = {}; sessionStorage.setItem('warmdesk_runtime_server_url', {});",
                            js_url, js_url
                        ));
                    }
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

                // On Windows, WebView2's autofill/credential service sends a
                // synchronous IPC message to its browser process on every
                // keystroke in any field it classifies as a password field.
                // ICoreWebView2Settings4 (WebView2 SDK 1.0.992+) exposes
                // IsPasswordAutosaveEnabled and IsGeneralAutofillEnabled —
                // disabling both eliminates that round-trip and removes the
                // typing lag on the login screen.
                #[cfg(target_os = "windows")]
                win.with_webview(|wv| {
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
                })?;
            }
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running WarmDesk");
}

#[derive(Clone)]
struct RuntimeSettings {
    runtime_server_url: Option<String>,
}

#[tauri::command]
fn runtime_server_url(state: tauri::State<'_, RuntimeSettings>) -> Option<String> {
    state.runtime_server_url.clone()
}

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
                if owned { "deb".to_string() } else { "portable".to_string() }
            }
            "redhat" => {
                // rpm -qf <path>: exit 0 when the path is owned by a package.
                let owned = std::process::Command::new("rpm")
                    .args(["-qf", exe_str.as_ref()])
                    .output()
                    .map(|o| o.status.success())
                    .unwrap_or(false);
                if owned { "rpm".to_string() } else { "portable".to_string() }
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
        "rhel", "fedora", "centos", "scientific", "slc", "ascendos",
        "cloudlinux", "psbm", "ol", "ovs", "amzn", "virtuozzo",
        "xenenterprise", "alinux", "euleros", "hce", "openeuler",
        "almalinux", "rocky", "tencentos", "eurolinux", "kylin",
        "miraclelinux",
    ];
    // Debian family: Debian, Ubuntu, and all known derivatives.
    const DEBIAN_IDS: &[&str] = &[
        "debian", "ubuntu", "raspbian", "neon", "linuxmint", "devuan",
        "kali", "parrot", "pop", "pardus", "deepin", "osmc",
        "univention", "cumulus-linux",
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

// Download a binary resource from `url` using reqwest (bypasses WebKit's
// broken Response.arrayBuffer() / ReadableStream handling on Linux GTK WebKit2)
// and return the bytes as a standard base64 string.
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
        s.push(if n > 1 { A[((v >> 6) & 63) as usize] as char } else { '=' });
        s.push(if n > 2 { A[(v & 63) as usize] as char } else { '=' });
    }
    s
}

fn print_help() {
    println!(
        "WarmDesk {}\n\nUsage:\n  warmdesk [OPTIONS]\n\nOptions:\n  -h, --help                   Show this help and exit\n  -V, --version                Show version and exit\n      --maximized              Start with the main window maximized\n      --url <URL>              Override server URL for this launch only\n      --url=<URL>              Same as above\n\nNotes:\n  - URL override is runtime-only and is not saved to settings.\n  - URL must start with http:// or https://",
        env!("CARGO_PKG_VERSION")
    );
}
