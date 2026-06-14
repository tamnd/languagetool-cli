package languagetool

import (
	"context"
	"strings"
	"unicode"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go registers the languagetool kit Domain so a blank import in a
// multi-domain host (ant) enables the driver:
//
//	import _ "github.com/tamnd/languagetool-cli/languagetool"
//
// The Domain also builds the standalone languagetool binary via NewApp.
func init() { kit.Register(Domain{}) }

// Domain is the LanguageTool driver. It carries no state; the per-run client
// is built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme and the identity the single-site binary inherits.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme:  "languagetool",
		Aliases: []string{"lt"},
		Hosts:   []string{Host, "languagetool.org"},
		Identity: kit.Identity{
			Binary: "languagetool",
			Short:  "Check text for grammar and style issues",
			Long: `languagetool checks text for grammar, spelling, and style issues using
the free LanguageTool API. No API key required.

Quick start:
  languagetool check "This are wrong"
  languagetool check "Je suis fatigue" --lang fr
  languagetool languages`,
			Site: "languagetool.org",
			Repo: "https://github.com/tamnd/languagetool-cli",
		},
	}
}

// Register installs the client factory and the two LanguageTool operations.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "check",
		Group:   "check",
		Summary: "Check text for grammar and style issues",
		Args:    []kit.Arg{{Name: "text", Help: "text to check for grammar issues"}},
	}, checkText)

	kit.Handle(app, kit.OpMeta{
		Name:    "languages",
		Group:   "info",
		Summary: "List supported languages",
	}, listLanguages)
}

// newClient builds a Client from the resolved kit Config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	return NewClient(c), nil
}

// --- input structs ---

type checkInput struct {
	Text   string  `kit:"arg" help:"text to check for grammar issues"`
	Lang   string  `kit:"flag" help:"language code (e.g. en-US, fr, de, es)" default:"en-US"`
	Client *Client `kit:"inject"`
}

type languagesInput struct {
	Client *Client `kit:"inject"`
}

// --- handlers ---

func checkText(ctx context.Context, in checkInput, emit func(*GrammarIssue) error) error {
	issues, err := in.Client.Check(ctx, in.Text, in.Lang)
	if err != nil {
		return err
	}
	for i := range issues {
		if err := emit(&issues[i]); err != nil {
			return err
		}
	}
	return nil
}

func listLanguages(ctx context.Context, in languagesInput, emit func(*Language) error) error {
	langs, err := in.Client.ListLanguages(ctx)
	if err != nil {
		return err
	}
	for i := range langs {
		if err := emit(&langs[i]); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver ---

// Classify turns any accepted input into the canonical (uriType, id).
// A language code (2-letter or IETF like en-US) maps to ("lang", input);
// anything else is treated as text to check and maps to ("text", input).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("languagetool: empty input")
	}
	if isLangCode(input) {
		return "lang", input, nil
	}
	return "text", input, nil
}

// Locate returns the canonical URL for a (uriType, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "text", "lang":
		return "https://languagetool.org/", nil
	default:
		return "", errs.Usage("languagetool has no resource type %q", uriType)
	}
}

// isLangCode returns true when s looks like a BCP 47 language tag: 2-3 letters
// optionally followed by a hyphen and a region subtag (e.g. "en", "en-US").
func isLangCode(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 2 || len(s) > 8 {
		return false
	}
	parts := strings.SplitN(s, "-", 2)
	// primary subtag: 2–3 ASCII letters
	primary := parts[0]
	if len(primary) < 2 || len(primary) > 3 {
		return false
	}
	for _, r := range primary {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	if len(parts) == 2 {
		// region subtag: 2 ASCII uppercase letters or 3 digits
		region := parts[1]
		if len(region) < 2 || len(region) > 3 {
			return false
		}
		for _, r := range region {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				return false
			}
		}
	}
	return true
}
