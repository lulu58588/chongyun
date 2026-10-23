# Agent 接线

<a id="agent-wiring"></a>

星鸣把一个 coding agent 连到自己的网关上，方式是改这个 agent 自己的配置。
这覆盖了 endpoint、key `starling-<agent>` 和模型，用 agent 自己的格式。当 agent
处于断开状态时，星鸣把它原先的样子放回去。包 `internal/agent` 描述了每个
受支持的 agent：它的配置在哪里、哪些字段重要、要提供哪些取值。它还记下星鸣
科技写过什么，这样之后由别的东西做出的改动就会显示为 drift。

## 职责与事实来源

<a id="responsibilities-and-sources-of-truth"></a>

| 部分 | 职责 | 来源 |
| --- | --- | --- |
| Agent 描述 | `Agent` 有一个 id、name、icon，用于检测的 `Bin` 和 `Dir`，以及 `Path`、`Fields` 和 `Native`。它还有一些可选的 hook：`Notice`、`UA`、`ListsModels`、`Sync`、`Unwire`、`Join`/`Joined`、`Routed`、`Follow`、`RenameRefs`、`Check`、`LastUsed`、`Reached`、`Import`/`Added`、`Launch`、`SplitSuffix`、`Spelled` 和 `ModelAliases`。每个 hook 的注释说明哪些 agent 需要它。 | [`agent.go`](../../internal/agent/agent.go) |
| 字段和选项 | 一个 `Field` 有 `Get`、`Set` 和 `Options`。`Set("")` 把字段恢复成 agent 的默认值，并把星鸣的接线撤掉。`Quiet` 字段在被设置之前跟随另一个字段（`Follows`，比如 Claude Code 的 tier）。一个 `Option` 带着选择器所需的事实：`Direct`、`Via`、`Same`、`Alias`、`Rate` 和 `Context`。GUI 和 TUI 的选择器列出 `Agent.Picks`：当该 agent 在设置的 `ownModelsHidden` 里时（模型列表顶部的开关，#1642），一个提供目录模型的字段（`Ref`）会把 agent 自己的模型排除（没有 `Ref`），但它自己的默认值和当前被设成的值除外。CLI 仍然可以按名字取 agent 自己的模型。 | [`agent.go`](../../internal/agent/agent.go), [`route.go`](../../internal/agent/route.go) |
| agent 列表 | `All` 列出星鸣知道的每个 agent，然后是 WSL agent（`wslAgents`）。`Clients` 列出网关按名字知道的每个 agent，也就是 `usage.Agents` 读的那份。每个 agent 一个文件：`claude.go`、`codex.go`、`agents.go` 里的 `opencode`，等等。 | [`agents.go`](../../internal/agent/agents.go) |
| 连接 | `ConnectHow` 选出一个 agent 启动时用的模型并说明是怎么选的。`How` 是其中之一：`kept`、`again`、`joined`、`starling`、`same`、`alike`、`default` 或 `first`（#726）。没有模型可给时，它返回 `NoModelsError`（错误码 `no_models`）。 | [`applied.go`](../../internal/agent/applied.go) |
| 星鸣写了什么 | `Apply` 设置一个字段并把它记进 `applied.json`。`Wired` 说明星鸣是否在配置里。`Drift` 报告 `unwired`、`replaced`、`unreachable`、`proxied` 或 `bypassed`。`Reapply` 再写一次那份记录。`Keep` 接受现在的配置原样。 | [`applied.go`](../../internal/agent/applied.go) |
| 之前是什么样 | stash（`stash.json`）记着在星鸣把配置指向网关之前它是什么样子，这样断开时是恢复而不是猜测。 | [`stash.go`](../../internal/agent/stash.go) |
| 断开与预览 | `Disconnect` 先跑 `Unwire`，再把每个仍然指向星鸣的字段恢复成默认值。最后它跑 `Restore`：Pi 及其同类（`piLike`：Pi、OmO、Prime Agent；以及 MothX，它只有一份 `settings.json`）会在星鸣第一次写入之前把 `settings.json` 和 `models.json` 复制一份（`*.before-starling`、`beforeStarling`），并在断开留下的内容与那份 JSON 相同时逐字节放回；原本不存在的文件在只剩下空对象时也重新删掉。`DisconnectPreview` 在一个临时 home 下的副本上、在一个子进程里（`DryRunArg`）跑真正的 Disconnect，然后对文件做 diff。预览绝不是对 Disconnect 行为的第二份记账。 | [`applied.go`](../../internal/agent/applied.go), [`preview.go`](../../internal/agent/preview.go) |
| 让配置保持最新 | `Reseat` 把 agent 从星鸣不再提供的模型上挪走（#200）。`OnGateway` 和 `Rewire` 在端口变化时把新的网关地址写进每个已连接的 agent。`Source` 和 `Stale` 说明未连接的 agent 跑在什么上，以及正在运行的副本是否还在用旧列表。 | [`reseat.go`](../../internal/agent/reseat.go), [`rewire.go`](../../internal/agent/rewire.go), [`connectinfo.go`](../../internal/agent/connectinfo.go) |
| Codex 后台压缩 | 长对话的 Codex 后台压缩，在 Agents 页面 Codex 打开的那一行里（`CodexAsyncCompact`、`CodexAsyncCompactPercent`，#2490），会在每次 `Sync` 时把 `[features.async_compaction]` 的 `enabled = true` 和 `model_async_compact_threshold_percent`（Codex 0.163 只在那张表里读它）写进星鸣路由到的每个 Codex 配置。Codex 只会在 `capabilities.remote_compaction` 为 `"v2"` 的 provider 上启动后台压缩，而它只给 OpenAI 和 Azure 这个能力，所以 `[model_providers.starling]` 也要拿到 `capabilities = { remote_compaction = "v2" }`，即使用户是手动打开这个功能的（#2493）；`putProvider` 再次写入那张表时会把它的 capabilities 带过去。星鸣写了什么记在 stash 的 `codex.async` 下：用户设过的值，或者之后改过的值，绝不会被覆盖写，切换关闭时也留着；一个单纯的 `async_compaction = true`、内联表或点号表是用户自己写的原样保留。 | [`codex_async.go`](../../internal/agent/codex_async.go) |
| 具名的 omp profile | omp 把一个 profile 的设置放在 `~/.omp/profiles/<name>` 下（或者 `PI_CONFIG_DIR` 下），每个都有自己的 `config.yml` 和 `models.yml`。每个这样的 profile 都是一个独立的 agent —— 在 Windows 上是 `omp#<name>`，在发行版里是 `omp#<name>@wsl:<distro>` —— 所以每个都能单独连接、单独看 drift、单独断开。它们共用 omp 的模型列表（`ListsFor`），而默认的 `omp` 行就是 `ompDir`/`ompIn` 已经选中的那个目录，所以一个该行已经指向的 `OMP_PROFILE` 不会被列两次。omp 拒绝的名字不成行：`default`、以点结尾，或是 Windows 保留设备名（`aux`、`nul`、`com1`、`lpt9.x`、`con.work`），在所有平台上都一样（`ompProfileOK`）。一个发行版的 profile 名字来自它的探测，保存在 `wsl.json` 里；一个停着的发行版绝不会被启动来列这些名字，它的 profile 行显示上次看到的东西（`asleep`）。 | [`omp.go`](../../internal/agent/omp.go), [`wsl.go`](../../internal/agent/wsl.go) |
| WSL 孪生 | 在 Windows 上，每个运行中的发行版被探测一次，不会启动停着的那个。在那里找到的 `wslKinds` agent 会变成各自的 agent `<id>@wsl:<distro>`，通过 `\\wsl.localhost\<distro>` 编辑。一种 kind 通过它在发行版 home 下的目录、或通过它在那里的命令来找到（不是 Windows 盘上的命令）。当另一个工具共用其中某个目录时，这种 kind 会被排除：Gemini CLI 只通过它的命令找到，因为 Antigravity 也占着 `~/.gemini`。Goose 只通过它的目录找到，因为 `goose` 命令可能是 pressly 的迁移工具。Reasonix 还需要它的命令给出的版本号（`accept`），因为旧的 npm 客户端也叫 `reasonix`。网关地址和 key 跟着发行版的网络走：镜像或 NAT（`place`、`keyAt`）。`ListsFor` 把孪生体的模型选择保存在 Windows 孪生体的 id 下（#927）。 | [`wsl.go`](../../internal/agent/wsl.go), [`gatewaykey.go`](../../internal/agent/gatewaykey.go) |
| 按使用情况调优的设置 | Usage › Context 的 "Tuned for you" 会按 agent 给出建议：在该范围内，用户自己的调用里哪些设置能花最少的 token，点一下就应用其中一个。包 `tune` 会用每个候选值（`Compact`、`CacheTTL`）重放 `sessions.Calls` 从各 agent 的 session 文件里读到的调用；`tuneOf` 挑选 agent 和模型（Claude Code 按模型，Codex 合并计算）。`TuneRead`、`TuneApply` 和 `TuneUndo` 知道每项设置放在哪里：Claude Code 的 `modelSettings.<model>.autoCompactWindow`（它的 `/autocompact` 存在那里）、`~/.claude/settings.json` 里的 `promptCacheTtl` 和 `subagentPromptCacheTtl`；Codex 的顶层 `model_auto_compact_token_limit`。建议只按 token 权衡：它被标注为 "Fewest tokens"，而它会改动的压缩窗口会说明它并不衡量 agent 在压缩之后把任务保持得有多好，以及更低的窗口会压缩得更频繁（#1854，`context-tune.test.cjs`）。一个压缩行的 "Other window…" 会打开应用菜单，里面是曲线上的每个取值，每个都带上相对最少者估算出的额外 token，以及它大概会造成多少次压缩（`Point.compacts`，来自同一次重放），并像 Use 那样把选中的那个应用上（#1860）。一个其文件给压缩计时的 agent，它的等待时间也被纳入权衡（HirotaNagai 报告）：Claude Code 的 `compact_boundary` 行给出 `compactMetadata.durationMs`，`sessions` 把它放在其后的第一次调用上，而 Codex 是用它自己的一次调用来压缩的（`compacted` 行指名的记录，2026-09 之后的 rollout），那次调用的时间就是压缩的时间（`Call.Compact`；更老 rollout 的压缩没有对应调用，就不计时间）。当一个模型有 3 个或更多计过时的（`minTimed`），`Facts.took` 是它们的中位数秒数，每个窗口给出自己的等待（`Point.wait`，它的压缩次数 × `took`），这一行会说明一次压缩花了多久、建议相对于当前窗口的等待，以及建议为了少等一会儿而跳过的那个更便宜的窗口（`Compaction.cheapest`、`wait_cheapest`），当它能省下 3% 或更多时；建议被标注为 "Tokens + wait"，它的注意事项说明只涉及 token 和等待，而选择器给出每个窗口的等待，最便宜的那个标着 "Fewest tokens"（`context-tune.test.cjs`）。计时读数更少的 agent 或模型则像以前一样给建议。 | [`internal/tune`](../../internal/tune/tune.go), [`tune.go`](../../internal/agent/tune.go), [`internal/gui/tune.go`](../../internal/gui/tune.go) |
| plugin 添加的 agent | 一个 plugin 的 `starling.agent` 模块（或一个 `*.agent.js` 文件）说明星鸣自己没有设置的 agent 把模型放在哪里（`config`、`format`、`model`、`prefix`），它的 `connect` 返回要写的其他键。没有 `model` 的只用于连接（#2374）：它唯一的字段是 `provider`，要么是星鸣要么为空，就像一个只有"provider 是星鸣"这一项设置的应用；`connect` 在没有模型的情况下被调用，必须返回键，并且目录变化时会再被调用一次（`Sync`）；`ListsModels` 被设置了，所以它的列表就是为它挑的模型。`agentplug` 用 moejs 编译每一个（通过 `jsmod`，和网关 middleware 一样），在 plugins.json 或该文件变化时重新加载它，并保留某个没能加载的原因。`plugged` 把每一个做成一个 `Agent` 并设好 `Plugin`，列在星鸣自己的 agent 之后；一个与星鸣自己某个 agent 同名的 id 会被拒绝（`PluginErrors`）。 | [`agentplug.go`](../../internal/agentplug/agentplug.go), [`plugged.go`](../../internal/agent/plugged.go) |
| 只靠环境的 agent | 只从环境里读网关的 agent 会得到一条 `Launch` 命令。Cursor Private Inference 则得到用户级变量（`cursorLocalOn`、`KeepCursorLocalEnv`），而星鸣不写 Cursor 自己的任何文件。它通过 apps 文件夹里的 product.json 找到，通过 Windows 已安装程序列表点名的那个目录找到（`cursorLocalInstalls`，安装器里选的任意目录），或者作为它的 Linux AppImage（`Cursor_Private_Inference-…AppImage`）在下载目录里找到（#1254）。它的 effort 字段保存在星鸣里（`provider.AgentEffort`）并由网关来问，它的等级就是它列出的那些模型的等级（`cursorLocalEfforts`，#1003）。 | [`cursorlocal.go`](../../internal/agent/cursorlocal.go) |
| Claude Science | 它的守护进程（`claude-science`）只从自己环境的 `ANTHROPIC_BASE_URL` 读 API 地址，没有对应的设置项。连接会用该变量指向网关来重启守护进程（`scienceOn`：先 `stop`，然后在 `status` 说它所在的端口上跑 `serve --detached --no-browser`），并把守护进程记在 `claude-science.json` 里（url、pid、port）；断开则不带该变量重启它，星鸣自己环境里的那个变量也会被去掉。不会为用户设置任何变量，也不会写这个 app 的任何文件。在星鸣的守护进程跑在这个网关上时再设置一次没有任何效果（`Pick` 在 `connect` 之后设置字段）。`KeepClaudeScience` 在它没在运行时，在网关上、按它自己的端口把它重新启动（电脑重启之后），并在星鸣提供服务期间每分钟做一次；它从不停掉正在跑的那个。`Check` 读守护进程的 `operon.lock`，而不是 `status`（一个 130 MB 的 Bun 程序要好几秒），并把 pid 不同的守护进程调用算作 drift：那个 app 的窗口在没有守护进程运行时打开，会自己启动一个不带该变量的。它在 Mac app 的 bundle 里（`Claude Science.app/Contents/Resources/bin`）、Windows 上的 `%LOCALAPPDATA%\Programs\ClaudeScience`、`~/.local/bin` 里找到，否则在 PATH 上找。这个 app 仍然需要它自己的 Claude 登录；网关从不把登录转交出去（见 gateway-routing）。 | [`claudescience.go`](../../internal/agent/claudescience.go), [`science.go`](../../internal/gateway/science.go) |

