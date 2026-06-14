---
title: "Introduction"
description: "What languagetool is and how it works."
weight: 10
---

`languagetool` is a command-line interface for the
[LanguageTool](https://languagetool.org/) grammar and style checking API.

It talks to the free public API at `api.languagetool.org` over HTTPS. No
account, no API key, no configuration needed to get started.

## What it does

- `languagetool check` submits text to the grammar check endpoint and prints
  any issues it finds: the problematic phrase, the rule that triggered, suggested
  replacements, and the category (Grammar, Style, Spelling, and so on).
- `languagetool languages` lists the 60+ languages LanguageTool supports.

## Output formats

By default, output is a table at the terminal. Pipe to another program and it
switches to newline-delimited JSON (JSONL) automatically, so
`languagetool check "..." | jq` works with no extra flags.

Pass `-o json` or `-o jsonl` to force a format.
