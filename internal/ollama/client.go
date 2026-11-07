package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Model represents an Ollama model
type Model struct {
	Name       string    `json:"name"`
	Parameters string    `json:"parameters,omitempty"`
	ModifiedAt time.Time `json:"modified_at"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	Model   string      `json:"model"`
	Message Message     `json:"message"`
	Done    bool        `json:"done"`
	Meta    *MetaFields `json:"meta,omitempty"`
}

// MetaFields holds metadata about the response
type MetaFields struct {
	TotalDuration   uint64 `json:"total_duration,omitempty"`
	LoadDuration    uint64 `json:"load_duration,omitempty"`
	PromptEvalCount uint   `json:"prompt_eval_count,omitempty"`
}

// OllamaClient provides an interface to Ollama API
type OllamaClient struct {
	mu          sync.RWMutex
	baseURL     string
	model       string
	httpClient  *http.Client
	initialized bool
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		initialized: false,
	}
}

// NewOllamaClientDefault creates a new Ollama client with default values
func NewOllamaClientDefault() *OllamaClient {
	return &OllamaClient{
		baseURL: "http://localhost:11434",
		model:   "llama3.2",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		initialized: false,
	}
}

// Initialize checks if the Ollama server is available
func (c *OllamaClient) Initialize() error {
	url := fmt.Sprintf("%s/api/tags", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama server returned status %d", resp.StatusCode)
	}

	c.initialized = true
	return nil
}

// IsAvailable checks if the Ollama server is available
func (c *OllamaClient) IsAvailable() bool {
	url := fmt.Sprintf("%s/api/tags", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return resp.StatusCode == http.StatusOK
}

// ListModels returns available models
func (c *OllamaClient) ListModels() ([]Model, error) {
	url := fmt.Sprintf("%s/api/tags", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list models: status %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models, ok := data["models"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	result := make([]Model, 0, len(models))
	for _, m := range models {
		model, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		var name string
		if n, ok := model["name"]; ok {
			name = fmt.Sprintf("%v", n)
		}
		var mod string
		if m, ok := model["modified"]; ok {
			mod = fmt.Sprintf("%v", m)
			_ = mod
		}
		result = append(result, Model{
			Name: name,
		})
	}

	return result, nil
}

// Chat sends a chat request and returns the response
func (c *OllamaClient) Chat(messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", c.baseURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return "", fmt.Errorf("Ollama API error: status %d, body: %s", resp.StatusCode, string(body[:n]))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return chatResp.Message.Content, nil
}

// ChatStream sends a chat request with streaming support
func (c *OllamaClient) ChatStream(messages []Message, callback func(string) error) error {
	reqBody := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", c.baseURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("Ollama API error: status %d, body: %s", resp.StatusCode, string(body[:n]))
	}

	decoder := json.NewDecoder(resp.Body)
	var fullResponse strings.Builder

	for {
		var chunk ChatResponse
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode chunk: %w", err)
		}

		if chunk.Done {
			break
		}

		fullResponse.WriteString(chunk.Message.Content)
		if err := callback(chunk.Message.Content); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}
	}

	return nil
}

// GenerateResponse generates a response for the given prompt
func (c *OllamaClient) GenerateResponse(prompt string) (string, error) {
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}
	return c.Chat(messages)
}

// GenerateStream generates a streaming response for the given prompt
func (c *OllamaClient) GenerateStream(prompt string, callback func(string) error) error {
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}
	return c.ChatStream(messages, callback)
}

// GetModel returns the configured model name
func (c *OllamaClient) GetModel() string {
	return c.model
}

// SetModel sets the model name
func (c *OllamaClient) SetModel(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.model = model
}

// GetBaseURL returns the base URL
func (c *OllamaClient) GetBaseURL() string {
	return c.baseURL
}

// SetBaseURL sets the base URL
func (c *OllamaClient) SetBaseURL(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.baseURL = url
}
