# MarkoWiki

A lightweight, self-hosted Markdown wiki that lets you open multiple folder locations at once — from anywhere on your system — and browse, edit, and search them all in one place.

Built for developers, writers, and anyone who keeps notes scattered across their machine.

![License](https://img.shields.io/badge/license-MIT-green)
![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)

---

## Features

- **Multiple vaults** — add any folder from anywhere on your system
- **CodeMirror editor** — Markdown syntax highlighting with 8 themes
- **Live preview** — split, editor-only, or preview-only modes
- **Math rendering** — inline `$...$` and block `$$...$$` via KaTeX
- **Full toolbar** — bold, italic, headings, lists, links, tables, code, math
- **Find & Replace** — built-in via Ctrl+F / Ctrl+H
- **Full-text search** — searches across all vaults simultaneously
- **Dark / Light mode** — toggle for the whole UI
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

---

## Getting Started

1. Click **"+ Add vault folder"** in the sidebar
2. Enter a display name and the absolute path to your folder
3. Browse the file tree and click any `.md` file to open it
4. Edit on the left, preview on the right

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

## License

MIT — see [LICENSE](LICENSE)

---

## Author

Created by [Nader Ereij](https://github.com/nader-ereij)  
A [Markologist LLC](https://github.com/markologist-llc) project