## 运行时路径

<a id="runtime-path"></a>

1. **连接。** Agents 页面的开关调用 `ConnectHow`（`POST /api/agents/{action}/{id}`，
   在 [`internal/gui/api.go`](../../internal/gui/api.go)）。它把当前的值存起来
   （`<id>.connect.was`）。它把关闭之前用过的模型拿回来（`reconnect`），或者在
   agent 保留自己模型的地方调用 `Join`（用 ChatGPT 登录的 Codex）。否则它按
   `ConnectHow` 规定的顺序挑一个模型，并通过 `Apply` 设置它。
2. **挑选。** `Pick` 用模型选择器设置一个字段。在一个还没连接的 agent 上挑选
   星鸣自己的模型，会先把它连接上。
3. **提示。** 改动之后，如果该 agent 只在启动时读自己的配置、需要重启，页面就
   显示 `Notice`。标着 `Live` 的字段（`Field.Live`：星鸣自己的设置，网关
   在每次请求时读；Claude Desktop 的 `subagent_model` 和 `subagent_effort`）不
   需要重启，所以 `POST /api/set` 和 `starling <agent> <field> <value>` 在它
   之后就不给提示了，除非这次挑选同时把 agent 连上了（在未连接的 agent 上
   `Pick`）。一个停着的 WSL 发行版里某个字段的副本保留这个标记（`asleep`）。
4. **Drift。** 在 Agents 页面绘制时，`Drift` 把配置和 `applied.json` 做比较。它
   先跑 `Check`，然后逐个字段。通过反向代理指向星鸣的 agent 仍然算在星鸣
   科技上（#1985，davisjiahao：某公司的 nginx 在
   `https://ai.corp.example/starling`）。`wiringOff`——大多数 agent 的 `Check`
   用比较 URL 的方式，而它们自己的检查（Claude Code 的托管设置、Reasonix 的
   provider）在 URL 以星鸣的路径结尾时取的不是星鸣那个，剩下的要么是
   另一台机器上的 http 或 https 地址（`onAnotherMachine`），要么是本机回环上
   带自己路径的一个（`loopbackWithPath`：报告者的合规代理在
   `http://127.0.0.1:8788/claude-code-anthropic/<route>`；`proxyBase`），而那里
   的 `GET <rest>/` 像网关自己的 `GET /` 那样回答（`{"name":"starling"}`，
   `starlingAt` 在 [`reach.go`](../../internal/agent/reach.go) 里，和 `reachProbe`
   一样在后台尝试）。那里是别的东西（CC Switch 放进去的一个中转）就是
   `unwired`，本机回环上一个不带路径的地址也一样——那是旧端口上的星鸣，
   永远不会被尝试。在答案已知之前它不算 drift，所以不会提供 Reconnect，也不会
   把代理写掉；端口迁移（`OnGateway`、`Rewire`）只重新设置 `unwired` 的 agent，
   所以它们也保留它。Codex 靠星鸣的 key 保留这样一个地址，和 #816 一样
   （`codexKeptGateway`：http 或 https、任意端口、星鸣路径之前有路径、
   带路径的回环）。在 Claude Code 里挑选的模型在星鸣在那里回答过之后也会
   保留它（`surelyStarling`；`claude.go` 里的 `routed` 和 `gwAt`），所以每个请求
   都必须走的代理不会被绕过；在那之前，或者在中转上，一次挑选会把 Claude Code
   指向这个星鸣。测试在 `TestMain` 里把 `starlingAt` 打桩成"已知且不是星鸣
   科技"。当配置正确且指向一个回环之外的地址（NAT 下的 WSL 孪生体，或者用户
   保留的地址上的本机 Codex，#816）时，它会尝试那个地址
   （[`reach.go`](../../internal/agent/reach.go)）：在后台，最多每 20 秒一次，
   从不等待。当那里没有回答而网关在回环上有回答时，drift 是 `unreachable`
   （#1013），并说明那里必须有什么在监听：在局域网上共享、用户的 portproxy、
   Windows 的防火墙。然后它检查该 agent 最近一次使用（`LastUsed`、`Reached`）
   是否到达了网关：一次 Codex 日志里说发去了它配置中星鸣所在主机的请求，
   或者网关在之后从它那里看到的（`lastSeen`），都算到达了。
5. **断开。** `Disconnect` 放回 stash 存着的配置，并忘掉那份记录。它还把这个
   agent 标记为已断开（stash 里的 `<id>.disconnected`，`Disconnected`），直到
   用户再次连接它（`connect`，或一个让它保持接线的 `Apply`）：星鸣自己做的
   任何事都不会在此期间把星鸣放回去，启动时的 `Sync` 不是，Codex 的账号
   故障转移不是，#701 的星鸣 API 接线也不是（paynezhuang 报告：Codex
   每次更新后 `openai_base_url` 都回来了）。

## 约束与失败行为

<a id="constraints-and-failure-behavior"></a>

- 星鸣只删除它自己写过的东西。agent 配置里的其他一切保持原样。
- 一次 sync 是把星鸣的模型更新到位，而不是更新 agent 在哪里连星鸣。
  `SyncCatalog` 跑的每一次 `Sync`（星鸣每次启动和更新、新增一个 provider，
  以及一个 plugin 登录、登出或续期，或列出的模型和上次不同：`plugin.OnChange`
  和 `plugin.OnListed` 会调用 `catalog.Touched`，#2373）都走 `synced`，它把
  每个写出的文件过一遍 `kept.away`（`edit.Filtered`，
  [`keptaway.go`](../../internal/agent/keptaway.go)）。这个过滤器在开启期间会
  看到任何通过 `edit` 写出的文件，来自任何 goroutine、在写者持有的任何锁下，
  所以它不持有写者可能持有的任何锁：网关地址和局域网 key 在它开启之前就被
  读走（`keeping`），星鸣自己的文件（`starlingFile`：它原样的目录，或者它
  的链接指向的目录，比如被链进一个 dotfiles 仓库的 `~/.config`）保持原样
  （ylorn 报告：`settings.Save` 在设置锁下去到了一个会读设置的过滤器，
  于是 app 每次启动后卡住约 10 秒；`TestOwnWritesDuringASyncWithALinkedConfig`）。
  在一个不含网关自己地址的文件里，如果用户把星鸣指向了另一台机器上的地址
  （NAS 的，或本机的局域网地址），并且没有 WSL 发行版被看到过的地址，那个地址
  就会留在 `gateway.URL()` 的位置上，连同星鸣每个 URL 在它前面的那段路径
  （`awayPath`：反向代理的 `/starling`；在回环上，是本地代理的路由，那里星鸣
  科技回答过或还没被问过），以及和它一起的 key 留在星鸣自己 key 的位置上
  （coeo91 报告：base URL 在更新或重启后回到了 127.0.0.1）。本机回环上
  的地址跟随网关；地址多于一个或 key 已经没了时，这次写入保持原样。在星鸣
  里挑选模型仍然会把 agent 指向这个星鸣，但对 Claude Code 来说是指向它的
  代理（见上）。ZCode 和 WorkBuddy 在它们自己的写入器里也保留这样的地址
  （`zcodeAddress`、`buddyWriteAt`）。
- `SyncCatalog` 会持有它构建出来的目录（`provider.Hold`），最多一秒钟，连带
  plugin 的 provider 和登录。那些 `catalog.Touched` 的丢弃在后台告知，所以
  一次 plugin 变化也会立刻丢掉被持有的东西（`plugin.OnChangeNow`，
  [`hold.go`](../../internal/provider/hold.go)）：宿主的 "auth" 事件在它对保存
  登录的回复之前就被读到，而登录返回之后再看账号就看得见它了（#2449：
  `starling plugin login` 在一次 sync 运行时找不到账号；
  `TestHeldGoesWithAPluginChange`）。
- JetBrains Air 的 `starling-opencode.json` 在 sync 和挑模型时会保留
  `provider.starling` 里用户加的字段，包括嵌套的选项（`air` 在
  [`air.go`](../../internal/agent/air.go) 里，用
  [`sync.go`](../../internal/agent/sync.go) 里的 `theirsKept`）。网关字段和
  完整的模型列表仍会被替换，所以被移除的模型会从列表里消失。
- 在每一个 OpenCode 形状的配置里（OpenCode、DevEco、Air、OpenChamber；
  [`agents.go`](../../internal/agent/agents.go) 里的 `openCodeKept`），用户放在
  星鸣某个模型上的变体在一次 sync 后依然保留，用户用星鸣的名字改写
  出来的也一样。一个完全按星鸣写法写在它名字下的变体（`openCodeVariant`）
  是星鸣的，所以模型不再有的那个会被移除（#2379，
  `TestOpenCodeCodexFastVariants`）。
- OpenCode 没有 `/fast`。一个 ChatGPT 账号自己的 `gpt-*` 模型（不是路由组，
  `catalog.Model.CodexFast`，在 `starlingModels` 里设置）会再次拿到每个推理
  等级，作为 `<level>-fast`（没有等级时就是单独的 `fast`），并像 Codex 的
  `/fast` 那样请求 `service_tier: "priority"`。这个等级被写两次：作为选项，
  OpenCode 1 的 `@ai-sdk/openai-compatible` 会原样发送；以及在 `extraBody` 里，
  OpenCode 2 会发送它并丢掉自己不认识的选项（用 OpenCode 2.0.25 对一个假服务器
  验证过）。当网关把请求重建为发给 ChatGPT 后端的 Responses 时，它会丢掉
  OpenCode 1 多余的 `extraBody` 字段（`TestOpenCodeFastVariantReachesCodex`）。
- 每一次把星鸣自己的块写进 agent 配置的 sync 和挑选，都会把它铺在现有内容
  上面，所以星鸣不写的键会保留（#2030，以及它的同类）。每个块都对照过该
  app 自己的 schema 或源码。星鸣的部分是有名字的；它不再写的键跟着它们一起
  走。
  - Qoder 的 `providers.starling` 保留 `routing`、各种超时、`type`、`authType`、
    `anthropic` 和一个非空的 `displayName`（`qoderKept`）。
  - fx 保留 `reviewer_model` 和一个值为 `omit` 的 `tool_choice_mode`。它的 `auth`
    整体是星鸣的，因为 fx 在 `type: none` 旁边拒绝 `env`（`fxKept`）。
  - Hermes 保留 `hermes model` 写在那里的 `default_model`、`key_cmd`、各种超时、
    `extra_body`，以及星鸣旁边的 name 和 headers（`hermesKept`）。这在每个
    profile 里都成立。
  - Hermes 的 `api_key` 可能是一个引用、`${VAR}` 或 `${env:VAR}`，Hermes 从配置
    旁边的 `.env`（它自己的或某个 profile 的）再从环境里展开它。`Check` 和一次
    sync 读它指名的 key（`hermesKey`）：一个"持有"的（`keyHolds`）保留下来且不算
    drift，指名一个星鸣看不到的变量的也一样（只有 shell profile 会设置它），
    那是未知。指名一个网关不接受的 key 则是 drift，一次 sync 保留那个引用，并
    把星鸣的 key 写进那个 `.env` 里对应的变量下，没有 `.env` 时就新建一个
    并设为 `0600`（`hermesKeptKey`；PaulT 报告，
    `TestHermesWithAKeyInItsEnvFile`）。一个里面含有不止那一个引用的值，会像
    以前一样被星鸣的 key 替换掉。
  - omp 的 `models.yml` 保留 `modelOverrides`、`discovery`、`disableStrictTools`、
    headers 和兼容标志（`ompKept`）。它的 `auth` 和 `apiKey` 是星鸣的，或者
    被删掉。
  - Kimi Code 的 `[providers.starling]` 就地设置星鸣的那三个键，所以
    `custom_headers`、`env` 和 `default_model` 保留（`kimiProvider`）。
  - Cline 的 `models.json` 保留 provider 的其他键，并且在每个模型上列出星鸣
    不写的东西，比如一个 VS Code 覆盖（`clineKept`）。
  - Goose 的自定义 provider 文件保留 name、`timeout_seconds`、`toolshim`、
    headers 之类（`gooseProviderKept`）。
  - ZCode 的 provider 规则保留它的 name、logo、可见性、API headers 和 key 的
    管理链接（`zcodeRules`）。
  - 星鸣写进 ZCode `providerModelRules` 的每个模型，在 provider 的格式是
    Anthropic Messages（ZCode 唯一会发送的格式）时，都打开原生联网搜索
    （`properties.supportsNativeWebSearch`，"原生联网搜索"）（`zcodeWebSearch`）。
    星鸣为哪些模型打开过它，记在 stash 的 `zcode.web_search` 下，所以用户在
    ZCode 里取消勾选的模型（ZCode 会把那个键从规则里去掉）保持关闭，规则里已有的
    值保持，手工建的规则不动，而读不出来的 stash 不会打开任何东西。换一种格式时，
    星鸣自己打开的会被去掉。断开会忘掉这份列表。
  - 那些 YAML 的走 [`sync.go`](../../internal/agent/sync.go) 里的 `yamlKept` 和
    `mergeKept`。每一个都有 `Test<Agent>KeepsTheUsersSettings` 这样的测试。
