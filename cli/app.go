// Package cli builds the languagetool command tree on top of the languagetool
// library and the any-cli/kit framework. Every command is a kit operation:
// declared once and exposed as a CLI subcommand, an HTTP route, and an MCP
// tool, with the output formats handled by the framework.
package cli

import (
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/languagetool-cli/languagetool"
)

// Build metadata, injected via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// NewApp assembles the kit application: identity, defaults, client factory,
// and the languagetool operations.
func NewApp() *kit.App {
	app := kit.New(kit.Identity{
		Binary:  "languagetool",
		Version: Version,
		Short:   "Check text for grammar and style issues",
		Long: `languagetool checks text for grammar, spelling, and style issues using
the free LanguageTool public API. No API key required.

Quick start:
  languagetool check "This are wrong"
  languagetool check "Je suis fatigue" --lang fr
  languagetool check "Das ist falsch" --lang de
  languagetool languages`,
		Site: "languagetool.org",
		Repo: "https://github.com/tamnd/languagetool-cli",
	}, kit.WithDefaults(func(c *kit.Config) {
		c.Rate = 500 * time.Millisecond
		c.Retries = 3
		c.Timeout = 30 * time.Second
		c.UserAgent = languagetool.DefaultUserAgent
	}))

	languagetool.Domain{}.Register(app)

	return app
}
