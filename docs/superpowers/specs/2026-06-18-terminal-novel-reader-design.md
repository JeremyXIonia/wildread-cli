# 终端小说阅读器 — 设计文档

**日期:** 2026-06-18
**状态:** 已批准，待实施

## 目标

开发一个跨 Windows 和 macOS 平台的终端小说阅读器。程序在终端中运行，支持打开 EPUB、TXT、Markdown 三种格式的小说，提供目录、阅读、键盘翻页、阅读进度记录、书架等核心功能。

## 用户选择的关键约束

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 语言 | Go | 编译为单二进制，跨平台分发简单，TUI 生态成熟 |
| 书架模式 | 监控目录 | 指定一个目录，自动扫描 `.epub`/`.txt`/`.md` 文件 |
| 数据存储 | SQLite | 单文件存储，结构清晰，支持复杂查询 |
| 键盘风格 | Vim 风格 | 程序员直觉，操作高效 |
| 阅读模式 | 分页 | 按终端高度分页，沉浸式阅读 |
| 配色 | 跟随终端 | 不自定义颜色，简洁，适配所有终端主题 |
| EPUB 解析 | 自实现 | EPUB 是 ZIP + HTML，用标准库 + `golang.org/x/net/html` 即可 |
| 编码处理 | 自动检测 + 轻量渲染 | 支持 UTF-8/GBK/GB18030，Markdown 提取纯文本 |

## 架构

```
┌─────────────────────────────────────────────┐
│                   main.go                    │
│              (入口 & 命令行参数)               │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│              app/ (应用层)                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────────┐ │
│  │ Bookshelf│ │  Reader  │ │   Settings   │ │
│  │  书架管理 │ │  阅读器   │ │   设置管理    │ │
│  └────┬─────┘ └────┬─────┘ └──────┬───────┘ │
└───────┼────────────┼──────────────┼─────────┘
        │            │              │
┌───────▼────────────▼──────────────▼─────────┐
│              models/ (数据模型)               │
│  Book / Chapter / ReadingProgress / Config  │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│              store/ (数据持久层)              │
│           SQLite: 书架 + 进度 + 书签          │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│             parser/ (文档解析层)              │
│  ┌──────────┐ ┌──────────┐ ┌─────────────┐ │
│  │   EPUB   │ │   TXT    │ │  Markdown   │ │
│  │ 解析器   │ │ 解析器   │ │  解析器      │ │
│  └──────────┘ └──────────┘ └─────────────┘ │
└─────────────────────────────────────────────┘
```

## 组件职责

| 组件 | 职责 |
|------|------|
| **main.go** | 解析命令行参数（`--dir <path>` 可选指定书籍目录），初始化数据库，启动 TUI 主循环 |
| **Bookshelf** | 扫描监控目录，展示书籍列表，支持搜索/过滤，回车进入阅读并恢复到上次进度 |
| **Reader** | 分页渲染文本，响应键盘操作，实时更新阅读进度到数据库，提供目录和书签功能 |
| **Parser** | 统一接口 `Parser.Parse(path) ([]Chapter, error)`，三种格式各一个实现 |
| **Store** | SQLite 操作层，管理 books、reading_progress、bookmarks 三张表 |

## 数据模型

```sql
-- books: 书架中的书籍
CREATE TABLE books (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path   TEXT NOT NULL UNIQUE,
    title       TEXT NOT NULL,
    author      TEXT,
    format      TEXT NOT NULL,    -- epub / txt / md
    added_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- reading_progress: 每本书的阅读进度（一本书一行）
CREATE TABLE reading_progress (
    book_id     INTEGER PRIMARY KEY REFERENCES books(id) ON DELETE CASCADE,
    chapter     INTEGER DEFAULT 0,
    page        INTEGER DEFAULT 0,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- bookmarks: 书签
CREATE TABLE bookmarks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id     INTEGER REFERENCES books(id) ON DELETE CASCADE,
    chapter     INTEGER NOT NULL,
    page        INTEGER NOT NULL,
    label       TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 界面流程

```
   [书架界面]                  [阅读界面]
