# 网关 session 聚合

<a id="gateway-session-aggregation"></a>

Sessions API 把以文件为后端的原生 session 和网关使用记录的只读投影合在一起。
一个网关 session 的身份恰好是 `AgentOf(record.Agent) + record.Session`，其中
session 是客户端发来的那个（`X-Magpie-Session`，或它自己的 session 头），无论
它以什么开头。响应在 `X-Magpie-Session` 里返回这个 id；列出的浏览器 origin
可以通过 CORS 读取这个头。崇云科技绝不为没有 session 的请求编造 id，所以这样的
请求没有 session，它的文本也不会被记录：挂在一个没有任何列表能列出的 id 下的
文本是无法打开的。prompt 文本、caller key 和网络地址都不会被用来猜对话归属。
因此崇云科技写下的配置让每个 agent 各自声明自己的 session：Pi 的 provider 块
设置 `compat.sendSessionAffinityHeaders`（Chat 和 Messages 上是
`x-session-affinity`；Responses 本来就发 `session_id`），omp 16.0.6 及以后则得到
`compat.promptCacheSessionHeader: x-session-affinity`。用户自己的
`sendSessionAffinityHeaders` 会被保留（`piBlockKept`）。验证：
`TestPiNamesItsSessionToTheGateway`、`TestOmpNamesItsSessionToTheGateway`。
没有 `Session` 的旧账本记录仍然被排除。这个投影是用量账本算出来的，不做持久化。
窗口投影和路由汇总复用带版本的日志快照；追加和重写会让它们失效，价格/provider
的变化也会让带价格的投影失效。单 session 查询使用缓存的身份索引。原生排除项
按调用方分别应用。

WorkBuddy 和 CodeBuddy Code 用 `X-Conversation-ID` 命名它们的对话。网关把它当作
一个原生 session 头读取，在已有原生头之后；`X-Magpie-Session` 仍然覆盖被记录的
session，但不会改变它的原生身份。每轮的 `X-Conversation-Request-ID`、每条消息的
`X-Conversation-Message-ID` 和 `X-Parent-Conversation-ID` 都不标识当前 session。
这让 Routing 能按原生 WorkBuddy session 读取器已经知道的同一个 id 给请求分组；
没有记录 session 的旧请求仍然无名。见 `TestWorkBuddySessionHeader`、
`TestWorkBuddySessionThroughGateway` 和 `TestRemoteMagpieKeepsTheSession`。

DSH 用 `x-deepseek-harness-session-id` 命名它的 session，同样在其余之后读取，
并用 `x-deepseek-harness-compact` 标记一个会压缩上下文的请求（#2505）。
`callKind` 把这样的调用记为 `compact`，即 Requests 列表里的 Compaction；取值
`0`、`false`、`no` 或 `off` 表示没有标记，而没有这个头的请求绝不会因为它的
prompt 被当成一次压缩。`passOnCaller` 会把这个标记连同 session 一起转给远端的
崇云科技，也不会转给任何厂商（`TestDshMarksSessionAndCompaction`）。

当同一个身份出现在一个原生 session 里时，原生 session 优先，网关投影被省略。在
`/api/sessions/manage` 里、在它的计数中、也在 `/api/sessions/delete` 里，一个
原生 session 就是 agent 保留的任何 session 文件，包括没说过话的那个，没有数量
限制（`sessions.NativeIDs`）；对各 agent 文件的一次遍历（`sessions.Manage`）只读
被展示的那个 agent 和那些有网关 session 的 agent（#2776）。网关投影没有路径、
没有恢复命令也没有终端动作：agent 自己那些 session 的文件不在本机上。它们那几行
带着 `gateway: true`；Sessions 页面把它们归到 "Seen through the gateway" 而不是
"No folder" 下，而它们的大小就是崇云科技为它们记录下来的文本
（`sessions.GatewaySizes`），只有在有文本时才显示。概览顶部的汇总
（`MergeExternalStats`）同样带着 `gateway: true`，而 Usage › Sessions 把这样的行
在没有文件夹的位置上标为 "Gateway session"，在没有保留文本时加上
"(prompt not recorded)"，这样它读起来不像一个 prompt 为空的原生 session
（#2285：一个 session 文件在临时目录里的 Pi subagent）。验证：
`TestTopSessionsMarkWhatOnlyTheGatewaySaw`、`sessions-overview.test.cjs`。它们的
模型、token 总量、开始/最后时间戳和有效价格，都来自 Usage 和统计端点所用的同一
批用量记录。

