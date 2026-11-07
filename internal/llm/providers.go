package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ProviderType represents the type of LLM provider
type ProviderType int

const (
	// ProviderNone represents no provider
	ProviderNone ProviderType = iota
	// ProviderOllama represents Ollama
	ProviderOllama
	// ProviderOpenAI represents OpenAI compatible API
	ProviderOpenAI
	// ProviderLMStudio represents LM Studio
	ProviderLMStudio
)

// String returns the string representation of ProviderType
func (pt ProviderType) String() string {
	switch pt {
	case ProviderOllama:
		return "ollama"
	case ProviderOpenAI:
		return "openai"
	case ProviderLMStudio:
		return "lmstudio"
	default:
		return "unknown"
	}
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ModelInfo represents an available model
type ModelInfo struct {
	Name     string `json:"name"`
	ID       string `json:"id,omitempty"`
	Provider string `json:"provider"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
	MaxTokens   *int      `json:"max_tokens,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Done    bool    `json:"done"`
	Raw     string  `json:"raw,omitempty"`
}

// LLMProvider defines the interface for LLM backends
type LLMProvider interface {
	// GetName returns the provider name
	GetName() string
	// GetType returns the provider type
	GetType() ProviderType
	// GetBaseURL returns the base URL
	GetBaseURL() string
	// SetBaseURL sets the base URL
	SetBaseURL(url string)
	// GetModel returns the model name
	GetModel() string
	// SetModel sets the model name
	SetModel(model string)
	// IsAvailable checks if the provider is available
	IsAvailable() bool
	// ListModels returns available models
	ListModels() ([]*ModelInfo, error)
	// Chat sends a chat request
	Chat(messages []Message) (string, error)
	// ChatStream sends a streaming chat request
	ChatStream(messages []Message, callback func(string) error) error
}

// LLMClient manages LLM interactions
type LLMClient struct {
	mu           sync.RWMutex
	provider     LLMProvider
	providers    map[string]LLMProvider
	currentModel string
	baseURL      string
}

// NewLLMClient creates a new LLM client
func NewLLMClient() *LLMClient {
	return &LLMClient{
		providers: make(map[string]LLMProvider),
	}
}

// RegisterProvider registers a new LLM provider
func (c *LLMClient) RegisterProvider(name string, p LLMProvider) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.providers[name] = p
	return nil
}

// GetProvider returns a provider by name
func (c *LLMClient) GetProvider(name string) (LLMProvider, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	provider, ok := c.providers[name]
	return provider, ok
}

// SetProvider sets the current provider
func (c *LLMClient) SetProvider(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	provider, ok := c.providers[name]
	if !ok {
		return fmt.Errorf("provider %s not found", name)
	}

	c.provider = provider
	return nil
}

// GetCurrentProvider returns the current provider
func (c *LLMClient) GetCurrentProvider() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.provider == nil {
		return ""
	}
	return c.provider.GetName()
}

// SetModel sets the model for the current provider
func (c *LLMClient) SetModel(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentModel = model
	if c.provider != nil {
		c.provider.SetModel(model)
	}
}

// GetModel returns the current model
func (c *LLMClient) GetModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentModel
}

// Chat sends a chat request to the current provider
func (c *LLMClient) Chat(messages []Message) (string, error) {
	c.mu.RLock()
	provider := c.provider
	c.mu.RUnlock()

	if provider == nil {
		return "", fmt.Errorf("no LLM provider configured")
	}

	return provider.Chat(messages)
}

// ChatStream sends a streaming chat request to the current provider
func (c *LLMClient) ChatStream(messages []Message, callback func(string) error) error {
	c.mu.RLock()
	provider := c.provider
	c.mu.RUnlock()

	if provider == nil {
		return fmt.Errorf("no LLM provider configured")
	}

	return provider.ChatStream(messages, callback)
}

// IsProviderAvailable checks if a provider is available
func (c *LLMClient) IsProviderAvailable(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	provider, ok := c.providers[name]
	if !ok {
		return false
	}

	return provider.IsAvailable()
}

// ListModels lists available models for a provider
func (c *LLMClient) ListModels(providerName string) ([]*ModelInfo, error) {
	c.mu.RLock()
	provider, ok := c.providers[providerName]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	return provider.ListModels()
}

// GetProviders returns all registered providers
func (c *LLMClient) GetProviders() map[string]LLMProvider {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.providers
}

// GetProviderNames returns all registered provider names
func (c *LLMClient) GetProviderNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.providers))
	for name := range c.providers {
		names = append(names, name)
	}
	return names
}