- JetBrains IDE 里的 Copilot 在 Copilot 语言服务器的 `byok.json` 里保留它的
  Bring Your Own Key provider，这份文件与 Xcode 和 Eclipse 的 Copilot 共用
  （[`copilotjetbrains.go`](../../internal/agent/copilotjetbrains.go) 里的
  `copilotJetBrains`）。星鸣在那里写自己的三个键——`starling-provider-config`、
  `starling-api-key` 和 `starling-models-config`——用服务器自己写的那种字节，而且
  只在它们不同时才写；用户的 provider 和 key 保留。用户 Manage Models 里隐藏的
  某个星鸣模型（`isRegistered: false`）在一次 sync 后依然隐藏。新建的文件放在
  0700 的目录里并设为 0600，和服务器的做法一样。
- dsh（DeepSeek Harness）为每条 llm-pi-ai 路由设置一种 API，所以星鸣按
  协议写一条路由（[`dsh.go`](../../internal/agent/dsh.go) 里的 `dshKinds`；
  01huadalang 报告）：大多数模型用 `starling`（Chat Completions），在
  Responses 上提供的模型用 `starling-responses`（`openai-responses`），在
  Messages 上提供的用 `starling-anthropic`（`anthropic-messages`，网关根路径，
  `forceAdaptiveThinking` 对应 `AdaptiveThinking`），由模型的 `APIs` 决定
  （`dshRouteOf`）。一个模型的 `APIs` 是它的 provider 最好地提供它的那种
  （`Provider.Native`，在 `starlingModels` 里，
  [`route.go`](../../internal/agent/route.go)），Pi、omp 和 droid 也同样读它。
  一个路由组在每个成员都最适合 Responses 时有 `responses`（`allResponses`，
  #1594），于是它的工具调用参数会边流边转发；其他任何组都没有，仍然留在 Chat
  上，只有 Messages 的组也一样，因为一个 Claude 成员的思考由它自己的模型决定。
  没有模型的路由会被去掉。用户为 dsh 的 Models 页面里的某条路由挑选的 api 会
  保留，他们在那里加的键也保留。启动项（`agent-default-model`）点名它模型的
  路由，Sync 会把一个位于星鸣某条路由上的启动项挪到它模型现在所在的路由
  （`dshStartOnItsRoute`），它的 effort 和标记也保留。Check 会请启动项点名的
  那条路由列出它的模型。一个在后来换了路由的模型上恢复的 dsh session，会要求
  重新挑选模型，因为 dsh 是按路由和 id 找模型的。
- dsh 把一个模型列出的 `maxTokens` 当作每次请求的输出并预留它，从
  `contextWindow` 里再留出 65536 的余量，然后才决定何时压缩
  （compaction-basic 的 `resolveCompactSpec`：窗口的 80% 和窗口 − maxTokens −
  65536 中较小的那个）。厂商的回复上限接近它的窗口时，dsh 曾在一个 500000 的
  组里于 41248 处压缩（#2823），或者在没有预算的情况下失败（524288/524288，
  LoKi 报告；131072/65536），所以星鸣把它在每条路由上以及在更老的
  `llm-deepseek` 条目里都列为限制到窗口的四分之一（`dshMaxTokens`），正如
  JetBrains 里的 VS Code 和 Copilot 所做。用户设的 Max 输出保持他们设的样子
  （`outputOr`），即使那会让 dsh 没有预算。`TestDshOutputLeavesCompactionRoom`。
- 在星鸣提供服务期间，`KeepDshWired` 每 30 秒重写一遍这些路由，因为 dsh
  会把它的 patch 列表覆盖掉。一轮什么都没写的结果会被记住（`dshRounds`），当
  它读的东西自那以后都没变时，下一轮就什么也不做。它检查 dsh 各文件的字节
  （每个 profile 的 `cordis.patch.yml`、home 层、`settings.yaml`、`.env`、
  `.credentials.yaml`、`config.yaml`）和网关地址。它还检查 `provider.StampNow`：
  没有 Changed、没有 plugin 变化、没有 `filememo.Forget`、环境相同，且 provider、
  设置和目录的文件都没变。当那个标记移动时模型会重新构建，并且至少每
  `dshModelsFor`（5 分钟）一次，以覆盖标记看不到的东西，比如另一个 app 的账号
  文件或一个按日期退役的模型。每一轮都构建整个目录和每条路由，在一个用户的
  35.7 GB 里 48 分钟分配掉了 12 GB（`TestDshRoundWithNothingChangedBuildsNothing`、
  `TestDshRoundAfterAChangeWritesAgain`）。
- DeepSeek Harness Desktop（macOS arm64 和 Windows x64；没有 Linux 构建）读的
  是与 CLI 相同的 `$DSH_HOME` 或 `~/.dsh`，作为它自己的 profile
  `profiles/desktop`（app 的 `lib/main.js` 里的 `resolveDshHome` 和
  `resolveDesktopPaths`），所以它不需要自己的接线就拿到了星鸣的路由
  （#2279，liangjinyuan-sihan）。它通过 `/Applications/DeepSeek Harness.app`、
  `%LOCALAPPDATA%\Programs\DeepSeek Harness`，或通过它当前用户安装器记录在
  `HKCU\Software\1bf39983-50d0-5fe0-9ef4-cece76f67c5e` 下的目录
  （`InstallLocation`）找到，其 Uninstall 键给出版本号
  （[`desktopapp.go`](../../internal/agent/desktopapp.go) 里的 `registered`），所以
  选了别的目录的用户也能找到它。app 装了但从未打开过时还没有 `~/.dsh`：这一行
  仍然会被列出，连接时会用 dsh 自己的模板（`dshProfileTemplate`）把带星鸣
  路由的 `profiles/desktop/cordis.patch.yml` 写进去，app 第一次启动时会保留它，
  因为它只写缺失的文件。提示说该 app 第一次打开时会在它上面启动
  （`noticeDsh5`）。在那次首次启动之前断开，会移除星鸣建的那个 profile
  （里面没有别的东西时）（`dshTidyUnopened`）；一旦 app 已经打开过它，这个 profile
  保留，只有星鸣的条目被移除（`TestDshDesktopNeverOpened`、
  `TestDshDesktopRegisteredFolder`）。
- Ante 的 `catalog.json` 以同样方式保留 `providers.starling` 里用户加的字段
  （[`ante.go`](../../internal/agent/ante.go) 里的 `ante`）；它的 `auth`、
  `http_headers`、`extra_body` 和 `stream_idle_timeout_secs`，是一个在局域网上
  共享、或从 WSL 访问的网关要发送一个 key 所需要的。Ante 按 `settings.json` 里的
  `provider` 和 `model` 启动一个 session，而不是按它的目录里存的东西（#1210），
  所以星鸣设置这一对，并把它在那里找到的放回去；用户在星鸣接线期间在
  Ante 里做的挑选归用户，保留。
- AstrBot 把它的 provider 放在它根目录下的 `data/cmd_config.json`
  （`$ASTRBOT_ROOT`，否则 `~/.astrbot`，否则桌面 app 的，否则 `~/astrbot`），
  用 UTF-8 BOM 写入，星鸣保留它（[`astrbot.go`](../../internal/agent/astrbot.go)
  里的 `astrbot`）。星鸣给 `provider_sources` 加一个 `starling` 条目，给
  `provider` 每个模型加一个 `starling/<model>` 条目，用户自己的条目逐字节保留，
  用户在星鸣条目里改的东西（超时、代理、启用）也保留，并设置默认对话模型
  （`agent_runner.config.model.provider_id`，或者在 agent runner 之前的
  `provider_settings.default_provider_id`），把它在那里找到的放回去。用户在星鸣
  科技接线期间做的挑选归用户，保留。第一次连接会把该文件复制为
  `cmd_config.json.before-starling`。AstrBot 只在启动时读该文件，它的 WebUI 保存时
  会把它整体写回去，所以 agent 的提示说在 WebUI 里保存之前要重启 AstrBot。Docker
  安装的根目录在它的容器里，星鸣到不了。
- jcode 把它的设置放在 `$JCODE_HOME` 里的 `config.toml`，否则 `~/.jcode`
  （[`jcode.go`](../../internal/agent/jcode.go) 里的 `jcode`）。星鸣按
  `jcode provider add` 写一个的方式添加一个具名 profile `[providers.starling]`：
  `type = "openai-compatible"`、网关的 `/v1`、一个点名 jcode 的内联 `api_key`
  （jcode 不发别的能区分它的东西）和 `requires_api_key = true`，因为 jcode 在没有
  key 时不会向回环 URL 发送 key。它的 `models` 列表每个模型一条，带上
  `context_window`、`input`，以及有 effort 等级的模型的 `reasoning`。用户在某个
  模型上设的 `reasoning_effort` 在该模型还有这个等级时保留。jcode 启动时用的
  模型是 `[provider] default_provider = "starling"` 和 `default_model`。原来在那里
  的东西被 stash 起来，在断开时放回去。在 jcode 位于星鸣上时挑中它自己的
  模型，会把星鸣摘掉。每个键都是单独设置的，所以用户的其他键和表保留，
  而 jcode 自己保存文件时写的 `[[providers.starling.models]]` 表会被那一份列表
  替换。没有任何要改的 sync 不会写文件。断开会整个丢掉这个 profile，包括它的
  子表。jcode 在启动时读该文件，所以提示说要重启打开的 session。
- Raycast AI 在 Mac 和 Windows 上都从用户目录下的
  `~/.config/raycast/ai/providers.yaml` 读 Custom Providers
  （[`raycast.go`](../../internal/agent/raycast.go) 里的 `raycast`；
  manual.raycast.com/ai/custom-providers）。星鸣给它的 `providers` 列表加
  一个条目，`id: starling`，在网关的 `/v1` 上，用一个 `api_keys` 别名装着一个
  点名 Raycast 的 key（它的 User-Agent 未知），并为目录里的每个模型加一个模型，
  写明 `context` 和全部五项 `abilities`，因为 Raycast 在某项没写时会假设工具
  和视觉是关的。这个列表以节点方式编辑（`EditYAMLNode`），所以用户的其他条目、
  `tools` 块（还有锚点）和注释都保留。用户加到星鸣条目或它某个模型上的键
  （`include_user_email`、一个 `description`）在一次 sync 后保留
  （`mergeYAMLKept`），而没有任何要改的 sync 不会写文件。断开会把条目拿走，如果
  它是列表里唯一一个就把列表拿走，并把只剩一个空壳的文件拿走。Raycast 会监视
  该文件，所以没有重启提示。模型是在 Raycast 里挑的，所以这个字段就是星鸣
  是否在它的选择器里（`ListsModels`）。在 Mac 上通过 `Raycast.app` 或
  `~/.config/raycast` 目录检测到。
- OpenPencil（ZSeven-W/openpencil，`openpencil-desktop` app）把它对话运行在内置
  agent 上的设置放在 Rust 的 `dirs::config_dir()` 加上 `openpencil` 得到的
  `settings.json` 里：Mac 上是 `~/Library/Application Support/openpencil`，Linux
  上是 `$XDG_CONFIG_HOME` 或 `~/.config`，Windows 上是 `%APPDATA%`
  （[`openpencil.go`](../../internal/agent/openpencil.go) 里的 `openPencil`；它
  的 `settings_io.rs` 和 `settings_payload.rs`）。星鸣给 `builtin_agents` 加
  一个条目，`id: starling`，`kind: openai-compat`，在网关的 `/v1` 上，带一个点名
  OpenPencil 的 key（它的 User-Agent 未知），`preset: custom`，以及按 OpenPencil
  写列表的方式给出的被挑模型：第一个写作 `model`，`models` 只有两个及以上才有，
  最多 64 个（v0.8.4 只读 `model`）。用户的其他条目和设置保留，用户给星鸣
  条目的 `display_name` 和 `enabled` 也保留；没有任何要改的 sync 不会写文件。
  OpenPencil 需要 `"version": 1`（没有它的文件会被当作损坏放到一边，保存也会停），
  所以星鸣在它自己新建的文件里写上它，而对版本是别的值或没有版本的文件不动。
  断开会把条目拿走，当只剩版本号和一个空列表时把文件也拿走。OpenPencil 启动时
  读一次文件，退出时把它整体写回，所以在 `openpencil-desktop` 运行期间改动会被
  拒绝（Windows 上问不到，所以提示说要先退出它）。模型是在 OpenPencil 的对话里
  挑的，所以这个字段就是星鸣是否在它的选择器里（`ListsModels`）。通过它的
  设置目录、`~/.openpencil`（它 CLI 的）、`OpenPencil.app`
  （`com.zseven-w.openpencil`）或 `%ProgramFiles%\OpenPencil` 检测到。同名的另一个
  app open-pencil/open-pencil 把它的 provider 放在 webview 的 localStorage 里，
  所以它不成行：它通过 Pi 或一个星鸣接线的 ACP agent 访问星鸣。
