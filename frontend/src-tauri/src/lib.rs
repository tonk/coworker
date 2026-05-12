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
        .invoke_handler(tauri::generate_handler![runtime_server_url])
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

fn print_help() {
    println!(
        "WarmDesk {}\n\nUsage:\n  warmdesk [OPTIONS]\n\nOptions:\n  -h, --help                   Show this help and exit\n  -V, --version                Show version and exit\n      --maximized              Start with the main window maximized\n      --url <URL>              Override server URL for this launch only\n      --url=<URL>              Same as above\n\nNotes:\n  - URL override is runtime-only and is not saved to settings.\n  - URL must start with http:// or https://",
        env!("CARGO_PKG_VERSION")
    );
}
