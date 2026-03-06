# MarkoWiki

A lightweight, self-hosted Markdown wiki that lets you open multiple folder locations at once — from anywhere on your system — and browse, edit, and search them all in one place.

Built for developers, writers, and anyone who keeps notes scattered across their machine.

![License](https://img.shields.io/badge/license-MIT-green)
![Version](https://img.shields.io/badge/version-1.1.0-blue)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)

---

## Features

- **Multiple vaults** — add any folder from anywhere on your system
- **CodeMirror editor** — Markdown syntax highlighting, find & replace, vertical scrollbar
- **Live preview** — split, editor-only, or preview-only modes
- **Math rendering** — inline `$...$` and block `$$...$$` via KaTeX
- **Full toolbar** — bold, italic, headings, lists, links, tables, code, math, collapsible
- **Full-text search** — searches across all vaults simultaneously (Ctrl+K)
- **Dark / Light mode** — toggle for the whole UI
- **5 accent color schemes** — Forest, Amber, Ocean, Rose, Mono
- **Auto-save** — configurable delay, saves silently while you type
- **Custom blocks** — reusable Markdown templates with smart placeholders
- **Single binary** — no installation, no dependencies, just run it

---

## Download

Go to the [Releases](https://github.com/markologist-llc/markowiki/releases) page and download the binary for your platform:

| Platform | File |
|---|---|
| Linux | `markowiki-linux` |
| macOS | `markowiki-macos` |
| Windows | `markowiki-windows.exe` |

---

## Usage

### Linux / macOS
```bash
chmod +x markowiki-linux
./markowiki-linux
```

### Windows
Double-click `markowiki-windows.exe`

Then open **http://localhost:8000** in your browser.

On first run, a `config.json` file is created next to the binary storing your vaults, settings, and custom blocks.

---

## Getting Started

1. Click **"+ Add vault folder"** in the sidebar
2. Enter a display name and the absolute path to your folder
3. Browse the file tree and click any `.md` file to open it
4. Edit on the left, preview on the right

---

## Custom Blocks

Custom blocks are reusable Markdown templates you define once and insert anywhere from the toolbar. They support three placeholder types:

| Syntax | Type | Behavior on insert |
|--------|------|--------------------|
| `{title}` | Free text | Appears as a clickable mark — click it and type to replace |
| `{status:Draft,Done,Review}` | Pick list | Click the mark → choose from your defined options |
| `{date}`, `{time}`, `{datetime}` | Auto | Replaced immediately with the current value |

### Auto keywords

| Keyword | Output |
|---------|--------|
| `{date}` | `2026-03-06` |
| `{time}` | `14:35` |
| `{datetime}` | `2026-03-06 14:35` |
| `{year}` | `2026` |
| `{month}` | `03` |
| `{day}` | `06` |

### Example block schemas

**Meeting note**
```
## {title} — {date}
**Status:** {status:Draft,In Progress,Done}
**Author:** {author}
```

**Callout**
```
> **{type:Note,Warning,Tip}:** {message}
```

**Dated signature**
```
---
*Written by {name} · {datetime}*
```

To manage blocks: click **Blocks ▾** in the toolbar → **Manage blocks…**

---

## Keyboard Shortcuts

| Shortcut | Action |
|---|---|
| `Ctrl+S` | Save file |
| `Ctrl+K` or `/` | Search vaults |
| `Ctrl+\` | Toggle sidebar |
| `Ctrl+F` | Find in editor |
| `Ctrl+H` | Find & Replace |
| `Ctrl+B` | Bold |
| `Ctrl+I` | Italic |
| `Ctrl+L` | Insert link |

---

## Building from Source

Requires [Go](https://go.dev) 1.21+
```bash
git clone https://github.com/markologist-llc/markowiki.git
cd markowiki
go build -o markowiki
./markowiki
```

---

## What's new in v1.1.0

- **Custom blocks** — reusable templates with auto, list, and free-text placeholders
- **5 accent color schemes** — Forest, Amber, Ocean, Rose, Mono
- **Auto-save** — configurable delay in Settings (gear icon)
- **Unified config** — vaults, settings, and blocks in a single `config.json`
- **Light mode** — full light theme with per-accent contrast adjustments
- **Collapsible toolbar** — toggle button always visible in topbar
- **Math block styling** — `$$` blocks render with code-block background
- **`<details>` / `<summary>` styling** — accent-colored collapsible blocks in preview
- **Fixed false-dirty bug** — opening a file no longer triggers the unsaved changes warning
- **Editor scrollbar** — native vertical scrollbar styled to match the UI

---

## License

MIT — see [LICENSE](LICENSE)

---

## Author

Created by [Nader Ereij](https://github.com/naderereij)
A [Markologist LLC](https://github.com/markologist-llc) project
