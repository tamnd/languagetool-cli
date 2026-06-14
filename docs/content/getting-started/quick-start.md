---
title: "Quick start"
description: "Run your first languagetool command."
weight: 30
---

## Check a sentence

```bash
languagetool check "This are wrong"
```

This submits the text to the LanguageTool API in English (the default) and
prints any grammar issues it finds.

## Check in another language

```bash
languagetool check "Je suis fatigue" --lang fr
languagetool check "Das ist falsch" --lang de
languagetool check "Esto es incorrecto" --lang es
```

## List supported languages

```bash
languagetool languages
```

Prints all 60+ languages LanguageTool supports, with their codes.

## JSON output

```bash
languagetool check "This are wrong" -o json | jq '.[] | .message'
```

Pipe to get JSONL automatically, or pass `-o json` for a JSON array.
