<div align="center">

# 星鸣 AI 网关 · Starling

### Every agent's model. One place.

**Claude Code on Kimi. Codex on DeepSeek. Gemini CLI on GLM. MiniMax Code on any model.**<br>
One local gateway serves them all: many providers, account failover, local models,<br>
and every protocol translated into the one an agent speaks.

</div>

---

## What this is

Every agent keeps its model in its own config file, with its own key and address, so a
subscription you paid for works in one agent only. Starling brings all of it into one local
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
go install github.com/lulu58588/starling@latest

# or run it from the sources
make cli && ./starling serve
```

```sh
starling provider add deepseek sk-…            # a preset provider needs only a key
starling claude deepseek/deepseek-chat         # Claude Code on DeepSeek
starling codex moonshot/kimi-k2.5               # Codex on Kimi
starling group add "daily" models=a,b routing=smart   # failover when a quota runs out
starling quota                                  # what each subscription has left
starling tui                                    # everything, in the terminal
```

The gateway listens on `127.0.0.1:3425` and speaks OpenAI Chat, OpenAI Responses,
Anthropic Messages and Gemini.

## What this fork changes

This project is a derivative of [yetone/magpie](https://github.com/yetone/magpie) (MIT),
rebranded as **Starling · 星鸣**, by 崇云科技 (Chongyun Technology).

**Removed:**

| Item | What this build does |
|---|---|
| Usage telemetry (PostHog) | **Removed entirely** — nothing is reported to any third party |
| The upstream update feed and auto-update | Off; point `STARLING_UPDATE_FEED` at a feed of your own |
| The upstream partners and sponsors list | Off |
| The upstream icon proxy | Fetched directly now (the host of a picture sees your IP; noted in the privacy section) |
| The upstream site and its cross-origin allowlist | The `site/` directory is gone; the allowlist is empty |
| **Reading other agents' sign-ins** (Claude Code / Codex / the keychain) | **Removed entirely** — no agent's sign-in is read, and no other app's login file is imported |
| The upstream icons, logo, product name and texts | All replaced with Starling's |

**Kept on purpose** (changing these breaks compatibility with the agents, and they are not
branding): the `~/.config/starling` config folder · the `/v1/starling/*` endpoints · the
`X-Starling-*` headers · the `STARLING_*` environment variables · the provider id `starling` in
every agent's config · the Go module path `github.com/lulu58588/starling`.

**Kept because the law requires it**: the MIT text and `Copyright (c) 2026 yetone` in the
root `LICENSE`, and the third-party notices in `internal/qoder/LICENSE` — all unchanged. The
fork's own notes are at the end of `LICENSE`.

## Safety

- On the loopback address the gateway **does not check keys** (upstream's design): any
  program on this machine can call it and spend your quota. Don't expose it to a network or
  to other people unless you put your own authentication in front.
- Reusing subscription accounts (Claude / ChatGPT / Copilot) is a grey area upstream. This
  build **reads no agent's sign-in** — that whole layer was removed and cannot be turned back
  on, so it cannot trip a vendor's own checks either. Accounts you sign into through Starling
  itself (Codex PKCE, the Copilot device code, Google OAuth) are unaffected.
- Sharing on the local network is off by default; when you turn it on, set
  `STARLING_ALLOW_IPS` (note: an empty allowlist allows every address).

## License

MIT — see [LICENSE](LICENSE). The upstream project's copyright belongs to yetone; this fork's
belongs to 崇云科技 (Chongyun Technology); the product is Starling · 星鸣.
