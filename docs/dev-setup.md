# Development environment setup

Install the three required tools before running or building WarmDesk:

| Tool | Minimum version | Required for |
|---|---|---|
| make | any | Running all build targets |
| Go | 1.26 | Backend |
| Node.js | 20 LTS | Frontend |
| Rust + Cargo | 1.85 | Desktop (AppImage / DMG / Windows installer) |

---

## Linux — Debian / Ubuntu

### make

```bash
sudo apt install -y build-essential
```

### Go

```bash
# Download and install (replace 1.26.0 with the latest 1.26.x release)
wget https://go.dev/dl/go1.26.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.0.linux-amd64.tar.gz

# Add to PATH (add this line to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin
```

### Node.js 20 LTS

```bash
# Via NodeSource
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs
```

### Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
# Follow the on-screen prompts, then reload your shell:
source "$HOME/.cargo/env"
```

### sccache (recommended for desktop builds)

`frontend/src-tauri/.cargo/config.toml` wraps every Rust compile with `sccache` to cache compiled objects between builds. Only needed when building the Tauri desktop client locally (`make appimage`, `npm run tauri:dev`/`tauri:build`) — not for the backend/frontend web dev loop. CI runners suppress this wrapper (`RUSTC_WRAPPER=""`) since they don't have it installed; local builds should install it so rebuilds are actually fast:

```bash
cargo install sccache --locked
```

If you'd rather skip it for a one-off build: `RUSTC_WRAPPER= cargo build` (matches what CI does).

### AppImage build dependencies

Only needed when building the Linux desktop package (`make appimage`):

```bash
sudo apt install -y \
    libwebkit2gtk-4.1-dev \
    libgtk-3-dev \
    librsvg2-dev \
    libayatana-appindicator3-dev \
    libssl-dev \
    patchelf \
    squashfs-tools \
    gstreamer1.0-tools \
    gstreamer1.0-plugins-base \
    gstreamer1.0-plugins-good \
    gstreamer1.0-plugins-bad \
    gstreamer1.0-pulseaudio
```

The `gstreamer1.0-*` packages aren't linked in — `make appimage` copies them
from this host straight into the built AppImage after `tauri build` (its
`_bundle-gst-appimage` step) so the app's camera/microphone device picker
actually finds devices. Without `plugins-base`/`plugins-good`/`pulseaudio`,
WebKitGTK's bundled GStreamer core has no `v4l2src`/`pulsesrc`/`autoaudiosrc`
elements, so `getUserMedia`'s device list comes back silently empty. Without
**`plugins-bad`** specifically (it provides the `webrtcbin` element),
`RTCPeerConnection` is missing entirely — calls fail with `ReferenceError:
Can't find variable: RTCPeerConnection` even though device selection still
works fine. `_bundle-gst-appimage` warns at build time if no `webrtc*.so`
plugin ends up bundled. Must be built on Ubuntu 24.04 for the copied plugins
to be ABI-compatible with the webkit2gtk GStreamer core also bundled from this
host.

---

## Linux — Fedora / RHEL / CentOS

### make

```bash
sudo dnf -y install make
```

### Go

```bash
# Download and install (replace 1.26.0 with the latest 1.26.x release)
wget https://go.dev/dl/go1.26.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.0.linux-amd64.tar.gz

# Add to PATH (add this line to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin
```

### Node.js 20 LTS

```bash
# Via NodeSource
curl -fsSL https://rpm.nodesource.com/setup_20.x | sudo bash -
sudo dnf -y install nodejs
```

### Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"
```

### sccache (recommended for desktop builds)

`frontend/src-tauri/.cargo/config.toml` wraps every Rust compile with `sccache` to cache compiled objects between builds. Only needed when building the Tauri desktop client locally (`make appimage`, `npm run tauri:dev`/`tauri:build`) — not for the backend/frontend web dev loop.

```bash
sudo dnf -y install sccache
# or, to build the latest release from source:
cargo install sccache --locked
```

If you'd rather skip it for a one-off build: `RUSTC_WRAPPER= cargo build` (matches what CI does).

### AppImage build dependencies

Only needed when building the Linux desktop package (`make appimage`).
Requires Fedora 39+ or RHEL 9+ for `webkit2gtk4.1`:

```bash
sudo dnf -y install \
    webkit2gtk4.1-devel \
    gtk3-devel \
    librsvg2-devel \
    libappindicator-gtk3-devel \
    openssl-devel \
    patchelf \
    squashfs-tools \
    gstreamer1-tools \
    gstreamer1-plugins-base \
    gstreamer1-plugins-good \
    gstreamer1-plugins-good-extras \
    gstreamer1-plugins-bad-free \
    gstreamer1-plugins-bad-free-extras
