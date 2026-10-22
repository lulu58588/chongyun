<div align="center">

# 崇云 AI 网关 · Chongyun

### Every agent's model. One place.

**Claude Code on Kimi. Codex on DeepSeek. Gemini CLI on GLM. MiniMax Code on any model.**<br>
One local gateway serves them all: many providers, account failover, local models,<br>
and every protocol translated into the one an agent speaks.

</div>

---

## What this is

Every agent keeps its model in its own config file, with its own key and address, so a
subscription you paid for works in one agent only. Chongyun brings all of it into one local
gateway:

- **Any model in any agent**: pick one from a list; the tools rewrite that agent's own
  config fields, keeping its comments and layout.
- **Every provider everywhere**: an API key, a local model, or a subscription account you
  are signed in to.
- **Failover when a quota runs out**: routing groups move between models and accounts; the
  agent sees no error.

macOS, Windows and Linux, with a menu bar app, a terminal TUI, a web app and a CLI.

## Quick start

```sh
# install from source (needs Go 1.26+)
go install github.com/lulu58588/chongyun@latest

# or run it from the sources
make cli && ./chongyun serve
```

```sh
chongyun provider add deepseek sk-…            # a preset provider needs only a key
chongyun claude deepseek/deepseek-chat         # Claude Code on DeepSeek
chongyun codex moonshot/kimi-k2.5               # Codex on Kimi
chongyun group add "daily" models=a,b routing=smart   # failover when a quota runs out
chongyun quota                                  # what each subscription has left
chongyun tui                                    # everything, in the terminal
```

The gateway listens on `127.0.0.1:3425` and speaks OpenAI Chat, OpenAI Responses,
Anthropic Messages and Gemini.

## What this fork changes

This project is a derivative of [yetone/magpie](https://github.com/yetone/magpie) (MIT),
rebranded for **Chongyun Technology**.

**Removed:**

| Item | What this build does |
|---|---|
| Usage telemetry (PostHog) | **Removed entirely** — nothing is reported to any third party |
| The upstream update feed and auto-update | Off; point `MAGPIE_UPDATE_FEED` at a feed of your own |
| The upstream partners and sponsors list | Off |
| The upstream icon proxy | Fetched directly now (the host of a picture sees your IP; noted in the privacy section) |
| The upstream site and its cross-origin allowlist | The `site/` directory is gone; the allowlist is empty |
| **Reading other agents' sign-ins** (Claude Code / Codex / the keychain) | **Off by default** — run `chongyun accounts discovery on` to turn it on |
| The upstream icons, logo, product name and texts | All replaced with Chongyun's |

**Kept on purpose** (changing these breaks compatibility with the agents, and they are not
branding): the `~/.config/magpie` config folder · the `/v1/magpie/*` endpoints · the
`X-Magpie-*` headers · the `MAGPIE_*` environment variables · the provider id `magpie` in
every agent's config · the Go module path `github.com/lulu58588/chongyun`.

**Kept because the law requires it**: the MIT text and `Copyright (c) 2026 yetone` in the
root `LICENSE`, and the third-party notices in `internal/qoder/LICENSE` — all unchanged. The
fork's own notes are at the end of `LICENSE`.

## Safety

- On the loopback address the gateway **does not check keys** (upstream's design): any
  program on this machine can call it and spend your quota. Don't expose it to a network or
  to other people unless you put your own authentication in front.
- Reusing subscription accounts (Claude / ChatGPT / Copilot) is a grey area upstream. This
  build **reads no agent's sign-in by default**; even when turned on by hand, it may trip a
  vendor's own checks. That is your call to make.
- Sharing on the local network is off by default; when you turn it on, set
  `MAGPIE_ALLOW_IPS` (note: an empty allowlist allows every address).

## License

MIT — see [LICENSE](LICENSE). The upstream project's copyright belongs to yetone; this fork's
belongs to Chongyun Technology.