Claude Desktop 的网关行即使没有 CLI 记录也可以有本地名字（#1675）。
[`DesktopTitles`](../../internal/sessions/desktop_titles.go) 读取 Desktop 数据目录
下的 `claude-code-sessions/*/*/local_*.json` 和
`local-agent-mode-sessions/*/*/local_*.json`（`DesktopDataDirs`，也就是 session
读取器用的同一个 `desktopdir.Everywhere`，含 Claude-3p 和 Windows 的 MSIX 目录）。
记录的 `title` 点名 Desktop 自己的 LocalSessionManager 算作这个 session 的那些
CLI id：它实时的 `cliSessionId` 和 `unarchivedCliSessionId`，它血统上的
`preClearCliSessionId` 和 `priorCliSessionIds`，以及在没有 `cliSessionId` 时
它去掉 `local_` 的 `sessionId`。先前的运行共享当前的名字，但仍是各自独立的用量
身份。一个实时句柄优先于一个血统句柄，同一级上名字冲突的则是未知。一条在那里
却读不了或解析不了的记录（Desktop 正在写它）保留它上一次解析的结果；消失了的
会带走它的名字；一个都没有时，这一行保留记录下来的首个 prompt 标题或它的 id。
文件通过 `filememo` 读取，名字在再次查看这些文件之前会保持一秒
（`desktopTitlesHeld`），所以一个组装了几百个原生 Desktop session 的列表只统计
它们一次，而 Desktop 里的一次改名在一秒内就会显示出来。这些名字会到达原生
Desktop session（`assemble`）、每一个网关列表（`MergeExternal`）、统计汇总
（`MergeExternalStats`）、Routing 的 `/api/gateway/session-titles`（`Titles`）和
Context 页面。它不会给网关行带来恢复或原生记录的能力。验证：
`TestDesktopSessionTitles`、`TestDesktopSessionTitleIdentity`、
`TestDesktopSessionTitleHandles`、`TestDesktopSessionTitleInStats`、
`TestDesktopTitlesReadOncePerList`、`TestDesktopSessionTitlesOnPages`。

一个网关投影可以被删除（lc 报告）。`sessions.DeleteGateway` 把它记录下的
文本目录（`gateway-conversations/<date>/<identity hash>`）移进崇云科技的 session
回收站下的 `trash/sessions/gateway/`，附带一条标着 `gateway` 的备注，并把
`{agent, id, last}` 加到配置目录里的 `gateway-sessions-hidden.json`。`/api/sessions`
和 `/api/sessions/manage` 会把一个隐藏的 session 排除到那次最后的请求为止；更新
的请求又会把它列出来。用量账本不变，所以 Usage、`/api/sessions/one` 和统计端点
仍然保留它的花费。最后一分钟内有请求的 session 会被当作活跃而拒绝。Restore 会
把文本移回去（其间过期或被清空的条目会被跳过；同一天里又被记录的 session 会被
拒绝而不是被覆盖）并取消隐藏；它只接受存储里属于那个 session 身份目录的条目。
永久删除会让它保持隐藏。录音的承诺在回收站里依然成立：每小时的清理会抹掉回收站
里超过七天的网关文本，"Delete saved conversations…" 也会抹掉它（备注留着，所以
Restore 仍能列出那个 session），而 `HasGatewayConversations` 会统计它。
实现：[`internal/sessions/gateway_delete.go`](../../internal/sessions/gateway_delete.go)、
[`internal/sessions/manage.go`](../../internal/sessions/manage.go) 里的 `Restore`，
以及 [`internal/gui/sessions_manage.go`](../../internal/gui/sessions_manage.go)。
全量统计包含比原生文件更早的网关历史。日历活跃度和输出总量使用每个 session 的
实际请求日期，包括中间的日期；扩展范围会以现有的原生日历索引为基准重新计算。

实现：[`internal/usage/gateway_sessions.go`](../../internal/usage/gateway_sessions.go)、
[`internal/sessions/external.go`](../../internal/sessions/external.go)，以及
[`internal/gui/sessions.go`](../../internal/gui/sessions.go)。

## 本地网关对话录音

<a id="local-gateway-conversation-recording"></a>

Sessions 页面提供一个明确的、默认关闭的录音开关。确认时会说明 prompt、回复和
工具结果可能包含私有文件和代码。这个开关是 session 列表上方的一张卡片，说明保留
什么、保留在哪里（仅本机、最多 7 天和 256 MiB）。它只在它适用的地方显示：某个
列表里有网关 session 的 agent 上（`/api/sessions/manage` 那一行上的 `gateway`）、
网关模式下，或者录音开着或文本被保留时。它的 "Delete saved conversations…" 在
`/api/sessions/manage` 的 `recorded` 说有文本被保留时显示
（`sessions.HasGatewayConversations`，目录读不出来时为真），并且先在崇云科技
自己的对话框里询问。`Settings.GatewayConversations` 是机器本地的同意：
`settings.SetGatewayConversations` 在设置文件锁下更新它。其它设置保存会在同一把
锁下保留最近存下的同意，这样一份更旧的快照不能在录音停止后把它重新打开。
`KeepOwn` 也会阻止同步/恢复去打开它。
`POST /api/sessions/recording` 改变这个开关；`clear: true` 关闭它并且只抹掉本地
录下的网关内容（存储和回收站里已删除 session 的文本），用量和原生文件不动。
已有的浏览器模式鉴权 guard 保护着这些 API；这是一个管理员视图，不是按 caller key
的访问控制。

