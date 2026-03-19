#!/usr/bin/env node

const fs = require("node:fs");
const path = require("node:path");
const { spawn } = require("node:child_process");

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows"
};

const ARCH_MAP = {
  arm64: "arm64",
  x64: "amd64"
};

function resolveBinaryPath() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    console.error(`cooledit: unsupported platform ${process.platform}/${process.arch}`);
    process.exit(1);
  }

  const binaryName = process.platform === "win32" ? "cooledit.exe" : "cooledit";
  return path.join(__dirname, "vendor", `${platform}-${arch}`, binaryName);
}

const binaryPath = resolveBinaryPath();

if (!fs.existsSync(binaryPath)) {
  console.error("cooledit: bundled binary is missing");
  console.error("Reinstall the package with `npm install -g cooledit`.");
  process.exit(1);
}

const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: "inherit"
});

child.on("error", (error) => {
  console.error(`cooledit: failed to launch binary: ${error.message}`);
  process.exit(1);
});

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }

  process.exit(code ?? 0);
});
