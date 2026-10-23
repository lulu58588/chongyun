<div align="center">

# 星鸣 AI 网关

### 所有 Agent 的模型，一处搞定。

**Claude Code 跑 Kimi，Codex 跑 DeepSeek，Gemini CLI 跑 GLM，MiniMax Code 跑任意模型。**<br>
在一个本地网关上统一：多家供应商、多账号额度自动切换、本地模型、以及全部协议互转。

</div>

---

## 这是什么

每个 Agent 都把模型写在自己的配置文件里、用自己的密钥和地址。你买的订阅只在一个 Agent 里能用。
星鸣 AI 网关把这些收进一个本地网关：

- **任意模型配任意 Agent**：一个列表里挑，工具直接改写该 Agent 自己的配置字段，保留原有注释与排版。
- **一套供应商到处用**：API Key、本地模型，或你已登录的订阅账号。
- **额度用完自动接力**：路由组在多个模型/账号之间切换，Agent 侧看不到错误。

支持 macOS、Windows、Linux，含菜单栏应用、终端 TUI、Web 界面与 CLI。

## 快速开始

```sh
# 从源码安装（需要 Go 1.26+）
go install github.com/lulu58588/starling@latest

# 或直接跑源码
make cli && ./starling serve
```

```sh
starling provider add deepseek sk-…            # 预设供应商只需一个 Key
starling claude deepseek/deepseek-chat         # 让 Claude Code 走 DeepSeek
starling codex moonshot/kimi-k2.5               # 让 Codex 走 Kimi
starling group add "日常" models=a,b routing=smart   # 额度用完自动接力
starling quota                                  # 各订阅还剩多少
starling tui                                    # 终端里的一切
```

网关默认只监听 `127.0.0.1:3425`，同时说 OpenAI Chat、OpenAI Responses、Anthropic Messages、Gemini 四种协议。

## 本二开版与上游的差异

本项目基于开源项目 [yetone/magpie](https://github.com/yetone/magpie)（MIT 协议）二开，产品名为 **星鸣（Starling）**，著作权属 **崇云科技（Chongyun Technology）**。

**已清理：**

| 项 | 处理 |
|---|---|
| 用户行为统计上报（PostHog） | **整包移除**，本版本不向任何第三方上报 |
| 官方更新源 / 自动更新 | 已关闭；如需自建更新源，设 `STARLING_UPDATE_FEED` |
| 官方合作伙伴/赞助商清单 | 已关闭 |
| 官方图标代理 | 改为直连（图片站可见你的 IP，已在隐私说明中注明） |
| 官方站点与跨域白名单 | 官网目录 `site/` 已移除，跨域白名单收敛为空 |
| **读取其他 Agent 的登录凭据**（Claude Code / Codex / 钥匙串） | **整层移除**：不再读取任何 Agent 的登录态，也不再从别家工具导入登录文件 |
| 上游的图标、Logo、产品名、文案 | 全部替换为星鸣 |

**故意保留的技术标识**（改动会破坏与各 Agent 的兼容性与互操作，不是品牌内容）：

`~/.config/starling` 配置目录 · `/v1/starling/*` 接口路径 · `X-Starling-*` 响应头 ·
`STARLING_*` 环境变量 · 各 Agent 配置里的 provider id `starling` ·
Go module path `github.com/lulu58588/starling`（导入链接的 scheme 已是 `starling://`）。

**必须保留的法律声明**：根目录 `LICENSE` 中 MIT 协议原文、`Copyright (c) 2026 yetone`
与 `internal/qoder/LICENSE` 中第三方作者的版权声明，均原样保留，二开说明见 LICENSE 末尾。

## 安全须知

- 网关在回环地址上**不校验密钥**（上游的既定设计）：本机任何程序都能调用它并消耗你的额度。
  不要把它开给局域网或多人共用，除非你自己加了一层鉴权。
- 订阅账号复用（Claude / ChatGPT / Copilot 等）在上游属于灰色用法，本版本**不读取任何 Agent 的登录态**
  （相关功能已整层移除，无法再开启），因此也不会因此触发厂商风控。
  星鸣自身签入的账号（Codex PKCE、Copilot 设备码、Google OAuth）不受影响。
- 局域网共享默认关闭；如开启，请务必配置 `STARLING_ALLOW_IPS` 白名单（注意：默认白名单为空即放行全部）。

## 许可

MIT，见 [LICENSE](LICENSE)。上游项目版权归 yetone 所有，二开版权归崇云科技（Chongyun Technology）所有，产品名为星鸣（Starling）。