┌──────────────┐          ┌──────────────┐
│ 📖 三体       │  Enter   │ 第3章 第4页   │
│ 📖 活着       │ ──────► │              │
│ 📖 围城       │          │ 夕阳照在...   │
│              │          │              │
│ q: 退出      │  Esc    │ Space: 翻页  │
│ /: 搜索      │ ◄────── │ q: 回书架    │
│ Enter: 阅读  │          │ o: 目录      │
└──────────────┘          └──────────────┘
                              │ o
                              ▼
                        ┌──────────────┐
                        │  [章节目录]    │
                        │ 第一章 地球    │
                        │ 第二章 三体    │
                        │ →第三章 红岸  │
                        │              │
                        │ j/k: 移动    │
                        │ Enter: 跳转  │
                        │ Esc: 返回    │
                        └──────────────┘
```

## 键盘快捷键

### 书架界面

| 键 | 功能 |
|----|------|
| `j` / `↓` | 向下移动光标 |
| `k` / `↑` | 向上移动光标 |
| `Enter` | 打开选中书籍，恢复上次进度 |
| `/` | 进入搜索模式（按书名过滤） |
| `q` | 退出程序 |

### 阅读界面

| 键 | 功能 |
|----|------|
| `j` / `Space` / `→` / `PageDown` | 下一页 |
| `k` / `←` / `PageUp` | 上一页 |
| `gg` | 跳到章节第一页 |
| `G` | 跳到章节最后一页 |
| `o` | 打开章节目录 |
| `n` | 下一章 |
| `p` | 上一章 |
| `m` | 添加当前页为书签 |
| `b` | 查看书签列表 |
| `Esc` / `q` | 返回书架（自动保存进度） |

### 目录/书签面板

| 键 | 功能 |
|----|------|
| `j` / `k` | 上下移动 |
| `Enter` | 跳转到选中项 |
| `d` | 删除（仅书签面板） |
| `Esc` / `q` | 返回阅读界面 |

## 解析器设计

### 统一接口

```go
type Parser interface {
    Parse(path string) (*Book, error)
}

type Book struct {
    Title    string
    Author   string
    Chapters []Chapter
}

type Chapter struct {
    Title   string
    Content string    // 纯文本，保留段落结构（用 \n\n 分隔）
}
```

### EPUB 解析策略

EPUB 本质是 ZIP 包：
1. 用 `archive/zip` 打开
2. 读 `META-INF/container.xml` 找到 OPF 文件路径
3. 解析 OPF 获取 spine（章节顺序）和 manifest（资源映射）
4. 按 spine 顺序读取每个 HTML 章节
5. 用 `golang.org/x/net/html` 解析 HTML，提取文本和段落结构
6. 忽略所有内联样式、脚本、图片；只保留 `<h1>`~`<h6>` 作为可选的章节标题

### TXT 解析策略

1. 读取文件前 4KB 用于编码检测（用 `golang.org/x/text/encoding/htmlindex`）
2. 支持 UTF-8、GBK、GB18030
3. 按空行分段为章节（无标题），或检测 `# 第X章` 模式（如果用户希望用 Markdown 风格的章节标记）
4. 文件名（去后缀）作为书名

### Markdown 解析策略

1. 用 `goldmark` 或 `gopkg.in/russross/blackfriday.v2` 解析为 AST
2. 提取纯文本，保留段落分隔（连续两个换行）
3. 提取 `#` `##` 标题作为章节切分点
4. 文件名（去后缀）作为书名

## 书架目录扫描

启动时扫描指定的书籍目录：
1. 遍历目录（含一层子目录）
2. 识别 `.epub`/`.txt`/`.md` 文件
3. 与数据库中已有记录对比
4. 新增文件 → 解析元数据并插入 books 表
5. 缺失文件 → 从 books 表中删除（同时级联删除进度和书签）
6. 在 UI 顶栏显示「扫描到 N 本书」提示

> **范围说明：** 启动时单次扫描即可，不使用 `fsnotify` 实时监控。如果用户在程序运行期间新增/删除书籍，需要重启程序才生效。

