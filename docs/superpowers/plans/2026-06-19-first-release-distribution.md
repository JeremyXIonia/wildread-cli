# First Release Distribution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a first-version release pipeline and user installation paths for `wildread-cli` using GitHub Releases, shell/PowerShell installers, and documented `go install`.

**Architecture:** GitHub Actions owns release builds triggered by version tags. Install scripts download the latest or requested GitHub Release artifact, install it under a user-writable bin directory, and print PATH guidance. Product docs describe recommended install paths and maintainer release steps.

**Tech Stack:** GitHub Actions, Go 1.25+, POSIX shell, PowerShell, GitHub Releases, SHA-256 checksums.

## Global Constraints

- Repository: `JeremyXIonia/wildread-cli`.
- Binary name: `wildread-cli` / `wildread-cli.exe`.
- Release artifacts:
  - `wildread-cli-darwin-amd64.tar.gz`
  - `wildread-cli-darwin-arm64.tar.gz`
  - `wildread-cli-windows-amd64.zip`
  - `checksums.txt`
- Release workflow triggers on tags matching `v*`.
- Install scripts install into user-writable directories by default, not system directories.
- macOS/Linux default install path: `$HOME/.local/bin/wildread-cli`.
- Windows default install path: `$HOME\bin\wildread-cli.exe`.
- Install scripts must not require admin privileges.
- README documents install script, manual download, and `go install github.com/JeremyXIonia/wildread-cli@latest`.
- `docs/release.md` documents the tag-based release flow.

---

## File Structure

- Create: `.github/workflows/release.yml` — tag-triggered test/build/package/release workflow.
- Create: `install.sh` — macOS/Linux installer.
- Create: `install.ps1` — Windows PowerShell installer.
- Create: `docs/release.md` — maintainer release process.
- Modify: `README.md` — user installation documentation.
- Modify: `.gitignore` if local release artifacts need to stay ignored.

---

### Task 1: GitHub Release Workflow

**Files:**
- Create: `.github/workflows/release.yml`

**Interfaces:**
- Produces GitHub Release assets with exact names from Global Constraints.

- [ ] **Step 1: Create release workflow**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    name: Build and publish release
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25.x'
          cache: true

      - name: Test
        run: go test ./... -v

      - name: Build artifacts
        shell: bash
        run: |
          set -euo pipefail
          mkdir -p dist
          LDFLAGS="-s -w"

          GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/wildread-cli .
          tar -C dist -czf dist/wildread-cli-darwin-amd64.tar.gz wildread-cli
          rm dist/wildread-cli

          GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/wildread-cli .
          tar -C dist -czf dist/wildread-cli-darwin-arm64.tar.gz wildread-cli
          rm dist/wildread-cli

          GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/wildread-cli.exe .
          (cd dist && zip wildread-cli-windows-amd64.zip wildread-cli.exe)
          rm dist/wildread-cli.exe

          (cd dist && sha256sum wildread-cli-darwin-amd64.tar.gz wildread-cli-darwin-arm64.tar.gz wildread-cli-windows-amd64.zip > checksums.txt)

      - name: Publish GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            dist/wildread-cli-darwin-amd64.tar.gz
            dist/wildread-cli-darwin-arm64.tar.gz
            dist/wildread-cli-windows-amd64.zip
            dist/checksums.txt
```

- [ ] **Step 2: Verify YAML is present**

Run:

```bash
test -f .github/workflows/release.yml
grep -n "wildread-cli-darwin-amd64.tar.gz\|wildread-cli-windows-amd64.zip\|checksums.txt" .github/workflows/release.yml
```

Expected: all artifact names appear.

---

### Task 2: macOS/Linux Installer

**Files:**
- Create: `install.sh`

**Interfaces:**
- Installs `wildread-cli` to `${WILDREAD_INSTALL_DIR:-$HOME/.local/bin}`.
- Accepts optional version argument, default `latest`.

- [ ] **Step 1: Create install.sh**

Create `install.sh`:

```sh
#!/bin/sh
set -eu

REPO="JeremyXIonia/wildread-cli"
BIN="wildread-cli"
VERSION="${1:-latest}"
INSTALL_DIR="${WILDREAD_INSTALL_DIR:-$HOME/.local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

case "$os" in
  darwin) os="darwin" ;;
  linux) echo "Linux release artifacts are not published yet. Use: go install github.com/JeremyXIonia/wildread-cli@latest" >&2; exit 1 ;;
  *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

asset="$BIN-$os-$arch.tar.gz"
base="https://github.com/$REPO/releases"
if [ "$VERSION" = "latest" ]; then
  url="$base/latest/download/$asset"
else
  url="$base/download/$VERSION/$asset"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

mkdir -p "$INSTALL_DIR"
echo "Downloading $url"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$url" -o "$tmp/$asset"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$url" -O "$tmp/$asset"
else
  echo "curl or wget is required" >&2
  exit 1
fi

tar -xzf "$tmp/$asset" -C "$tmp"
install -m 0755 "$tmp/$BIN" "$INSTALL_DIR/$BIN"

echo "Installed $BIN to $INSTALL_DIR/$BIN"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "Add $INSTALL_DIR to PATH to run '$BIN' from anywhere." ;;
esac
```

- [ ] **Step 2: Make executable and syntax-check**

Run:

```bash
chmod +x install.sh
sh -n install.sh
```

Expected: no output.

---

### Task 3: Windows PowerShell Installer

**Files:**
- Create: `install.ps1`

**Interfaces:**
- Installs `wildread-cli.exe` to `$env:USERPROFILE\bin` by default.
- Accepts optional `-Version`, default `latest`.

- [ ] **Step 1: Create install.ps1**

Create `install.ps1`:

```powershell
param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:USERPROFILE\bin"
)

