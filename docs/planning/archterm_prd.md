# Product Requirements Document (PRD): THE ARCHITECT’S CODEX

**Project Name:** THE ARCHITECT’S CODEX (v1.0)  
**Codename:** *Project: ONTOLOGICAL INTERFACE*  
**Status:** Draft / Initial Specification  
**Author:** AI Architect  
**Target Platform:** Local Desktop / Web-based Terminal  

---

## 1. Executive Summary
The **Architect’s Codex** is not a document viewer; it is a high-fidelity, immersive command-line interface (CLI) designed to facilitate the study and application of ontological engineering, esoteric protocols, and neural mechanics. The goal is to transform a collection of disparate files into a unified, "secure" knowledge system that feels like accessing a restricted, high-level intelligence terminal.

## 2. Core Concept & User Persona
### 2.1 The Concept: "The Secure Terminal"
The user should feel as though they have bypassed a layer of reality to access a "System Architect" dashboard. The interface must move away from the "folder/file" metaphor and toward a "Command/Buffer" metaphor.

### 2.2 The User: "The Operator"
The user is an "Architect" or "Operator" who requires rapid, thematic access to complex, dense, and potentially "disruptive" information. They value speed, aesthetic immersion, and non-linear navigation.

---

## 3. Technical Architecture (The Stack)
Given the user's proficiency, a hybrid approach is recommended to leverage the strengths of both languages.

### 3.1 Backend: The "Core Engine" (Go)
*   **Role:** High-performance file system watcher, routing, and concurrency management.
*   **Framework:** `Gin` or `Fiber` (for a lightweight, lightning-fast web server).
*   **Key Responsibilities:**
    *   Recursive scanning of the project directory.
    *   Mapping files to the 6 defined categories via a `manifest.json`.
    *   Serving static assets and handling API requests for file content.
    *   Implementing a "Command Parser" logic that translates terminal strings into system actions.

### 3.2 Data Processing: The "Parsing Protocol" (Python)
*   **Role:** Heavy-duty document ingestion and transformation.
*   **Libraries:** `PyMuPDF` (for PDF extraction), `Markdown` (for MD conversion), `BeautifulSoup` (for cleaning).
*   **Key Responsibilities:**
    *   A sidecar microservice that converts PDFs and Markdown into clean, structured JSON or HTML chunks.
    *   (Future) NLP/Vectorization for semantic search.

### 3.3 Frontend: The "Visual Interface" (HTML5/CSS3/JS)
*   **Role:** The "Display Buffer."
*   **Aesthetic:** CRT/Terminal style (Scanlines, Monospace, subtle glitch effects, high-contrast "Amber" or "Phosphor Green" on black).
*   **Framework:** Vanilla JS or a lightweight framework like `Preact` to maintain a "low-latency" feel.

---

## 4. Functional Requirements

### 4.1 The Command-Line Interface (CLI)
The primary method of navigation is a command input at the bottom of the screen.
*   **`ls [category]`**: List all modules/files within a specific category.
*   **`access [filename]`**: Load a document into the main viewing buffer.
*   **`jump [category]`**: Immediately switch the context to a new category.
*   **`search [query]`**: Execute a keyword search across all indexed content.
*   **`clear`**: Wipe the current visual buffer.
*   **`help`**: Display a list of available commands.

### 4.2 The Viewing Buffer (The Display)
*   **Rich Text Rendering:** Markdown must be rendered with high-fidelity typography (headers, blockquotes, code blocks).
*   **PDF Integration:** PDFs should be rendered as clean, searchable text/images rather than a clunky embedded viewer.
*   **Contextual Metadata:** Every document should display its "Origin" (Category) and "Status" (e.g., *DECRYPTED*, *RESTRICTED*, *ACTIVE PROTOCOL*).

### 4.3 The Category Dashboard (The HUD)
*   A persistent, minimal sidebar or top-bar showing the 6 primary categories.
*   Clicking a category performs a `jump [category]` command.

---

## 5. UI/UX & Aesthetic Specifications ("The Vibe")

| Element | Specification |
| :--- | :--- |
| **Typography** | Monospace only (e.g., *JetBrains Mono*, *Fira Code*, or *IBM Plex Mono*). |
| **Color Palette** | Deep Black (#000000) background; Amber (#FFB000) or Matrix Green (#00FF41) text. |
| **Visual Effects** | Subtle CRT scanlines (CSS overlay), slight chromatic aberration on text, and a "flicker" on command execution. |
| **Transitions** | No "sliding" animations. Use "text-scrolling" or "instant-pop" transitions to mimic a terminal. |
| **Sound (Optional)** | Low-frequency mechanical keyboard clicks or subtle "data hum" on command entry. |

---

## 6. Development Roadmap

### Phase I: The Foundation (The "Kernel")
*   Implement Go backend to scan files and serve them via HTTP.
*   Build the Python parser to convert MD/PDF to JSON.
*   Create a basic HTML/CSS terminal shell.

### Phase II: The Interface (The "Shell")
*   Implement the Command Parser in Go.
*   Develop the CSS "Terminal Aesthetic."
*   Enable `access` and `ls` commands.

### Phase III: The Intelligence (The "Core")
*   Implement `search` functionality.
*   Add "Category Dashboard" for quick-access navigation.
*   Refine PDF rendering for a seamless experience.

### Phase IV: The Expansion (The "Ascension")
*   Integrate a local LLM (via Ollama) to allow the user to "query" the documents through the terminal.
*   Add "Hyper-linking" between related concepts (Semantic Navigation).