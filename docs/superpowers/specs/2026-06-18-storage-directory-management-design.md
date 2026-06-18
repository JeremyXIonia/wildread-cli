# Storage Directory Management Design

## Goal

Move app data to a stable user-level data directory, store managed book directories in SQLite, and add a bookshelf-accessible TUI screen for viewing, adding, deleting, and rescanning book directories.

## Background

The current app defaults to paths relative to the process working directory:

- books: `./books`
- database: `./novel-reader.db`

This makes the user's bookshelf and reading progress depend on where they launch `reader`. The app should instead use a stable data location by default and let users manage one or more book directories from inside the TUI.

## Defaults and Path Semantics

Default app data directory:

```text
~/.cli-read
```

Default contents:

```text
~/.cli-read/
├── novel-reader.db
└── .book/
```

Rules:

- `reader` with no flags uses `~/.cli-read/novel-reader.db` for SQLite data.
- The default managed book directory is `~/.cli-read/.book`.
- The app creates the data directory and default book directory if needed.
- The README documents Go 1.25+ and these default paths.

CLI flags:

- `--data-dir <dir>` sets the app data directory. The default database becomes `<dir>/novel-reader.db`; the default book directory becomes `<dir>/.book`.
- `--db <file>` remains as an advanced override for the database file.
- `--dir <dir>` remains for compatibility as a one-session temporary scan directory. It does not write to `library_dirs`; the UI should hint that persistent directories can be added from directory management.

Path normalization:

- Expand `~` to the current user's home directory.
- Convert relative paths to absolute paths before saving.
- Clean paths with `filepath.Clean`.
- Store normalized absolute paths in SQLite.

## SQLite Schema

Add a managed directory table:

```sql
CREATE TABLE IF NOT EXISTS library_dirs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    is_default INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Store-layer API:

```go
type LibraryDir struct {
    ID        int64
    Path      string
    IsDefault bool
    CreatedAt string
}

func (s *Store) ListLibraryDirs() ([]models.LibraryDir, error)
func (s *Store) AddLibraryDir(path string, isDefault bool) (int64, error)
func (s *Store) DeleteLibraryDir(id int64) error
func (s *Store) DeleteBooksUnderDir(dir string) error
```

`DeleteBooksUnderDir` deletes all `books` rows whose `file_path` is under the normalized directory. Existing foreign-key cascade removes corresponding `reading_progress` and `bookmarks` rows.

## Startup Flow

1. Parse flags.
2. Resolve paths:
   - `dataDir`
   - `dbPath`
   - `defaultBookDir`
   - optional temporary `--dir`
3. Create `dataDir` and `defaultBookDir` if needed.
4. Open SQLite database.
5. Load `library_dirs`.
6. If no managed directories exist:
   - insert `defaultBookDir` with `is_default = true`
   - show bookshelf status: `未配置书籍目录，已使用默认目录 <path>`
7. Scan all managed directories plus any temporary `--dir`.
8. Sync scanned books into the store.
9. Render the bookshelf.

If a configured directory no longer exists, scanning should not crash the app. The app should keep the directory in the list and show a status message that one or more directories could not be scanned.

## Directory Management UI

Add an independent Bubble Tea model, not a bookshelf sub-mode:

```text
app/directory_manager.go
app/directory_manager_test.go
```

Root app mode:

```go
modeBookshelf
modeReader
modeDirectoryManager
```

Bookshelf entry point:

- Press `D` from the bookshelf to open directory management.

Directory manager layout:

```text
书籍目录

> ~/.cli-read/.book        默认
  ~/Books
  /Volumes/Novel

a 添加目录  d 删除目录  r 重新扫描  q 返回书架
```

Keys:

| Key | Behavior |
| --- | --- |
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `a` | Enter add-directory input |
| `d` | Enter delete confirmation for selected directory |
| `r` | Rescan all directories immediately |
| `q` / `Esc` | Return to bookshelf |

## Adding Directories

Pressing `a` opens a path input focused on paste/type workflows:

```text
添加目录：
/path/pasted/from/file-manager
```

Behavior:

1. User pastes or types a path.
2. On Enter, normalize the path.
3. If the path already exists in `library_dirs`, show `目录已存在`.
4. If the path exists and is a directory, save it.
5. If the path does not exist, ask whether to create it:
   - `y`: create directory, save it
   - `n` / `Esc`: cancel
6. After a successful add, immediately rescan all managed directories and refresh the bookshelf.
7. Stay in directory management and show a status message like `已添加目录 <path>`.

## Deleting Directories

Deleting a directory always deletes related book data. There is no "remove only from directory list" mode.

Pressing `d` on a selected directory shows an explicit confirmation:

```text
删除目录：~/Books

这会删除该目录下已入库的书籍、阅读进度和书签。
目录中的原始文件不会被删除。

输入 y 确认删除，Esc/q 取消
```

Behavior:

- `y`: delete the `library_dirs` row, delete all `books` rows under that directory, rely on FK cascade for `reading_progress` and `bookmarks`, rescan remaining directories, refresh bookshelf, stay in directory management.
- `Esc` / `q`: cancel and return to the directory list.
- Do not delete files from disk.
- If the selected directory is the only directory, deletion is allowed. After deletion, startup/default-directory logic should ensure a default directory exists again before the next scan; in the current session, immediately recreate and re-add the default directory so the app never has zero managed directories.

## Refresh and Sync

Directory changes refresh immediately.

On add, delete, or `r`:

1. Load managed directories.
2. Ensure at least one directory exists, re-adding the default if needed.
3. Scan all managed directories.
4. Sync books.
5. Rebuild the bookshelf model with the current `ListBooks()` result.
6. Keep the user in directory management unless they explicitly press `q`/`Esc`.

Book sync should distinguish between two operations:

- Normal full scan: prune books missing from all currently managed directories.
- Explicit directory deletion: before the normal scan, delete all book records under the removed directory.

## Error Handling

- Data directory creation failure: print to stderr and exit before TUI starts.
- Database open failure: print to stderr and exit.
- Invalid added path: show status message in directory manager.
- Duplicate path: show status message in directory manager.
- Directory creation failure: show status message.
- Scan failure for one directory: continue scanning other directories and show a warning status.
- Delete failure: show status message and do not refresh.

## Testing

Add tests for:

1. Path resolution:
   - default `~/.cli-read`
   - `--data-dir`
   - `--db` override
   - `~` expansion
   - relative path normalization
2. Store APIs:
   - `library_dirs` CRUD
   - duplicate path handling
   - deleting books under a directory cascades progress and bookmarks
3. Startup/default behavior:
   - empty `library_dirs` creates and inserts default `.book`
   - temporary `--dir` is scanned but not persisted
4. Directory manager model:
   - list rendering
   - `a` enters input mode
   - pasted path add saves and triggers refresh message
   - nonexistent path asks to create
   - `d` enters delete confirmation
   - `y` confirms destructive delete
   - `q`/`Esc` cancels delete or exits manager
5. Integration:
   - multiple managed directories scan into one bookshelf
   - deleting a managed directory removes books/progress/bookmarks under that directory and refreshes bookshelf

## Non-Goals for the First Version

- File-system browsing UI.
- Per-directory enable/disable toggles.
- Directory ordering controls.
- Multi-profile support.
- Deleting original book files from disk.
- Cloud sync or import/export.