// Prevents an extra console window from opening on Windows in release mode.
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

fn main() {
    // On Linux, some WebKitGTK / Wayland configurations produce a blank window
    // because a wrong version of libwayland-client is picked up at dlopen() time.
    // Preloading the system library forces the correct symbol resolution order.
    // This must happen before run() — LD_PRELOAD is honoured by the OS loader
    // at exec() time, so we re-exec ourselves once with the variable set.
    #[cfg(target_os = "linux")]
    preload_wayland();

    warmdesk_lib::run()
}

/// Re-exec the current binary with LD_PRELOAD pointing at libwayland-client.so
/// so that WebKitGTK picks up the correct Wayland symbols on every distribution.
///
/// Does nothing when:
/// - the sentinel env var WARMDESK_PRELOAD_DONE is already set (prevents loops)
/// - libwayland-client.so cannot be found in any of the well-known locations
/// - the library is already present in LD_PRELOAD
#[cfg(target_os = "linux")]
fn preload_wayland() {
    // Already re-execed — nothing to do.
    if std::env::var("WARMDESK_PRELOAD_DONE").is_ok() {
        return;
    }

    // Well-known locations, ordered by preference.
    // The unversioned .so.0 is the runtime library; the unversioned .so is the
    // -devel symlink.  We try both so that this works on base installs (no -devel).
    let candidates: &[&str] = &[
        "/usr/lib64/libwayland-client.so.0",   // Fedora / RHEL (64-bit)
        "/usr/lib64/libwayland-client.so",      // Fedora -devel symlink
        "/usr/lib/x86_64-linux-gnu/libwayland-client.so.0", // Ubuntu / Debian multiarch
        "/usr/lib/x86_64-linux-gnu/libwayland-client.so",
        "/usr/lib/aarch64-linux-gnu/libwayland-client.so.0", // ARM64
        "/usr/lib/aarch64-linux-gnu/libwayland-client.so",
        "/usr/lib/libwayland-client.so.0",      // other distros (non-multiarch)
        "/usr/lib/libwayland-client.so",
    ];

    let lib_path = match candidates
        .iter()
        .find(|&&p| std::path::Path::new(p).exists())
    {
        Some(&p) => p,
        None => return, // Not found; skip preload.
    };

    // Skip if this library is already in LD_PRELOAD.
    let current = std::env::var("LD_PRELOAD").unwrap_or_default();
    if current.split(':').any(|e| e == lib_path) {
        return;
    }

    // Prepend to LD_PRELOAD.
    let new_preload = if current.is_empty() {
        lib_path.to_string()
    } else {
        format!("{lib_path}:{current}")
    };

    // Re-exec: replace the current process image.  All existing environment
    // variables are inherited; we only override LD_PRELOAD and set the sentinel.
    use std::os::unix::process::CommandExt;
    let exe = match std::env::current_exe() {
        Ok(e) => e,
        Err(e) => {
            eprintln!("WarmDesk: could not determine executable path: {e}");
            return;
        }
    };
    let args: Vec<String> = std::env::args().collect();

    // exec() replaces the current process; it only returns on failure.
    let err = std::process::Command::new(&exe)
        .args(&args[1..])
        .env("LD_PRELOAD", &new_preload)
        .env("WARMDESK_PRELOAD_DONE", "1")
        .exec();

    // If we get here, exec failed — log and continue without the preload.
    eprintln!("WarmDesk: wayland preload re-exec failed: {err}");
}