- DevEco Code（华为的 OpenCode 分支，#1885）按 OpenCode 的方式接线
  （[`agents.go`](../../internal/agent/agents.go) 里的 `deveco`）：星鸣的
  provider 进到 `~/.config/deveco` 下 `deveco.jsonc`、`deveco.json` 和
  `config.json` 里的第一个（DevEco 最后合并的那个，也就是生效的那个），或者在
  `$DEVECO_CONFIG_DIR` 是绝对路径时进到那里，在它们之后读。0.2.3 会把
  `deveco/0.2.3 ai-sdk/…` 发给它配置里的某个 provider，但它保留的 OpenCode 代码
  里有很大一部分仍会发送 OpenCode 自己的 User-Agent（`opencode/<version>`），
  所以它 provider 的 key 是 `starling-deveco`（`openCodeKeyAt`），不是 OpenCode 和
  MiMo Code 发的那个普通的 `starling`，而 Check 要的就是那个 key。它的 npm 包没有
  Linux 二进制，所以那里不提供安装命令（`cliSpec.only`）。
- Grok App（RongleCat/grok-app，一个通过 ACP 启动 Grok Build 的桌面前端）把它的
  provider 通道放在它数据目录下的 `agent-home/config.toml`：`$GROK_APP_HOME`，
  否则是它的 `directories` crate 放的位置，否则是那里正好存在的 `~/.grok-app`
  回退（[`grokapp.go`](../../internal/agent/grokapp.go) 里的 `grokApp`）。它的
  编辑器只列 Grok Build 的官方目录，从不列 Grok Build 那一行的
  `[model."starling/…"]` 表，所以星鸣把自己加成一个通道 `[model.starling]`，带上
  该 app 的 Providers 页面所写的那些键：网关的 `/v1`、一个点名 Grok App 的 key、
  `api_backend = "chat_completions"`、`app_provider_mode = "generic"`，以及
  `app_models`——一个 JSON 列表，列出星鸣模型及它们的 `contextWindow`、
  `supportsVision` 和 effort 等级（没有 `none`；用户在 app 里标出的默认值保留，
  否则是 medium）。被挑模型的 `context_window`（一个裸整数）和
  `supports_reasoning_effort`（一个裸 bool，Grok Build 用它把守
  `--reasoning-effort`）写在通道上，就像该 app 在编辑器里挑模型时那样放。
  `[models] default` 是路由，而星鸣不设置它：把它设成 `"starling"` 会让该 app
  丢掉自己的登录（`ensure_independent_for_custom_route`）。路由保持该 app 原来
  的样子。在 Accounts 下对星鸣按 Use 才会切换它。一个已经设过路由的星鸣
  会在下一次 sync 或连接时把旧路由放回去一次（`grokAppReleased`）；那之后的 Use
  保留。断开只丢掉通道，并且只在路由仍然是 `"starling"` 时才放回路由：连接时 stash
  的那个，否则是 `connect.was` 持有的那个。在星鸣的模型字段里挑了该 app 自己
  的某条路由，就会把它写成路由并摘掉星鸣的通道，就像在别处挑另一个 provider
  一样。一次 sync 在没有通道且路由不是 `"starling"` 时什么都不写，星鸣的 stash
  也不例外，而在路由是 `"starling"` 但通道没了时 Check 会这么说。此后在该 app 的
  编辑器里挑的模型归用户，在一次 sync 后仍然保留，而没有任何要改的 sync 不会写
  文件。共享模式把 `GROK_HOME` 保持在 `~/.grok`（Grok Build 的库那一行）。独立
  模式（Use 会打开它）读 agent-home：Library 把 `AGENTS.md`、同一个 `config.toml`
  里的 `[mcp_servers]` 和那里的 `skills/` 写出来，Grok App 也读
  `~/.agents/skills`。RTK 没有 Grok 钩子，所以 Library 把 Grok App 列出来但开关是
  关的。该 app 运行期间，提示说明在 Settings › Account 里星鸣是一个自定义
  provider，对话可以在模型选择器里挑模型，并且该 app 在 Use 之前会留在它当前的
  路由上。
- JSON/JSONC 配置编辑及其预览会拒绝格式错误的输入，而不是去编辑一个只解析了一
  部分的文档。对象编辑和 VS Code 的模型组数组编辑在解析出错时不会改动原文件；
  合法的注释和尾随逗号仍然受支持（`internal/edit`）。
- 一个以 UTF-8 字节序标记开头的配置（PowerShell、记事本或某个通道切换器保存的
  Windows `settings.json` 或 `config.toml`，#1534）会被不带这个标记地读取
  （`edit.Read`），写入时则保留它（`WriteAtomic` 里的 `keepBOM`）：用户的文件
  保留它被保存时的编码。这涵盖 `internal/edit` 写的每种格式（JSON/JSONC 及其
  数组、TOML、YAML、env）；Codex 也能透过 BOM 读 TOML 文件，所以星鸣不再拒绝
  这样的文件。以别的方式读的文件用 `edit.TrimBOM` 去掉它（Claude Code 和 CC
  Switch 的导入）。从 CRLF 文件里移除的一行会连同它的 `\r\n` 一起移除。
- 用 ChatGPT 登录的 Codex 会留在它的登录旁边（`openai_base_url`），除非用户挑
  星鸣 API。星鸣只有在 Codex app 持有那个活跃账号时才会成为它的
  provider，而它对一个被持有的账号什么都不发：`/wham/usage` 说它不被允许，
  而且它没有额度、也不在工作区超额之内（`codexHeld`，由 `provider.CodexUsedUp`
  读）。在用量看不出这件事的地方（一个实验的门禁、一份预留、保留实验下 Team
  套餐的超额），它也算被持有：动错了要花 ChatGPT 的钱，漏掉了每一轮都受影响。
  一次失败读取后保留下来的读数，在它用完的窗口重新开始后就不再说被持有了；一个
  没有窗口用完的持有会一直持续到下一次全新的读取。于是该 app 会失去它的
  ChatGPT 状态（持久 thread、远程控制），所以仅仅一个 100% 的窗口（额度还能
  过去的那种）不会移动它（`codex.out` 标记这次移动，而 Sync 会在账号有空间后
  撤销它）。
- 把一个字段从星鸣的某个模型挪到另一个模型（agent 自己的选择器）不算 drift。
  把它挪离星鸣才算。
- 一个在为空时跟随另一个字段的字段（`Follows`：Claude Code 的 tier，以及它主
  模型上的 subagent）被 `Drift` 读成在那个字段的模型上。主模型在 Claude Code
  自己的 `/model` 里被挪到星鸣曾为某个 tier 设定的那个模型上，不算 drift
  （#1050）。这个 tier 在这次移动和下一次移动中都保留那个模型作为它自己的
  （stash 里的 `tier_own`），无论在新主模型上它本会跟随什么：fable 不跟随任何
  不是 Fable 的模型，haiku 跟随 Haiku 上的 Claude 模型，被读成在跟随，并在下一次
  移动时失去它们的模型（#2257）。
- 一个被给了自己模型的 Claude Code tier 或它的 subagent，在主模型挪到那个模型
  上、再挪走时都保留它（`claudeIn` 里的 `ownKey`），无论主模型是在星鸣里挑的
  还是在 Claude Code 的 `/model` 里挑的。为它们挑的"与模型相同"会让它们重新跟随。
- 在本机的 Claude Desktop 位于星鸣上时，Claude Code 的共享设置会得到
  `CLAUDE_CODE_DISABLE_EXPLORE_INHERIT_CAP="1"`（`claudeIn` 里的 `writeExplore`，
  [`claude.go`](../../internal/agent/claude.go)）。否则 Claude Code 2.1.295 的
  Explore 继承上限会把 Desktop 的 `mythos-starling` 兼容别名换成 Opus。这禁用的是
  那个上限，不是所有 subagent 的模型选择。它在只有 Desktop 或旁边有已连接的
  Claude Code 时生效，包括模型挑选和 Sync；WSL 孪生体不会得到它。用户已有的
  值，包括空字符串，保留。stash（`claude.explore_inherit_cap`）只标记星鸣的
  值：把 Claude Code 挪离星鸣时，在 Desktop 还在的情况下它留着；把 Desktop
  挪走时，只有在用户没改过它的情况下才移除。由于这个设置文件是共享的，其他
  读到它的 Claude Code session 在 Desktop 连接期间也会看到这个标记。
- Claude Code 的 tier 和 subagent 模型（用户为自己的 endpoint 设置的，
  `claudeOwnEnv`）在星鸣接线时连同那个 endpoint 和 token 一起被 stash，
  在星鸣退出时（通过断开或通过 Claude Code 自己的模型）一起回来（#1050）。
  它们在环境里的主模型也一样（`ANTHROPIC_MODEL`，在
  `CLAUDE_CODE_USE_BEDROCK` 旁边对应一个 Bedrock 模型 id），星鸣在接线期间会
  把它摘掉，因为它比 `/model` 里的挑选优先。
- Claude Code 的 auto mode 字段（`auto_mode`，
  [`claude.go`](../../internal/agent/claude.go) 里的 `claudeAutoModeEnv`）是
  `CLAUDE_CODE_AUTO_MODE_MODEL`，也就是 Claude Code 2.1.302 用来判定每个动作的
  客户端 auto mode 分类器模型（Linka 报告）。它对服务端的 auto mode 毫无
  影响，而 Anthropic 为账号配置的分类器模型排在它前面。没有设置时，Claude Code
  挑自己的（sonnet tier 的模型，否则是 Sonnet），所以它不跟随任何字段，也不是
  `Follows`。它只在 Claude Code 路由经过星鸣时被设置，只会设成星鸣的
  某个模型，像其他模型键那样写成短横线形式，而且绝不会设成 Sonnet 5.5、Opus 5.5
  或 Opus 5（`claudeNoClassifier`）——Claude Code 2.1.303 在那里会拒绝并换成它
  自己挑的。星鸣路由 Claude Code 之前用户就有的值被 stash 起来，并随它们的
  tier 一起摘掉（`claudeOwnEnv`），在星鸣退出时回来，除非他们之后又设了
  一个；星鸣自己的值（stash 里的 `claude.auto_mode_model`，或星鸣任意
  一个模型）那时会被移除，而接线期间用户设的值不动。Claude Code 边走边读它，
  所以它是 Quiet 的，不需要重启。
- Claude Code 的 provider 开关（`agentenv.ClaudeProviders`：
  `CLAUDE_CODE_USE_BEDROCK`、`_VERTEX`、`_FOUNDRY`、`_ANTHROPIC_AWS`、
  `_ANTHROPIC_GOOGLE_CLOUD`、`_MANTLE`，Claude Code 2.1.300）把它送到那个云
  自己的 endpoint 并把 `ANTHROPIC_BASE_URL` 传过去，所以只要有一个开着，它就
  永远到不了星鸣（jw1dev 报告：Bedrock 拒绝了星鸣的模型，而星鸣什么
  都没显示）。Claude Code 把 1、true、yes 和 on 都读成开（`claudeSwitchOn`）。在
  星鸣所写的设置里开着的那个，会在每个写入接线的路径上被摘掉
  （`writeTiers` 和 `Sync` 里的 `dropClouds`），它的原始值被 stash
  （`claude.cloud.<var>`，WSL 孪生体则在该发行版的 id 下），在星鸣退出时按
  原样放回（`unroute` 里的 `putClouds`），除非用户之后又设过。那个云自己的变量
  （`AWS_*`、`ANTHROPIC_VERTEX_PROJECT_ID` 等）和关着的开关保留。在星鸣不写
  它的地方，Check 会点名那个变量、那个云和它的位置（`claudeCloud`）：托管设置
  （它们胜过用户的）；`~/.claude.json` 的 env，在用户设置之下；以及对本机的
  Claude Code 来说，登录 shell 的环境（`proc.LoginEnv`，只读，从不在星鸣里
  设置），在每个设置文件之下。项目的 `.claude/settings.json` 不读：星鸣不知道
  Claude Code 在哪些目录里运行。T3 Code 的 Check 对用户的 `~/.claude` 设置也
  说同样的事，那些设置会被盖在它的星鸣实例的环境之上，而开在这个实例上的
  开关会在它下一次写入时生效；那个 Check 首先说明该实例的 `ANTHROPIC_BASE_URL`
  是否还连得到网关（`t3EnvOf`），就像每个 agent 的 Check 一样，所以端口迁移也会
  重写它。drift 的细节通过 `tNotice` 显示在这一行的小字和它的标签上
  （`claude-cloud-drift.test.cjs`）。Claude Desktop 的 Code 标签页带着
  `CLAUDE_CODE_PROVIDER_MANAGED_BY_HOST` 运行，在它之下 Claude Code 忽略设置
  文件里的开关，所以 Desktop 两者都不需要。
- Claude Code 的登录（`login`）是星鸣的 key（`""`，默认）或 claude.ai
  （`claudeai`）。claude.ai 会把 `ANTHROPIC_AUTH_TOKEN` 写成空，Claude Code 把它
  读成没有 key：它保留自己的 claude.ai 登录并把它发给星鸣，而星鸣从不
  把它转交出去（网关里的 `claudeSignIn`）。之后每次写入都让 token 保持为空
  （`wiredKey`），而手动放在那里的网关自己的 caller key 会被当成星鸣的，
  Check 和模型挑选都一样，并按用户留下的样子保留（`keyHolds`，#2545）；网关会
  拒绝的 key 会被 Check 报告出来并被接线覆盖，星鸣自己造出来的共享 key
  （`access.Default`）也一样，在共享关闭之后——星鸣自己的 key，尽管网关还
  接受它。它只在网关接受任何 key 的地方被提供（`gwKey() == gateway.Token`：
  回环、镜像过来的 WSL 发行版），而在星鸣在网络上共享时从 NAT 下的发行版
  则不提供，那里只有共享 key 进得来。这个选择被 stash，所以它能活过断开和端口
  迁移。只有紧挨着星鸣最后写入的地址时，一个空 token 才算星鸣的
  （stash 里的 `claude.wired_at`，Rewire 会把它带到新端口，`keyOurs`）：模型回到
  Claude Code 自己的模型时，用户自己那个带空 token 的 endpoint 仍然归用户。
