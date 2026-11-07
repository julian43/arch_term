package markdown

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// Renderer converts markdown to HTML
type Renderer struct{}

// NewRenderer creates a new markdown renderer
func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render converts markdown content to HTML
func (r *Renderer) Render(content string) string {
	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	// Split into lines for block-level processing
	lines := strings.Split(content, "\n")
	var result bytes.Buffer

	inCodeBlock := false
	inList := false
	inBlockquote := false
	codeBlockLang := ""
	listItems := 0

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Handle code blocks (``` syntax)
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if inCodeBlock {
				result.WriteString("</code></pre>\n")
				inCodeBlock = false
				codeBlockLang = ""
			} else {
				if inList {
					result.WriteString("</ul>\n")
					inList = false
					listItems = 0
				}
				if inBlockquote {
					result.WriteString("</blockquote>\n")
					inBlockquote = false
				}
				codeBlockLang = strings.TrimRight(strings.TrimLeft(strings.TrimSpace(line), "`"), "`")
				result.WriteString(fmt.Sprintf("<pre><code class=\"language-%s\">", codeBlockLang))
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			result.WriteString(escapeHTML(line))
			result.WriteString("\n")
			continue
		}

		// Close list if we're no longer in one
		if inList && !strings.HasPrefix(strings.TrimSpace(line), "- ") && !strings.HasPrefix(strings.TrimSpace(line), "* ") && strings.TrimSpace(line) != "" {
			result.WriteString("</ul>\n")
			inList = false
			listItems = 0
		}

		// Handle blockquotes
		if strings.HasPrefix(strings.TrimSpace(line), "> ") {
			if !inBlockquote {
				result.WriteString("<blockquote>\n")
				inBlockquote = true
			}
			result.WriteString("<p>" + r.inlineFormat(strings.TrimSpace(line)[2:]) + "</p>\n")
			continue
		} else if inBlockquote {
			result.WriteString("</blockquote>\n")
			inBlockquote = false
		}

		// Handle empty lines
		if strings.TrimSpace(line) == "" {
			if inList {
				result.WriteString("</ul>\n")
				inList = false
				listItems = 0
			}
			result.WriteString("\n")
			continue
		}

		// Handle headers
		if m := matchHeader(line); m != nil && m[1] != "" {
			if inList {
				result.WriteString("</ul>\n")
				inList = false
				listItems = 0
			}
			level := len(m[1])
			text := strings.TrimSpace(m[2])
			id := strings.ToLower(strings.ReplaceAll(text, " ", "-"))
			id = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(id, "")
			result.WriteString(fmt.Sprintf("<h%d id=\"%s\">%s</h%d>\n", level, id, r.inlineFormat(text), level))
			continue
		}

		// Handle horizontal rules
		if matched := matchHR(line); matched {
			if inList {
				result.WriteString("</ul>\n")
				inList = false
				listItems = 0
			}
			result.WriteString("<hr>\n")
			continue
		}

		// Handle list items
		if strings.HasPrefix(strings.TrimSpace(line), "- ") || strings.HasPrefix(strings.TrimSpace(line), "* ") {
			if !inList {
				result.WriteString("<ul>\n")
				inList = true
				listItems = 0
			}
			listItems++
			text := strings.TrimPrefix(strings.TrimSpace(line), "- ")
			text = strings.TrimPrefix(text, "* ")
			result.WriteString(fmt.Sprintf("<li>%s</li>\n", r.inlineFormat(text)))
			continue
		}

		// Handle paragraphs
		if inList {
			result.WriteString("</ul>\n")
			inList = false
			listItems = 0
		}
		result.WriteString(fmt.Sprintf("<p>%s</p>\n", r.inlineFormat(line)))
	}

	// Close any open blocks
	if inCodeBlock {
		result.WriteString("</code></pre>\n")
	}
	if inList {
		result.WriteString("</ul>\n")
	}
	if inBlockquote {
		result.WriteString("</blockquote>\n")
	}

	return result.String()
}

// escapeHTML escapes HTML special characters
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&")
	s = strings.ReplaceAll(s, "<", "<")
	s = strings.ReplaceAll(s, ">", ">")
	return s
}

// inlineFormat processes inline markdown elements
func (r *Renderer) inlineFormat(text string) string {
	// Escape HTML first
	text = escapeHTML(text)

	// Inline code: `code`
	codeRe := regexp.MustCompile("`([^`]+)`")
	text = codeRe.ReplaceAllString(text, "<code>$1</code>")

	// Bold: **text** or __text__
	boldRe := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	text = boldRe.ReplaceAllString(text, "<strong>$1</strong>")
	boldRe2 := regexp.MustCompile(`__([^_]+)__`)
	text = boldRe2.ReplaceAllString(text, "<strong>$1</strong>")

	// Italic: *text* or _text_
	italicRe := regexp.MustCompile(`\*([^*]+)\*`)
	text = italicRe.ReplaceAllString(text, "<em>$1</em>")
	italicRe2 := regexp.MustCompile(`_([^_]+)_`)
	text = italicRe2.ReplaceAllString(text, "<em>$1</em>")

	// Strikethrough: ~~text~~
	strikeRe := regexp.MustCompile(`~~([^~]+)~~`)
	text = strikeRe.ReplaceAllString(text, "<del>$1</del>")

	// Links: [text](url)
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	text = linkRe.ReplaceAllString(text, `<a href="$2">$1</a>`)

	// Images: ![alt](url)
	imgRe := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	text = imgRe.ReplaceAllString(text, `<img src="$2" alt="$1">`)

	return text
}

// matchHeader matches markdown headers (#, ##, ###, etc.)
var headerRe = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

func matchHeader(line string) []string {
	return headerRe.FindStringSubmatch(line)
}

// matchHR matches horizontal rules (---, ***, ___)
var hrRe = regexp.MustCompile(`^(\*{3,}|-{3,}|_{3,})$`)

func matchHR(line string) bool {
	return hrRe.MatchString(strings.TrimSpace(line))
}
