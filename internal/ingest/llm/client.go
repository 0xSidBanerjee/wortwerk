package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/0xSidBanerjee/wortwerk/internal/ingest/compile"
)

// Client calls an OpenAI-compatible chat API.
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

// Usage reports token counts from one request.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// NewFromEnv builds a client from WORTWERK_LLM_* variables.
func NewFromEnv() (*Client, error) {
	base := strings.TrimRight(os.Getenv("WORTWERK_LLM_BASE_URL"), "/")
	key := os.Getenv("WORTWERK_LLM_API_KEY")
	model := os.Getenv("WORTWERK_LLM_MODEL")
	if base == "" || key == "" || model == "" {
		return nil, fmt.Errorf("set WORTWERK_LLM_BASE_URL, WORTWERK_LLM_API_KEY, WORTWERK_LLM_MODEL")
	}
	return &Client{
		BaseURL: base,
		APIKey:  key,
		Model:   model,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}, nil
}

// Annotate sends a chunk payload and decodes compile.Annotation from the model JSON response.
func (c *Client) Annotate(ctx context.Context, userJSON string) (compile.Annotation, Usage, error) {
	var empty compile.Annotation
	body := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt()},
			{"role": "user", "content": userJSON},
		},
		"temperature": 0.2,
		"response_format": map[string]string{"type": "json_object"},
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return empty, Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	res, err := c.HTTP.Do(req)
	if err != nil {
		return empty, Usage{}, err
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return empty, Usage{}, err
	}
	if res.StatusCode >= 300 {
		return empty, Usage{}, fmt.Errorf("LLM HTTP %d: %s", res.StatusCode, truncate(string(respBody), 400))
	}

	var envelope struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return empty, Usage{}, err
	}
	if len(envelope.Choices) == 0 {
		return empty, Usage{}, fmt.Errorf("empty LLM response")
	}
	content := strings.TrimSpace(envelope.Choices[0].Message.Content)
	var ann compile.Annotation
	if err := json.Unmarshal([]byte(content), &ann); err != nil {
		return empty, Usage{}, fmt.Errorf("decode annotation JSON: %w", err)
	}
	usage := Usage{
		InputTokens:  envelope.Usage.PromptTokens,
		OutputTokens: envelope.Usage.CompletionTokens,
	}
	return ann, usage, nil
}

func systemPrompt() string {
	return `You annotate German learner stories. You receive sentences with token ids and dictionary candidates.
Output JSON only with keys "sentences" and "lexicon".
Never invent token ids. Every word token must appear in exactly one unit.
Rules: fixed phrases one unit; separable verbs one unit; compounds one unit with parts; articles separate; names kind "name".
Keep meanings under 60 characters. Provide English sentence translations.`
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
