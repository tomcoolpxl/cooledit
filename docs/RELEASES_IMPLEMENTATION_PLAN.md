# cooledit Release Pipeline — Implementation Plan

This document translates the architecture in `releases.md` into concrete, step-by-step
implementation tasks. Every file that needs to be created or modified is listed with
its full contents or exact changes.

**Target release**: v0.8.0
**Starting state**: no CI, no GoReleaser config, no install script, no WinGet manifest

---

## Table of Contents

1. [Prerequisites and one-time repository setup](#1-prerequisites-and-one-time-repository-setup)
2. [Version embedding](#2-version-embedding)
3. [GoReleaser configuration](#3-goreleaser-configuration)
4. [GitHub Actions workflows](#4-github-actions-workflows)
5. [install.sh — Linux/macOS installer](#5-installsh--linuxmacos-installer)
6. [WinGet manifest](#6-winget-manifest)
7. [README installation section](#7-readme-installation-section)
8. [Release workflow (end-to-end steps)](#8-release-workflow-end-to-end-steps)
9. [Maintenance rules](#9-maintenance-rules)

---

## 1. Prerequisites and one-time repository setup

### 1.1 Fix the version discrepancy

`cmd/cooledit/main.go` currently hard-codes `version = "0.1.0"` but `CHANGELOG.md` shows
`0.5.0` as the last release. The hard-coded constant will be replaced by build-time
injection (see §2), so the constant itself just needs a sensible placeholder:

```go
// cmd/cooledit/main.go — change:
const version = "0.1.0"
// to:
var version = "dev" // overridden at build time via -ldflags
```

The real value is injected by GoReleaser at build time.

### 1.2 GitHub repository

The repository must be hosted on GitHub under a predictable path, e.g.:

```
https://github.com/tomcoolpxl/cooledit
```

Replace `tomcoolpxl` with the actual GitHub username/org throughout this document.

### 1.3 No token setup required

The workflow uses `secrets.GITHUB_TOKEN`, which GitHub Actions injects automatically
into every workflow run at no cost. No PAT, no extra secrets, no account upgrade needed.
The `permissions: contents: write` declaration in `release.yml` is all that is required.

### 1.4 Install GoReleaser locally (for testing)

```sh
# macOS
brew install goreleaser/tap/goreleaser

# Linux (one-shot)
curl -sfL https://goreleaser.com/static/run | bash

# or via go install (v2)
go install github.com/goreleaser/goreleaser/v2@latest
```

---

## 2. Version embedding

Replace the hard-coded `version` constant with a build-time variable injected via
`-ldflags`. GoReleaser sets `{{ .Version }}` automatically.

### 2.1 Change `cmd/cooledit/main.go`

```go
// Replace:
const version = "0.1.0"

// With:
var version = "dev" // overridden at build time via -ldflags
```

### 2.2 Local dev build (manual)

```sh
go build -ldflags "-X main.version=$(git describe --tags --always --dirty)" \
    -o cooledit ./cmd/cooledit
```

The GoReleaser config (§3) handles this automatically for releases.

---

## 3. GoReleaser configuration

Create `.goreleaser.yaml` in the repository root.

```yaml
# .goreleaser.yaml
version: 2

project_name: cooledit

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - id: cooledit
    main: ./cmd/cooledit
    binary: cooledit
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      # Skip windows/arm64 until there is demand
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w -X main.version={{ .Version }}

archives:
  - id: cooledit-archive
    builds:
      - cooledit
    name_template: "cooledit_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        formats: [zip]
    files:
      - LICENSE
      - README.md

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

release:
  github:
    owner: tomcoolpxl        # replace with actual GitHub username/org
    name: cooledit
  draft: false
  prerelease: auto        # tags like v0.8.0-rc1 become pre-releases automatically
  name_template: "cooledit {{ .Version }}"
  footer: |
    ## Install

    **Linux/macOS (curl | sh)**
    ```sh
    curl -fsSL https://raw.githubusercontent.com/tomcoolpxl/cooledit/main/install.sh | sh
    ```

    **Windows (WinGet)**
    ```
    winget install cooledit.cooledit
    ```

    **Verify checksum**
    ```sh
    sha256sum -c checksums.txt
    ```

changelog:
  sort: asc
  use: github
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
      - Merge pull request
      - Merge branch
```

### 3.1 Test locally before first real tag

```sh
goreleaser release --snapshot --clean
```

Artifacts appear in `dist/`. Verify binary names, archive contents, and checksums.txt
before creating the first real tag.

---

## 4. GitHub Actions workflows

Create the `.github/workflows/` directory with two workflow files.

### 4.1 CI workflow — runs on every push and PR

**File**: `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: ["main"]
    tags-ignore: ["v*"]      # releases handled by release.yml
  pull_request:
    branches: ["main"]

jobs:
  test:
    name: Test
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest]
        go: ["1.25.x"]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test ./...

      - name: Build
        run: go build -o /dev/null ./cmd/cooledit   # linux/mac
        if: runner.os != 'Windows'

      - name: Build (Windows)
        run: go build -o NUL ./cmd/cooledit
        if: runner.os == 'Windows'
```

### 4.2 Release workflow — runs on version tag push

**File**: `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags:
      - "v[0-9]*.[0-9]*.[0-9]*"

permissions:
  contents: write     # needed to create GitHub Release and upload assets

jobs:
  release:
    name: GoReleaser
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0    # required for GoReleaser changelog generation

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Run tests
        run: go test ./...

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 5. install.sh — Linux/macOS installer

Create `install.sh` in the repository root. This is what users pipe into `sh`.

```sh
#!/usr/bin/env sh
# cooledit installer
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/tomcoolpxl/cooledit/main/install.sh | sh
#   COOLEDIT_VERSION=v0.8.1 curl -fsSL ... | sh
set -eu

REPO="tomcoolpxl/cooledit"
BINARY="cooledit"
INSTALL_DIR=""

# ── helpers ──────────────────────────────────────────────────────────────────

die() { echo "error: $*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "required tool not found: $1"
}

# ── detect OS and architecture ────────────────────────────────────────────────

detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Linux)  OS="linux" ;;
    Darwin) OS="darwin" ;;
    *)      die "unsupported OS: $OS" ;;
  esac

  case "$ARCH" in
    x86_64)          ARCH="amd64" ;;
    aarch64|arm64)   ARCH="arm64" ;;
    *)               die "unsupported architecture: $ARCH" ;;
  esac
}

# ── resolve version ───────────────────────────────────────────────────────────

resolve_version() {
  if [ -n "${COOLEDIT_VERSION:-}" ]; then
    VERSION="$COOLEDIT_VERSION"
    return
  fi

  need curl

  VERSION="$(curl -fsSL \
    "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"

  [ -n "$VERSION" ] || die "could not determine latest version"
}

# ── resolve install dir ───────────────────────────────────────────────────────

resolve_install_dir() {
  if [ "$(id -u)" = "0" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
  fi
}

# ── download and verify ───────────────────────────────────────────────────────

install_binary() {
  need curl
  need sha256sum

  ARCHIVE_NAME="cooledit_${VERSION#v}_${OS}_${ARCH}.tar.gz"
  BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
  TMPDIR="$(mktemp -d)"
  trap 'rm -rf "$TMPDIR"' EXIT

  echo "Downloading cooledit ${VERSION} for ${OS}/${ARCH}..."
  curl -fsSL "${BASE_URL}/${ARCHIVE_NAME}" -o "${TMPDIR}/${ARCHIVE_NAME}"
  curl -fsSL "${BASE_URL}/checksums.txt"   -o "${TMPDIR}/checksums.txt"

  echo "Verifying checksum..."
  # checksums.txt contains lines like: <hash>  cooledit_VERSION_OS_ARCH.tar.gz
  # grep for just our archive name, then verify
  (
    cd "$TMPDIR"
    grep "${ARCHIVE_NAME}" checksums.txt | sha256sum -c -
  ) || die "checksum verification failed"

  echo "Extracting..."
  tar -xzf "${TMPDIR}/${ARCHIVE_NAME}" -C "$TMPDIR"

  echo "Installing to ${INSTALL_DIR}/${BINARY}..."
  install -m 755 "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

  echo ""
  echo "cooledit ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"

  # Warn if install dir is not in PATH
  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *) echo "NOTE: Add ${INSTALL_DIR} to your PATH to use cooledit" ;;
  esac
}

# ── main ──────────────────────────────────────────────────────────────────────

detect_platform
resolve_version
resolve_install_dir
install_binary
```

Make it executable in the repo:

```sh
chmod +x install.sh
git add install.sh
```

### 5.1 Usage variants

```sh
# Latest release
curl -fsSL https://raw.githubusercontent.com/tomcoolpxl/cooledit/main/install.sh | sh

# Specific version
COOLEDIT_VERSION=v0.8.0 curl -fsSL \
  https://raw.githubusercontent.com/tomcoolpxl/cooledit/main/install.sh | sh

# As root (installs to /usr/local/bin)
curl -fsSL https://raw.githubusercontent.com/tomcoolpxl/cooledit/main/install.sh | sudo sh
```

---

## 6. WinGet manifest

WinGet manifests live in the
[`microsoft/winget-pkgs`](https://github.com/microsoft/winget-pkgs) repository under
`manifests/<first-letter>/<publisher>/<package>/<version>/`.

### 6.1 Package identifier

```
cooledit.cooledit
```

Manifests live at:
```
manifests/c/cooledit/cooledit/<version>/
```

### 6.2 Three manifest files per release

Every WinGet submission requires three YAML files.

#### `cooledit.cooledit.installer.yaml`

```yaml
PackageIdentifier: cooledit.cooledit
PackageVersion: 0.8.0
Platform:
  - Windows.Desktop
MinimumOSVersion: "10.0.0.0"
InstallerType: zip
NestedInstallerType: portable
NestedInstallerFiles:
  - RelativeFilePath: cooledit.exe
    PortableCommandAlias: cooledit
Installers:
  - Architecture: x64
    InstallerUrl: https://github.com/tomcoolpxl/cooledit/releases/download/v0.8.0/cooledit_0.8.0_windows_amd64.zip
    InstallerSha256: <SHA256-of-windows-amd64-zip>  # from checksums.txt
ManifestType: installer
ManifestVersion: 1.6.0
```

#### `cooledit.cooledit.locale.en-US.yaml`

```yaml
PackageIdentifier: cooledit.cooledit
PackageVersion: 0.8.0
PackageLocale: en-US
Publisher: cooledit
PackageName: cooledit
License: GPL-3.0-or-later
LicenseUrl: https://github.com/tomcoolpxl/cooledit/blob/main/LICENSE
ShortDescription: A terminal-based text editor with modern keyboard shortcuts and syntax highlighting
Description: >
  cooledit is a terminal-based text editor inspired by nano but with better UI,
  keyboard shortcuts, syntax highlighting, and theme support.
Tags:
  - terminal
  - text-editor
  - cli
  - editor
ReleaseNotesUrl: https://github.com/tomcoolpxl/cooledit/releases/tag/v0.8.0
ManifestType: defaultLocale
ManifestVersion: 1.6.0
```

#### `cooledit.cooledit.yaml`

```yaml
PackageIdentifier: cooledit.cooledit
PackageVersion: 0.8.0
DefaultLocale: en-US
ManifestType: version
ManifestVersion: 1.6.0
```

### 6.3 Submission process per release

1. After the GitHub Release is published, get the SHA256 for the Windows amd64 zip from
   the release's `checksums.txt`:
   ```sh
   curl -fsSL https://github.com/tomcoolpxl/cooledit/releases/download/v0.8.0/checksums.txt \
     | grep windows_amd64.zip
   ```

2. Fork `microsoft/winget-pkgs` on GitHub.

3. Create the three manifest files under:
   ```
   manifests/c/cooledit/cooledit/0.8.0/
   ```

4. Validate locally with the WinGet validator (optional but recommended):
   ```sh
   winget validate --manifest manifests/c/cooledit/cooledit/0.8.0/
   ```

5. Open a PR to `microsoft/winget-pkgs`. Title format:
   ```
   New package: cooledit.cooledit version 0.8.0
   ```

6. The WinGet bot validates automatically. Merge typically takes 1–3 business days.

### 6.4 Automating WinGet submission (optional, future)

The [WinGet automation action](https://github.com/vedantmgoyal9/winget-releaser) can
submit manifests automatically from the release workflow. Add it to `release.yml` after
GoReleaser succeeds if the manual process becomes tedious.

---

## 7. README installation section

Replace the current placeholder "Download from releases page." with:

```markdown
## Installation

### Linux / macOS

```sh
curl -fsSL https://raw.githubusercontent.com/tomcoolpxl/cooledit/main/install.sh | sh
```

Install a specific version:

```sh
COOLEDIT_VERSION=v0.8.0 curl -fsSL \
  https://raw.githubusercontent.com/tomcoolpxl/cooledit/main/install.sh | sh
```

The script installs to `~/.local/bin` (non-root) or `/usr/local/bin` (root) and
verifies the SHA256 checksum automatically.

### Windows

```
winget install cooledit.cooledit
```

### Manual download

Download the archive for your platform from the
[Releases page](https://github.com/tomcoolpxl/cooledit/releases), extract, and place
the `cooledit` binary somewhere on your `PATH`.

### Build from source

```sh
git clone https://github.com/tomcoolpxl/cooledit.git
cd cooledit
go build -o cooledit ./cmd/cooledit
```

Requires Go 1.25 or later.
```

---

## 8. Release workflow (end-to-end steps)

These are the exact steps a developer takes for each release.

### 8.1 Prepare the release

```sh
# 1. Ensure all tests pass
go test ./...

# 2. Update CHANGELOG.md
#    Add a new [X.Y.Z] section at the top with today's date
#    Follow Keep a Changelog format

# 3. Update the version placeholder comment in main.go if desired
#    (the actual version is injected by ldflags — no code change required)

# 4. Commit everything
git add CHANGELOG.md
git commit -m "chore: prepare release v0.8.0"
git push origin main
```

### 8.2 Tag and push

```sh
git tag -a v0.8.0 -m "Release v0.8.0"
git push origin v0.8.0
```

This triggers `.github/workflows/release.yml` which:
1. Runs `go test ./...`
2. Runs GoReleaser → builds all targets, packages archives, generates checksums.txt
3. Creates and publishes the GitHub Release with all assets

### 8.3 Verify the release

```sh
# 1. Check GitHub Release page — all 3 archives present + checksums.txt

# 2. Test the installer
curl -fsSL https://raw.githubusercontent.com/tomcoolpxl/cooledit/main/install.sh | sh
cooledit --version

# 3. Test the Windows zip manually if possible
```

### 8.4 Submit WinGet manifest

Follow §6.3. Do this after verifying the GitHub Release is live.

### 8.5 If the release fails

```sh
# Delete the tag locally and remotely, fix the issue, retag
git tag -d v0.8.0
git push origin :refs/tags/v0.8.0

# Fix the issue
# ...

# Retag
git tag -a v0.8.0 -m "Release v0.8.0"
git push origin v0.8.0
```

Note: if a GitHub Release was partially created before failure, delete it from the
GitHub Releases UI before retagging to avoid asset conflicts.

---

## 9. Maintenance rules

These rules must be followed to avoid breaking existing installs and WinGet references.

| Rule | Reason |
|------|--------|
| Never rename archive artifacts between versions | install.sh and WinGet manifests use predictable names |
| Never delete published releases | WinGet manifests reference specific release URLs permanently |
| Never change a published checksums.txt | Verification breaks for users who cached it |
| Tags are the only release trigger | Never run GoReleaser from main branch pushes |
| Patch releases (v0.8.1) for critical bugs | Don't bundle unrelated changes into a hotfix |
| Always test `goreleaser release --snapshot --clean` before tagging | Catches build problems before they create a broken release |

### 9.1 File summary — what gets created

| File | Location | Purpose |
|------|----------|---------|
| `.goreleaser.yaml` | repo root | Build matrix, archive config, release settings |
| `.github/workflows/ci.yml` | repo | Tests on every push/PR |
| `.github/workflows/release.yml` | repo | Automated release on tag push |
| `install.sh` | repo root | curl-pipe installer for Linux/macOS |
| `manifests/c/cooledit/cooledit/<ver>/` | `microsoft/winget-pkgs` fork | WinGet distribution (per release) |

### 9.2 Secrets required

No secrets need to be manually created. The workflow uses `secrets.GITHUB_TOKEN`,
which GitHub provides automatically to every Actions run.
