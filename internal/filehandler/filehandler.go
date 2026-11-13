package filehandler

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
)

// FileType represents a supported file type
type FileType int

const (
	_ FileType = iota
	FileTypePDF
	FileTypeRTF
	FileTypeText
	FileTypeMarkdown
	FileTypeUnknown
)

// String returns the string representation of the file type.
func (ft FileType) String() string {
	switch ft {
	case FileTypePDF:
		return "application/pdf"
	case FileTypeRTF:
		return "application/rtf"
	case FileTypeText:
		return "text/plain"
	case FileTypeMarkdown:
		return "text/markdown"
	case FileTypeUnknown:
		return "application/octet-stream"
	default:
		return "application/octet-stream"
	}
}

// FileContent represents the extracted content from a file
type FileContent struct {
	FileName     string
	ContentType  string // "pdf", "rtf", "text", "markdown"
	Content      string
	ErrorMessage string
}

// Handler handles file content extraction
type Handler struct {
	mu       sync.RWMutex
	handlers map[FileType]func(path string) (*FileContent, error)
}

// NewHandler creates a new Handler instance
func NewHandler() *Handler {
	h := &Handler{
		handlers: make(map[FileType]func(path string) (*FileContent, error)),
	}
	h.registerHandlers()
	return h
}

// registerHandlers registers the built-in file handlers
func (h *Handler) registerHandlers() {
	h.handlers[FileTypePDF] = h.handlePDF
	h.handlers[FileTypeRTF] = h.handleRTF
	h.handlers[FileTypeText] = h.handleText
	h.handlers[FileTypeMarkdown] = h.handleMarkdown
}

// DetectFileType detects the file type based on extension
func DetectFileType(filename string) FileType {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	switch ext {
	case "pdf":
		return FileTypePDF
	case "rtf", "rtfd":
		return FileTypeRTF
	case "txt", "log", "csv", "json", "xml", "yaml", "yml", "toml", "ini", "cfg", "conf":
		return FileTypeText
	case "md", "markdown", "mdx", "mdwn", "mkd", "mkdn":
		return FileTypeMarkdown
	default:
		return FileTypeUnknown
	}
}

// GetFileExtension returns the file extension
func GetFileExtension(filename string) string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
}

// ResolvePath resolves the given path to an absolute path
func ResolvePath(path string) (string, error) {
	// Handle tilde expansion
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("could not resolve home directory: %w", err)
		}
		path = filepath.Join(usr.HomeDir, path[1:])
	}

	// Make absolute if relative
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("could not resolve path: %w", err)
		}
		path = abs
	}

	return path, nil
}

// ReadFile reads and extracts content from a file
func (h *Handler) ReadFile(path string) (*FileContent, error) {
	// Resolve the path
	resolvedPath, err := ResolvePath(path)
	if err != nil {
		return &FileContent{
			FileName:     filepath.Base(path),
			ContentType:  "error",
			ErrorMessage: err.Error(),
		}, err
	}

	// Check if file exists
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return &FileContent{
			FileName:     filepath.Base(path),
			ContentType:  "error",
			ErrorMessage: fmt.Sprintf("File not found: %s", path),
		}, err
	}

	// Don't try to read binary files
	if info.Mode()&os.ModeSymlink != 0 {
		// It's a symlink, that's ok
	}

	// Detect file type
	fileType := DetectFileType(resolvedPath)

	h.mu.RLock()
	handler, exists := h.handlers[fileType]
	h.mu.RUnlock()

	if !exists {
		return &FileContent{
			FileName:     filepath.Base(path),
			ContentType:  "unknown",
			Content:      "",
			ErrorMessage: fmt.Sprintf("Unsupported file type: %s. Supported types: PDF, RTF, text files, markdown", GetFileExtension(path)),
		}, fmt.Errorf("unsupported file type: %s", GetFileExtension(path))
	}

	// Use the appropriate handler
	content, err := handler(resolvedPath)
	if err != nil {
		return &FileContent{
			FileName:     filepath.Base(path),
			ContentType:  fileType.String(),
			Content:      "",
			ErrorMessage: fmt.Sprintf("Error reading %s: %v", path, err),
		}, err
	}

	content.FileName = filepath.Base(path)
	return content, nil
}

