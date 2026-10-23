# provider 与 plugin 归属

有些内置订阅已经废弃，并有提供同样服务的社区 plugin。一个订阅仍然可以用它的内置实现，也可以迁到它的 plugin。这份归属决定了登录、刷新、模型发现、请求、用量和错误分别在哪里处理。

## 职责与事实来源

| 部分 | 职责 | 来源 |
| --- | --- | --- |
| provider 迁移 | 把订阅与 plugin 对应起来、记录归属、转移账户并维护所需的 plugin 版本 | [`migrate.go`](../../internal/provider/migrate.go) 和同目录下的 `migrate_*.go` |
| provider 集成 | 暴露由 plugin 支撑的 provider 卡片、用量、登录，以及 plugin 自己的每日签到 | [`plugins.go`](../../internal/provider/plugins.go)、[`plugin_usage.go`](../../internal/provider/plugin_usage.go)、[`pluginsignin.go`](../../internal/provider/pluginsignin.go)、[`plugin_checkin.go`](../../internal/provider/plugin_checkin.go) |
| plugin host | 加载 plugin 并执行它们的代码 | [`internal/plugin`](../../internal/plugin)、[`host.js`](../../internal/plugin/host.js) |
| 网关 | 把请求分发给拥有该订阅的实现 | [`internal/gateway`](../../internal/gateway) |
| GUI | 展示可用的订阅、迁移状态和 plugin 登录 | [`app.js`](../../internal/gui/assets/app.js) |
| 社区 plugin | 实现已迁移订阅的上游行为 | [magpie-community/plugins](https://github.com/magpie-community/plugins)，`packages/<name>` |

在 `internal/provider/migrate*.go` 里组装起来的 `movers` 映射表是订阅到 plugin 的权威对应关系。每个 mover 的 `min` 是星鸣要求的最低 plugin 版本。[AGENTS.md](../../AGENTS.md#built-in-subscriptions-that-a-plugin-serves) 里的已迁移订阅清单是给贡献者查的；改动它时要在源码里确认映射关系。

## 归属与运行时路径

在 `migrations.json` 里迁移状态为 `plugin` 的订阅，由 `provider.Moved(id)` 识别出来。在 `/api/providers` 里，它那一行暴露 `move.state == "plugin"`；GUI 也会在 `onPlugins` 里收到它的 id。

已退役的内置实现（[`migrate.go`](../../internal/provider/migrate.go) 里的 `retiredBuiltins`，`provider.BuiltinRetired(id)`）不再有内置代码：无论迁移跑过没有，都由 plugin 提供服务，它那一行的 `move.state` 是 `plugin`。过去问 `Moved(id)` 的关卡现在问 `provider.PluginServes(id)`，也就是已迁移或已退役。已退役订阅里还没迁移的内置账户没有任何东西在服务；`/api/providers` 把它们列在 `unmoved` 里（每一项是一个 `MoveCandidate`，带上失败迁移的 `error` 和 `why`），GUI 在供应商列表上方显示一行，带一个「迁移」按钮（`app.js` 里的 `unmovedLine`）。内置实现是一个一个退役的，每个都在删掉它代码的那个提交里退役。

| 操作 | 内置实现拥有该订阅 | plugin 拥有该订阅 |
| --- | --- | --- |
| 登录、刷新和模型 | 内置 provider 实现 | 通过星鸣的 plugin 集成的社区 plugin |
| 上游请求与协议转换 | 内置网关适配器 | 经 plugin host 的 plugin fetch |
| 用量与上游错误 | 内置实现 | plugin，经星鸣的 provider 集成暴露 |
| 迁移状态与所需版本 | 星鸣 | 星鸣 |

对已迁移的订阅，它的内置上游实现不再运行。plugin 账户使用 `Account.Agent == "plugin"`；请求经 `Provider.Do` 到达 `plugin.Fetch`，绕过内置的订阅适配器。星鸣仍然负责路由、协议转换、配额处理和兼容性调整，以及 plugin host、provider 集成、迁移代码和 GUI 的 plugin 路径。

在 [`quotas.go`](../../internal/provider/quotas.go) 里的配额聚合，只有当一个 key 的 GLM Coding Plan 卡片与某个 ZCode 订阅卡片的可比重置窗口一致时才会被隐藏：内置或已迁移的 `zcode`，或者单独安装的 `zcode-plugin`。[`PlanQuotas`](../../internal/provider/planquota.go) 通过配额端点识别 GLM 计划，包括自定义 provider id。互不相关厂商的重置时间相同不会隐藏卡片。在这一对里，重置仍然只是共享账户的启发式判断；计划的 `User` 是 key 的标签或掩码，不是登录身份。出错、重置冲突或没有可比重置时，计划照常显示。这适用于用量页、托盘、配额报告、告警和配额等待；它不改变 plugin 归属或上游请求。

## plugin 延迟收集的用量

如果一个 plugin 只能在某一步结束之后才拿到这一步的用量，它就会随后面某一步一起上报，放在 `usage.starling_deferred` 里（`prompt_tokens`、`completion_tokens`、`cached_tokens`、`cache_write_tokens`），与那一步自己的用量字段并列。Cursor 的 plugin 就是这样：工具步骤的用量事件在步骤结束后才出现，所以这一步写 0，而对话的下一步收走这个事件（[#1053](https://github.com/yetone/magpie/issues/1053)）。

网关把这个字段解析进 `Usage.Deferred`（[`cUsage`](../../internal/gateway/chat.go)），并且只在计费的地方加上它：用量记录、OTel 的 attempt span 和路由定价，后者把它当作排在请求自身之前的一档来定价（[`routeUsage`](../../internal/gateway/trace.go)、[`Usage.billed`](../../internal/gateway/ir.go)）。告诉客户端的仍然只有请求自身的数值，因为 Agent 把 `prompt_tokens` 当作上下文有多满来读：透传路径原样转发 plugin 的字节，其标准字段就是那一步自己的；翻译过的回复则只按自身计数编码。`promptCounted` 同样读请求自身的值。没有这个字段的回复按老样子处理。用较旧版本的星鸣时这个字段会被忽略，欠下的那一步在账本里不计费。[`deferred_usage_test.go`](../../internal/gateway/deferred_usage_test.go) 覆盖了透传、翻译后的回复、账本、OTel span 和定价档位。

## plugin 的每日签到

<a id="a-plugins-daily-check-in"></a>

plugin 可以自己去按厂商的每日签到：它的 `auth` hook 在 `usage` 之外多出 `checkin(getAuth, provider)`。宿主会把带 `checkin: true` 的 provider 列出来（`plugin.Provider.Checkin`），并通过像 `usage` 那样刷新账户、带着账户的 auth 调用该 hook，来响应 `checkin({provider, account})`（`plugin.AccountCheckin`）。

该 hook 返回 `{outcome, credit?, streak?, message?}`。`outcome` 是下列之一：`claimed`（刚刚签到）、`done`（今天已经签过）、`ineligible`、`inactive`（现在没有签到事件）、`captcha` 或 `failed`。抛出异常或给出不在列表里的 outcome，都算 `failed`，并带上原因。`captcha` 意味着厂商要验证码：星鸣从不解验证码，只会提示「到它自己的应用里签到」。plugin 同样不许解验证码。

[`plugin_checkin.go`](../../internal/provider/plugin_checkin.go) 让这些走内置签到（[`checkin.go`](../../internal/provider/checkin.go)）用的同一个 `checkiner`：每个已登录、在用的账户每个北京日一次，启动 2 分钟后首次查看、之后每 30 分钟一次，`failed` 的结果 30 分钟后再问一次，其他 outcome 当天就此落定。结果按 `provider|<sign-in key>` 存在 `plugin-checkin.json` 里（`pluginCheckinKey`：`plugin-auth.json` 里保存该登录所用的 key，`autoclaw` 或 `autoclaw#x1y2z3`），绝不按显示名保存：两个登录可以同名（AutoClaw 的默认名改不了），若按名字做键，历史上只有第一个被签到过，无论是被循环签到的还是被「立即签到」签到的（plugins#254）。登录 key 在刷新之后、再次登录同一账户之后都保持不变（`host.js` 里的 `settle` 会把它放回自己的 key 下）；同一天里某个登录被登出之后再做的登录可以占用空出来的 key，连同当天的结果一起。此前按 `provider|<display name>` 保存的条目，会在 plugin 的账户被列出之后被承接过来（`adoptLegacyPluginCheckins`）：签到会落到该名字下第一个在用的登录，也就是旧 key 问过的那个，且只落到它，所以同名的其他登录会被问到，且不会被问两次；`settings.CheckinOff` 里的一个 `Off` 则落到该名字的每一个登录上，因为关闭期间它们谁都没被签到。读不了或解析不了的 `plugin-checkin.json` 保持原样，等能读了再承接过来；在那之前没有任何 plugin 账户被签到，每个账户都会说明原因（`readCheckins`，和每个厂商的签到文件一样）。失败被记为失败，绝不重置账户、key 或前一天的结果。每个账户的用量卡片上都有签到这一行（`checkinBy: "plugin:<provider>"`），设置 → 用量里有一个插件标签页，每个 provider 一个「每日自动签到」开关（`settings.PluginCheckins`、`POST /api/settings/plugin-checkin {provider, on}`；「立即签到」是 `POST /api/usage/plugin-checkin {provider}`）。`starling accounts checkin` 和 TUI 用量页上的 `c` 也包含它们。每一次签到，无论是内置的还是 plugin 的，都会跳过按文件 key 列在 `settings.CheckinOff` 里的账户（#1902）：无论循环还是手动按，`checkinNow` 都不问它；`CheckinAccounts()` 为设置页列出每个账户各自的 `on`（`checkinAccounts`，某厂商有两个及以上账户时每个账户一行显示在其开关下面，`POST /api/settings/checkin-account {key, on}`），它的卡片上带 `checkinOff`。

这个开关默认关闭。对那些星鸣此前通过 plugin 的 fetch 签到的厂商所属的 provider（WorkBuddy、Trae CN、MiniMax Code、Qoder），它跟随该厂商已有的开关，直到单独为它设置；而星鸣自己的签到把该 provider 的账户交给 plugin（`pluginChecksIn`），所以一个账户不会被签到两次。

没有 `auth.checkin` 的 Trae CN plugin 通过 plugin 的 fetch 使用 [`trae_checkin.go`](../../internal/provider/trae_checkin.go)：先 `/status`，然后在签到开着且今天还没签过时 `/claim`。claim 返回 9074（当前参与用户太多，请稍后再试）时记为 `own-app`，和其他非 `failed` 的回答一样当天落定：任何重试都没能越过它（#808），所以当天循环和「立即签到」都不会再 claim，而卡片、设置、TUI 和 CLI 会说 Trae CN 拒绝了签到（9074），让你去 Trae CN 的应用里签，而不是说 Trae 只认自己的应用——Buddy Switch 已经推翻了这个说法（#1920）。从 plugin 0.2.14 起，本机 Trae CN 桌面客户端登录的那个账户在签到时会带上客户端自己的设备 id（其 `storage.json` 里 `iCubeAuthInfo://icube-dc:<id>` 这个键，只读）；其他每个账户都保留自己的设备，不会编造也不会轮换（#1705）。`/status` 上的 9074 和其他每个代码都保持原有的处理方式。`own-app` 不是 plugin hook 的 outcome。

WorkBuddy 的签到（[`workbuddy_checkin.go`](../../internal/provider/workbuddy_checkin.go)，内置的或经 plugin 的 fetch）能把 401 区分开。WorkBuddy 的网关对拒绝的 token 只回一个裸的 HTML 401，而卡片对此只显示「Unauthorized」并每 30 分钟重试一次。当 `/v2/billing/meter/checkin-activity-status` 或 `/daily-checkin` 返回 401 时，用同样的请求头读该账户的额度（`/billing/meter/get-user-resource-summary`）：读得到，说明登录是好的，只是签到拒绝了星鸣（WorkBuddy 自己的应用用自己安全 SDK 的 `X-Device-Token` 发它的签到，而且只发那一个），于是当天记为 `own-app`；读不到，则是 `failed`，并提示「重新登录」。

没有 `auth.checkin` 的 Qoder plugin 仍然通过 plugin 的 fetch 使用 [`qoder_checkin.go`](../../internal/provider/qoder_checkin.go)。是否 claim 由服务端的 `CLAIMABLE` 状态决定；本地的 `[startAt, endAt)` 检查只用来识别已经 claim 过的活动及其到期时间，包括跨越北京午夜的那个。成功的结果把当前活动的结束时间保存在 `qoder-checkin.json` 的可选字段 `until` 里。定时检查最多复用同一北京日的结果 30 分钟，且不会越过那个结束时间，这样一个长活动就不能无限期地遮住新的或重置的活动。现有循环每 30 分钟查看一次，并在北京日切换时查看；午夜刷新会把今天的结果保存下来给用量卡片和 TUI 用。`inactive` 的结果现在 30 分钟后会再问一次，而不是当天落定；失败仍按原有的重试计划。手动签到总是立刻去问。没有 `until` 的已保存结果依然可读，也用同样的 30 分钟刷新，不需要迁移。有自己签到的其他厂商和 plugin 保持它们每天落定的做法。

## plugin 自己的选项

<a id="a-plugins-own-options"></a>

plugin 可以为自己的 provider 声明开关（#2038）：它的 `auth` hook 多出 `options: [{key, label, description?, default?, confirm?}]`。宿主（[`host.js`](../../internal/plugin/host.js) 的 `optionsOf`）保留 `key` 匹配 `[a-z0-9_-]{1,40}` 且 `label` 非空的选项，丢弃出现过的重复 key，把 label 限制在 60 个字符、description 和 `confirm` 限制在 300 个字符，`default` 只在它是 `true` 时才采用，并且只保留前 16 个。provider 会带着它们一起列出（`plugin.Provider.Options`）。

用户的值按 plugin 自己的 provider id 和选项的 key 保存在 `settings.PluginOptions` 里，用 `POST /api/settings/plugin-option {provider, key, on}` 设置，该接口用同样的方式检查 key。宿主在 `init` 时拿到它们（`options`），每次改动之后再拿一次（`setProviderOptions`、`plugin.SetProviderOptions`，只发给已经在运行的宿主）。plugin 通过 `input.starling.option(provider, key)`（用户的值，没有则用默认值；它没声明过的 key 返回 false）或 `input.starling.options(provider)`（所有声明过的 key）读取它们。两者每次都读当前值，所以改动无需重启就能到达 plugin。为旧宿主写的 plugin 要先检查 `input.starling?.option`。选项变化时不会触发任何运行：星鸣只是把值保存下来并交出去。签到开关是另一回事，没有变化。

provider 的编辑器把它们列在*插件选项*下（[`app.js`](../../internal/gui/assets/app.js) 里的 `pluginOptionsField`，值来自 `/api/providers` 的 `plugins[].options[].on`），每个 plugin 的 label 和 description 旁边各有一个关/开，作为文本填入。带 `confirm` 的选项在开启前会先确认。改动要等保存，保存时会先逐个提交改过的选项，再提交 provider。

## 上游自行搜索的 plugin

处在 Anthropic 或 OpenAI 自家 API 前面的 plugin（比如一个 CPA 或 sub2api 中转）可以声明它的上游会自己应付联网搜索工具（#2027）。它的 `config` hook 在 provider 上设置 `searches: true`；宿主把它透传（`plugin.Provider.Searches`），[`pluginProvider`](../../internal/provider/plugins.go) 则设置 provider 的 `Searches`，就像自定义 provider 的「自己能联网搜索」那样（#359）。[`searchesItself`](../../internal/gateway/search.go) 只对 plugin 的 provider、只在 Anthropic 或 Responses 的 API 上认这个标记，而且只对 plugin 有模型的 provider 认：客户端的 `web_search` 于是原样发给 plugin 的 `fetch`，不经过星鸣的搜索。别的登录的 `Searches` 从不被读取，Codex 和 Grok 的保持原样。和中转一样，`searchRank` 会把它排除在外，所以星鸣绝不会挑它去替别的模型搜索；设置里可以手动指定（`namedSearcherRank`）。

plugin 也可以改为只给自己列表里的单个模型标上 `searches: true`（在 `config` hook 的模型上或 `models` hook 的列表里）：也就是用普通说法问它时它会自己联网搜索的模型，比如 Trae 的搜索模型。宿主把它透传（`plugin.Model.Searches`），`pluginCatalog` 把目录模型的 `WebSearch` 标为「原生」，网关在设置 › 联网搜索的选择器里提供该 provider，自己的挑选里则只提供带这些模型的（`pluginSearches`、`partSearcher`）。

## plugin 的服务档位

plugin 用模型上的 `serviceTiers` 说明它的厂商以哪些服务档位提供这个模型（#2514）：诸如 `"priority"` 和 `"ultrafast"` 的 id，或带 `id` 的对象。宿主保留这些 id（`host.js` 里的 `tiersOf`，写入 `plugin.Model.ServiceTiers`），`pluginCatalog` 把它们放到目录模型的 `Tiers` 上，[`CodexTiers`](../../internal/provider/codex_models.go) 则按 `codexcat` 的叫法把它们提供给 Codex（`priority` 就是 Fast）。在 [`forwardTranslated`](../../internal/gateway/gateway.go) 里，翻译成 plugin 的 Chat 或 Responses 的请求只有在 `service_tier`（`Request.OwnTier`）属于该模型的档位时才保留它（`provider.PluginTiers`）；用自身协议中转的请求则仍按客户端写的样子发送，和以前一样。一个模型在自己的 Chat 或 Responses API 上以 `priority` 提供时就有快速模式（[`CanFast`](../../internal/provider/groupfast.go)，经 `pluginTierOn`），无论它属于哪个 plugin：它可以是某个 group 的快速成员，也可以是某个 Agent 的快速首选，当它同时以 `ultrafast` 提供时还能被挑成 Ultrafast（`UltraOn`）。Claude 意义上的快速模式（速度 `"fast"`，或在 Claude 的 API 上由某个 Agent 作出的快速首选）在模型以该档位提供时以 `priority` 档位到达 plugin，Ultrafast 首选则是 `ultrafast`（Discord 上的 Lovest）。Cursor 的 `fast` 标志和 `-fast` 系列的 plugin 仍和以前一样享有快速模式。

一个翻译到 Grok plugin 的 Responses API、关闭了工具且没有声明任何工具的请求（Chat 的 `tool_choice: "none"`、Anthropic 的 `{"type":"none"}`、Gemini 的 `functionCallingConfig.mode` `NONE`）到达 plugin 的 `fetch` 时带着 `tool_choice: "none"` 且没有 `tools`（#3110）。以前它是不带 `tool_choice` 出去的，和没有任何偏好的请求一样，于是添加自己服务端工具的 plugin（Grok 的 `x_search`，magpie-community/plugins#263）没法尊重这个设置。[`buildResponses`](../../internal/gateway/responses.go) 保留了 `"none"`，而 [`forwardTranslated`](../../internal/gateway/gateway.go) 对其他每一个上游都不带上单独的 `tool_choice`（`keepsLoneNone`）：每个厂商、xAI 和 Copilot（`refusesLoneToolChoice`，#378），以及其他每一个 plugin——它们的厂商还没试过单独的 `"none"`。所以发给厂商的内容没有变化。Grok 的 Responses 后端同样不收它：Grok plugin 的请求体由 `grokPluginBody` 塑形，它保留了内置实现里 `grokBody` 会丢掉的单独那个 `"none"`，而 plugin 在请求发给 Grok 之前把它去掉（0.1.18，mover 的 `min`）。Chat、Anthropic 和 Gemini 的请求在没有工具时依旧不带 tool choice，OpenAI 的 Chat 会拒绝这种请求（"'tool_choice' is only allowed when 'tools' are specified"）。[`tool_choice_none_test.go`](../../internal/gateway/tool_choice_none_test.go) 里的 `TestToolsOffWithoutToolsReachesPlugin`、`TestToolsOffWithoutToolsNotSentToOtherPlugins`、`TestToolsOffWithoutToolsNotSentToVendor`、`TestToolsOffWithoutToolsLeftOutForXAI` 和 `TestToolsOffWithoutToolsPerBuilder`，以及 `TestGrokPluginBodyKeepsLoneNone`，覆盖了这些。

## plugin 的决策 API

plugin 可以为自己的 provider 提供 TypeSafe 的 System One（#1514）。它的 `config` hook 在 provider 上设置 `decide: true`，并在回答问题而非对话的模型上设置 `decides: true`（在 `config` hook 里，或在 `models` hook 的列表里）。宿主把两者都透传（`plugin.Provider.Decide`、`plugin.Model.Decides`）。[`pluginProvider`](../../internal/provider/plugins.go) 随后把 provider 的 `Decide` 设成 `plugin://<id>/v1`，也就是它自己的 base：`pluginFetch` 只往加载器的宿主发请求，所以 plugin 自己的决策地址是够不着的。`AskSystemOne` 和网关的分类器用 `Provider.Do` 发送 `/v1/systemone`，这会带着账户的登录到达加载器的 `fetch`，形式是 `{baseURL}/systemone`。

在 [`decide.go`](../../internal/provider/decide.go) 里，plugin provider 的决策模型是它标记的那些（`pluginMarks`，从 plugin 当前的列表读出），以及按名字认的 Jev 的模型（`jev-*`）。它们不提供给 Agent（`buildEntries` 里的 `isDecision`），用于聊天会被拒绝（「只做决策」），作为 group 里与对话模型并列的成员也会被拒绝（单独由决策模型组成的 group 会回答 `/v1/systemone`：[gateway 路由](gateway-routing.md#remote-starling-apis)），而是作为某个 group 的分类器提供（`Deciders`、`DecisionModels`）。只有 provider 声明了 `decide` 时，标记才起作用；否则每个模型仍旧是可以聊天的。plugin provider 的编辑器（账户编辑器）没有地址字段，所以这个地址从不会被手输；GUI 根据 `deciders` 标出决策模型。

## plugin 的 fetch 里，Agent 在哪里工作

<a id="where-the-agent-works-for-a-plugins-fetch"></a>

plugin 的 fetch 跑在星鸣的 plugin host 里，所以它的 `process.cwd()` 是星鸣自己的目录，而不是发起请求的 Agent 的目录。某个 plugin 把用户的环境告诉厂商时——比如 ZCode plugin 的 Start Plan 请求在它的 `# Environment` 段里那样做——说成了星鸣的 `~/.config/starling` 是工作目录，于是模型就去那儿找了（magpie-community/plugins#86）。

网关会告诉 plugin。每一次发往某个 plugin 的 provider 的请求，`forwardOnce` 都调用 `tellWorkPlace`（[`workdir.go`](../../internal/gateway/workdir.go)），它按与按目录使用判定相同的 `workPlace`（[gateway 路由](gateway-routing.md)）读取发给 plugin 的请求，并在 plugin 的 fetch 上设置两个请求头：

| 请求头 | 值 |
| --- | --- |
| `x-starling-work-dir`（`provider.WorkDirHeader`） | Agent 说自己工作的目录，百分号编码的 UTF-8（由 `decodeURIComponent` 读取）。 |
| `x-starling-work-git`（`provider.WorkGitHeader`） | `yes` 或 `no`：那个目录是否在一个 git 仓库里。Agent 自己说过的话优先，其次是目录旁边的说明（Claude Code 的 `Is a git repository: true`/`false`，OpenCode 的 `Is directory a git repo: yes`/`no`），都不行时，若目录在本机上，则看目录里或目录上层有没有 `.git`（[`workplace.go`](../../internal/provider/workplace.go) 里的 `provider.GitSaid`、`provider.GitOnDisk`）。 |

各自在无法获知时不设置：Agent 没说目录，或者说了但那个目录不在这台机器上、也没说它是不是仓库。于是 plugin 就不写它不知道的东西，而不是改写成星鸣的值。宿主的 `fetch`（`host.js` 里的 `unsaid`）会把 plugin 发出的每个请求上的这两个请求头都去掉，所以照着收到的请求头转发它们的 plugin，不会把它们发给任何厂商。

`TestPluginToldWhereTheAgentWorks`（[`plugin_workdir_test.go`](../../internal/gateway/plugin_workdir_test.go)）把抓到的 Claude Code 和 Codex 请求，以及 Pi 风格的请求，从一个在仓库内的目录和一个不在的目录，通过假 plugin 发出去，检查 plugin 读到了什么，以及厂商两个请求头都没收到。

## 一个 provider id 对应两个 plugin

宿主对每个 provider id 只运行一个 plugin。当两个已安装的 plugin 对同一个 id 都有 `auth` hook 时，比如一个第三方 plugin 和用户自己按目录加进来的副本，`host.js` 里的 `auths()` 取 `plugins.json` 的 `prefer` 映射里为该 id 选定的那个（provider id → spec）。没有选定就取列表里的最后一个，和 OpenCode 一样。只有提供服务的那个 plugin 的 `provider.models` hook 会为该 id 运行。每个 plugin 的 `config` hook 仍然都运行，所以另一个 plugin 依然可以往 `config.provider[id]` 里加模型或名字。

在此之前，落败的那个 plugin 会悄无声息地消失：它那一行显示「登录到星鸣用不了的东西」，而 `keepUnloaded` 可能把同一个 id 列出两次，一次来自旧的 `plugin-providers.json`（Discord 上的 neiko）。现在：

- `init` 返回每个 plugin 的 `provides`（它的 auth hook 点名的 id）和 `servedBy`（id → 另一个 spec，对应它不服务的每个 id）。
- `plugin.Clashes` 把这些整理成每个 spec 的 `{id, by, with}`。
- `/api/plugins` 给每一行它的 `clashes`，以及 provider 列出的名字。
- 当现在有别的 plugin 服务某个 id 时，`keepUnloaded` 不再保留此前的那个 provider。

两行在插件页上都保持可见：

- 提供服务的那一行写着「{other} 也登录到 {name}；这个由它来服务」。
- 另一行以琥珀色写着「{name} 由 {other} 提供服务，后者也登录到它」。它有一个 **用于 {name}** 按钮，调用 `POST /api/plugins/prefer {spec, provider}`（`plugin.Prefer`）并重启宿主。

在 CLI 里，`starling plugin list` 打印同样那两行，`starling plugin use <name> <provider>` 做出选择。

存储层如何处理选择和名字：

- 移除一个 plugin 会丢掉它的选择。
- 关闭一个 plugin 会把 id 交给另一个 plugin，直到它被重新打开。
- 以新版本或新 spec 重新添加一个 plugin 会保留它的选择。
- 移除、关闭、选项和 prefer 都先按精确 spec 找 plugin，再按包名找（`store.go` 里的 `indexOf`）。这样对一个 plugin 的操作绝不会波及另一个同名包的 plugin。

## 归属如何变化

迁移记录有五种状态：

| 状态 | 含义 |
| --- | --- |
| `moving` | 迁移正在进行或被中断。显式重试会先回滚被中断的那次迁移。 |
| `plugin` | plugin 拥有这个内置 provider id。这是 `Moved` 唯一认的状态。 |
| `back` | 较旧版本的星鸣把归属还给了内置实现。`MoveRetiring` 会再迁一次。 |
| `failed` | 迁移失败，内置实现继续服务。自动重试等待六小时。 |
| `removed` | 用户移除了已迁移订阅的 plugin。它的账户仍留在 plugin 的登录下，记录也留着它们。没有任何东西会自行迁移或安装它；重新安装该 plugin（`HandOver`）或编辑器里的迁移（`Adopt`）会让它重新变成 `plugin`，连同记录一起。 |

编辑器里的「迁移到 plugin」动作调用 `provider.Adopt`。有账户时，adopt 委托给 `Move`，它会安装 plugin、导入每个账户并落实 plugin 归属。每个账户都通过 plugin 试一遍，由于内置实现已退役，没有任何账户会挡住这次迁移：厂商拒绝的账户会带着「已失效」的标记一起过去，本来就已失效或因其他原因检查失败（厂商不可达）的账户则未经尝试一起过去，交给 plugin 自己的读取去判断。内置实现曾服务的任何模型都不会拿去和 plugin 的列表比对。只有安装失败、plugin 不服务该 id，或某个登录 plugin 不肯导入，才会让迁移失败，这时会把一切放回去（`putBack`），六小时后再试一次。较旧版本的星鸣在失败迁移时留下的 `why`（`lapsed`、`unserved`、`account`）仍会被读出并由页面翻译。没有账户时，adopt 只验证包和它服务的 provider，然后记录 plugin 归属，不做账户检查。

单独安装该 plugin 也可以通过 `provider.HandOver` 交出归属。这个包必须处于启用状态、达到 mover 的最低版本，并服务预期的 provider。添加、更新或升级之后，GUI 会调用 `HandOver(ctx, false)`，它可以接管一个没有账户的内置实现。已存在的内置账户要等网关维护时的 `HandOver(ctx, true)`。

`KeepRetiringMoved` 首次在 20 秒后运行，之后每小时一次；如果某个已退役内置实现的失败迁移到了重试时间，则更早运行（`nextRetiringRun`）。每个周期维护已迁移 plugin 的版本、调用 `MoveRetiring`，然后执行自动交接。`MoveRetiring` 会迁移每一个有 mover（`retiring`，即 `MovableIDs`）且还有内置账户的订阅。一次失败的尝试要等六小时（`moveRetry`），内置实现已退役时等十分钟（`retiredRetry`、`retryWait`），因为在迁移之前没有东西服务那些账户：因为 npm 不可达或 plugin 尚未安装而失败的迁移，很快会再试一次。内置实现正在被逐一退役（负责人，10-16：陆续把内置Provider下线掉），所以 `back` 的记录也会再迁一次；`removed` 的则留着，直到用户安装该 plugin。交接会把插件已安装且为最新版的 `removed` 记录变回 `plugin`，保留一个已用独立 `-plugin` id 登录的 plugin，并跳过已有的迁移记录，只有一个 `failed` 记录老到可以重试时才动它。

没有往回迁：`provider/moveback` 和 `starling plugins move-back` 都回答内置实现已退役，移除或关闭 plugin 也不会把任何东西迁回去。关闭时，plugin 仍把订阅保持在 `plugin`，只是它自己也关着。被移除（`RemovePlugin`）时，订阅被记为 `removed`，它的账户留在 plugin 的 auth 存储里，因为 `plugin.Remove` 会保留登录，直到用户登出。插件页和 `starling plugins rm`/`off` 会说明该订阅停止工作、账户被保留。回滚一次失败的迁移（`putBack`，经每个 mover 的 `back` 和 `settle`）是唯一会写内置实现登录的路径。

移除一个 plugin 还会把它那些 provider 留在 `providers.json` 里的模型选择另行保存（#2057，`plugin_left.go`）。以前没有任何地方列出这些记录，`starling provider rm` 也找不到它们，于是同步和备份每次运行都要求为它们登录（`NotSignedIn`）。

- `RemovePlugin` 在该 plugin 仍装着的时候读取它的 provider 有哪些 id（`pluginRecordIDs`：每个 provider 上 `plugin.Listed` 最后为它那个 spec 返回的 `PluginID`，关着的 plugin 也算）。它会略去：
  - 内置实现的 id；
  - 另一个已安装 plugin 也服务的任何 id。
- `plugin.Remove` 之后，只有那些 id 下仅含选择的记录（`OnlyPicks`）会移进 `providers.json` 的 `pluginLeft`，放在包名之下。带端点或 key 的记录是用户自己的 provider，保留不动。
- 已知没有任何列表时（既无缓存，也读不到 `plugin-providers.json`），什么都不会另行保存。
- `provider.AddPlugin` 是 GUI、CLI 和 `AddPlugins`（备份与同步）添加时走的入口。它把该包的记录放回去，每个 id 当前没有记录的才放。此后用户又给了某个 provider 的记录则仍另行保存。
- 关闭一个 plugin 不会另行保存任何东西，因为重新打开就能撤销。这次改动之前移除留下的记录保持原样；它们无法和同步为一个尚未在此安装的 plugin 带来的选择区分开。

新用户不会在添加面板里看到没人用的已废弃内置实现。`unusedSub` 在它没有账户时把它略去。在安装好的替代品接管归属之前，`replacedSub` 让添加面板显示该 plugin 的卡片。插件页是安装入口。

`keepMovedCurrent` 为已迁移的订阅维护所需的 plugin 版本。在发布 plugin 修复之后调高所属 mover 的 `min`，就会触发这条更新路径。更新失败会记入日志；被禁用的包和从本地目录安装的不会被自动替换。

编辑器里的*更新*按钮和*全部更新*安装 npm 上的最新版本；用户设的固定版本会保留到他自己更新为止，更新后就不再固定。`keepMovedCurrent` 是唯一一条不用用户按任何东西就会行动的路径，所以 `installPlugin` 保留用户的选择：固定在低于该次迁移 `min` 的 npm 版本不会被重装，迁移会说明它被固定了、需要哪个版本；git 安装的按它自己的 spec 重装；没固定的 npm 包走 `plugin.Upgrade`（npm 上的最新版本；bun 会说明为什么在 `minimumReleaseAge` 下装不了，而 `Add` 的 `@latest` 会悄悄给出旧版本）；用户关掉的 plugin 不动。

npm 上没有的 `@magpie-community` 包从社区仓库的 `dist` 分支安装（[`dist.go`](../../internal/plugin/dist.go)）。没有所有者的 2FA，npm 无法接受一个新包的首个版本（Trusted Publishing 只能为已存在的包配置），于是审核过的 plugin 一直没能合并，等着有人去发布（magpie-community/plugins#84）；2026-09-03 之后创建的 Trusted Publisher 只允许发布到暂存区，zed 0.1.7 和 zen-free 0.1.7 就这样被拒了。现在仓库的 `publish.yml` 会把每个合并的版本打包到 `dist` 上（`scripts/dist.mjs`：`tgz/<file>` 和带每个版本 sha512、清单与 README 的 `index.json`），不需要任何人的凭据，npm 也发布到它旁边。`install` 和以前一样先问 npm；只有在 `bun add` 对某个社区包的 `latest` 或指定版本失败时（`distWanted`），且失败原因不是用户的 `minimumReleaseAge`（那个拦截保留），`installDist` 才去取 tarball，除非它的 sha512 与索引里的一致否则拒绝，把它留在 `plugins/dist` 并执行 `bun add ./dist/<file>`。之后从 npm 的安装会取代它，`pruneDist` 会删掉 `package.json` 里不再提到的 tarball。`npmAsk`（市场里的版本、*更新*、每小时检查）在 npm 出错或落后时取 `dist` 上的最新版本，`Readme` 则回落到它自己的 README。开启「国内镜像」开关时通过它的 jsDelivr 副本读取该分支（`fetchJSONFaithful`），否则走 raw.githubusercontent.com。[`dist_test.go`](../../internal/plugin/dist_test.go) 和 [`update_held_test.go`](../../internal/plugin/update_held_test.go) 里的 `TestDistInstallsAPackageNPMDoesntHave`、`TestDistServesAVersionNPMRefused`、`TestDistRefusesATarballNotItsIndexs` 和 `TestUpdateBunHoldsBackSaysWhy`（让 `dist` 持有被拦下的版本）覆盖了这些。

列表上有一个包不在 `plugins/` 里的 plugin，会由 `plugin.InstallMissing` 按它自己的 spec 重装，不管它是开着还是关着。`KeepUpdated` 在启动 30 秒后运行它，之后每小时一次，在 `CheckUpdates` 之前。以前没有别的东西做这件事。一次更新会跳过没有已安装版本的 plugin，于是宿主每次启动都报 `no such plugin: …\node_modules\…`，直到用户移除该 plugin 再重新添加（#2002，Windows，更新之后）。每一条会改动 `plugins/` 的 bun 命令都在 `installMu` 下一次只跑一条：`install`（连同它写的第一个 `package.json`）、`reinstall` 的 `bun update`，以及 `Remove` 和 `add` 的 `bun remove`。启动时 `KeepUpdated` 和 `keepMovedCurrent` 都会跑 `bun add`。两个同时跑会各自在对方写入之前读完 `package.json`，于是有一次改动丢失（在 Windows 上用 Bun 1.4.3 复现），而且依赖被删掉的 git plugin 就找不到了。[`plugin_missing_test.go`](../../internal/plugin/plugin_missing_test.go) 里的 `TestMissingPluginsInstalledAgain` 和 `TestBunCommandsOneAtATime` 覆盖了这两点。

用户关掉的 plugin 在每条安装路径上都保持关闭：`add` 替换已有条目时保留 `Off`，所以添加、更新、全部更新和全部升级都按用户留下的样子对待它。

## 修复时的约束

先确定归属，再选修复位置。影响已迁移订阅的 bug 通常需要改社区 plugin。只改内置实现够不到已迁移的用户。

修 plugin 时，发布修正后的社区包，并把所属 mover 的 `min` 提到那个版本。当修复必须覆盖尚未迁移的用户时，内置实现也要一起改。

排查故障时，把迁移、宿主和上游行为分开看。plugin 的 bug 可能需要改社区代码；宿主、provider 卡片、登录编排或迁移的 bug，即使对已迁移的用户也可能需要改星鸣本身。

宿主拿到的只是星鸣以 `STARLING_*_PROXY` 形式给的代理，并把它自己应用到每一次 fetch（`proxyFor`），以及 plugin 打开的每个 HTTP/2 会话（`http2.connect`、`tunnel`，#1035）。处在用户自己网络上的宿主会绕过该代理，就像星鸣自己的请求那样（`host.js` 里的 `onLAN`，是 `internal/netproxy/lan.go` 那套的一个版本：私有地址、链路本地地址、回环地址或 CGNAT 地址，`.local` 名字，或每个地址都是上述之一的域名，最多查 400 毫秒并保留一分钟；#1572）。provider 自己的代理仍照原样使用。这两条规则保持一致：改一边也要改另一边。plugin 启动的程序，比如 Grok plugin 的 `grok login`，会从 `host.js` 的 spawn 包装器那里以 `HTTPS_PROXY`/`HTTP_PROXY` 拿到该代理，除非 plugin 自己设了一个。内置实现里能登录、经 plugin 却失败的登录，可能就是这样一处宿主差异，而不是 plugin 的 bug。

星鸣会下载适合本机的 Bun 构建（[`bun.go`](../../internal/plugin/bun.go) 里的 `bunTarget`）：x64 Linux 和 Windows 上是 baseline 构建，ARM 上的 Windows 是原生的 `bun-windows-aarch64`（在 #2183 之前它因为没有对应构建而被拒绝），musl Linux（比如 Alpine，`libc.Musl`，Claude Code 的下载也用同样的检查）则是 musl 版，即 `bun-linux-x64-musl-baseline` 或 `bun-linux-aarch64-musl`。glibc 构建会带上 glibc 的加载器，而 Alpine 没有它，所以在那里启动不了，exec 报「no such file or directory」（#1877）。musl Linux 把它的 Bun 和 `state.json` 放在 `bun/musl` 下，所以 C 库变过的机器不会去跑另一个的构建，也不会取用另一个另行保存的版本；其他地方它们留在 `bun` 里，之前下载的 Bun 也在那里。磁盘上有一个启动不了的 bun 时，它会说自己是为另一种 C 库构建的，并指出 `STARLING_BUN`（`cantStart`），既不回退也不另行保存：前一个版本是同一个构建。Bun 的 musl 构建需要 libstdc++ 和 libgcc；在缺少它们的 Alpine 上，`bun add` 和宿主的启动会指出缺哪个库，以及 `apk add libstdc++ libgcc`（`missingLibs`）。[`bunmusl_test.go`](../../internal/plugin/bunmusl_test.go) 里的 `TestBunTargetForMusl`、`TestMuslBunKeptApartFromGlibc`、`TestCantStartSaysCLibrary`、`TestHostOnUnstartableBunSaysCLibrary`、`TestMissingLibsNamesThem` 和 `TestHostOnBunLackingLibsSaysWhich` 覆盖了这些。

Bun 只接受 http:// 和 https:// 代理。SOCKS5 代理，比如 Mac 的系统 SOCKS 代理，会以一个前置的回环 HTTP 代理提供给星鸣启动的每个 bun（`netproxy.Bridge`）：通过 `hostEnv` 给宿主，通过 `bunCommand` 给 `bun add`、`bun update` 和 `bun remove`（`netproxy.EnvForBun`）。以前，只有 SOCKS 系统代理时的安装会以 `UnsupportedProxyProtocol` 失败（#1409）。由 bun 安装的 Agent CLI 也通过这个桥更新（`internal/agent/cliupdate.go` 里的 `updateEnv`）。用 Bun 构建的 Claude Code 在 Claude 订阅的运行和登录时会拿到它（`netproxy.EnvForBunWith`，#2651；见 Claude 订阅桥）。

宿主在启动时读取一次它的代理。当星鸣会给新宿主的代理与正在运行的宿主拿到的那个不同时（例如登录时 Clash 还没起来而设了系统代理，或者设置变了），下一次调用宿主就会替换它（[`host.go`](../../internal/plugin/host.go) 里的 `proxyMoved`，最多每 15 秒查一次；宿主没有代理时每 2 秒查一次：`proxyLookNone`，#1518）。为没人要求的读取发起的调用（额度只在被要求时才读，`netproxy.Held`）不会到达宿主。在此之前，没有代理启动的宿主就一直没有代理，于是 Grok plugin 的 `grok models` 无法续期它的 token，每次重启后账户都显示为已登出（#1363）。

plugin 用 `console.log`、`console.warn` 等写出的内容会进宿主的 stderr，星鸣逐行记录它（[`host.go`](../../internal/plugin/host.go) 里的 `drainStderr`）。超过 4 MiB 的一行会终止记录：星鸣会记下它终止了，然后把剩下的读完丢掉，这样宿主不会因为管道写满而阻塞。以前出现这种行之后没人再读管道：在 Windows 上宿主下一次写入时会挂住，在 Linux 上 Bun 会把之后写入的一切都留在内存里。

plugin 的某个账户若被厂商拒绝登录（`lapsed`，或者读取额度时得到这样的说法），它那一行仍保留「重新登录」和「移除」，无论它是在用还是 Agent 自己的账户（`app.js` 里 `renderAccounts` 的 `signedOut`）。

plugin 登录页面只会被打开一次。`authorize` 的结果可能带 `opened: true`（这不是 OpenCode 的字段）：说明 plugin 自己的 CLI 已经打开过该页面。Grok 的 `grok login --device-auth` 在 macOS 和 Windows 上就是这样，而那里没有东西会关掉它，星鸣再去打开就会看到两个页面（X 上的 andyfmyth；plugin 0.1.17，在其他系统上则给 grok 一个什么也不打开的 `BROWSER`）。`host.js` 把它作为 `Authorization.Opened` 传下去，再变成 `SignInState.Opened`，于是桌面端的 `/api/plugin-signin`、`starling plugin login` 和在浏览器标签里打开的窗口（`startPluginSignIn`，以及安装后的 `followSignIn`）都不再打开任何页面；页面链接仍然显示（`TestPluginSignInPageOpensOnce`、`plugin-signin-opened.test.cjs`）。

GUI 的 plugin 集成包含 `subOf`、`pluginSubs` 和 `startPluginSignIn`。对登录或 provider 展示的改动，除了它影响的内置路径，也必须沿着这些路径走。

## 验证

<a id="verification"></a>

plugin 也可以声明一个与 provider 无关的 `starling.hooks` 入口，包括纯网关包和 provider/Pi 混合的包。它的异步请求契约和验证见[网关请求 hook](gateway-request-hooks.md)。已有的 provider hook 仍走它们平常的入口路径。

[`migrate_notice_test.go`](../../internal/provider/migrate_notice_test.go) 里有 `TestMovedBuiltinsSayTheirPlugin`，检查已迁移的内置实现会告诉贡献者由哪个 plugin 服务。迁移测试放在 [`migrate.go`](../../internal/provider/migrate.go) 旁边。

网关的等价性测试在 [`plugin_parity_test.go`](../../internal/gateway/plugin_parity_test.go) 和其他 `plugin_*_test.go` 文件里，用来比较或演练 plugin 路径。只靠一个内置实现的测试并不能证明已迁移用户的行为。要检查每个测试的夹具，确认它覆盖了被改动的 provider 和操作。

[`plugin_decide_test.go`](../../internal/gateway/plugin_decide_test.go) 用带决策 API 的假 plugin（`FAKE_DECIDE`）运行：它标记的模型和 `jev-*` 模型是分类器，不提供给 Agent；向其聊天会被拒绝；一个 group 保存其一作为分类器，而不是把它和聊天模型并列；单独由决策模型组成的 group 也是如此；每个分类器的问题都通过 plugin 的 fetch 到达 `/v1/systemone` 上的假厂商。用 `FAKE_DECIDE=models` 时，没有 `decide` 的标记不起任何作用。

[`plugin_checkin_test.go`](../../internal/provider/plugin_checkin_test.go) 让假 plugin 的 `auth.checkin`（`FAKE_CHECKIN`）按计划跑一遍：关闭、各个 outcome、同一天不按第二次、失败只在重试窗口之后重试、第二天、卡片与设置，以及厂商签到让位；`TestPluginCheckinTellsSignInsOfOneNameApart` 用两个名字登录四个账户、每个名字两个（plugins#254），检查每个都被签到、被保留、出现在卡片上并能单独关闭；`TestPluginCheckinCarriesOverDisplayNameEntries` 则升级一个文件和一个按显示名保留的 `Off`。`plugin-checkin.test.cjs` 覆盖卡片那一行和设置标签页。

[`plugin_options_test.go`](../../internal/plugin/plugin_options_test.go) 在真实宿主里加载 `testdata/options`：哪些声明的选项被保留、它们如何被限长、启动时用户的值优先于默认值，以及 `SetProviderOptions` 之后 plugin 能在不重启的情况下读到改动。[`plugin_option_api_test.go`](../../internal/gui/plugin_option_api_test.go) 覆盖该接口、非法的 key，以及设置保存时保留数值。`plugin-options.test.cjs` 在 Chromium 和 WebKit 里覆盖编辑器，宽 1100 和 420，所有语言都测。

[`plugin_cliproxy_test.go`](../../internal/plugin/plugin_cliproxy_test.go) 的 `TestPluginHostTakesANewProxy` 启动一个没有代理的真实宿主，设上一个代理，并检查 plugin 启动的程序能拿到它。[`plugin_lan_test.go`](../../internal/plugin/plugin_lan_test.go) 的 `TestPluginLeavesTheUserNetworkDirect` 用真实宿主检查：在设了代理的情况下，对本机私有地址的 fetch（http:// 和 https://）以及 HTTP/2 会话都直连到达，而公网地址和无法解析的名字仍然走代理。`plugin-signed-out-account.test.cjs` 覆盖已迁移的 Grok 账户在使用中以及作为 Agent 自己的账户时被登出的情形，在 Chromium 和 WebKit 里，英文和中文，宽 420 和 1100。

[`stderr_drain_test.go`](../../internal/plugin/stderr_drain_test.go) 的 `TestDrainStderrKeepsTheChildGoing` 检查子进程往 stderr 写 8 MiB 且没有换行时仍能结束，并检查日志说明它的 stderr 被丢弃了。[`stderr_e2e_test.go`](../../internal/plugin/stderr_e2e_test.go) 的 `TestHostStartsPastAnOverlongStderrLine` 启动一个真实宿主，它的 plugin 在加载期间记了一行 5 MiB 的日志，检查宿主能启动；没有这个修复时它只在 Windows 上失败。

这些测试覆盖一个 provider id 上有两个 plugin 的情形：

- [`plugin_clash_test.go`](../../internal/plugin/plugin_clash_test.go) 的 `TestTwoPluginsOneProvider` 在真实宿主里加载两份假 plugin。它检查 `provides`/`servedBy`、`Clashes`、该 id 只被列一次，以及这个选择能挺过 Prefer、重启、关/开和移除。
- `TestPluginNamedExactlyFirst` 和 `TestKeepUnloadedLeavesAServedID` 覆盖存储层和 provider 缓存。
- `TestPluginsClashRows`（`internal/gui`）和 `TestPluginListSaysClash`（CLI）检查用户看到的东西。
- `plugin-clash.test.cjs` 覆盖两行和「用于」，在 Chromium 和 WebKit 里，英文、中文、繁中、日文和德文，宽 1100 和 560。

[`quotas_dedupe_test.go`](../../internal/provider/quotas_dedupe_test.go) 检查无关厂商的重置冲突和 ZCode 的去重，包括两个 plugin id 的用量窗口换算。[`TestPlanQuotas`](../../internal/provider/planquota_test.go) 检查自定义 GLM 端点、缓存的卡片，以及从磁盘恢复的兜底读取。这些夹具不联系真实厂商，也不运行社区的 ZCode plugin。

用隔离的主目录运行相关的包测试。在改 HOME 之前先解析 GOPATH 和 Go 的缓存，这样多次运行能复用依赖和已下载的工具链。子 shell 让临时文件可写，并在退出时清理它的主目录（测试失败时也一样），并返回测试命令的退出状态。

```sh
(
	verification_home=$(mktemp -d) || exit
	trap 'chmod -R u+w "$verification_home"; rm -rf "$verification_home"' EXIT
	verification_gopath=$(go env GOPATH) || exit
	verification_modcache=$(go env GOMODCACHE) || exit
	verification_buildcache=$(go env GOCACHE) || exit
	HOME="$verification_home" GOPATH="$verification_gopath" GOMODCACHE="$verification_modcache" GOCACHE="$verification_buildcache" \
		go test -tags nogui ./internal/provider ./internal/gateway ./internal/plugin
)
```

有些迁移和网关 plugin 测试要求 PATH 上有 Bun，并使用本地 plugin 夹具；Bun 不可用时它们会跳过。要报告任何被跳过的测试，并在社区包的实现有变动时单独验证它。
