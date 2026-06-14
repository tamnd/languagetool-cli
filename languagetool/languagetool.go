// Package languagetool is the library behind the languagetool command: the HTTP
// client, request shaping, and the typed data models for the LanguageTool API.
//
// The public API at api.languagetool.org is open: no API key, no auth. This
// package wraps it with a rate-limited client that the kit operations consume.
package languagetool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultUserAgent identifies the client to LanguageTool.
const DefaultUserAgent = "languagetool/dev (+https://github.com/tamnd/languagetool-cli)"

// Host is the LanguageTool API hostname.
const Host = "api.languagetool.org"

// GrammarIssue is one grammar or style match from the LanguageTool check endpoint.
type GrammarIssue struct {
	Sentence     string   `kit:"id" json:"sentence"`
	Offset       int      `json:"offset"`
	Length       int      `json:"length"`
	Message      string   `json:"message"`
	ShortMessage string   `json:"short_message"`
	Replacements []string `json:"replacements"`
	RuleID       string   `json:"rule_id"`
	Category     string   `json:"category"`
	IssueType    string   `json:"issue_type"`
	Context      string   `json:"context"`
}

// Language is one supported language entry from the languages endpoint.
type Language struct {
	Code     string `kit:"id" json:"code"`
	Name     string `json:"name"`
	LongCode string `json:"long_code"`
}

// Config holds constructor parameters for Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults for the LanguageTool public API.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://api.languagetool.org",
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Retries:   3,
		Timeout:   30 * time.Second,
	}
}

// Client is a rate-limited HTTP client for the LanguageTool API.
type Client struct {
	cfg  Config
	http *http.Client
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Check submits text to the grammar check endpoint and returns any matches.
func (c *Client) Check(ctx context.Context, text, lang string) ([]GrammarIssue, error) {
	form := url.Values{}
	form.Set("text", text)
	form.Set("language", lang)

	body, err := c.post(ctx, c.cfg.BaseURL+"/v2/check", form)
	if err != nil {
		return nil, fmt.Errorf("check: %w", err)
	}

	var wire wireCheckResult
	if err := json.Unmarshal(body, &wire); err != nil {
		return nil, fmt.Errorf("check: decode: %w", err)
	}

	out := make([]GrammarIssue, 0, len(wire.Matches))
	for _, m := range wire.Matches {
		ctx := m.Context.Text
		lo := m.Context.Offset
		hi := lo + m.Context.Length
		if lo < 0 {
			lo = 0
		}
		if hi > len(ctx) {
			hi = len(ctx)
		}
		if lo > hi {
			lo = hi
		}
		excerpt := ctx[lo:hi]

		repls := make([]string, 0, len(m.Replacements))
		for _, r := range m.Replacements {
			repls = append(repls, r.Value)
		}

		out = append(out, GrammarIssue{
			Sentence:     m.Sentence,
			Offset:       m.Offset,
			Length:       m.Length,
			Message:      m.Message,
			ShortMessage: m.ShortMessage,
			Replacements: repls,
			RuleID:       m.Rule.ID,
			Category:     m.Rule.Category.Name,
			IssueType:    m.Rule.IssueType,
			Context:      excerpt,
		})
	}
	return out, nil
}

// ListLanguages fetches the list of supported languages.
func (c *Client) ListLanguages(ctx context.Context) ([]Language, error) {
	body, err := c.get(ctx, c.cfg.BaseURL+"/v2/languages")
	if err != nil {
		return nil, fmt.Errorf("languages: %w", err)
	}

	var raw []struct {
		Name     string `json:"name"`
		Code     string `json:"code"`
		LongCode string `json:"longCode"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("languages: decode: %w", err)
	}

	out := make([]Language, 0, len(raw))
	for _, r := range raw {
		out = append(out, Language{
			Code:     r.Code,
			Name:     r.Name,
			LongCode: r.LongCode,
		})
	}
	return out, nil
}

// --- wire types ---

type wireCheckResult struct {
	Language struct {
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"language"`
	Matches []struct {
		Message      string `json:"message"`
		ShortMessage string `json:"shortMessage"`
		Replacements []struct {
			Value string `json:"value"`
		} `json:"replacements"`
		Offset  int `json:"offset"`
		Length  int `json:"length"`
		Context struct {
			Text   string `json:"text"`
			Offset int    `json:"offset"`
			Length int    `json:"length"`
		} `json:"context"`
		Sentence string `json:"sentence"`
		Type     struct {
			TypeName string `json:"typeName"`
		} `json:"type"`
		Rule struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			IssueType   string `json:"issueType"`
			Category    struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"category"`
		} `json:"rule"`
	} `json:"matches"`
}

// --- HTTP internals ---

func (c *Client) post(ctx context.Context, rawURL string, form url.Values) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			wait := time.Duration(attempt) * 500 * time.Millisecond
			if wait > 5*time.Second {
				wait = 5 * time.Second
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}
		body, retry, err := c.doPost(ctx, rawURL, form)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("post %s: %w", rawURL, lastErr)
}

func (c *Client) doPost(ctx context.Context, rawURL string, form url.Values) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			wait := time.Duration(attempt) * 500 * time.Millisecond
			if wait > 5*time.Second {
				wait = 5 * time.Second
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}
		body, retry, err := c.doGet(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) doGet(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace blocks until at least Rate has elapsed since the last request.
func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}
