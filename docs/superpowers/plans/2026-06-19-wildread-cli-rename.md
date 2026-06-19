# Wildread CLI Rename Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rename the project to `wildread-cli` across module imports, build outputs, docs, and default app data paths.

**Architecture:** This is a focused mechanical rename. Runtime code keeps the same structure; only the Go module path, import strings, constants, build artifact names, docs, and tests for default paths change.

**Tech Stack:** Go 1.25+, standard Go module imports, shell build script, Markdown docs.

## Global Constraints

- Project name: `wildread-cli`.
- Go module path: `github.com/JeremyXIonia/wildread-cli`.
- CLI binary names: `wildread-cli` and `wildread-cli.exe`.
- Release artifact prefix: `wildread-cli`.
- Default app data directory: `~/.wildread-cli`.
- Default database file name: `wildread-cli.db`.
- Keep old ignore patterns for `reader*` and `novel-reader*` to avoid tracking stale local build artifacts.
- Do not rewrite historical plan/spec documents except this rename plan and product-facing docs.

---

## File Structure

- Modify: `go.mod` — module path.
- Modify: all `.go` files importing `github.com/xuanchong/cli-read/...` — update imports.
- Modify: `config/paths.go` and `config/paths_test.go` — default data dir and DB filename.
- Modify: `build.sh` — output artifact names.
- Modify: `.gitignore` — add/keep wildread-cli artifact ignore patterns.
- Modify: `README.md` — title, build/run commands, storage defaults, tree name.
- Modify: `CLAUDE.md` — build/run commands and design decisions.

---

### Task 1: Module Path and Imports

**Files:**
- Modify: `go.mod`
- Modify: all Go source files under project root except generated/vendor files

**Interfaces:**
- Produces module path `github.com/JeremyXIonia/wildread-cli`.

- [ ] **Step 1: Update module path**

Change `go.mod` line 1 from:

```go
module github.com/xuanchong/cli-read
```

to:

```go
module github.com/JeremyXIonia/wildread-cli
```

- [ ] **Step 2: Update imports**

Replace all Go import strings:

```text
github.com/xuanchong/cli-read
```

with:

```text
github.com/JeremyXIonia/wildread-cli
```

Run:

```bash
grep -R "github.com/xuanchong/cli-read" -n --include='*.go' . || true
```

Expected: no output.

- [ ] **Step 3: Format and test compile**

Run:

```bash
gofmt -w $(find . -name '*.go' -not -path './.git/*' -not -path './.claude/*')
go test ./... -v
```

Expected: PASS.

---

### Task 2: Default Paths and Tests

**Files:**
- Modify: `config/paths.go`
- Modify: `config/paths_test.go`

**Interfaces:**
- Produces default data dir `~/.wildread-cli` and DB `wildread-cli.db`.

- [ ] **Step 1: Update path tests first**

In `config/paths_test.go`, change expectations:

```go
wantData := filepath.Join(home, ".wildread-cli")
```

and:

```go
filepath.Join(wantData, "wildread-cli.db")
```

Run:

```bash
go test ./config -v
```

Expected before implementation: FAIL because code still returns `.cli-read` and `novel-reader.db`.

- [ ] **Step 2: Update constants**

In `config/paths.go`, change:

```go
DefaultDataDirName = ".wildread-cli"
DefaultDBFileName  = "wildread-cli.db"
```

- [ ] **Step 3: Run path tests**

Run:

```bash
go test ./config -v
```

Expected: PASS.

---

### Task 3: Build Artifacts and Ignore Rules

**Files:**
- Modify: `build.sh`
- Modify: `.gitignore`

**Interfaces:**
- Produces binary names `wildread-cli`, `wildread-cli.exe`, `bin/wildread-cli-*`.

- [ ] **Step 1: Update build.sh artifact names**

Change outputs to:

```bash
go build -ldflags "$LDFLAGS" -o wildread-cli .
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/wildread-cli-windows-amd64.exe .
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/wildread-cli-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o bin/wildread-cli-darwin-arm64 .
ls -la bin/wildread-cli-*
```

- [ ] **Step 2: Update ignore rules**

Ensure `.gitignore` includes:

```gitignore
wildread-cli*
reader*
novel-reader*
```

Do not remove old `reader*` / `novel-reader*` patterns.

- [ ] **Step 3: Verify shell syntax**

Run:

```bash
bash -n build.sh
```

Expected: no output.

---

### Task 4: Product Docs Rename

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`

**Interfaces:**
- Product-facing docs use `Wildread CLI`, `wildread-cli`, `~/.wildread-cli`, and `wildread-cli.db`.

- [ ] **Step 1: Update README product name and commands**

In `README.md`:

- Title becomes:

```markdown
# Wildread CLI — 终端小说阅读器
```

- Build commands use:

```bash
go build -o wildread-cli .
```

and Windows:

```powershell
go build -o wildread-cli.exe .
```

- Usage examples use:

```bash
wildread-cli
wildread-cli --data-dir /path/to/app-data
wildread-cli --dir /path/to/books
wildread-cli --db /path/to/db.sqlite
```

- Storage docs use:

```text
~/.wildread-cli/
├── wildread-cli.db
└── .book/
```

- Project tree root becomes:

```text
wildread-cli/
```

- Install examples copy/move `wildread-cli` or `wildread-cli.exe`.

- [ ] **Step 2: Update CLAUDE.md**

In `CLAUDE.md`, build/run commands become:

```bash
go build -o wildread-cli .
./wildread-cli --dir ./books
```

Windows:

```powershell
go build -o wildread-cli.exe .
.\wildread-cli.exe --dir .\books
```

Design decision becomes:

```markdown
- App data defaults to `~/.wildread-cli` (`wildread-cli.db` + `.book`)
```

- [ ] **Step 3: Verify product docs**

Run:

```bash
grep -R "Novel Reader\|novel-reader.db\|~/.cli-read\|go build -o reader\|reader --data-dir\|reader --dir\|reader --db" -n README.md CLAUDE.md || true
```

Expected: no output.

---

### Task 5: Full Verification and Commit

**Files:**
- All modified files

- [ ] **Step 1: Run full tests**

Run:

```bash
go test ./... -v
```

Expected: PASS.

- [ ] **Step 2: Verify old active names are gone from code/docs**

Run:

```bash
grep -R "github.com/xuanchong/cli-read" -n --include='*.go' --include='go.mod' . || true
grep -R "~/.cli-read\|novel-reader.db" -n README.md CLAUDE.md config || true
```

Expected: no output.

- [ ] **Step 3: Review diff**

Run:

```bash
git diff --stat
git status --short
```

Expected: only rename-related text changes and the new plan file.

- [ ] **Step 4: Commit**

Run:

```bash
git add -A
git commit -m "chore: rename project to wildread-cli" -m "Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Self-Review

- Spec coverage: module path, imports, binary names, release artifacts, defaults, docs, and ignore rules are covered.
- Placeholder scan: no placeholders remain.
- Type consistency: no API signature changes beyond module/import path and path constants.
