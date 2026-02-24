Below is a structured publication plan for cooledit v0.8 using:

* GitHub Releases as the distribution source of truth
* GoReleaser for automation
* A curl | sh installer (Linux + optional macOS)
* WinGet for Windows

This is architecture and process only, not implementation details.

---

# High-Level Architecture

Single source of truth:

* Main repository (source, go.mod, CI)
* GitHub Releases containing:

  * Versioned binaries per OS/arch
  * SHA256 checksums file
  * Signed checksums (optional but recommended)

Everything else (installer script, WinGet) points to those release assets.

No separate “publication repo” required except the WinGet manifest PR to Microsoft’s repository.

---

# Versioning and Release Strategy

## Versioning Model

* Semantic Versioning: v0.8.0, v0.8.1, etc.
* Git tags drive releases.
* Tag format: vX.Y.Z (important for automation compatibility).

Decision:

* No "latest" mutable binary links. Always versioned.
* GitHub “latest release” is just convenience, not a dependency target.

---

# Build and Artifact Strategy

## Target Matrix

Define explicitly:

Linux:

* linux_amd64
* linux_arm64

Windows:

* windows_amd64
* optionally windows_arm64 (if relevant)

Keep the matrix minimal unless you have a reason to expand.

## Binary Strategy

* Fully static builds (CGO disabled unless required).
* Version embedded at build time.
* Deterministic file names:

  * cooledit_<version>*<os>*<arch>.tar.gz (Linux)
  * cooledit_<version>_windows_amd64.zip

Inside archive:

* cooledit binary
* LICENSE
* README (optional)

---

# GitHub Releases Plan

Each release contains:

* All archives for all targets
* checksums.txt (SHA256 for each artifact)
* Optionally:

  * checksums.txt.sig (GPG signature)

Release notes:

* Short changelog
* Install instructions:

  * curl | sh
  * winget install cooledit

This page becomes your official distribution entry point.

---

# GoReleaser Plan

GoReleaser responsibilities:

1. Cross-compile for defined OS/arch matrix.
2. Package binaries into archives.
3. Generate SHA256 checksums.
4. Create GitHub Release.
5. Upload artifacts.
6. Optionally sign checksums.

Triggered by:

* Git tag push: vX.Y.Z

CI flow:

* On tag:

  * Run tests
  * Run GoReleaser
  * Publish release

Failure strategy:

* If release fails, delete tag and retag (document this in your internal process).

---

# curl | sh Installer Plan (Linux)

Purpose:

* Zero package manager dependency.
* Works on most distros.
* Simple, deterministic, auditable.

## Installer Responsibilities

1. Detect OS and architecture.
2. Resolve latest version OR install specific version.
3. Download matching archive from GitHub Release.
4. Download checksums.txt.
5. Verify SHA256 of downloaded archive.
6. Extract binary.
7. Install to:

   * ~/.local/bin (default non-root)
   * /usr/local/bin (if root)
8. Print post-install path notice.

## Security Model

* Always verify SHA256.
* Do not skip checksum verification.
* Consider optional:

  * GPG verification of checksums.txt.

## Script Hosting

Two acceptable approaches:

Option A:

* Store install.sh in repo root.
* Users curl from raw GitHub URL.

Option B:

* Serve from GitHub Pages (cleaner public endpoint).

Do not hardcode unstable GitHub “latest asset” URLs. Always resolve via GitHub API or documented predictable pattern.

---

# Windows Strategy: WinGet

WinGet is the official Windows package manager and should be primary.

## Distribution Model

You do not upload binaries to WinGet.
WinGet manifest references your GitHub Release URLs.

## Requirements

You must provide:

* HTTPS download URL
* SHA256 of installer/archive
* Versioned file
* Stable naming

WinGet manifest includes:

* Package identifier (e.g., cooledit.cooledit)
* Version
* Installer type:

  * zip (portable)
  * exe (if you later create an installer)

For a CLI, portable zip is sufficient.

## Process Per Release

For each new version:

1. Release published on GitHub.
2. Generate new WinGet manifest.
3. Submit PR to:

   * microsoft/winget-pkgs repository.
4. Wait for validation + merge.
5. Users can run:

   * winget install cooledit
   * winget upgrade cooledit

This is semi-manual unless automated via GitHub Actions.

## Stability Considerations

* Never change asset URLs after publishing.
* Never delete old releases that are referenced by WinGet.

---

# Release Workflow (End-to-End)

Developer steps:

1. Update version references.
2. Commit.
3. Tag: v0.8.0
4. Push tag.
5. CI runs GoReleaser.
6. GitHub Release created with binaries + checksums.
7. Update WinGet manifest and submit PR.
8. Done.

No separate repo required except interacting with winget-pkgs.

---

# Maintenance Strategy

## Backward Compatibility

* Never rename artifacts between versions.
* Keep naming stable forever.

## Deprecation

* If changing archive layout:

  * Introduce change in minor version.
  * Update installer script first.
  * Then release.

## Automation Safety

* Do not auto-publish from main branch without tag.
* Tags are the only release trigger.

---

# Future Expansion (Optional)

Later, you can add:

* Scoop (easy once GitHub releases exist).
* deb/rpm packages via GoReleaser.
* Snap.

None of these require changing your core architecture.

---

# Final Structural Summary

You do not need:

* npm
* cargo
* a separate publication repository

You need:

1. One canonical GitHub repository.
2. GoReleaser configuration.
3. One install.sh script.
4. WinGet manifest submissions per release.
