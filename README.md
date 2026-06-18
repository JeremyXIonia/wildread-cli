# Novel Reader — 终端小说阅读器

一个跨平台（Windows / macOS）的终端小说阅读器，支持 EPUB、TXT、Markdown 格式。

## 功能

- 📚 **书架**：扫描指定目录下的所有书籍，自动同步
- 📖 **阅读**：按终端高度分页，支持中文宽字符
- 📑 **目录**：按章节跳转，Vim 风格操作
- 🔖 **书签**：随时添加、查看、删除
- 💾 **进度**：自动保存阅读进度，下次打开继续阅读
- ⌨️ **Vim 风格快捷键**：全键盘操作

## 环境要求

- [Go 1.25+](https://go.dev/dl/)
- 无需 CGO、无需 C 编译器（MSVC / Xcode Command Line Tools 均不需要）

## 编译

### 日常开发

```bash
# 运行测试
go test ./...

# 编译当前平台
go build .
```

如需指定输出文件名：

```bash
# macOS / Linux
go build -o reader .

# Windows
go build -o reader.exe .
```

### 交叉编译示例

```bash
# 在 macOS / Linux 上编译 Windows 版本
GOOS=windows GOARCH=amd64 go build -o reader.exe .

# 在 Windows CMD 中编译 macOS Apple Silicon 版本
set GOOS=darwin
set GOARCH=arm64
go build -o reader-mac-arm64 .
```

### 一键构建发布产物

```bash
./build.sh        # macOS / Linux / Git Bash / WSL
```

生成当前平台、Windows (amd64)、macOS (Intel + Apple Silicon) 的二进制文件。

> **注意：** 使用 `modernc.org/sqlite`（纯 Go SQLite），无需安装 MSVC 或任何 C 编译器。

## 安装到全局路径

编译完成后，将 `reader`（或 `reader.exe`）放到系统 PATH 中，即可在任意位置运行。

### Windows

**方法一：移动到已有 PATH 目录**

```powershell
copy reader.exe C:\Windows\System32\
```

**方法二：添加自定义目录到 PATH（推荐）**

```powershell
# 创建目录
mkdir C:\tools

# 移动文件
move reader.exe C:\tools\

# 添加到用户 PATH（永久生效，需重启终端）
setx PATH "$env:PATH;C:\tools"
```

### macOS

**方法一：`/usr/local/bin`（推荐）**

```bash
sudo cp reader /usr/local/bin/
```

**方法二：用户目录（不需要 sudo）**

```bash
mkdir -p ~/.local/bin
cp reader ~/.local/bin/

# 添加到 PATH（~/.zshrc 是 macOS 默认 shell 配置）
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

## 使用

```bash
# 默认扫描 ./books 目录
reader

# 指定书籍目录
reader --dir /path/to/books

# 指定数据库路径
reader --db /path/to/db.sqlite
```

## 键盘快捷键

### 书架

| 键 | 功能 |
|----|------|
| `j` / `↓` | 向下移动 |
| `k` / `↑` | 向上移动 |
| `Enter` | 打开选中书籍 |
| `/` | 搜索书名 |
| `q` | 退出 |

### 阅读

| 键 | 功能 |
|----|------|
| `j` / `Space` / `↓` / `PgDn` | 下一页 |
| `k` / `↑` / `PgUp` | 上一页 |
| `gg` | 跳到章节开头 |
| `G` | 跳到章节末尾 |
| `o` | 打开章节目录 |
| `n` | 下一章 |
| `p` | 上一章 |
| `m` | 添加书签 |
| `b` | 查看书签 |
| `q` / `Esc` | 返回书架 |

### 目录 / 书签面板

| 键 | 功能 |
|----|------|
| `j` / `k` | 上下移动 |
| `Enter` | 跳转 |
| `Esc` / `q` | 返回阅读 |

## 支持的格式

- **TXT**: 自动检测 UTF-8 / GBK / GB18030 编码
- **Markdown**: 按 `#` / `##` 标题切分章节
- **EPUB**: 自实现解析器（ZIP + OPF + HTML→文本）

## 数据存储

所有数据保存在一个 SQLite 文件中（默认 `./novel-reader.db`）：

- `books` — 书架
- `reading_progress` — 每本书的阅读进度
- `bookmarks` — 书签

## 项目结构

```
cli-read/
├── main.go           # 入口
├── app/              # TUI 组件（书架、阅读器）
├── models/           # 数据模型
├── store/            # SQLite 持久层
├── parser/           # 文档解析器（TXT/MD/EPUB）
├── pager/            # 文本分页
├── ui/               # 键盘映射和样式
├── e2e/              # 端到端测试
└── testdata/         # 测试样本
```

## 限制

- 不做实时文件监控，新增/删除书籍需重启程序
- 不做云同步
- 不做复杂 Markdown 渲染（表格、代码块、公式等）
- 界面文案为中文

## 许可

MIT