// OllamaProvider implements LLMProvider for Ollama
type OllamaProvider struct {
	mu         sync.RWMutex
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetName returns the provider name
func (p *OllamaProvider) GetName() string {
	return "ollama"
}

// GetType returns the provider type
func (p *OllamaProvider) GetType() ProviderType {
	return ProviderOllama
}

// GetBaseURL returns the base URL
func (p *OllamaProvider) GetBaseURL() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.baseURL
}

// SetBaseURL sets the base URL
func (p *OllamaProvider) SetBaseURL(url string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.baseURL = url
}

// GetModel returns the model name
func (p *OllamaProvider) GetModel() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.model
}

// SetModel sets the model name
func (p *OllamaProvider) SetModel(model string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.model = model
}

// IsAvailable checks if the provider is available
func (p *OllamaProvider) IsAvailable() bool {
	url := fmt.Sprintf("%s/api/tags", p.baseURL)
	resp, err := p.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return resp.StatusCode == http.StatusOK
}

// ListModels returns available models
func (p *OllamaProvider) ListModels() ([]*ModelInfo, error) {
	url := fmt.Sprintf("%s/api/tags", p.baseURL)
	resp, err := p.httpClient.Get(url)
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

	result := make([]*ModelInfo, 0, len(models))
	for _, m := range models {
		model, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		var name string
		if n, ok := model["name"]; ok {
			name = fmt.Sprintf("%v", n)
		}
		result = append(result, &ModelInfo{
			Name:     name,
			Provider: ProviderOllama.String(),
		})
	}

	return result, nil
}

// Chat sends a chat request
func (p *OllamaProvider) Chat(messages []Message) (string, error) {
	reqBody := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
		"stream":   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", p.baseURL)
	resp, err := p.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
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

	var chatResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	message, ok := chatResp["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid message content")
	}

	return content, nil
}

// ChatStream sends a streaming chat request
func (p *OllamaProvider) ChatStream(messages []Message, callback func(string) error) error {
	reqBody := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
		"stream":   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", p.baseURL)
	resp, err := p.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
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
	for {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			break
		}

		done, ok := chunk["done"].(bool)
		if ok && done {
			break
		}

		message, ok := chunk["message"].(map[string]interface{})
		if !ok {
			continue
		}

		content, ok := message["content"].(string)
		if !ok {
			continue
		}

		if err := callback(content); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}
	}

	return nil
}

// OpenAIProvider implements LLMProvider for OpenAI compatible APIs
type OpenAIProvider struct {
	mu         sync.RWMutex
	baseURL    string
	model      string
	apiKey     string
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI compatible provider
func NewOpenAIProvider(baseURL, model, apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		baseURL: baseURL,
		model:   model,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetName returns the provider name
func (p *OpenAIProvider) GetName() string {
	return "openai"
}

// GetType returns the provider type
func (p *OpenAIProvider) GetType() ProviderType {
	return ProviderOpenAI
}

// GetBaseURL returns the base URL
func (p *OpenAIProvider) GetBaseURL() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.baseURL
}

// SetBaseURL sets the base URL
func (p *OpenAIProvider) SetBaseURL(url string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.baseURL = url
}

// GetModel returns the model name
func (p *OpenAIProvider) GetModel() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.model
}

// SetModel sets the model name
func (p *OpenAIProvider) SetModel(model string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.model = model
}

// IsAvailable checks if the provider is available
func (p *OpenAIProvider) IsAvailable() bool {
	url := fmt.Sprintf("%s/v1/models", p.baseURL)
	resp, err := p.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return resp.StatusCode == http.StatusOK
}

// ListModels returns available models
func (p *OpenAIProvider) ListModels() ([]*ModelInfo, error) {
	url := fmt.Sprintf("%s/v1/models", p.baseURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
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

	models, ok := data["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	result := make([]*ModelInfo, 0, len(models))
	for _, m := range models {
		model, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		var id, name string
		if i, ok := model["id"]; ok {
			id = fmt.Sprintf("%v", i)
		}
		if n, ok := model["object"]; ok {
			name = fmt.Sprintf("%v", n)
		}
		result = append(result, &ModelInfo{
			Name:     name,
			ID:       id,
			Provider: ProviderOpenAI.String(),
		})
	}

	return result, nil
}

// Chat sends a OpenAI compatible API request
func (p *OpenAIProvider) Chat(messages []Message) (string, error) {
	reqBody := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
		"stream":   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/chat/completions", p.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return "", fmt.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(body[:n]))
	}

	var chatResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	choices, ok := chatResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid message content")
	}

	return content, nil
}

