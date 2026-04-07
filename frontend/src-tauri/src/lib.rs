#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let args: Vec<String> = std::env::args().collect();

    if args.iter().any(|a| a == "--version" || a == "-V") {
        println!("WarmDesk {}", env!("CARGO_PKG_VERSION"));
        std::process::exit(0);
    }

    let maximized = args.iter().any(|a| a == "--maximized");

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
    }

    tauri::Builder::default()
        .plugin(tauri_plugin_http::init())
        .plugin(tauri_plugin_opener::init())
        .setup(move |app| {
            if let Some(win) = tauri::Manager::get_webview_window(app, "main") {
                if maximized {
                    win.maximize()?;
                }

                // Disable WebKit hardware acceleration on Linux to work around
                // a COLRv1 font rendering crash in webkit2gtk/Skia (Fedora 43,
                // webkit2gtk 2.50.x). Forces software compositing.
                #[cfg(target_os = "linux")]
                win.with_webview(|webview| {
                    use webkit2gtk::{HardwareAccelerationPolicy, SettingsExt, WebViewExt};
                    if let Some(settings) = WebViewExt::settings(&webview.inner()) {
                        settings.set_hardware_acceleration_policy(
                            HardwareAccelerationPolicy::Never,
                        );
                    }
                })?;
            }
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running WarmDesk");
}