[`recordConversation`](../../internal/gateway/conversations.go) 在协议转换和网关
middleware 之外包装 `serveAgent`。它记录一次面向客户端的往返，而不是每次重试，
并且只记在客户端点名的那个 session 下，使用与用量投影相同的 `usage.AgentOf`
身份，于是每个被记录的对话都有一行能打开它的 Sessions 行。Codex 直接的
`/backend-api/codex/responses` 中转也会被捕获。不经过这些入口的内部辅助调用、
token 计数、embeddings、模型列表和压缩路径不会被记录。录音不添加任何头也不改动
任何头，所以路由和 prompt 缓存亲和性看到的就是客户端发来的请求。已有的协议
读取器处理 Chat、Anthropic、Responses 和 Gemini；流式回复使用已有的解码器，
已完成的 Responses 条目保留自定义工具调用。只有真正经由网关发出的内容是可用的。
图片以占位符显示；隐藏的推理、仅客户端的操作和不受支持的线上条目无法重建。这
不是原始请求归档，也不是远端 session 同步，已有的 S3 归档不会被导入。

[`SaveGatewayTurn`](../../internal/sessions/gateway.go) 把归一化、擦除过的往返存进
崇云科技的配置目录下的每 `gateway-conversations/<UTC date>/<identity hash>/` 一个
只追加的 `turns.jsonl`，目录用 0700、文件用 0600。每一行是一个有界的轮次；
旧的按轮次分的 JSON 文件在过期之前仍可读。一次没写完的末尾追加不会进入读取
（被标为截断），并在下次追加前被移除；完整的更早的行仍然可用。不存储任何头。
已知的机密和名字像机密的 JSON 字段会被擦除，但任意的敏感文本不保证能被移除。
响应捕获使用一个 8 MiB 的临时缓冲；输入同样限制在 8 MiB。归一化后的轮次在磁盘
上限制为 1 MiB，单个部分为 64 KiB，每个每日 session 文件为 256 MiB。某个每日
session 文件满了之后会拒绝当天进一步的录音。每小时的后台清理移除过期的桶和
超过 256 MiB 的最旧文件；这是两次清理之间的软性存储上限。保存一个轮次不会重新
扫描整个存储，后台的遍历和排序也不持有保存锁。捕获、解析或持久化失败不会改变
模型响应；文件系统失败会被记录，但不记对话正文。

超过七天的内容不会进入读取。提供服务的网关在启动时和每小时清理过期的 UTC 日桶，
录音关着时也一样；按天分目录意味着物理清理可能比读取截止晚一天。关闭并清除与
写入共用一把锁，并重新检查同意的代次，这样即使在那个请求完成之前录音又被打开，
一个在途请求也无法把已清空的存储重新填满。

`GET /api/sessions/transcript` 仍然优先匹配一个原生 session。否则它读取一个已知
网关 session 最后保留的 128 次往返。一段完整且完全一致的历史前缀会从下一个请求
里去掉，而重复的新消息会被保留。系统/工具定义的上下文在未变化时单独折叠。被
压缩、编辑或增量的上下文可能重复内容；它们绝不会被猜成一段唯一的标准历史。已有
的记录显示限制仍然适用，截断会被标出。空的/过期的/缺失的内容有明确的空状态，
而所有网关记录都标明自己的来源。页面在重新加载或清除时作废待处理的记录读取。
一次完成的读取在重绘后更新当前打开的记录，而更旧的响应不能替换清除存储之后才
取到的内容。

`GET /api/sessions/markdown` 和 `POST /api/sessions/export` 做同样的选择：原生
session 从它自己的文件写出（`sessions.WriteMarkdown`），否则一个已知的网关
session 按它的记录所读到的东西写出（`sessions.WriteGatewayMarkdown`，
"Recorded by: magpie's gateway"，并在轮次或某个部分的结尾没有被保留时加一条
备注）。一个被别人当作远端崇云科技访问的崇云科技没有那些 agent 的 session 文件，
所以这是它导出它们的唯一办法。验证：`TestGatewaySessionExportsMarkdown`。

一个没有保存 Desktop 名字的网关 session，按它最初被问的内容命名（YOU 报告），
而不是按它的 `ses_…` id：`sessions.GatewayTitle` 读取它在这 7 天保留期内最旧的
那个记录文件的前 8 个完整轮次，取第一段不是 `<…>` 上下文标签、不是
`# AGENTS.md` 也不是图片的用户文本部分，并把它做成一行（`untagged`）。一个已完成
的文件的答案按路径缓存；一个还在增长、尚未有标题的 `turns.jsonl` 会被再读一次。
完全什么都没记录时（存储的目录缺失）它只花一次 stat。`gatewayListed` 在每一个
列表里应用它（`/api/sessions`、`/manage`、`/one`、统计汇总）；一个既没有保存
Desktop 名字也没有记录文本的 session 保留它的 id 作为标题。