// ChatStream sends a streaming chat request
func (p *OpenAIProvider) ChatStream(messages []Message, callback func(string) error) error {
	reqBody := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
		"stream":   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/chat/completions", p.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(body[:n]))
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			break
		}

		done, ok := chunk["done"]
		if doneStr, ok := done.(string); ok {
			if strings.ToLower(doneStr) == "true" {
				break
			}
		}
		if doneBool, ok := done.(bool); ok && doneBool {
			break
		}

		usage, ok := chunk["usage"]
		if ok {
			if err := callback(fmt.Sprintf("%v", usage)); err != nil {
				return fmt.Errorf("callback error: %w", err)
			}
		}
	}

	return nil
}

// GetHTTPClient returns the HTTP client
func (p *OpenAIProvider) GetHTTPClient() *http.Client {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.httpClient
}

// SetHTTPClient sets the HTTP client
func (p *OpenAIProvider) SetHTTPClient(client *http.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.httpClient = client
}

// LMStudioProvider implements LLMProvider for LM Studio
type LMStudioProvider struct {
	mu         sync.RWMutex
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewLMStudioProvider creates a new LM Studio provider
func NewLMStudioProvider(baseURL, model string) *LMStudioProvider {
	return &LMStudioProvider{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetName returns the provider name
func (p *LMStudioProvider) GetName() string {
	return "lmstudio"
}

// GetType returns the provider type
func (p *LMStudioProvider) GetType() ProviderType {
	return ProviderLMStudio
}

// GetBaseURL returns the base URL
func (p *LMStudioProvider) GetBaseURL() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.baseURL
}

// SetBaseURL sets the base URL
func (p *LMStudioProvider) SetBaseURL(url string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.baseURL = url
}

// GetModel returns the model name
func (p *LMStudioProvider) GetModel() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.model
}

// SetModel sets the model name
func (p *LMStudioProvider) SetModel(model string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.model = model
}

// IsAvailable checks if the provider is available
func (p *LMStudioProvider) IsAvailable() bool {
	url := fmt.Sprintf("%s/v1/models", p.baseURL)
	resp, err := p.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return resp.StatusCode == http.StatusOK
}

// ListModels returns available models
func (p *LMStudioProvider) ListModels() ([]*ModelInfo, error) {
	url := fmt.Sprintf("%s/v1/models", p.baseURL)
	resp, err := p.httpClient.Get(url)
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

	models, ok := data["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	result := make([]*ModelInfo, 0, len(models))
	for _, m := range models {
		model, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		var id, name string
		if i, ok := model["id"]; ok {
			id = fmt.Sprintf("%v", i)
		}
		if n, ok := model["name"]; ok {
			name = fmt.Sprintf("%v", n)
		}
		result = append(result, &ModelInfo{
			Name:     name,
			ID:       id,
			Provider: ProviderLMStudio.String(),
		})
	}

	return result, nil
}

// Chat sends a chat request
func (p *LMStudioProvider) Chat(messages []Message) (string, error) {
	reqBody := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
		"stream":   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/chat/completions", p.baseURL)
	resp, err := p.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return "", fmt.Errorf("LM Studio API error: status %d, body: %s", resp.StatusCode, string(body[:n]))
	}

	var chatResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	choices, ok := chatResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid message content")
	}

	return content, nil
}

// ChatStream sends a streaming chat request
func (p *LMStudioProvider) ChatStream(messages []Message, callback func(string) error) error {
	reqBody := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
		"stream":   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/chat/completions", p.baseURL)
	resp, err := p.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		return fmt.Errorf("LM Studio API error: status %d, body: %s", resp.StatusCode, string(body[:n]))
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			break
		}

		done, ok := chunk["done"]
		if doneStr, ok := done.(string); ok {
			if strings.ToLower(doneStr) == "true" {
				break
			}
		}
		if doneBool, ok := done.(bool); ok && doneBool {
			break
		}

		usage, ok := chunk["usage"]
		if ok {
			if err := callback(fmt.Sprintf("%v", usage)); err != nil {
				return fmt.Errorf("callback error: %w", err)
			}
		}
	}

	return nil
}

// GetHTTPClient returns the HTTP client
func (p *LMStudioProvider) GetHTTPClient() *http.Client {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.httpClient
}

// SetHTTPClient sets the HTTP client
func (p *LMStudioProvider) SetHTTPClient(client *http.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.httpClient = client
}