- Claude Code 的 auto mode 评审（`CLAUDE_CODE_AUTO_MODE_SERVER`，
  [`claude.go`](../../internal/agent/claude.go) 里的 `claudeReviewEnv`）在 Claude
  Code 路由经过星鸣、且它在那里能到达的模型没有一个是 Anthropic 评审的对象
  时写成 `"0"`（`claudeReviewed`，由 `gateway.AutoModeReviewed` 判定）：被展示
  给它的那些模型（以及 Desktop 在星鸣上时展示给它的）、一个被展示的组的
  成员、那些 provider 的 fallback，以及它被设成的那些模型，每一个都按它解析到
  的 provider（或组，或 fallback）来判定，就像网关判定它的请求那样，所以一个
  单独位于 Copilot 或 Kiro 上的 Claude 模型会写成 `"0"`，而一个星鸣不知道的
  名字按名字判定（`TestClaudeReviewedJudgesTheSetModelWhereItGoes`）。关掉时，它
  从一开始就本地分类，就像在一个不回答的网关后面本来那样，也不会出现"因为你的
  请求走 127.0.0.1:3425 所以不符合条件"的提示（1EFLOPS 报告）。任何被评审的模型
  都让它保持未设置，因为 `/model` 可以把 session 带过去；对 Claude Code 隐藏的
  模型不算。它在每次设置模型和每次 `Sync` 时被重写，所以新增一个 Claude
  provider 会把评审重新打开。星鸣的值单独存在 stash 里
  （`claude.auto_mode_server`），用户自己的值不动，并在 Claude Code 被挪离星鸣
  科技时移除（`desktopOnly`、`unroute`）。T3 Code 的星鸣实例通过自己的环境
  在网关上运行 Claude Code，它在那里按自己列出的模型带着同样的 `"0"`
  （[`t3code.go`](../../internal/agent/t3code.go) 里的 `t3InstanceJSON`）；用户
  `~/.claude` 设置里的值排在它后面。
- Claude Code 对每个它不认识的模型采用同一个窗口（`CLAUDE_CODE_MAX_CONTEXT_TOKENS`，
  [`claude.go`](../../internal/agent/claude.go) 里的 `claudeContextEnv`）。只有
  Claude Code 位于星鸣上时，它被告知自己主模型的窗口（`claudeWindow`）。在
  Claude Desktop 也位于星鸣上时，它的 Code 标签页用同一份
  `~/.claude/settings.json` 运行 Claude Code，带着 Desktop 列出的每个模型，所以
  星鸣把那些模型和 Claude Code 自己模型里最大的窗口告诉它，并用
  `modelSettings.<id>.autoCompactWindow` 把每个低于它的模型限制在自己的窗口
  （[`claude_budget.go`](../../internal/agent/claude_budget.go) 里的 `claudePlan`，
  #2025：最少的那个把一个 1M 模型限制到 131K）。一个以 `[1m]` id 传进来的模型
  （Desktop 的，在 `DesktopLongest` 下，或 Claude Code 自己的）也按它的 1M 计，
  所以当 Claude Code 不再接受这个 id 表示 1M 时（1M beta 曾经被拒一次，
  `CLAUDE_CODE_DISABLE_1M_CONTEXT`），它仍然跑在 1M 上；而如果不算，一个 1M 模型
  旁边有 272K 的 GPT 时会被告知 272K（dumplings 报告）。Claude Code 2.1.296
  按它拿到的 id（转小写且不含 `[1m]`）匹配那个键；它排在顶层 `autoCompactWindow`
  之前，后者服务于没有专属项的模型。`CLAUDE_CODE_AUTO_COMPACT_WINDOW` 会优先于
  两者，所以在这个模式下星鸣的工作窗口改放进顶层 `autoCompactWindow`，而
  只有当某个模型的窗口或它自己的阈值与没有它时会跑的值不同，才写一个条目。
  星鸣的条目和顶层值被 stash（`claude.model_budgets`、
  `claude.compact_default`），并在 Desktop 离开星鸣时移除；用户自己的条目、
  顶层值和变量保留，用户越过其模型窗口的条目（或 `auto`）会把最大的限制到那个
  窗口，而用户自己的 `CLAUDE_CODE_AUTO_COMPACT_WINDOW` 存在时，最小的窗口像
  以前一样被告知。
- Claude Code 通过 `CLAUDE_CODE_MODEL_CAPABILITIES` 得知星鸣每个模型能做什么
  （[`claude.go`](../../internal/agent/claude.go) 里的 `claudeCapsEnv`、
  `claudeCapabilities`）：对展示给它的那些模型是
  `<model>=effort[,xhigh_effort][,max_effort]`，而在 Claude Desktop 位于星鸣
  上时，也包括 Desktop 的 Code 标签页递给它的 `mythos-starling-<number>` id。
  Claude Code 认识的 Claude 模型不会得到比 Claude Code 自己给的更多（Opus 和
  Sonnet 4.6：没有 xhigh），没有已知等级的模型根本不会被点名。一个在没有 xhigh
  时就能到 max 的模型（GLM-5.3 的 low/high/max、DeepSeek V4 的 high/max）也会被
  说成有 xhigh：网关为它在 max 上跑 xhigh（`fitEffort`），而 Claude Code 2.1.299
  没有 `xhigh_effort` 就会把 Ultracode 关掉（"Ultracode isn't available on …"），
  在 Desktop 里那就是 Ultracode 等级和动态 workflow（leesunrise2364 报告；
  `TestClaudeModelCapabilities`）。用户自己的值不动。
- MiniMax Code 的模型 ref 整体只允许一个斜杠（一个 SubAgent 的 `model` 必须匹配
  `^[^/\s]+\/[^/\s]+$`，#1387），而星鸣在那里的 ref 是
  `custom_provider:starling/<model>`。所以 `custom_provider.starling` 下的模型在写入时
  用 `~` 代替它目录 id 里的每个斜杠（`bb-codex~gpt-6.1-sol`，
  `provider.FlatID`），但只在这种写法恰好对应一个被提供的 id 时
  （`provider.Flattened`）；网关会把它还原。一个更老的星鸣写过的配置会在下一次
  sync 时迁移过来：用户在某个模型上的键转到它的新键，`defaultModel`、
  `defaultLightModel` 和 `memory.captureModel`/`consolidationModel` 跟着走
  （`mcodeRefs`），而带斜杠的那个键保留但被关掉（`enabled: false`），这样在它上面
  开始的 session 仍能恢复。星鸣绝不改用户自己的 `agents/<name>/agent.md`。
  没有发现别的 agent 有这个限制（[`minimax.go`](../../internal/agent/minimax.go)）。
- Claude Code 的 `/model` 列表（它 settings.json 里的 `modelPicker`）是星鸣的，
  只要它在 stash 里有标记（`claude.model_picker`，WSL 发行版则是
  `claude@wsl:<distro>.model_picker`），或者每一行的描述都以 "via starling" 结尾
  ——星鸣写的每一行都是这样（[`claude.go`](../../internal/agent/claude.go) 里的
  `starlingsPicker`）。星鸣在每次设置模型和每次 `Sync` 时把自己的列表保持最新，
  并在 Claude Code 被挪离星鸣时把它移除；用户自己写的列表不动。这个指纹
  也覆盖那些标记被更老的星鸣的 stash 弄丢了的列表，否则它们会一直缺少此后
  新增的每个模型（#2046）。
- Claude Code 2.1.296 把版本号里带点的 Claude 模型（`relay/claude-opus-4.6-us`）
  读成 Claude Opus 4：一条退役警告、预算思考、32000 输出 token、没有 effort。
  用短横线（`relay/claude-opus-4-6-us`）它就认识这个模型（misteer_ti 报告）。
  所以星鸣写进它设置的每个模型（`model`、各个 tier 的 env、
  `ANTHROPIC_SMALL_FAST_MODEL`、`CLAUDE_CODE_SUBAGENT_MODEL`、
  `CLAUDE_CODE_AUTO_MODE_MODEL`、`modelPicker`）、T3 Code 的 Claude slug 和 Claude
  Desktop 的 Code 标签页 id（`claudeLooking`）都把带点的 Claude 版本号写成短横线
  形式（`provider.ClaudeSpelled`），但只在这种写法恰好对应一个被提供的 id 时。
  一个带推理等级列给 Desktop 的 Claude 模型给出的形式是
  `starling-<n>.anthropic.<model>`，它的带点版本号也写成短横线（`claudeModel`：
  Copilot 的 `claude-opus-5.5` 是 `…anthropic.claude-opus-5-5`，Claude Code 认识它；
  它以前是 `mythos-starling-<n>`，一个未知模型，而 Haiku 4.5 得到了 Copilot 拒绝的
  一个 effort，#1819）；对它的请求由 `<n>` 单独提供（`aliased`），所以这种写法
  不必唯一。Mythos 在这里也算 Claude：Claude Code 用子串匹配 `claude-mythos-5-1`，
  所以 `claude-mythos-5.1` 会被读成 Mythos 5。对那些键的每次读取都会把它取回
  （`provider.ClaudeRead`：`claudeIn` 的 `env`、`claudeMain`、`claudeStandInAt`），
  所以星鸣的字段、应用状态和 Drift 都保持那个带点的 id。用户自己的值被
  stash 起来并按原样还给用户。更老的星鸣写成带点形式的值会在下一次 Sync 时
  被重写。别的厂商的 id 永远不会被改写（[`claudedash.go`](../../internal/provider/claudedash.go)）。
- Codex 的 "subagents" 字段是 `[agents] default_subagent_model`：只有当一个
  subagent 的 lead 在 `spawn_agent` 里没有点名模型时，subagent 才从这个模型启动
  （codex-rs `core/src/agent/child_config.rs`）。没有设置时，它不是"与模型相同"：
  lead 仍可能点名另一个，所以 GUI 把它叫作 lead 的挑选，否则就是模型
  （app.js 里的 `leadPicks`），它的标题指向 "subagent model" 字段。
- Codex 的 "subagent model" 字段不往 Codex 的配置里写任何东西：它的 lead 在
  `spawn_agent` 里点名每个 subagent 的模型，而 `[agents]` 无法推翻它。它是星鸣
  科技的设置（`settings.CodexSubagentModel`，
  [`codex_subagent.go`](../../internal/provider/codex_subagent.go)），网关会把
  subagent 的请求挪到它上面（见 [gateway routing](gateway-routing.md)）。它的选项
  是 Codex 自己列表里显示的每个模型和路由组（`CatalogFor("codex")`，所以从它们
  中移除或没被挑中的模型在这里也被排除），ChatGPT 账号排在前面（#1633：一个位于
  chat provider 上的 lead 也希望它的 subagent 在那里）。一个由 ChatGPT 支撑的 lead
  封存的 task 只能在 ChatGPT 账号上打开，所以当所设的那个模型属于另一个 provider
  时，网关会让这样的 subagent 留在它自己的模型上。它的 setter 接受任何
  `provider/model` 或 `group/<id>`，也接受星鸣现在不提供的，所以放到另一台
  电脑上的 profile 或备份仍然生效（`ApplyFields` 在第一个失败的 Set 处停下），并且
  拒绝 Codex 自己的裸名字。
- Claude Desktop 的 "subagent model" 字段（`subagent_model`，
  [`claudedesktop.go`](../../internal/agent/claudedesktop.go) 里的
  `desktopSubagentField`）不往 Desktop 的配置里写任何东西：Desktop 用
  `CLAUDE_CODE_PROVIDER_MANAGED_BY_HOST=1` 运行它 Code 标签页的 Claude Code，
  在它之下 Claude Code 会忽略设置文件 `env` 里的 `CLAUDE_CODE_SUBAGENT_MODEL`、
  `CLAUDE_CODE_SUBAGENT_MODEL_FORCE` 和各个 tier 变量（"Ignoring … this session's
  provider routing is managed by the host"，Claude Code 2.1.295），而 Desktop 为
  Code 标签页构建的环境里没有它的键（Desktop 2.31226）。它是星鸣的设置
  （`settings.DesktopSubagentModel`，
  [`desktop_subagent.go`](../../internal/provider/desktop_subagent.go)），网关会
  把每个 subagent 的一轮挪到它上面（见 [gateway routing](gateway-routing.md)）。
  它被标成和 Codex 的一样，所以 Agents 页面给它同样的方块。它的选项是 Desktop
  列表显示的星鸣模型和组（`viaStarling("claude-desktop")`），在 Desktop 位于
  星鸣上时如此，和 tier 一样。被挑中的 Desktop 别名按它的目录 id 保留
  （`DesktopCatalogID`）。和 Codex 的一样，它的 setter 接受任何 `provider/model`
  或 `group/<id>`，也接受星鸣现在不提供的，所以 profile 或备份仍然生效，并且
  拒绝裸名字。它是 `Live` 的：网关在每次请求时读它。
- Claude Desktop 的 "subagent effort"（`desktopSubagentEffortField`，
  `settings.DesktopSubagentEffort`）同样是 Quiet 且 Live 的。默认值保留请求所要的
  effort，包括关闭的思考；它绝不会另选一个默认值。被挑的等级与 subagent 模型
  无关：没有挑模型、挑了同一个模型、挑了一个不可用的模型，都仍然采用那个
  effort。接线期间，该字段提供所选模型的等级，或在它的模型是继承来的或不可用时
  提供标准等级；被标成无推理的模型不提供任何等级。存下来的挑选能活过模型变更。
  一个断开状态下的偏好可以在不连接 Desktop、也不显示重启建议的情况下保存或清除，
  这样 profile 恢复和断开就能恢复连接前保存的值；重新连接会回到最后一次生效的
  挑选。`subEffortTitle` 给 Desktop 一个和 Claude Code 相同的默认说明；Codex 的
  默认含义不变。网关把每次尝试适配到它实际的模型上，并保留固定的组成员 effort
  （见 [gateway routing](gateway-routing.md)）。