`GET /api/sessions/search?agent=&q=` 按说过的话找出某个 agent 的 session，供
Sessions 页面的筛选使用。查询会被转小写并把空格合并成一个（`sessions.NeedleOf`）；
少于 2 个 rune 时它什么也不搜。文件会被读取的 agent（`HasTranscript`）的原生
session 从它们自己的文件读取（`SearchSession`），先经过一个字节级预筛
（`mayHold`：查询里最长的词，ASCII 大小写折叠且不复制，按 1 MiB 分块；对 zstd
文件跳过，对 JSON 可能转义的查询跳过，以及对一个其路径不是对话所在文件的
session 跳过，`linesFile`：一个 OpenCode session 的信息文件或数据库，一个
Cursor 对话的 SQLite 存储，其最新页在它的 `-wal` 里；`search_db_test.go`），
然后只解码到用户或模型所说内容中的第一个匹配处。对 Claude 或 Codex 的文件，
只有包含查询里每个词的那些行会被解码（`readTranscriptOf`，一个部分就是一行）；
Pi 的目录整体读取。没有这个机制时，一个 1.3 GB 的 Claude session 就要占满 8 秒；
有了它，用户的全部 58 个（3 GB）在负载满的机器上 3–7 秒就能搜完。网关 session，
以及文件不被读取的 agent 的原生 session（那时是 `unread: true`），在它们记录的
转录里搜索（`SearchGateway`）。系统 prompt、工具定义（`role: system`、`kind:
context`）和图片绝不会被匹配。session 由最多 4 个 worker 按最新优先读取，在
200 个命中或 8 秒时停下（`cut: true`，带 `searched` 和 `total`）；请求的 context
会取消读取，所以页面下一次查询会结束上一次查询。每个命中带上匹配处以及它前面
最多 36 个、后面最多 120 个 rune，截断的末尾标上 "…"。页面在最后一次击键后等待
300 毫秒，中止在途请求，立刻显示标题、文件夹、id 或模型匹配，把搜索到的 session
连同片段加上（匹配处放在 `<mark>` 里，作为文本节点构建），并说明它是在搜索、
提前截断还是仅限已记录的文本。
实现：[`internal/sessions/search.go`](../../internal/sessions/search.go)、
[`internal/gui/sessions_manage.go`](../../internal/gui/sessions_manage.go) 里的
`searchSessions`、
[`internal/gui/assets/plugins/sessions/client.js`](../../internal/gui/assets/plugins/sessions/client.js)。
验证：`TestSearchSession*`（`TestSearchSessionReadsOnlyLinesWithTheWords` 对应行
过滤）、`TestContainsFold`、`TestMayHoldSkipsAFileWithoutTheWords`、
`TestGatewayTitleAndSearch`、`TestSessionsSearch*`；浏览器
`sessions-search.test.cjs`。

验证：`TestGatewayConversation*`（`TestGatewayConversationNeedsTheClientsSession`
和 `TestGatewayConversationClientRequestPrefixedSession` 对应上面两条规则）、
`TestGatewaySessionHistoryAndCalendar`、`TestGatewaySessionsNativeWinsBeforeLimits`、
`TestGatewayStatsRebaseNativeCalendar`、`TestCORSKeyThroughTheServer`、
`TestSessionsDeleteGatewaySession`、`TestGatewayDeleteTrashExpires`、
`TestGatewayPurgeKeepsItHidden`、`TestGatewayDeleteNothingRecorded` 和
`TestGatewayRestoreRefusesForeignItems`；浏览器 `gateway-conversations.test.cjs`、
`sessions-talk.test.cjs`、`sessions-gateway-delete.test.cjs`，以及 locale 检查。

## 经由网关的用量，以及 subagent

<a id="usage-over-the-gateway-and-subagents"></a>

`GET /v1/magpie/usage` 和 `GET /v1/magpie/usage/requests`
（[`internal/gateway/usage_api.go`](../../internal/gateway/usage_api.go)）用与
窗口的 `/api/usage` 和 `/api/usage/requests`
（[`internal/usageapi`](../../internal/usageapi/usageapi.go)）相同的构建器，提供
Usage 页面的汇总和请求账本，只读且不含对话文本。它们遵循 `/v1/magpie/quotas`
的 guard：本机，或者另一台持有启用了的网关 key 的机器，在同样的 Host 检查之后。
一个被持有的 key（`heldKey`：一份预算、一些模型或一些账号）会被拒绝访问汇总，
并让账本按它自己的 `CallerKey` 过滤，筛选列表也裁剪到它自己那几行
（`usageapi.OwnPage`）：账本的分面来自这段时间内的每一次调用，不裁剪的话它们会
提到别的 key。`TestUsageOverGateway` 覆盖这两种情况，去掉任一个裁剪都会变红。
一个被持有的 key 会被拒绝（`refuseHeld`，403 `provider.KeyHeldCode`）访问配额、
它们的历史、保留的卡片和刷新，以及 `/v1/magpie/concurrency`（#1981）；
`/v1/magpie/route` 只告诉它那些 `callerKey` 是它自己的路由。`Route.callerKey`
由 `trace.beginFor` 从请求的调用方设置，所以一条没有它就开始的路由不会告诉任何
被持有的 key。`TestHeldKeyToldNoAccounts` 从回环和局域网两种情况覆盖这里每一条，
针对三种方式各自被持有的 key 和一个被持有到什么都不剩的 key。
一个被持有的 key，若其所有者设置了 `access.Key.Quotas`（#2216），会通过
`scopeQuotas`（`quotas_scope.go`）被告知那四条配额路由：那些每一个目录模型都被
`modelAllowed` 允许使用的 provider 的卡片，而在这些 provider 之中，只有那些被
`AllowsAccount` 允许使用的账号，按卡片的 `User` 与 `AccountIDs` 匹配（两个账号
共用的名字必须都被允许；一个没有账号与之对应的卡片，只在该列表没有提到任何
该 provider 的账号时才允许）。远端卡片（`From`）绝不会被告知。刷新一张在范围
之外的卡片会被拒绝，一个没点名任何账号的只读该范围的卡片，而一个没有卡片的
范围会以"被持有"来回答卡片路由。`TestHeldKeyToldItsOwnQuotas` 覆盖每一条路由：
针对一个 provider 范围的 key、一个拥有某 provider 部分模型的 key、一个按账号
被持有的 key、某 provider 唯一的 key，以及一个什么都被告知的 key，去掉它对应的
检查时每个都会变红。

