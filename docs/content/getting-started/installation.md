---
title: "Installation"
description: "How to install the languagetool binary."
weight: 20
---

## Go install

```bash
go install github.com/tamnd/languagetool-cli/cmd/languagetool@latest
```

## Prebuilt binaries

Download a binary for your platform from the
[releases page](https://github.com/tamnd/languagetool-cli/releases).

Archives are available for Linux (amd64, arm64, armv7, 386), macOS (amd64,
arm64), Windows (amd64, arm64), and FreeBSD.

## Linux packages

Debian/Ubuntu:

```bash
# Download the .deb from the releases page, then:
sudo dpkg -i languagetool_*.deb
```

RPM-based (Fedora, RHEL):

```bash
sudo rpm -i languagetool_*.rpm
```

Alpine:

```bash
sudo apk add --allow-untrusted languagetool_*.apk
```

## Container image

```bash
docker run --rm ghcr.io/tamnd/languagetool check "This are wrong"
```
