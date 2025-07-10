package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const OPENAI_API_BASE = "https://api.openai.com/v1"
const OPENAI_MODEL = "gpt-4o"

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatChoice struct {
	Index        int             `json:"index"`
	Message      ChatMessage     `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	N           int           `json:"n,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	PresencePenalty float64   `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	LogitBias   map[string]int `json:"logit_bias,omitempty"`
	User        string        `json:"user,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatChoice           `json:"choices"`
	Usage   *ChatCompletionUsage   `json:"usage,omitempty"`
}

func (s Service) GptResponse(ctx context.Context, prompt string, max_tokens int32) (res ChatCompletionResponse, err error) {
	payload := ChatCompletionRequest{
		Model: OPENAI_MODEL,
		Messages: []ChatMessage{
			{
				Role: "user",
				Content: prompt,
			},
		},
		Temperature: 0.2,
		MaxTokens: int(max_tokens),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return ChatCompletionResponse{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", OPENAI_API_BASE), bytes.NewBuffer(body))
	if err != nil {
		return ChatCompletionResponse{}, fmt.Errorf("could not create new request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Tokens.OpenAiApiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ChatCompletionResponse{}, fmt.Errorf("gpt request resulted in an error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		dump, _ := io.ReadAll(resp.Body)
		return ChatCompletionResponse{}, fmt.Errorf("non-200 response from gpt: %d - %s", resp.StatusCode, string(dump))
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return ChatCompletionResponse{}, fmt.Errorf("failed to parse gpt response: %w", err)
	}
	return res, nil
}
