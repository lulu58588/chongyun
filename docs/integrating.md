# 把你的应用或 Agent 接入星鸣网关

面向希望与星鸣网关良好配合的 AI 应用、Agent 与客户端开发者。接入分三个层级，从完全不需要改动网关这一侧，到在 Agents 页面拥有自己的一行。

| 你的目标 | 路线 | 对星鸣网关的改动 |
| --- | --- | --- |
| 让你的应用请求由星鸣网关提供服务，并在用量里记在你自己的名下 | [与网关对话](#1-talk-to-the-gateway) | 无 |
| 在 Agents 页面拥有一行：可被识别、可在那里选模型、断开后还原 | [一个 Agent 适配器](#2-a-row-on-the-agents-page) | 提交一个 PR |
| 新的模型来源（某个订阅、某个厂商），或改写请求 | [一个插件](#3-plugins) | 无 |

## 1. 与网关对话

<a id="1-talk-to-the-gateway"></a>

星鸣在本机运行一个 LLM 网关。用户在此配置好的每一个模型（订阅、API Key、本地模型、路由组）都在这里以各自唯一的 id 对外提供，因此你的应用只需要一条供应商配置，而不是每家厂商各写一条。

**找到它。** 网关监听 `127.0.0.1:3425`，或星鸣 `settings.json` 里 `port` 指定的端口（`~/.config/starling/settings.json`；Windows 下是用户目录下的同一路径；若设置了 `$XDG_CONFIG_HOME/starling` 则在该目录下）。`GET /api/hello` 在网关为星鸣时返回 `{"name":"星鸣","version":"…"}`。只读取这个文件，不要写它。星鸣网关需处于运行状态（应用本身或 `starling serve`）。若某个应用在此什么也没找到，就照常提供它自己的供应商。

**接口。** 当厂商本身就说你的 API 时请求直接透传，否则会被翻译，流式输出、工具调用、图片与推理都一并处理：

| 路径 | API |
| --- | --- |
| `/v1/chat/completions` | OpenAI Chat Completions |
| `/v1/responses` | OpenAI Responses |
| `/v1/messages`、`/v1/messages/count_tokens` | Anthropic Messages |
| `/v1beta/models/{model}:generateContent`（以及 `:streamGenerateContent`） | Google Gemini |
| `/v1/models` | 模型目录 |

因此对 OpenAI 客户端，base URL 是 `http://127.0.0.1:<port>/v1`；对 Anthropic 或 Gemini 客户端是 `http://127.0.0.1:<port>`。

**你的应用使用的密钥名。** 在本机上网关接受任意密钥。发送 `starling-<your-app-id>`（`Authorization: Bearer starling-acme`，或 `x-api-key`），每次调用就会在用量、请求日志以及按 Agent 匹配的路由规则里都记作 `acme`。不带它的话，星鸣网关会按你 `User-Agent` 的第一个词计名（`Acme/1.4 (darwin)` 记作 `Acme`）；SDK 自带的 User-Agent（`OpenAI/JS …`、`ai-sdk/…`）则会被记成那个 SDK。从另一台电脑访问（设置 → 在局域网上共享）时，密钥必须是用户的网关密钥之一，并由 User-Agent 指明你的应用。

**模型。** `GET /v1/models` 列出用户已启用的内容：`id`（形如 `provider/model`，或路由组的名字——原样发送即可）、`display_name`、`owned_by`（供应商）、`reasoning` 与 `supported_reasoning_levels`（`[{"effort":"low"},…]`，`reasoning_levels` 里是同样内容的平铺写法）、已知时的 `context_window` 与 `max_output_tokens`、以及星鸣知道该模型是否接受图片时的 `modalities`。请从那里读取列表，不要自己维护一份：用户会在星鸣里增删供应商。推理强度要用你这套 API 自己的字段来传（`reasoning_effort`、`reasoning.effort`、Anthropic 的 `thinking` 或 `output_config.effort`）。加 `?format=text` 可让 id 一行一个。

**可选。** 对同一会话的请求带上 `X-Starling-Session: <id>`，然后 `GET /v1/starling/route?session=<id>` 会说明路由组为这一轮挑了哪个模型、尝试过哪些回退（在第一个 token 之前就能拿到，长轮询参数为 `after=<seq>&wait=<s>`）。`GET /v1/starling/quotas` 列出用户的订阅额度；`GET /v1/starling/usage` 与 `/v1/starling/usage/requests` 给出用量页面的数字与请求明细（可带 `session`，以及 Claude Code 子 Agent 调用时的 `subagent`/`parent_agent`）。这些接口只对本机作答，或对持有网关密钥的另一台电脑作答；一个受预算、模型或账号限制的密钥只能看到自己的调用与路由，除非它的持有者允许它查看可访问对象的额度（`starling gateway-key quotas <id> on`，#2216；它自身的上限是 `GET /v1/starling/limit`）。你的应用直接向厂商发起的调用，也可以用 `POST /v1/starling/usage/requests` 上报（一次一条，或一次最多 5000 条，格式与 `usage.jsonl` 相同，每条带 `rid`；过去时间也会保留，因此可以回填历史）。它们在用量页面上显示为「已上报」，且永不计入某个密钥的上限。详见[接入其他任何东西](reference.md#connecting-anything-else)。

**保留用户自己的供应商。** 把星鸣网关作为你应用里众多供应商中的又一个（命名为 Starling）并列加入，而不是替换掉用户已有的供应商，这样"关掉星鸣"就等于改选另一个供应商。

## 2. 在 Agents 页面拥有一行

<a id="2-a-row-on-the-agents-page"></a>

星鸣列出的 Agent（Claude Code、Codex、OpenCode、dsh、Alma、Cindy……）是 [`internal/agent`](../internal/agent) 里的适配器，一个应用一个文件，统一登记在 [`agents.go`](../internal/agent/agents.go)。这样一行会显示你应用的名字与图标、反映你是否已安装、允许用户从星鸣的模型目录里挑选它的模型（以及推理强度）、跟随目录随供应商变化，并在断开时还原你应用原有的设置。没有运行时注册机制：新增一行需要向本仓库提交 PR，随下一个版本发布。

一个适配器是一个 [`Agent`](../internal/agent/agent.go) 值。它的形态由你的应用能提供什么决定：

- **星鸣直接改写你的配置文件**（多数适配器如此；[`fx.go`](../internal/agent/fx.go)、[`empryo.go`](../internal/agent/empryo.go)）。`Dir`、`Path` 与 `Bin` 用于定位应用；每个 `Field`（`model`、`effort` 等）都有 `Get`、`Set` 与 `Options`。接入时会在网关侧加一个 `starling` 供应商，键为 `starling-<id>`，并把你的模型设为星鸣的某个模型；`Set("")`（Agent 的默认值）或 `Unwire` 会移除星鸣写入的条目并还原原状。只有星鸣设置过的键会被改动，用户其他设置、供应商与注释都保持不变。`Notice` 说明你的应用何时需要重启才能看到改动；`Sync` 用于改写那些把模型列表存在文件里、而不是通过 `/v1/models` 获取的应用。
- **你的应用只通过自己的导入链接接入供应商，且由用户在你的应用里确认**（[`cindy.go`](../internal/agent/cindy.go)）。`Import` 返回那个链接（`yourapp://provider/import?…`，其中带着网关地址、密钥 `starling-<id>` 与各接口），这一行的按钮会打开它，用户在你的应用里点一次同意。`Added` 以只读方式读取你的应用里是否已经接入星鸣。若你的应用没有公开的状态读取器（正如 [Lody](../internal/agent/lody.go) 所做的），就留空 `Added`；此时这一行会请用户在你的应用里确认，而不会声称自己知道导入状态。星鸣不会写入你应用的任何东西。这适合把供应商存在数据库或加密存储里的应用。
- **你的应用在运行时有一个本地 API**（Alma，在 `localhost:23001`）：星鸣通过它把自己加为供应商，并设置默认模型。

一个 PR 需要包含：适配器及其登记、`UA`（你的 User-Agent 以什么开头，小写），让用量能认出你的调用；`internal/gui/assets/icons` 里的图标；在临时 `HOME` 下运行的测试（接入、切换、断开后还原原文件）；以及 [Agents 表格](reference.md#agents)里的一行。对我们最有帮助的是：稳定且有文档的配置格式或导入链接，以及你的应用在各个系统上分别把它存在哪里。

## 3. 插件

<a id="3-plugins"></a>

插件是星鸣运行的一个 npm 包；用户在设置 → 插件里安装，或用 `starling plugin add <package>` 安装。插件分两类（详见[插件指南](subsystems/plugin-runtime.md)）：

- **供应商插件**，即 OpenCode 的 `auth` 钩子或一个 pi 包：提供新的模型来源，例如星鸣自己不登录的某个订阅。它们负责登录、列出模型并发起请求。
- **网关中间件**：在请求与应答经过时改写它们（`onRequest`、`onProvider`、`onEvent`、`onResponse`）。

插件无法在 Agents 页面新增一行，也无法改写其他应用的配置，所以它不是应用自我登记的方式。

## 没有的部分

- 没有任何 API 或链接能让其他应用在未经用户同意的情况下自我登记为 Agent，或自行安装插件：插件由用户在星鸣里安装，Agent 行来自星鸣源码中的适配器。
- `starling://import` 链接添加的是*供应商*（一个模型来源，带自己的密钥），而不是 Agent 或应用；见[导入链接](reference.md#import-links)。

有疑问，或希望先讨论某个适配器：请在 <https://github.com/lulu58588/starling/issues> 提交 issue。