- Claude Desktop 的 3p 模式跑在它自己的 userData 里（`Claude-3p`），那里的
  `config.json`（Desktop 自己的设置，不是 `claude_desktop_config.json`）开头没有
  `locale`，所以位于星鸣上的 Desktop 用的是系统语言，而不是用 Anthropic 登录
  时挑的语言（memory 和 Osal 报告）。把 Desktop 切到星鸣时，会从登录
  用户的 userData 的 `config.json`（`Claude`，Windows 上是 `%APPDATA%\Claude`，
  取第一套而不是其他）里把那一个键 `locale` 复制到 `Claude-3p/config.json`，
  当那个文件缺失或是一个不含它的 JSON 对象时（[`claudedesktop.go`](../../internal/agent/claudedesktop.go)
  里的 `desktopCarryLocale`）；只有 Desktop 为一个目录项取的名字会被带过去。
  `Claude-3p` 里已有的、Desktop 自己的或手动设置的 `locale` 绝不会被覆盖写，
  而那个文件的其他键也永远不会被写。关闭时，只有当该文件是星鸣建的、且里面
  除了带过来的语言什么都没有时才移除它（`TestClaudeDesktopCarriesTheLanguage`、
  `TestClaudeDesktopKeepsThe3pLanguage`、`TestClaudeDesktopUserEditsThenReapply`）。
  Desktop 只在它的构建打包了那个目录项时才显示该语言（2.7032 有 11 个，没有中文）。
- Claude Desktop 按它自己的目录（"Opus 5.5"）给带推理等级的 Claude 模型命名，
  而不是按星鸣列出的 `display_name`，因为星鸣把它列成
  `starling-<number>.anthropic.<model>` 以保住它的 effort 选择器
  （[`desktop.go`](../../internal/gateway/desktop.go) 里的 `claudeLooking`）。于是
  同一个模型来自两个 provider 的行，在 Desktop 菜单里只差那个编号，所以 provider
  编辑器把它显示在模型的小标签上（[`providers.go`](../../internal/gui/providers.go)
  里的 `gateway.DesktopNumber`、`desktopNumbers`，#2578）。这个小标签只在 Claude
  Desktop 已安装且展示了该模型时才显示它。Desktop 展示的每个其他模型在小标签上
  都没有编号：Desktop 按 `display_name` 命名它，而那个名字已经说明了 provider，
  或者按它自己的 id 命名。
- Codex 的 `spawn_agent` 工具只把它目录里的前五个模型告诉 lead（codex-rs
  `MAX_SPAWN_AGENT_MODEL_OVERRIDES` in `child_config.rs`，
  `spawn_agent_models_description`），Codex 自己的排在星鸣之前，尽管它按名字
  接受任何目录模型（`find_spawn_agent_model_name`）。所以一个 lead 可能把它的
  subagent 放在星鸣的模型之外（Gearclickist 报告）。网关把交给那个工具的
  目录里其余的模型补进它的描述，放在 Codex 自己列表之后，最多 200 个
  （[`codex_spawn_models.go`](../../internal/gateway/codex_spawn_models.go) 里的
  `spawnAgentModels`）。一个能搜索工具的模型（`supports_search_tool`）不会在请求的
  tools 里收到它：Codex 把它的多 agent V1 工具放在 `tool_search` 后面（codex-rs
  `core/src/tools/spec_plan.rs`），而 lead 得到的是 Codex 放进对话并在每一轮重发的
  `tool_search_output` 里的 `spawn_agent`，所以那一份也被加宽了。这段文本会随
  prompt 一起送到上游，对 ChatGPT lead 连 OpenAI 也会送到，所以它的头部
  （`spawnMore`）保持中立：既不提星鸣，也不提 Codex 的上限。在 `CodexPath`
  处它是 Codex 自己列出的模型减去隐藏的那些，然后是 `CodexListed`，除非 Codex 被
  限制为自己。在 `/v1` 处它是该 key 的额度下的 `CodexCatalog`。一个由 ChatGPT
  后端以多 agent V2 回答的 lead 会为 ChatGPT 账号封存它的 subagent 的 task
  （#619），所以它只得到 Codex 自己的模型。Codex 放进上下文的模型目录
  （`model_catalog_in_context`，开发中，默认关闭）同样上限五个，也不加宽。
- Snow CLI 一次只跑一个 profile（`~/.snow` 或 `$SNOW_CONFIG_DIR` 下的
  `profiles/<name>.json`，名字记在 `active-profile.json` 里），并在每次启动时把
  当前那个复制成 `config.json`，`snow --ask` 也一样
  （[`snow.go`](../../internal/agent/snow.go)）。星鸣写一个 `starling` profile、
  把它设为当前那个，并自己把它复制成 `config.json`，这样已经在跑的 Snow 和 Snow
  App 的 "Sync Snow CLI API config" 看到的是同一个。它由用户当时所在的 profile
  生成，保留他们的设置，去掉它的 OAuth 登录。那个 profile 的一个视觉模型只有在
  带着自己的 endpoint 和 key 时才一起带上（`snowOwnVision`），因为用 profile 上的
  那个现在会走网关。`maxContextTokens`、`maxTokens` 和 `supportsVision` 跟着目录里
  的模型走（`snowSynced`；`supportsVision` 像给每个 agent 一样给出，在星鸣的
  Vision 描述图片期间为真）；星鸣最后写入的这些值被 stash，而用户此后在 Snow
  自己的设置里改过的那个，在后续的挑选和 sync 中保留。小模型字段是星鸣
  profile 的 `basicModel`，Snow 用于摘要、压缩和文件搜索的轻量模型：跟随主模型时
  为空，一个经由星鸣的自有模型在换主模型和 sync 之后仍然保留；在用户自己的
  profile 上它什么都不提供并被隐藏。思考字段是星鸣 profile 的 `chatThinking`
  （`{enabled: true, reasoning_effort}`），是 Snow 在对话协议上唯一读的思考设置；
  为空表示删除它，即 Snow 的默认，也就是发送关闭的思考。用户从自己的 profile
  带过来的 chatThinking 会被展示，不会被覆盖写。他们当时所在的 profile（或者没有
  `active-profile.json` 的情况）以及他们自己的一个 `starling.json` 被 stash 起来，
  并在断开时放回。profile 出现之前的 Snow（只有 `config.json`）会得到 Snow 自己
  会为它生成的 `default.json`，这样它下次启动就不会拿星鸣的去替代用户的。挑中
  用户另一个 profile 的模型会把 Snow 切到那个 profile。在星鸣的 profile 上，
  一个既不是经由星鸣的模型、也不是用户某个 profile 的模型的值会被拒绝并且
  什么也不改（#1457：`starling snow --help` 过去会把星鸣摘掉，并把 `--help`
  写成它退回的那个 profile 的模型）。出于同样的理由，Gemini CLI 的 auth 字段会在
  取消路由之前拒绝一个它不认识的 auth。在 CLI 上，任何命令或 agent 之后的 `-h`/
  `--help` 只打印用法，在 `run` 读取或移动任何东西之前，而且给 `starling <agent> …`
  的任何其他 flag 都会被拒绝（`help_cli.go`；`TestHelpOnEveryCommandDoesNothing`
  遍历 `run` 的 switch 和每个 agent）。
- omp 的 `config.yml` 里的 `enabledModels` 是它把默认角色绑定的列表：不在其中的
  模型不会被使用，无论 `modelRoles.default` 怎么说（#1769）。所以为默认角色挑的
  模型会被加进那个存在的、且还没有按名字或按 omp 的某个 glob（`*` 不跨 `/`，如同
  `Bun.Glob`）收进它的列表；一旦 `config.yml` 里没有任何东西在星鸣上，它那些
  普通的 `starling/…` 条目就会被摘掉（空掉的列表被删除，绝不留下 `[]`）；一个被
  改名的 provider 的条目会被改名一次，目录限定的条目不动
  （[`omp_scope.go`](../../internal/agent/omp_scope.go) 里的 `ompScope*`）。没有
  列表就没有变化：omp 于是使用每个模型。一个把默认角色的模型排除在外的列表就是
  drift，Reapply 会修好它。
- Pi 的 `settings.json`（以及 OmO 和 Prime Agent 的，`piLike`）里的
  `enabledModels` 是新 session 开始时用的 Ctrl+P 列表。一个被挑了但列表没收进去
  的模型会被加上（[`pi.go`](../../internal/agent/pi.go) 里的 `piScopeWith`），而
  一个被改名的 provider 的 `starling/<old>/` 条目会跟着改名，glob 或思考等级也包括
  在内，刚被这次挑选加进去的那份副本被丢掉（[`rename.go`](../../internal/agent/rename.go)
  里的 `renameRefList`，#1919，`TestPiEnabledModelsFollowRename`）。用户的其他条目
  保留；一个没有提到该 provider 任何模型的列表不会被重写。
- 在目录同步之后，Pi 的（以及它各分支的）、StepCode 的或 omp 的 `enabledModels`
  里，一个直接点名星鸣某个模型（不是 glob；思考等级会被读掉）而该模型星鸣
  科技已不再提供给该 agent 的条目会被摘掉：Pi 过去在每次启动时为每一条警告
  "No models match pattern"（#2055）。它可能是用户自己的，所以它被记在
  `stash.json`（`<agent>.scope.gone`）里，并在该模型重新被提供时回到列表末尾；
  一次改名也会移动这些被记下的条目，而断开会忘掉它们。glob、其他 provider 的条目
  和 omp 的目录限定条目保留。一个被留空的列表会被删除，并在它的条目回来时带着
  这期间挑的模型重新建好；用户自己删掉的列表不会。当 agent 文件里星鸣的模型
  列表读不出来时，什么都不会被摘掉（[`scope_gone.go`](../../internal/agent/scope_gone.go)，
  `TestPiEnabledModelsLoseDeselected`）。对 StepCode 来说，接线状态里最后写入的
  范围会跟着走，所以断开时仍会把用户自己的列表放回去。
- 一个 agent 写成本是它自己 provider 的模型的值，绝不会被当成星鸣的一个，
  即使星鸣有同名的 provider（`Spelled`，#835）。
- 反过来，一个点名星鸣某个模型而该 agent 并没有被展示它（因为
  `starling visible` 把它藏起来了，或某个模型没为它勾选）、或者点名星鸣某个
  provider 的一个不在它列表上的模型（一个远端星鸣的组在列表取回之后加进去
  的），或者点名一个星鸣没有对应组的 `group/…`，都会被 `Spell` 拒绝
  （`notOffered`），什么都不会变。它过去会被写成 agent 自己的：Pi 把
  `remote/group/smart` 读成它自己的 provider `remote` 和模型 `group/smart`，
  于是被从星鸣里摘了出去（#2376）。一个被展示给该 agent、但某个字段提供的
  比它少的模型（Claude Desktop 的 tier）仍然会被采用，一个其 provider 部分是该
  agent 自己的某个模型所走的值也一样（Pi 里的 `openai/…`）
  （[`agent.go`](../../internal/agent/agent.go)，
  `TestSpellRefusesStarlingModelNotOffered`）。
- 对 `unwired` 执行 `Reapply` 会重新设置每个位于星鸣某个模型上的字段，以及
  每个被设成星鸣本身的字段（`starlingID`，`Wired` 也按它判断）：Claude Desktop
  的 drift 会显示在某个 tier 上，而它的 `provider` 字段负责把网关写回去
  （dumplings 报告）。Claude Desktop 的 `Check` 只判断它上次运行所在的那套
  目录（`desktopRead`，由 `desktopdir.LastRun` 给出），一次都没运行时判断所有套：
  星鸣会写 MSIX 安装的两套，而 2.31226 读的是真正的
  `%LOCALAPPDATA%\Claude-3p`，所以包里那份被清空并不表示 Desktop 离开了星鸣。
- `unreachable` 从不设置任何东西：配置是对的，而页面上的小标签只说明那里必须
  有什么在监听。例外是一个 WSL 发行版以前到达 Windows 的地址（`distro.Was`，
  保存在 `wsl.json` 里），而它现在到达 Windows 用的是另一个地址：小标签会给出
  前一个，而 `Reapply` 会把配置挪到那里（`Agent.move`，Codex 的
  `codexMoveGateway`）。没有任何发行版被看到过的地址是用户自己的，永远不会被移动。
  未知答案（还没探测过）和网关在回环上也停掉的情况都不是 drift。测试在
  `TestMain` 里给 `reachProbe` 打桩。
