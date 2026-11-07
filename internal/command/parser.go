package command

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// CommandType represents the type of command
type CommandType int

const (
	// CmdUnknown represents an unknown command
	CmdUnknown CommandType = iota
	// CmdHelp represents the help command
	CmdHelp
	// CmdLs represents the ls command
	CmdLs
	// CmdAccess represents the access command
	CmdAccess
	// CmdSearch represents the search command
	CmdSearch
	// CmdClear represents the clear command
	CmdClear
	// CmdJump represents the jump command
	CmdJump
	// CmdCat represents the cat command
	CmdCat
	// CmdPwd represents the pwd command
	CmdPwd
	// CmdWhoami represents the whoami command
	CmdWhoami
	// CmdDate represents the date command
	CmdDate
	// CmdExit represents the exit command
	CmdExit
)

// Command represents a parsed command
type Command struct {
	Type CommandType
	Args []string
	Raw  string
}

// Result represents a command result
type Result struct {
	Success bool
	Message string
	Err     error
	Data    interface{}
}

// HelpText represents a help entry
type HelpText struct {
	Command     string
	Description string
	Usage       string
}

// Parser handles command parsing
type Parser struct {
	mu       sync.RWMutex
	commands map[string]CommandType
	handlers map[CommandType]func(args []string) *Result
}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	p := &Parser{
		commands: make(map[string]CommandType),
		handlers: make(map[CommandType]func(args []string) *Result),
	}
	p.registerBuiltin()
	return p
}

// registerBuiltin registers the built-in command mappings
func (p *Parser) registerBuiltin() {
	p.commands["help"] = CmdHelp
	p.commands["ls"] = CmdLs
	p.commands["access"] = CmdAccess
	p.commands["search"] = CmdSearch
	p.commands["clear"] = CmdClear
	p.commands["jump"] = CmdJump
	p.commands["cat"] = CmdCat
	p.commands["pwd"] = CmdPwd
	p.commands["whoami"] = CmdWhoami
	p.commands["date"] = CmdDate
	p.commands["exit"] = CmdExit

	// Register default handlers
	p.handlers[CmdHelp] = p.handleHelp
	p.handlers[CmdLs] = p.handleLS
	p.handlers[CmdAccess] = p.handleAccess
	p.handlers[CmdSearch] = p.handleSearch
	p.handlers[CmdClear] = p.handleClear
	p.handlers[CmdJump] = p.handleJump
	p.handlers[CmdCat] = p.handleCat
	p.handlers[CmdPwd] = p.handlePwd
	p.handlers[CmdWhoami] = p.handleWhoami
	p.handlers[CmdDate] = p.handleDate
	p.handlers[CmdExit] = p.handleExit
}

