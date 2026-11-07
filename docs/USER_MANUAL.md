# ArchTerm - User Manual

> **THE ARCHITECT'S CODEX** — A terminal-style file browser and document viewer with semantic search capabilities.

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Installation](#installation)
3. [Launching the Server](#launching-the-server)
4. [Using the Terminal Interface](#using-the-terminal-interface)
5. [Available Commands](#available-commands)
6. [Command Details](#command-details)
7. [Keyboard Shortcuts](#keyboard-shortcuts)
8. [File Categories](#file-categories)
9. [API Reference](#api-reference)
10. [Troubleshooting](#troubleshooting)

---

## Getting Started

ArchTerm transforms your project directory into an immersive, terminal-style interface for browsing files, searching content, and viewing documents — all from your browser.

### Prerequisites

- **Go 1.21+** — Required to build the server binary
- **Python 3.8+** (optional) — Required only if using the Python document parser
- **Modern browser** — Chrome, Firefox, Edge, or Safari

### Quick Start (5 minutes)

```bash
# 1. Clone the repository
git clone https://github.com/yourusername/archterm.git
cd archterm

# 2. Initialize Go modules
go mod tidy

# 3. Build the binary
make build

# 4. Start the server
./archterm -port 8080

# 5. Open your browser
# Navigate to http://localhost:8080
```

That's it. You now have a fully functional terminal-style file browser running in your browser.

---

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/archterm.git
cd archterm

# Install Go dependencies
go mod tidy

# Build the application
make build

# (Optional) Install to GOPATH
make install
```

### Using Make

| Target | Description |
|--------|-------------|
| `make build` | Build the binary |
| `make run` | Build and run on port 8080 |
| `make dev` | Development mode (build + run) |
| `make clean` | Remove build artifacts |
| `make install` | Install to GOPATH/bin |

### Using Go Directly

```bash
go build -o archterm ./cmd/archterm/
./archterm
```

---

## Launching the Server

### Default Port (8080)

```bash
./archterm
```

### Custom Port

```bash
./archterm -port 9090
```

### Server Behavior

When started, ArchTerm:

1. **Scans the project directory** — Recursively walks `./docs`, `./plans`, `./cmd`, `./internal`, `./webapp`, and `./python_parser`
2. **Indexes files** — Registers each text-based file with its category (documents, code, config, images, data, other)
3. **Builds a search index** — Creates a TF-IDF index for semantic search
4. **Starts the HTTP server** — Listens on the specified port
5. **Serves the web interface** — Available at `http://localhost:<port>`

### Server Output

```
Starting file watcher...
Watching path: ./docs
Watching path: ./plans
Watching path: ./cmd
Watching path: ./internal
Watching path: ./webapp
Watching path: ./python_parser
Initial scan complete. Indexed 42 files.
Starting server on port 8080...
```

---

## Using the Terminal Interface

Once the server is running, open your browser to `http://localhost:<port>` (e.g., `http://localhost:8080`).

You'll see a CRT-styled terminal with:

- **Banner** — ArchTerm logo and version
- **Command history** — Your typed commands and their output
- **Input line** — Where you type commands (prompt: `arch >`)
- **Status bar** — Shows current status and file count

### First Steps

1. Type `help` and press Enter to see all available commands
2. Type `ls` and press Enter to list all indexed files
3. Type `ls documents` to list only document files
4. Type `access README.md` to view a file's contents
5. Type `search architecture` to search across all indexed files

---

## Available Commands

| Command | Description | Example |
|---------|-------------|---------|
| `help` | Show available commands | `help` |
| `ls [category]` | List files by category | `ls code` |
| `access <file>` | View file contents | `access README.md` |
| `access <file> render` | View markdown with rendered HTML | `access docs/planning/archterm_prd.md render` |
| `view <file>` | Open file in overlay viewer | `view cmd/archterm/main.go` |
| `metadata <file>` | Show file metadata | `metadata go.mod` |
| `related <file>` | Find related documents | `related cmd/archterm/main.go` |
| `search <query>` | Search indexed files | `search "file watcher"` |
| `clear` | Clear terminal | `clear` |
| `cat <file>` | Display file contents | `cat go.mod` |
| `pwd` | Show current directory | `pwd` |
| `whoami` | Show current user | `whoami` |
| `date` | Show current date/time | `date` |
| `exit` | Close terminal | `exit` |

---

## Command Details

### `help`

Displays a list of all available commands, categories, and usage examples.

```
arch > help
```

### `ls [category]`

Lists files in the specified category. If no category is given, lists all files.

**Categories:**

| Category | File Types |
|----------|-----------|
| `documents` | `.pdf`, `.md`, `.txt`, `.doc`, `.docx` |
| `code` | `.go`, `.js`, `.ts`, `.py`, `.rs`, `.java` |
| `config` | `.json`, `.yaml`, `.yml`, `.toml`, `.ini`, `.cfg` |
| `images` | `.png`, `.jpg`, `.jpeg`, `.gif`, `.svg`, `.ico` |
| `data` | `.csv`, `.xml`, `.sql`, `.db` |
| `all` | All files (default) |

**Examples:**

```
arch > ls
arch > ls documents
arch > ls code
arch > ls config
```

### `access <filename>`

Displays the raw contents of a file in the terminal output.

**Examples:**

```
arch > access README.md
arch > access docs/planning/archterm_prd.md
arch > access cmd/archterm/main.go
```

### `access <filename> render`

For markdown files (`.md`, `.markdown`), this renders the content as formatted HTML with:

- **Headers** — Styled with accent colors and borders
- **Bold/Italic** — Highlighted text
- **Code blocks** — Syntax-highlighted with dark background
- **Lists** — Properly indented bullet lists
- **Blockquotes** — Left-bordered quote blocks
- **Links** — Underlined and clickable
- **Horizontal rules** — Divider lines

**Examples:**

```
arch > access docs/planning/archterm_prd.md render
arch > access README.md render
```

### `view <filename>`

Opens a file in a full-screen overlay viewer for focused reading.

**Examples:**

```
arch > view cmd/archterm/main.go
arch > view internal/server/server.go
```

### `metadata <filename>`

Shows detailed metadata for a file:

- File size
- Last modified timestamp
- File category
- File mode/permissions

**Example:**

```
arch > metadata go.mod

--- Metadata for go.mod ---
  Name:       go.mod
  Path:       go.mod
  Category:   config
  Size:       128 bytes
  Modified:   2024-01-15T10:30:00
  Mode:       -rw-r--r--
---
```

### `related <filename>`

Finds documents related to the specified file using concept similarity analysis. Shows:

- Related file names
- Similarity percentage
- Key concepts shared

**Example:**

```
arch > related cmd/archterm/main.go

Related documents for "cmd/archterm/main.go":
──────────────────────────────────────────
  1. [85%] internal/server/server.go
     Path: internal/server/server.go
     Concepts: [████████░░] server, http, port
  2. [62%] internal/config/config.go
     Path: internal/config/config.go
     Concepts: [██████░░░░] config, manifest
```

### `search <query>`

Performs a TF-IDF based semantic search across all indexed files.

**Examples:**

```
arch > search architecture
arch > search "file watcher"
arch > search http server
```

### `cat <filename>`

Alias for `access`. Displays raw file contents.

**Example:**

```
arch > cat go.mod
```

### `clear`

Clears the terminal output.

```
arch > clear
```

### `pwd`

Shows the current working directory.

```
arch > pwd
/Users/juliancowell/dev/arch_term
```

### `whoami`

Shows the current user.

```
arch > whoami
arch
```

### `date`

Shows the current date and time.

```
arch > date
Mon Jan 15 10:30:00 UTC 2024
```

---

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Enter` | Execute command |
| `Tab` | Auto-complete command |
| `↑` (Up Arrow) | Previous command in history |
| `↓` (Down Arrow) | Next command in history |
| `Ctrl+L` | Clear terminal |
| `Escape` | Close file viewer overlay |

---

## File Categories

ArchTerm automatically categorizes files based on their extension:

| Category | Extensions | Icon |
|----------|-----------|------|
| **Documents** | `.pdf`, `.md`, `.markdown`, `.txt`, `.doc`, `.docx` | 📄 |
| **Code** | `.go`, `.js`, `.ts`, `.py`, `.rs`, `.java` | 🔧 |
| **Config** | `.json`, `.yaml`, `.yml`, `.toml`, `.ini`, `.cfg` | ⚙️ |
| **Images** | `.png`, `.jpg`, `.jpeg`, `.gif`, `.svg`, `.ico` | 🖼️ |
| **Data** | `.csv`, `.xml`, `.sql`, `.db` | 📊 |
| **Other** | Any unrecognized extension | 📎 |

---

## API Reference

ArchTerm exposes a REST API for programmatic access.

### Base URL

```
http://localhost:<port>/
```

### Endpoints

#### Get Manifest

```
GET /api/manifest
```

Returns the full file manifest with all indexed files.

#### List Files

```
GET /api/ls?category=all
```

**Query Parameters:**

| Parameter | Description | Default |
|-----------|-------------|---------|
| `category` | File category filter | `all` |

**Response:**

```json
{
  "category": "code",
  "files": [
    {
      "name": "main.go",
      "path": "cmd/archterm/main.go",
      "category": "code",
      "metadata": {
        "size": "1234",
        "modified": "2024-01-15T10:30:00",
        "mode": "-rw-r--r--"
      }
    }
  ]
}
```

#### Get File Content

```
GET /api/files/{filepath}
GET /api/files/{filepath}?render=true
```

**Query Parameters:**

| Parameter | Description | Default |
|-----------|-------------|---------|
| `render` | Render markdown as HTML | `false` |

**Response (raw):**

```
Content-Type: text/plain

[file contents]
```

**Response (rendered):**

```
Content-Type: text/html

[rendered HTML]
```

#### Search

```
GET /api/search?q=query
```

**Query Parameters:**

| Parameter | Description | Required |
|-----------|-------------|----------|
| `q` | Search query | Yes |

**Response:**

```json
{
  "query": "architecture",
  "results": [
    {
      "document": {
        "id": "docs/architecture.md",
        "title": "Architecture",
        "content": "System architecture documentation...",
        "source": "docs/architecture.md",
        "category": "documents"
      },
      "score": 0.8542
    }
  ]
}
```

#### Get Related Documents

```
GET /api/related?doc=filename
```

**Query Parameters:**

| Parameter | Description | Required |
|-----------|-------------|----------|
| `doc` | Source file path | Yes |

**Response:**

```json
{
  "source": "cmd/archterm/main.go",
  "related": [
    {
      "target_title": "Server",
      "target_source": "internal/server/server.go",
      "similarity": 85,
      "shared_concepts": ["server", "http", "port"]
    }
  ]
}
```

#### Get Concepts

```
GET /api/concepts?doc=filename
```

**Query Parameters:**

| Parameter | Description | Required |
|-----------|-------------|----------|
| `doc` | File path | Yes |

---

## Troubleshooting

### Server won't start

**Problem:** Port is already in use.

**Solution:** Use a different port:

```bash
./archterm -port 9090
```

### 404 on web interface

**Problem:** Browser shows "404 page not found".

**Solution:** Ensure the server is running and you're accessing the correct URL:

```bash
# Check if server is running
lsof -i :8080

# Restart the server
./archterm -port 8080
```

### `access` command returns "File not found"

**Problem:** The file path is incorrect.

**Solution:** Use the full relative path from the project root:

```bash
# Correct
arch > access docs/planning/archterm_prd.md

# Incorrect (wrong directory name)
arch > access documents/archterm_prd.md
```

### Markdown not rendering

**Problem:** `render` flag doesn't work.

**Solution:** Ensure the file has a `.md` or `.markdown` extension:

```bash
# Works
arch > access docs/planning/archterm_prd.md render

# Does not work (not a markdown file)
arch > access README.txt render
```

### No files showing in `ls`

**Problem:** The file watcher hasn't indexed any files.

**Solution:** Wait a moment for the initial scan to complete. Check the server logs:

```
Initial scan complete. Indexed 42 files.
```

If still empty, verify the watched directories exist in [`internal/filesystem/watcher.go`](internal/filesystem/watcher.go:27).

---

## Appendix: Project Structure

```
archterm/
├── cmd/archterm/          # Main application entry point
│   └── main.go
├── internal/              # Internal packages
│   ├── command/           # Command parser
│   ├── config/            # Configuration and manifest
│   ├── filesystem/        # File watcher
│   ├── llm/               # LLM provider integration
│   ├── markdown/          # Markdown renderer
│   ├── ollama/            # Ollama client
│   ├── parser/            # Document parser
│   ├── search/            # Semantic search (TF-IDF)
│   └── server/            # HTTP server
├── webapp/                # Frontend assets
│   ├── index.html
│   ├── css/terminal.css
│   └── js/terminal.js
├── docs/                  # Documentation
├── plans/                 # Planning documents
├── python_parser/         # Python document parser (optional)
├── go.mod
├── Makefile
├── manifest.json
└── README.md
```
