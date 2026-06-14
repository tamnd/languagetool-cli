---
title: "languagetool"
description: "A command line for LanguageTool grammar and style checking."
heroTitle: "Grammar checking, from the command line"
heroLead: "A command line for LanguageTool. One pure-Go binary, no API key, output that pipes into the rest of your tools."
heroPrimaryURL: "/getting-started/quick-start/"
heroPrimaryText: "Get started"
---

`languagetool` checks text for grammar, spelling, and style issues using the
free LanguageTool public API. No API key required.

```bash
languagetool check "This are wrong"
languagetool check "Je suis fatigue" --lang fr
languagetool languages
languagetool check "text" -o json | jq
```

Output is a table when you are at a terminal and JSONL when you pipe.

## Where to go next

- New here? Read the [introduction](/getting-started/introduction/), then the
  [quick start](/getting-started/quick-start/).
- Installing? See [installation](/getting-started/installation/) for prebuilt
  binaries, packages, and one-line installers.
- Need every flag? The [CLI reference](/reference/cli/) is the full surface.
