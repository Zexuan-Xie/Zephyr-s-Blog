package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const ProviderQwen = "qwen"

type EmbeddingProvider interface {
	Embed(context.Context, string) ([]float32, error)
}

type QwenProvider struct {
	baseURL    string
	apiKey     string
	model      string
	dimensions int
	client     *http.Client
}

func NewQwenProvider(baseURL, apiKey, model string, dimensions int, client *http.Client) *QwenProvider {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &QwenProvider{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		model:      model,
		dimensions: dimensions,
		client:     client,
	}
}

func (p *QwenProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	if strings.TrimSpace(p.apiKey) == "" {
		return nil, errors.New("dashscope api key is not configured")
	}
	payload := map[string]any{
		"model":           p.model,
		"input":           input,
		"dimensions":      p.dimensions,
		"encoding_format": "float",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+p.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := p.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("embedding request failed: status=%d body=%s", response.StatusCode, strings.TrimSpace(string(message)))
	}
	var decoded struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	if len(decoded.Data) == 0 || len(decoded.Data[0].Embedding) == 0 {
		return nil, errors.New("embedding response did not include an embedding")
	}
	if len(decoded.Data[0].Embedding) != p.dimensions {
		return nil, fmt.Errorf("embedding dimensions = %d, want %d", len(decoded.Data[0].Embedding), p.dimensions)
	}
	return decoded.Data[0].Embedding, nil
}
