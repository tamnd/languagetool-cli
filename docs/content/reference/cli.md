---
title: "CLI reference"
description: "Every command and flag."
weight: 10
---

## languagetool check

Check text for grammar and style issues.

```
languagetool check <text> [--lang <code>] [-o <format>]
```

**Arguments:**

- `text` — the text to check (required)

**Flags:**

- `--lang` — language code, e.g. `en-US`, `fr`, `de`, `es` (default: `en-US`)
- `-o` — output format: `table` (default at terminal), `jsonl`, `json`

**Examples:**

```bash
languagetool check "This are wrong"
languagetool check "Je suis fatigue" --lang fr
languagetool check "Das ist falsch" --lang de -o json
```

## languagetool languages

List all languages supported by LanguageTool.

```
languagetool languages [-o <format>]
```

**Flags:**

- `-o` — output format: `table` (default at terminal), `jsonl`, `json`

**Example:**

```bash
languagetool languages
languagetool languages -o json | jq '.[] | select(.code | startswith("en"))'
```
