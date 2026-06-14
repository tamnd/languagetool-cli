# languagetool-cli

A command line for LanguageTool grammar and style checking. One pure-Go binary, no API key, output that pipes into the rest of your tools.

## Install

```bash
go install github.com/tamnd/languagetool-cli/cmd/languagetool@latest
```

Or download a prebuilt binary from the [releases page](https://github.com/tamnd/languagetool-cli/releases).

## Quick start

```bash
# Check a sentence for grammar issues (English by default)
languagetool check "This are wrong"

# Check in a different language
languagetool check "Je suis fatigue" --lang fr
languagetool check "Das ist falsch" --lang de

# List all supported languages
languagetool languages

# Pipe output as JSON
languagetool check "This are wrong" -o json
```

## Output

Output is a table at the terminal and JSONL when you pipe, so `languagetool check "..." | jq` works with no flags.

## Languages

LanguageTool supports over 60 languages. Use `languagetool languages` to see all available language codes.

## License

Apache 2.0. See [LICENSE](LICENSE).
