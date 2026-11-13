# ArchTerm - Maintainer's Manual

> **THE ARCHITECT'S CODEX** — Development and maintenance guide for ArchTerm contributors.

---

## Table of Contents

1. [Overview](#overview)
2. [Development Environment Setup](#development-environment-setup)
3. [Project Architecture](#project-architecture)
4. [Codebase Structure](#codebase-structure)
5. [Building and Testing](#building-and-testing)
6. [Adding New Features](#adding-new-features)
7. [Adding New File Categories](#adding-new-file-categories)
8. [Extending the Markdown Renderer](#extending-the-markdown-renderer)
9. [API Development](#api-development)
10. [Frontend Development](#frontend-development)
11. [Configuration](#configuration)
12. [Network & Tailscale](#network--tailscale)
13. [Deployment](#deployment)
14. [Troubleshooting](#troubleshooting)
15. [Release Process](#release-process)

---

## Overview

ArchTerm is a hybrid Go/JavaScript application that provides a terminal-style file browser and document viewer accessible via web browser. The backend is written in pure Go (standard library only), and the frontend uses vanilla JavaScript with no framework dependencies.

### Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Backend** | Go 1.21+ | HTTP server, file watching, search indexing |
| **Frontend** | HTML5/CSS3/ES6+ | Terminal emulation, command interface |
| **Search** | Custom TF-IDF | Semantic file search |
| **Markdown** | Custom renderer | Markdown-to-HTML conversion |
| **Build** | Makefile | Build automation |

### Design Philosophy

- **Zero external dependencies** — Backend uses only Go standard library
- **Minimal frontend** — No frameworks, no build tools, no bundlers
- **Self-contained** — All assets embedded in the repository
- **Terminal-first UX** — Command-driven interface with keyboard shortcuts

---

## Development Environment Setup

### Prerequisites

```bash
# Go 1.21+
go version  # Should output go1.21 or higher

# Python 3.8+ (optional, for document parser)
python3 --version

# Make (macOS)
make --version

# Make (Linux)
sudo apt-get install make
```

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/archterm.git
cd archterm

# Run development setup
make dev-setup
```

This will:
1. Initialize Go modules (`go mod tidy`)
2. Install Python dependencies (`pip3 install -r python_parser/requirements.txt`)
3. Display environment information

### Manual Setup

```bash
# Go modules
go mod tidy

# Python dependencies (optional)
pip3 install -r python_parser/requirements.txt
```

---

## Project Architecture

### High-Level Flow

```
┌──────────────────────────────────────────────────────────────┐
│  Browser (http://localhost:8080)                             │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Terminal UI (terminal.js) + CRT CSS (terminal.css)   │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────┬───────────────────────────────────┘
                           │ HTTP REST API
┌──────────────────────────▼───────────────────────────────────┐
│  archterm (cmd/archterm/main.go)                            │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────────┐  │
│  │ HTTP Server │  │ File Watcher │  │  Search Indexer   │  │
│  │(server.go)  │  │(watcher.go) │  │ (semantic.go)     │  │
│  └─────────────┘  └──────────────┘  └───────────────────┘  │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────────┐  │
│  │ Manifest    │  │ Markdown     │  │  LLM Providers    │  │
│  │ Manager     │  │ Renderer     │  │  (ollama/, llm/)  │  │
│  └─────────────┘  └──────────────┘  └───────────────────┘  │
└──────────────────────────────────────────────────────────────┘
                           │
┌──────────────────────────▼───────────────────────────────────┐
│  File System                                                  │
│  docs/  plans/  cmd/  internal/  webapp/  python_parser/     │
└──────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Package | Responsibility |
|-----------|---------|----------------|
| **Main** | `cmd/archterm` | Application entry point, signal handling |
| **HTTP Server** | `internal/server` | Request routing, API handlers, static file serving |
| **File Watcher** | `internal/filesystem` | Directory scanning, file categorization, binary detection |
| **Manifest** | `internal/config` | File manifest data structures |
| **Search** | `internal/search` | TF-IDF indexing and query execution |
| **Markdown** | `internal/markdown` | Markdown-to-HTML rendering |
| **Command Parser** | `internal/command` | Terminal command parsing |
| **LLM Providers** | `internal/llm` | LLM provider interfaces |
| **Ollama Client** | `internal/ollama` | Ollama API integration |
| **Network Detector** | `internal/network` | Tailscale detection, network info, startup banner |

---

## Codebase Structure

### Backend (Go)

```
internal/
├── command/
│   └── parser.go          # Terminal command parsing logic
├── config/
│   └── config.go          # Data structures (Manifest, ManifestEntry)
├── filesystem/
│   └── watcher.go         # File system watcher and categorization
├── llm/
│   └── providers.go       # LLM provider interfaces
├── markdown/
│   └── renderer.go        # Markdown-to-HTML converter
├── ollama/
│   └── client.go          # Ollama API client
├── parser/
│   └── (empty)            # Reserved for document parsers
├── search/
│   └── semantic.go        # TF-IDF search engine
├── server/
│   └── server.go          # HTTP server and API handlers
└── network/
    └── detector.go        # Tailscale detection, network info, startup banner
cmd/
└── archterm/
    └── main.go            # Application entry point
```

### Frontend (Web)

```
webapp/
├── index.html             # Terminal UI structure
├── css/
│   └── terminal.css       # CRT terminal styling
└── js/
    └── terminal.js        # Terminal class and command handling
```

### Documentation

```
docs/
├── planning/
│   └── archterm_prd.md    # Product Requirements Document
├── USER_MANUAL.md         # User documentation (this repo)
└── MAINTAINER_MANUAL.md   # Developer documentation (this file)
plans/
└── development-plan.md    # Development roadmap
```

---

## Building and Testing

### Build Commands

```bash
# Production build (stripped binary)
make build

# Build and run
make run

# Development mode
make dev

# Clean build artifacts
make clean
```

### Manual Build

```bash
# Standard build
go build -o archterm ./cmd/archterm/

# Build with debug info
go build -o archterm ./cmd/archterm/

# Build with version info
go build -ldflags="-X main.version=1.0.0" -o archterm ./cmd/archterm/
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -cover ./...

# Run specific package tests
go test -v ./internal/search/
go test -v ./internal/markdown/
```

### Code Quality

```bash
# Run go vet
make lint

# Format code
make fmt

# Tidy modules
make mod-tidy
```

---

## Adding New Features

### Adding a New API Endpoint

1. **Define the handler** in [`internal/server/server.go`](internal/server/server.go):

```go
func (s *Server) handleNewFeature(w http.ResponseWriter, r *http.Request) {
    // Parse request
    query := r.URL.Query().Get("q")
    
    // Validate input
    if query == "" {
        http.Error(w, "Missing required parameter 'q'", http.StatusBadRequest)
        return
    }
    
    // Process request
    result := s.processNewFeature(query)
    
    // Return response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

2. **Register the route** in [`registerRoutes()`](internal/server/server.go:88):

```go
s.mux.HandleFunc("/api/newfeature", s.handleNewFeature)
```

3. **Add frontend support** in [`webapp/js/terminal.js`](webapp/js/terminal.js):

```javascript
case 'newfeature':
    await this.handleNewFeature(args);
    break;
```

### Adding a New Terminal Command

1. **Add the case** in [`processCommand()`](webapp/js/terminal.js:95):

```javascript
case 'mycommand':
    await this.handleMyCommand(args);
    break;
```

2. **Implement the handler**:

```javascript
async handleMyCommand(args) {
    if (!args[0]) {
        this.print('Usage: mycommand <param>', 'warning');
        return;
    }
    // ... implementation
}
```

3. **Update help text** in [`handleHelp()`](webapp/js/terminal.js:157):

```
  mycommand <param>       Description of my command
```

### Adding a New Internal Package

1. **Create the package directory**:

```bash
mkdir -p internal/newpkg
```

2. **Create the Go file**:

```go
package newpkg

// Your code here
```

3. **Import in server.go**:

```go
import (
    "archterm/internal/newpkg"
)
```

---

## Adding New File Categories

### Step 1: Update the Watcher

Edit [`internal/filesystem/watcher.go`](internal/filesystem/watcher.go:29):

```go
categoryMap: map[string]string{
    // ... existing mappings ...
    ".xyz": "newcategory",  // Add new extension mapping
}
```

### Step 2: Update the Manifest

Edit [`manifest.json`](manifest.json):

```json
{
  "categories": {
    "newcategory": {
      "label": "New Category",
      "extensions": [".xyz", ".abc"],
      "icon": "🆕"
    }
  }
}
```

### Step 3: Update Frontend Styling

Edit [`webapp/css/terminal.css`](webapp/css/terminal.css) to add category-specific colors:

```css
.output-line.code { color: #7dd3f8; }
.output-line.newcategory { color: #ff6b6b; }
```

### Step 4: Update the Watched Paths

Edit [`internal/filesystem/watcher.go`](internal/filesystem/watcher.go:27):

```go
paths: []string{
    "./docs", "./plans", "./cmd", "./internal",
    "./webapp", "./python_parser",
    "./newdir",  // Add new watched directory
}
```

---

## Extending the Markdown Renderer

The markdown renderer is in [`internal/markdown/renderer.go`](internal/markdown/renderer.go). It supports:

### Block-Level Elements

| Element | Syntax | Supported |
|---------|--------|-----------|
| Headers | `# H1`, `## H2`, etc. | Yes |
| Code blocks | ``` ... ``` | Yes |
| Blockquotes | `> text` | Yes |
| Lists | `- item` or `* item` | Yes |
| Horizontal rules | `---`, `***`, `___` | Yes |
| Paragraphs | Plain text | Yes |

### Inline Elements

| Element | Syntax | Supported |
|---------|--------|-----------|
| Bold | `**text**`, `__text__` | Yes |
| Italic | `*text*`, `_text_` | Yes |
| Code | `` `code` `` | Yes |
| Strikethrough | `~~text~~` | Yes |
| Links | `[text](url)` | Yes |
| Images | `![alt](url)` | Yes |

### Adding New Syntax

To add a new inline element (e.g., highlight):

```go
// In inlineFormat() function:
highlightRe := regexp.MustCompile(`==([^=]+)==`)
text = highlightRe.ReplaceAllString(text, `<mark>$1</mark>`)
```

To add a new block element (e.g., callout boxes):

```go
// In Render() function, before paragraph handling:
if strings.HasPrefix(strings.TrimSpace(line), ">>> ") {
    // Handle callout
}
```

---

## API Development

### API Design Principles

- **RESTful** — Use appropriate HTTP methods
- **JSON responses** — All API responses are JSON (except file content)
- **Error handling** — Return proper HTTP status codes
- **Query parameters** — Use for filtering and options

### Adding a New Endpoint

1. **Define the handler** in [`server.go`](internal/server/server.go):

```go
func (s *Server) handleNewEndpoint(w http.ResponseWriter, r *http.Request) {
    // Only allow GET requests
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Get query parameters
    param := r.URL.Query().Get("param")
    
    // Process
    result := s.process(param)
    
    // Respond
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

2. **Register the route**:

```go
s.mux.HandleFunc("/api/newendpoint", s.handleNewEndpoint)
```

### Testing the API

```bash
# Test manifest endpoint
curl http://localhost:8080/api/manifest

# Test search
curl "http://localhost:8080/api/search?q=architecture"

# Test file access
curl http://localhost:8080/api/files/README.md

# Test rendered markdown
curl "http://localhost:8080/api/files/docs/planning/archterm_prd.md?render=true"

# Test related documents
curl "http://localhost:8080/api/related?doc=cmd/archterm/main.go"
```

---

## Frontend Development

### Terminal Class Structure

The terminal is implemented as a class in [`webapp/js/terminal.js`](webapp/js/terminal.js):

```javascript
class Terminal {
    constructor() { ... }      // Initialization
    init() { ... }             // Setup event listeners
    handleKeyDown(e) { ... }   // Process keyboard input
    processCommand(cmd) { ... } // Route commands
    handleHelp() { ... }       // Help command
    handleLS(args) { ... }     // LS command
    handleAccess(args) { ... } // Access command
    handleSearch(args) { ... } // Search command
    handleRelated(args) { ... } // Related command
    handleMetadata(args) { ... } // Metadata command
    handleCat(args) { ... }    // Cat command
    handleView(args) { ... }   // View command
    getFileIcon(file) { ... }  // Get file icon
    getFileColor(file) { ... } // Get file color
    print(text, type) { ... }  // Output to terminal
    printRenderedHTML(html) { ... } // Rendered HTML output
    clear() { ... }            // Clear terminal
    scrollToBottom() { ... }   // Auto-scroll
    handleTab() { ... }        // Tab completion
    handleArrowKeys(e) { ... } // History navigation
}
```

### Adding Terminal Styles

Edit [`webapp/css/terminal.css`](webapp/css/terminal.css):

```css
/* Output line types */
.output-line.command { color: var(--accent); }
.output-line.success { color: #27ae60; }
.output-line.warning { color: #e6c600; }
.output-line.error { color: #e06c75; }
.output-line.info { color: var(--text-primary); }
.output-line.dim { color: var(--text-secondary); }
```

### Adding CSS for Rendered Content

```css
.output-line.rendered-html {
    /* Styles for rendered markdown */
}

.output-line.rendered-html h1,
.output-line.rendered-html h2 {
    /* Header styles */
}

.output-line.rendered-html code {
    /* Inline code styles */
}

.output-line.rendered-html pre {
    /* Code block styles */
}
```

---

## Configuration

### Server Configuration

The server is configured via command-line flags in [`cmd/archterm/main.go`](cmd/archterm/main.go):

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `8080` | Port to listen on |
| `-bind` | `:8080` | Address to bind. Default is `:port` (all interfaces). Examples: `127.0.0.1:9090` (localhost only), `:9090` (all interfaces) |
| `-banner` | `true` | Show network info banner on startup (includes Tailscale IP). Set to `false` to suppress. |
| `-config` | `manifest.json` | Path to manifest file |
| `-manifest` | (empty) | Override manifest path |

### Network Configuration

The network layer is implemented in [`internal/network/detector.go`](internal/network/detector.go) and uses Go's standard library only.

**Key components:**

- `GetNetworkInfo()` — Enumerates all network interfaces and detects Tailscale IPs (in the `100.64.0.0/10` range)
- `GetTailscaleIP()` — Returns the first Tailscale IP, or empty string if not connected
- `GetTailscaleIPs()` — Returns all Tailscale IPs
- `BuildAccessURL(port)` — Constructs a URL for remote access (returns empty string if no Tailscale)
- `PrintStartupBanner(port)` — Prints a formatted banner with Tailscale IP and local interfaces

**Tailscale detection** works by checking all IP addresses on all active network interfaces against the Tailscale subnet allocator CIDR (`100.64.0.0/10`). If any IP falls within this range, the machine is considered to be on a Tailnet.

### Watched Directories

Configured in [`internal/filesystem/watcher.go`](internal/filesystem/watcher.go:27):

```go
paths: []string{
    "./docs",
    "./plans",
    "./cmd",
    "./internal",
    "./webapp",
    "./python_parser",
}
```

### File Categories

Configured in [`internal/filesystem/watcher.go`](internal/filesystem/watcher.go:29):

```go
categoryMap: map[string]string{
    ".pdf":      "documents",
    ".md":       "documents",
    ".go":       "code",
    ".js":       "code",
    // ... etc
}
```

---

## Network & Tailscale

### Architecture

The network layer (`internal/network/detector.go`) uses only Go standard library:

```go
package network

import (
	"fmt"
	"net"
	"sort"
	"strings"
)
```

No external network libraries, no Tailscale SDK dependency. It simply scans network interfaces for IPs in the Tailscale range (`100.64.0.0/10`).

### Startup Banner System

The startup banner is printed during server initialization in [`cmd/archterm/main.go`](cmd/archterm/main.go):

```go
// Print network startup banner
if *showBanner {
	network.PrintStartupBanner(*port)
}
```

The banner displays:
- Tailscale IP (if detected)
- Access URL for remote clients
- Local network interfaces (first 3 non-loopback)
- A warning if no Tailscale is detected

### Binding Behavior

By default, the server binds to all interfaces (`:port`). This is intentional — it makes ArchTerm immediately accessible on the Tailscale network without any configuration.

To restrict to localhost only:
```bash
./archterm -bind 127.0.0.1:8080
```

To bind to a specific interface:
```bash
./archterm -bind 192.168.1.100:8080
```

### Multi-Instance Setup

To run ArchTerm on multiple machines and browse each from one browser:

1. **Install Tailscale** on all machines
2. **Run ArchTerm on each** with a unique port:
   ```bash
   # Machine A (port 8080)
   ./archterm -port 8080
   
   # Machine B (port 8081)
   ./archterm -port 8081
   ```
3. **Access each** from your browser at the Tailscale URL shown in each instance's banner
4. **Create projects** — ArchTerm indexes files on the machine where it runs

### Testing Network Functionality

```bash
# Test: default (all interfaces, banner on)
./archterm -port 8080

# Test: localhost only (no network exposure)
./archterm -port 8080 -bind 127.0.0.1:8080

# Test: custom port on all interfaces
./archterm -port 9090 -bind :9090

# Test: show banner without running server (quick check)
./archterm -port 8080 -banner 2>&1 &
sleep 2 && kill $! 2>/dev/null
```

Verify the output shows the Tailscale IP and access URL. If no Tailscale is detected, the banner will show a warning message instead.

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make net-run` | Build and run on all interfaces (Tailscale-ready) |
| `make net-local` | Build and run on localhost only (no network exposure) |

---

## Deployment

### Production Deployment

```bash
# Build optimized binary
make build

# Copy binary to target server
scp archterm user@server:/opt/archterm/

# Start server (use systemd, supervisor, etc.)
./archterm -port 8080
```

### Docker Deployment

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o archterm ./cmd/archterm/

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/archterm .
COPY --from=builder /app/webapp ./webapp
COPY --from=builder /app/manifest.json .
EXPOSE 8080
CMD ["./archterm", "-port", "8080"]
```

Build and run:

```bash
docker build -t archterm .
docker run -p 8080:8080 archterm
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ARCHTERM_PORT` | `8080` | Server port |
| `ARCHTERM_MANIFEST` | `manifest.json` | Manifest file path |

---

## Troubleshooting

### Development Issues

#### Build fails with "undefined" errors

**Cause:** Missing import or typo in package name.

**Solution:** Check the import path in [`server.go`](internal/server/server.go:3):

```go
import (
    "archterm/internal/markdown"  // Ensure this matches the package path
)
```

#### Server not serving static files

**Cause:** The `handleStatic` function path resolution is broken.

**Solution:** Verify the `webapp/` directory exists relative to the working directory:

```bash
ls webapp/index.html
```

#### File watcher not detecting files

**Cause:** Watched directories don't exist or have wrong permissions.

**Solution:** Check [`internal/filesystem/watcher.go`](internal/filesystem/watcher.go:27):

```go
paths: []string{"./docs", "./plans", ...}
```

Ensure these directories exist and are readable.

### Runtime Issues

#### Port already in use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Use a different port
./archterm -port 9090
```

#### Search returns no results

**Cause:** The search index may be empty or the query has no matches.

**Solution:** Check the server logs for indexing messages:

```
Initial scan complete. Indexed 42 files.
```

If 0 files indexed, check the watched paths.

### Performance Issues

#### Slow search

**Cause:** Large number of files with long content.

**Solution:** The TF-IDF index is built in memory. For large codebases:
- Reduce watched directories
- Exclude large files from indexing
- Consider adding pagination to search results

#### High memory usage

**Cause:** All file contents are loaded into memory for indexing.

**Solution:** The `isBinary()` function skips binary files. For very large projects:
- Add file size limits
- Implement lazy loading
- Consider streaming the index

---

## Release Process

### Version Bumping

1. Update version in [`cmd/archterm/main.go`](cmd/archterm/main.go:17):

```go
const version = "1.1.0"
```

2. Update version in [`manifest.json`](manifest.json):

```json
{
  "version": "1.1.0",
  ...
}
```

### Changelog

1. Document all changes in a `CHANGELOG.md` file
2. Group changes by: Added, Changed, Fixed, Removed

### Tagging

```bash
# Create a tag
git tag -a v1.1.0 -m "Release v1.1.0"

# Push the tag
git push origin v1.1.0
```

### Publishing

1. Build the binary for target platforms:

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -o archterm-linux-amd64 ./cmd/archterm/

# macOS amd64
GOOS=darwin GOARCH=amd64 go build -o archterm-darwin-amd64 ./cmd/archterm/

# macOS arm64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o archterm-darwin-arm64 ./cmd/archterm/

# Windows amd64
GOOS=windows GOARCH=amd64 go build -o archterm-windows-amd64.exe ./cmd/archterm/
```

2. Create archives:

```bash
tar czf archterm-linux-amd64.tar.gz archterm-linux-amd64
tar czf archterm-darwin-amd64.tar.gz archterm-darwin-amd64
tar czf archterm-darwin-arm64.tar.gz archterm-darwin-arm64
```

3. Upload to release page with checksums:

```bash
sha256sum archterm-*.tar.gz > checksums.txt
```

---

## Contributing Guidelines

### Code Style

- Follow [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Run `make fmt` before committing
- Run `make lint` to check for issues
- Keep functions focused and under 50 lines where possible

### Commit Messages

Use conventional commits:

```
feat: add rendered markdown support to access command
fix: resolve port flag not being applied
docs: add user manual and maintainer manual
refactor: simplify handleStatic function
test: add unit tests for markdown renderer
```

### Pull Requests

1. Create a feature branch
2. Make changes and test locally
3. Update documentation as needed
4. Submit PR with clear description
5. Request review from maintainers

---

## Appendix: Go Package Reference

### `archterm/cmd/archterm`

Entry point. Parses flags, initializes components, starts server.

### `archterm/internal/config`

Data structures:
- `Config` — Application configuration
- `ManifestEntry` — Single file entry
- `Manifest` — Top-level manifest

### `archterm/internal/server`

HTTP server:
- `Server` — HTTP server struct
- `NewServer()` — Constructor
- `Start(addr)` — Start listening on the given address. The address format is `":port"` (all interfaces) or `"host:port"` (specific interface). Default when called is `:port`.
- `Shutdown()` — Graceful shutdown
- `handleStatic()` — Serve web assets
- `handleManifest()` — `/api/manifest`
- `handleFiles()` — `/api/files/`
- `handleLS()` — `/api/ls`
- `handleSearch()` — `/api/search`
- `handleRelated()` — `/api/related`
- `handleConcepts()` — `/api/concepts`

### `archterm/internal/filesystem`

File watcher:
- `Watcher` — File watcher struct
- `NewWatcher()` — Constructor
- `Start()` — Start watching
- `walkFn()` — filepath.Walk callback
- `registerFile()` — Register file in manifest
- `getCategory()` — Get file category by extension
- `isBinary()` — Detect binary files

### `archterm/internal/search`

TF-IDF search:
- `SemanticSearcher` — Search engine
- `Document` — Indexed document
- `ScoredDocument` — Search result
- `NewSemanticSearcher()` — Constructor
- `AddDocument()` — Add document to index
- `Search()` — Execute search query

### `archterm/internal/markdown`

Markdown renderer:
- `Renderer` — Markdown renderer
- `NewRenderer()` — Constructor
- `Render()` — Convert markdown to HTML

### `archterm/internal/network`

Network detection and banner:
- `GetNetworkInfo()` — Enumerate interfaces and detect Tailscale IPs
- `GetTailscaleIP()` — Return first Tailscale IP (or empty string)
- `GetTailscaleIPs()` — Return all Tailscale IPs
- `BuildAccessURL(port)` — Construct remote access URL
- `PrintStartupBanner(port)` — Print formatted startup banner