## 关键库依赖

| 库 | 用途 |
|----|------|
| `github.com/charmbracelet/bubbletea` | TUI 框架，基于 Elm Architecture |
| `github.com/charmbracelet/bubbles` | 组件库（list、viewport、textinput） |
| `github.com/mattn/go-sqlite3` | SQLite 驱动 |
| `golang.org/x/text` | 编码检测与转换（GBK→UTF-8） |
| `golang.org/x/net/html` | HTML 解析（EPUB 内容提取） |
| `github.com/yuin/goldmark` | Markdown 解析 |

## 命令行参数

```
novel-reader [--dir <path>] [--db <path>]

选项：
  --dir <path>   书籍目录（默认：./books 或 ~/.local/share/novel-reader/books）
  --db <path>    SQLite 数据库文件路径（默认：./novel-reader.db）
```

## 错误处理

| 场景 | 处理 |
|------|------|
| 书籍文件不存在 | 删除对应数据库记录，UI 不显示 |
| EPUB 文件损坏 | 跳过该文件，控制台打印警告，UI 显示「解析失败」标记 |
| TXT 编码无法识别 | 尝试按 UTF-8 处理，失败则提示「无法识别编码」 |
| 数据库读写错误 | 显示错误信息并退出（致命错误） |
| 目录不存在 | 提示用户创建目录或指定其他目录 |

## 测试策略

| 测试类型 | 覆盖内容 |
|----------|----------|
| 单元测试 | 三个解析器、SQLite store、文本分页逻辑、编码检测 |
| 集成测试 | 解析 → 存储 → 读取进度的完整流程（用临时目录和内存 SQLite） |
| 端到端测试 | 启动程序，加载样本书籍，模拟按键序列（用 `teatest` 或类似工具） |
| 样本数据 | `testdata/` 目录下放真实的 EPUB/TXT/MD 文件用于测试 |

## 项目结构

```
cli-read/
├── main.go
├── go.mod
├── go.sum
├── README.md
├── docs/
│   └── superpowers/
│       └── specs/
│           └── 2026-06-18-terminal-novel-reader-design.md
├── app/
│   ├── bookshelf.go
│   ├── reader.go
│   ├── toc.go
│   └── bookmarks.go
├── models/
│   └── book.go
├── store/
│   ├── store.go
│   ├── schema.sql
│   └── store_test.go
├── parser/
│   ├── parser.go         # 接口定义
│   ├── epub.go
│   ├── txt.go
│   ├── markdown.go
│   ├── html2text.go      # HTML→纯文本工具
│   └── *_test.go
├── pager/
│   ├── pager.go          # 分页逻辑
│   └── pager_test.go
├── ui/
│   ├── styles.go         # 极简样式
│   └── keys.go           # 快捷键定义
└── testdata/
    ├── sample.epub
    ├── sample.txt
    └── sample.md
```

## 实施顺序（建议）

1. **基础骨架**：go.mod、main.go、命令行参数解析
2. **store 层**：SQLite schema、CRUD 函数、单元测试
3. **parser 层**：TXT（最简单）→ Markdown → EPUB；每个实现带测试
4. **pager 层**：文本分页逻辑 + 单元测试
5. **书架 UI**：bubbletea 列表、扫描、显示、搜索
6. **阅读 UI**：分页渲染、键盘响应、进度保存
7. **章节目录 UI**：跳转功能
8. **书签功能**：添加、列表、删除、跳转
9. **端到端测试**：用样本数据走完整流程
10. **跨平台构建**：用 Go 交叉编译生成 Windows 和 macOS 二进制
11. **README 编写**：使用说明、键盘快捷键参考

## 范围之外（暂不做）

- 自定义配色和字体
- 实时文件系统监控
- 多语言界面（界面文案为中文）
- 复杂的 Markdown 渲染（表格、代码块高亮等）
- 书籍元数据编辑
- 导入/导出书架
- 同步阅读进度到云端
- 阅读统计和历史记录
