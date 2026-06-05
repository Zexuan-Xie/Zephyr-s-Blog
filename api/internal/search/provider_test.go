package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQwenProviderSendsOpenAICompatibleEmbeddingRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/compatible-mode/v1/embeddings" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["model"] != "text-embedding-v4" || payload["encoding_format"] != "float" || int(payload["dimensions"].(float64)) != 1024 {
			t.Fatalf("payload = %#v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		embedding := make([]float32, 1024)
		embedding[0] = 0.25
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{{"embedding": embedding}}})
	}))
	defer server.Close()

	provider := NewQwenProvider(server.URL+"/compatible-mode/v1", "test-key", "text-embedding-v4", 1024, server.Client())
	embedding, err := provider.Embed(context.Background(), "name\n/path\nkw\ntext")
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}
	if len(embedding) != 1024 || embedding[0] != 0.25 {
		t.Fatalf("embedding len/first = %d/%v", len(embedding), embedding[0])
	}
}
