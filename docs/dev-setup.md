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
    squashfs-tools
```

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
    squashfs-tools
```

---

## macOS

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

No extra system libraries are needed for `make dmg` on macOS.

---

## Windows

[winget](https://learn.microsoft.com/en-us/windows/package-manager/winget/) is included with Windows 11 and recent Windows 10 builds. Run the commands below in **PowerShell** or **Windows Terminal**.

### make

Windows does not ship with `make`. Install GnuWin32 make via winget:

```powershell
winget install --id GnuWin32.Make
```

Then add `C:\Program Files (x86)\GnuWin32\bin` to your `PATH` in **System Properties → Environment Variables**.

Alternatively, use [WSL](https://learn.microsoft.com/en-us/windows/wsl/install) (Windows Subsystem for Linux) and follow the Debian/Ubuntu instructions above — this is the recommended approach for a full Linux-compatible build environment on Windows.

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