### 从 agent 自己的文件读到的调用

<a id="calls-read-from-the-agents-own-files"></a>

一个 agent 在没有崇云科技时做出的调用，只有当它自己的文件被读取时才到达 Usage
（`sessions.Calls`，来自 [`calls.go`](../../internal/sessions/calls.go) 里的
`callSources`）。读取器有 Claude Code 的、Claude Desktop 的和 Cowork 的、Codex
的（以及 WSL 里两者），还有一次读一个文件的（[`call_cache.go`](../../internal/sessions/call_cache.go)
里的 `wholeCallReaders`）、OpenCode 的（#680）、ZCode 的、DeepSeek Harness 的
（`dsh`）、WorkBuddy 的、CodeBuddy Code 的（#1779；它的 session 也在 Sessions
页面上，来自同样按 WorkBuddy 那样读的 `projects/<cwd>/<id>.jsonl` 文件，可用
`codebuddy --resume <id>` 恢复并且可删除，#2657：`TestCodeBuddy`、
`TestCodeBuddySessionsListedUnderTheirName`）、MiniMax Code 的（#1605）、Kimi
Code 的（#1607，[`calls_kimi.go`](../../internal/sessions/calls_kimi.go)）以及 Pi
的和 omp 的，在这里和 WSL 里都有（#1951，
[`calls_pi.go`](../../internal/sessions/calls_pi.go)）。每一个都会跳过该 agent
通过崇云科技自己的入口做出的那些调用，那是网关用应答的账号记下的：Kimi Code 的
是那些 `usage.record` 模型以 `magpie/` 开头的。Kimi Code 的调用是它每个 agent 的
`wire.jsonl` 里的 `usage.record`，位于 `$KIMI_CODE_HOME/sessions`（否则
`~/.kimi-code`）下，协议 1.4 和 1.5；一个 fork 会重复它父级的记录，所以每一条
按 agent、时间、模型和用量为键，只计一次。Pi 和 omp 的调用是 session 文件里的
assistant 消息，以及 Sessions 也一并统计的那些调用（"usage" 和 omp 的
"model_usage" 条目、一次压缩的、一个分支摘要的和一个工具结果的用量，用在当时
的模型和 provider 上）；provider `magpie`（或 `dial`）是崇云科技的入口，一次
失败的调用没有用量，而 fork 复制来的它父级的条目（写在其头之前）属于父级。
Sessions 页面读取的其他 agent（Cline、Qoder、Grok、Droid、Cursor、Hermes、Alma）
还没有读取器（#1607）。Context 页面读的是网关自己的 trace，所以它不显示这些
调用中的任何一个。验证：`TestMiniMaxCodeCalls`、`TestKimiCode*`（报告者真实的
wire 文件）、`TestCodeBuddyCalls`、`TestWorkBuddyCalls*`、`TestPiCalls*`
（Pi 1.1.0 和 omp 18.9.1 自己的文件）和 `TestWSLSessionsListed`。

一次对所有调用的读取会解码每个文件缓存的分片，在长历史上那是几万个文件
（#2278：68k 个分片，415 MB）。在一次读取进行中对这些相同调用的请求会等它，并
拿到它答案的副本（`sharedCalls`），而读取一次只跑一个。`AgentCalls` 只读一个
agent 的文件。Sessions 概览里的 Codex 模型和 effort（`usage.CodexModelEffort`）
只由 Codex 的调用和网关的记录构建，因为只有它们点名了一个 Codex 行为者。它每个
周期同时只构建一次，而构建期间的请求会拿到同一台主机上一次的答案。当 usage.jsonl
缺失或在构建期间增长时答案会被保留，最多保留 8 个周期。页面对已经在等的概览只
问一次（`plugins/usage/page.js` 里的 `sessOverview`）。在这之前，网关写入期间的
每个请求都会重建整个账本，app 增长到了 85 GB。验证：
`TestConcurrentCallsShareOneRead`、`TestAgentCallsReadsOneAgent`、
`codex_effort_shared_test.go` 里的 `TestCodexModelEffort*`（只有 codex 的答案
等于整个账本的答案），以及 `sessions-overview-once.test.cjs`。

### app 自行上报的调用

<a id="calls-an-app-reports-itself"></a>

