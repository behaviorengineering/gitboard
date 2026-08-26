package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnabled(t *testing.T) {
	var nilC *Client
	if nilC.Enabled() {
		t.Fatal("nil enabled")
	}
	if (&Client{}).Enabled() {
		t.Fatal("empty base enabled")
	}
	if !(&Client{BaseURL: "http://x"}).Enabled() {
		t.Fatal("want enabled")
	}
}

func TestChat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "m",
			"choices": []map[string]any{
				{"message": map[string]string{"content": "hi"}},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c := &Client{BaseURL: srv.URL, Model: "fallback", HTTP: srv.Client()}
	content, model, err := c.Chat(context.Background(), "sys", "user")
	if err != nil {
		t.Fatal(err)
	}
	if content != "hi" || model != "m" {
		t.Fatalf("content=%q model=%q", content, model)
	}
}
