package languagetool_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/languagetool-cli/languagetool"
)

func TestDefaultConfig(t *testing.T) {
	cfg := languagetool.DefaultConfig()
	if cfg.Rate != 500*time.Millisecond {
		t.Errorf("Rate = %v, want 500ms", cfg.Rate)
	}
	if cfg.Retries <= 0 {
		t.Errorf("Retries = %d, want > 0", cfg.Retries)
	}
	if cfg.Timeout <= 0 {
		t.Errorf("Timeout = %v, want > 0", cfg.Timeout)
	}
	if cfg.UserAgent == "" {
		t.Error("UserAgent is empty")
	}
}

func TestNewClientNotNil(t *testing.T) {
	c := languagetool.NewClient(languagetool.DefaultConfig())
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestGrammarIssueRoundTrip(t *testing.T) {
	want := languagetool.GrammarIssue{
		Sentence:     "This are wrong.",
		Offset:       0,
		Length:       7,
		Message:      "Grammar error",
		ShortMessage: "Grammatical error",
		Replacements: []string{"These"},
		RuleID:       "THIS_NNS",
		Category:     "Grammar",
		IssueType:    "non-conformance",
		Context:      "This",
	}
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got languagetool.GrammarIssue
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Sentence != want.Sentence || got.RuleID != want.RuleID || got.Category != want.Category {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
	if len(got.Replacements) != 1 || got.Replacements[0] != "These" {
		t.Errorf("replacements = %v, want [These]", got.Replacements)
	}
}

func TestLanguageRoundTrip(t *testing.T) {
	want := languagetool.Language{
		Code:     "en-US",
		Name:     "English (US)",
		LongCode: "en-US",
	}
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got languagetool.Language
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Code != want.Code || got.Name != want.Name {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestHostConstant(t *testing.T) {
	if languagetool.Host != "api.languagetool.org" {
		t.Errorf("Host = %q, want api.languagetool.org", languagetool.Host)
	}
}

func TestCheckFromTestServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", ct)
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("request has no User-Agent")
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		if r.FormValue("text") == "" {
			t.Error("text form field is empty")
		}
		if r.FormValue("language") == "" {
			t.Error("language form field is empty")
		}
		resp := map[string]any{
			"language": map[string]string{"name": "English (US)", "code": "en-US"},
			"matches": []map[string]any{
				{
					"message":      "Grammar error",
					"shortMessage": "Grammatical error",
					"replacements": []map[string]string{{"value": "These"}},
					"offset":       0,
					"length":       4,
					"context":      map[string]any{"text": "This are wrong", "offset": 0, "length": 4},
					"sentence":     "This are wrong",
					"type":         map[string]string{"typeName": "Hint"},
					"rule": map[string]any{
						"id":        "THIS_NNS",
						"issueType": "non-conformance",
						"category":  map[string]string{"id": "GRAMMAR", "name": "Grammar"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := languagetool.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	c := languagetool.NewClient(cfg)

	issues, err := c.Check(context.Background(), "This are wrong", "en-US")
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("got %d issues, want 1", len(issues))
	}
	if issues[0].RuleID != "THIS_NNS" {
		t.Errorf("RuleID = %q, want THIS_NNS", issues[0].RuleID)
	}
	if len(issues[0].Replacements) != 1 || issues[0].Replacements[0] != "These" {
		t.Errorf("Replacements = %v, want [These]", issues[0].Replacements)
	}
}

func TestListLanguagesFromTestServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		langs := []map[string]string{
			{"name": "English (US)", "code": "en-US", "longCode": "en-US"},
			{"name": "French", "code": "fr", "longCode": "fr"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(langs)
	}))
	defer srv.Close()

	cfg := languagetool.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	c := languagetool.NewClient(cfg)

	langs, err := c.ListLanguages(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(langs) != 2 {
		t.Fatalf("got %d languages, want 2", len(langs))
	}
	if langs[0].Code != "en-US" || langs[0].Name != "English (US)" {
		t.Errorf("langs[0] = %+v, want {en-US, English (US)}", langs[0])
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := languagetool.DefaultConfig()
	cfg.Rate = 0
	cfg.Retries = 0
	c := languagetool.NewClient(cfg)

	_, err := c.Check(ctx, "test text", "en-US")
	if err == nil {
		t.Error("Check with cancelled context returned nil error")
	}
}

func TestRetryOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		langs := []map[string]string{{"name": "English", "code": "en", "longCode": "en"}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(langs)
	}))
	defer srv.Close()

	cfg := languagetool.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := languagetool.NewClient(cfg)

	langs, err := c.ListLanguages(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(langs) != 1 {
		t.Fatalf("got %d languages after retries, want 1", len(langs))
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}