`POST /v1/magpie/usage/requests`
（[`internal/gateway/usage_ingest.go`](../../internal/gateway/usage_ingest.go)，
#1874）接收那些从未经过崇云科技、由某个 app 直接发给它的厂商的调用，于是 Usage
页面也覆盖它们。请求体是一条记录或最多 5000 条的数组（最多 32 MiB），形状与
`usage.jsonl` 相同：`t`、`agent`、`provider`、`model` 和 `rid` 是必需的；`in`、
`out`、`cache_read`、`cache_write`、`cache_write_1h`、`reasoning`、`ms`、
`ttft_ms`、`first_text_ms`、`flow_ms`、`status`、`cost`、`effort`、`session` 和
`err` 是可选的。任何其它字段、一个带引号的数字、一个负数或超范围的计数、一个
比当前早于或晚于 2000-01-01T00:00:00Z 超过一小时的时间，都会让整个提交被一个
400 拒绝，该 400 点名那条记录（`error.index`）和那个字段（`error.param`），什么
都不会被写入。一个过去的 `t` 一律照收，无论多旧，这样客户端的历史可以回填。

- **谁可以提交：** GET 路由的那条规则，通过同一个 `lanGuard`：本机，或者共享
  期间另一台持有启用了的网关 key 的机器。除此之外，一个浏览器页面（`fromBrowser`）
  无论在哪都需要一个 key，因为 `corsGuard` 放行一个 localhost 页面时不需要 key
  （`Origin: null` 在那里也需要 key，#2242），而一个被持有的 key 会被拒绝。
  `TestUsageIngestAccess`。
- **写入：** `usage.IngestReported`
  （[`internal/usage/reported.go`](../../internal/usage/reported.go)）在和
  `usage.Append` 相同的互斥锁下、往同一个文件里追加，每一行带
  `source: "reported"` 且没有 caller key。一个日志里已经作为上报记录存在的
  `rid`，或这次提交里重复的 `rid`，会被跳过。上报的 rid 集合是从日志本身读出来
  的，所以它能活过重启：重启后的第一次提交扫一遍日志，只保留带上报标记的行，
  之后每次提交只读此后追加的部分。回复会统计 `accepted` 和 `skipped`，并给出
  每条记录的 `index`、`rid` 和 `status`，所以重跑一批是安全的。
- **定价：** 一条带 `cost` 的记录保留它（`Record.OwnCost`，打包行上的
  `rowOwnCost` 标志）。没有它就按崇云科技自己的价格定价，和网关行一样。
- **速度：** `first_text_ms` 和 `flow_ms`（#2637）的含义与 `usage.jsonl` 里一样
  （`Record.FirstText`、`Record.Flow`），所以一条上报行的速度就是
  `usage.DecodeOf` 给出的，正如网关行的一样，在服务端（`Row.Decode`，汇总的
  `DecodeN`）和在 `page.js` 的 `ledDecode` 里都是。一条做过推理的行在没有
  `first_text_ms` 时不报告任何速度；`ThinkPace` 和（有 `flow_ms` 时的）
  `HeldShare` 说明一个回复是成批放出的。`first_text_ms` 不能早于 `ttft_ms`
  （首段文本也是内容），也不能晚于 `ms`；`flow_ms` 是内容时间的中间 80% 乘以
  10/8，正如 `firstToken.flowMs` 所测量的，所以它被 `ms` × 10/8 而不是 `ms` 所
  界。`TestUsageIngestTellsSpeed`（报告者的记录）、`TestUsageIngestTimesChecked`。
- **展示：** 上报行计入每一个总量。它们还有一个自己的来源（`SourceReported`）：
  `RequestPage.Reported`、`?via=reported`，以及 Requests 标签页上的第四格和一枚
  "Reported" 徽章。
- **没有实时部分：** 一条上报行只出现在 Usage 页面上。`Record.IsReported` 让它
  不进入网关 key 的预算和花费上限（`budget.counts`）、订阅周期的占用（以及决定
  一个周期能否计算的那个日志起点，`firstOwn`）、Sessions 页面的网关 session、
  一个 session 被路由到了哪里（`Vias`），以及 Codex 封存的答案和记住的模型。
  除此之外不为它做任何事：没有 OTel span，没有限流，没有休息，没有告警。

`TestUsageIngest*` 和 `TestIngestReported*` 覆盖这些，包括在网关自己写入旁边的
并发提交，以及一整批 5000 条记录的回填落在正确的 7d/30d/all 天数上。
`TestReportedTouchesNothingLive` 和 `TestUsageIngestTouchesNoLimit` 在任何实时
guard 被移除时变红。浏览器：`usage-ledger-reported.test.cjs`。

在 20 万条上报行（一个 47 MB 的日志）下，40 次每次 5000 条的提交总共花了 0.4 秒。
重启后的第一次提交花 0.45 秒把 rid 读回来，之后一次约 2 毫秒。Usage 页面的第一次
读取会构建日志索引（2.1 秒，和它为网关自己的行做的一样）；之后一页全部需要
27 毫秒，全量汇总需要 215 毫秒。