$ErrorActionPreference = "Stop"
$Repo = "JeremyXIonia/wildread-cli"
$Asset = "wildread-cli-windows-amd64.zip"

if ($Version -eq "latest") {
    $Url = "https://github.com/$Repo/releases/latest/download/$Asset"
} else {
    $Url = "https://github.com/$Repo/releases/download/$Version/$Asset"
}

$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $TempDir | Out-Null
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

try {
    $ZipPath = Join-Path $TempDir $Asset
    Write-Host "Downloading $Url"
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath

    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force
    $ExePath = Join-Path $TempDir "wildread-cli.exe"
    Copy-Item -Path $ExePath -Destination (Join-Path $InstallDir "wildread-cli.exe") -Force

    Write-Host "Installed wildread-cli.exe to $InstallDir"
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (($UserPath -split ';') -notcontains $InstallDir) {
        Write-Host "Add $InstallDir to your user PATH to run wildread-cli from anywhere."
        Write-Host "You can run: [Environment]::SetEnvironmentVariable('Path', `$env:Path + ';$InstallDir', 'User')"
    }
} finally {
    Remove-Item -Recurse -Force $TempDir -ErrorAction SilentlyContinue
}
```

- [ ] **Step 2: PowerShell parse check**

Run:

```bash
pwsh -NoProfile -Command '$null = [System.Management.Automation.Language.Parser]::ParseFile("install.ps1", [ref]$null, [ref]$errs); if ($errs.Count) { $errs | ForEach-Object { $_.Message }; exit 1 }'
```

Expected: exit 0. If `pwsh` is unavailable locally, record that parser check was skipped and manually inspect the script.

---

### Task 4: User and Maintainer Documentation

**Files:**
- Modify: `README.md`
- Create: `docs/release.md`

**Interfaces:**
- Documents install script, manual download, `go install`, and tag release process.

- [ ] **Step 1: Update README installation section**

Replace the install section with:

```markdown
## 安装

### 推荐：安装脚本

macOS：

```bash
curl -fsSL https://raw.githubusercontent.com/JeremyXIonia/wildread-cli/master/install.sh | sh
```

Windows PowerShell：

```powershell
iwr https://raw.githubusercontent.com/JeremyXIonia/wildread-cli/master/install.ps1 -UseB | iex
```

安装脚本默认安装到用户目录，不需要管理员权限：

- macOS: `~/.local/bin/wildread-cli`
- Windows: `%USERPROFILE%\bin\wildread-cli.exe`

### Go 用户

```bash
go install github.com/JeremyXIonia/wildread-cli@latest
```

### 手动下载

从 [GitHub Releases](https://github.com/JeremyXIonia/wildread-cli/releases) 下载对应平台压缩包，解压后将 `wildread-cli` 或 `wildread-cli.exe` 放入 PATH。
```

Keep the build-from-source section after installation.

- [ ] **Step 2: Add release docs**

Create `docs/release.md`:

```markdown
# Release

## 发布新版本

确保当前在 `master` 且测试通过：

```bash
go test ./...
```

创建并推送 tag：

```bash
git tag v0.1.0
git push origin v0.1.0
```

推送 tag 后，GitHub Actions 会自动：

- 运行测试
- 构建 macOS Intel、macOS Apple Silicon、Windows amd64 二进制
- 打包 Release artifacts
- 生成 `checksums.txt`
- 创建 GitHub Release

## 版本号

使用 SemVer 风格 tag：

- `v0.1.0`
- `v0.2.0`
- `v1.0.0`

## Release artifacts

每次 release 应包含：

- `wildread-cli-darwin-amd64.tar.gz`
- `wildread-cli-darwin-arm64.tar.gz`
- `wildread-cli-windows-amd64.zip`
- `checksums.txt`
```

- [ ] **Step 3: Verify docs mention install paths**

Run:

```bash
grep -n "install.sh\|install.ps1\|go install github.com/JeremyXIonia/wildread-cli@latest\|GitHub Releases" README.md docs/release.md
```

Expected: matching lines are printed.

---

### Task 5: Verification and Commit

**Files:**
- All modified files

- [ ] **Step 1: Run tests and syntax checks**

Run:

```bash
go test ./... -v
bash -n build.sh
sh -n install.sh
```

Expected: PASS/no output for syntax checks.

If `pwsh` is installed, also run:

```bash
pwsh -NoProfile -Command '$null = [System.Management.Automation.Language.Parser]::ParseFile("install.ps1", [ref]$null, [ref]$errs); if ($errs.Count) { $errs | ForEach-Object { $_.Message }; exit 1 }'
```

- [ ] **Step 2: Review diff**

Run:

```bash
git diff --stat
git status --short
```

Expected: release workflow, install scripts, README, release docs, and this plan file.

- [ ] **Step 3: Commit**

Run:

```bash
git add .github/workflows/release.yml install.sh install.ps1 README.md docs/release.md docs/superpowers/plans/2026-06-19-first-release-distribution.md
git commit -m "feat: 添加首版发布安装流程" -m "Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Self-Review

- Spec coverage: release workflow, install.sh, install.ps1, README install docs, and docs/release.md are covered.
- Placeholder scan: no placeholders remain.
- Type consistency: artifact names match between workflow, scripts, README, and release docs.
