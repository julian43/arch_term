# ArchTerm - The Architect's Codex

> A terminal-style file browser and document viewer with semantic search capabilities.

[![Go Version](https://img.shields.io/badge/go-1.21+-brightgreen.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Overview

ArchTerm is a high-performance, terminal-based file browser and document viewer designed for software architects and developers. Built with Go for the backend and vanilla JavaScript for the frontend, it provides a lightweight, dependency-free experience for exploring and searching through codebases and documentation.

### Key Features

- **Terminal-Style Interface**: Authentic terminal experience with CRT monitor aesthetics
- **Semantic Search**: TF-IDF based search engine for finding relevant files
- **Multi-Format Support**: View PDF, Markdown, and source code files
- **Zero External Dependencies**: Pure Go backend using only standard library
- **Real-time File Watching**: Automatic indexing of project files
- **Command Palette**: Quick navigation with keyboard-driven commands

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (HTML/CSS/JS)                    │
│  ┌─────────────┐ ┌──────────────┐ ┌─────────────────────────┐  │
│  │ Terminal UI │ │ Command Palet│ │    File Viewer Overlay  │  │
│  │   (CRT      │ │   (ls,       │ │   (PDF/MD/Code       │  │
│  │  Effects)   │ │    search)   │ │    Rendering)          │  │
│  └─────────────┘ └──────────────┘ └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │ WebSocket / REST
┌─────────────────────────────────────────────────────────────────┐
│                        Backend (Go)                              │
│  ┌─────────────┐ ┌──────────────┐ ┌─────────────────────────┐  │
│  │   HTTP      │ │  File       │ │    Semantic           │  │
│  │  Server     │ │  Watcher    │ │    Search Engine      │  │
│  │(net/http)   │ │ (fsnotify)  │ │    (TF-IDF)           │  │
│  └─────────────┘ └──────────────┘ └─────────────────────────┘  │
│  ┌─────────────┐ ┌──────────────┐ ┌─────────────────────────┐  │
│  │  Command    │ │  Manifest   │ │    Document            │  │
│  │  Parser     │ │  Manager    │ │    Store                │  │
│  │            │ │             │ │                        │  │
│  └─────────────┘ └──────────────┘ └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Component Overview

| Component | Technology | Purpose |
|-----------|------------|---------|
| HTTP Server | Go `net/http` | Handle HTTP requests/responses |
| File Watcher | Go `os/fs` | Monitor file system changes |
| Search Engine | Custom TF-IDF | Semantic search implementation |
| Command Parser | Custom | Parse and execute terminal commands |
| Frontend | Vanilla JS | Browser-based terminal emulation |

## Installation

### Prerequisites

- Go 1.21 or higher
- Python 3.8+ (optional, for document parsing)
- Modern web browser (Chrome, Firefox, Edge, Safari)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/archterm.git
   cd archterm
   ```

2. **Initialize Go modules**
   ```bash
   go mod init archterm
   go mod tidy
   ```

3. **Build the application**
   ```bash
   go build -o archterm ./cmd/archterm/main.go
   ```

4. **Run the server**
   ```bash
   ./archterm -port 8080
   ```

5. **Open your browser**
   Navigate to `http://localhost:8080`

## Usage

### Available Commands

| Command | Description | Example |
|---------|-------------|---------|
| `help` | Show available commands | `help` |
| `ls [category]` | List files by category | `ls documents` |
| `access <file>` | Open a file in viewer | `access README.md` |
| `search <query>` | Search indexed files | `search architecture` |
| `clear` | Clear terminal | `clear` |
| `cat <file>` | Display file contents | `cat go.mod` |
| `pwd` | Show current directory | `pwd` |
| `whoami` | Show current user | `whoami` |
| `date` | Show current date/time | `date` |
| `exit` | Close terminal | `exit` |

### Categories

| Category | Extensions | Description |
|----------|-----------|-------------|
| `documents` | `.pdf`, `.md`, `.txt` | Document files |
| `code` | `.go`, `.js`, `.ts`, `.py` | Source code files |
| `config` | `.json`, `.yaml`, `.toml` | Configuration files |
| `images` | `.png`, `.jpg`, `.svg` | Image files |
| `data` | `.csv`, `.xml`, `.db` | Data files |
| `all` | - | Show all files |

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Enter` | Execute command |
| `Tab` | Auto-complete |
| `↑` / `↓` | Browse command history |
| `Ctrl+L` | Clear terminal |
| `Escape` | Close viewer |

## API Reference

### Search Files

```http
GET /api/search?q=query
```

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

### List Files

```http
GET /api/ls?category=all
```

**Response:**
```json
{
  "category": "all",
  "files": [
    {
      "name": "README.md",
      "path": "README.md",
      "category": "documents",
      "metadata": {
        "size": "4096",
        "modified": "2024-01-15T10:30:00"
      }
    }
  ]
}
```

### Get File Content

```http
GET /api/files/{filepath}
```

## Semantic Search

ArchTerm includes a built-in semantic search engine powered by TF-IDF (Term Frequency-Inverse Document Frequency) algorithm.

### How It Works

1. **Tokenization**: Text is tokenized and stop words are removed
2. **Indexing**: Documents are indexed with term frequencies
3. **Scoring**: Query terms are scored against the index
4. **Ranking**: Results are sorted by relevance score

### Adding Documents

```go
searcher := search.NewSemanticSearcher()
searcher.AddDocument(search.Document{
    ID:       "docs/architecture.md",
    Title:    "Architecture Documentation",
    Content:  "System architecture...",
    Source:   "docs/architecture.md",
    Category: "documents",
})
```

### Performing a Search

```go
results := searcher.Search("system design", 10)
// Returns []SearchResult{...}
```

## Project Structure

```
archterm/
├── cmd/
│   └── archterm/          # Main application entry point
│       └── main.go        # CLI entry point
├── internal/
│   ├── command/            # Command parser
│   │   └── parser.go      # Command parsing logic
│   ├── config/             # Configuration management
│   │   └── config.go      # Manifest configuration
│   ├── filesystem/         # File system operations
│   │   └── watcher.go     # File watching logic
│   ├── search/             # Semantic search
│   │   └── semantic.go    # TF-IDF search engine
│   └── server/             # HTTP server
│       └── server.go      # HTTP handlers
├── webapp/                 # Frontend assets
│   ├── index.html          # Main HTML file
│   ├── css/
│   │   └── terminal.css   # CRT terminal styles
│   └── js/
│       └── terminal.js    # Terminal emulation
├── docs/                   # Documentation
│   └── planning/
│       └── archterm_prd.md
├── go.mod
├── go.sum
├── manifest.json           # File manifest
└── README.md
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ARCHTERM_PORT` | HTTP server port | `8080` |
| `ARCHTERM_WATCH_PATHS` | Comma-separated watch paths | `./docs,./plans` |
| `ARCHTERM_INDEX_PATHS` | Comma-separated index paths | `./docs,./plans` |

### Manifest Configuration

The `manifest.json` file defines file categories and their properties:

```json
{
  "version": "1.0.0",
  "categories": {
    "documents": { "label": "Documents", "extensions": [".pdf", ".md", ".txt"] },
    "code": { "label": "Code", "extensions": [".go", ".js", ".ts", ".py"] },
    "config": { "label": "Configuration", "extensions": [".json", ".yaml", ".toml"] }
  }
}
```

## Development

### Adding a New Command

1. Define the command type in `internal/command/parser.go`:
   ```go
   const (
       // ... existing commands
       CmdNewCommand CommandType = iota
   )
   ```

2. Register the command handler:
   ```go
   p.handlers[CmdNewCommand] = p.handleNewCommand
   ```

3. Add the command to the help texts:
   ```go
   func (p *Parser) GetHelpTexts() []HelpText {
       return []HelpText{
           // ... existing help texts
           {Command: "newcmd", Description: "New command description", Usage: "newcmd [args]"},
       }
   }
   ```

### Adding a New File Type

1. Add the extension to `manifest.json`:
   ```json
   {
     "categories": {
       "newtype": {
         "label": "New Type",
         "extensions": [".ext"]
       }
     }
   }
   ```

2. Update the file watcher in `internal/filesystem/watcher.go`:
   ```go
   case ".ext":
       category = "newtype"
   ```

### Running Tests

```bash
go test ./...
```

### Linting

```bash
go install golang.org/x/lint/golint@latest
golint ./...
```

## Roadmap

- [x] Basic file browsing
- [x] Semantic search
- [ ] LLM-powered code analysis (Ollama integration)
- [ ] Real-time collaboration
- [ ] Plugin system
- [x] Terminal aesthetic
- [ ] Mobile responsiveness
- [ ] Dark/Light theme toggle

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit changes (`git commit -am 'Add feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/archterm.git
cd archterm

# Install dependencies
go mod tidy

# Run in watch mode
go run ./cmd/archterm/main.go -port 8080
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by classic terminal interfaces and CRT monitor aesthetics
- TF-IDF algorithm implementation for semantic search
- Go standard library for zero-dependency backend

---

*The Architect's Codex - Version 1.0.0*