- `proxied`（Muma102kMichael 报告：星鸣接管 Codex 之后每个项目、每个新对话
  都是 502）指的是本机的 Codex 把它的请求发给网关时经过了一个不转发它们的代理，
  而网关在这里是有回答的（[`codex_proxy.go`](../../internal/agent/codex_proxy.go)）。
  Codex 的 reqwest 取 `HTTP_PROXY`，否则取系统对 `http://` 的代理（只有 macOS 的
  HTTP 条目，绝不是它的 HTTPS 或 SOCKS 那个；Windows 的 `ProxyServer` 只有在它
  对所有协议都是同一个地址时），否则取 `ALL_PROXY`，并且只让 `NO_PROXY` 里一个
  精确的 IP 或一段 CIDR 把 127.0.0.1 排除在外（在 Windows 上，没有 `NO_PROXY` 时
  用去掉了 `*.` 的旁路列表）：`127.*` 和 `localhost` 不行，macOS 的例外也不被
  读取。带上 `[features] respect_system_proxy` 时它只问系统，自行匹配它的列表，
  通配符和 `<local>` 也包括在内，并在那里没有设代理时直连；只有一个 macOS SOCKS
  条目时它用不了，于是改取 `HTTP_PROXY`/`ALL_PROXY`（codex-rs `outbound_proxy`；
  用 codex-cli 0.162.1 试过）。它的 reqwest 没有 SOCKS，所以一个 `socks5://` 代理
  会被当成 HTTP 代理来说话，探测也一样。环境是 Codex 的：登录 shell 的 `*_PROXY`
  （`proc.LoginEnv`，终端里的 Codex 和 ChatGPT app 都有；只有导出的那些，因为探测
  是向一个子 `/bin/sh` 问它们的，所以 fish 里没导出的通用 `http_proxy` 不算，
  #2627）优先于星鸣自己的，而从 Dock 启动的星鸣没有那些（liu_dyj 报告：
  profile 里的 `NO_PROXY=127.0.0.1`，或者一个没有 HTTP 条目的 macOS 代理，会在
  Codex 直连时发出警告）。`codexProxyFor` 用同样的方式算出那个代理，而穿过它的
  `GET /` 必须以星鸣的样子返回（`proxyProbe`），像 `reachable` 那样被尝试。
  它的小标签像 `unreachable` 的那样说明怎么修，点名那个代理以及它是从哪个变量读
  来的，`http_proxy` 或 `HTTP_PROXY` 照原样（`Addr`、`Detail`），通过
  `netproxy.Redact` 处理，不带它的用户名和密码（#1886），而探测是整个穿过去的，
  那个对话框里有从星鸣断开。星鸣不请自写就写入的账号故障转移会等那次
  探测（`codexProxyOK`），在代理不转发 Codex 时不会被写入，已经写入的会被移除。
  测试在 `TestMain` 里给环境、系统代理和 `proxyProbe` 打桩。
- 列举从不会启动一个停着的 WSL 发行版。一个停着的发行版的 agent 显示上次在那里
  看到的东西（`asleep`，`wsl.json`）；设置它的一个字段会写入它的文件，这会启动
  那个发行版。
- Codex 的登录（`login`）是 ChatGPT（`""`，默认）、Always ChatGPT（`chatgpt`）或
  星鸣 API（`api`）。用 ChatGPT 时，星鸣的模型会加入 Codex 自己的模型，
  就在登录旁边（`openai_base_url`）；而在 Codex app 持有 Codex 登录的那个 ChatGPT
  账号时（`codexUsedUp`，`provider.CodexUsedUp`：OpenAI 不再允许该账号且它没有
  剩余额度，仍在超额之内的工作区除外），星鸣会取代成为 Codex 的 provider
  （`codex.out`）：Codex app 在它持有的账号上什么都不发，包括一个星鸣模型的
  轮次（#540），而作为它自己的 provider，星鸣已经越过了那条线。一个 100%
  但还有额度的窗口不算，Codex 保持登录。Sync 会在账号有空间后把它放回登录旁边。
  Always ChatGPT 永不这样：Codex 留在它的 ChatGPT 状态里，Codex app 可能一直不
  发东西直到账号有空间，而 Codex CLI 的轮次仍通过 base URL 到达星鸣，由它的
  路由和其他账号来服务它们。用星鸣 API 时，Codex 的 `getAuthStatus` 不返回
  token（它的 provider 没有 `requires_openai_auth`），于是它的 Chrome 控制会拒绝；
  这个选项的说明说请选 ChatGPT 以保留它（#1824）。在星鸣是 Codex 的 provider
  时（星鸣 API，或 `codex.out`），`Joined` 报告 false，尽管 `codex.joined` 标记
  还在，为 Sync 保留回头的路：Agents 那一行说 Codex 自己的模型会留在 `/model` 里
  星鸣模型旁边，而 ChatGPT 状态会阻止 Codex app 发送东西，只有 app 重启后
  才离开那个状态（#2267，`TestCodexProviderIsNotJoined`）。一个从另一台机器到达
  网关的 Codex（NAT 下的 WSL 发行版，或 `STARLING_ADDR` 固定在回环之外，
  `onAnotherMachine(at.gw())`）被接成没有 ChatGPT 登录的那种（那里 `chatGPT()`
  是 false）：绝不 joined 也不在登录旁边，因为在登录旁边它会把 ChatGPT token 发到
  `/backend-api/codex`，而从那里网关只接受一个具名 key，于是每个请求都是 401
  （#2145，`TestCodexBeyondLoopbackNotJoined`）。星鸣是它的 provider，它发送
  `gwKey()`，也就是局域网共享 key；挑中它自己某个模型的请求会直接按它的登录走
  到 OpenAI。
- 用 ChatGPT 登录但没连接的 Codex，只要在星鸣里启用了更多它的 ChatGPT 账号，
  仍然会拿到 `openai_base_url`（`failover`，stash `codex.failover`；以 CC Switch
  镜像为 provider 时则是镜像的 `base_url` 和 `codex.mirror_failover`），这样它自己
  账号没有额度的轮次就继续走到下一个——但对那种从另一台机器到达网关的 Codex 绝不
  这样，在那条路上它只接受一个具名 key，故障转移过去的轮次会回来一个 401
  （#2145，`TestCodexFailoverBeyondLoopbackNotWritten`）；Sync 会把它之前写过的移除。
  那个 Codex 处于 `FailingOver`：不是 `Wired`，但它的请求经过星鸣。Agents
  那一行会说明这一点以及怎么关掉它（CLI 和 TUI 里的 `FailoverSaid`、
  `/api/state` 里的 `failover`、GUI 里通往 Providers › Codex 的一个按钮），网关只
  从 `/backend-api/codex/models` 把它自己的模型给它（`CodexOwnOnly`，由 main 设置
  在 `gateway.CodexOwnOnly` 上；它的 ETag 带着 `provider.CodexOwnListTag`，而 Sync
  会让一个没有它的 `models_cache.json` 变旧）。一个已连接的 WSL Codex 共用网关
  这条路，所以只要有一个连着，每个 Codex 都拿到星鸣的模型（#1385）。用户
  手写的 base URL（#856）不是故障转移的，保持一个普通的未连接行。用户断开过的
  Codex 在他们再次连接之前得不到故障转移。提供网关的那个 app 退出时会把故障转移
  移除（`OnShutdown` 里的 `EndFailover`、`EndCodexFailover`），所以一个被留在自己
  那边的 Codex 会直接去 OpenAI，而不是去重试一个已经没了的网关（Leon920717 报告）；
  下次启动时的 Sync 会再写一次。Codex 的 app-server 在启动时读到了旧的 base URL
  并保留它，而 `KeepCodexDaemonCurrent` 以星鸣结束，所以在没有 codex session
  用着它时 `EndCodexFailover` 会在那里重启它（`codexDaemonCheck`）；有 session 用着
  时，它会留在那个已经没了的网关上，直到用户重启它（thisisvainxx 报告："重启动了好
  几次"）。一个已连接的 Codex 在退出时保留星鸣：它的模型是星鸣的某个模型，
  而把它自己的配置放回去是断开做的事。一个不以退出方式结束的星鸣（被杀、崩溃、
  `starling serve`）会把故障转移留在原处。
- 一个 Codex thread 保留它启动时所在的 provider。当一个星鸣模型在上面时，CC
  Switch 的 provider 表也经过星鸣（`takeTables`），并在 Codex 离开星鸣时
  拿回它们自己的 base URL（`giveTables`，stash `codex.tables`）。CC Switch 的
  OpenAI 镜像（另一个 id 的表，`name = "OpenAI"`，`requires_openai_auth`，没有
  `base_url`）不是被 `openai_base_url` 到达的，后者只把内置 provider 指向星鸣
  科技：在它上面开始又被重新打开的 thread 会直接去 ChatGPT，而 ChatGPT 拒绝了里面
  挑的星鸣模型，星鸣从没看到这个请求（#1796）。当 Codex 通过 base URL
  走向星鸣时，`takeMirrors` 给每个这样的镜像把星鸣的 ChatGPT 后端作为
  它的 `base_url`（stash `codex.mirrors`），而 `giveMirrors` 又把它摘掉；一个被
  某个 profile 点名的镜像，或者一个 `base_url` 之后被改过的镜像，保持原样。把
  这些 thread 挪到另一个 provider 上是 Sessions 页面的 Move（#887）。
- 一个 agent 的配置按已安装的那个版本能读的形式写，取自那个版本声明的东西，而不
  是最新的。VSCodium 没有自己的 Copilot Chat，而 Marketplace 上最后一个
  （0.48.1）没有 Custom Endpoint provider，只有 OpenAI Compatible（`customoai`，
  #1014）。星鸣从 `~/.vscode-oss/extensions` 下安装的 `github.copilot-chat`
  包的 package.json 里读 provider 列表（`vscodeChatOf`）。当声明了
  `customendpoint` 时它为自己的组写 `customendpoint`，只有 `customoai` 时就为
  `customoai` 写，并把每个模型的窗口写成 `maxInputTokens` + `maxOutputTokens`，
  把 token 写成 `x-api-key`（那个版本会丢掉 `requestHeaders` 里的 Authorization）。
  扩展缺失或读不出来时保留 `customendpoint`。Sync 会重写为另一个 provider 写的
  组，Check 也会报告它，正如它对一个既没声明 Copilot Chat 也没声明 provider 的
  情况的处理。那个 Copilot Chat 只有在用个人 Copilot 套餐登录 GitHub 时才注册它的
  provider，Notice 会说明这一点。
- 调优过的设置只在用户点击时写入，绝不由 Sync 或 Connect 写入。那里的原值会被
  stash 一次（`stash.json` 里的 `tune.<agent>.<knob>[.<model>]`），这样 Undo 能把
  它放回去，或者把这个键拿掉（以及留空的 `modelSettings` 条目）。用户在点击之后
  改过的值归用户：这时 Undo 只是忘掉这份 stash。一个排在设置之前的变量
  （`CLAUDE_CODE_AUTO_COMPACT_WINDOW`、`FORCE_PROMPT_CACHING_5M`、
  `CLAUDE_CODE_PROMPT_CACHE_TTL`、`CLAUDE_CODE_SUBAGENT_PROMPT_CACHE_TTL`，在
  settings.json 的 `env` 里或星鸣自己的 env 里）会被显示为 `locked`，Apply 会
  拒绝，因为那样设置根本不会改变什么。Claude Code 在所设窗口和模型自己的窗口
  中较小的那个上压缩（它的 `/autocompact` 写着 "capped to 200k by model"），所以
  建议从那个开始，而一个不是 Claude 的模型在这里没有窗口（由星鸣自己的压缩
  接线为它服务）。Codex 的窗口是星鸣的目录为它跑过的每个模型列出的那个
  （标注的，或厂商的），而当 config.toml 里没有时，它现在按星鸣告诉 Codex 的
  那个跑（`provider.CompactFor`，否则是 agent 打开那一行里的 Long conversations，
  由 `codexcat` 写入）；星鸣某个窗口未知的模型，除了 1M（或用户自己更大的
  值）之外不限制任何东西，绝不会按见过的最大 prompt 来限制（zhemu 报告：一个
  1M 模型在 750K 时被提供了最多到 410K），而只有 Codex 自己的模型（在目录之外）
  才保留 272K 或见过的最大 prompt（`codexWindow`，`TestTuneWindowsAreTheUsers`）。
  建议只在重放能省下 3% 或更多时才改变一个值，而在最便宜的 1% 之内的窗口里取
  最大的那个；有压缩计时的情况下，它绝不建议一个比现在压缩得更频繁的窗口，也就是
  绝不让等待变长：在那些压缩不更频繁的窗口里取最便宜的，并在那些与它相差 3%
  以内的（`sameCost`，与"完全改变"同一个门槛）里取压缩得最少的那个，两个压缩
  一样频繁的取较大的。token 和秒数没有文件所测量的共同单位，所以建议不会按猜
  出来的比率把一个换另一个；一个更便宜但等待更长的窗口会在这一行和选择器里带上
  它的等待，由用户来选。每次压缩的等待按每个窗口都一样来取：文件只在你用的那个
  窗口附近给压缩计时，太窄了，容不下它如何随 prompt 增长。token 按计费来权衡：
  对 Claude 是缓存读 0.1、5 分钟写 1.25、小时写 2、输出 5；其他是读 0.1、写 1、
  输出 8。
- 一个 plugin 的 agent 就地写每个键（JSON、YAML 或 env），用户的注释和其他键保留。
  `connect` 返回的一个对象或数组被当作它的 JSON 文本（`agentplug.keys`），所以它
  按 plugin 给的顺序写入，而不是像 Go map 那样排序：Kilo 按键顺序列出它的 effort
  变体（#1890，`TestPluggedKeepsPluginOrder`）。每个键在第一次写入前的内容被 stash
  在 `plugin:<id>:<path>` 下，带着类型，而它自己的模型或断开会放回每个键并拿走
  那些原本不存在的；文件绝不会被整体写入。一个抛异常或接管的 `connect` 什么都不
  写，并说明原因。它的网关 key 是 `starling-<id>`（`agentKeyAt`），用量按那个 key 或
  按它的 `ua` 词来称呼它的请求（`usage.PluginAgents`）。它的 `agent.library`
  （说明、`mcp` 连同 `mcpFormat`、skills）也给了它 Library 的那三样，由
  `internal/library` 写，和任何 agent 的一样（[library](library.md)）；一个在
  `agentplug.MCPFormats` 之外的 `mcpFormat`，或者一个没有 `mcpFormat` 的 `mcp`，
  会让这个 plugin 无法加载。
