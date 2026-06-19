# Installer PATH Guidance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Update installer output so users receive copyable PATH configuration examples after installing `wildread-cli`.

**Architecture:** Keep installers non-mutating beyond installing the binary. After install, each script checks whether the install directory is already available in PATH and prints modern platform-specific guidance only when needed. README documents that installers print PATH setup examples.

**Tech Stack:** POSIX shell for `install.sh`, PowerShell for `install.ps1`, Markdown for README, Go test suite for regression confidence.

## Global Constraints

- Do not automatically modify user shell profiles or environment variables.
- macOS guidance targets modern zsh users and `~/.zshrc`.
- Windows guidance targets modern PowerShell users and User PATH.
- Installers remain user-directory installs and require no administrator privileges.
- Existing release artifact names and download URLs remain unchanged.

---

## File Structure

- Modify: `install.sh` — print macOS zsh PATH setup commands after successful install when needed.
- Modify: `install.ps1` — print PowerShell User PATH setup commands after successful install when needed.
- Modify: `README.md` — mention that installers print PATH setup examples.

---

### Task 1: macOS install.sh PATH Guidance

**Files:**
- Modify: `install.sh`

**Interfaces:**
- Consumes: existing `INSTALL_DIR` and `BIN` variables.
- Produces: installer output with copyable zsh commands when `$INSTALL_DIR` is not in `$PATH`.

- [ ] **Step 1: Update post-install PATH guidance**

Replace the existing block at the end of `install.sh`:

```sh
echo "Installed $BIN to $INSTALL_DIR/$BIN"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "Add $INSTALL_DIR to PATH to run '$BIN' from anywhere." ;;
esac
```

with:

```sh
echo "Installed $BIN to $INSTALL_DIR/$BIN"
case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    echo "You can now run: $BIN"
    ;;
  *)
    echo ""
    echo "To run '$BIN' from anywhere, add $INSTALL_DIR to PATH."
    echo "For modern macOS zsh, run:"
    echo ""
    echo "  mkdir -p \"$INSTALL_DIR\""
    echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
    echo "  source ~/.zshrc"
    echo ""
    echo "Or reopen your terminal after updating ~/.zshrc."
    ;;
esac
```

- [ ] **Step 2: Verify shell syntax**

Run:

```bash
sh -n install.sh
```

Expected: no output and exit code 0.

---

### Task 2: Windows install.ps1 PATH Guidance

**Files:**
- Modify: `install.ps1`

**Interfaces:**
- Consumes: existing `$InstallDir` variable.
- Produces: installer output with copyable PowerShell commands when `$InstallDir` is not in User PATH.

- [ ] **Step 1: Update post-install PATH guidance**

Replace this block inside `install.ps1`:

```powershell
Write-Host "Installed wildread-cli.exe to $InstallDir"
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (($UserPath -split ';') -notcontains $InstallDir) {
    Write-Host "Add $InstallDir to your user PATH to run wildread-cli from anywhere."
    Write-Host "You can run: `$userPath = [Environment]::GetEnvironmentVariable('Path', 'User'); [Environment]::SetEnvironmentVariable('Path', `$userPath + ';$InstallDir', 'User')"
}
```

with:

```powershell
Write-Host "Installed wildread-cli.exe to $InstallDir"
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (($UserPath -split ';') -contains $InstallDir) {
    Write-Host "You can now run: wildread-cli"
} else {
    Write-Host ""
    Write-Host "To run wildread-cli from anywhere, add $InstallDir to your User PATH."
    Write-Host "For modern PowerShell, run:"
    Write-Host ""
    Write-Host '  $userPath = [Environment]::GetEnvironmentVariable(''Path'', ''User'')'
    Write-Host "  [Environment]::SetEnvironmentVariable('Path', \"`$userPath;$InstallDir\", 'User')"
    Write-Host ""
    Write-Host "Then reopen PowerShell."
}
```

- [ ] **Step 2: Verify PowerShell syntax if available**

Run:

```bash
pwsh -NoProfile -Command '$null = [System.Management.Automation.Language.Parser]::ParseFile("install.ps1", [ref]$null, [ref]$errs); if ($errs.Count) { $errs | ForEach-Object { $_.Message }; exit 1 }'
```

Expected: exit code 0. If `pwsh` is unavailable, record that this check was skipped.

---

### Task 3: README Installer Note and Verification

**Files:**
- Modify: `README.md`

**Interfaces:**
- Produces: installation docs that tell users installers print PATH setup examples.

- [ ] **Step 1: Add README note**

After the default install path bullets in `README.md`, add:

```markdown
安装完成后，如果安装目录还不在 PATH 中，脚本会输出可复制的 PATH 配置命令示例。
```

- [ ] **Step 2: Run full verification**

Run:

```bash
sh -n install.sh
go test ./... -v
```

Expected: shell syntax check exits 0; Go tests pass.

- [ ] **Step 3: Review diff**

Run:

```bash
git diff -- install.sh install.ps1 README.md docs/superpowers/plans/2026-06-19-installer-path-guidance.md
```

Expected: only PATH guidance and README note changes appear.

---

## Self-Review

- Spec coverage: install.sh, install.ps1, README note, and verification are covered.
- Placeholder scan: no TODO/TBD/placeholders.
- Type consistency: shell variables `INSTALL_DIR`/`BIN` and PowerShell variables `$InstallDir`/`$UserPath` match existing scripts.
