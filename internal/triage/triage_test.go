package triage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/behaviorengineering/gitboard/internal/llm"
)

func TestParseAnswer(t *testing.T) {
	raw := "SUMMARY: first line\ncontinues here\nROOT_CAUSE: because\nof that\nFIX_STEPS:\n1. do one\n2) do two\nCONFIDENCE: medium\n"
	got := parseAnswer(raw, "m1")
	if got.Model != "m1" {
		t.Fatalf("model: %q", got.Model)
	}
	if got.Summary != "first line continues here" {
		t.Fatalf("summary: %q", got.Summary)
	}
	if got.RootCause != "because of that" {
		t.Fatalf("root: %q", got.RootCause)
	}
	if len(got.FixSteps) != 2 || got.FixSteps[0] != "do one" || got.FixSteps[1] != "do two" {
		t.Fatalf("steps: %+v", got.FixSteps)
	}
	if got.Confidence != "medium" {
		t.Fatalf("confidence: %q", got.Confidence)
	}
	if got.RawAnswer != raw {
		t.Fatal("raw answer should be preserved")
	}
}

func TestEnabled(t *testing.T) {
	var nilA *Analyzer
	if nilA.Enabled() {
		t.Fatal("nil should be disabled")
	}
	if (&Analyzer{}).Enabled() {
		t.Fatal("empty llm disabled")
	}
	if (&Analyzer{LLM: &llm.Client{}}).Enabled() {
		t.Fatal("empty base url disabled")
	}
	if !(&Analyzer{LLM: &llm.Client{BaseURL: "http://localhost"}}).Enabled() {
		t.Fatal("configured should be enabled")
	}
}

func TestAnalyzeEmptyLog(t *testing.T) {
	a := &Analyzer{LLM: &llm.Client{BaseURL: "http://example.invalid"}}
	_, err := a.Analyze(context.Background(), Request{Log: "  "})
	if err == nil {
		t.Fatal("want empty log error")
	}
}

func TestAnalyzeDisabled(t *testing.T) {
	a := &Analyzer{}
	resp, err := a.Analyze(context.Background(), Request{Log: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Unavailable == "" {
		t.Fatalf("want unavailable: %+v", resp)
	}
}

func TestAnalyzeHappy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "srv-model",
			"choices": []map[string]any{
				{"message": map[string]string{
					"content": "SUMMARY: boom\nROOT_CAUSE: flake\nFIX_STEPS:\n1. retry\nCONFIDENCE: low\n",
				}},
			},
		})
	}))
	t.Cleanup(srv.Close)

	a := &Analyzer{LLM: &llm.Client{BaseURL: srv.URL, Model: "fallback", HTTP: srv.Client()}}
	resp, err := a.Analyze(context.Background(), Request{Log: "FAIL", JobName: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Summary != "boom" || resp.RootCause != "flake" || resp.Confidence != "low" {
		t.Fatalf("resp: %+v", resp)
	}
	if len(resp.FixSteps) != 1 || resp.FixSteps[0] != "retry" {
		t.Fatalf("steps: %+v", resp.FixSteps)
	}
	if resp.Model != "srv-model" {
		t.Fatalf("model: %q", resp.Model)
	}
}