```

As on Debian/Ubuntu, these aren't linked at build time — `make appimage`'s
`_bundle-gst-appimage` step copies them from this host into the built
AppImage after `tauri build` so camera/microphone selection actually finds
devices (without `plugins-base`/`plugins-good`, WebKitGTK's bundled GStreamer
core has no `v4l2src`/`pulsesrc`/`autoaudiosrc` elements to enumerate with).
Without **`plugins-bad-free`** specifically (it provides the `webrtcbin`
element), `RTCPeerConnection` is missing entirely and calls fail with
`ReferenceError: Can't find variable: RTCPeerConnection`. `_bundle-gst-appimage`
warns at build time if no `webrtc*.so` plugin ends up bundled.

---

## macOS

### Don't have a Mac?

Apple's license only permits virtualizing macOS on top of genuine Apple hardware, so a QEMU/KVM-style VM (like the one described for Windows above) isn't an option here without breaking that license.

Since this repo is public, GitHub gives unlimited free minutes on macOS runners — real Apple hardware, at no cost. `.github/workflows/macos-dev-shell.yml` turns that into an on-demand interactive session instead of just automated CI: trigger it manually (`gh workflow run macos-dev-shell.yml`, or **Actions → macOS dev shell (manual) → Run workflow**), then grab the SSH connection string tmate prints in the job log. You land in a shell with the repo checked out and Rust/Node already set up, ready for `npm run tauri:dev` inside `frontend/`. It only ever runs when triggered by hand, and the SSH session is restricted to whoever's GitHub account triggered it.

Tradeoffs: ephemeral (nothing persists between runs) and capped at 6 hours per job (GitHub's hard limit for hosted runners). If you need persistent, always-on access instead, the cheapest paid fallback is a rented Mac mini (e.g. Scaleway's Mac mini M-series, billed per hour with no minimum commitment).

