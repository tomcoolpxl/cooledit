#!/usr/bin/env node

const crypto = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const https = require("node:https");
const { pipeline } = require("node:stream/promises");

const AdmZip = require("adm-zip");
const tar = require("tar");

const pkg = require("../package.json");

const REPO = "tomcoolpxl/cooledit";
const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows"
};
const ARCH_MAP = {
  arm64: "arm64",
  x64: "amd64"
};

function fail(message) {
  console.error(`cooledit: ${message}`);
  process.exit(1);
}

function platformInfo() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    fail(`unsupported platform ${process.platform}/${process.arch}`);
  }

  return { platform, arch };
}

function releaseVersion() {
  if (!pkg.version || pkg.version === "0.0.0-dev") {
    fail("package version is not set to a releasable version");
  }

  return {
    plain: pkg.version,
    tag: `v${pkg.version}`
  };
}

function archiveDetails(platform, arch, version) {
  const isWindows = platform === "windows";
  const extension = isWindows ? "zip" : "tar.gz";
  const binaryName = isWindows ? "cooledit.exe" : "cooledit";
  const archiveName = `cooledit_${version}_${platform}_${arch}.${extension}`;

  return { archiveName, binaryName, extension };
}

function request(url) {
  return new Promise((resolve, reject) => {
    https.get(
      url,
      {
        headers: {
          "user-agent": "cooledit-npm-installer"
        }
      },
      (response) => {
        if (
          response.statusCode &&
          response.statusCode >= 300 &&
          response.statusCode < 400 &&
          response.headers.location
        ) {
          response.resume();
          resolve(request(response.headers.location));
          return;
        }

        if (response.statusCode !== 200) {
          response.resume();
          reject(new Error(`download failed with status ${response.statusCode} for ${url}`));
          return;
        }

        resolve(response);
      }
    ).on("error", reject);
  });
}

async function downloadFile(url, destination) {
  const response = await request(url);
  await pipeline(response, fs.createWriteStream(destination));
}

async function downloadText(url) {
  const response = await request(url);
  const chunks = [];
  for await (const chunk of response) {
    chunks.push(chunk);
  }
  return Buffer.concat(chunks).toString("utf8");
}

function expectedChecksum(checksums, archiveName) {
  const line = checksums
    .split("\n")
    .map((entry) => entry.trim())
    .find((entry) => entry.endsWith(archiveName));

  if (!line) {
    fail(`checksum entry not found for ${archiveName}`);
  }

  return line.split(/\s+/)[0];
}

function actualChecksum(filePath) {
  const hash = crypto.createHash("sha256");
  hash.update(fs.readFileSync(filePath));
  return hash.digest("hex");
}

async function extractArchive(archivePath, extension, extractDir) {
  if (extension === "zip") {
    const zip = new AdmZip(archivePath);
    zip.extractAllTo(extractDir, true);
    return;
  }

  await tar.x({
    file: archivePath,
    cwd: extractDir
  });
}

function findBinary(rootDir, binaryName) {
  const entries = fs.readdirSync(rootDir, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = path.join(rootDir, entry.name);
    if (entry.isDirectory()) {
      const nested = findBinary(fullPath, binaryName);
      if (nested) {
        return nested;
      }
      continue;
    }

    if (entry.isFile() && entry.name === binaryName) {
      return fullPath;
    }
  }

  return null;
}

async function main() {
  if (process.env.COOLEDIT_SKIP_BINARY_DOWNLOAD === "1") {
    console.log("cooledit: skipping binary download because COOLEDIT_SKIP_BINARY_DOWNLOAD=1");
    return;
  }

  const { platform, arch } = platformInfo();
  const version = releaseVersion();
  const details = archiveDetails(platform, arch, version.plain);

  const vendorDir = path.join(__dirname, "vendor", `${platform}-${arch}`);
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), "cooledit-npm-"));
  const archivePath = path.join(tempDir, details.archiveName);
  const extractDir = path.join(tempDir, "extract");
  const binaryTarget = path.join(vendorDir, details.binaryName);
  const baseUrl = `https://github.com/${REPO}/releases/download/${version.tag}`;

  fs.mkdirSync(extractDir, { recursive: true });

  try {
    console.log(`cooledit: downloading ${version.tag} for ${platform}/${arch}`);

    await downloadFile(`${baseUrl}/${details.archiveName}`, archivePath);
    const checksums = await downloadText(`${baseUrl}/checksums.txt`);

    const expected = expectedChecksum(checksums, details.archiveName);
    const actual = actualChecksum(archivePath);

    if (expected !== actual) {
      fail(`checksum mismatch for ${details.archiveName}`);
    }

    await extractArchive(archivePath, details.extension, extractDir);

    const extractedBinary = findBinary(extractDir, details.binaryName);
    if (!extractedBinary) {
      fail(`could not locate ${details.binaryName} in downloaded archive`);
    }

    fs.rmSync(vendorDir, { recursive: true, force: true });
    fs.mkdirSync(vendorDir, { recursive: true });
    fs.copyFileSync(extractedBinary, binaryTarget);

    if (process.platform !== "win32") {
      fs.chmodSync(binaryTarget, 0o755);
    }

    console.log(`cooledit: installed binary to ${binaryTarget}`);
  } catch (error) {
    fail(error.message);
  } finally {
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
}

main();