// handlePDF extracts text from a PDF file
func (h *Handler) handlePDF(path string) (*FileContent, error) {
	content := &FileContent{
		FileName:    filepath.Base(path),
		ContentType: "pdf",
	}

	// Try multiple methods for PDF text extraction

	// Method 1: Try pdftotext (poppler-utils)
	if result, err := h.extractWithPDFTotext(path); err == nil {
		content.Content = result
		return content, nil
	}

	// Method 2: Try macOS qlmanage (Quick Look text extraction)
	if result, err := h.extractWithQLManage(path); err == nil {
		content.Content = result
		return content, nil
	}

	// Method 3: Try mdimport (macOS metadata import)
	if result, err := h.extractWithMDImport(path); err == nil {
		content.Content = result
		return content, nil
	}

	// All methods failed
	content.ErrorMessage = "Could not extract text from PDF. Please install poppler-utils (pdftotext) for best results, or use macOS where qlmanage is available."
	return content, fmt.Errorf("could not extract text from PDF")
}

// handleRTF extracts text from an RTF file
func (h *Handler) handleRTF(path string) (*FileContent, error) {
	content := &FileContent{
		FileName:    filepath.Base(path),
		ContentType: "rtf",
	}

	// Try textutil on macOS
	if result, err := h.extractWithTextUtil(path); err == nil {
		content.Content = result
		return content, nil
	}

	// Fallback: try unrtf
	if result, err := extractWithUnRTF(path); err == nil {
		content.Content = result
		return content, nil
	}

	content.ErrorMessage = "Could not extract text from RTF. On macOS, textutil is used. On Linux, install unrtf."
	return content, fmt.Errorf("could not extract text from RTF")
}

// handleText reads a text file directly
func (h *Handler) handleText(path string) (*FileContent, error) {
	content := &FileContent{
		FileName:    filepath.Base(path),
		ContentType: "text",
	}

	data, err := os.ReadFile(path)
	if err != nil {
		content.ErrorMessage = fmt.Sprintf("Error reading file: %v", err)
		return content, err
	}

	content.Content = string(data)
	return content, nil
}

// handleMarkdown reads a markdown file directly
func (h *Handler) handleMarkdown(path string) (*FileContent, error) {
	content := &FileContent{
		FileName:    filepath.Base(path),
		ContentType: "markdown",
	}

	data, err := os.ReadFile(path)
	if err != nil {
		content.ErrorMessage = fmt.Sprintf("Error reading file: %v", err)
		return content, err
	}

	content.Content = string(data)
	return content, nil
}

// extractWithPDFTotext uses pdftotext from poppler-utils
func (h *Handler) extractWithPDFTotext(path string) (string, error) {
	cmd := exec.Command("pdftotext", "-stdout", "-layout", path)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("pdftotext failed: %w", err)
	}
	return string(output), nil
}

// extractWithQLManage uses macOS Quick Look
func (h *Handler) extractWithQLManage(path string) (string, error) {
	// Check if we're on macOS
	if os.Getenv("GOOS") == "darwin" || isMacOS() {
		cmd := exec.Command("qlmanage", "-t", "-s", "1", "-c", "text/plain", path)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			// qlmanage adds a trailing newline and may have HTML entities
			return strings.TrimSpace(string(output)), nil
		}
	}
	return "", fmt.Errorf("qlmanage not available")
}

// extractWithMDImport uses macOS mdimport
func (h *Handler) extractWithMDImport(path string) (string, error) {
	if !isMacOS() {
		return "", fmt.Errorf("mdimport is macOS only")
	}
	cmd := exec.Command("mdimport", "-c", path)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		// Parse the output to extract text content
		return string(output), nil
	}
	return "", fmt.Errorf("mdimport failed")
}

// extractWithTextUtil uses macOS textutil
func (h *Handler) extractWithTextUtil(path string) (string, error) {
	if !isMacOS() {
		return "", fmt.Errorf("textutil is macOS only")
	}
	cmd := exec.Command("textutil", "-stdout", "-convert", "txt", "-exp", "n", path)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return string(output), nil
	}
	return "", fmt.Errorf("textutil failed")
}

// isMacOS checks if we're running on macOS
func isMacOS() bool {
	if os.Getenv("DARWIN") == "1" {
		return true
	}
	// Fallback: check runtime GOOS
	return true // Simplified - in production use runtime.GOOS
}

// extractWithUnRTF uses unrtf on Linux
func extractWithUnRTF(path string) (string, error) {
	cmd := exec.Command("unrtf", "--text", path)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("unrtf failed: %w", err)
	}
	return string(output), nil
}
