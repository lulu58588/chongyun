# 星鸣参考手册

<a id="starling-reference"></a>

完整手册：星鸣接线的每一个 agent、每一个 provider 选项、网关、路由、plugin、CLI、Docker、同步，以及星鸣保存的文件。概览请看 [README](../README.md)。

改过项目文件夹的名字？见[迁移 Claude Code 会话](session-relocation.md)。

一处挑定每个 agent 的模型：Codex 跑 DeepSeek、Claude Code 跑 Kimi、Gemini CLI 跑 GLM，都在菜单栏里完成。[GitHub](https://github.com/lulu58588/starling)

`starling` 是一块屏幕，列出你机器上的每个 AI agent 以及它被设成的模型。点一个值，挑一个模型。整个应用就这些。

它住在菜单栏里：点一下图标，面板就落下来；同一块屏幕也能当作普通窗口打开（`starling`，或托盘菜单里的 *Open starling*），并且会停在你上次离开的那一页，就像 Mac 应用回到它原来的样子那样；另外还有一个终端版（`starling tui`）和一个纯 CLI。当没有别的星鸣实例（应用本身、`starling web`、`starling serve`）在服务网关时，终端版会在开着期间负责服务网关，并在状态行里说明这一点；连到星鸣网关上的 agent 会在它退出时失去连接。

```
  ◉ starling

  ▸ Claude Code   claude-fable-5-1[1m]                        ~/.claude/settings.json
    Codex         gpt-6-astra   effort medium
    Gemini CLI    gemini-3.1-pro
    OpenCode      anthropic/claude-sonnet-5   small anthropic/claude-haiku-4-5
    MiMo Code     anthropic/claude-sonnet-5
    Pi            openrouter/z-ai/glm-5.2:batch
    Goose         anthropic/claude-sonnet-5
    Cursor CLI    auto
    Copilot CLI   claude-fable-5

  ↑↓ agent  ·  ←→ field  ·  ↵ change  ·  s save profile  ·  p profiles  ·  q quit
```

- **一个二进制。** 带上桌面应用大约 30 MB（它通过 [Wails](https://wails.io) 使用系统 webview，不打包任何东西；在 Mac 上下载 15 MB），纯终端版本则不到 30 MB。支持 macOS、Linux 和 Windows。
- **外科手术式地改配置文件。** 只有你改动的那一个键会被碰到；你的 `settings.json`、`config.toml`、`opencode.jsonc` 或 `config.yaml` 里的注释、顺序和缩进都原样保留。写入是原子的。
- **每个 agent 只有一个 endpoint。** 星鸣跑一个本地网关，它说 OpenAI chat completions、OpenAI Responses 和 Anthropic Messages API，并把请求转发给真正提供该模型的厂商。Codex、Claude Code、OpenCode 以及其余 agent 全都指向 `http://127.0.0.1:3425/v1`，从同一份目录里挑模型；API 之间的翻译在星鸣这里完成，流式输出和工具调用也一样。
- **你的订阅，人人共享。** 在 Claude Code、Codex（ChatGPT）或 Copilot 里登录之后，这个登录就会作为一个 provider 出现：其它每个 agent 都能通过网关使用它的模型，既不用复制什么，也不用粘贴任何 key。
- **只有一个字段的 provider。** 选一个预设（Anthropic、OpenAI、Gemini、Amazon Bedrock 及其 Mantle endpoint、DeepSeek、Kimi、GLM、MiniMax、StepFun、Qwen、百度千帆、腾讯云 Token Plan、华为云 MaaS、火山方舟、Mistral、Groq、Nous Portal、xAI、OpenRouter、Together、Fireworks、SiliconFlow、NVIDIA NIM、ModelScope、AiHubMix、PipeLLM、302.AI、Ollama、LM Studio……），粘贴一个 key，完事。自定义厂商需要一个名字和一个 base URL。星鸣绝不从你的 shell 环境里读取 key。
- **真实的模型列表，没有编译进去的东西。** 手上有 key 之后，星鸣会问厂商它提供哪些模型，然后原样给出这些；[models.dev](https://models.dev) 目录补上名字、推理强度，以及那些自己没有列表的厂商的模型清单（厂商自己的条目没列出的模型，就按其它提供它的 provider 的叫法来命名，所以 glm-5-turbo 在 Zhipu 下和在 ZCode 下都叫 GLM-5-Turbo；Azure 部署则保留它自己的名字），并且在数据变旧之后于后台自我刷新。你可以挑每个 provider 暴露哪些模型，也可以全部暴露——今天上午刚发布的模型，下一次刷新后就在选择器里了。
- **每个 agent 自己的模型列表。** 在 Agents 页面里某个 agent 名字下面，"Showing 5 / 32 models"（显示 5 / 32 个模型）会打开它的列表：点一个模型，把它从那个 agent 的选择器里拿掉（Codex 的 `/model` 也在内，它的 ChatGPT 模型同样），或者再放回去；其它 agent 照旧能用它，而新模型会被显示出来。它的 "Only models I pick"（只显示我勾选的模型）开关把逻辑反过来：开启时勾上的模型会留下，而之后新加进来的模型——不管来自哪个 provider——在被勾上之前都不会进入那个 agent 的选择器和它的配置（`starling visible <agent> --only-picked`，用 `--show-new` 恢复）。按名字点名的模型无论哪种情况都照常可用。
- **配置档。** 把每个 agent 的设置按一个名字存档，然后一步全切回去。正在用的那个档——也就是切过去也不会有任何变化的那个——会被打上勾，托盘面板的 Profiles 按钮显示它的名字，`starling profiles` 也会给它标上 ✓。
- **真图标，不套框架。** 系统 webview 上的纯 HTML；品牌图标来自 [lobehub/icons](https://github.com/lobehub/lobe-icons)。

## Agent 列表

<a id="agents"></a>

| Agent        | 文件                              | 字段            |
| ------------ | --------------------------------- | --------------- |
| Claude Code  | `~/.claude/settings.json`         | provider、model、opus/sonnet/haiku/fable、自动模式、登录（通过星鸣） |
| Claude Desktop | `Claude/` + `Claude-3p/configLibrary/`，在 `~/Library/Application Support` 下（Windows 上是 `%LOCALAPPDATA%`，Linux 上是 `~/.config`）。以 MSIX 包形式装在 Windows 上的 Desktop 把它们放在自己的包里（`%LOCALAPPDATA%\Packages\Claude_<publisher>\LocalCache\Local`，或者 Store 包的 `AnthropicPBC.Claude_<publisher>`，它自己的数据在 `LocalCache\Roaming\Claude`）。2.31226 及以后打包版的 Desktop 改读真正的 `%LOCALAPPDATA%\Claude-3p`，所以对着打包版 Desktop，星鸣两处都写（[`desktopdir`](../internal/desktopdir/desktopdir.go)） | provider（它的第三方网关模式：Code 和 Cowork 走星鸣，不登录 Anthropic；需要重启 Desktop）、opus/sonnet/haiku/fable、subagent 模型与 subagent effort（星鸣的设置，由网关在 subagent 的工具轮次上套用；无需重启） |
| Codex        | `~/.codex/config.toml`            | provider、model、effort |
| Gemini CLI   | `~/.gemini/settings.json`、`~/.gemini/.env` | auth、model |
| OpenCode     | `~/.config/opencode/opencode.json(c)`（`$OPENCODE_CONFIG_DIR`） | model、small |
| OpenChamber  | `~/.config/openchamber/preferences.json`（`$OPENCHAMBER_DATA_DIR`；星鸣的 provider 写在 OpenCode 的配置里） | model、small（用它自己的默认值，覆盖 OpenCode 的） |
| MiMo Code    | `~/.config/mimocode/mimocode.json(c)` | model、small |
| DevEco Code  | `~/.config/deveco/deveco.json(c)`（`$DEVECO_CONFIG_DIR`；它的 key 是 `starling-deveco`，因为它的一部分请求会带上 OpenCode 的 User-Agent） | model、small |
| Pi           | `~/.pi/agent/settings.json`       | model           |
| StepCode     | `~/.stepcode/config.toml`（外加 `models.json`；`$STEP_CODING_AGENT_DIR`、`$STEPCODE_CONFIG_DIR`） | model |
| MothX        | `~/.mothx/settings.json`（`$MOTHX_DIR`；Windows 上是 `%APPDATA%\mothx`） | model、thinking（`defaultProvider`/`defaultModel`/`defaultThinkingLevel`；星鸣作为 provider `starling` 写在用户自己的那个旁边，`openai-chat` 带 `thinkingFormat: "openai"`） |
| Aside        | `~/.aside/u/0/settings.json`（外加 `models.json`；第一个账号，也是星鸣唯一接线的那个） | model、effort、fast、standard、deep、visual、image |
| OmO (omo-ai) | `~/.omo/agent/settings.json`（外加 `models.json`；`$OMO_CODING_AGENT_DIR`、`$SENPI_CODING_AGENT_DIR`） | model |
| Prime Agent  | `~/.prime/agent/settings.json`（外加 `models.json`；`$PRIME_AGENT_CODING_AGENT_DIR`） | model、thinking（Pi 的字段；它的 key `starling-prime-agent` 指明了它，因为它的请求不带 User-Agent；只改全局设置，`allowedModels` 和 `subagentDefaultModel` 保持原样） |
| Goose        | `~/.config/goose/config.yaml`     | model           |
| Cursor CLI   | `~/.cursor/cli-config.json`       | model           |
| Zed          | `~/.config/zed/settings.json`（Linux 上是 `$XDG_CONFIG_HOME`，Windows 上是 `%APPDATA%\Zed`） | model（一个 OpenAI 兼容的 `starling` provider；它的目录出现在 Zed 的选择器里） |
| VS Code (Chat) | `~/Library/Application Support/Code/User/settings.json` + `chatLanguageModels.json`（Linux 上是 `~/.config/Code/User`，Windows 上是 `%APPDATA%\Code\User`） | model（`chat.defaultModel`；一个 `starling` Custom Endpoint 组，它的目录出现在 Chat 的模型选择器里；VS Code 1.122+，不需要 Copilot 登录或 key）。每个 profile 在 `User/profiles/<id>/` 下自己那一对文件（在 `globalStorage/storage.json` 里列出）也得到同样的设置，除非该 profile 用的是默认值那一套 |
| VS Code Insiders (Chat) | 同样的文件，位置在 `Code - Insiders/User` 下，代替 `Code/User` | 与 VS Code 相同，自己单独一行；它的模型发送 token `starling-vscode-insiders`，因为它的 chat 的 User-Agent 就是 VS Code 的 |
| VSCodium (Chat) | 同样的文件，位置在 `VSCodium/User` 下，代替 `Code/User` | 与 VS Code 相同，自己单独一行；它的模型发送 token `starling-vscodium`，所以用量里把它们记作 VSCodium；它的 Chat 功能必须打开（`chat.disableAIFeatures=false`，以及在它的 product.json 里为 GitHub.copilot-chat 设置 `defaultChatAgent` 和 `trustedExtensionAuthAccess`）。用应用市场里的 GitHub Copilot Chat 0.48.1 时，它没有 Custom Endpoint provider，于是这个组改为写给它的 OpenAI Compatible provider（`customoai`，token 以 `x-api-key` 发送），而这个 provider 只有在 Copilot Chat 用个人 Copilot 套餐登录 GitHub 时才列出模型 |
| Zoo Code (VS Code) | VS Code 的 `Code/User/settings.json`（以及每个 profile 的那份）：`zoo-code.autoImportSettingsPath` 指向它旁边的 `globalStorage/zoocodeorganization.zoo-code/settings/starling-import.json` | model（一个 provider 配置 `starling`，OpenAI Compatible，由 Zoo Code 在启动时导入） |
| Copilot (JetBrains) | `~/.config/github-copilot` 下的 `byok.json`（设了 `$XDG_CONFIG_HOME/github-copilot` 时用它，Windows 上是 `%USERPROFILE%\AppData\Local\github-copilot`），这是 Copilot language server 自己的文件，Xcode 的和 Eclipse 的 Copilot 也读它 | provider（用户自己那个旁边的一个 `starling` custom endpoint provider：网关 `/v1/chat/completions` 上的每一个星鸣模型都带着 `starling-copilot-jetbrains` 发送；在 Manage Models 里被藏起来的模型，同步之后仍然藏起来；关掉时只删掉星鸣那三个 key）。由 JetBrains IDE 或 Android Studio 里的 `github-copilot-intellij` 插件找到。IDE 重启后才会列出，且只在 Copilot 用支持自带 key（Bring Your Own Key）的套餐登录 GitHub 时才列出（Free、Pro、Pro+；组织允许时还有 Business 和 Enterprise）；模型在 Copilot Chat 里挑选 |
| JetBrains Air | `~/Library/Application Support/JetBrains/Air` 下的 `acp.json`（Linux 上是 `~/.config/JetBrains/Air`，Windows 上是 `%APPDATA%\JetBrains\Air`）+ 它旁边的 `starling-opencode.json` | model（一个 `Starling` ACP agent：OpenCode 的 `opencode acp` 只走星鸣的 provider，它的模型和路由组出现在 Air 的模型菜单里；需要装好 OpenCode） |
| Copilot CLI  | `~/.copilot/settings.json`        | model           |
| Crush        | `~/.config/crush/crush.json`      | large、small    |
| DeepSeek Harness (dsh) | `~/.dsh/profiles/*/cordis.patch.yml`（`$DSH_HOME`；一个 custom provider，Starling，它的 key `STARLING_GATEWAY_KEY` 放在 `~/.dsh/.env` 里，dsh 自己的 key 仓库 `~/.dsh/.credentials.yaml` 也放一份，桌面应用读的就是它；当模型是星鸣的一个模型时，它的 `web-search-deepseek` 那一行也走网关，由网关用该模型或设置 › 网络搜索去搜索，除非 dsh 自己有 `DEEPSEEK_API_KEY` 或你自己的一行），dsh 0.1.5 之前则是 `~/.dsh/config.yaml` | model、effort |
| Reasonix Studio（2.x，或原生 Go 的 `reasonix` 2.x / 1.39.x CLI） | `~/.reasonix/config.toml` + `.env`（Windows 上是 `%APPDATA%/reasonix`；`$REASONIX_HOME`） | model（Executor）、planner（Plan）、effort（一个专门的 `starling` provider，Studio 和原生 CLI 共用） |
| Command Code | `~/.commandcode/settings.json`（外加 `providers.json`） | model |
| fx           | `~/.fx/settings.json`             | model（一个不要 key 的 `starling` provider） |
| omp (oh-my-pi) | `~/.omp/agent/config.yml`（外加 `models.yml`） | model |
| Devin        | `~/.config/devin/config.json`（Windows 上是 `%APPDATA%\devin\config.json`） | model |
| Hermes Agent | `~/.hermes/config.yaml`（Windows 上是 `%LOCALAPPDATA%\hermes`；`$HERMES_HOME`），以及 `profiles/` 下每个 profile 的那份 | model |
| Mister Morph | `~/.morph/config.yaml`（`$MISTER_MORPH_CONFIG`） | model、effort（网关上的 `llm` 用 `openai_response_compatible`，也就是 Responses API；切走之后它原来的设置会回来） |
| Kimi Code    | `~/.kimi/config.toml`（`$KIMI_SHARE_DIR`） | model（一个 `starling` provider；星鸣的模型出现在 Kimi 的 /model 里） |
| Qwen Code    | `~/.qwen/settings.json`（`$QWEN_HOME`） | model（星鸣的模型作为 `modelProviders.openai` 条目，靠一个 `STARLING_QWEN_API_KEY` 环境变量；settings.model；接线期间 `security.auth.selectedType` 为 openai） |
| Muse Code    | `~/.config/muse/settings.json`（`$XDG_CONFIG_HOME`） | model（endpoint_transport 指向网关，auth 为 none；星鸣的模型出现在 Muse 的列表里） |
| Empryo       | `~/.empryo/config.json` | defaultModel（`providers` 里一个指向网关的 `starling` provider） |
| Ante         | `~/.ante/catalog.json` 和 `settings.json`（`$ANTE_HOME`） | provider、model（一个基于 OpenAiCompatible 的 `starling` provider，不带星鸣的 auth：网关放这台机器用任何 token 进来，而 Ante 把没有 auth 的 provider 视为已认证；星鸣的模型出现在 Ante 的模型选择器里，而 Ante 启动时是否落在星鸣上，由它的设置里的 `provider` 决定） |
| AstrBot      | `$ASTRBOT_ROOT` 下的 `data/cmd_config.json`，否则是 `~/.astrbot`（桌面应用的那份） | model（`provider_sources` 里一个 `starling` 条目，星鸣的模型写进 `provider`，用户自己的条目保留；默认对话模型写进 `agent_runner.config.model.provider_id`，AstrBot 4.28 之前的 agent runner 则是 `provider_settings.default_provider_id`；接上之后要重启 AstrBot；Docker 装的那些够不着） |
| jcode        | `$JCODE_HOME` 下的 `config.toml`，否则是 `~/.jcode` | model（一个 `[providers.starling]` 配置，`openai-compatible`，里面放着星鸣的模型；`[provider] default_provider`/`default_model` 指向它，断开时把原来那份放回去；配置里用户自己的键，比如 `headers` 和某个模型的 `reasoning_effort`，保留；接上之后要重启已打开的 jcode 会话） |
| MiniMax Code (mcode) | `~/.minimax/config.yaml`（`$MINIMAX_DATA_DIR`） | model（一个 `starling` custom provider；星鸣的模型出现在它的 /model 里） |
| Droid (Factory) | `~/.factory/settings.json`（`$FACTORY_HOME_OVERRIDE`） | model（星鸣的模型作为 BYOK 的 `customModels`，出现在 Droid 的 /model 里） |
| Cline (CLI)  | `~/.cline/data/settings/providers.json`（`$CLINE_DIR`） | model、effort（星鸣接管它的 openai-compatible provider） |
| Qoder (CLI)  | `~/.qoder/settings.json`（`$QODER_CONFIG_DIR`） | model、effort（一个 `starling` custom provider；需要带 BYOK 的 Qoder 套餐） |
| Qoder CN (CLI) | `~/.qoder-cn/settings.json`（`$QODERCN_CONFIG_DIR`） | model、effort（同 Qoder；它自己的账号，一个带 BYOK 的 Qoder CN 套餐） |
| Grok Build   | `~/.grok/config.toml`（`$GROK_HOME`） | model、effort |
| Grok App     | `agent-home/config.toml`，在它的数据文件夹里（`$GROK_APP_HOME`，否则是 `~/Library/Application Support/com.grokapp.grok-app`、`%APPDATA%\grokapp\grok-app\data`、`$XDG_DATA_HOME/grok-app`，或 `~/.grok-app`） | model（Accounts 下一个 `[model.starling]` custom provider，`app_models` 里放着星鸣的模型及其 effort 档位；路由 `[models] default` 保持不变，直到用户在那里对星鸣按 Use（在星鸣里挑的 App 自己的某条路由会被写成这条路由，而星鸣这个 channel 则被去掉）；旧版星鸣设成 `starling` 的路由会被还原一次；断开时会去掉这个 channel，只有当路由仍然是 `starling` 时才把它放回去，否则在没有剩下路由的情况下回到 Connect 之前那一条；在 App 里挑的模型或默认 effort 保留；指令、MCP 和 skills 用的是资源库里 agent-home 上的 Grok App 那一行；RTK 对它和对 Grok Build 都没有 hook） |
| ZCode        | `~/.zcode/v2/provider_config.json` 和 `config.json`，在 ZCode 设置里改过的那个数据文件夹下（`dataBaseDir`）；设了 `$HOME` 时，`~` 指向它，Windows 上也一样，因为 ZCode 先读它再读 `%USERPROFILE%` | provider（星鸣的模型出现在 ZCode 的选择器里） |
| WorkBuddy    | `~/.workbuddy/models.json`（`$WORKBUDDY_CONFIG_DIR`） | provider（星鸣的模型出现在 WorkBuddy 的选择器里） |
| WorkBuddy AI | `~/.workbuddy-ai/models.json`（国际版） | provider（星鸣的模型出现在 WorkBuddy AI 的选择器里） |
| CodeBuddy Code | `~/.codebuddy/models.json`（`$CODEBUDDY_CONFIG_DIR`）、`settings.json` | model（星鸣的模型出现在它的列表里；对话模型） |
| Raycast      | `~/.config/raycast/ai/providers.yaml`，在用户的文件夹里，Mac 和 Windows 都是 | provider（Raycast AI 的 Custom Providers 里一个 `starling` 条目，星鸣的模型出现在它的模型选择器里；需要 Raycast Pro；用户自己的 provider、`tools` 和注释保留） |
| OpenPencil   | 配置文件目录下的 `openpencil/settings.json`：`~/Library/Application Support`（macOS）、`$XDG_CONFIG_HOME` 或 `~/.config`（Linux）、`%APPDATA%`（Windows） | provider（它的 `builtin_agents` 里一个 `starling` 条目，星鸣的模型出现在它的 chat 的模型选择器里；用户自己的条目和设置保留；OpenPencil 运行时拒绝操作，因为它退出时会把这个文件写回去） |
| T3 Code      | `~/.t3/userdata/settings.json`（`$T3CODE_HOME/userdata`） | provider（一个基于 Claude Code 的 `starling` provider 实例，星鸣的模型作为它的自定义模型） |
| OpenHanako   | `~/.hanako/provider-catalog.json` + `agents/<id>/config.yaml`（`$HANA_HOME`；它在运行时提供本地 API） | model（主 agent 的；星鸣的模型作为一个 provider） |
| AtomCode     | `~/.atomcode/config.toml`（`$ATOMCODE_HOME`） | model、effort（一个 `starling` provider 账号，按目录里的每个模型各写一张模型表，就像它自己的登录流程那样） |
| Snow CLI     | `~/.snow/profiles/starling.json`（`$SNOW_CONFIG_DIR`），把它设为当前配置并复制到 `config.json` | model（用户其它配置里的模型也一样；Snow App 通过 "Sync Snow CLI API config" 导入这个配置） |
| Alma         | Alma 的本地 API（`localhost:23001`，在 Alma 运行时；alma-server 的数据在 `$ALMA_DATA_DIR`、Linux 上是 `$XDG_DATA_HOME/alma` 或 `~/.local/share/alma`） | model（Alma 的默认值；星鸣的模型作为一个 provider），以及 Image Generation 的模型在 Auto 时，就是星鸣用来画图的那个模型 |
| Lody | 它自己的确认对话框（`lody://provider/import`） | 在本机上添加所选的 Claude、Codex、Pi、DSH、Kimi Code、Grok 和 Bub provider；星鸣不改动任何设置 |

按 provider 取值的 agent（OpenCode、MiMo Code、DevEco Code、Pi、OmO、StepCode、MothX、Aside、Goose、Crush、omp、Hermes Agent）取的是 `provider/model`。

只显示已经安装或已经配置过的 agent。

### 关于几个 agent 的说明

<a id="notes-on-some-agents"></a>

#### StepCode

<a id="stepcode"></a>

StepCode 通过 [`stepcode.go`](../internal/agent/stepcode.go) 配置。星鸣管理 `~/.stepcode/config.toml` 里的全局 `defaultProvider` 和 `defaultModel`（或者一个绝对路径 `STEP_CODING_AGENT_DIR` 的父目录），并用 Pi 的 provider 格式把自己的目录写进 `~/.stepcode/models.json`。星鸣只管选模型；思考档位由 StepCode 自己控制。项目本地的 `.stepcode/config.toml` 覆盖层优先于全局设置。改完模型后要重启 StepCode 或开一个新会话。检测严格按目录来（`~/.stepcode`），所以 `PATH` 上装着的 Smallstep 那个无关的 `step` 证书命令行工具永远不会被当成 StepCode。

可选的真实 CLI 验证测试这样跑：`STARLING_TEST_STEPCODE_BIN=/path/to/step go test ./internal/agent -run '^TestStepCodeRealCLISandbox$' -count=1`。它在一个隔离的临时沙箱里运行 StepCode，遥测关闭，也不带线上凭据。

#### MothX

<a id="mothx"></a>

MothX（oschina/mothx，原名 vibecoding）通过 [`mothx.go`](../internal/agent/mothx.go) 配置。星鸣把自己的 provider `starling` 写进 MothX 全局 `settings.json` 里的 `providers`（`$MOTHX_DIR`，否则 Windows 上是 `%APPDATA%\mothx`，再否则是 `~/.mothx`），并设置 `defaultProvider`、`defaultModel` 和 `defaultThinkingLevel`。这个 provider 用 `openai-chat` 配 `thinkingFormat: "openai"`，因此思考档位对每个模型都以 `reasoning_effort` 发出去；没有它的话，MothX 会从一个模型 id 猜这是 Qwen 还是 DeepSeek 的思考字段。用户往星鸣那一个块里加的东西（headers、他自己的 `thinkingFormat`）保留。key `starling-mothx` 指明了它，它的 User-Agent（`Vibecoding Client/<version>`）也一样。项目的 `.mothx/settings.json` 覆盖全局那份，并且不受改动。MothX 在启动时读自己的设置，所以改完要重启已打开的 `mothx` 会话。若期间没有别人改过 `settings.json`，断开时它会逐字节放回原样；若期间 MothX 或用户改过它（它的 `/model` 和 `/statusline … global` 会写整个文件），断开时只摘掉星鸣的 provider、它的选择，以及它留下的一个空的 `providers` 对象。

#### Reasonix Studio

<a id="reasonix-studio"></a>

Reasonix Studio 从它的桌面安装或原生 Go 的 `reasonix` 2.x、1.39.x CLI 来检测；历史上的 npm 包装和仅有一个共享配置都不算。已用 [Studio 2.24.0](https://github.com/esengine/DeepSeek-Reasonix/releases/tag/studio-v2.24.0) 测试过。

在应用里挑 Executor 的模型，或者用 `starling reasonix starling/deepseek/deepseek-chat`（路由组则是 `starling reasonix starling/group/code`）。Plan 单独选，用 `starling reasonix planner starling/deepseek/deepseek-chat`；用 `starling reasonix planner off` 关掉独立的 planner。`starling reasonix effort high` 设置一个对外声明的 Executor 推理档位。`starling reasonix default` 恢复之前的 Executor 选择；`starling reasonix planner default` 恢复 Plan。两个角色都不再需要时，星鸣的 provider 和那个放在 `.env` 里的私有 key 会被移除。新会话要重启 Studio；项目/会话级的覆盖仍然优先。其它 provider 和凭据保留。一个名字叫 `starling` 但不由本程序管理的 provider 属于冲突，只作报告，不会覆盖它。Studio 和原生 CLI 共用这些文件，所以只升级 Studio 并不会把它们的设置隔离开。

已发布客户端的流式/工具契约测试这样跑：`STARLING_TEST_REASONIX_CLI=/path/to/reasonix go test ./internal/agent -run '^TestReasonixStudioCLIIntegration$' -count=1`。它使用隔离的设置和一个本地的测试上游，不带厂商凭据。

#### VSCodium Chat

<a id="vscodium-chat"></a>

VSCodium 的 Chat 功能默认是关着的。要用 Chat 的模型选择器接上星鸣，在 VSCodium 的设置里把 `"chat.disableAIFeatures": false` 设上，并按 [VSCodium 的 Copilot 指南](https://github.com/VSCodium/vscodium/blob/master/docs/ext-github-copilot.md) 的要求，把 `defaultChatAgent` 和 `trustedExtensionAuthAccess` 两个条目加进 VSCodium 的 `product.json`。那份指南也解释了怎么装一个兼容的 GitHub Copilot Chat 扩展，因为 Open VSX 的注册表通常不提供微软的扩展。改完这些文件后重启 VSCodium，或者执行 **Developer: Reload Window**。之后星鸣的模型就出现在 `starling` 这个 custom endpoint 组下。

#### Zoo Code

<a id="zoo-code"></a>

Zoo Code（VS Code 扩展 `ZooCodeOrganization.zoo-code`）把它的 provider 配置存在 VS Code 的加密 secret 存储里，星鸣写不了那里。所以星鸣改用 Zoo Code 自己的 `zoo-code.autoImportSettingsPath` 设置：它写一个设置文件 `starling-import.json`，放在 Zoo Code 的 `globalStorage/zoocodeorganization.zoo-code/settings` 目录下，里面有一个名叫 `starling` 的 OpenAI Compatible 配置（网关的 `/v1`、key `starling-zoo-code`、选好的模型），并在 VS Code 的 `settings.json` 以及每个 profile 的那份里把设置指向它。Zoo Code 每次启动都会导入那个文件，所以改完要执行 **Developer: Reload Window**。连接期间，每次启动都会把 Zoo Code 切回 `starling` 配置。断开时把设置还原成原样并删掉那个文件，但导入的 `starling` 配置仍留在 Zoo Code 里：在那里挑另一个配置，或者在 Zoo Code 的设置里删掉它。只接线了 VS Code（Stable）。JetBrains 版 Zoo Code 启动扩展时用的是空配置，所以这个设置在那里够不着。

#### 兼容 Zed 的 agent 路径

<a id="zed-compatible-agent-paths"></a>

星鸣可以使用一个兼容 Zed 的 fork 或安装，只要它的可执行文件或配置目录与上游默认的不同。启动星鸣之前先设好这些变量：

```sh
export STARLING_ZED_BIN=/Applications/ZedG.app/Contents/MacOS/zedg
export STARLING_ZED_CONFIG_DIR="$HOME/.config/zed"
export STARLING_ZED_PROCESS_NAMES=zedg,ZedG
starling serve
```

`STARLING_ZED_BIN` 控制安装检测，`STARLING_ZED_CONFIG_DIR` 选定装着 `settings.json` 的目录，`STARLING_ZED_PROCESS_NAMES` 给出逗号分隔的进程名，用于重启提示。现有的 Zed 配置适配器被复用，所以这套变量是给那些保留了 Zed 的 `settings.json` 和 Agent 模型 schema 的 fork 用的，比如 [ZedG](https://github.com/x6nux/zed-globalization)。这些变量只影响设置它们的那个星鸣进程；在 systemd 或 Docker 下运行星鸣时，把它们放进服务环境里。它们配置的是与星鸣同一台机器上的兼容 Zed 的 agent；它们不会发现或修改跑在另一台主机上的 agent。在 macOS 上，如果配置的二进制解析不到一个 `.app` bundle，星鸣在授权网关凭据时会退回 Zed 应用的标准位置。

## Provider 与网关

<a id="providers-and-the-gateway"></a>

agent 能挑的每个模型都写成 `provider/model`，由星鸣的网关来提供，所以 agent 从来不持有厂商的 key 或厂商的地址。加一个 provider，它的模型就出现在每个 agent 的选择器里：

```sh
starling presets                        # the vendors starling knows, grouped: vendors, relays, local
starling provider add deepseek sk-…     # a preset needs only the key
starling provider add ollama            # local servers need none
starling provider add "My Relay" url=https://relay.example.com/v1 key=sk-… models=gpt-5.5,claude-sonnet-5
starling providers                      # host, key, exposed models, who uses what
starling provider deepseek              # one provider in detail
starling provider models deepseek       # re-fetch the vendor's list (add ids to choose which to expose)
starling provider models deepseek +deepseek-v4 -deepseek-chat # expose one more, take one out; the rest stay
starling provider models deepseek a b c # the whole list of models to expose, replacing it (all: the default)
starling provider refresh deepseek      # re-fetch it, and drop picks it no longer has (the TUI: m)
starling provider refresh-every relay 6h # the gateway fetches its list every 6h (1d, 30m; off; at least 10m)
starling provider auto-include relay 'gpt-*' 're:claude-.*' # a newly listed model these match is exposed (none clears)
starling provider test deepseek         # one tiny request per API, with latency
starling provider key deepseek sk-…     # replace the key
starling provider rm deepseek
starling models                         # the catalog agents see
starling claude deepseek/deepseek-chat  # use it
```

自定义 provider 可以取 `url=`（一个 OpenAI 兼容的 base）、`anthropic=`（一个 Anthropic 兼容的 base），或者两者都要；厂商另有 Responses endpoint 时再加 `responses=`，用 `catalog=` 借用一份 models.dev 清单，用 `models=` 指定要暴露的模型。预设不知道的任何东西，都可以用同样的方式覆盖。

某个模型如果厂商回答它已废弃或已下线（OpenCode Zen 的 410 "Model exo-free has been deprecated."），它仍会留在那个 provider 的模型里、留在 `starling models`、`/v1/models` 以及每个 agent 的列表里，即便厂商自己的清单还列着它。等一次以 `provider/model` 请求它的调用被应答之后，它就回来了，或者一周之后回来。

`starling provider set <id> header.<Name>=<value>` 会给 key+URL 型 provider 的每个请求都带上你自己指定的一个 header（值为空表示移除；已登录的账号会忽略它们）。它会替换掉星鸣本来要发的同名 header，`anthropic-beta` 除外——那是一个列表：你加的 beta 会排在这次请求里 agent 申请的那些之后（Claude Code 的、1M 上下文的 `context-1m-2025-08-07`（`[1m]` 模型时）、快速模式的），每样只加一次。provider 拒绝掉的某个 beta（`Unexpected value(s) … for the anthropic-beta header`）会从重试里去掉，并且从此以后你的和 agent 的都不再发；你的其它 header 则一直都会发。agent 自己登录带来的 beta（`oauth-2025-04-20`，Claude Code 登录 claude.ai 时申请的）永远不发，因为登录本身就不发；但你自己设置的那个会发。

请求发往厂商时用的是星鸣自己的 User-Agent（`starling/<version>`）。只服务它点名的那些 agent 的编程套餐，则认得出它们各自的：`starling provider set <id> user-agent=agent`（编辑器里的「Send the agent's own User-Agent」）会给 key+URL 型 provider 带上 agent 发给星鸣的那个 User-Agent，原样带上，在每一个 API 上都是如此（chat、Anthropic、Responses、Gemini、count_tokens、embeddings、images、videos）。agent 发请求时没带 User-Agent 的那些，以及星鸣自己发的那些（分类器的、图片工具的），仍然以星鸣的身份发。你设置的 `header.User-Agent` 比这两个都优先。`user-agent=starling` 可以再把它关掉；默认就是关的。已登录的账号、plugin 的、以及那些「星鸣以它自己的客户端身份去走」的厂商（OpenCode、Cline、Kilo、另一个星鸣），都保留它们登录时发送的 User-Agent，不接受这个设置。

`starling usage`（还有 `--csv`）可以取 `today`、`7d`、`30d` 或 `all`，每个都从本地时间的午夜算起，或者取一个到此刻为止的窗口：`--last 24h`（任意 Go duration，`90m`）或 `--days 1.5`；这时 CSV 里就只有那个窗口的调用（#2075）。

`starling usage` 还会列出**上游 provider key**，方便核对上游账单。每次调用都记下真正应答它的那把 key 的指纹和保存下来的名字，包括图片调用和账号/key 故障切换。CSV 里多了 `provider_key_id` 和 `provider_key_name`，JSON 里用 `providerKeyId` 和 `providerKeyName`，它们和网关的调用方 key 是两回事。System One 的调用用同样的归属方式。用量记录里不存任何原始凭据。轮换 provider 的第一把 key 不会把旧的用量迁到它的替代品上；被删掉的 key 保留它们各自的历史身份。更早的记录显示为**key not recorded**，绝不会用今天配置的 key 去反推。这些是上游凭据，不是客户端用来调用星鸣的 key。

CSV 里的 `cache_write_tokens` 仍然是全部缓存写入；新增的两列 `cache_write_5m_tokens` 和 `cache_write_1h_tokens` 按 Anthropic 保留它多久来拆开（在拆分之前记的算全是 5 分钟那档）。应用会在某次请求的 Cache 上以及用量磁贴的提示里显示这个拆分。

**请求与失败**按请求计数，每条按它最终的结果算一次：agent 拿到错误就算一次失败。后来某次重试应答了它（组里的下一个成员、另一个账号），就不算失败；用量页面、Requests 卡片、它的排名和组行，以及 `starling usage`，都把它记成重试后被应答（JSON 里是 `retried`）。在 Requests 卡片上，每个计数都是下面列表里对应请求的按钮（Failed 和 Retried 两个过滤器，`failed=1` 和 `retried=1`）。厂商拒绝或挡回的那次尝试会保留自己的一行，记下它被计费多少、为什么，并标记为已移交，且不计入请求数。CSV 的最后几列是 `tries`（在最后一行，也就是这次请求用了不止一次时）、`handed_on`，然后是 `tier_asked`、`tier_sent` 和 `tier_served`：客户端要求的服务档位（Claude 的快速速度读作 `fast`）、那次尝试实际发出去时的档位（选了快速就发 `priority`，或者在档位选成 Ultrafast 时发 `ultrafast`；套餐不提供 Ultrafast 时 Ultrafast 变成 `priority`），以及厂商回话说它实际跑在哪个档位（回话没提就是空）。JSON 日志里它们是 `tier_asked`、`tier_sent` 和 `tier_served`，而 `/v1/starling/route` 里的每次尝试都有 `tier` 和 `tier_served`。最后一列 `upstream_stop` 是上游说它的回话为什么结束，用它自己的说法（`stop`、`tool_calls`、`end_turn`、Responses 回话的 `completed`），它没说就是空。Requests 页面上某次请求的那一行把它显示为 Upstream stop reason。

**输出速度**（用量页和路由页上的 tok/s，`starling usage`）是输出 token 除以第一个 token 之后的时间，按请求各自求和：每个计了时的请求的 token 之和，除以它们各自窗口之和，所以并发的子 agent 不会互相累加。一个推理过的回话只算它的答案部分，从它第一次出现文本开始算（`usage.DecodeOf`）：隐藏的推理是在流里显示任何东西之前写下的，把它算进一个从它之后才开始的窗口里，会让 gpt-6.1-sol 显示成每秒几百 tok/s。一个推理之后只写了工具调用（且不是流式的）的回话，以及一次爆发（低于 100 ms，或高于 10,000 tok/s），都说明不了速度。厂商扣住之后在末尾一下放出来的回话同样说明不了，无论那下爆发花了多久才来：内容在它窗口的四分之一之内就到的那个（`flow_ms`，从它字节的十分之一到十分之九，再按整体缩放），或者推理之后、它的答案本该比推理快 20 倍以上到达的那个（它的文本是跟着爆发一起来的）。一次爆发花了多久才到，不是它写了多久，那个时间也看不见。历史也用同样的方式读，因为它保存了推理、第一次文本和流量；在开始保存流量之前的记录，只能靠推理的节奏来判断。

一次从 agent 的会话文件里读出来、而不是网关亲眼看见的调用，没有第一个 token：Claude Code 的文件只有在请求回话的那一行和回话的最后一行都写下之后才有。所以 Requests 标签页上这样的行改为显示它的回话至少是按什么速度写出来的（`≥98 tok/s`）：它的输出 token 除以整段时间，其中就包含等第一个 token 的那段时间，所以第一个 token 之后的速度绝不会低于它。只有写出 1,000 token 或更多、且用了一秒或更久的回话才会显示一个；更短的回话里等待占的时间太大，说明不了任何事；能说明问题的是一个 token 数，不是时长（一个 5 秒的回复里可能有 4 秒都在等）。Claude Code 的、Claude Desktop 的和 Codex 的行会显示：Claude 那条请求回话的行是在发出之前写下的，而一轮里有好几次调用的 Codex，其计时从该次调用第一个输出项之前模型最后拿到的东西算到最后一项，绝不从该次调用自己跑的工具的输出来算——更早的 rollout 会把那个写在调用的 token 数之前。OpenCode 的行不显示。一个由会话文件计时的调用（一轮里只有一次调用的 Codex）在第一个 token 之后的时间达到一秒或更多、且在那期间写了 100 token 或更多时，显示它自己的速度（`usage.CodexWindowMs`、`usage.CodexBodyTokens`）。那段时间可能只是一次网络推送而不是在写：一个第一个 token 出现在 6.8 秒、其余 0.2 秒后到的回话，读作 1,900 tok/s。当那次调用做了推理时，文件里有它的第一个 token 但没有它的第一次文本，所以测不出速度；它改为显示它的答案至少是按什么速度写出来的：它的输出减去推理，全部写在第一个 token 之后，除以从第一个 token 到它结束的时间，同样以一秒和 100 token 为下限。它在整个输出上除以那段时间会把写在第一个 token 之前的推理也算进去，读出来的速度偏高。一轮里只有一次调用、两者都说明不了的，显示它至少是按什么速度写出来的，除以整轮时间（其中包含等第一个 token 的那段时间）（1,200 token 用 10 秒就是 `≥120 tok/s`），用的是任何「对整段时间取下限」都用的同一条规则：1,000 token 或更多、且一秒或更多。低于这个门槛时等待占的时间太大，说明不了任何事，所以一个 11 token 的答案花 2.75 秒就不显示，815 token 的 Claude Code 回话也不显示。网关的行和其它 agent 的行保留它们自己的速度。下限永远不会混进一个实测速度里：Speed 趋势和 `starling usage` 只统计计了时的调用。

一个组的速度（用量页面排名里的一行，按模型、provider 或 agent，它的 Output speed 卡片，以及概览里的那些列表）说的是它由多少个回话构成，遵循一条规则：对于速度被测出来的回话，给出它们的速度和它们有多少个（`62 tok/s · 2 timed replies`，而不只是旁边写「260 requests」）；对于只能给出下限的行，给出它们至少写出了多少（`≥100 tok/s · 30 replies`）；两者都有时，两个都给，回复多的那个在前（`≥50 tok/s · 259 replies; 125 tok/s · 1 timed reply`），因为少数被测出来的回话未必能代表其余的；两者都没有时，什么都不给。计数用文字表达，绝不是括号里的一个裸数字。Output speed 卡片把第一个作为它的数值，把它的计数放在旁边，把另一个放在下面，并把第一个 token 的平均等待时间单独放一行，写明它是多少个计时的回话的平均。提示文字说明每一个是什么。按速度排的排名，用它最先给出的那个数值来给一个组排序和画条（`usage.Totals` 的 `decode_n`、`floor_out`、`floor_ms`、`floor_n`；`usage.Row.Floor`）。

它也列出**账号**：每个 Codex、Claude 或其它订阅账号的 token 和成本，按真正应答的那个账号归——故障切换后接手的那个，或者 `X-Starling-Account` 钉住的那个。账号的名字就是路由视图里的叫法（它的邮箱或登录名，绝不是 token）；CSV 里多了 `provider_account`（JSON 里是 `providerAccount`），`starling usage --csv --account <name>` 只保留某一个账号的调用，而应用的用量页面上有一个 Accounts 列表和 Requests 上的 Account 过滤器。更早的记录在它那台主机说明的情况下会写出它的调用是以哪个账号发出的（`chatgpt.com as dee@example.com`）；否则它显示为 **account not recorded**，绝不用今天的登录态去反推。OTLP 导出从不带账号。

一个星鸣可以同时服务好几台电脑（办公室那台、家里那台）：把它共享到网络上（设置 → 在局域网上共享），然后在其它每台电脑上把它加成**Remote starling**——在应用的添加面板里，在它自己那个标题「Another computer's Starling」下，排在「On this machine」之后；在 TUI 的 Providers 页面上（按 `a`，然后是它的地址和 key；`w` 改地址，`m` 重新取它的清单），或者用 `starling provider add remote-starling sk-starling-… url=http://192.168.1.20:3425 id=office`。每台电脑自己的星鸣仍然接线它自己的 agent，而 provider、路由组（`office/group/…`）和用量则来自共享的那台。请求会沿着 agent 说的那个 API 继续走——Anthropic Messages、Responses、Chat Completions、token 计数——而共享那台的 provider 只在另一个 API 上提供的模型，会被转成那个 API 一次，不会两台电脑都转。它的清单是共享那台给它的 agent 展示的模型，每个都带上那里的 provider 名字（`Claude Sonnet 5 · Relay A · office`）——id 还是 id（`office/relay-a/claude-sonnet-5`），只是这些标签带上了名字。它的图片和视频模型会被自动发现，并通过它的图片生成/编辑和视频创建/状态/内容 API 提供。System One 也自动在它的 `/v1/systemone` 配置好：刷新远端 provider 的模型清单就能发现它的决策模型，包括 Jev、Clef 和带自定义名字的模型。这些显示为决策模型，而不是对话模型。像 `office` 这样一个裸的 provider id 会请求清单里第一个决策模型；它不会为了满足某个 key 的额度限制而改挑别的模型。如果这个 key 只允许另一个模型，就用它完整的 `office/provider/model` id，免得撞上 403。Embeddings、rerank 和语音转文字的请求走远端对应的 API，而它暴露出来的检索模型在取清单时会保留。这些在 agent 的模型列表里和远端上一样；列出来并不意味着一个 embedding 或 rerank 模型能撑起一场对话。远端必须提供被请求的那个 API 和模型，并允许所给的那把 key 使用它们。Responses 走 HTTP（包括 SSE），而不是 WebSocket。没有决策模型发现的旧远端仍然提供它已有的那些 API；升级它并刷新清单，就能在这边暴露 System One 的模型。它的额度也会显示：用量页面、菜单栏和 `starling quota` 会列出共享那台的卡片，并带上它的名字（`Codex · office`，id `office/codex`），内容是它最近一次读到的——只有持有登录态的那台电脑才会去问厂商，而它是用自己保存下来的内容作答（`GET /v1/starling/quotas/cards`）。在这里刷新一张卡片，会让它再读那张卡片一次（`POST /v1/starling/quotas/refresh`，同一张卡片 30 秒内最多一次）；在它读到任何东西之前，远端的那张卡片会这么说。它们的历史也各自带着自己的显示。Codex 的重置或签到是在那台上按的，不是在这里。

Codex 自带的图片工具会先向给它这一轮对话提供服务的那个 provider 要那个图片模型。如果那个 provider 没有列出这个模型，或者认不出是哪一轮，它就改用设置 → 图片 → 图片生成（包括「自动」）。关掉这一项只是关掉这个兜底，不会让这一轮 provider 上本来匹配的模型也没了。显式写出的 `provider/model` 保持它的去处；上游的图片错误不会通过这个设置去重试。不需要星鸣 Image MCP。

WorkBuddy（中国版）的每日签到可以从应用的用量卡片（*Check in now*）、`starling accounts checkin`，或者 TUI 用量页面上的 `c` 来按，后者会一次给本机登录的所有账号签到；它的 WorkBuddy 各行会说明今天的情况。共享那台的账号要在那台上签到，用它自己的应用、TUI 或 CLI。

有两个或更多账号的厂商，会在**设置 → 用量**下的「每日签到」里把每一个都列出来，各自带自己的开/关（#1902）。被关掉的账号会被排除在每日签到和*Check in now*之外，CLI 和 TUI 的也一样，而厂商的其它账号照常签到；它的用量卡片会说明它的自动签到已关。这个选择保存在 `settings.json` 的 `checkinOff` 里，由该账号在厂商签到文件（`workbuddy|<uid>`）里的条目决定，并用 `POST /api/settings/checkin-account {key, on}` 设置。

`auth` 钩子里带 `checkin` 的 plugin 也用同样的方式给自己的账号签到：只要**设置 → 用量 → Plugins**下它的开关是开的，每天一次，结果显示在每个账号的用量卡片上。厂商要求验证码的签到会照实标出来（去厂商的应用里签）；星鸣绝不破解验证码。见 [provider-plugins.md](subsystems/provider-plugins.md#a-plugins-daily-check-in)。

plugin 也可以为自己的 provider 声明自己的开关（`auth.options`）。它们列在 provider 编辑器的 **Plugin options** 下，每一个都是开/关，配上 plugin 自己的措辞，并随 provider 一起保存。这些值保存在 `settings.json` 的 `pluginOptions` 里，plugin 在工作时读它们。改动其中一个本身不会启动任何东西。见 [provider-plugins.md](subsystems/provider-plugins.md#a-plugins-own-options)。

网关为客户端签发**网关 key**，它们和 provider 的上游 API key 是两回事。打开**设置 → 在局域网上共享**，然后进入**网关 → 网关 key → 添加网关 key**。这一块只在共享打开时出现。为每个客户端建一把有名字的 key，并从它那一行复制。要从另一个网关搬过来、又不动它的客户端，就把它已经在发送的那把 key 填进 **Your own key (optional)**（8–256 个字符，不能有空格，不能以 `sk-starling-` 开头，这一点和其它任何 key 都不同）；留空的话，星鸣会自己造一把。列表只显示它最后四个字符，而它和其它任何 key 一样会被计数、限额和限制，从这台电脑上用也一样。各把 key 可以独立改名、禁用、轮换或删除；轮换和删除会要求确认。轮换保留名字、启用状态和用量历史；其它 key 不变。共享打开时，**网关 → 连接**会给出回环地址和共享地址，以及已启用的网关 key，供所有连接示例使用。共享关闭时，连接页保留原来的 **API key** 字段和本地 `starling` token，没有网关 key 选择器。它那个任意的本地选项是 **This computer**，与那个具名的 **Starling** key 不同。

要一个无界面的网关，就在暴露端口之前先用 CLI：

```sh
starling gateway-key add "Remote laptop" # prints the new credential once
starling gateway-key add Family --key - # keeps a key clients already send, read from stdin
starling gateway-key list              # ids, names, enabled state and masked keys
starling gateway-key rotate <id>       # prints the replacement; identity stays the same
starling gateway-key remove <id>       # revokes remote access
starling gateway-key limit <id> week --tokens 2m --cost 5 # its own limit
starling gateway-key limit <id> list   # each limit, used, left and reset
starling gateway-key limit <id> 10d --tokens 5m # every 10 days, from now
starling gateway-key limit <id> 5h --tokens 200k # every 5 hours, from now
starling gateway-key limit <id> add week --cost 20 # one more limit; all must pass
starling gateway-key limit <id> remove 2 # take off the 2nd limit
starling gateway-key limit <id> reset  # count its usage from 0 again (reset 2: only the 2nd)
starling gateway-key limit <id> off    # no limit
starling gateway-key models <id> openai/gpt-5 anthropic/* # only these models
starling gateway-key models <id> all   # every model
starling gateway-key accounts <id> codex/me@example.com # only these accounts
starling gateway-key accounts <id> all  # every account
starling gateway-key quotas <id> on    # told the quotas of what it may use
```

每把网关 key 都可以有自己的**限额**：一个 token 总量、一个按美元估算的成本，或两者都要，按天、按周或按月（本地时间的日历窗口：从午夜、从周一、从 1 号），每 N 天一次（#1509；N 从 1 到 3650）或每 N 小时一次（#2102；N 从 1 到 8760，比如订阅的 5 小时节奏就取 5）。一个 N 天或 N 小时的周期从设定限额的那一刻开始，跑 N×24 或 N 小时，然后下一个周期从它结束的地方接着开始，所以重启、改时区或夏令时切换都不会让它挪位。只改上限数值或 **Count cache reads too** 会保留这个周期；改 N 或周期长度，或者把限额打开，则从那时起开始一个新的周期。用 key 那一行的 **Limit** 来设置（随 Save 一起保存），或者用 `starling gateway-key limit`；那一行显示这个 key 用了多少、还剩多少以及何时重置。一把 key 最多可以有五条限额，各自在自己的窗口里（#2227），比如每 5 小时 $3、每周 $20，就像订阅的两个窗口那样：在编辑器里用 **Add a limit**，或者用 `starling gateway-key limit <id> add`。只要每一条限额都还有余额，请求就放行；它的预留会在每个窗口里被占住，而它的 429 会指出最后重置的那条已经花完的限额。两条限额不能共用一个窗口：一个窗口的 token 和成本上限写在同一条限额里。不带 `add` 设置某个窗口，改的是这个 key 在那个窗口上的限额，或者它仅有的那一条。`GET /v1/starling/limit` 在 `limits` 里给出每一条，在 `limit` 里给出占用这个 key 最多的那一条，早于 #2227 的星鸣读的是后者；这样的星鸣也只保留第一条限额（caller-keys.json 里的 `limit`，其余在 `moreLimits` 下）。被计入的 token 是一次调用的未缓存输入、输出和缓存写入，再加上开启 **Count cache reads too** 时的缓存读取。成本是按用量页面的价格估算的；没有已知价格的调用不计入任何成本。调用算在它开始的那个窗口里。一旦某个 key 花完，它的请求会在问任何 provider 之前被拒绝，用 API 自己的错误格式返回 429，并写出那个 key、那条限额和重置时间，外加 `Retry-After`；其它 key 不受影响。这些计数是从用量日志里读出来的，所以重启后仍在。一路在飞的请求会占住一个预留（它请求体的 token 数加上该 key 每次调用的平均输出），所以同时发出的请求会超出一整个调用左右；流式的回话在结束时按厂商报上来的用量结算。key 的限额编辑器里的 **Reset usage**（或者 `starling gateway-key limit <id> reset`）从此刻起把该 key 的 token 和成本一起从 0 重新计：编辑器的重置它当前打开的那条限额，CLI 的重置所有限额，或者用 `reset <n>` 重置第 n 条。N 天或 N 小时的 key 从那时起开始一个新的周期。按天、按周或按月的 key 保留它窗口的终点（午夜、周一、1 号）：只有重置之前的调用不再计入，所以它不会偏离日历。重置状态和 key 一起保存（caller-keys.json 里的 `since`，星鸣自己写的：客户端发来的值会被忽略），因此重启后仍在；用量日志本身不动。更早保存的 key 保留它们原来的 JSON。一个 key 可以用 `GET /v1/starling/limit` 读自己的状态；不管请求里写了什么，它只会被告知自己的情况，不会知道别的 key 的。把这一台作为 Remote starling provider 的星鸣，就用这种方式在它的用量页面上显示每一把 key 的限额（#1547）。从这台电脑发出、不带网关 key 的请求不受限额；用在这台电脑上的网关 key 受限。

网关 key 也可以被限制在某些**模型**上（#882）：用 key 那一行的 **All models** 徽章来挑，或者运行 `starling gateway-key models <id> openai/gpt-5 anthropic/*`（`all` 取消这个限制）。一个模式是 `<provider>/<model>` 或 `<provider>/*`，它按真正提供这次调用的那个 provider 来匹配，所以一个裸的模型名会先被解析出来；或者是一个路由组 `group/<id>`（`group/*` 表示所有组）。一把点名的 key 可以用这个组里的每一个成员，但它不能用那些被点名单独请求的成员。这样的 key 在 `/v1/models`、`/v1/codex/models` 以及 Anthropic 和 Gemini 的清单里只看得到自己的模型；它没点名的路由组只有在它能用组内每个成员时才看得见；别的模型一律在问任何 provider 之前就被拒绝，用 API 的错误格式返回 403；它无权使用的回退会被跳过。一把没有列出任何模型的 key 可以用所有模型。

网关 key 也可以被限制在某些**账号**上（#905）：它的请求可以用的账号和 key，在同一个菜单里、紧跟在该 provider 的模型之后挑——账号按登录的人来认，key 按它的指纹来认——或者用 `starling gateway-key accounts <id> codex/me@example.com openai/<key id>`（`all` 取消这个限制）。这份清单把一把 key 限制在**它点名的那些 provider 的**某些账号上：它点名了却没有账号的 provider，key 还是照旧那样用。清单按账号的稳定 id 保留它们，这个 id 跨改名保持不变，界面上按登录的人显示；后来没了的账号——被轮换的 key、被登出的账号、一个不留登录信息的 agent 的账号改了名——仍按原样保留，谁都匹配不上，所以这把 key 被收得更紧，绝不会更松。key 点名的路由组只通过它可以用的那些账号来覆盖它的成员：点名一个组不等于点名它的账号，而星鸣替这把 key 发出的调用——一次联网搜索、一张描述出来的图片、一个 Codex 标题——用的是用户挑的模型，但花的钱记在某个账号上，所以账号也管着它们。当一个模型背后所有的账号或 key 都是这把 key 不许用的，请求会在问任何 provider 之前被拒绝，用 API 自己的错误格式返回 403，并写出那个模型和它可以用的那些账号——它在一个用量上限的 429 之后、在为该模型被禁用的账号之前——而 `/v1/models` 给这把 key 列出的只有它可以用的某个账号或 key 所提供的模型，路由视图的「未列出」清单标出这把 key 排除了什么。一把没有列出任何账号的 key 可以用所有账号，就像 key 一直以来那样。

在局域网共享打开期间，或者 `STARLING_ADDR` 把网关放到网络上（一个不是回环也不是 `localhost` 的主机：Docker 镜像里的 `0.0.0.0`、一台服务器的地址），远端请求需要一把已启用的网关 key，以 Bearer、`x-api-key`、`x-goog-api-key` 或 `?key=` 的形式发送；不带就得到 401。对端不是回环的请求算远端请求，这包括 Docker 的端口发布（对端是网桥，不是主机的回环）和 NAT 下的 WSL 发行版。回环依然宽松：任何 token 都行，包括过期的或已禁用的网关 key。只有有效且已启用的 key 才会归属到它那个具名身份上。既没共享也不在网络上时，远端请求被拒绝（403）。`STARLING_ADDR` 固定到某个网卡的地址时，即便是这台电脑上的 agent 调用那个地址也要 key；而 `0.0.0.0` 配 `127.0.0.1` 上的 agent 则不用。Claude 订阅自己那些回调星鸣的工具调用，靠它们那次运行的一次性 token 放行，不需要 key（#1546）。共享会监听每一个网卡，除非 `STARLING_ADDR` 指定了它自己的某个主机：`STARLING_ADDR=127.0.0.1:3425` 放在 Tailscale Serve 后面，或者某个网卡的地址，在共享期间留在原处，而从别处到达它的请求仍然需要一把已启用的网关 key（#1112）。

**另一台电脑上的 Codex**（#1281）可以把共享的星鸣指定为它的模型 provider，配一把网关 key，并从 `GET /v1/codex/models` 按 Codex 自己的目录格式读到星鸣的模型。`/v1/models` 是 OpenAI 的清单，Codex 读不了；`/backend-api/codex/models` 是 ChatGPT 后端的，给的是登录在网关自己那台电脑上的 Codex 用的，网关 key 在那里会得到 401。

```toml
model_provider = "starling"
model = "provider/model"

[model_providers.starling]
name = "Starling"
base_url = "http://192.168.1.20:3425/v1"
model_catalog_url = "http://192.168.1.20:3425/v1/codex/models"
env_key = "STARLING_API_KEY"   # the gateway key
wire_api = "responses"
supports_websockets = false
```

这份清单是共享的星鸣给 Codex 看的模型（它的 Agents 页面挑的那些），每一个都以 `model_catalog_json` 的形式给出——那是星鸣为在自己电脑上的 Codex 写下的描述——并且只有这把 key 可以用的那些模型。Codex 0.161 默认就读它；0.160 需要 `[features] api_key_model_discovery = true`。Codex 最多读 1 MiB 的目录，而每个条目都带着 Codex 的提示词，所以超过大约 125 个模型之后，清单只保留放得下的前几个，而应答的 `X-Starling-Left-Out` 响应头会说明它排除了多少个：给 Codex 少挑一些，或者把 key 限制得更少。

一个经由本机的代理或隧道到达回环地址的请求（Cloudflare Tunnel 的 `cloudflared`、ngrok、Tailscale serve 或 funnel、frp 的 HTTP 代理、Caddy、nginx）算是别人的请求，并按来自另一台机器的请求来应答（#1022）：它带一个转发 header（`Forwarded`、`X-Forwarded-For`、`X-Real-IP`、`Cf-Connecting-IP`、`True-Client-IP`、`Tailscale-User-Login`），而没有 agent 会发这些。没共享时它被拒绝；共享打开、或者 `STARLING_ADDR` 在网络上时，它需要一把已启用的网关 key，和从网络上来的请求一样。这覆盖每一条路由：模型 API、额度、MCP 登录。要通过 Cloudflare Tunnel 提供星鸣，打开**设置 → 在局域网上共享**，为每个客户端加一把网关 key，把隧道的服务指向 `http://127.0.0.1:3425`，并把隧道的地址连同那把 key 给客户端。只转发裸 TCP 的隧道（frp 的 `tcp`、`ssh -R`、socat）不加任何 header，无法和这里的 agent 区分：把它指向这台电脑的局域网地址并打开共享，而不是回环。本机上一个自己给客户端签入的代理可以用 `STARLING_TRUST_PROXY=1` 信任，它转发的东西仍算本机，和以前一样。

要只让某些地址的其它电脑进来，就用 IP 和 CIDR 段把它们列在**设置 → 网络与共享 → 允许的地址**里，或者列在 `STARLING_ALLOW_IPS`（`203.0.113.7,10.0.0.0/8,2001:db8::/32`）里，后者优先。其它任何地址来的请求都被 403 拒绝，即便带着已启用的网关 key，拒绝信息会写明它来自哪个地址。这台电脑永远放行。地址取自连接本身，绝不是转发 header，所以在代理或隧道后面，每个客户端都是代理的地址：要在代理那边限制。`STARLING_ALLOW_IPS` 里写错一个条目，结果是什么都进不来，而不是什么都能进来。settings.json 读不出来、或者解析不了又没有可用备份时，上次读到的清单继续生效；如果星鸣启动后一次也没读到过，那么在文件修好之前只有这台电脑能进。`starling web` 的页面也遵守同一份清单。

**HTTPS。** 一台通过互联网访问到的星鸣（比如当中继用的服务器）可以自己提供 HTTPS，前面不需要代理：

```sh
starling serve --tls-cert /etc/letsencrypt/live/relay.example.com/fullchain.pem \
               --tls-key  /etc/letsencrypt/live/relay.example.com/privkey.pem
```

`STARLING_TLS_CERT` 和 `STARLING_TLS_KEY` 做同样的事（给 Docker 用），`starling web` 和 `starling daemon install` 也接同样的开关；服务在不带它们的情况下重新安装时会保留这一对证书，而 `--no-tls` 会把它去掉。这些文件是 PEM：证书后面跟着它的证书链，以及它的私钥。缺其中一个，或者文件加载不了，就拒绝启动，而不是改用普通 HTTP。任何一个文件变化时证书都会重新读取，所以续签（certbot、acme.sh）在下一次连接时就生效，不用重启；那时如果文件加载不了（一次续签只写了一半），就沿用之前那一份，日志会说明。其它电脑访问网关、以及 `starling web` 的页面，只能走 `https://`，而星鸣打印的网络链接也写 `https://`；从它们发来的普通 HTTP 会得到 400。这台电脑仍然可以在同一个端口上用 `http://127.0.0.1` 调用，就像这里接线的 agent、MCP server 和状态检查那样，因为服务器名字对应的证书覆盖不到回环。通过 NAT 地址调用的容器或 WSL 发行版算是另一台电脑，走 `https://`。网关 key、共享和允许地址与在 HTTP 上时一样工作。

Caddy 是另一个选择，而且证书也由它替你搞定：开着共享、每个客户端一把网关 key，运行 `relay.example.com { reverse_proxy 127.0.0.1:3425 }`。本机上的代理会加转发 header，所以它的请求按前面说的算另一台电脑的。

**用量 → 概览 → 网关 key**按客户端的 key 把调用分组，绝不按 provider 的凭据分组。**用量 → 请求**提供同样的过滤器；CSV 里包含 `caller_key_id` 和 `caller_key_name`。被删掉的 key 保留它们的历史。通常那个本地 `starling` token 和更早的记录保持不归属。调用方归属涵盖对话、图片、视频创建和 System One。provider 归属仍然独立；CSV 里 `provider_key_*` 排在 `caller_key_*` 前面。一个已有的局域网 key 会变成 **Starling**，凭据不变。`lanKey` 仍留在设置里，供旧版星鸣读取。禁用或删除默认 key 会把那面镜子换成一个随机的非空吊销值；轮换不会把它重新启用。把共享关掉再打开，既不会重新启用一个被禁用的默认 key，也不会让一个被删掉的 key 回来；只有去申请一把新的共享 key 才会造一把。key 仓库会记下迁移已完成，哪怕 `lanKeyId` 存不下去，这样读取就不会一直重试那次写入。迁移写入失败会被记下，但不会妨碍网关启动、CLI 的 key 管理，或者设置页和 key 列表页打开。如果设置是只读的，改默认 key 会失败且不改动它：把 `settings.json` 变成可写再重试，免得旧版本继续接受它原来的凭据。独立的具名 key 仍然可以管理。网关凭据留在 `~/.config/starling/caller-keys.json`（感知 XDG，权限 `0600`），绝不出现在用量记录或清单响应里。

百度千帆的 [Token Plans](https://cloud.baidu.com/doc/qianfan/s/Dmrabu8b6) 以 `baidu-qianfan` 的形式提供：个人版（个人版）和企业版（企业版）两个套餐，加上按量付费，每一个都有自己的 Chat Completions、Responses 和 Anthropic Messages endpoint，以及一把只在自己那个套餐上生效的 key。用 `starling provider add baidu-qianfan <api-key>` 添加——它第一天用的 id `qianfan-token-plan` 也会被占用。这些套餐不提供模型清单，所以预设里带着它们有文档的模型；按量付费在 `/v2/models` 提供它自己的。

阿里的 Token Plan 在两处销售，各有各的 `sk-sp-` key，且只有它自己的主机认：在 Qwen AI 平台上叫 `qwen-token-plan`（`token-plan.maas.qianwenaiapi.com`），在阿里云百炼上叫 `bailian-token-plan`（`token-plan.cn-beijing.maas.aliyuncs.com`）。两者都在 `/compatible-mode/v1` 提供 Chat Completions，在 `/apps/anthropic` 提供 Anthropic Messages，并带着这些套餐有文档的文本模型，以备主机一个都不列的时候。百炼那个还会在 System One 的 `/compatible-mode/v1/systemone` 上用同一把 key 应答它的决策模型 `decision-model-preview`，所以路由组可以挑它当分类器。同一个模型在 Qwen AI 平台的按量付费里叫 `qwen-decision`（`maas.qianwenaiapi.com`），在百炼的工作空间里叫 `bailian-decision`。两个套餐都没有查额度的 API：额度在控制台的订阅页面里看，所以星鸣不为这些 key 显示套餐用量。

### Google Vertex AI

<a id="google-vertex-ai"></a>

Google Vertex AI（`google-vertex`，谷歌的文档现在称它为 Gemini Enterprise Agent Platform）用你自己的 Google Cloud 项目提供谷歌的 Gemini 模型，费用记在那个项目上。它不要 API key：每个请求都用一个由你的 Google 凭据签出的 token 来签名。先在项目里打开 Agent Platform API（`aiplatform.googleapis.com`）（项目上必须开着结算），然后用 `gcloud auth application-default login` 登录，或者准备一个服务账号的 key 文件（[谷歌的指南](https://docs.cloud.google.com/gemini-enterprise-agent-platform/models/start/gcp-auth)）。[星鸣到哪里找这些凭据，以及它读哪些种类](subsystems/providers-accounts.md#responsibilities-and-sources-of-truth)。

```sh
starling provider add google-vertex project=my-project
starling provider set google-vertex location=us-central1      # global (the default), us, eu or a region
starling provider set google-vertex credentials=~/keys/vertex.json
starling provider set google-vertex impersonate=vertex@my-project.iam.gserviceaccount.com
starling provider set google-vertex impersonate=              # empty clears it
starling provider test google-vertex
```

`project=` 是项目的 id。`location=` 是 `global`（默认值，空值也会回到它）、`us`、`eu` 或某个区域（比如 `us-central1`）。`credentials=` 指定一个凭据文件，用来代替 gcloud 的登录来签名，比如服务账号的 key；相对路径会按命令所在的目录补成完整路径。`impersonate=` 让请求以那个服务账号的身份发出；登录的账号需要在它身上有 Service Account Token Creator 角色。应用的编辑器在 **Project ID**、**Location**、**Credentials file** 和 **Service account** 里问同样的东西；TUI 的 `a` 只问项目，其余由 CLI 设置。key 会被拒绝，地址（`url=` 之类）也一样：项目和位置决定了它的 token 唯一会发往的那个地址。添加、列出和显示它都不会问谷歌任何东西；凭据是在给请求签名时才读的。

每个位置提供它自己的模型，而 Vertex AI 没有一张清单可以问，所以星鸣给每个位置配的是它在 2026-10-06 应答过的那些模型（以及 `gemini-4-argon`，从 2026-10-13 起在 `global` 上）：

| 位置 | 模型 |
| --- | --- |
| `global` | `gemini-3.8-flash`、`gemini-3.7-flash`、`gemini-3.6-flash`、`gemini-3.5-flash`、`gemini-3.5-flash-lite`、`gemini-4-argon`、`gemini-3.1-pro-preview`、`gemini-3.1-pro-preview-customtools`、`gemini-3.1-flash-lite`、`gemini-3-flash-preview`、`gemini-2.5-pro`、`gemini-2.5-flash`、`gemini-2.5-flash-lite` |
| `us`、`eu` | `gemini-3.8-flash`、`gemini-3.7-flash`、`gemini-3.6-flash`、`gemini-3.5-flash`、`gemini-3.5-flash-lite`、`gemini-3.1-flash-lite` |
| 一个区域 | `gemini-2.5-pro`、`gemini-2.5-flash`、`gemini-2.5-flash-lite` |

某个区域如果多提供一个模型（比如 asia-northeast1 提供 `gemini-3.5-flash`），就把它和你想要的其余模型一起敲进去：

```sh
starling provider models google-vertex gemini-2.5-pro gemini-2.5-flash gemini-2.5-flash-lite gemini-3.5-flash
starling provider models google-vertex all   # back to the location's list
```

`starling provider set google-vertex location=…` 会保留已经挑好的模型，所以搬过之后要再挑一次。

谷歌的 [Priority PayGo](https://docs.cloud.google.com/gemini-enterprise-agent-platform/models/priority-paygo) 和 [Flex PayGo](https://docs.cloud.google.com/gemini-enterprise-agent-platform/models/flex-paygo) 用请求头来指定。编辑器的 **Headers** 里两者都提供，可以添加：`X-Vertex-AI-LLM-Request-Type` 和 `X-Vertex-AI-LLM-Shared-Request-Type`；它们的取值就是谷歌那些页面给的。在 CLI 上是：`starling provider set google-vertex header.<Name>=<value>`。[关于 Priority PayGo 观察到的行为](subsystems/providers-accounts.md#constraints-and-failure-behavior)。

请求失败时：

- `no Google credentials at <file>`：那个位置上没有凭据文件。如果它就是 provider 自己那个，就把 `credentials=` 改对，或者只写 `credentials=` 把它清掉；否则运行 `gcloud auth application-default login`，或者设置 `credentials=`。
- `GOOGLE_APPLICATION_CREDENTIALS names <file>, which isn't there`，或者 `GOOGLE_APPLICATION_CREDENTIALS is "…", not a full path`（`CLOUDSDK_CONFIG` 也一样）：这个变量指名的文件星鸣读不了，而 gcloud 自己那份也不会替它来读。给出完整路径（开头是 `~/` 就是家目录）、把它取消设置，或者设置 `credentials=`。
- `the Google credentials in <file>: the sign-in was refused`，并附带重新登录的建议：登录过期了或被吊销了。运行那条建议为该文件指名的命令：gcloud 的 Application Default Credentials 用 `gcloud auth application-default login`（文件里冒充了某个服务账号时加上 `--impersonate-service-account=`），gcloud 的 `legacy_credentials` 下的文件用 `gcloud auth login <account>`。你自己的文件就重新写一份，或者用 `credentials=` 给另一个。下一次请求就会读新文件，不用重启。
- 一个点名 `aiplatform.endpoints.predict` 的 403 `PERMISSION_DENIED`：签名的那个账号，或者它冒充的服务账号，可能没有权限在该项目里使用 Vertex AI。给它在项目上授予 Agent Platform User（`roles/aiplatform.user`）（[这个角色](https://docs.cloud.google.com/gemini-enterprise-agent-platform/machine-learning/general/access-control#aiplatform.user)）。
- `impersonating …: the sign-in was refused (403)`：登录的账号需要在那个服务账号上有 Service Account Token Creator（`roles/iam.serviceAccountTokenCreator`）。

### Plugin

<a id="plugins"></a>

星鸣自己不签入的某个订阅，可以来自一个 [OpenCode](https://opencode.ai) provider plugin 或一个 [pi](https://github.com/earendil-works/pi) 包：OpenCode 用户为了签入某个套餐而安装的那些 npm 包（它们的 `auth` 钩子），以及注册 provider 的那些 pi 包（`pi.registerProvider`），在星鸣里和在那边一样工作。星鸣用 [Bun](https://bun.sh) 跑它们，插件第一次需要时才下载；插件负责签入、列出模型并发出每个请求；星鸣像对待任何 provider 那样把它们提供给 agent。

```sh
starling plugin add opencode-gemini-auth # an npm package, or a path to a plugin of your own
starling plugin add pi-antigravity       # a pi package, the same way
starling plugin                         # the plugins, what each signs in to, and whether you are
starling plugin login google-plugin     # its sign-in: the method, its questions, the browser or a key
starling plugin logout google-plugin
starling plugin off opencode-gemini-auth # on brings it back; rm removes it; update updates them all
```

星鸣已经有的 provider id（google、openai、anthropic）会变成 `<id>-plugin`。在应用里，设置 → Plugins 负责添加和移除它们，而它们签入的那些 provider 在「添加 provider → 来自 plugin」里。

有时两个装好的 plugin 会签入同一个 provider id，比如一个第三方 plugin 和你自己那份副本。其中一个服务那个 id：你挑的那个，或者最后加上的那个。两个 plugin 都会留在列表里，Plugins › Installed 里每一行都写明是哪一个在服务这个 provider。另一行有 **Use for <provider>**，点它就把 provider 切过去。`starling plugin` 打印的也是同一件事，而

```sh
starling plugin use ~/dev/my-acme-auth acme # this plugin serves acme from now on
```

就完成了这次挑选。每个 plugin 仍然可以单独关掉或删除。

一个包是不是 pi 的，看它的 `package.json` 里有没有 `pi` 清单、有没有 `pi-package` 关键字，或者它是否依赖 `@earendil-works/pi-coding-agent`。星鸣把 pi 装在旁边，并用 pi 自己的加载器加载它；它注册的 provider 就是星鸣使用的那些，而它的命令、工具和渲染器都不去动。它的签入按 pi 里的方式运行：OAuth 登录的提问和页面，以及它显示的对话框和选择器（`ctx.ui`，还有它的终端组件），都在星鸣的签入流程里问；用 API key 也行。pi 自己的文件（它的 `auth.json` 以及一个包保存的东西）在星鸣配置目录下的 `pi/` 里，而不是你的 `~/.pi`，除非 `PI_CODING_AGENT_DIR` 另有说法。请求按 pi 发送的样子到达这个包，它的回话保留自己的思考和工具调用。

#### 给写 plugin 的人

<a id="for-plugin-authors"></a>

plugin 就是个 OpenCode plugin 或 pi 包；不需要任何星鸣特有的东西。对 OpenCode plugin，星鸣会多读几个字段，而 OpenCode 会忽略它们：

- **provider 的图标**：`auth` 钩子上的 `icon`，或者 plugin 的 `package.json` 里的 `"starling": { "icon": "…" }`（对它签入的每一个没有指名的 provider 都算）。一个 `https://` 的图片 URL，指向某个公开主机上的图片，星鸣取一次并保存；或者一个 `data:image/…` URI；PNG、JPEG、GIF、WebP、ICO 或 SVG，最大 1 MB。别的都被忽略，改用 plugin 市场为这个 plugin 列出的图标。`package.json` 里的图标同时也是 Discover → Unofficial 下那张卡片上的图片，在 GitHub 上则对应打了 `starling-plugin` 标签的仓库。
- **API key 的输入框**：`type: "api"` 方法的 `label` 是那个 key 输入框的标题，就像 OpenCode 的对话框那样（只写「API key」的会读作「<provider> API key」），它的 `placeholder` 是输入框里的提示（也是 `starling plugin login` 里那条提问后面的提示）。方法的 `prompts` 会先被问出来，和在 OpenCode 里一样，然后送进 `authorize(inputs)`。`type: "password"` 的提问（星鸣自己加的；OpenCode 有 `text` 和 `select`）是不回显输入的：在应用里是一个遮蔽的输入框，在 `starling plugin login` 里也不回显（#1982）。

```js
export const LemonPlugin = async () => ({
  auth: {
    provider: "lemon",
    icon: "https://lemon.example/icon.png", // or "data:image/svg+xml;base64,…"
    methods: [
      { type: "api", label: "Lemon API key (lemon.example/keys)", placeholder: "sk-lemon-…" },
    ],
  },
})
```

在 TypeScript 里，`icon` 和 `placeholder` 不在 OpenCode 的类型中：把这个钩子建成一个变量（或者做一次类型断言），或者把图标放进 `package.json`。

#### 网关 middleware

<a id="gateway-middleware"></a>

plugin 也可以充当网关 middleware：`onRequest(body, ctx)` 看到 agent 发出的每个请求（在路由之前），`onProvider(body, ctx)` 看到它在每个被尝试的 provider 上发出的请求体（`ctx.provider` 里有它的 `id`、`name`、`account` 和 `model`；每次尝试都从 `onRequest` 留下的内容开始），`onEvent(event, ctx)` 看到流式回话的每个事件，`onResponse(body, ctx)` 看到整个回话，全部都在 agent 自己的 API 里。钩子返回改过的对象，或者什么都不返回表示不改；`onEvent` 返回 `null` 表示丢掉这个事件，而 `ctx.reject(status, message)` 会把请求挡回去。在 `package.json` 里用 `"starling": { "middleware": "./mw.js" }` 声明它，或者放一个单独的 `*.middleware.js` 文件：

```js
// alias.middleware.js — starling plugin add ./alias.middleware.js
export const events = ["message_start"]
export function onRequest(body, ctx) {
  ctx.state.asked = body.model
  if (body.model === "fast") body.model = "deepseek/deepseek-chat"
  return body
}
export function onEvent(ev, ctx) {
  ev.message.model = ctx.state.asked // the agent sees the name it asked for
  return ev
}
```

Middleware 在网关内部用 [moejs](https://github.com/Calcium-Ion/moejs) 运行，这是一个用 Go 写的 JavaScript 引擎（每个流式事件大约一微秒），而不是在 Bun 上：没有计时器、`fetch` 或 Node API，只能 import 旁边的文件。抛异常或耗时过久的 hook（一个请求 250 ms，一个事件 50 ms）会让它拿到的东西保持原样。设置 → Plugins 会显示每个 middleware 的调用次数、耗时和失败情况。剩下的内容见 [plugin 指南](subsystems/plugin-runtime.md)。

Middleware 通过内置 plugin `middleware`（Plugins › Built-in 里的「Gateway middleware」）运行，只要有一个 middleware 开着它就保持开启。关掉之后，没有任何 middleware 运行，agent 的请求按原样走。隐私脱敏是内置 plugin `redact`（Privacy masking），只要设置 → 隐私要求任何脱敏它就开着；如果要求了脱敏而它没有在服务，请求会得到 503，什么也不发出去。

`ctx.options` 是 `plugins.json` 里 middleware 那一条的 `options`：用 Plugins › Installed 里的 **Options** 按钮设置，或者

```sh
starling plugin options model-map '{"mapping": {"fast": "deepseek/deepseek-chat"}}'
starling plugin options model-map off # clears them
```

名字可以是包的短名（`@magpie-community/middleware-model-map` 就是 `model-map`）。一个包的 `"starling": { "options": {…} }` 是编辑器起步时的示例。

现成的 middleware 在 Plugins › Discover › Gateway middleware 下（`@magpie-community/middleware-<name>`），多半是 New API 的渠道设置、只是用了同样的 JSON：`param-override`（`param_override`）、`model-map`（`model_mapping`）、`system-prompt`、`word-guard`（敏感词）和 `think-tags`（去掉 `<think>…</think>`，或 `thinking_to_content`），以及 `antigravity-cloak`，它把 Antigravity 用来挡掉 Claude Code 系统提示的那些词，从它单独发给 Antigravity 的内容里去掉。middleware 在社区 `registry.json` 里的那一条有 `"kind": "middleware"`，也没有 `providers`。

#### Agent plugin

<a id="agent-plugins"></a>

plugin 可以增加一个星鸣自己并没有接入的 agent。它的 `package.json` 指名那个模块（`"starling": { "agent": "./agent.js" }`），或者它就是一个单独的 `*.agent.js` 文件：

```js
export const agent = {
  id: "aider", name: "Aider", bin: "aider",
  config: "~/.aider.conf.yml",   // json, yaml or env
  model: "model",                // the key of the model
  prefix: "openai/",             // written before starling's model id
  ua: ["aider", "litellm"],
}
export function connect({ gateway, model, models }) {
  return { "openai-api-base": gateway.v1, "openai-api-key": gateway.key }
}
```

之后它就和星鸣自己的那些一样出现在 Agents 页面上：挑一个星鸣的模型会写入模型和 `connect` 返回的东西，而选它自己的模型或断开连接会把每个键放回原样。它和 middleware 一样跑在 moejs 里。被星鸣自己的 agent 用着的 id 会被拒绝。在市场里，agent 的那一条有 `"kind": "agent"`（`@magpie-community/agent-<name>`）。完整参考在本仓库的 plugin 文档里（docs/subsystems/plugin-runtime.md）的 Agent plugin 一节。

### 一个模型花多少钱

<a id="what-a-model-costs"></a>

一次调用按它的**有效价格**计入：如果你为那个 provider 和模型设了价格，就用你设的；否则用那个 provider 自己的目录列的；再否则用 models.dev 为这个模型的厂商列的。你只按地址添加、没指定目录的 provider，它自己的价格取 models.dev 在那个地址上列的那一个（它的 `api`），前提是恰好只有一个，而且它不是一个套餐的：位于 `https://developer.amd.com.cn/radeon/api/v1` 的 provider，价格取自 models.dev 的 `amd`，不管你叫它什么。开箱即用的情况下，最后这一条就是全部故事，而对任何不按目录价收费的 provider 来说它是错的数字——一个打折转售或按倍数转售的中继，会按这个模型的厂商收的价来计。写下一个 provider 实际收多少，用量账本和会话汇总就会用它：

```sh
starling model price relay-a/gpt-5.5                      # what it is counted at, and where that came from
starling model price relay-a/gpt-5.5 0.12,0.60,0.01,0.15 # input,output,cache read,cache write
starling model price relay-a/gpt-5.5 --reset              # take your price off this model
starling model prices                                     # every model you priced
```

这四个数字都是每百万 token 的美元价。四个都要填，因为少填一个会让每次调用的其余部分都被低估；`0` 表示这个模型不花钱地提供服务，那是一个价格，而不是没有价格。小数可以用点，或者在应用的输入框里用逗号（`0,25`）。

对某个没有已知价格的模型的调用，会按回话声称作答的那个模型在同一 provider 上的价格来计，只要那个模型有价格：一个把 `kimi-k3` 当作 `moonshot-kimi-k3` 卖的中继，回话里写的是 `kimi-k3`，所以它的调用按 `relay-a/kimi-k3` 的价格计。一个自己就有价格的模型（包括 `0`），不管回话怎么说，都保留自己的价格。

`starling model price provider/model` 会说明一个价格是从哪来的；对没有价格的模型，它也给出能给它一个价格的那两条命令。

完全没有价格的调用会被排除在某个合计的成本之外，绝不会被当成免费：用量页面的合计旁边会写它没能定价多少次调用（"≈$1.20+"，接着是 "3 calls unpriced"；按模型定价的 Sessions 写的是多少个模型），而 `starling usage` 和终端应用的用量页面在各自合计之后说同样的话。星鸣不会从一个相似的名字去猜某个变体的价格（`-auto`、`-max`、`-fast` 各自可能价格不同）；用 `starling model price` 给它一个，或者在 provider 的编辑器里那个模型下面填。

**Kimi Code** 会员的那些模型，按各自对应的 Kimi API 模型计，而不是按 models.dev 为这个套餐列的 $0 来计：`k3` 和 `k3-256k` 按 `kimi-k3` 计，`kimi-for-coding-highspeed` 按 `kimi-k2.7-code-highspeed` 计。`kimi-for-coding` 是 K2.8 Preview，API 不卖它，所以在你给它一个价格之前它没有价格（`starling model price kimi-code/kimi-for-coding …`）。

**第五个数字**是 1 小时缓存写入的价格。Anthropic 对保留 5 分钟的缓存写入按输入价的 1.25× 计费，对保留一小时的按输入价的 2× 计费，而它的用量会说明哪些是哪一种（`cache_creation.ephemeral_5m_input_tokens` / `ephemeral_1h_input_tokens`，来自 API 和 Claude Code 的记录）。第四个数字是 5 分钟那档的价格；第五个没给时，对 Claude 模型就是输入价的 2×——Anthropic 的规则。其它模型除非你给了，否则没有 1 小时的价格：它报出的 1 小时写入按 5 分钟那档的价格计，而 `starling model price` 和应用的输入框都不会为它显示 1 小时价格。在星鸣开始保留这个拆分之前记录的调用，所有写入都按 5 分钟那档的价格计，就像当时一样。

**长上下文价格**是指请求的提示词超过某个大小之后，整个请求要付多少钱，就像 OpenAI 对超过 272K 输入的 gpt-6-astra、gpt-6.1-sol 和 gpt-6-luna 那样计费（输入和缓存 2×，输出 1.5×）：

```sh
starling model price openai/gpt-6-astra 10,50,1,12.5 --tier 272k 20,75,2,25
```

这个大小按 OpenAI 的算法算提示词：未缓存输入 + 缓存读取 + 缓存写入，而档位在它**超过**那个大小时才适用（272,000 不算，272,001 算）。这时调用的每一部分——输入、输出、缓存——都按该档位的价格，而不只是超出那个大小的那些 token。`--tier` 可以再给一次，换一个大小。内置价格只在 models.dev 列出了档位的地方才带档位（`cost.tiers` 里类型为 `context` 的），那三个模型就是如此。会话和按天的合计汇总很多次调用，所以它们按基础价格计；账本和每一次请求则按它们达到的那个档位计。在应用里，一个模型的价格行下面有一行安静的 **Long-context price**，列表里的档位会灰着显示出来，没有档位时则是一个链接。

查找价格的顺序是：**这个模型的价格 → `<provider id>/*` 的价格，它覆盖那个 provider 的每个模型 → 那个 provider 自己的目录列的 → models.dev 为厂商列的 → models.dev 为「你说它等同于」的那个模型的厂商列的。** 最后一条是给那些没有任何目录定价的 provider 自己的 id 用的，比如 WorkBuddy 的 `kimi-k3-1` 或某个 plugin 的 `qfmodel`（#2344）：在 provider 的模型编辑器里给这个模型设上 **Same as**（settings.json 里的 `"modelSameAs": {"workbuddy/kimi-k3-1": "kimi-k3"}`），它的调用就按 Kimi K3 的目录价计，用这个 provider 的价格率，绝不会用别的 provider 的资费。没有 Same as 的模型绝不会靠名字去匹配另一个。`starling model price <provider/model>` 会说明一个价格是不是来自它。`--reset` 会移除第一条，并在 `<provider id>/*` 的价格仍然生效时把这一点说出来；把那个按名字也重置掉就能一并去掉。

价格是**一个 provider 对一个模型的资费**，不是模型自己的：同一个模型经过两个 provider 就是两个价格，各保留各的。agent 能看到的东西没有任何变化。模型清单、agent 自己的设置，以及为后台任务挑选模型的那些选择器——一张图片、一次联网搜索、一段描述——都照旧从目录里工作；只有成本报告会读那个有效价格。

有两件事值得知道。账本和会话汇总是在被读取时重新定价，所以新增或改动一个价格会把先前的数字重述一遍：它们是按有效价格估的数，不是结清的账单。而且价格是按 provider 和模型定的。现在的记录会标出上游 API key，但一个按 key 收不同资费的 provider，仍然没法用一个 provider 级别的价格精确算出成本。

### 一个模型要什么

<a id="what-a-model-takes"></a>

一个提供了 models.dev 没有列、或列错了大小的模型的 provider，它的上下文窗口和回话上限是星鸣无从得知的。说清它们是多少：

```sh
starling model context relay-a/gpt-5.5 262144  # the window a request may hold, or 1m
starling model output  relay-a/gpt-5.5 131072  # the most a reply may hold, or 128k
starling model context "relay-a/*" 200000      # every model of that provider
starling model output  relay-a/gpt-5.5 --reset # take your limit off this model
```

`*` 要加引号，因为 zsh 会把一个展开不了的名称当成一条失败的命令，而不是像 bash 那样把名称传下去。

两者都按这个顺序查找：**这个模型 → 这个 provider 的 `*` → provider 自己的清单 → models.dev。** `--reset` 只移除这个模型自己那个值。

你让它不列出、或关掉的 provider，和其它 provider 一样取这两个值：数字会保留下来，等它重新开始服务时，它的模型就按这些数字来。

上下文窗口既是给 agent 看的**数字，也是一个路由输入**：到窗口的 95% 时，一个卡在某个路由组成员上的请求会转到一个能吃下更多的成员，所以把窗口说大会让这次转移动得太晚。回话上限会在 `/models` 里声明，而一个组声明的是它里面最小的那个。网关不会因为 agent 发了某个上限就把它降低。一个没有指明上限、又发往 Anthropic Messages API（那个 API 需要一个上限，比如 Codex 的 Responses 轮次、Chat 客户端）的请求，会在 agent 被告知的那个窗口内，把它作为 `max_tokens` 发出去。只有上限谁都不知道的模型才按 16384 走。一个撞上它的回话以 `stop_reason: max_tokens` 结束，Responses 客户端把它读作 `incomplete_details.reason: "max_output_tokens"`（#1792）。

保存一个窗口或一个回话上限，会把那个数字写进星鸣在各个 agent 自己的文件里保存的模型清单——Pi 的 `contextWindow` 和 `maxTokens`、OpenCode 的 `limit`，以及 Crush 的、droid 的、Cline 的、Qoder 的和 Zcode 的——agent 在启动时读它。所以一个已经在跑的会话会保持它开始时的那个窗口，而网关自己的 `/models` 以及从那之后的每个请求则立刻就是对的。

### 厂商怎么称呼一个模型

<a id="the-name-a-vendor-knows-a-model-by"></a>

中继常常用它自己的 id 提供某个模型——一个用来做命名空间的前缀、一个带日期的名字、一个和 models.dev 的叫法不同的 `-latest`。说清该用哪个名字去要：

```sh
starling model wire relay-b/model-2                         # what the vendor is asked for
starling model wire relay-b/model-2 vendor-c/model-2-preview # ask for it by this
starling model wire 'relay-b/*' 'vendor-c/*'                # every model, * being the model
starling model wire relay-b/model-2 --reset                 # ask for it by its own name again
starling model wires                                        # every name your vendors are asked for models by
```

带 `*` 的参数要加引号：zsh 会把裸的 `*` 当成 glob，并回答 `no matches found`。在名字里，`*` 代表模型本身，所以 `'relay-b/*' 'vendor-c/*'` 会为 `model-3` 发 `vendor-c/model-3`，为 `model-2` 发 `vendor-c/model-2`——给一个给模型加了命名空间的中继一个名字，而每个模型仍然按它自己的名字去要。不带 `*` 的名字会把那个 key 的每个模型都用这一个名字去发，这只有在那个中继确实把它们一视同仁地提供时才对。模型自己的 key 比 provider 的更优先，而 `--reset` 只拿走它被给到的那个，所以在 `'relay-b/*'` 生效时重置 `relay-b/model-2`，provider 的那个名字仍然生效——CLI 在每次改动之后都会说这两个里哪个生效，也会指出那些被「整个 provider 的一个名字」留给自身的模型；而对一个从未给过的名字做 `--reset`，它会照实说明，而不是勾掉一次什么都没去掉的删除。

只有发出去的那个请求会带上这个名字，它出现在 chat、Responses 或 Anthropic Messages 请求的 model 字段里，也出现在 Anthropic 的 token 计数里；图片请求仍然用星鸣认识这个模型的那个名字。Gemini CLI 和 Antigravity 的登录走的是 Code Assist，在那里，一个 effort 挑中的模型是星鸣所认识的那个模型的一个变体——`gemini-3.7-flash` 在 `high` 下会作为 `gemini-3.7-flash-high` 发出——所以名字里的 `*` 指的就是那个变体：`'antigravity/*' 'vendor-c/*'` 会分别按各自的名字去要 `vendor-c/gemini-3.7-flash-high` 和 `vendor-c/gemini-3.7-flash-low`，和别处一样。不带 `*` 的名字，或者给 `antigravity/gemini-3.7-flash` 本身起的名字，在每一层都是那一个名字。

其它一切都保留星鸣认识这个模型的那个名字：agent 挑选用的目录、路由组的成员、`GET /models`，以及一次调用被记录和被定价时用的名字。厂商自己的回话声称作答的是什么，则并排保存在账本的 `served_model` 里——那正是它的用处——并且它会跟当时向厂商要的名字做比较，所以一个用中继自己的 id 作答的模型不会被读成换了模型。这个比较用的是读账本时生效的那些名字，因为一条记录保存的是厂商作答的东西，而不是请求发出时用的名字：事后给一个模型改名会重新评判那次调用，而它随后就被读成了换模型。某个 provider **支持**什么——它接不接受 temperature、哪些推理档位合适——仍然是按星鸣所认识的那个模型来问的，所以上游改名不会改变星鸣对这个模型的行为。

上游名字是**某一个 provider 的**，不是模型自己的：另一个提供同样 id 的 provider，会用它自己的名字或者这个名字去要它。它被赋予的那个模型必须是那个 provider 提供的——`<provider>/*` 就是「每个模型都这样」的说法——因为给一个星鸣绝不会向该 provider 去要的模型起名字，那不算它自己的名字：这个模型会按星鸣认识它的名字去要，而且任何地方都不会说给它起的那个名字不是生效的那个。一个既用自己 id 提供某个模型、**又**把它列出来的中继，确实会得到这个名字；被拒绝的是 provider 的清单里根本没有的模型。模型测试也用这个名字，所以一个只认自己 id 的中继不会把一个好用的模型报成坏的。图片模型是遵循上面那条规则的例外：它按星鸣认识它的名字来测试，在 images API 上，以及在它回退到的那个 chat 上，因为画图就是这么去要的。

### 路由组

<a id="routing-groups"></a>

路由组是若干个模型（来自一个 provider 或多个），agent 把它们当作一个来挑：`group/<id>`。网关会在每个成员的 key 和账号上一起路由每一次请求。两个 provider 用同一个名字提供的模型会自己变成一个组；应用里的路由视图和 `starling group` 可以组出别的：

```sh
starling groups                           # yours, then those starling found
starling group add "Opus anywhere" models=claude/claude-opus-5-5,copilot/claude-opus-5.5 routing=order stays=session
starling group opus-anywhere              # one group, its models in order
starling group set opus-anywhere models+=openrouter/anthropic/claude-opus-5.5 routing=usage
starling group set opus-anywhere models-=copilot/claude-opus-5.5
starling group set opus-anywhere routing=weight weight=claude/claude-opus-5-5=3,copilot/claude-opus-5.5=1
starling group copy opus-anywhere "Opus test"  # the same models, routing and rules, to change on its own
starling group rm opus-anywhere           # one starling found is hidden; starling group restore <id> brings it back
starling claude group/opus-anywhere       # use it
```

`routing=` 可以是 `smart`（默认：在那些还有余额的订阅里，先用最早重置的那个）、`order`（先用第一个模型，直到它答不了再换下一个）、`rotate`（每一轮轮到下一个成员）、`usage`（先用得最少的）或 `pace`（按周节奏：在重置之前，优先用「每小时剩余本周额度最多」的那个账号，这样重置时损失的周更少），或者 `weight`（每个模型按自己的权重分摊请求，平滑交错：`weight=claude/claude-opus-5-5=3` 表示它每拿到三个请求，权重为 1 的模型才拿到一个；权重取值 1 到 1000，单独写 `weight=` 会清空全部，而组里的另一个组算一个成员，`weight=group/fast=2`）。一把 key 没有可衡量的额度，所以它按自己的顺序，或者按星鸣最近发给它的量来走——但如果是它的主人给了 5 小时、按天或 7 天限额的 sub2api key：当 provider 的 Balance URL 设成中继的 `/v1/usage` 时，它的卡片会显示那些窗口，路由也像对待订阅那样按它们来衡量它，在中继说出 `api key 7天限额已用完` 之后让它休息——或者直到它的窗口下一次被读到还没满、它的限额被提高、或者它的用量被重置。按套餐买的 key——Kimi Code、GLM Coding Plan 或 Z.ai、MiniMax Coding Plan、OpenCode Go、Command Code、Ollama Cloud——同样按它那个套餐的窗口来衡量。只有一把 key 或一个账号时，路由页面仍然会显示它还剩多少。`stays=` 是一段对话留在应答它的那把 key 或那个账号上多久：`auto`（默认，只要厂商那边的缓存还值得留着）、`session`、`turn`、`off` 或 `strict`（绝不换到别的 key 或账号，哪怕已经超过用量上限：它的错误直接交给 agent）。`models=` 会按顺序替换整份清单；只有一个 provider 提供某个模型时，裸的模型 id 就够用。

`listed=no`（组编辑器里的 **Model lists: The group alone**）会把一个组点名的模型、以及它的模式匹配到的那些模型，挡在 `starling models`、`/v1/models` 和每个 agent 的清单之外，这样客户端只提供这个组。它们仍然会被服务：这个组在它们之上路由，而一个以 `provider/model` 点名其中某个的请求会被应答。里面的一个组保留它自己的设置，被关掉的组则什么都不挡。`listed=yes` 会把它们重新列出来。要把整个 provider 留给路由组，请改用它那个 **Only through routing groups** 开关。

一个组也可以只由决策模型组成（Jev 之类：TypeSafe 的、百炼的 `decision-model-preview`、某个 plugin 标记的模型）——绝不与能对话的模型混在一起。它不会提供给 agent；它是在 `POST /v1/systemone` 处用 `"model":"group/<id>"`（或者它的 id 或名字）来问的，而一个对话组可以拿它当分类器。它每个模型的 key 和账号都按这个组的路由顺序去问：某个像对话请求那样失败的（一个 429、一个 5xx、一个连不上的主机、一把用尽的 key）会休息，然后问下一个，而路由页面的 Requests 列表会显示每一次尝试。它不接受模式、规则、分类器、effort 或快速模式。

```sh
starling group add Jevs models=typesafe/jev-latest,bailian-token-plan/decision-model-preview routing=order
curl -s localhost:3425/v1/systemone -H 'Authorization: Bearer starling' \
  -d '{"model":"group/jevs","state":{"message":"fix the login bug"},"questions":{"intent":{"type":"choice","choices":["bug","feature"]}}}'
```

路由页面的 Requests 列表默认是按时间排序的「按请求」视图。选「按会话」就会按 agent 的会话 ID 把调用分组；这个选择会被页面记住，重新加载也还在。Codex 的标题助手如果带明确的父级或 fork 来源，会并入它原本的那段对话，保留它的标题徽标，并计入那段对话的成本。没有来源关系的标题，以及普通的分叉对话，仍各自独立。Codex 的对话名来自它的本地名称索引，并跟随改名。其它 agent 的会话（Claude Code 以及 Sessions 页面会读的那些）采用 Sessions 页面为它们显示的名字，这些名字是从它们自己在这台电脑上的文件里读出来的。未知或只有远端才有的名字会退回用 ID；完整 ID 仍然保留在标题的提示里。展开一个会话就能看到每一次请求。每次请求和每个会话都显示按有效模型价格估算的成本，包含缓存读写。会话合计只覆盖列出来的那些请求（实时追踪或所选那天保留下来的历史），一个 `+` 标记表示这是部分估算。没有会话 ID 的调用单独列出；没有 token 档位的老历史，或者没有已知价格的模型，显示 `—`。

打开 Providers 页面**不会**再扫描本机其它应用里的 provider 配置：星鸣版不检查 Alma、CC Switch、Claude Code 或 Codex，也没有「Import local configurations」提示。供应商请手工添加，或用 `starling://import?…` 导入链接添加。

「从其它应用导入」对话框**已随订阅复用功能一并移除**。Claude Code 的 `settings.json`（设了 `CLAUDE_CONFIG_DIR` 时用它）、Codex 的 `config.toml`（设了 `CODEX_HOME` 时用它）里的 provider 都不会再被读取进星鸣。Codex 那边带内联 `experimental_bearer_token` 的自定义 `[model_providers.*]` 条目、`[model_providers.*.http_headers]` 里的 header，以及来自 `[profiles.*]` 或 `model_catalog_json` 的模型，同样不再被导入。供应商请手工添加。

OpenCode 自己的登录信息（`$XDG_DATA_HOME/opencode` 下的 `auth.json`，没设时是 `~/.local/share/opencode`，Windows 上也一样）**也不再被读取**。OpenCode Go、OpenCode Zen 以及各个厂商预设下的 API key，请在星鸣里手工添加。订阅账号（Claude Pro/Max、Copilot、企业服务器）请用星鸣自己的登录流程获取。

### 星鸣如何挑选账号

<a id="how-starling-picks-an-account"></a>

当一个订阅开着好几个账号（比如 Claude Code 或 Codex 保存的那些账号），或者一个路由组有好几个成员时，网关会按下面这几步为每个请求把它们排好序。路由页面会显示每个请求的结果：谁排在了最前面，以及为什么。

1. **谁能应答。** agent 当前登录的那个账号（除非它被暂停），以及其它被勾上的账号。被设成不提供这个模型的账号、或者停在用量上限上的账号，永远不会被尝试。套餐里没有列出这个模型的账号（比如排在 Plus 后面的 Free），只有在没有别的账号列出它时才会被尝试。一个需要重新登录的 Claude 账号仍留在列表里，但轮到它时会被跳过，而且不会因此休息。
2. **路由方式。** 剩下的按那个订阅或组所用的路由方式排序（见下文）。
3. **被限流的排最后**，而且只在 *Sink* 打开时如此：一个还有余额却被 429 限流的账号，会排在所有此后没有被限流的账号后面。
4. **provider 的回退模型**排在它自己的账号之后。
5. **在休息的排最后。** 失败后正在休息的账号排到最末尾，在所有回退之后。它永远不会被丢掉：只有它一个时，照样会尝试它。
6. **对话所在的账号优先。** 一个账号一旦应答过一段对话，就会按照 *Stays* 的说法，被挪到这段对话接下来几次请求的最前面。在自动模式下，这是在一轮之内；而跨轮则要满足：上一个请求从厂商的缓存里读过、又写回了至少 1,024 个 token，并且是最近 5 分钟内应答的。第一轮不读任何东西，只写入整个提示词，所以第二轮会留下来读它；某个把提示词缓存住却不报告的厂商（OpenAI 的自动前缀缓存、Gemini）被视为写入了它收到的东西。这绝不会把正在休息的账号带回来，也不会带回来那个已经用掉某个窗口 98% 的账号（顺序模式下是 100%）。*Stays: Strict* 会带回来，而且不问别的账号：这段对话留在它那个账号上，无论它是在休息、已用尽还是已过上限，它的错误直接交给 agent。

所以路由只对还没人应答过的对话决定谁先上。一个已经用到 90% 的账号，在新对话里会排在其它账号后面，但已经在它上面的一段对话会一直待到 98%。

**这些路由方式。** Smart、Least used first 和 Weekly pace 都按「算上这个模型的最满的那个窗口」的已用份额来算，5 小时的窗口也算在内。Smart 和 Weekly pace 在某个窗口达到 90% 或以上时把一个账号排到其它账号后面，在某个窗口达到 98% 或以上时再排到那些账号后面，每一组内部按已用份额排序。

- **Smart**（默认）：在那些低于 90% 的账号里，最早重置额度的那个先上，因为它剩下的东西会在重置时作废。由最大的窗口决定：先是周，只有当几周在同一小时重置时才看 5 小时。重置时间按钟表比到小时：13:55 和 14:05 是不同的小时，14:05 和 14:55 是同一个。没有周的账号（Claude Enterprise）按它的 5 小时来算。在窗口重置之前就耗尽的额度，或者根本没有重置的额度（WorkBuddy 的那些套餐，各自在自己那天结束），算作在耗尽时重置：在几个 WorkBuddy 账号里，额度剩余最早结束的那个先上。Free 套餐的每日额度不计入，因为在每个账号上都是同一个午夜结束。星鸣还一无所知的 Claude 账号会先上一次，好从应答里弄清它还剩多少；其它未知账号排在那些已知账号之后。既有一周又有一个月的账号（Factory）按它的月来算：当这个月里剩下的比它那几周加起来还少时（按某位用户的观察，Factory 大约是两周的量），一周剩下的东西不会在周重置时作废，因为这个月里还留着它可以之后花。它在路由页面上那一行写的是最满那个窗口的份额，以及那个窗口何时重置："32% used · renews in 3 days" 说的是那一周。
- **In order**：按账号自己的顺序。第一个接下每一个请求，直到它用尽（100%）或失败。
- **In turn**：一段对话的每一轮都轮到下一个账号。一段新对话会落到「最近 30 分钟里其它对话最少」的那个账号上。
- **Least used first**：先用已用份额最低的；然后是星鸣最近发得 token 最少的那个账号（发出去的一半在一小时后就不再计入）。未知的账号算作没被用过。
- **Weekly pace**：先用节奏最高的那个账号。节奏是「本周剩余份额」除以「离本周重置还有多少小时」（至少按 1 小时算）：还剩 60%、离重置 20 小时，就是每小时 3。当有两周窗口都算上这个模型时（Opus 自己的那个和通用的那个），取节奏较低的那个。没有周的账号按它的 5 小时来算。节奏落在自己那一档最高值 0.1 以内的账号算作一样，而在它们之中，星鸣最近发得 token 最少的那个先上。还不认识的 Claude 账号像在 Smart 里一样先上一次；其它未知账号算作一个全新的周。

除 Smart 之外的这些路由方式由内置 plugin `rotation` 提供（Plugins › Built-in 里的「Routing strategies」）。只要有 provider 或组用它们中的某个来路由，它就一直开着。关掉之后，选择器只提供 Smart，而仍被设成其中某个的 provider 或组会按它列出的顺序来尝试。

你自己的路由组由内置 plugin `groups` 提供（Plugins › Built-in 里的「Routing groups」），只要你有一个组它就一直开着。星鸣为两个 provider 都提供的模型自动找到的那些组不需要它。关掉之后，网关不再列出也不提供你自己的组，路由页面上也没有「新建组」按钮；`starling groups` 仍然列出它们。

**Make first** 设定这些账号的顺序。这个顺序有多重要，取决于路由方式：

| 路由方式 | 这个顺序决定什么 |
| --- | --- |
| In order | 全部：第一个接下它能接的每一个请求。 |
| In turn | 轮转接下来轮到谁。 |
| Smart | 只在那些窗口都在同一批小时里重置的账号之间，或已用份额相等的账号之间。 |
| Least used first | 只在已用份额相同、且最近发出的 token 也相同的账号之间。 |
| Weekly pace | 只在同一节奏档里、且星鸣最近发出的 token 相同的账号之间，实际上从星鸣启动以来没有这种。 |

对 Claude Code 和 Codex，*Make first* 还会把 agent 登录到那个账号上，因为 agent 登录着的那个账号排在最前。打开 *Keep … signed in to* 时，*Make first* 只改变网关的顺序，agent 仍然登录在它被留在的那个账号上。

**当一个账号用尽时。** 一个说明额度已用完的 429 会让该账号休息到它恢复，而同一个请求会立刻转给下一个候选，只要回话还没发出任何内容。休息多久取自下面这些里第一个已知的：拒绝信息里的重置时间（Claude Code 的 `usage limit reached|<time>`、ChatGPT 的 `resets_at`）、星鸣最后一次读到已用尽的某个窗口的重置时间（98%，顺序模式下是 100%）、厂商的 `Retry-After` 或限流重置响应头（最多一小时），否则 15 分钟。某个 key 的 provider 可以自己休息得更久：`starling provider quota-rest <id> 6h`，或者对一个在某个钟点回来的额度写 `until 08:00`（按厂商的时钟写就是 `until 00:00+08:00`）。它自己的休息时长压过要求超过一小时的 `Retry-After`。它永远不会超过 8 天。这个账号的窗口会立刻重新读一遍，一旦某次读到发现它填满的那个窗口又重新开始了（对 Claude 来说是一次新的 `/usage`；对任何订阅来说，是它的重置时间过去了），它就立刻恢复，而不是等到拒绝信息里说的那个时刻。五小时的窗口重新开始、但它那一周仍然用尽，并不会让它回来。一个只是短时限流的 429 会让账号休息厂商要求的时间（最多一小时），或者一分钟；如果它每次休息刚结束就又回来，这个时间就翻倍，最多 30 分钟。欠费会让它休息半小时。任何其它失败休息一分钟，每次再失败就加长一点，最多 10 分钟；一个因为窗口用尽而失败的订阅会一直休息到那个窗口重置。`starling provider rest-cap <id> 4h` 会把这两个上限（30 分钟和 10 分钟）对某个 provider 的 key 和账号一起调整，从 1 分钟（每次休息一分钟）到 8 天；`default` 把它们还原。`starling provider rest-cap <id> off` 是给那些偶尔冒出来的、原因不明的错误的中继或账号池用的：它某个失败却没有说明原因的 key 或账号（一个 5xx、一次连接掉线）完全不休息，下一个请求还会再问它。被拒绝的 key（401、403）、它不提供的模型、额度或余额用尽、或者限流，仍然按上面那样让它休息。

**当安全过滤器拒绝时。** 一个在说出任何内容之前就被厂商的安全过滤器拒绝的回话，不会让任何人休息：它关乎被问的是什么，而不是账号。默认情况下，同一个请求会问这个模型的其它账号，然后问路由组的其它成员，直到有一个应答。设置 › 模型 › 回话 › *拒绝后再问一次* 设定了还要再问多少个：1、2 或 3，或者 *永不再问*。当不再问别人时，走 Anthropic Messages API 的 agent（Claude Code）会收到厂商原样给出的那个拒绝，带上分类的 `stop_reason` `refusal`，而 Claude Code 会像在 Anthropic 的 API 上那样，切换到它自己的回退模型。Codex 和 Chat 客户端会收到一个 400，说明模型拒绝了这次请求，它们不会再问一次。任何其它 400 都不会再问。

**走网关，还是自己单独用。** 上面这些全都是网关的事，针对经由星鸣发出的请求。单独使用的 Claude Code 或 Codex 会用它登录着的那个账号直接和厂商说话，路由根本看不到。为此，星鸣会在它启动一分钟后看一次 agent 登录着的那个账号，之后每 5 分钟看一次：一旦它用掉了某个窗口的 98%（顺序模式下是 100%），或者到达了它的用量上限，星鸣就把 agent 登录到下一个被勾上的账号，按它们的顺序，挑那个额度已知且低于这个份额的。星鸣还一无所知的已保存 Claude 账号（见下文）不会被换上去。等它把 agent 换下来的那个账号所有窗口又都低于 90% 了，它就把 agent 登录回那个账号。*Keep … signed in to* 会关掉这一套：agent 留在第一个账号上，或者你挑的那个账号上，不管它还剩多少。这对网关的路由没有任何影响。

开着不止一个 Codex 账号时，星鸣还会在 Codex 的 config.toml 里设置 `openai_base_url`，这样登录着的那个账号没有额度的请求，就会通过网关转给其它账号。Providers › Codex 下的 *Only switch Codex's sign-in*（或者 `starling accounts switch-only codex on`）会把这一项排除掉：那样星鸣只按上面说的移动 Codex 自己的登录态（`auth.json`），而 Codex 的请求直接发给 OpenAI，绝不经过星鸣（#1907）。你自己写的 `openai_base_url` 会保留。Codex 应用和正在运行的 `codex app-server` 都在启动时读登录态，所以已经开着的那个会留在它那个账号上，直到再次打开；新的 `codex` 会话则用新的那个。接上星鸣时，Codex 的请求照旧经过它。

**额度多久读一次。** 路由用的是上一次读到的内容，绝不为一次读取而等待，星鸣启动后的第一次除外（最多 3 秒）。超过一分钟的旧数据会在路由请求时于后台重新读一次，而因额度失败的账号会立刻重读；如果失败时已经有一次读取在进行，账号会在那次读取回来后立刻重读。用量页面做的读取也算一次，它的年龄从做完那一刻算起。Codex 和大多数其它订阅都用这种方式从厂商读每个账号。Codex 账号还可以从 ChatGPT 为它发给星鸣的每个回话里得知，那个回话会说明这个账号用了多少：一个接近用量上限的账号从下一轮起就不再被使用（#1295）。用量上限设在账号那一行上，可以按它的每个窗口分别设；而一个窗口也可以有它自己的上限，从它的计量里取：五小时窗口 50%、按周的 40%，或者某个窗口不设上限。只要有任何一个窗口超过了它自己的上限，这个账号就被搁置。用量上限仍然只是对星鸣已读内容的约束，不是保证：一轮已经在进行的对话可能把这个账号推过它，所以 99% 的上限并不承诺还剩 1%。Claude 不一样：星鸣从不直接问 Anthropic。它只读 Claude Code 登录着的那个账号，做法是运行 Claude Code 的 `/usage`：

- 星鸣启动后一次；
- 你打开或刷新用量页面时（在应用里或 TUI 里），或者运行 `starling quota` 或 `starling accounts` 时，最多每 30 秒一次；
- 否则在 5 到 15 分钟之后再来一次（每次随机取一个），而且只有这段时间里用过 Claude Code 才来。

一个没有登录的已保存 Claude 账号永远不会被读。它还剩多少，只能从 Claude Code 以那个账号应答经由网关发出的请求时所说的话里知道，在那之前它的用量卡片也是这么说的。所以路由对这样一个账号所依据的窗口可能已经是几小时前的；一个重置时间已经过去的窗口算作又空了。

### 把已登录的 agent 当作 provider

<a id="signed-in-agents-as-providers"></a>

**只有星鸣自己签入的账号**才会被当作 provider 提供：Codex（PKCE 流程）、Copilot（设备码流程）、Google OAuth（Gemini CLI / Antigravity），以及插件提供的账号。星鸣**不读取任何别家 Agent 已登录的凭据** —— Claude Code 在 macOS 钥匙串里的 OAuth 登录、Codex 的 `~/.codex/auth.json`、Copilot 的 `~/.config/github-copilot/apps.json`、Devin 的 `~/.local/share/devin/credentials.toml` 都不会再被读，也不再被服务。Qoder 由星鸣用它的 OAuth 设备流程登录，保存在星鸣自己的配置里；Qoder CN 是它旁边的另一个订阅，给那些用阿里云账号或手机号在 qoder.cn 上开的、没法在 qoder.com 上开的账号用。

在别处登录的账号**不能再从它们的文件导入**：`starling accounts import` 及其全部入口已移除，窗口里对应的「从文件导入账号…」也没有了。账号只能通过星鸣自己的登录流程获得。

### 接入其他任何东西

<a id="connecting-anything-else"></a>

网关监听 `127.0.0.1:3425`（`STARLING_ADDR` 可以改它），并随应用启动；`starling serve` 则单独运行它。对于反向代理或容器部署，把 `STARLING_PUBLIC_URL=https://starling.example.com` 设成控制台和 CLI 里显示的那个 base URL，连接示例也用它。在允许别的机器进来时（在局域网上共享，或者 `STARLING_ADDR` 在网络上），网关页面的「连接」一开始就是它，直到你在那里另选一个地址；设置 › 关于会把它列在本地 URL 旁边，而 `starling web` 会把它显示为网关的地址。本机 agent 的配置仍然用本地网关地址。

要做一个用上星鸣、或想在 Agents 页面上有一行的应用或 agent：见[把应用或 Agent 接入星鸣网关](integrating.md)。

反向代理后面的外部客户端需要一把已启用的网关 key（网关 → 网关 key）：当设置 → 在局域网上共享开着、或者 `STARLING_ADDR` 在网络上时，星鸣会问你要一把；否则它会拒绝它们——除非这个代理自己给客户端做认证（`STARLING_TRUST_PROXY=1`，见下文）。一个不带自己端口的公开 URL——比如反向代理的 `https://starling.example.com`——也是 `starling web` 为它自己的页面打印的地址，所以这个代理必须把 `/v1` 和 `/v1beta` 转发到网关的端口，其余的转到页面的端口。当代理和星鸣跑在同一台机器上时，它经回环转发过来的请求带着转发 header，因此也需要一把网关 key，和从网络上来的请求一样（见[Provider 与网关](#providers-and-the-gateway)）；一个自己给客户端做认证的代理可以用 `STARLING_TRUST_PROXY=1` 来信任。

它暴露这些：

| 路径 | API |
| ------------------------ | -------------------------- |
| `/v1/chat/completions`   | OpenAI chat completions    |
| `/v1/responses`          | OpenAI Responses（HTTP，以 SSE 流式返回；WebSocket 升级请求会得到 426） |
| `/v1/messages`           | Anthropic Messages         |
| `/v1/messages/count_tokens` | Anthropic token 计数 |
| `/v1beta/models/{model}:generateContent` | Google Gemini（还有 `:streamGenerateContent`、`:countTokens`） |
| `/v1/audio/transcriptions`、`/v1/audio/translations` | OpenAI 语音转文字（multipart，原样透传） |
| `/v1/models`、`/v1beta/models` | 模型目录 |
| `/v1/codex/models`       | 按 Codex 的 `model_catalog_url` 读取的模型目录（[另一台电脑上的 Codex](#providers-and-the-gateway)） |
| `/api/version`、`/api/tags`、`/api/show` | 按 Ollama 服务器列出的模型目录，给那些只把自定义 endpoint 当作 Ollama 的客户端用（见下文） |

每个 `/v1/models` 条目都包含 `reasoning` 和 `supported_reasoning_levels`（`[{"effort":"low"}, ...]`），以及同样那些档位以纯名字形式出现在 `reasoning_levels`（`["low", ...]`，没有时是 `[]`），读它们的是采用 Catwalk schema 的客户端（Crush 及其同类）（#2576）。只要有一个成员会推理，一个路由组就标为 `reasoning`；`supported_reasoning_levels` 列出那些会推理的成员共有的档位。已知不会推理的成员不会把它们拿走（#2127），档位谁都不知道的成员也不会（#597）。当这个组路由到一个星鸣确知不会推理的成员时，不会向它发送推理请求，于是这个组被要求的 effort 就能到达那些接受它的成员，而不会让厂商以 400 拒掉；谁都不代表的那个成员则照旧收到 agent 要的东西。`native_endpoints`（`["/v1/messages"]`）点名那些「对这个模型的请求会原样透传」的 API；路由组里不写它，对那种反正所有请求都会被翻译的模型也不写。

只把自定义 endpoint 当作 Ollama 服务器的客户端（Visual Studio 和 VS Code 的 Copilot：Manage Models，然后 Ollama）会拿到 `http://127.0.0.1:3425` 作为 Ollama URL，不带 `/v1`。它从 `/api/tags` 读模型，从 `/api/show` 读每个模型的工具、视觉、思考和窗口，在 `/v1/chat/completions` 上对话。Ollama 自己的对话接口（`/api/chat`、`/api/generate`，也就是 `ollama run`）不提供。Copilot 不发 key，所以这只在这台电脑上有效；在只有一个公开 URL 的反向代理后面，`/api` 会转到 `starling web` 的页面，所以要给出网关自己的端口。

当厂商说的正是 agent 的那个 API 时，请求会原样透传，否则会被翻译，流式输出、工具调用和推理都一并处理。key 是 `starling`（任何值都行；网关只监听回环地址），模型写成 `provider/model`。

- 路由组可以用它的 id（`group/<id>`）或它的名字来问。
- 裸的模型 id（`glm-5.3`）也可以。当好几个 provider 都提供它时，它会走星鸣为它找到的那个路由组，那个组按 Providers 页面上的顺序去问它们，失败就换下一个。关掉「自动找到的组」时，它只走那个顺序里的第一个。
- 语音转文字（`/v1/audio/transcriptions`、`/v1/audio/translations`）接受 OpenAI 的 multipart 表单，它的 `model` 填一个星鸣的 id，比如 `openai/whisper-1`、`groq/whisper-large-v3`，或某个自定义 provider 自己的（一个 vLLM 的 `qwen3-asr`）。这个表单会原样发到那个 provider 的 OpenAI 风格 base 上，音频一个字节都不改；只有 model 被改成该 provider 自己的 id，而 `prompt` 会按你的脱敏设置遮住。答案以厂商自己的格式回来（`json`、`text`、`srt`、`vtt`）。没有 OpenAI 风格 API 的 provider 会被拒绝；没有那个 endpoint 的 provider 会自己返回它的 404。文本转语音不提供。
- 每个回话都在 `X-Starling-Provider`（那个 provider）和 `X-Starling-Model`（`provider/model`）里写明是谁应答的，用量日志也一样。

任何带 base URL 设置的工具都能用它：

| 工具说的 | Base URL | 环境变量 |
| ----------- | -------------------------- | --------------------------------------------- |
| OpenAI      | `http://127.0.0.1:3425/v1` | `OPENAI_BASE_URL`、`OPENAI_API_KEY=starling`      |
| Anthropic   | `http://127.0.0.1:3425`    | `ANTHROPIC_BASE_URL`、`ANTHROPIC_API_KEY=starling` |
| Gemini      | `http://127.0.0.1:3425`    | `GOOGLE_GEMINI_BASE_URL`、`GEMINI_API_KEY=starling` |

可选的 `X-Starling-Account: <account>` 响应头（账号的邮箱或登录名，或它在路由页面上的 id）会把一个请求钉到某个有多个账号的订阅中的那一个上：只有它会被尝试，而未知的账号、套餐里没有这个模型的账号、或者正在休息的账号，会得到一个错误，而不是另一个账号的回话。这个响应头不会被转发给厂商。

状态栏可以在第一个 token 到达之前就显示一轮对话去了哪里：发请求时带上 `X-Starling-Session: <id>`（agent 自己的会话头也可以，比如 Pi 的或 Claude Code 的），然后读 `GET /v1/starling/route?session=<id>`。它给出这个会话最近一次请求在路由上目前的状态——`asked`（agent 点名的模型）、`group`、`rule`（匹配上的那个组的规则，以及 `rule.pick`，它的决策模型挑中的 effort）、`model` 和 `effort`（正在尝试的那个成员，写成 `provider/model`，以及它被发出去时的推理档位），以及 `tries`，每试一个成员一条，其中失败的每一个都带着它的 `fail` 标为回退——然后在回话结束后是 `done`、`status` 和 `served`；会话还没有路由时 `route` 是 `null`。路由一旦定下来就会出现，在问厂商之前。`after=<seq>&wait=<seconds>`（最多 60）会把回答挂住，直到路由的变化超过上一次那个 `seq`，这样 UI 可以一次一个请求地跟着走一轮对话。只有被点名的那个会话会被告知；和 `/v1/starling/quotas` 一样，它回答这台电脑，而别的机器要用局域网上共享的那个网关的 key。一把受预算、受某些模型或受某些账号限制的 key，只会被告知它自己那些请求的路由：别的 key 的会话得到 `route: null`。

要在没有会话的情况下显示当前在用的那个，`GET /v1/starling/quotas`（以及 `starling quota --json`，同一份清单）在最近 30 天里经由网关应答过请求的每个订阅账号、套餐和 key 上都有 `lastServedAt`，最新那个还有 `last: true`。它保存在 `providers.json` 旁边的 `served.json` 里，所以重启后仍在。

那里每个订阅窗口还有用量页面的「整个窗口」估算（#2333）：`holds`（`tokens`、`cost`（按 API 目录价）、`priced`），它是根据星鸣自己从窗口开始以来经由该账号发出的调用，按已用份额算出来的。在没法如实说明的情况下，`holdsWhy` 会改为说明原因：`low`（用得不足 5%）、`clock`（没有重置时钟）、`log`（这个窗口开始得比星鸣的日志还早）、`none`（里面没有路由过任何东西）或 `credits`（里面某次调用可能是用额度付的）。两者都没有的窗口，就是不会给出整体数字的窗口：一个号池、一个家庭、一个金额，或者一张余额在别的电脑上读到的卡。

`GET /v1/starling/quotas?kept=1` 用星鸣上次读到的内容给出同一份清单，不问任何厂商，也绝不等待一次读取，适合那些经常来问的东西。Claude Code 的状态行就建立在它之上（#1997）：Claude Code 在每条消息上运行它，把会话的 JSON 从 stdin 送进来，它打印一行——这个会话最近一轮去了哪个模型（来自 `/v1/starling/route`），然后是每个订阅和套餐以及它们各个窗口的使用量，最后应答的那个带 `*`：
`→ claude/claude-opus-5-5 | claude a@example.com* 5h 62% 7d 41% | claude B 5h 8% 7d 30%`。
最近 30 天里经由网关应答过的账号都会显示，一个都没有时全部显示；`starling statusline claude`（一个 provider，或者某个账号的邮箱、登录名或名字）只保留那些。它不管 Claude Code 是怎么登录网关的都有效，不像 `anthropic-ratelimit-*` 那些响应头——Claude Code 只有在登录 claude.ai 时才读它们。星鸣不把它写进 Claude Code 的设置（那里可能有另一个状态行）；请自己把它加到 `~/.claude/settings.json`：
`"statusLine": {"type": "command", "command": "starling statusline"}`。网关没在跑时，它打印一个空行。

`starling heap [<file>]` 是用来写一份「星鸣占了太多内存」的报告的：对应用做一次 macOS 的 `sample`，Go 的栈帧会显示成 `???`。它向那个正在跑、并在服务网关的星鸣（应用、菜单栏、`starling serve` 或 `starling web`）请求 `GET /v1/starling/heap`——一份在一次回收之后取的 Go 堆 profile（`runtime/pprof` 的 `heap`，gzip 过的 protobuf，`go tool pprof` 能读）——并把它写到 `<file>`，没给文件就写到工作目录下的 `starling-heap-<date>-<time>.pb.gz`，绝不会覆盖那里已有的文件。它打印这个文件，以及网关在 `X-Starling-Heap-Stats` 里发来的运行时数字：进程、`HeapAlloc`（以及回收之后还活着的部分）、`HeapInuse`、`HeapSys`、`StackSys`、`Sys`（Go 从系统拿走的全部；远超这个数的占用是在 Go 之外的，比如 webview）、`NumGC` 和 goroutine 数量。这个路由只回答这台电脑上以 `localhost` 或某个回环地址调用它的程序：别的机器、某个代理或隧道的客户端、任何带着浏览器标记的请求（一个 Origin，无论是否在白名单里、`null`、`Sec-Fetch-*`、一个 `Mozilla/` 的 user agent）以及任何其它 `Host`，都被 403 拒绝，带着网关 key 也一样。当 `STARLING_ADDR` 设在一个不是回环的地址上时，这次调用就不是从回环来的，所以同样被拒绝。

`starling cpu [<file>] [--seconds N]` 是用来写一份「星鸣占了太多 CPU」的报告的（X 上的 @LeiCheng42 就遇到过，当时在跑 Codex）。它向同一个正在运行的星鸣请求 `GET /v1/starling/cpu?seconds=N`——一份记录 30 秒的 Go CPU profile，或者 N 取 5 到 120（`runtime/pprof` 的 CPU profile，gzip 过的 protobuf，`go tool pprof` 能读）。在它记录期间，照常使用星鸣，哪怕它正忙着，比如在跑 Codex：网关会继续服务其它每一次调用，而这份 profile 记录的就是那段时间。它把 profile 写到 `<file>`，没给文件就写到工作目录下的 `starling-cpu-<date>-<time>.pb.gz`，并且在开始记录之前就拒绝那里已有的文件，而不是之后。它打印这个文件，以及网关在 `X-Starling-CPU-Stats` 里发来的数字：进程、在记录的这些秒里用掉的 CPU 时间（用户态和系统态，按系统的算法，并带上占一个核心的比例）以及从启动到现在用掉的、它的核心数和 `GOMAXPROCS`，还有 goroutine 数量。Go 一次只记录一份 CPU profile，所以在一份正在记录时再跑一次 `starling cpu` 会被 409 拒绝；调用者挂断的那一份则停止记录。这个路由回答的调用者和 `/v1/starling/heap` 一样，拒绝的那些也一样是 403。

一把受预算、受某些模型或受某些账号限制的网关 key，不会被告知星鸣这些账号中的任何一个（#1981）：`/v1/starling/quotas`、`/v1/starling/quotas/history`、`/v1/starling/quotas/cards`、`POST /v1/starling/quotas/refresh` 和 `/v1/starling/concurrency` 都会给它 403 和 `"code": "gateway_key_held"`，无论来自这台电脑还是别的机器。它从 `GET /v1/starling/limit` 读自己的限额。另一个用到这样一把 key 的星鸣只显示那把 key 自己的卡片。

它的持有者可以允许这样一把 key 被告知它可以使用的那些东西的额度（#2216）：key 那一行上的 **Quotas** 徽章（鼠标悬停可见；打开时它写着*Quotas shown*），或者 `starling gateway-key quotas <id> on`（再用 `off` 关掉，或不带参数看当前是哪种）。这四条额度路由随后会告诉它那些「每个模型它都能用」的 provider 的卡片——一把被限制在某个 provider 的一部分模型上的 key，不会被告知那个 provider 的任何东西，因为它不能用那些模型花的是同一份额度——而且只告诉其中它可以用的那些账号和 key，包含账号名和余额。它没被告知的卡片会被拒绝刷新（403 `gateway_key_held`），而一个不指名任何卡片的刷新只重读它被告知的那些卡片。一把一张卡片都没被告知的 key 得到的 `/v1/starling/quotas/cards` 响应和之前一样是 403 `gateway_key_held`，这样另一个星鸣显示的是它自己的卡片，而不是「什么都没读到」。`/v1/starling/concurrency` 和 `/v1/starling/usage` 仍然拒绝它。

你自己的应用也能读到用量页面上的数字。`GET /v1/starling/usage?period=<today|week|month|all>` 是这个页面的汇总，和那个窗口的 `/api/usage` 是同一份 JSON（给别的机器时 `path` 会被去掉），其中 `sessions` 里是这个窗口按 token 列出的前 100 个会话；`sessionCount` 是会话总数。`GET /v1/starling/usage/requests` 一次给一页请求（`offset`、`limit` 最多 500），过滤器和 Requests 日志的一样：`period`、`day`、`agent`、`provider`、`model`、`account`、`callerKey`、`purpose`、`failed`、`retried`、`q`、`via`、`computer`。`callerKey=none` 保留那些没有任何网关 key 发出的调用，而 `by.callerKey` 按 key 把页面的调用分开，其中也包括 `none`（#2521）。一次调用里说了什么，这两处都没有。和 `/v1/starling/quotas` 一样，它们回答这台电脑，别的机器要用网关 key。一把受预算、受某些模型或受某些账号限制的 key 只会被告知它自己的调用：`/v1/starling/usage` 会用 403 拒绝它，而 `/v1/starling/usage/requests` 只保留它的那些行、过滤列表，以及按 key 拆分时它的那一份。

现在正在飞的东西，给像 SketchyBar 这样的状态栏用（#2132）：`GET /v1/starling/concurrency` 给出 `requests`，即网关正在处理的每一个 agent 请求，不管有没有设限额（等位置的也在其中）；`waiting`，其中有多少个在某个 provider 或账号的限额下等着位置；以及 `tools`，即经由 Claude 订阅运行的 Claude Code 轮次里那些在等调用方工具结果的（它们不是请求）。空闲时每一个都是 `0`。`data` 是每个受限 key 或账号各自的计数，没有设限额时为空。`active` 把那些请求一条条列出，最老的在前，空闲时是 `[]`：`id`、`agent`、`model` 和 `provider`（路由之后才有）、agent 发了会话时的 `session`、`started` 以及此后的 `ms`，还有只在网关知道时才有的 `phase`：`slot`（在某把 key 或某个账号的队列里等着）、`rpm`（等着这一分钟里的名额）、`upstream`（已经发到厂商，或者它的回话正在流式返回）。里面没有提示词、回话或 key。这次读取不问任何厂商，也不计入自身，所以可以每秒轮询一次。它像 `/v1/starling/quotas` 那样作答：回答这台电脑，别的机器要用网关 key，而受限的 key 一律不答。

一个自己直接调用厂商、而不经过星鸣的应用，可以把这些调用上报，好让用量页面把它们算进去：`POST /v1/starling/usage/requests`，带 `Content-Type: application/json`。请求体是一条记录，或者最多 5000 条的数组（最多 32 MiB），形状和 `usage.jsonl` 里的一行一样：

```sh
curl -s http://127.0.0.1:3425/v1/starling/usage/requests \
  -H 'Content-Type: application/json' \
  -d '[{"t":"2026-10-13T08:00:00Z","agent":"my-app","provider":"anthropic",
        "model":"claude-sonnet-4-6","in":1200,"out":340,"cache_read":9000,
        "cache_write":0,"ms":2400,"ttft_ms":610,"status":200,"rid":"req_01"}]'
```

- **字段：** `t`（RFC 3339）、`agent`、`provider`、`model` 和 `rid` 是必填的。`in`、`out`、`cache_read`、`cache_write`、`cache_write_1h`、`reasoning`、`ms`、`ttft_ms`、`first_text_ms`、`flow_ms`、`status`、`cost`（美元）、`effort`、`session` 和 `err` 可以不填。
- **速度：** Requests 页面根据 `out`、`ms` 和 `ttft_ms` 判断一个回话的速度（`ttft_ms` 是从请求到回话的第一段内容所用的毫秒数：文本、推理或一次工具调用）。一个推理过的回话（`reasoning` 大于 0）是从它的第一次文本开始计时的，所以只有配上 `first_text_ms`（从请求到它第一次文本的毫秒数）才能给出速度，而它可能早于 `ttft_ms`、也可能晚于 `ms`。`flow_ms` 是内容花了多久才到，从计时速度的那个点算起：从它十分之一到十分之九的那段时间，乘以 10/8（最多 `ms` × 10/8）。一个内容在它窗口的四分之一之内就到位的回话是被扣住后一下放出来的，它给不出速度。不知道的就别填。
- **定价：** 没有 `cost` 的记录按星鸣为它那个 provider 和模型的价格计价。
- **过去的时间：** `t` 就按给定的来，不管多早，这样历史可以回填；但它不能早于 2000-01-01T00:00:00Z，也不能比现在晚出一个小时以上。
- **重复：** 已经上报过的 `rid` 会被跳过，重启之后也一样，所以同一批可以再发一次。
- **拒绝：** 一条坏记录会让整次提交被 400 拒绝，并指出是哪一条，写在 `error.index` 和 `error.param` 里（`record 3: in: must not be negative`），什么也不会被写入。
- **应答：** `{"accepted":1,"skipped":0,"records":[{"index":0,"rid":"req_01","status":"accepted"}]}`。
- **显示：** 上报的行带着 `source: "reported"`，计入每一个合计，也可以用 `via=reported` 单独列出（GET 应答里的 `reported` 是它们的合计）。
- **不算进实时的东西：** 它们绝不会计入某把网关 key 的限额或预算、某个订阅窗口、某个提醒、某个限流或路由。
- **谁能上报：** 和那些 GET 一样：这台电脑，或者带着网关 key 的别的机器。网页即使在这台电脑上也需要一把 key。

某一行的 `session` 是它所属的那段对话。Claude Code 子 agent 的调用会保留它父对话的那一份，并加上 `subagent`（它的 agent id，即它那些行的 `agentId` 以及它的 `<session>/subagents/agent-<id>.jsonl` 文件），还有 `parent_agent`：启动它的那个子 agent（Claude Code 的 `x-claude-code-parent-agent-id`，或者那个子 agent 的 `.meta.json` 里的 `parentAgentId`）；主对话发起的则没有。这些字段出现之前写入的记录两者都没有。Codex 的子 agent 只靠 `kind` 区分。

`starling quota refresh <provider> [<account>]` 会再读一次那张卡片，就像用量页面上的刷新按钮那样，并单独打印它；其它厂商不会被问。账号可以用它的登录名或别名来指，也可以写成 `<provider>/<account>`。30 秒内对同一张卡片的第二次刷新不问任何人，只打印读到的内容，并在 stderr 上说明，所以状态行可以把它放在循环里跑。`--json` 像 `starling quota --json` 那样打印那张卡片，带上 `readAt`（什么时候读的）和 `error`（为什么读不到）。

`starling quota wait <provider|account>` 会一直阻塞，直到那个订阅（星鸣手上的它任何一个账号）或那一个账号又有额度——没有任何挡住它的窗口处于用尽状态——然后以 0 退出，所以一个被自己的限额挡住的长任务可以无人值守地继续：

```sh
until codex exec "…"; do starling quota wait codex || break; done
```

它自己读厂商，不管网关在不在跑，并且会在它知道的最近一次重置之后不久再读一次（每 1 到 10 分钟；读不到的额度每次被问得更少），在 stderr 上说明它在等什么、等多久。账号可以用邮箱或登录名来指，两个订阅共用同一个账号时写成 `<provider>/<account>`。`--timeout 6h` 在超时先到时以 1 退出，`--quiet` 什么都不说；不认识的名字以 2 退出，Ctrl+C 则是 130。

一个持有额度的 ChatGPT 账号，在一个用量窗口用尽之后仍会继续应答：厂商会花掉那些额度，所以任务能继续。这是默认行为。`starling quota credits <account> off`（或者该账号用量卡片上的 *Use credits*）会为那个账号关掉它：一旦星鸣最近一次（大约每分钟刷新一次）对这些窗口的读取显示有一个用尽，网关就把它当作用尽，一直扣到那个窗口重置，请求则转给其它账号、组和回退；一个都不剩时，请求会得到一个说明原因的 429——除非这个账号自己花掉了它的重置、而它那一周又已用尽，这时会先花掉一个重置。`starling quota credits` 列出那些被设成不花额度的账号，`starling quota credits <account>` 说明某一个。它只改变网关的路由，绝不改变 Codex 登录着的那个账号。账号持有的额度会在 `starling quota`、`starling accounts`、用量页面和菜单栏面板里显示在它的窗口旁边。

一个账号可以有一个你自己的名字，用来区分厂商起了同样名字的席位（比如一个团队里那些用遮蔽邮箱的 GLM 席位）：在它那个 provider 的编辑器里点账号的名字，或者用 `starling accounts alias <agent> <email> <name>`（`--clear` 把它去掉）。这个名字是星鸣的，保存在 providers.json 里；厂商或 plugin 的登录态绝不会被写进去。它显示在该账号的用量卡片上、菜单栏里、`starling accounts`、`starling quota` 以及 `GET /v1/starling/quotas`（`alias`）里。一个既以 key、又以已登录的 ZCode 账号两种方式加入的 GLM Coding Plan 席位只显示一次，在那个账号的卡片上（两张卡片会把它那份额度数两遍）；在你亲自给它起名之前，那张卡片会以那把 key 的 provider 来称呼它，并说明这一点（`seat`）。当两者被发现是同一个席位之后，任何一边读取失败都不会把 key 的那张卡片带回来：只有账号被登出，或者窗口不再一致，才会。

用量页面上的卡片和菜单栏面板的 *Allowances* 标签页用的是同一个顺序：在用量页面上拖某张卡片的图标（或者在它上面按 Alt+方向键），或者在面板那个标签页的底部点 *Arrange* 然后在那里移动各行。同一个 *Arrange* 还能把某个订阅从面板的标签页里隐藏（*Hide*，用 *Show* 放回来），标签页底部会写明藏了几个。隐藏只影响面板：路由、上限、用量页面和菜单栏自己的格子都照样有它，而菜单栏里它的那个格子仍然能打开它的卡片。两样都保存在星鸣的设置里（`usageOrder`、`panelUsageHidden`）。菜单栏自己在图标旁边显示哪些额度，是在设置里挑的。同一份清单还有 *Tokens and cost today* 和 *Tokens and cost, last 7 days*：每个 agent 的请求按用量页面 Requests 标签页那样汇总，包括 agent 自己会话文件里的那些调用，它们的全部 token（输入、输出和缓存，标着 Σ）以及按你的货币折算的成本（`trayUsages` 里的 `usage:today`、`usage:7d`）；提示会把 token 拆成输入、输出和缓存。点一下就打开那个时段的 Requests 标签页。

应用里的 *Gateway* 标签页把这些做成复制按钮和现成的代码片段（shell、curl、Python、Node），每个 API 一份，外加模型 id 清单和最近的调用；`STARLING_DEBUG=1` 会把每一次调用都记到终端。

**Claude Code** 会得到 `ANTHROPIC_BASE_URL`、`ANTHROPIC_AUTH_TOKEN` 以及 `settings.json` 的 `env` 块里的模型相关变量；挑一个原生模型（`opus`、`sonnet`……）会把它们移除，并还原原来那里的内容。`ANTHROPIC_AUTH_TOKEN` 里的一把 key 会让 Claude Code 在经由星鸣运行时退出 claude.ai：claude.ai 套餐在 `/usage` 里的限制、它的连接器、语音和 `/teleport` 都会关掉。把它的登录设成 claude.ai 就能保住那个登录：星鸣会把 token 写成空的，Claude Code 把它的 claude.ai 登录发给星鸣，而星鸣绝不转发出去（连它的 `oauth-2025-04-20` beta 也不）。这时一个拒绝星鸣自己那把 key 的 provider 会被报给 Claude Code 一个 502，而不是「你的登录被拒」，所以 Claude Code 不会登出。Remote Control 和 ultrareview 一直关着：Claude Code 只在 Anthropic 自己的地址上才有它们。这个选项出现在网关接受任何 key 的地方，所以当星鸣在网络上共享时，NAT 下的 WSL 发行版里不会提供它。

**Codex** 会得到一张 `[model_providers.starling]` 表、一个指向 `~/.codex/starling-models.json` 的 `model_catalog_json`（由目录写出，所以模型会出现在 Codex 自己的清单里），以及一个有效的 `model`/`effort`；挑一个原生模型或断开连接会移除 `model_provider` 和那份目录。`[model_providers.starling]` 这张表会留着，所以在星鸣上开起来的一个线程仍然能打开。当 `model_provider = "starling"` 没有对应的表时，Codex 干脆不加载自己的配置（"Model provider `starling` not found"）。如果别的工具留下了这种状态，星鸣会在下一次同步时把那张表写回去。登录在 ChatGPT 上时（登录字段的默认值），Codex 保留自己的 provider 和登录态，星鸣的模型则通过 `openai_base_url` 加入它的清单。这时星鸣只有在 Codex 应用还占着那个账号期间才成为 Codex 的 provider：OpenAI 已经不允许了，而它也没有额度了，除非是一个还在超额范围内的 workspace 账号。被搁置的账号应用不发出任何请求，于是星鸣顶上，等那个账号又有空间了就退回去。一个用尽 100% 但还有额度的窗口不算：Codex 保持登录，继续在那些窗口上发送。你的 ChatGPT 登录态绝不会被动。Codex 在启动时读它的模型清单，所以切换之后要重启它。

**OpenCode、Pi、Crush** 会得到一个 `starling` provider 条目和 `starling/provider/model`。

**Gemini CLI** 在 API key、Google 账号和 Vertex 之间切换 `auth`；那个 API key 写到 `~/.gemini/.env`。挑一个目录里的模型会把 `GOOGLE_GEMINI_BASE_URL` 指向网关（它说 Gemini API），把 `auth` 设成用网关 token 的 API key，并在 `settings.json` 里指明模型；挑一个原生模型会把原来的 auth 放回去。

### 导入链接

<a id="import-links"></a>

某个厂商或中继可以把一个现成的 provider 以链接的形式交给它的用户：

```
starling://import?preset=deepseek&key=sk-…
starling://import?name=Acme%20Relay&chat=https://api.acme.example/v1&anthropic=https://api.acme.example&key=sk-…&models=gpt-5.5,claude-sonnet-5
```

打开一个链接会唤起星鸣，并显示这个链接会添加什么：名字、你的提示词和 key 会发往的主机、模型。在你按 *Add* 之前什么都不会被保存。`starling import <link>` 在终端里做同样的事。

| 参数 | 含义 |
| ----------- | ------------------------------------------------------------------ |
| `preset` | 一个预设 id（`starling presets`）；用它那些 endpoint。不是 `google-vertex`，那个是在你自己的 Google Cloud 项目上问的 |
| `region` | 对于有区域的预设，指哪一个 |
| `name` | provider 的名字；没有预设时必填 |
| `id` | 它的 id；没给时由名字推导 |
| `key` | 那把 API key；没给时由用户粘贴 |
| `chat` | OpenAI Chat Completions 的 base URL（`…/v1`） |
| `responses` | OpenAI Responses 的 base URL（`…/v1`） |
| `anthropic` | Anthropic Messages 的 base URL（根地址，不带 `/v1`） |
| `gemini` | Google Gemini API 的 base URL（`…/v1beta`），应用的编辑器和 TUI 就是这样取的 |
| `models` | 要暴露的模型 id，逗号分隔 |
| `catalog` | models.dev 的 provider id，用来取模型名字和推理档位 |
| `website`、`keys` | 厂商的站点和它的 API key 页面（https） |
| `icon` | 厂商自己的一张 https 图片（PNG、JPEG、GIF、WebP、ICO、SVG，最大 1 MB）。在你确认导入之后，星鸣会把它下载一次，放进它的图标目录；没有的话就用目录里的 logo 或者一个普通标记 |

Base URL 必须是 https（明文的 http 只允许指向这台电脑或局域网）。网页和 GitHub 对自定义 scheme 的链接并不可靠，所以在那样的地方请把 `starling://import?<同样的参数>` 作为纯文本给出，让用户自己复制粘贴。参数留在 fragment 里，而浏览器从不会把 fragment 发给服务器。完整的参数说明见上面那张表。

## 安装

<a id="install"></a>

星鸣版没有官网下载页。macOS、Windows 或 Linux 的构建可以从[仓库的 releases](https://github.com/lulu58588/starling-releases/releases) 下载，或者在终端里从源码安装（Linux 上装了 WebKitGTK 4.1 就是桌面应用，否则是命令行版）：

```sh
go install github.com/lulu58588/starling@latest
```

在防火墙后面、或者访问不了 GitHub 时，`starling update` 接受 `--proxy <url>`（http、https、socks5、socks5h；或者 `STARLING_PROXY`，以及通常的 `HTTPS_PROXY` / `ALL_PROXY`）和 `--mirror <prefix>`——一个你自己挑的 GitHub 下载镜像，会放在 release 的 github.com URL 前面（或者 `STARLING_MIRROR`；`starling update mirror <prefix>|off` 为每次更新保留一个，应用程序自己的那次也算）：

```sh
starling update --proxy http://127.0.0.1:7890
starling update --mirror https://mirror.example/
```

除非你给一个，否则不会用镜像；而每次下载都要比对的 SHA-256 来自更新源本身，绝不来自镜像。

Mac 的发布版是签名并公证过的；Windows 和 Linux 的构建目前还没签名（Windows SmartScreen 可能在第一次运行前问一次）。星鸣版不内置官方更新源，所以不会自己去找新版本：把 `STARLING_UPDATE_FEED` 指向你自己的更新源之后，应用程序就会在后台下载新版本，并在你重启（菜单里的 *Restart to Update*）或退出时安装它；`starling update` 则在终端里做同样的事。每个发布版都在 [lulu58588/starling-releases](https://github.com/lulu58588/starling-releases/releases) 上。

从源码：

```sh
go install github.com/lulu58588/starling@latest
```

或者在本地构建：

```sh
make build            # ./starling with the desktop app (needs cgo + the platform webview)
make app              # macOS: starling.app, a menu bar app with no Dock icon
make cli              # terminal-only build, no cgo, cross-compiles anywhere
make release          # dist/: native app build + cli builds for every platform
make release-windows  # dist/: the Windows app, amd64 and arm64 (cross-compiles)
make release-linux    # dist/: the Linux app for this machine's arch
```

Linux 上构建应用需要 `libgtk-3-dev` 和 `libwebkit2gtk-4.1-dev`（Makefile 会加上 `gtk3` tag；直接用 `go build` 时请传 `-tags gtk3`）；Windows 用的是系统自带的 WebView2 运行时。

Termux 用的是 Android 的终端构建：go install 会把它装进 Go 的 bin 目录（GOBIN 或 GOPATH/bin），而 `make build` 会自动选它。直接用 `go build` 或 `go install` 时请传 `-tags nogui`；界面用 `starling web` 或 `starling tui`。`starling autostart on` 需要 Termux:Boot。

### 作为后台服务

<a id="as-a-background-service"></a>

`starling daemon install` 会把 `starling serve` 作为你自己的一个服务来跑，没有窗口也没有桌面会话：一台服务器，或者一台你通过 SSH 访问的机器。系统会在登录时启动它（开机启动见下文），并在它以失败告终时再把它启动起来。

```sh
starling daemon install    # write the service and start it; again, to take a new PATH or proxy
starling daemon status     # running or not, its pid, file and logs, whether the gateway answers
starling daemon restart    # after updating starling
starling daemon stop|start
starling daemon uninstall  # stop it and remove the file
starling daemon web on     # the service serves the browser page too; off stops it, no word shows it
```

这个服务默认只提供网关，除非你要求它也提供那个页面：`starling daemon install --web`（或者装好之后用 `starling daemon web on`）会运行 `starling serve --web`，由一个进程同时提供网关和 `starling web` 的页面，于是那个页面显示的是这个服务自己的网关、provider 和账号。它监听 `127.0.0.1:3430`；`--lan` 把它放到 `0.0.0.0:3430`，`--addr host:port` 放到任意位置，警告和 `starling web --lan` 的一样：任何拿到这个链接的人都能改你的设置。它的 key 只生成一次，并保存在服务文件里（作为 `STARLING_WEB_KEY`，只有你能读），所以这个链接在重启和重装之后都不变；`daemon web off` 会丢掉它，下一次 `on` 会生成一个新的。`starling daemon status` 会打印这个链接。`--no-web` 表示不带它安装。

在运行着的服务旁边跑 `starling web` 不会占走网关的端口：当服务本身提供那个页面时，它打印那个页面的链接然后退出；否则它自己的页面会使用服务的网关，并说明这一点。一个占着端口却应答很慢的星鸣（忙的服务器，或者还在启动中的那个）会被等待，而不是去争抢：新来的那个只有在端口空出来之后才提供网关。

- **Linux：** 一个 systemd 用户单元，`~/.config/systemd/user/starling.service`（若设置了 XDG_CONFIG_HOME，则在 `$XDG_CONFIG_HOME/systemd/user`），用 `systemctl --user` 运行，不需要 root。它的日志：`journalctl --user -u starling -f`。systemd 在用户第一次登录时启动他的单元、在他最后一次登出时停止它们，所以要让网关在没人登录时也能开机运行，就执行一次 `sudo loginctl enable-linger $USER`；没开时 `starling daemon status` 会说明这一点。从 `su` 或 `sudo -u` 过来的话，`systemctl --user` 够不着你的会话：请以那个用户登录，或者设置 `XDG_RUNTIME_DIR=/run/user/$(id -u)`。在 WSL 上，先把 systemd 打开（`/etc/wsl.conf` 里的 `[boot] systemd=true`）。没有 systemd 的系统会被告知这一点；请用它自己有的那个管理器来跑 `starling serve`。
- **macOS：** 一个 launch agent，`~/Library/LaunchAgents/com.starling.ai.gateway.plist`，用 `launchctl bootstrap gui/$UID` 加载。它的日志：`~/Library/Logs/starling/serve.log`。
- **Windows：** 不做成服务；`starling autostart on` 会在你登录时打开星鸣，由它提供网关。在 WSL 里面，`starling daemon` 的行为和 Linux 上一样。

这个服务会拿到运行 `starling daemon install` 时的环境：`PATH`（它在那里找各个 agent 的 CLI 和 bun）、代理相关的变量、`STARLING_*` 设置（比如 `STARLING_ADDR`，以及 `--tls-cert`/`--tls-key` 这一对，对应 `STARLING_TLS_CERT` 和 `STARLING_TLS_KEY`）、XDG 目录，以及各个 agent 自己的目录变量（`CODEX_HOME`、`CLAUDE_CONFIG_DIR`……）。它还会给这个 `PATH` 加上你终端里有、而 shell 里可能没有的东西：nvm 的默认 Node（不是它最新的那个）、其它版本管理器以及 npm 的目录，还有你登录 shell 自己的 `PATH`，这样 plugin 和它们启动的 MCP server 就能找到 `node`，不管是 nvm 懒加载还是通过 SSH。连这样都找不到 `node` 时，安装会说明这一点。这个文件只有你能读，因为代理的地址里可能带着密码。改了其中任何一项之后，请重新安装一次。它会用 `PATH` 上那个名字来运行星鸣——只要那是同一个文件（Homebrew 的链接在更新之后依然正确）——并且拒绝 `go run` 构建出来的版本。`starling autostart` 开着时它会被拒绝：登录时打开的应用同样提供网关，两者会轮流在端口上失败。服务在跑时手动打开的那个应用会使用服务的网关。

### Docker

<a id="docker"></a>

`docker build` 会做一个服务器镜像：Debian 12 slim 上的纯终端二进制，带上 bash 和 curl，这样某个厂商的 `curl … | bash` 安装脚本能跑起来，它装的那些 CLI（Cursor 的、Devin 的、Grok Build）也能找到它们链接的 glibc，plugin 用的 Bun 同理；还可以在容器里开一个终端（`docker exec -it starling bash`，或者用 NAS 面板的终端；`starling` 在它的 PATH 上）。它以 nonroot 运行，把它保存的一切都放在 `/config` 上的一个卷里：星鸣自己的文件（`/config/starling`）、各自 agent 保存位置的登录态（HOME 是 `/config/home`，所以 `~/.codex`、`~/.claude`……都在里面），以及跑 Bun plugin 的缓存（`/config/cache`，只下载一次）。用旧镜像做的卷依然能用：星鸣启动时会把这些目录加进它，只有用那个旧镜像做的登录态（当时在卷外面）需要重新做一次。

```sh
docker build -t starling .
docker run -d --name starling -p 127.0.0.1:3425:3425 -p 127.0.0.1:3430:3430 -v starling-config:/config starling
```

3425 是给 agent 用的网关。在容器内部它接受任何 key；从容器外面来的请求（主机自己发来的也算）必须带一把已启用的网关 key 当作 API key，不带就得到 401。在浏览器界面里（网关 → 网关 key）或者用 `docker exec starling starling gateway-key add "Docker client"` 建一把。共享（在局域网上共享）在这里并不需要；打开它还会额外创建一把名为 **Starling** 的 key。上面那些端口发布在主机的回环地址上；Docker 的 `-p 3425:3425` 会把网关放到主机的每一个网卡上，供其它机器访问。从旧镜像升级过来时：容器外那个发 `starling` 的 agent，或者别的什么不是已启用网关 key 的 key，现在会得到 401（"the API key sent is not an enabled starling gateway key"）；请给它配一把网关 key。
在容器里星鸣只能看到容器自己的地址（Docker 的 172.17.x），所以要设 `-e STARLING_PUBLIC_URL=http://<主机或 NAS 的地址>:3425`（在主机上发布的那个端口），这样它显示和打印出来的地址才是别的机器会用的那个；在反向代理后面，把它设成那个对外的 base URL，并遵循[上面的认证要求](#connecting-anything-else)，尤其是在代理经回环访问星鸣的时候。

要用浏览器界面，就用 `starling web --addr 0.0.0.0:3430 --no-open` 代替默认的 `serve` 来运行镜像，然后打开 `http://localhost:3430/?k=<docker logs starling 里的 key>`（设 `STARLING_WEB_KEY` 可以让一个 key 在重启之后保持不变；这样浏览器会在 400 天里保持登录）。浏览器是按主机而不是按端口保存 cookie 的，所以同一个地址上的另一个服务可能把这个登录抢走：清掉它的站点数据，或者把主机的 cookie 上限填满（这个登录是 Chrome 最后才丢的那个）。没有带这个 key 打开的页面（比如一个书签）会说明自己为什么被挡回来，并在某个字段里收下这个 key，然后回到同一个页面。在那里你添加 provider、登录订阅，并从文件导入登录态。厂商会把登录的浏览器送回 `localhost`（ChatGPT 送到 `http://localhost:1455/auth/callback?code=…`），那是你自己的机器，不是容器，所以那个页面加载不出来：把地址栏里的整个地址复制下来，粘进登录的 *Callback URL* 字段。`docker exec -it starling /starling accounts add codex` 在终端里做同样的事：打开它打印的链接，然后把浏览器最后停下的那个地址粘进去。key、登录态和 plugin 都住在那个卷里，所以重启、或者在同一个卷上起一个新容器，都会保留它们。

镜像里带一个 `HEALTHCHECK`：`starling healthcheck` 会在网关于 `STARLING_ADDR` 上应答时以 0 退出，`serve` 和 `web` 都一样，于是 `docker ps` 会把容器显示为 healthy（Compose：`depends_on: condition: service_healthy`），而镜像里不需要 curl。用 bind mount 把一个文件夹挂到 `/config` 来代替具名卷也可以；它必须能被 uid 65532 写入。

#### NAS 上跑着别的机器上的 agent

<a id="nas-with-agents-on-other-machines"></a>

即使使用它的 agent（Claude Code、Codex、OpenCode 等等）跑在别的电脑上，星鸣也可以在一台无头的 NAS 上作为共享网关运行。容器不会去发现 NAS 主机或远端机器上装的 agent：请把那些 agent 配置成使用 NAS 的地址和一把已启用的网关 key。各协议对应的 base URL 见[接入其他任何东西](#connecting-anything-else)。请把 `3425` 放在防火墙或 VPN 后面，不要在没有网关 key 认证的情况下把它暴露到公网。

这样的页面处于**网关模式**：它显示 Providers、Gateway、Routing、Usage、Plugins 和 Settings，没有 Agents、Sessions 和 Library 标签页，而 Settings 里省掉了写入本机各个 agent 的东西（模型名里的 provider）以及桌面的提醒和托盘。Codex 和 Claude Code 自己的设置（Codex 子 agent、长对话、Codex 线程标题等等）在 Agents 页面上那个 agent 展开的行里，所以网关没有这些。当 `starling web` 在它那台机器上找不到任何 agent 时（比如在容器里），或者用 `starling web --gateway` 时，它就自己处于这种模式。*设置 › 通用 › 网关模式* 为这台机器挑选自动、开启或关闭（它比前面两种都优先）；关闭会把每一个页面都带回来。应用自己的窗口永远不会处于网关模式。

对于 bind mount 的配置目录，该目录必须能被非 root 的容器用户写入（uid 65532）：

```sh
sudo chown -R 65532:65532 /volume1/docker/starling/config
```

一个 Docker Compose 项目示例是：

```yaml
services:
  starling:
    image: ghcr.io/lulu58588/starling:latest
    restart: unless-stopped
    ports:
      # 127.0.0.1: the host alone; every caller outside the container needs a gateway key.
      - "127.0.0.1:3425:3425"
      - "127.0.0.1:3430:3430"
    environment:
      STARLING_ADDR: 0.0.0.0:3425
      STARLING_PUBLIC_URL: http://<nas-address>:3425
      STARLING_WEB_KEY: <random-key>
    volumes:
      - ./config:/config
    command: [web, --addr, 0.0.0.0:3430, --no-open]
```

通过 SSH 隧道打开 `http://127.0.0.1:3430/?k=<random-key>` 来配置 provider，然后为每一台远端 agent 主机创建一把网关 key。要让远端 agent 能访问，就把网关的映射改成 `3425:3425`（或者只绑定到 NAS 的那个网卡或 VPN 地址），并重建容器。要接入互联网，请优先用 VPN，比如 Tailscale 或 WireGuard；经回环转发到星鸣的反向代理或隧道，它的客户端同样需要一把网关 key，就像从网络上来的那样。

### 开发

<a id="developing"></a>

```sh
make dev
```

它用 `-tags dev` 构建，并打开应用，UI 直接从 `internal/gui/assets` 提供：保存 `app.css`、`app.js` 或 `index.html`，窗口会自己重载。装了 `fswatch` 的话（`brew install fswatch`），改动 Go 文件也会重新构建并重启应用。开发版用自己那个网关端口（`DEV_ADDR`，默认 127.0.0.1:3426），所以你已经跑着的那个星鸣会继续为你的 agent 服务。把它指向一个临时的家目录，就能把你真正的 agent 配置挡在外面：

```sh
HOME=/tmp/starling-home XDG_CONFIG_HOME=/tmp/starling-home/.config make dev
```

`STARLING_THEME=light|dark` 强制指定配色，`STARLING_DEBUG=1` 会打印网关做的是哪些翻译。

## 使用

<a id="use"></a>

```sh
starling                        # open the app: a window plus the menu bar icon
starling tray                   # menu bar icon only (use this in your login items)
starling tui                    # the same thing, in the terminal; serves the gateway while open when no other starling does
starling web                    # the app's window in a browser (WSL, a server over SSH); --lan, --addr, --no-open, --gateway, --tls-cert/--tls-key
                                # (a new key each run; STARLING_WEB_KEY keeps one, for a page run as a service)
                                # (behind a reverse proxy, STARLING_WEB_URL=https://<the page there> prints the link through it)
starling serve                  # the gateway alone, no window; --web[=host:port] serves the page too, from the same process
                                # --tls-cert <file> --tls-key <file>: https:// for other computers (see HTTPS under sharing)
starling daemon install         # run starling serve as a service (systemd, launchd); --web, --lan, --addr serve the page too; --tls-cert/--tls-key, --no-tls
starling ls                     # list every agent and its current settings
starling claude opus            # set a model (agent names accept prefixes: cc, oc, gem …)
starling codex gpt-5.6-sol
starling codex effort high      # other fields
starling codex effort           # one field, as it is now
starling codex xhigh            # bare effort levels are recognised too
starling codex deepseek/deepseek-chat # any catalog model, through the gateway
starling claude moonshot/kimi-k2.5
starling claude haiku deepseek/deepseek-v4-flash # one tier on its own model
starling claude haiku ""        # back to the main model
starling gemini auth api-key
starling opencode anthropic/claude-sonnet-5
starling oc small anthropic/claude-haiku-4-5
starling mimo anthropic/claude-sonnet-5

starling save work              # snapshot everything as a profile
starling use work               # switch back
starling profiles
starling rm work

starling sync                   # refresh the models.dev catalog and every live model list

starling quota                  # what is left of every subscription, plan and key balance
starling quota refresh codex    # read Codex's cards again now (once per 30s)
starling quota wait codex       # block until a Codex account has allowance again
starling quota credits me@example.com off # hold a ChatGPT account at its limit instead of spending credits
```

在应用里，点任何一个值就会打开一个带过滤的清单；输入可以搜索，也可以输入一个没列出来的东西；`esc` 关闭这个面板。配置档是底部那些小标签：点一下套用，点 `×` 删除，*+ save current* 新增。窗口的 *Providers* 标签页列出你的 provider 以及每个上面挂着哪些 agent；点某一行可以改它的 key 或暴露的模型、*Test* 它，或者点某个 agent 图标，把那个 agent 指向它的一个模型。*Add provider* 把预设摆成一块块磁贴：挑一个，粘贴 key。

在*用量 → 请求*和托盘的*用量*标签页里，点某一天的柱子就能看那天的合计和明细。其它天会变灰；再点一次选中的那天就回到整个时段。换了时段，选择会被清掉。

在 macOS 上，*设置 → 偏好设置 → 会话终端*决定在*用量 → 会话*里按终端按钮时用哪个已安装的应用打开会话。这份清单包含那些注册为能打开 `.command` 文件的应用，当前系统默认的那个只列一次。没有保存过选择时，星鸣跟随那个默认值；如果默认值不是终端，就用 Terminal。*Resume* 按钮仍然会复制那条命令。

终端版里的按键：

| 按键 | 动作 |
| ---------- | ------------------------------------- |
| `↑` `↓`    | 选择 agent |
| `←` `→`    | 选择字段（model、effort、small……） |
| `↵`        | 打开选择器 |
| 输入 | 过滤；回车接受自定义的值 |
| `s`        | 把当前设置存成一个配置档 |
| `p`        | 套用或删除（`ctrl+d`）一个配置档 |
| `S`        | 同步模型目录 |
| `q`        | 退出 |

agent 在启动时读自己的配置，所以一个正在跑的会话会保持它的模型，直到你开一个新的。

### 迁移到另一台机器

<a id="moving-to-another-machine"></a>

```sh
starling backup                   # writes starling.starling-backup, asks for a passphrase twice
starling backup --no-keys ~/b.starling-backup   # the same with no API keys in it
starling restore starling.starling-backup         # on the other machine
starling restore --no-agents b.starling-backup  # providers, settings, profiles; agents left as they are
starling restore --no-library b.starling-backup # the library here left as it is
```

一份备份里有你的 provider（带它们的 key，除非 `--no-keys`）、为它们挑的图片、设置、配置档、每个 agent 的模型，以及资源库（除非 `--no-library`）：指令集、MCP server，以及带自己文件、模板和其它二进制的 skills（超过 16 MB 的文件会被排除，skills 的文件加起来超过 32 MB 后最大的那些也会被排除；恢复时那台机器已有的那种文件会保留它自己的副本）。不带 key 时，某个服务器那些看起来像 key 的环境变量和 header 会变成空的。网关凭据、它们的名字、id 和禁用状态也随设置一起加密带过去。恢复设置会把网关 key 仓库换成备份里的那一个。`--no-keys` 会把网关凭据和它们的旧版镜像排除在外；这样恢复时会保留目标机器上已有的 key。恢复资源库会替换掉原来那份——被替换的会随资源库的备份留下——并把它写进那台机器上的各个 agent。它在你的机器上加密（AES-256-GCM，密钥用 PBKDF2-SHA256 从口令派生）；没有口令，里面任何东西都读不出来。恢复会替换 id 相同的 provider，并加上其余的；其中一个没带 key 来的，会保留那里已有的 key。agent 的模型只为装在那台机器上的 agent 设置。订阅不在里面：请在每台机器上分别登录。它们挑的模型和设置会跟着来，登录后就会显示出来；恢复过程、以及另一台机器上同步的状态，会点名还有哪些没登录。从 npm 或 git 添加的 plugin 也会跟着来，并装在它被放回的地方（从文件夹添加的那个留在它自己那台机器上；不带 key 时，名字看起来像 key 的选项会被排除）；plugin 的登录态不会带过去，已经在那里的 plugin 保持原样，而在一台机器上被删掉的 plugin 在其它机器上仍然留着。用管道传进去时，口令是 stdin 的第一行。

### 让多台机器保持同步

<a id="keeping-machines-in-sync"></a>

*设置 → 同步与备份 → WebDAV、S3 或 GitHub 同步*把同一份备份放在服务器上，并让每台机器都与它保持最新，网关运行期间每 3 分钟一次。挑其中一种：

- **WebDAV**：一个 WebDAV 服务器上的文件夹，比如坚果云、Nextcloud 或群晖。
- **S3**：AWS S3、Cloudflare R2、Backblaze B2、MinIO、Garage、NAS 或任何其它兼容 S3 的服务器上的一个存储桶。
- **GitHub**：一个已有的仓库，用一把带 Contents 读写权限的细粒度 token。仓库必须是私有的：同步会拒绝公开的仓库，因为它的历史会让任何人拿到每一个封存的版本。地址里可以带一个文件夹，比如 `github://owner/repo/starling-sync`。每次更新都是往那个分支的一次提交，所以这个仓库也提供了加密备份的历史。

这个文件在你的机器上用口令封好，所以服务器存的永远只有密文。每台机器只覆盖它读到的那一版（一次条件写入），所以中间被别的机器做过的更新会被合并，而不会丢失。

```sh
starling webdav on https://dav.jianguoyun.com/dav/ user=me@example.com
starling s3 on s3://my-bucket/starling endpoint=https://<account>.r2.cloudflarestorage.com access-key-id=…
starling s3 on s3://backups endpoint=http://nas.local:9000 path-style=yes access-key-id=…
starling s3                       # where it syncs to and how the last sync went; starling s3 now, off
starling github on github://owner/private-backups/settings branch=main
starling github                   # where it syncs to and how the last sync went; starling github now, off
```

关于 S3：

- `endpoint` 对 AWS 是空的。
- `region` 默认是 `us-east-1`，在 R2 上默认是 `auto`。
- `path-style=yes` 把桶名放进路径里，MinIO 和大多数自己搭的服务器都需要这样。
- 那个密钥像 WebDAV 密码一样被问出来并保存。它只用在它被给予的那组 endpoint 和 access key 上。
- 这个存储桶必须已经存在。
- 这把 access key 需要能读写 `<prefix>/starling/`。在 AWS 上它还需要能列出这个桶。
- 没有条件写入的服务器也支持。在那里星鸣会在每次写入之前检查那个对象的 ETag。两台在同一时刻写入的电脑可能都通过这个检查（阿里云 OSS，#1776），而较晚的那次写入会被保留。另一台电脑在几分钟后的下一次同步时才会发现。它发现的文件不是自己写的那一个，于是它从自己写入之前的状态重新合并：它改动的东西会和另一台电脑的改动并排放回去。它绝不会把对方的文件当成已经见过自己的。
- 共享的用量只有在一份完整的 S3 清单之后才会对账。续传 token 缺失或重复，或者列了 100 页之后清单仍被截断，都会报一个同步错误，并把之前下载过的那些用量天原样保留。

关于 GitHub：

- 在启用同步之前先创建仓库。Branch 留空就用它的默认分支。对一个空仓库，第一次加密备份会在那个默认分支上创建第一个提交。选了默认分支以外的某个分支时，那个分支必须已经存在。
- 给一把细粒度 token 这个仓库的访问权，带 Contents 读写权限。它由 CLI 问出来，绝不在命令行里接收，并且和资源库的那个 GitHub token 分开保存。
- Folder 是可选的。文件放在 `<folder>/starling/` 里；每次有变化的备份都会创建一次提交。没有变化的设置不会创建备份提交。
- 在每台电脑上用同一个加密口令。它必须和那个 GitHub token 不同。token 和同步的连接设置留在各自的电脑上。
- 文件上限是 64 MiB。仓库的历史会保留加密的旧版本，开启共享时也包括用量文件。
- 在 WebDAV、S3 和 GitHub 之间切换时，会保留其它服务的设置和凭据，方便再切回去；同一时间只有一个服务在同步。

## OTLP 导出

<a id="otlp-export"></a>

设置 → 可观测性可以把网关请求的元数据通过 OTLP/HTTP（JSON）导出。导出默认关闭。先设好采集器的 base URL 和可选的响应头，再打开 **OTLP 导出**。**导出指标**是另外单独关闭的；采集器接受时长和 token 直方图时才打开它。保存的设置不需要重启。这个导出由内置 plugin `otel` 提供（Plugins › Built-in 里的「Observability」），只要设置或环境变量要求导出它就保持开启。

对 `starling serve`，环境变量会覆盖保存下来的偏好：

```sh
STARLING_OTEL_ENABLED=true STARLING_OTEL_ENDPOINT=http://localhost:4318 starling serve
```

- `STARLING_OTEL_ENABLED`：`true` 或 `false`；只给一个 endpoint 并不会开启导出。
- `STARLING_OTEL_ENDPOINT`：一个 HTTP(S) 的 base URL；`/v1/traces` 和 `/v1/metrics` 会被附加在后面。
- `STARLING_OTEL_HEADERS`：逗号分隔的 `name=value` 对，比如 `Authorization=Bearer%20token`。值里的空格和逗号要做百分号编码。
- `STARLING_OTEL_METRICS`：`true` 或 `false`，默认关闭。
- `STARLING_OTEL_SESSIONS`：`true` 或 `false`，默认关闭；从新记录的会话事件里追踪受支持的本地 agent 交互。
- `STARLING_OTEL_BODIES`：`true` 或 `false`，默认关闭；把每次调用的请求和回话作为该 trace 的 Langfuse 输入与输出发送。
- `STARLING_OTEL_BODIES_WHOLE`：`true` 或 `false`，默认关闭；开启正文时保留完整正文，而不是在 256 KB 处截断。较长的回话会写进一个临时文件；正文对采集器来说仍然太大时照样被拒绝。

对 Langfuse，用 `https://<your-langfuse-host>/api/public/otel` 作为 base URL，用 `Authorization=Basic%20<base64(public-key:secret-key)>` 作为响应头。指标保持关闭。这用的是 Langfuse 的 OTLP 摄入端点。Langfuse v4 还要加上响应头 `x-langfuse-ingestion-version=4` 以启用实时摄入。

Trace 里包含 agent、provider、模型、token 数（含缓存和推理）、HTTP 状态、耗时和路由 ID。路由 ID 相同的几次尝试共用一个 trace ID。Chat、Responses、Anthropic 和 Gemini 的网关请求还会导出一个父请求 span，以及每一次路由尝试的子 span，包括不计费的失败、重试和回退，这样 Langfuse 就能画出一张瀑布图。一个有效的、来自调用方的 W3C version-00 `traceparent` 会把这些 span 接到调用方那条 trace 上。尝试的 span 带着 token 用量；父 span 不重复它。尝试的耗时包含等待并发位置的任何时间，而路由和重试的延迟则作为父 span 内部的空隙保留下来。调用方内部的工具执行不在网关这条 trace 里。

模型调用的 span 还会通过 `langfuse.observation.cost_details` 把估算的美元输入成本、输出成本和总成本发给 Langfuse，用的是和星鸣用量账本相同的有效价格（先自定义价格，再目录价格）。输入成本包含缓存读取和写入；推理 token 已经计入输出成本。明确为零的价格也会导出。不知道的价格会被省略，这样 Langfuse 可以用它自己的模型定价。父 span 和工具 span 不带任何成本，避免重复计数。按目录价算的订阅成本是等价于 API 的估算，不是订阅账单。

打开 **Trace agent conversations** 可以改为每次用户交互导出一条 trace，把模型调用和工具执行归在一个 agent 根之下。会话 ID 把这些 trace 归成 Langfuse 的会话。在一个可见的本地存储覆盖了那个确切的 agent/会话之前，网关 trace 仍然可用。对 Codex 和 Pi，一个可读的会话头就足以表明已就绪，不必等模型或工具的 span。第一次请求在需要时会检查一个上限 256 KiB 的头（包括 Codex 的基础指令）；之后的轮询会在文件仍属于读取器最新的 200 个主文件之列时，让空闲的会话保持就绪。Claude 的子进程有单独的 200 文件配额，共用同一个 8 MiB 的轮询读取预算。头部检查从不上传历史或正文。发现在各个请求 ID 之间共享两秒。刷新发现的那个请求会等这次扫描；其它并发的、尚未解决的请求则保留网关 trace，不等待。轮询按路径、大小和 mtime 复用头部。其它适配器仍然使用近期的观测（五分钟）。只有不带网关 key 的回环请求才能被去重。客户端原生的会话 ID 优先于星鸣的路由覆盖；未知的会话保留网关 trace，其中包括 WSL 镜像客户端和 Docker Desktop 客户端那些在本地看不见的会话。没有会话 ID 的请求会退回使用那个 agent 近期的观测。Claude Code/Cowork 的读取器还会跟随 `<session>/subagents/*.jsonl`，以及某个工作流（ultracode）的 `<session>/subagents/workflows/<run>/agent-*.jsonl`（不是它的 `journal.jsonl`）。子交互与父会话共用会话，但使用各自的 trace/span ID，即使复制了同一个用户 UUID 也一样。子进程的最终回话会跨内容块合并，并在记录两秒不变之后导出；不需要 `turn_duration`。它们的开始时间仍然是推断出来的；父工具链接不会被猜测。因此主线程的调用可以去重。Claude 那些不带工具的小请求（`max_tokens` 最多 4096），包括标题和 haiku 的助手，保留网关 trace，因为它们可能不会进入记录。这条保守规则也可能保留一个短的、不带工具的主请求。被明确归类的辅助调用，在其它 agent 上同样保留网关 trace。这包括带明确 kind 的 Codex 调用（subagent、review、memgen、title 或 compact）：即使子进程有一份可读的 rollout，这些也可能出现两次。只有它那些未归类的调用才使用本地会话去重；单凭一份 rollout 并不能证明每个辅助请求都被记录过。带 key 的和不在回环上的客户端始终保留网关 trace。星鸣每两秒从本地会话存储里读一次新事件；开启时它不会上传已完成的历史。重启或改了目的地会开启一个观测窗口。Span ID 在反复记录之间保持稳定。

Pi 和 Codex 已经用真实客户端端到端测试过。其它适配器只有格式 fixture 覆盖，**没有做过端到端测试**。

受支持的客户端和格式：

| 客户端 | 本地存储 | 计时 |
| --- | --- | --- |
| Codex | JSONL rollout（`token_usage_record`、`item_completed`、`response_item`） | 记录下来的操作和配对的原生工具；推断出的模型/原生工具区间 |
| Pi | Version-3 JSONL；可选的 `timing-final` | 记录下来的模型时间；推断出的工具区间 |
| Oh My Pi | 兼容 Pi 的 JSONL，含 `model_usage` | Pi 的计时；辅助调用的时长是推断出来的零 |
| Claude Code / Cowork | `projects/*/*.jsonl`、`<session>/subagents/*.jsonl` 和 `<session>/subagents/workflows/*/agent-*.jsonl`，重复的 assistant 块和配对的工具结果 | 推断出的模型/工具开始时间；记录下来的记录边界 |
| OpenCode | SQLite V1/V2（`message`/`part` 或 `session_message`）；旧版 JSON 存储 | 记录下来的模型和工具时间戳 |
| Gemini CLI | `~/.gemini/tmp/*/chats/session-*.json[l]`，含补丁和回退 | 从消息/工具事件边界推断出的区间 |

这些适配器遵循上游的 schema：
[OpenCode V1](https://github.com/anomalyco/opencode/blob/dev/packages/schema/src/v1/session.ts)、
[OpenCode V2](https://github.com/anomalyco/opencode/blob/dev/packages/schema/src/session-message.ts)、
[Gemini CLI](https://github.com/google-gemini/gemini-cli/blob/main/packages/core/src/services/chatRecordingTypes.ts)、
[Oh My Pi](https://github.com/can1357/oh-my-pi/blob/main/packages/coding-agent/src/session/session-manager.ts)。
Claude 的记录条目是拿本地的 Claude Code 会话比对的；它的 [hook 文档](https://code.claude.com/docs/en/hooks) 描述了记录的位置。独立的子 agent 记录不会被合并进来：父 agent 的 task/delegation 工具会被包含，但子进程内部的工作目前还没有被链接。不需要任何 hook 或客户端改动。

轮询使用一个软性的 8 MiB 读取预算，每个存储最多 200 个近期文件。JSONL 文件是增量读取的；没写完的行留到下一次轮询。OpenCode 的 SQLite 以只读方式打开，包含 WAL 里的改动，没有变化的存储会被跳过。SQLite 的 message/part 更新在各自独立的只读连接上按有界的批次增量读取；过大的对话在下次轮询时继续。OpenCode 每个会话最多保留 256 条消息和一个 8 MiB 的正文窗口；单条超过 8 MiB 的记录只导出元数据。旧版 JSON 和 Gemini 的快照读取上限是 32 MiB；Gemini 保留当前提示词加上最近 255 条消息。已完成的观测会在有界的元数据范围内去重。Claude 的最终回话和 Gemini 的最终 token 记录可能要多一轮轮询才落定。

推断出的计时带着 `starling.timing.source=inferred`。工具的参数/结果和用户的输入/输出遵循**包含请求与回话正文**这一项，会遮住机密，并遵守完整正文的偏好。会话的目录/标题元数据、凭据和账号名不会被导出。

会话追踪会取代这些本地客户端的独立网关 trace，以免 token 用量重复。不受支持的客户端和远端转发的请求保留网关追踪；本地的用量账本不变。这种模式展示的是 agent 的活动，而不是网关的重试或 provider 路由的细节。它读的是这台电脑上的本地会话，包括那些没有把星鸣当代理发出的调用。

指标按 agent、provider、模型、操作和错误状态分组，给出时长和输入/输出 token 的直方图。只有开启**包含请求与回话正文**时才导出提示词/回话文本（`STARLING_OTEL_BODIES=true`）；机密会被遮住，每份正文上限 256 KiB，除非开启**包含完整正文**。会话 ID 只在会话追踪开启时导出；provider 的账号名/key 永远不导出。

完整正文会在读回、清除机密和 JSON 编码时增加临时的内存和分配开销；一份 32 MiB 的请求和回话相比截断导出，总分配大约会翻倍。导出把排队的正文限制在 128 MiB，但这并不限制处理中的内存。最近的调用只保留每份正文的前 256 KiB。

导出在后台运行，用一个有界的队列（128 条记录）和每批最多 32 条记录的批次，每五秒冲刷一次。队列满了就丢掉遥测数据，不会拖慢网关的请求。网络错误和 HTTP 429/502/503/504 最多重试两次，重试延迟上限 60 秒；其它错误和部分拒绝会被记下，但不带采集器的响应正文。每次 HTTP 尝试三秒后超时。网关优雅关闭最多给三秒排空；`starling serve` 目前收到 SIGTERM 就退出，不做排空，所以它最后一批可能会丢。这是尽力而为的导出；失败时本地用量记录仍然可用。请求遵循星鸣的代理设置，回环上的采集器则直连。排队的记录会在导出被关闭、或者在发送前目的地或凭据变了时被丢弃。不跟随重定向。不带 key 的备份会省略 OTLP 响应头；恢复这样一份备份时，只有采集器 endpoint 没变才会保留这台机器的响应头。

## 镜像

<a id="mirrors"></a>

Plugins 页面的 **Mirrors in China**（国内镜像）开关，会让那个页面要下载的东西先问国内的镜像、再问它的官方地址：npm 的 registry（plugin 包、它们的版本和搜索）在 `registry.npmmirror.com`，Bun 的发布版在 npmmirror 的二进制镜像上，plugin 清单在 jsDelivr（`cdn.jsdelivr.net/gh/magpie-community/plugins`）。每一个都是同一个文件的副本——Bun 的 zip 仍然要对着 Bun 的 SHA-256 校验，包也要对着 npm 给的完整性值校验——而镜像还没复制过来的东西就去问官方地址。npm 的每周下载量仍然取 npmjs 自己的。`bun add` 也被指向 npmmirror，除非 `BUN_CONFIG_REGISTRY` 或 `NPM_CONFIG_REGISTRY` 已经指了某个 registry。这个开关绝不会把任何东西经由 gh-proxy 之类的代理发送。

镜像回退默认关闭。设 `STARLING_MIRRORS=on` 来启用它。开启时先试官方来源；只有发生网络错误、HTTP 429、HTTP 5xx 或 GitHub 的限流 403 之后才会去试镜像。默认是 npm 用 `registry.npmmirror.com`，GitHub 用 `gh-proxy.com`。`STARLING_NPM_REGISTRY` 和 `STARLING_GITHUB_MIRROR` 可以替换它们。

只有两处下载会用不来自镜像的校验和来检查：星鸣自己的更新产物（它们的 SHA-256 来自未经镜像的更新源）和 Bun 的 zip。Bun 的默认版本把它的 SHA-256 写在代码里，所以即使 Bun 官方的 `SHASUMS256.txt` 够不着，它也能通过镜像下载；更新的 Bun 仍然需要那个官方文件。Skill 的 tar 包总是用官方来源，plugin 市场的 registry 也是，除非打开 Mirrors in China 开关（上面说的 jsDelivr）。npm 的元数据、plugin 的 README/搜索结果和版本检查都不做校验和；镜像可以报一个不同的版本。

官方来源和镜像共用一份请求预算，所以官方来源卡住不会让整个等待时间翻倍。

带凭据的请求，比如带 `Authorization`、`Cookie`、API key，或 URL 里的 userinfo/查询 token，绝不使用镜像；它们只去官方来源。

```sh
STARLING_MIRRORS=on \
STARLING_NPM_REGISTRY=https://registry.example.com \
STARLING_GITHUB_MIRROR=https://gh.example starling serve
```

## 文件

<a id="files"></a>

星鸣的目录 `~/.config/starling` 是刻意保持扁平的：备份、同步、Docker 的卷挂载和旧版星鸣都在那里按名字找到每个文件（#2028）。下面每个文件都会说明它是干什么的。文件旁边的 `.lock` 在它被写入期间被持有，跨各个星鸣实例；`.bak` 是写入之前做的那份副本。

你设置的：

- `settings.json` — 设置
- `providers.json` — 你的 provider，包含它们的 key（0600）；里面的 `pluginLeft` 会保留被移除的 plugin 的模型选择，直到它被再次添加
- `profiles.json` — 保存的配置档
- `library.json`、`library/` — 资源库：它的指令、skills、MCP server 和 hooks，以及每一项要去往哪些 agent
- `plugins.json`、`plugins/` — 添加的 plugin，以及它们的包
- `plugins.yml` — 哪些内置 plugin 是开的，以及每个的配置
- `caller-keys.json` — 你为其它调用方建的网关 key（0600）

账号与登录：

- `logins.json` — 通过星鸣登录的订阅账号（0600）
- `plugin-auth.json` — plugin 的登录态（0600）
- `mcp-signins.json` — 到 MCP server 的登录态（0600）
- `cli-identity.json` — 某个 agent 自己的 CLI 上次说登录着的是谁（`cursor-agent about`），这样启动时不必等那些 CLI
- `redact.key` — 隐私脱敏用来替换被遮内容的那些占位符背后的 key

星鸣在你的 agent 上设置的：

- `stash.json` — 星鸣替换掉的那些值，切回去时还原
- `applied.json` — 星鸣最后在每个 agent 上设了什么，以及什么时候
- `migrations.json` — 哪些订阅迁到了它们的 plugin（[内置订阅迁到 plugin](../AGENTS.md)）

网关与用量：

- `usage.jsonl` — 每个请求一行：用量页面上的那些数字
- `routing/` — 每天的路由尝试（`<date>.jsonl`），路由页面的历史
- `affinity.json` — 每段对话是由哪把 key 或哪个账号应答的，这样重启之后它还在那里，而厂商那边也还缓存着
- `served.json` — 每个账号和 key 上次应答的时间
- `quotas.json` — 每个账号上一次报出的额度
- `quota-history.json`、`balance-history.json` — 一段时间里的额度和余额读数，给用量页面的图表用
- `usage-alerts.json` — 已经显示过的用量提醒，这样重启后不会再显示一次
- `client-shapes.json` — 网关最后为每个客户端放行的那些 header，供 CLI 那边测试用
- `gateway-sessions-hidden.json` — 你删掉的网关会话

为窗口保留的：

- `icons/` — provider 的图标，只取一次
- `plugin-icons.json`、`plugin-providers.json`、`plugin-updates.json` — plugin 的图标、每个提供的 provider，以及最近一次查看 npm 得到的结果（等着更新的，和已经做好的）
- `trash/` — 删掉的会话，直到你清空它或者把它们恢复
- `last-version`、`whatsnew-quiet-day` — 上一次运行的版本，供「更新了什么」用，以及你说不要再显示它的那一天
- `start-menu-shortcut` — 开始菜单的快捷方式已经建好（Windows）

在 `~/.cache/starling` 里：

- `bun/<version>/starling-bun` — 跑 Bun plugin 用的那个东西，名字起成这样，好让某个代理应用的 PROCESS-NAME 规则能匹配到它（Windows 上是 `starling-bun.exe`）
- `models.json` — models.dev 目录（存在 OpenCode 的那份缓存 `~/.cache/opencode/models.json` 时用它）
- `models/<provider>.json` — 从各个厂商取回的模型清单

`XDG_CONFIG_HOME` 和 `XDG_CACHE_HOME` 都会被尊重。

**图标从哪来。** plugin、导入链接和资源库的图标都放在各自作者挑的主机上。星鸣版不内置图标代理：图片直接从图标所在的那台主机取，不经过任何第三方代理，因此那台主机会看到取图者的地址。相关代码在 [`internal/iconproxy`](../internal/iconproxy/iconproxy.go)：它的默认代理地址是空的，也就是直连；把 `STARLING_ICON_PROXY` 设成自己的代理，才会改从那里取。

## 社区

<a id="community"></a>

问题、值得分享的配置、想法、bug：请在本仓库的 [issues](https://github.com/lulu58588/starling/issues) 和讨论里聊聊。Issue 和 PR 都欢迎。

## 许可

<a id="license"></a>

MIT。见 [LICENSE](../LICENSE)。