一条用量记录的 `Session` 保持就是那场对话。一个 Claude Code subagent 的记录会
加上 `Subagent` 和 `ParentAgent`。经由网关时它们来自 Claude Code 为 subagent
的请求发送的 `x-claude-code-agent-id` 和 `x-claude-code-parent-agent-id` 头
（`subagentOf`，在 `appendUsage` 里读取）；主线程两个都不发。从 agent 自己文件
读出的一行，`Subagent` 来自文件名字，`<session>/subagents/agent-<id>.jsonl` 或
`<session>/subagents/workflows/<run>/agent-<id>.jsonl`
（`sessions.SubagentOf`），所以没有缓存版本变化，而 `ParentAgent` 来自它旁边
`agent-<id>.meta.json` 的 `parentAgentId`，那是 Claude Code 为一个 subagent 启动
的 subagent 写的（`sessions.SubagentParent`，通过 `ccMeta`：一个文件读一次；
一个没被读过的 meta 文件是未知，会再读一次）。一个匹配到文件行的网关行保留网关
自己的。两者的正确性都用 Claude Code 2.1.296 从一个 subagent 启动另一个 subagent
核对过。

这些读取器进入的目录可能是一个链接：符号链接，或者在 Windows 上是一个 junction，
Go 把它列为 irregular 而不是目录或符号链接。每个挑选自己要进入的条目的读取器
（一个 Claude Code、Qoder 或 Cowork 的项目目录、一个 session 自己的目录和一个
workflow 运行的目录、一个 Hermes profile、一个 Antigravity CLI 对话）都会问
`dirent.IsFolder`，它用 `os.Stat` 跟随这样一个条目；它们都不会递归穿过它。一个
是链接的遍历根会沿着它指向的地方遍历，它的文件仍以 agent 的目录名下命名
（`walkLinked`，#2233）。`TestIsFolder` 和那些 `*BehindALink` 测试会创建真实的
链接，在 Windows 上是 junction（#2268）。

Codex 的 session 身份是从原生 thread 元数据读出的，而不是从共享的 session 头或
`Kind`（`x-openai-subagent`）推断的。sessions API 增加 `role`（`main`、
`subagent` 或 `unknown`），并在原生元数据能确立时增加 `parent_session` 和
`root_session`。一个 root 未知的子级保留空的 `root_session`；没有显式身份元数据
的 Codex 条目、只有网关的条目和身份冲突的条目把 `role` 设为 `unknown`，不会获得
一个猜出来的父级或 root。这些字段对其他 agent 省略。这是 Codex 专属的（#1591）。

一个其 agent 把 subagent 的对话存在自己旁边的 session，会在 sessions API
（`/api/sessions`、`one`、`manage`；#2261，
[`subagents.go`](../../internal/sessions/subagents.go)）的 `subagents` 里列出它们，
但只对有转录的 agent，按最旧优先。每一个有 `id`、`title`（交给它的任务）、
`kind`（它以哪个 agent 身份运行）、`parent`（启动它的 subagent，对 session 本身
为空）和 `start`/`last`；它的文件绝不会被发送。

- Claude Code、Qoder 和 Qoder CN：每个 `subagents/agent-<id>.jsonl`（以及
  `workflows/<run>/` 下的），带 `agentType`、`description` 和 `parentAgentId`，
  取自它旁边的 meta 文件（`ccMeta`）。没有 meta 文件（Qoder），或有一个没点名
  `description` 的，标题就是文件里第一行用户内容（`ccTask`，最多读 16 行）。只
  在那时才为它读文件，而读到的东西（一个任务或没有）会被保留，直到文件的大小或
  时间变化；一次失败的读取保留上一次的结果。大小和时间是在读取之前取的，所以
  读取期间写入的一行会在下一次询问时被读到。`agent-acompact-<id>.jsonl` 是
  一次压缩自己的运行，被排除在外。
- OpenCode：一个子 session（`parent_id`，来自 JSON 存储或 `opencode.db`）是它
  root 的一个 subagent，而不是它自己的 session；`ocRoots` 记录它的父级。它的
  标题拆开 1.x 的 `<description> (@<agent> subagent)`；2.x 只用描述作标题
  （task 工具，opencode 2.0.25），而默认的 `Child session - <time>` 不算标题。
  OpenCode 2 在 `session_v2.agent` 里点名 agent（2.0.29），设置时它就是 kind，
  否则是标题里的 `@<agent>`（`ocV2DB.agent`，找到后保留；#2290）。ZCode 共用这个
  解析器但没有转录，所以它不列出任何东西。
- Kimi Code：每个 `agents/agent-<n>/wire.jsonl`。1.5 的 `subagent.spawned`
  记录写在启动方 agent 的 wire 里，给出 `subagentName`、`description` 和
  `parentAgentId`（`kimiWireOf`，只有当文件的大小或时间变化时才再读）；1.4 不写
  任何东西，标题是 subagent 的第一个 `turn.prompt`。在 subagent 自己的 wire 里，
  一个来自名为 `subagent` 的 `system_trigger` 的 `turn.prompt` 是交给它的任务，
  作为它的第一条用户消息打开，而不是被折叠的上下文（`kimiTask`，#2290）；其它
  system trigger，以及 `agents/main` 里的任何 trigger，都保持折叠。