[Homebrew](https://brew.sh) is the recommended package manager. Install it first if you don't have it:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### make

`make` is part of the Xcode Command Line Tools:

```bash
xcode-select --install
```

### Go

```bash
brew install go@1.26
# If brew installs a newer major version, check: go version
```

### Node.js 20 LTS

```bash
brew install node@20
brew link node@20
```

### Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source "$HOME/.cargo/env"
```

### sccache (recommended for desktop builds)

`frontend/src-tauri/.cargo/config.toml` wraps every Rust compile with `sccache` to cache compiled objects between builds. Only needed when building the Tauri desktop client locally (`make dmg`, `npm run tauri:dev`/`tauri:build`) — not for the backend/frontend web dev loop.

```bash
brew install sccache
# or, to build the latest release from source:
cargo install sccache --locked
```

If you'd rather skip it for a one-off build: `RUSTC_WRAPPER= cargo build` (matches what CI does).

No extra system libraries are needed for `make dmg` on macOS.

---

## Windows

[winget](https://learn.microsoft.com/en-us/windows/package-manager/winget/) is included with Windows 11 and recent Windows 10 builds. Run the commands below in **PowerShell** or **Windows Terminal**.

### Getting a Windows machine (VM)

If you don't have native Windows hardware, run a VM:

- **Linux host** — prefer QEMU/KVM over VirtualBox; it uses hardware virtualization directly and is noticeably faster. On Fedora/RHEL:
  ```bash
  sudo dnf install -y qemu-kvm libvirt virt-manager
  sudo systemctl enable --now libvirtd
  ```
  On Debian/Ubuntu: `sudo apt install -y qemu-kvm libvirt-daemon-system virt-manager`, then enable/start `libvirtd` the same way. Open `virt-manager`, create a new VM (use the Q35 chipset with UEFI + TPM 2.0 enabled — Windows 11 setup requires both) from a Windows ISO, or import Microsoft's free prebuilt [Windows 11 dev environment VM](https://developer.microsoft.com/en-us/windows/downloads/virtual-machines/) (90-day evaluation license, no product key needed).
- **macOS/Windows host** — VMware Fusion/Workstation, Parallels, or VirtualBox all work; import the same Microsoft dev VM image, or install from an ISO.

Once Windows is booted, install the toolchain below inside the VM.

### make

Windows does not ship with `make`. Install GnuWin32 make via winget:

```powershell
winget install --id GnuWin32.Make
```

Then add `C:\Program Files (x86)\GnuWin32\bin` to your `PATH` in **System Properties → Environment Variables**.

Alternatively, use [WSL](https://learn.microsoft.com/en-us/windows/wsl/install) (Windows Subsystem for Linux) and follow the Debian/Ubuntu instructions above — this is the recommended approach for a full Linux-compatible build environment on Windows.

#### Important: run `make` from Git Bash, not PowerShell/cmd

The Makefile's recipes use Unix shell syntax (`||`, `2>/dev/null`, single-quoted globs) — for example, the version stamp:
```make
VERSION = $(shell git describe --tags --always --match 'v*' 2>/dev/null || echo "dev")
```
GnuWin32 `make` needs a real POSIX `sh` on `PATH` to execute this correctly. If it can't find one (which is the case in a plain PowerShell or cmd session, even with Git for Windows installed), it silently falls back to interpreting recipes with `cmd.exe` instead — which mangles this line, `git describe` doesn't run as intended, and `VERSION` collapses to the literal string `"dev"`.

Symptoms of this happening:
- `The system cannot find the path specified.` printed near the very start of the build (`cmd` trying to treat `/dev/null` as a literal folder)
- `Stamped desktop version: dev`
- Tauri then fails with `` failed to parse config: `tauri.conf.json > version` must be a semver string `` — because `"dev"` isn't valid semver

**Fix:** run `make` targets from a **Git Bash** terminal instead — search "Git Bash" in the Start menu, or right-click the repo folder → **Git Bash Here**. Git for Windows ships its own `sh.exe`; once `make` is launched from a shell where that's on `PATH`, it correctly interprets the Makefile's Unix syntax and `git describe` produces a real version string.

#### Git Bash startup/config files

Git Bash launches as a **login shell**, so its startup lookup is:

1. System-wide first: `C:\Program Files\Git\etc\profile`, which sources anything in `C:\Program Files\Git\etc\profile.d\*.sh`
2. Then the first per-user file found (only one is read): `~/.bash_profile`, else `~/.bash_login`, else `~/.profile` — `~` is `C:\Users\<you>` (via `HOME`)

It does **not** automatically read `~/.bashrc` the way a non-login interactive bash would. Put aliases/functions directly in `~/.bash_profile`, or chain to `.bashrc` by adding this line to `~/.bash_profile`:
```bash
[ -f ~/.bashrc ] && . ~/.bashrc
```

PATH entries added via Windows' System Properties → Environment Variables (e.g. the GnuWin32 `make` fix above) don't need to be duplicated here — Git Bash's MSYS2 layer converts the Windows PATH into POSIX form automatically on startup.

### Go

```powershell
winget install --id GoLang.Go
```

After installation, open a new terminal so `%GOPATH%\bin` is on `PATH`.

### Node.js 20 LTS

```powershell
winget install --id OpenJS.NodeJS.LTS
```

### Rust

```powershell
winget install --id Rustlang.Rustup
```

Follow the on-screen prompts, then open a new terminal.

Rust on Windows links with MSVC by default, which Tauri requires. `rustup-init` detects if the MSVC linker is missing and offers to install it; you can also install it up front:

```powershell
winget install --id Microsoft.VisualStudio.2022.BuildTools --override "--wait --add Microsoft.VisualStudio.Workload.VCTools"
```

This installs the "Desktop development with C++" workload (several GB download) — required to link any Rust/Tauri binary.

#### Troubleshooting: `link.exe not found`

This means `rustc` can't locate the MSVC linker — usually the C++ workload isn't actually installed, even if the Build Tools installer itself ran (the silent `winget --override` can fail quietly, e.g. on a network hiccup).

1. Check whether the workload is really present:
   ```powershell
   & "${env:ProgramFiles(x86)}\Microsoft Visual Studio\Installer\vswhere.exe" -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath
   ```
   Empty output = the workload isn't installed.
2. Open the **Visual Studio Installer** GUI (Start menu) → **Modify** on Build Tools 2022 → check **Desktop development with C++** → confirm "MSVC v143 build tools" and a **Windows 10/11 SDK** are both selected under it → Modify/Install.
3. Open a brand-new terminal (or reboot) afterwards — `rustc` locates `link.exe` via the registry at process start, so an existing terminal won't pick up a fresh install.
4. Re-check with `cargo build --target x86_64-pc-windows-msvc -v` — `-v` shows the exact linker invocation if it still fails.

If `vswhere` still comes back empty after a Modify + reboot, add the Windows SDK component explicitly — the C++ workload can install without one in some minimal-override scenarios.

### sccache (recommended for desktop builds)

`frontend/src-tauri/.cargo/config.toml` wraps every Rust compile with `sccache` to cache compiled objects between builds — this applies to the Windows target too, so a local `npm run tauri:dev`/`tauri:build` (or `make windows-installer`) will fail to find it unless it's installed. Only needed for the desktop client, not the backend/frontend web dev loop:

```powershell
cargo install sccache --locked
# or
winget install --id Mozilla.sccache
```

If you'd rather skip it for a one-off build, clear the wrapper first: `$env:RUSTC_WRAPPER=""` (matches what CI does).

### WebView2 Runtime (desktop app testing)

Required to run the Tauri desktop app (`npm run tauri:dev`, or a built installer/portable zip). Pre-installed on Windows 11 and on Windows 10 builds from 2018 onward, so most VMs need nothing here. On a stripped-down or older image:

```powershell
winget install --id Microsoft.EdgeWebView2Runtime
```

### NSIS (Windows installer builds only)

Only needed when building the Windows installer (`make windows-installer`):

```powershell
winget install --id NSIS.NSIS
```

---

## Verify your installation

Run these after completing the steps above to confirm everything is in place:

```bash
make --version  # GNU Make 4.x ...
go version      # go version go1.26.x ...
node --version  # v20.x.x
npm --version   # 10.x.x
rustc --version # rustc 1.85.x ...
cargo --version # cargo 1.85.x ...
```
