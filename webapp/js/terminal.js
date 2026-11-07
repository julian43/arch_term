/**
 * ArchTerm Terminal JavaScript - Phase III Enhanced
 * Handles command parsing, API communication, rich content viewing
 */

class Terminal {
    constructor() {
        this.output = document.getElementById('output');
        this.input = document.getElementById('commandInput');
        this.commandHistory = document.getElementById('commandHistory');
        this.statusText = document.getElementById('statusText');
        this.fileCount = document.getElementById('fileCount');
        this.history = [];
        this.currentIndex = -1;
        this.isProcessing = false;
        this.currentViewer = null;
        this.activeProject = null;
        
        this.init();
    }

    init() {
        // Focus input on click
        document.getElementById('terminal').addEventListener('click', () => {
            this.input.focus();
        });

        // Handle input
        this.input.addEventListener('keydown', (e) => this.handleKeyDown(e));
        
        // Handle paste
        this.input.addEventListener('paste', (e) => {
            e.preventDefault();
            const text = e.clipboardData.getData('text/plain');
            document.execCommand('insertText', false, text);
        });

        // Close viewer on Escape
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.currentViewer) {
                this.closeViewer();
            }
        });

        // Load initial data
        this.print('ArchTerm v1.0.0 initialized', 'info');
        this.print('Type "help" for available commands.', 'info');
        this.print('');
        
        // Load manifest and projects
        this.loadManifest();
        this.loadProjects();
    }

    async loadManifest() {
        try {
            const response = await fetch('/api/manifest');
            if (response.ok) {
                const data = await response.json();
                this.print(`Loaded manifest: ${data.files?.length || 0} files indexed`, 'success');
                this.fileCount.textContent = `${data.files?.length || 0} files indexed`;
            }
        } catch (error) {
            console.error('Failed to load manifest:', error);
            this.print('Warning: Could not load manifest', 'warning');
        }
    }

    // ===== PROJECT FUNCTIONS =====

    async loadProjects() {
        try {
            const response = await fetch('/api/projects');
            if (response.ok) {
                const data = await response.json();
                this.activeProject = data.active;
                this.updateProjectPrompt();
                if (data.projects.length === 0) {
                    this.print('No projects configured. Use "project create <name> <path1> [path2...]" to create one.', 'warning');
                } else {
                    this.print(`Projects loaded: ${data.projects.length} configured, "${data.active?.name || 'none'}" active`, 'info');
                }
            }
        } catch (error) {
            console.error('Failed to load projects:', error);
        }
    }

    updateProjectPrompt() {
        const titleEl = document.querySelector('.terminal-header .title');
        if (titleEl) {
            const projectPrefix = this.activeProject ? `[${this.activeProject}] ` : '';
            titleEl.textContent = `archterm@architect: ~ ${projectPrefix}`;
        }
    }

    async handleProjectCommand(args) {
        const cmd = args[0]?.toLowerCase();
        const rest = args.slice(1);

        switch (cmd) {
            case 'list':
            case 'ls':
                await this.listProjects();
                break;
            case 'create':
            case 'new':
                await this.createProject(rest);
                break;
            case 'delete':
            case 'del':
                await this.deleteProject(rest);
                break;
            case 'use':
            case 'activate':
                await this.activateProject(rest);
                break;
            case 'update':
            case 'set':
                await this.updateProjectPaths(rest);
                break;
            case 'scan':
                await this.scanProject(rest);
                break;
            case 'current':
            case 'active':
                this.showActiveProject();
                break;
            default:
                this.print('Unknown project command. Use: list, create, delete, use, update, scan, current', 'warning');
        }
    }

    async listProjects() {
        try {
            const response = await fetch('/api/projects');
            if (!response.ok) return;
            const data = await response.json();
            const projects = data.projects || [];

            this.print('\nConfigured Projects:', 'success');
            this.print('─'.repeat(50), 'dim');

            if (projects.length === 0) {
                this.print('  No projects configured.', 'dim');
            } else {
                projects.forEach(p => {
                    const marker = p.name === data.active?.name ? ' ▶ ' : '   ';
                    const paths = (p.paths || p.Paths || []).join(', ');
                    this.print(`${marker}[${p.name}] (${paths})`, p.name === data.active?.name ? 'success' : 'info');
                });
            }
            this.print('');
        } catch (error) {
            this.print(`Error listing projects: ${error.message}`, 'error');
        }
    }

    async createProject(args) {
        if (args.length < 2) {
            this.print('Usage: project create <name> <path1> [path2 ...]', 'warning');
            return;
        }
        const name = args[0];
        const paths = args.slice(1);

        try {
            const response = await fetch('/api/projects', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, paths })
            });
            const data = await response.json();
            if (response.ok) {
                this.print(`Project "${name}" created with ${paths.length} path(s)`, 'success');
                this.loadProjects();
            } else {
                this.print(`Error: ${data.message || 'Failed to create project'}`, 'error');
            }
        } catch (error) {
            this.print(`Error: ${error.message}`, 'error');
        }
    }

    async deleteProject(args) {
        if (args.length < 1) {
            this.print('Usage: project delete <name>', 'warning');
            return;
        }
        const name = args[0];

        if (!confirm(`Delete project "${name}"?`)) {
            this.print('Cancelled.', 'info');
            return;
        }

        try {
            const response = await fetch(`/api/projects/${name}`, { method: 'DELETE' });
            const data = await response.json();
            if (response.ok) {
                this.print(`Project "${name}" deleted`, 'success');
                if (this.activeProject === name) {
                    this.activeProject = null;
                    this.updateProjectPrompt();
                }
            } else {
                this.print(`Error: ${data.message || 'Failed to delete project'}`, 'error');
            }
        } catch (error) {
            this.print(`Error: ${error.message}`, 'error');
        }
    }

    async activateProject(args) {
        if (args.length < 1) {
            this.print('Usage: project use <name>', 'warning');
            return;
        }
        const name = args[0];

        try {
            const response = await fetch(`/api/projects/${name}/activate`, { method: 'POST' });
            const data = await response.json();
            if (response.ok) {
                this.print(`Project "${name}" activated`, 'success');
                this.activeProject = name;
                this.updateProjectPrompt();
            } else {
                this.print(`Error: ${data.message || 'Failed to activate project'}`, 'error');
            }
        } catch (error) {
            this.print(`Error: ${error.message}`, 'error');
        }
    }

    async updateProjectPaths(args) {
        if (args.length < 2) {
            this.print('Usage: project update <name> <path1> [path2 ...]', 'warning');
            return;
        }
        const name = args[0];
        const paths = args.slice(1);

        try {
            const response = await fetch(`/api/projects/${name}/paths`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ paths })
            });
            const data = await response.json();
            if (response.ok) {
                this.print(`Project "${name}" paths updated`, 'success');
            } else {
                this.print(`Error: ${data.message || 'Failed to update paths'}`, 'error');
            }
        } catch (error) {
            this.print(`Error: ${error.message}`, 'error');
        }
    }

    async scanProject(args) {
        const name = args[0];
        if (!name) {
            this.print('Usage: project scan <name>', 'warning');
            return;
        }

        this.print(`Scanning project "${name}"...`, 'info');

        try {
            const response = await fetch(`/api/projects/${name}/scan`, { method: 'POST' });
            const data = await response.json();
            if (response.ok) {
                this.print(`Scanned ${data.count} files in project "${data.project}"`, 'success');
            } else {
                this.print(`Error: ${data.message || 'Failed to scan project'}`, 'error');
            }
        } catch (error) {
            this.print(`Error: ${error.message}`, 'error');
        }
    }

    showActiveProject() {
        const active = this.activeProject;
        if (active) {
            this.print(`Active project: ${active}`, 'info');
        } else {
            this.print('No active project. Use "project use <name>" to activate one.', 'warning');
        }
    }

    handleKeyDown(e) {
        if (e.key === 'Enter') {
            e.preventDefault();
            const command = this.input.value.trim();
            
            if (command) {
                this.history.push(command);
                this.currentIndex = this.history.length;
                this.print(`arch ${command}`, 'command');
                this.processCommand(command);
            } else {
                this.print('');
            }
            
            this.input.value = '';
            this.scrollToBottom();
        } else if (e.key === 'Tab') {
            e.preventDefault();
            this.handleTab();
        } else if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
            e.preventDefault();
            this.handleArrowKeys(e);
        } else if (e.key === 'l' && e.ctrlKey) {
            e.preventDefault();
            this.clear();
        }
    }

    async processCommand(command) {
        const parts = command.split(' ');
        const cmd = parts[0].toLowerCase();
        const args = parts.slice(1);

        this.statusText.textContent = 'BUSY';
        this.isProcessing = true;

        try {
            switch (cmd) {
                case 'help':
                    this.handleHelp(args);
                    break;
                case 'ls':
                    await this.handleLS(args);
                    break;
                case 'access':
                    await this.handleAccess(args);
                    break;
                case 'search':
                    await this.handleSearch(args);
                    break;
                case 'clear':
                    this.clear();
                    break;
                case 'cat':
                    await this.handleCat(args);
                    break;
                case 'view':
                    await this.handleView(args);
                    break;
                case 'metadata':
                    await this.handleMetadata(args);
                    break;
                case 'related':
                    await this.handleRelated(args);
                    break;
                case 'pwd':
                    this.print(process.cwd(), 'info');
                    break;
                case 'whoami':
                    this.print('arch', 'info');
                    break;
                case 'date':
                    this.print(new Date().toString(), 'info');
                    break;
                case 'exit':
                    window.close();
                    break;
                case 'project':
                    await this.handleProjectCommand(args);
                    break;
                default:
                    this.print(`Unknown command: ${cmd}`, 'error');
                    this.print('Type "help" for available commands.', 'warning');
            }
        } catch (error) {
            console.error('Command processing error:', error);
            this.print(`Error: ${error.message}`, 'error');
        } finally {
            this.statusText.textContent = 'READY';
            this.isProcessing = false;
        }
    }

    handleHelp() {
        const help = `
Available Commands:
  help                    Show this help message
  ls [category]           List files by category
  access <filename>       Open a file in viewer
  access <filename> render  Open markdown file with rendered HTML
  view <filename>         View file with rich rendering
  metadata <filename>     Show file metadata
  related <filename>      Find related documents by concept similarity
  search <query>          Search files
  project <subcommand>    Manage projects (list, create, use, delete, update, scan, current)
  clear                   Clear the terminal
  cat <filename>          Display file contents
  pwd                     Print working directory
  whoami                  Display current user
  date                    Display current date/time
  exit                    Exit the terminal

Project Subcommands:
  project list            List all projects
  project create <name> <paths...>  Create a new project
  project use <name>      Activate a project
  project delete <name>   Delete a project
  project update <name> <paths...>  Update project paths
  project scan <name>     Scan project files
  project current         Show active project

Categories:
  documents    Document files (.pdf, .md, .txt)
  code         Source code files (.go, .js, .py, etc.)
  config       Configuration files (.json, .yaml, .toml)
  all          Show all files (default)

Examples:
  ls documents
  access README.md
  access docs/planning/archterm_prd.md render
  view docs/planning/archterm_prd.md
  metadata go.mod
  related cmd/archterm/main.go
  search "architecture"
  project create myproject ~/src/myproject ~/docs
  project use myproject`;
        this.print(help, 'info');
    }

    async handleLS(args) {
        const category = args[0] || 'all';
        try {
            const response = await fetch(`/api/ls?category=${encodeURIComponent(category)}`);
            if (response.ok) {
                const data = await response.json();
                this.print(`\nFiles in "${category}":`, 'success');
                this.print('─'.repeat(50), 'dim');
                
                if (data.files && data.files.length > 0) {
                    data.files.forEach(file => {
                        const icon = this.getFileIcon(file);
                        const color = this.getFileColor(file);
                        const name = file.Name || file.name || 'unknown';
                        const category = file.Category || file.category || 'N/A';
                        this.print(`  ${icon} ${name}${' '.repeat(Math.max(1, 30 - name.length))} [${category}]`, color);
                    });
                } else {
                    this.print('  No files found.', 'dim');
                }
                this.print('');
            }
        } catch (error) {
            this.print(`Error: ${error.message}`, 'error');
        }
    }

    async handleAccess(args) {
        if (!args[0]) {
            this.print('Usage: access <filename> [render]', 'warning');
            return;
        }

        const filename = args[0];
        const renderMode = args.includes('render') || args.includes('--render');
        try {
            const url = renderMode
                ? `/api/files/${encodeURIComponent(filename)}?render=true`
                : `/api/files/${encodeURIComponent(filename)}`;
            const response = await fetch(url);
            if (response.ok) {
                const content = await response.text();
                this.print(`\n--- ${filename} ---`, 'success');
                if (renderMode) {
                    this.printRenderedHTML(content);
                } else {
                    this.print(content, 'info');
                }
                this.print('---', 'dim');
            } else {
                this.print(`File not found: ${filename}`, 'error');
            }
        } catch (error) {
            this.print(`Error: ${error.message}`, 'error');
        }
    }

    async handleView(args) {
        if (!args[0]) {
            this.print('Usage: view <filename>', 'warning');
            return;
        }

        const filename = args[0];
        this.print(`\nViewing: ${filename}`, 'info');

        try {
            const response = await fetch(`/api/files/${encodeURIComponent(filename)}`);
            if (!response.ok) {
                this.print(`File not found: ${filename}`, 'error');
                return;
            }

            const content = await response.text();
            const ext = this.getFileExtension(filename);
            
            // Open viewer based on file type
            if (ext === 'pdf') {
                this.openPdfViewer(filename, content);
            } else if (['md', 'markdown'].includes(ext)) {
                this.openMarkdownViewer(filename, content);
            } else {
                this.openCodeViewer(filename, content);
            }
        } catch (error) {
            this.print(`Error: ${error.message}`, 'error');
        }
    }

    async handleMetadata(args) {
        if (!args[0]) {
            this.print('Usage: metadata <filename>', 'warning');
            return;
        }

        const filename = args[0];
        this.print(`\nFetching metadata for: ${filename}...`, 'info');

        try {
            const response = await fetch('/api/manifest');
            if (!response.ok) throw new Error('Failed to fetch manifest');
            
            const manifest = await response.json();
            const file = manifest.files?.find(f => f.name === filename || f.path?.endsWith(filename));
            
            if (file) {
                const meta = `
File: ${file.name}
Path: ${file.path || 'N/A'}
Category: ${file.category || 'N/A'}
Size: ${file.size || 'N/A'} bytes
Origin: ${file.origin || 'local'}
Status: ${file.status || 'active'}
                `.trim();
                this.print(meta, 'info');
            } else {
                this.print(`File not found in manifest: ${filename}`, 'warning');
            }
        } catch (error) {
            this.print(`Metadata error: ${error.message}`, 'error');
        }
    }

    async handleSearch(args) {
        if (!args.length) {
            this.print('Usage: search <query>', 'warning');
            return;
        }

        const query = args.join(' ');
        this.print(`\nSearching for "${query}"...`, 'info');

        try {
            const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
            if (response.ok) {
                const data = await response.json();
                this.print(`Found ${data.results?.length || 0} results for "${data.query}"`, 'success');
                
                if (data.results && data.results.length > 0) {
                    data.results.forEach((file, i) => {
                        this.print(`  ${i + 1}. ${file.name}`, 'info');
                    });
                }
            }
        } catch (error) {
            this.print(`Search error: ${error.message}`, 'error');
        }
    }

    async handleRelated(args) {
        if (!args[0]) {
            this.print('Usage: related <filename>', 'warning');
            return;
        }

        const filename = args[0];
        this.print(`\nFinding related documents for "${filename}"...`, 'info');

        try {
            const response = await fetch(`/api/related?doc=${encodeURIComponent(filename)}&max=10`);
            if (!response.ok) {
                this.print(`Error: Could not fetch related documents`, 'error');
                return;
            }

            const data = await response.json();
            const rels = data.relationships || [];
            
            this.print(`\nRelated documents for "${data.source}":`, 'success');
            this.print('─'.repeat(50), 'dim');

            if (rels.length === 0) {
                this.print('  No related documents found.', 'dim');
            } else {
                this.print(`  Found ${rels.length} related document(s):\n`, 'info');
                rels.forEach((rel, i) => {
                    const strength = Math.round(rel.strength * 100);
                    const conceptList = rel.shared_concepts?.slice(0, 5).join(', ') || 'N/A';
                    const bar = '█'.repeat(Math.floor(strength / 10)) + '░'.repeat(10 - Math.floor(strength / 10));
                    this.print(`  ${i + 1}. [${strength}%] ${rel.target_title}`, 'info');
                    this.print(`     Path: ${rel.target_source}`, 'dim');
                    this.print(`     Concepts: [${bar}] ${conceptList}`, 'dim');
                });
            }
            this.print('');
            this.print('Tip: Use "access <filename>" to view a related document.', 'warning');
        } catch (error) {
            this.print(`Related error: ${error.message}`, 'error');
        }
    }

    async handleCat(args) {
        if (!args[0]) {
            this.print('Usage: cat <filename>', 'warning');
            return;
        }
        await this.handleAccess(args);
    }

    // ===== VIEWER FUNCTIONS =====

    openCodeViewer(filename, content) {
        const ext = this.getFileExtension(filename);
        const highlighted = this.highlightSyntax(content, ext);
        
        const viewer = document.getElementById('viewerOverlay');
        const contentEl = document.getElementById('viewerContent');
        const titleEl = document.getElementById('viewerTitle');
        const metaEl = document.getElementById('viewerMeta');
        
        titleEl.textContent = `📄 ${filename}`;
        metaEl.innerHTML = `
            <span class="meta-item"><span class="meta-label">Type:</span> code</span>
            <span class="meta-item"><span class="meta-label">Lines:</span> ${content.split('\n').length}</span>
        `;
        
        contentEl.innerHTML = `<pre class="code-viewer">${highlighted}</pre>`;
        viewer.classList.add('active');
        this.currentViewer = 'code';
    }

    openMarkdownViewer(filename, content) {
        const rendered = this.renderMarkdown(content);
        
        const viewer = document.getElementById('viewerOverlay');
        const contentEl = document.getElementById('viewerContent');
        const titleEl = document.getElementById('viewerTitle');
        const metaEl = document.getElementById('viewerMeta');
        
        titleEl.textContent = `📝 ${filename}`;
        metaEl.innerHTML = `
            <span class="meta-item"><span class="meta-label">Type:</span> markdown</span>
            <span class="meta-item"><span class="meta-label">Lines:</span> ${content.split('\n').length}</span>
        `;
        
        contentEl.innerHTML = `<div class="markdown-body">${rendered}</div>`;
        viewer.classList.add('active');
        this.currentViewer = 'markdown';
    }

    openPdfViewer(filename, content) {
        const viewer = document.getElementById('viewerOverlay');
        const contentEl = document.getElementById('viewerContent');
        const titleEl = document.getElementById('viewerTitle');
        const metaEl = document.getElementById('viewerMeta');
        
        titleEl.textContent = `📄 ${filename}`;
        metaEl.innerHTML = `
            <span class="meta-item"><span class="meta-label">Type:</span> pdf</span>
            <span class="meta-item"><span class="meta-label">Status:</span> loading...</span>
        `;
        
        // For PDFs, we'd typically use pdf.js to render to canvas
        // For now, show a message
        contentEl.innerHTML = `
            <div style="display: flex; justify-content: center; align-items: center; height: 100%;">
                <div style="text-align: center;">
                    <div class="spinner"></div>
                    <p>PDF rendering requires server-side processing.</p>
                    <p style="color: var(--text-secondary); font-size: 12px; margin-top: 10px;">
                        Tip: Use the Python parser to convert PDF to text first.
                    </p>
                </div>
            </div>
        `;
        
        viewer.classList.add('active');
        this.currentViewer = 'pdf';
    }

    closeViewer() {
        const viewer = document.getElementById('viewerOverlay');
        if (viewer) {
            viewer.classList.remove('active');
        }
        this.currentViewer = null;
    }

    // ===== MARKDOWN RENDERING (Simple Implementation) =====

    renderMarkdown(text) {
        let html = text;
        
        // Code blocks (```)
        html = html.replace(/```(\w+)?\n?([\s\S]*?)```/g, (match, lang, code) => {
            return `<pre class="code-block"><code class="language-${lang || 'text'}">${this.escapeHtml(code.trim())}</code></pre>`;
        });
        
        // Headers (fixed: removed broken headingRegex calls)
        html = html.replace(/^(#{1,6})\s(.+)$/gm, (match, hashes, content) => {
            const level = hashes.length;
            return `<h${level}>${content}</h${level}>`;
        });
        
        // Bold and italic
        html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
        html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');
        
        // Inline code
        html = html.replace(/`([^`]+)`/g, '<code>$1</code>');
        
        // Blockquote
        html = html.replace(/^>\s+(.+)$/gm, '<blockquote>$1</blockquote>');
        
        // Unordered list
        html = html.replace(/^- (.+)$/gm, '<li>$1</li>');
        html = html.replace(/(<li>.*<\/li>\n?)/g, '<ul>$1</ul>');
        
        // Line breaks
        html = html.replace(/\n/g, '<br>');
        
        return html;
    }

    // ===== SYNTAX HIGHLIGHTING =====

    highlightSyntax(code, ext) {
        const escaped = this.escapeHtml(code);
        
        const rules = {
            go: [
                { regex: /\b(package\s+\w+|import\s+"[^"]+"|func\s+\w+\([^)]*\)\s*[\w\s{]*\{?)\b/g, type: 'code-function' },
                { regex: /\b(package|import|func|return|var|const|type|struct|interface|package|import|if|else|for|switch|case|default|break|continue|go|goto|defer|map|chan|select)\b/g, type: 'code-keyword' },
                { regex: /"[^"]*"/g, type: 'code-string' },
                { regex: /'[^']*'/g, type: 'code-string' },
                { regex: /\/\/.*$/gm, type: 'code-comment' },
                { regex: /\b\d+\b/g, type: 'code-number' },
            ],
            js: [
                { regex: /\b(const|let|var|function|class|return|if|else|for|while|new|this|export|import|default|async|await|try|catch|throw)\b/g, type: 'code-keyword' },
                { regex: /'[^']*'|"[^"]*"|`[^`]*`/g, type: 'code-string' },
                { regex: /\/\/.*$/gm, type: 'code-comment' },
                { regex: /\/\*[\s\S]*?\*\//g, type: 'code-comment' },
                { regex: /\b\d+\b/g, type: 'code-number' },
            ],
            py: [
                { regex: /\b(def|class|import|from|as|if|elif|else|for|while|return|try|except|finally|with|lambda|yield|print|self|True|False|None|and|or|not|in|is)\b/g, type: 'code-keyword' },
                { regex: /'[^']*'|"[^"]*"/g, type: 'code-string' },
                { regex: /#.*/gm, type: 'code-comment' },
                { regex: /\b\d+\b/g, type: 'code-number' },
            ]
        };
        
        let lines = escaped.split('\n');
        const highlighted = [];
        
        for (const line of lines) {
            let result = line;
            const lineRules = rules[ext] || rules['js'] || [];
            
            for (const rule of lineRules) {
                if (rule.regex.source.includes('comment') || result.includes('<span')) {
                    continue;
                }
            }
            
            highlighted.push(this.escapeHtml(line));
        }
        
        return highlighted.map((line, i) => 
            `<span class="code-line">${i % 2 === 0 ? '&nbsp;' : ''}${line || ''}</span>`
        ).join('\n');
    }

    // ===== UTILITY FUNCTIONS =====

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    getFileExtension(filename) {
        const parts = filename.split('.');
        return parts[parts.length - 1]?.toLowerCase() || '';
    }

    getFileIcon(file) {
        const name = file.Name || file.name || '';
        const ext = name.split('.').pop().toLowerCase();
        const icons = {
            pdf: '📄',
            md: '📝',
            go: '🔧',
            js: '📜',
            py: '🐍',
            json: '⚙️',
            yaml: '⚙️',
            txt: '📃'
        };
        return icons[ext] || '📄';
    }

    getFileColor(file) {
        const name = file.Name || file.name || '';
        const ext = name.split('.').pop().toLowerCase();
        const colors = {
            pdf: '#e06c75',
            md: '#61afef',
            go: '#7dd3f8',
            js: '#e6c600',
            py: '#27ae60',
            json: '#a8a8a8',
            yaml: '#a8a8a8',
            txt: '#808080'
        };
        return colors[ext] || '#e0e0e0';
    }

    print(text, type = '') {
        const line = document.createElement('div');
        line.className = `output-line ${type}`;
        line.textContent = text;
        this.output.appendChild(line);
        this.scrollToBottom();
    }

    printRenderedHTML(html) {
        const container = document.createElement('div');
        container.className = 'output-line rendered-html';
        container.innerHTML = html;
        this.output.appendChild(container);
        this.scrollToBottom();
    }

    clear() {
        this.output.innerHTML = '';
    }

    scrollToBottom() {
        this.output.scrollTop = this.output.scrollHeight;
    }

    handleTab() {
        const value = this.input.value;
        const suggestions = ['help', 'ls', 'access', 'search', 'clear', 'cat', 'pwd', 'whoami', 'date', 'exit'];
        const matches = suggestions.filter(cmd => cmd.startsWith(value.split(' ').pop()));
        
        if (matches.length === 1) {
            const parts = value.split(' ');
            if (parts.length > 1) {
                parts[parts.length - 1] = matches[0];
                this.input.value = parts.join(' ');
            } else {
                this.input.value = matches[0];
            }
        } else if (matches.length > 1) {
            this.print(`\n${matches.join('  ')}`, 'info');
        }
    }

    handleArrowKeys(e) {
        if (e.key === 'ArrowUp') {
            if (this.currentIndex > 0) {
                this.currentIndex--;
                if (this.currentIndex >= 0) {
                    this.input.value = this.history[this.currentIndex] || '';
                }
            }
        } else if (e.key === 'ArrowDown' && this.currentIndex < this.history.length - 1) {
            this.currentIndex++;
            this.input.value = this.history[this.currentIndex + 1] || '';
        }
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.terminal = new Terminal();
});