这些是在解析过的 session 缓存旁边读取的，不在其中，所以它的版本不变。只有
Sessions 列表和某个 session 自己的查询会构建它们（`assemble`）；统计、原生
session 键、`Titles` 和 id 查询用的是 `assembleListed`，它在恢复、转录和 subagent
之前就停下，所以一次统计轮询不会读任何 subagent 的文件。
`GET /api/sessions/transcript` 接受 `sub=<id>` 来取一个 subagent 的对话；这个 id
会在该 session 自己的 `subagents` 里查找（`sessions.SubagentTranscript`），绝不会被
当作路径使用，而没有列出的一个是错误。见 `TestClaudeSubagents`、
`TestQoderSubagents`、`TestOpenCodeSubagents`、`TestOpenCodeV2SubagentKind`、
`TestKimiSubagents`、`TestKimiSubagentTask`、
`TestSubagentTaskReadOnlyWhenTheFileChanges`、
`TestSubagentWithMetaTaskIsNotReadForIt`、`TestStatsReadNoSubagentFile` 和
`TestSessionSubagentTranscript`。

Sessions 页面在顶层只画主 session。一个 session 的 subagent 折叠在它那一行下面
（"2 subagents"）；展开后，每一个是它任务和 kind 的一行，缩进在启动它的 subagent
下面，点击显示它的对话（`sub=`）。一个 `parent_session` 链到达某个已列出 session
的 Codex thread 会被画出来，作为一个有着自己详情、Resume 和 Delete 的 session，
在同样的折叠里放在它的父级下面；父级成环、链深于八层或父级未被列出的，会让它
留在顶层。那个目录的计数和复选框覆盖着其他 session 下面的那些。见
`sessions-subagents.test.cjs`。

Codex 保持现有的 session 标题优先级。只有当那让标题为空时，一个子 thread 才可以用
它父级的 `spawn_agent` `task_name`，按同一个 `call_id` 和一个 `item_completed`
事件匹配，那个事件的 item 类型是 `SubAgentActivity`、kind 是 `started`、有同一个
`id`，以及子级的 `agent_thread_id`。一个旧式 rollout（Codex 分页之前的默认历史）
把那个 started 活动保留成一个 `event_msg` 的 `sub_agent_activity`，它的
`event_id` 就是那个 call id，链接方式相同。一个链接只从子级自己父级的文件
（`parent_session`）算起：一个 fork 出来的子级文件会重复它父级的历史，包括它
兄弟的 spawn，而那份副本既不会点名一个子级，也不会从一个子级那里拿走名字。
原生的顶层 `session_meta.agent_nickname` 作为 `nickname` 单独暴露在 list、one、
manage 和 overview 的响应里；它不替代标题。解析过的 session 缓存针对 nickname 和
任务标题元数据做了版本化，Codex 调用缓存针对压缩也做了版本化，所以更旧的解析会
被重建。

`GET /api/sessions`、`/api/sessions/one` 和 `/api/sessions/manage` 还为每个 Codex
thread 返回 `model_effort`，按标准 Usage 账本的 `model` 和实际推理 effort 分组。
它带着那个 thread 的输入、输出、缓存和缓存写入 token、推理 token、费用，以及
分组的每次调用是否都定了价。Usage 账本的标准行提供这些值：一个与原生
`response_id` 配对的网关响应贡献一次，用网关行的 `model`、effort、token 和价格，
而相同的本地副本折叠为一次。
`/api/sessions/overview` 用概览里匹配的本地日历范围，把同样的字段加到它的
`tokens`、`cost` 和 `active` 顶部汇总里；`days=0` 使用全部历史。现有的 `models`
字段和 session 统计保持它们原先的聚合方式和含义。
这些是现有 API 响应上的字段；没有新增端点或统计结构。

在 Usage 请求 API 里，`session_role` 把标准的 Codex 行为者分类为 `main`、
`subagent` 或 `unknown`。对于一个已确认的身份，`session` 是它的 root thread ID；
当行为者已知但它的 root 未知时，`session` 为空。一个没有配对的网关行保留它
原始的 `session` 值，并带有 `session_role` `unknown`。现有的 `subagent` 和
`parent_agent` 字段被 Codex 复用：它们带着子 thread ID，以及在父级已知且本身是
一个 subagent 时的父 thread ID。`native_session` 保持是记录下来的原生头。新的
`session_role` 对其他 agent 省略，它们的现有字段语义保持不变。含糊或没有配对的
调用不会从共享的 session 元数据继承任何关系。

`GET /api/usage/requests` 和 `GET /v1/magpie/usage/requests` 在一个原生 Codex
调用确立了标记状态时，把 Codex 的 `compaction` 暴露为 `true` 或 `false`，在没有
匹配的原生证据时省略它。只有一个从 `token_usage_record` 读出的调用（它有
`response_id`）能确立它：Codex 把那条记录和一次压缩的 `compaction_response_id`
一起写出，所以一个在更老的 rollout 里仅凭 `token_count` 得知的调用会把它略去，
而不是说 `false`。一个唯一的 `response_id` 匹配把这个元数据加到网关行上，同时
保留该行的模型、用量、费用和 kind。
压缩已经计入 Codex 的总消耗，所以这个标记不增加 token 也不增加费用。这是
Codex 专属的；Claude 的行为和 POST `/v1/magpie/usage/requests` 的上报白名单不变。