// Parse parses a command string and returns a Command
func (p *Parser) Parse(raw string) *Command {
	if raw == "" {
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	// Split the command into parts
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return nil
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	if cmd == "" {
		return nil
	}

	cmdType, exists := p.commands[cmd]
	if !exists {
		cmdType = CmdUnknown
	}

	return &Command{
		Type: cmdType,
		Args: args,
		Raw:  raw,
	}
}

// Execute executes a command and returns the result
func (p *Parser) Execute(raw string) *Result {
	cmd := p.Parse(raw)
	if cmd == nil {
		return &Result{
			Success: false,
			Message: "Empty command",
		}
	}

	// Look up the handler
	handler, exists := p.handlers[cmd.Type]
	if !exists {
		handler = p.handleDefault
	}

	return handler(cmd.Args)
}

// RegisterHandler registers a handler for a command type
func (p *Parser) RegisterHandler(cmdType CommandType, handler func(args []string) *Result) {
	p.handlers[cmdType] = handler
}

// GetHelpText returns the help text for a command
func (p *Parser) GetHelpText(cmdType CommandType) *HelpText {
	switch cmdType {
	case CmdHelp:
		return &HelpText{
			Command:     "help",
			Description: "Show help message",
			Usage:       "help",
		}
	case CmdLs:
		return &HelpText{
			Command:     "ls",
			Description: "List files by category",
			Usage:       "ls [category]",
		}
	case CmdAccess:
		return &HelpText{
			Command:     "access",
			Description: "Open a file",
			Usage:       "access <filename>",
		}
	case CmdSearch:
		return &HelpText{
			Command:     "search",
			Description: "Search files",
			Usage:       "search <query>",
		}
	case CmdClear:
		return &HelpText{
			Command:     "clear",
			Description: "Clear the terminal",
			Usage:       "clear",
		}
	case CmdJump:
		return &HelpText{
			Command:     "jump",
			Description: "Jump to a category",
			Usage:       "jump <category>",
		}
	case CmdCat:
		return &HelpText{
			Command:     "cat",
			Description: "Display file contents",
			Usage:       "cat <filename>",
		}
	case CmdPwd:
		return &HelpText{
			Command:     "pwd",
			Description: "Print working directory",
			Usage:       "pwd",
		}
	case CmdWhoami:
		return &HelpText{
			Command:     "whoami",
			Description: "Display current user",
			Usage:       "whoami",
		}
	case CmdDate:
		return &HelpText{
			Command:     "date",
			Description: "Display current date/time",
			Usage:       "date",
		}
	case CmdExit:
		return &HelpText{
			Command:     "exit",
			Description: "Exit the terminal",
			Usage:       "exit",
		}
	default:
		return &HelpText{
			Command:     "unknown",
			Description: "Unknown command",
			Usage:       "unknown",
		}
	}
}

// GetHelpTexts returns all help texts
func (p *Parser) GetHelpTexts() []HelpText {
	return []HelpText{
		{Command: "help", Description: "Show help message", Usage: "help"},
		{Command: "ls [category]", Description: "List files by category", Usage: "ls [category]"},
		{Command: "access <filename>", Description: "Open a file", Usage: "access <filename>"},
		{Command: "search <query>", Description: "Search files", Usage: "search <query>"},
		{Command: "clear", Description: "Clear the terminal", Usage: "clear"},
		{Command: "jump <category>", Description: "Jump to a category", Usage: "jump <category>"},
		{Command: "cat <filename>", Description: "Display file contents", Usage: "cat <filename>"},
		{Command: "pwd", Description: "Print working directory", Usage: "pwd"},
		{Command: "whoami", Description: "Display current user", Usage: "whoami"},
		{Command: "date", Description: "Display current date/time", Usage: "date"},
		{Command: "exit", Description: "Exit the terminal", Usage: "exit"},
	}
}

// Command handlers
func (p *Parser) handleHelp(args []string) *Result {
	texts := p.GetHelpTexts()
	var sb strings.Builder
	sb.WriteString("Available Commands:\n")
	for _, t := range texts {
		sb.WriteString(fmt.Sprintf("  %-20s %s\n", t.Command, t.Description))
	}
	return &Result{
		Success: true,
		Message: sb.String(),
	}
}

func (p *Parser) handleLS(args []string) *Result {
	category := "all"
	if len(args) > 0 {
		category = args[0]
	}
	return &Result{
		Success: true,
		Message: fmt.Sprintf("Listing files in category: %s", category),
		Data:    map[string]interface{}{"category": category},
	}
}

func (p *Parser) handleAccess(args []string) *Result {
	if len(args) == 0 {
		return &Result{
			Success: false,
			Message: "Usage: access <filename>",
		}
	}
	filename := args[0]
	return &Result{
		Success: true,
		Message: fmt.Sprintf("Opening file: %s", filename),
		Data:    map[string]interface{}{"filename": filename},
	}
}

func (p *Parser) handleSearch(args []string) *Result {
	if len(args) == 0 {
		return &Result{
			Success: false,
			Message: "Usage: search <query>",
		}
	}
	query := strings.Join(args, " ")
	return &Result{
		Success: true,
		Message: fmt.Sprintf("Searching for: %s", query),
		Data:    map[string]interface{}{"query": query},
	}
}

func (p *Parser) handleClear(args []string) *Result {
	return &Result{
		Success: true,
		Message: "__CLEAR__",
	}
}

func (p *Parser) handleJump(args []string) *Result {
	if len(args) == 0 {
		return &Result{
			Success: false,
			Message: "Usage: jump <category>",
		}
	}
	category := args[0]
	return &Result{
		Success: true,
		Message: fmt.Sprintf("Jumping to category: %s", category),
		Data:    map[string]interface{}{"category": category},
	}
}

func (p *Parser) handleCat(args []string) *Result {
	if len(args) == 0 {
		return &Result{
			Success: false,
			Message: "Usage: cat <filename>",
		}
	}
	filename := args[0]
	return &Result{
		Success: true,
		Message: fmt.Sprintf("Displaying file: %s", filename),
		Data:    map[string]interface{}{"filename": filename},
	}
}

func (p *Parser) handlePwd(args []string) *Result {
	return &Result{
		Success: true,
		Message: "/home/arch",
	}
}

func (p *Parser) handleWhoami(args []string) *Result {
	return &Result{
		Success: true,
		Message: "arch",
	}
}

func (p *Parser) handleDate(args []string) *Result {
	return &Result{
		Success: true,
		Message: time.Now().Format("2006-01-02 15:04:05"),
	}
}

func (p *Parser) handleExit(args []string) *Result {
	return &Result{
		Success: true,
		Message: "Exiting...",
	}
}

func (p *Parser) handleDefault(args []string) *Result {
	return &Result{
		Success: false,
		Message: "Unknown command: " + args[0],
	}
}
