package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
	"github.com/pricetra/api/graph/gmodel"
)

func (s Service) GetAiTemplate(ctx context.Context, template_type model.AiPromptType) (template model.AiPromptTemplate, err error) {
	qb := table.AiPromptTemplate.
		SELECT(table.AiPromptTemplate.AllColumns).
		FROM(table.AiPromptTemplate).
		WHERE(table.AiPromptTemplate.Type.EQ(
			postgres.NewEnumValue(string(template_type)),
		)).
		LIMIT(1)
	err = qb.QueryContext(ctx, s.DbOrTxQueryable(), &template)
	return template, err
}

func (s Service) CreateAiResponseEntry(
	ctx context.Context,
	user gmodel.User,
	req ChatCompletionRequest,
	res ChatCompletionResponse,
	prompt_type model.AiPromptType,
) (entry model.AiPromptResponse, err error) {
	req_json, err := json.Marshal(req)
	if err != nil {
		return model.AiPromptResponse{}, fmt.Errorf("could not marshal req data: %s", err.Error())
	}
	res_json, err := json.Marshal(res)
	if err != nil {
		return model.AiPromptResponse{}, fmt.Errorf("could not marshal res data: %s", err.Error())
	}

	request := string(req_json)
	response := string(res_json)
	qb := table.AiPromptResponse.INSERT(
		table.AiPromptResponse.Type,
		table.AiPromptResponse.Request,
		table.AiPromptResponse.Response,
		table.AiPromptResponse.UserID,
	).MODEL(model.AiPromptResponse{
		Type: prompt_type,
		Request: &request,
		Response: &response,
		UserID: user.ID,
	}).RETURNING(table.AiPromptResponse.AllColumns)
	if err = qb.QueryContext(ctx, s.DB, &entry); err != nil {
		return model.AiPromptResponse{}, err
	}
	return entry, nil
}

func (s Service) GoogleVisionOcrData(ctx context.Context, image_url string) (ocr_data string, err error) {
	res, err := s.GoogleVisionApiClient.AnnotateImage(ctx, &visionpb.AnnotateImageRequest{
		Image: &visionpb.Image{
			Source: &visionpb.ImageSource{
				ImageUri: image_url,
			},
		},
		Features: []*visionpb.Feature{
			{
				Type: visionpb.Feature_TEXT_DETECTION,
			},
			{
				Type: visionpb.Feature_LOGO_DETECTION,
			},
		},
	})
	if err != nil {
		return "", err
	}
	if res == nil {
		return "", fmt.Errorf("response was empty")
	}
	if res.Error != nil {
		return "", err
	}
	if res.FullTextAnnotation == nil {
		return "", fmt.Errorf("full text annotation was null")
	}

	ocr_data = strings.TrimSpace(res.FullTextAnnotation.Text)
	ocr_data = strings.ReplaceAll(ocr_data, "\n", " ")
	if len(ocr_data) == 0 {
		return "", fmt.Errorf("data was empty")
	}

	return ocr_data, nil
}

const OPENAI_API_BASE = "https://api.openai.com/v1"
const OPENAI_MODEL = "gpt-4.1-mini"

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

func (s Service) GptResponse(ctx context.Context, prompt string, max_tokens int32) (payload ChatCompletionRequest, res ChatCompletionResponse, err error) {
	payload = ChatCompletionRequest{
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
		return ChatCompletionRequest{}, ChatCompletionResponse{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", OPENAI_API_BASE), bytes.NewBuffer(body))
	if err != nil {
		return ChatCompletionRequest{}, ChatCompletionResponse{}, fmt.Errorf("could not create new request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.Tokens.OpenAiApiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ChatCompletionRequest{}, ChatCompletionResponse{}, fmt.Errorf("gpt request resulted in an error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		dump, _ := io.ReadAll(resp.Body)
		return ChatCompletionRequest{}, ChatCompletionResponse{}, fmt.Errorf("non-200 response from gpt: %d - %s", resp.StatusCode, string(dump))
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return ChatCompletionRequest{}, ChatCompletionResponse{}, fmt.Errorf("failed to parse gpt response: %w", err)
	}
	return payload, res, nil
}

func ParseRawGptResponse[T any](gpt_res ChatCompletionResponse) (res T, err error) {
	if len(gpt_res.Choices) != 1 {
		return res, fmt.Errorf("unexpected choice response. expected 1 got %d", len(gpt_res.Choices))
	}

	response_content := gpt_res.Choices[0].Message.Content
	response_content = strings.ReplaceAll(response_content, "`", "")
	response_content = strings.Replace(response_content, "json", "", 1)
	if err := json.Unmarshal([]byte(response_content), &res); err != nil {
		return res, fmt.Errorf("could not parse choice response. %w", err)
	}
	return res, nil
}