- StepCode 的配置根解析为一个绝对 `STEP_CODING_AGENT_DIR` 的父目录（agent 资源
  目录与配置根不同；相对值被忽略），或者在原生环境下 home 下的
  `STEPCODE_CONFIG_DIR`，默认是 `~/.stepcode`
  （[`stepcode.go`](../../internal/agent/stepcode.go)）。在 StepCode 里项目配置覆盖
  全局配置，但星鸣只管理全局配置（`config.toml` 和 `models.json`）。检测
  严格只用目录，通过 `agentDir`，以避免和 `PATH` 上 Smallstep 的裸 `step` 命令
  冲突。接线做一次原子的双文件编辑（`models.json` 的 provider 块和 `config.toml`
  的顶层默认值/范围），由 `stash.json` 兜底。`providers.starling` 里的自定义字段
  在 sync 期间保留，断开时原来的 provider 块会被恢复。一个被改名的 provider 的
  `enabledModels` 里 `starling/<old>/` 条目跟着它，glob 或思考等级也包括在内，为
  断开记住的范围也跟着；用户的其他条目保留。断开恢复原来的默认值，保留用户自己的
  原生改动和自定义范围编辑，并且只在为空时移除新建的文件。
- MothX 的目录就是 `$MOTHX_DIR`（没有 `~`），否则 Windows 上是
  `%APPDATA%\mothx`，否则是 `~/.mothx`；一个 WSL 发行版的是它的 `~/.mothx`
  （[`mothx.go`](../../internal/agent/mothx.go)）。它的 `thinkingFormat` 默认是
  `openai`，除非用户设了别的：为时，MothX 按模型 id 挑 Qwen 的
  `enable_thinking`、DeepSeek 的 `thinking` 或 Doubao 的等级，而这些 id 对星鸣
  来说并不说明提供它们的厂商。MothX 的 `/model` 和 `/statusline … global` 会重写
  整个文件，所以在那之后断开要拿走星鸣的 provider、这次挑选，以及一个被留空
  的 `providers`（MothX 会省略空的），其余保留。
- Zoo Code 的 provider profile 存在 VS Code 的 secret storage 里，所以星鸣为它
  写一个设置文件（`starling-import.json`，在 Zoo Code 的
  `globalStorage/zoocodeorganization.zoo-code/settings` 里，不在星鸣的目录里，
  这样一次 sync 会保留用户挪过去的地址），并在 VS Code 的 `settings.json` 和每个
  profile 的设置里把 Zoo Code 的 `zoo-code.autoImportSettingsPath` 指向它，用户自己
  的路径被 stash 并在断开时放回（[`zoocode.go`](../../internal/agent/zoocode.go)）。
  Zoo Code 每次启动都会导入它（`importSettingsFromPath`），把它合并到用户的 profile
  之上，并让 `starling` 成为当前那个。断开停止导入并删除该文件；已经导入的 profile
  仍留在 Zoo Code 里，星鸣到不了那里。
- 一个 agent 通过它的配置文件、它的目录，或它在星鸣 PATH 上的命令被检测到
  （`Detected`）。那个 PATH 是登录 shell 的加上用户工具所在的目录
  （`proc.UserPath`：`~/.local/bin`、npm 的各 prefix、Node nvm 的默认 alias 解析到的
  位置（没指定时是装得最高的那个）、bun、Homebrew、……），所以一个终端找不到的
  CLI 仍然能被找到并列出：装在一个终端已经不再使用的 Node 下的 npm 包。
  `/api/agents/cli` 随后给出它在哪里（`path`），以及当登录 shell 的 PATH
  （`proc.LoginPath`）缺它的目录时的 `offPath`，Agents 那一行会说 "Not on your
  terminal's PATH"，它的提示会点名那个文件（悠悠哥 报告：Codex 0.162.1 被列出
  时 PATH 上没有 codex，也没有 `~/.codex`）。未知的登录 PATH 什么都不说。另一个工具
  也写的目录不算（`dirShared`）：Pi 的那个被 omp 用，`~/.copilot` 被 Copilot CLI 用
  （GitHub Copilot 的 JetBrains 和 VS Code 插件也写它），Qoder 和 Qoder CN 的那个被
  Qoder 的 JetBrains 插件填充，`~/.cursor` 被 Cursor CLI 用（Cursor 编辑器自己会
  建它）（`TestCursorFolderFromTheEditorIsNotTheCLI`，@G_Z_Y 报告）；只有
  `cli-config.json` 或 `cursor-agent` 才算 Cursor CLI。一个别的工具先写的配置文件，
  只有在 `pathOwn` 说它是该 agent 的自己的时候才算：只装着 `enabledPlugins` 的
  Qoder `settings.json` 是插件的（`qoderPluginOnly`；#1495，
  `TestCopilotFolderFromAnIDEPluginIsNotTheCLI`、
  `TestQoderSettingsFromAnIDEPluginIsNotTheCLI`）。Qoder app（桌面 app，它把
  Qoder 的 CLI 作为自己的 agent 用同一个目录）通过它写在那里的
  `.qoder-app-status.json` 找到（`sign`，`qoderApp`；#2138，`TestQoderAppIsQoder`），
  因为它的 `settings.json` 也可能只是它自带的插件的 `enabledPlugins`。只有 GUI 和
  `daemon install` 会这样扩大 PATH；`starling serve`、`web`、`tui` 和 CLI 保持它们
  启动时的那个，所以一个 agent 的 CLI（`cliBin`，它的版本和 `omp` 的最大 effort 都
  从那里读）在每个系统上都会在这些目录里找（`proc.FindTool`），而不只在 Windows
  上（whqtian 报告：Ubuntu 上 `~/.bun/bin` 里的 omp 被读成没有版本，于是提供
  了 xhigh 而不是 max；`TestOmpVersionOffPath`）。
- 检测是按一个 agent 的目录走的，而卸载会留下那个目录，所以只靠它找到的 agent 会
  显示为已不在（`CLIMissing`；#843、#2177、#2173）：它的 CLI 不在任何 PATH 上，
  也不在 `proc.FindTool` 的任何目录里，更不在它自己的安装器放的地方
  （`~/.claude/local`、`~/.codex/packages/standalone`、`~/.opencode/bin`、
  Hermes 的 `hermes-agent` 检出目录），而且这里没有别的东西会读那个目录
  （`cliHeld`）：它的桌面 app（`HasApp`）、编辑器的扩展
  （`~/.<editor>/extensions/<id>-*`：`anthropic.claude-code`、`openai.chatgpt`、
  `saoudrizwan.claude-dev`）、Mac 上的 Claude.app、OpenCode.app 和 Goose.app、
  JetBrains IDE 的设置（它的 AI Assistant 的 agent 和 Cline 的插件）或 Zed 的
  `external_agents`（它的 Claude Code 和 Codex 适配器）。在星鸣分不清的地方
  （每个共享目录，`cliShared`，在 Windows 上，那里的读者可能是 Store app、便携
  编辑器，或者只有 shell profile 才设置的 PATH，#1482；以及那里一个被该 agent 的
  某个桌面 app 读到的，`desktopApps`：DeepSeek Harness Desktop 可以装到任意目录
  并把它的 `dsh.cmd` 放在用户挑的某个目录里，#2300、#2304；不在 Mac 上的 OpenCode
  和 Goose、Gemini、一个用自己方式或用其他工具共享的命令名检测到的 agent、一个
  没有安装命令但有 `cliOnly` 的 Hermes），它就从来不是"不在了"。GUI 里一个
  "不在了"的行显示 "CLI not found" 并且不显示它的任何设置，也没有 Apply；而在连接
  状态下它保留自己的开关和 Disconnect，并且在没有设置之后它折叠在 Not set up
  下面。Get started 永远不提供它（`connectKind` 是 "no"）。TUI 把它排除，而
  `starling ls` 说 "CLI not found"（`TestCLIMissing*`、
  `agent-cli-missing.test.cjs`、`getting-started.test.cjs`）。
- 测试从不碰真实机器。`TestMain` 使用 `testenv`，它把 HOME、USERPROFILE、XDG 目录
  和 APPDATA 指向一个临时目录，清空 `agentenv.Vars`，并把钥匙串工具和 agent CLI 的
  失败替身放到 PATH 上。agent 特有的副作用（Aside、Zed 钥匙串、Codex 的用量检查）
  也在 `TestMain` 里被打桩。

## 验证

<a id="verification"></a>

```sh
go test -tags nogui ./internal/agent -run '^TestStepCode'
go test -tags nogui ./internal/agent -run '^TestMothX'
go test -tags nogui ./internal/agent -run 'TestClaudeAutoModeReview|TestT3CodeAutoModeReview|TestConnect|TestDisconnect|Drift|TestKeptAddress|TestWSLAgentAddress|TestReachProbe|TestDryRunKnowsTheAskersProviders|TestReseat|TestPort|TestWSL|TestClaudeKeepsItsClaudeAISignIn|TestClaudeSignIn|TestClaudeOwnEndpointWithAnEmptyKey|TestClaudeAtAnOlderGatewayAddress'
go test -tags nogui ./internal/agent
go test -tags nogui ./internal/agent -run '^TestAir'
go test -tags nogui ./internal/agent -run 'TestReverseProxy|TestClaudeManagedProxyIsStillStarling|TestStarlingProbe|TestCodexBehindReverseProxy|TestCodexSignedInBehindReverseProxy|TestAwayPath|TestClaudeBehindLocalProxyKeepsIt'
go test -tags nogui ./internal/agent ./internal/edit -run 'TestSyncKeepsAnAddressOnAnotherMachine|TestKeptAwayOnlyWhatIsSure|TestKeyValuesFormats|TestFilteredWrites|TestOwnWritesDuringASyncWithALinkedConfig'
go test -tags nogui ./internal/edit -run 'BOM|CRLF'
go test -tags nogui ./internal/agent ./internal/proc -run 'ClaudeCloud|TestT3CodeSaysAClaudeCloudSwitch|TestLoginEnvClaudeProvidersNotSet'
go test -tags nogui ./internal/agent -run '^TestAnte'
go test -tags nogui ./internal/agent -run '^TestAstrBot'
go test -tags nogui ./internal/agent -run '^TestJcode'
go test -tags nogui ./internal/agent -run '^TestRaycast'
go test -tags nogui ./internal/agent -run '^TestOpenPencil'
go test -tags nogui ./internal/agent -run '^TestGrokApp'
go test -tags nogui ./internal/agent -run 'ClaudeScience'
go test -tags nogui ./internal/gateway -run 'TestClaudeScience'
go test -tags nogui ./internal/agent -run '^TestPlugged'
go test -tags nogui ./internal/agent -run 'TestCodexFailoverOnlyKeepsOwnModels|TestCodexOwnModelThroughStarlingWhileAccountsOn|TestCodexHandWrittenBaseURLStays|TestCodexProxyFor|TestRealProxyProbe|TestCodexEnvTakesLoginShell|TestCodexFailoverHeldWhileProxyDoesNotPassOn|TestCodexProxiedDrift|TestCodexFailoverEndsWithStarling|TestCodexMirrorFailoverHeldAndEndsWithStarling'
go test -tags nogui ./internal/agent -run 'TestDisconnectStaysThroughSync|TestCodexDisconnectStaysWithAccountsOn|TestCodexAPIDisconnectStays'
go test -tags nogui ./internal/tune ./internal/agent -run 'TestCompact|TestCacheTTL|TestTune'
go test -tags nogui ./internal/sessions -run 'CompactTook'
node --test internal/gui/tests/context-tune.test.cjs
node --test internal/gui/tests/agent-connect.test.cjs internal/gui/tests/agent-disconnect-preview.test.cjs internal/gui/tests/claude-signin.test.cjs internal/gui/tests/agent-failover.test.cjs internal/gui/tests/agent-proxied.test.cjs internal/gui/tests/plugin-agent.test.cjs
```

## Lody 导入交接

<a id="lody-import-handoff"></a>

[`lody.go`](../../internal/agent/lody.go) 通过它的 home 目录检测 Lody
（`~/.lody`、`~/.lody-oss`），但只在那个目录里放着 Lody 建的某个条目时才这样
（`workspace-catalog.json`、`loro-repo`、`agent-binaries` 或 `code-collab-v2`），
所以另一个工具建的同名空目录不是 Lody；也可以通过 OS 配置目录下的 Electron 数据
目录检测（Lody、Lody OSS、Lody Nightly：`~/Library/Application Support`、
`%APPDATA%`、`~/.config`）；或者通过一个 id 以 `ai.lody.` 开头的 macOS bundle
（真正的 app 是 `ai.lody.desktop`）。一个从未打开过的 Windows 或 Linux 安装还没有
数据目录，不会被显示。它那一行打开
`lody://provider/import?v=1&data=<base64url JSON>`，带着星鸣的信封和七个
运行时目标。只提供回环 HTTP 网关；key 始终是公开的归因 token `starling-lody`，
绝不会是局域网的机密。当 `STARLING_ADDR` 固定在回环之外时没有链接：那一行没有
按钮，`starling lody add` 会说明原因并且什么都不打开
（`TestLodyLinkCarriesNoSecret`）。由操作系统选择已安装的 Lody 处理器。一个兼容的
Lody 构建必须实现这条导入路由。

Lody 会问要添加哪些 Providers，用 `-starling` 命名它们，并且只在用户确认之后才写到
当前桌面版的本地机器上。星鸣不读写 Lody 的私有文档存储，不在那里选择模型，
也不声称打开链接就完成了导入。Lody 没有公开的状态读取器：`Agent.Added` 保持为
nil，JSON 那一行省略 `added`。GUI 对这个未知状态显示一条确认提示；Cindy 仍会
报告它已知的真/假状态。重新打开链接能让 Lody 处理重复项。Bub 需要它单独安装的
ACP plugin。
