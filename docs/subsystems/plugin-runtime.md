# plugin 运行时

星鸣的内置 plugin 是并非人人需要的那些部分，每一个都是 `plugins/` 下一个用户可以关掉的包。`internal/pluginrt` 里的运行时保存它们的注册表，按各自服务要求的顺序启用开着的那些，并在某个 plugin 被关闭时释放它注册的东西。`plugins.yml` 说明哪些是开着的。本页描述 P0（运行时、这个文件和窗口里的开关）和 P1（窗口的内核：plugin 的页面、它的错误边界、它用来绘制的 kit 和它的设置表单，其中网关页是第一个自己画自己页面的 plugin），以及目前的 P2（整个从核心里搬出来的页面：它的代码、样式和文案：会话、资源库、路由和用量）、P3（网关的扩展点：观察者，以 OTel 导出为它的第一个 plugin；路由策略，以 Smart 之外的路由和用户的 group 为主；以及 middleware 点，以 JS 网关 middleware 和隐私脱敏为主；工具点，以联网搜索为主；以及预算点，以 key 限额为主）、P4（在每一个这样的点上都有社区 plugin）、P5（星鸣自己页面里的社区面板，以及占据内置实现位置的 plugin 页面）和 P6（provider 行菜单和 Agent 行菜单里的 plugin 项）。OpenCode plugin（plugin host 里的 npm 包）是另一回事；见[provider 与 plugin 归属](provider-plugins.md)和[网关 middleware](gateway-middleware.md)。

## 职责与边界

| 部分 | 作用 | 来源 |
| --- | --- | --- |
| 注册表 | plugin 在 `init` 里调用 `pluginrt.Register(Manifest)`。`Registered` 按 id 列出它们；两个用同一个 id 会 panic。`plugins/all` import 每一个 plugin，而只有 `main.go` import `plugins/all`，所以核心代码及其测试都不带任何 plugin 运行。 | [`pluginrt.go`](../../internal/pluginrt/pluginrt.go)、[`plugins/all/all.go`](../../plugins/all/all.go) |
| 清单 | `ID`、`API`（宿主 API 的插入符号范围，`"^1"`；`pluginrt.API` 是 `"1.0"`）、`Requires` 和 `Provides`（服务）、`Essential`、`Name`/`Sub`（英文；窗口负责翻译）、`One`/`Many`（在用的数量，1 或 `{n}`）、`InUse`、`Apply`、`Parts`（`{Slot, Key}`：它占窗口的哪些位置）、`Defaults`（它的配置）、`Settings`（它的设置结构体，在 `Register` 时做成 `Defaults` 和一份 schema）和 `Client`（它在窗口里的文件，放在 `internal/gui/assets/plugins/<id>/` 下：脚本、它的 `i18n.js` 和一个 `.css`）。 | [`pluginrt.go`](../../internal/pluginrt/pluginrt.go) |
| 生命周期 | `Runtime.Reconcile(on, config)`。先关：每一个不再被需要的 plugin，以及每一个依赖了这些 plugin 曾经提供的服务的 plugin，都按最新的优先释放。然后开：每一个需求已被满足（由 `Core`、`settings`、`providers`、`gateway`、`window` 或某个活跃的 plugin 提供）的所需 plugin 被应用，一遍又一遍直到没有更多能被应用。状态有：`off`、`pending`（等待某个服务，会记录日志）、`active`、`failed`（Apply 返回了错误或 panic，或者 API 范围不匹配；它此前注册的东西会被撤销）、`disposed`。失败的 plugin 不会影响其他 plugin。 | [`pluginrt.go`](../../internal/pluginrt/pluginrt.go) |
| 上下文 | `Apply` 拿到的东西：`OnDispose(f)`、`Contribute(slot, key)`、`Provide(name, v)`（它消失时一并撤回）、`Service(name)`，以及组合出来的 `Config`。释放器按相反顺序运行，每个的 panic 被记录，其余的照常运行。 | [`pluginrt.go`](../../internal/pluginrt/pluginrt.go) |
| plugins.yml | `appdir.Config()/plugins.yml`：`preset`（`essentials` 或 `all`；除非用户选了 Essentials，否则是 `all`）、`settled`（这个文件上次被更新到的默认值）和 `plugins`，即 `{id, enabled, config}` 的行，只保存用户设置过的内容。 | [`config.go`](../../internal/pluginrt/config.go) |
| 设置 schema | `SchemaOf(Settings)` 按字段顺序（`x-order`）把结构体做成一个 JSON Schema 对象：每个带 json 名的导出字段都是一项设置，类型为 `bool`、整数、`float64` 或 `string`，带 `title`（英文；窗口负责翻译）、`description`、`enum`（逗号分隔）和 `min`/`max` 标签。它的值就是默认值。表单画不出来的字段，或解析不了的标签，都会在 `Register` 时 panic。`Schema.Validate` 只接受 schema 允许的值（整数用 JSON float64 或 YAML int）；`SetConfig(id, config)` 通过它写入用户的那一行，而 `Compose` 会略掉用户手工改进 plugins.yml、但 schema 不允许的值，于是默认值仍然生效。 | [`schema.go`](../../internal/pluginrt/schema.go)、[`config.go`](../../internal/pluginrt/config.go) |
| 内置 plugin | `groups`、`library`、`otel`、`sessions`、`warmups`（P0），`gateway`（P1），网关页，它是 Essential 的，只要那里有网关 key 就在用（`access.List`），以及 `routing`（P2），路由页，Essential 的，只要用户有路由 group 就在用，因为 group 就是在那个页面里创建和修改的。`rotation`（P3），Smart 之外的路由，Essential 的，只要某个 provider 或 group 由其中之一路由就在用。`middleware`（P3），JS 网关 middleware，Essential 的，只要 plugins.json 里有一个 middleware 开着就在用（`middleware.On`）。`redact`（P3），隐私脱敏，Essential 的，只要设置 → 隐私要求任何脱敏就在用（`masking.Wanted`）。`search`（P3），星鸣的联网搜索，Essential 的，只要配好了搜索 API 或选定了搜索者就在用。`budget`（P3），key 限额，Essential 的，只要有网关 key 带限额就在用。`usage`（P2），用量页，同样 Essential，但没有 `InUse`：它显示的东西网关本来就在统计，而且托盘面板的用量标签页在它关闭时仍显示额度。网关页唯一的设置 `callsPage`（10、20、50 或 100；20），决定页面一次显示多少条最近调用。其余五个不是 Essential。每个的 `InUse` 统计用户用它做了什么：不是 Auto 也没有隐藏的路由 group、分给 Agent 的资源库条目、发过的采集器、记录下来的网关对话、开启的预热和签到。 | [`plugins/`](../../plugins) |

这条边界由 `TestPluginsImportOnlyCore` 检查：`plugins/` 之外的包不 import 它（`main.go` 除外，它通过 `plugins/all`），plugin 之间从不互相 import，而 `plugins/all` import 每一个 plugin 目录。plugin 之间只能通过服务互相到达。

## 窗口的内核（P1）

`internal/gui/assets/kernel.js` 在 `i18n.js` 之后、plugin 的脚本和 `app.js` 之前加载，都是共享同一作用域的 classic script。它就是 `window.starling`：

- `starling.plugin(id, setup)`：plugin 的脚本注册自己，`setup` 拿到一个上下文：`config`（来自 boot.js 的 `pluginConfig`，然后是插件 › 内置）、`onConfig(fn)`、`onDispose(fn)`、`slots.add(slot, part)` 和 `ui`。
- **slot。** 唯一的 slot 是 `page`：一个 part `{key, html, loads, mount, …hooks}` 会变成 `<main class="view" id="view-<key>" data-feature="<plugin>">`，也就是 index.html 里 `<template data-slot="page">` 所在的位置，于是 app.js 找页面时和找它自己的页面一样。app.js 在准备好时调用 `starling.start()`，它挂载每一个页面，然后通过 `starling.call(key, hook, …)` 调用各个 hook：对于 `loads: "providers"` 的页面，`loading`、`draft` 和 `render` 围绕着 `/api/providers`；`show` 打开页面时触发 `open`（它回给的一个 promise 在落定后把读者放回原处）；语言在页面上变化时触发 `relang`。`title` 是页面的标题（英文；`show` 负责翻译）。`starling.pages()` 把 plugin 的页面加进 `show` 和 `?view=` 认识的那些视图里。
- **文案。** plugin 的 `i18n.js` 调用 `starling.strings({zh, "zh-TW", ja, de})`，在 app.js 翻译任何东西之前加进 i18n.js 的 `I18N`。已经存在的 key 会被记录日志并且不采用。`tests/strings.cjs` 像页面那样读取它们，用于 `gui-de`、`gui-ja`、`gui-zh-tw` 和 `i18n-keys`（没有 key 出现在两个文件里，plugin 文件里的每个 key 在四种语言里都有）。
- **错误边界。** 抛出异常的 hook，或 promise 失败的 hook，会在 `run` 里被捕获：记录为 `[starling] <plugin>: <key>.<hook>`，页面被换成一张卡片，写着「这个页面停止工作了」、错误信息和一个「重试」。坏掉的页面的后续 hook 什么也不做，所以不会回来一个画了一半的东西，其他页面照常。重试会把页面的标记放回去、翻译它、挂载它，并用它当时拿到的东西重跑那个出错的 hook。`setup` 抛异常的 plugin 会被记录日志，并且没有页面。
- **kit**（`starling.ui`）：`el`、`button`、`toggle`（插件用的开关，`role="switch"`）、`menu`（通过 `openProtoMenu` 实现的 app 菜单，绝不用原生的 `<select>`）、`row`（一行设置）和 `form(schema, values, onChange)`：布尔值用开关，枚举用菜单（数字按窗口的语言习惯显示），数字用输入框（限制在它的范围内，不在范围内时标 `aria-invalid`），其余用文本。
- **设置。** 插件 › 内置在每个开着的 plugin 那一行下面用 `ui.form` 画出行（`.feat-set`），并配一份 schema；改动会发送 `{id, config}`；`applyFeatures` 把每个 plugin 的 config 交给 `starling.configure`，它只在值变了时才调用其 `onConfig`。

网关页是 `plugins/gateway`，其 `Client: keys.js, client.js`（`internal/gui/assets/plugins/gateway/`）：`keys.js` 是原来的 `caller-keys.js`，即网关 key 管理器；`client.js` 是原来在 app.js 里的页面代码，以 `starling.plugin("gateway", …)` 收尾。仍在 app.js 和 app.css 里的有：它的样式、它的侧边栏和页头按钮（index.html 里的 `data-feature="gateway"`），以及模型信息卡共用的 `decideEntry`。`TestBuiltinClientsLoaded` 要求 index.html 在 kernel.js 和 app.js 之间加载每个 plugin 的 `Client` 脚本，并把它的样式表放在 head 里，各自带上自己的 `?v=`。

## 搬出来的页面（P2）

每个页面整体地在自己的一个提交里搬走：代码搬进 `assets/plugins/<id>/client.js`，样式搬在它旁边，而只由它说的文案从 i18n.js 搬进它自己的 `i18n.js`（app.js、其他页面、index.html 或 Go 也说的 key 仍留在 i18n.js）。搬走之后一直如此：`i18n-keys` 的「plugin 的文案归它自己」在 plugin 目录之外有人说它的某句文案时就会失败，那句要搬回 i18n.js。

- **会话**（`plugins/sessions`，`Client: i18n.js, sessions.css, client.js`）：原来的 `sessions.js` 那个页面，它的标记全部由自己的代码画，所以它的 part 没有 `html`。它的 hook：`open` 和 `relang` 加载它，`reveal(s, e)` 是用量页进入它的方式（某个 session 那一行和「全部会话」，#752），通过 `starling.call("sessions", "reveal", …)` 调用，所以没有会话 plugin 时它什么也不做。关掉会话 plugin 时，用量列表的某一行改为就地展开它的详情（`sessionItem`，#2311）。仍在 app.js 里的：用量页的会话标签页（`/api/sessions`）和设置里的 session 终端那一行（`data-feature="sessions"`）。
- **资源库**（`plugins/library`，`Client: i18n.js, library.css, client.js`）：原来的 `library.js` 那个页面。它的代码只在 `mount` 时运行，在那里接上它的监听器（对话框的背景和 Escape、焦点和可见性变化时重新读取、它那些列表的滚动），因为脚本在 app.js 之前就在了。它的 hook：`open` 和 `relang` 读取 `/api/library`，`redraw` 在状态说明哪些 Agent 被隐藏后重新画一遍，而 `close`、`dirty` 和 `discard` 是读者离开前 `show()` 要问的。`window.loadLibrary`、`redrawLibrary`、`closeLibraryModal`、`libraryDirty` 和 `discardLibrary` 都没了。
- **路由**（`plugins/routing`，`Client: i18n.js, routing.css, client.js`）：原来的 `routing.js` 那个页面，Essential 的。它那 part 的 `html` 是原来 index.html 里的页面头部和方框；页面其余部分由 `routingPage` 画出，它在 `mount` 时运行，因为它建立在 app.js 之上。之后它会像以前一样维持自身（它自己的 `hidden`、语言、窗口焦点），所以它的 hook 只有 `dirty` 和 `discard`，用于正在编辑的路由 group，`show()` 和窗口关闭通过 `starling.call("routing", …)` 询问它们：`window.routingDirty` 和 `discardRouting` 都没了。它在托盘面板里的标签页和面板的那一节都标了 `data-feature="routing"`。其他页面用到的部分仍在 `window` 上（context.js 的 `openRoute`；用量请求列表的 `kindName`、`swapWhy` 和 `routedWhy`；供应商页的 `newGroupWith`；面板的 `panelRoutingShown`）。plugin 关闭时，通向该页面的入口也跟着没了：导航和面板的标签页、指南里的「打开路由」、供应商页的「建立路由 group…」（`data-feature="routing"`）、用量列表的「查看路由」以及把时间做成链接，还有上下文卡的「最新请求」；停在路由标签页上的面板会显示 Agent。
- **用量**（`plugins/usage`，`Client: i18n.js, client.js, page.js`）：那个曾是 app.js 最大一部分的页面（约 3000 行：概览的卡片、请求账本、会话标签页和 Agent 列表）。它那 part 的 `html` 是原来 index.html 里 `#view-usage` 的页面标记。它的代码在顶层建立在 app.js 之上（它的计时器、`?view=usage` 的标签页和筛选器），所以 `page.js` 是 plugin 在 app.js 之后加载的那一个脚本：index.html 加载 kernel.js、plugin 的其他脚本、app.js 和它旁边的脚本，然后是各自的 `page.js`，最后是 `start.js`，也就是 app.js 原来的尾巴（`starling.start()`、第一个显示的页面、第一次 `load()`），这样第一个页面就有自己的代码。`TestBuiltinClientsLoaded` 要求这个顺序。它的 hook：`open` 加载它，`refresh(again)` 是 `load()` 对它的读取（回到概览就是打开额度，#1588），`reload` 是「立即刷新」，`costs` 在设置的货币变化时重画费用，`tab` 说明它在哪个标签页上，供 `focusQuotaCard` 使用，而 `leave` 是 `show()` 在离开某页时调用的，它把点击借给请求列表的东西还回去（`ledLeave`，Discord 上的 TJHHHH）：排名或图表点击设下的筛选器，或者由地址栏或路由页的「查看用量」带来的筛选器，在读者离开请求列表去别的页面或标签页时就失效。在自己菜单里选的或手输的则保留。仍在 app.js 里的，因为托盘面板也要画它们：额度卡片、`quotasOpened`、`panelQuotaFocus` 和 `panelUseFocus`。plugin 关闭时它的导航标签页没了，通向该页面的方式也没了（`data-feature="usage"`）：托盘面板的「打开用量」和路由页的「查看用量」。任何地方残留的 `show("usage")` 都让位给 Agent。面板的额度标签页保留。

## 网关扩展点（P3）：观察者

P3 在五个点上向 plugin 开放网关：观察者、路由策略、middleware、工具和预算。观察者是第一个。它只读，所以最先迁移它也最安全。`plugins/otel` 是它的第一个使用者，关掉该 plugin 时网关不做任何观察工作。

| 部分 | 作用 | 来源 |
| --- | --- | --- |
| `usage.Observer` | `Serve() (stop)`、`Bodies()`、`Begin(Request) Watch` 和 `Ledger(Record)`。观察者观察网关服务了什么、账本写了什么，两者都不改。`Observe(o)` 注册一个观察者直到它的 remove；`Observers` 列出它们。它住在 `internal/usage` 而不是网关里，因为每一行都是一条用量 `Record`，账本的其他写入者也会到达它。 | [`observe.go`](../../internal/usage/observe.go) |
| 提供服务 | `ServeObservers()` 启动每一个已注册的观察者，以及之后注册的每一个，直到它最后一个停下（它会被计数）。只有提供网关服务的进程调用它（`Server.ListenAndServe`），所以写一条用量记录的 CLI 不导出任何东西。Remove 会停下一个正在服务的观察者。 | [`observe.go`](../../internal/usage/observe.go)、[`gateway.go`](../../internal/gateway/gateway.go) |
| `Begin` 与 `Watch` | 请求开始时，网关拿 `usage.Request` 问每个观察者（`observeRequest`）：它的请求头、Agent、session（Agent 点名时用原生的那个）、调用种类（Claude 无工具的辅助调用是 `auxiliary`）和 `Plain`（回环、没有网关 key、不是从另一个 starling 传来的）。它返回的 `Watch` 会为每次路由尝试、未计费的失败和重试各得到一个 `Attempt`，最后得到一次 `End`。没有观察者、或没有观察者在看时，请求什么都不带。 | [`observe.go`](../../internal/gateway/observe.go) |
| 请求体 | `Bodies()` 是 `NoBodies`、`CapturedBodies`（前 256 KiB，用 `usage.BodyCut` 截断）或 `WholeBodies`（回复在流式输出中完整缓冲，#538）。网关取任何观察者当下想要的最多者（`ObservedBodies`）：只有有人要完整回复时才保留整个回复，只有有人要任何请求体时才把请求体放到某次尝试的记录上。 | [`observe.go`](../../internal/gateway/observe.go)、[`capture.go`](../../internal/gateway/capture.go) |
| `Ledger` 与 `Watched` | `Append` 写的每一行都会传给每个观察者。某个观察者看过的那次请求，它的行会设上 `Watched`（`appendUsage`），于是发了该请求的那些尝试的观察者不会再发它那一行。 | [`usage.go`](../../internal/usage/usage.go)、[`lan.go`](../../internal/gateway/lan.go) |
| `internal/otlp` | OTLP/HTTP 导出器，整个从 `internal/usage` 搬出来：`otlp.New()` 是一个 `Exporter`，也是一个 `usage.Observer`。它的 Begin 在 session 追踪读取时，把本地 Agent 自己的回合留给 Agent 自己的记录，并取一个 W3C `traceparent` 作为父级。设置照旧每次调用都读。 | [`otlp/`](../../internal/otlp) |
| `plugins/otel` | 需要 `window`、`settings` 和 `gateway`。Apply 注册 `otlp.New()`，关闭时释放它。`InUse` 在 `settings.OTelExport` 说要发送时为 1（settings.json 或 `STARLING_OTEL_*`），所以只通过环境变量导出的用户在 Essentials 下也保持它开着。 | [`otel.go`](../../plugins/otel/otel.go) |

## 网关扩展点（P3）：路由策略

<a id="gateway-extension-points-p3-router-strategy"></a>

路由策略这个点规定一个路由池的尝试顺序：某个 provider 的账户或 key，或某个 group 成员的池头。网关仍然列出这个池，并保留休息、亲和、规则、钉住以及对顺序本身的遍历；策略只负责排序。默认的 Smart 留在核心里。其余六种路由属于内置 plugin `rotation`。

| 部分 | 作用 | 来源 |
| --- | --- | --- |
| `routes.Strategy` | `Order(pool, seats) []int`：座位索引的一个排列。`pool` 是某个 provider 的 id 或 `group/<id>`，用于每个池各轮一次的策略。`Exact` 保持给定的顺序；否则更贴合请求的 provider API key 排在前面，各自按自己的顺序。`ByUse` 把每个座位近期的 token 数放进追踪里。 | [`routes.go`](../../internal/routes/routes.go) |
| `routes.Seat` | 策略称量的东西，在路由状态的锁下由 `weighRouted` 读取：它的休息 key、它是否是账户、它的额度是否已知或在作答时顺便了解、已用的份额、它的窗口何时刷新、它的周节奏、它 key 的权重、它现在是否在休息以及它最近几小时的 token 数。 | [`routes.go`](../../internal/routes/routes.go)、[`routing.go`](../../internal/gateway/routing.go) |
| `Register`、`Find`、`Apply` | `Register(routing, s)` 为一个路由提供服务直到它的 remove；重复注册会叠加，最后一个生效。`Find("")` 永远是 `Smart`。`Apply` 失败放行：panic 的策略，或者给出不是排列的东西，会让池保持它列出的顺序，并记录日志。每条路由仍会被尝试。 | [`routes.go`](../../internal/routes/routes.go) |
| 策略 | `Smart`（核心），以及 `routes.Builtin` 里的 `Ordered`、`Rotate`、`LeastUsed`、`Pace` 和 `Weighted`。`Rotate` 为每个池保留一轮，`Weighted` 保留一个平滑的加权轮（`Round`、`MoveRound`），两者都从网关搬出；`Forget` 清掉它们。 | [`strategies.go`](../../internal/routes/strategies.go)、[`round.go`](../../internal/routes/round.go) |
| 关闭 | 没有策略服务的路由按它列出的顺序走。网关里的 `routingOf` 在核心询问某个路由的任何地方（某一轮的标记、`sinks`、轮换 group 的 session、group 的池头）都按 `order` 读它，所以路由不会有一半留在开启状态。 | [`routing.go`](../../internal/gateway/routing.go) |
| `plugins/rotation` | 需要 `providers` 和 `gateway`，`Essential` 的。Apply 注册每一个 `routes.Builtin`，关闭时移除它们。`InUse` 统计 providers.json 里由其中之一路由的 provider 和 group（`provider.Routings`，只看文件，不看目录）；providers.json 读不了就是错误，所以它保持开着。网关的 `KeepFresh` 也盯着 providers.json，所以从 CLI 或 TUI 设的路由 2 秒内就会让它打开。关闭时，窗口里的路由选择器（provider 编辑器、路由页的池和 group）只提供 Smart：它们的选项标着 `data-feature="rotation"`（`routingSegs`）。 | [`rotation.go`](../../plugins/rotation/rotation.go)、[`provider.go`](../../internal/provider/provider.go)、[`app.js`](../../internal/gui/assets/app.js) |

用户自己的路由 group 属于这个点上的 `groups` plugin。星鸣自己找到的 group（两个 provider 都提供的模型，`auto/<model>`）仍归核心，和 Smart 一样：它们让多个 provider 都提供的模型成为一个带故障转移的模型，用 Essentials 预设的用户就有这些，而那个预设里 `groups` 是关的。`planGroup` 以及对一个 group 成员的遍历无论何时都归核心：策略只给池排序，而一个关掉的 group 根本不会被提供。

| 部分 | 作用 | 来源 |
| --- | --- | --- |
| `GateGroups`、`ServeGroups` | 网关在提供服务期间持有 `GateGroups`（`ListenAndServe`），plugin 在开着期间持有 `ServeGroups`。闸门关着而没有服务时，`groupsIn` 会略去每一个不是 `Auto` 的 group，于是 `Groups`、`FindGroup`、`GroupFor`、目录、`/v1/models` 和某个 group 的 `Unlists` 都没有它们：请求这样的 group 就是请求一个没有提供的模型。不提供网关服务的进程（CLI、TUI）从不被闸门挡住，会列出全部。没有任何基于 group 建的东西在一次调用之后被保留，所以一次开关下一次查找就能看到。闸门关着时，某次查找若先找到用户自己的 group，会先问闸门的刷新（网关的是 `pluginrt.Refresh`），最多每秒一次且绝不从一次查找内部发起，然后重新查找：刚刚建好的 group 会让 plugin 打开，并由请求它的那次请求提供服务，而不是等 `KeepFresh` 下一次查看之后（`TestOwnGroupRefreshesTheGate`）。 | [`group_serve.go`](../../internal/provider/group_serve.go)、[`group.go`](../../internal/provider/group.go)、[`gateway.go`](../../internal/gateway/gateway.go) |
| `plugins/groups` | 需要 `window` 和 `providers`，不是 Essential 的。Apply 持有 `ServeGroups`。`InUse` 是 `provider.NamedGroups`：providers.json 里既不是 Auto 也没隐藏的 group，直接从文件读，从不经过闸门（否则关着时会一个都数不到而保持关闭）；文件读不了就是错误，所以它保持开着。`routing` plugin 的 `InUse` 是同一个计数。网关的 `KeepFresh` 盯着 providers.json，所以用 `starling group add` 建的 group 2 秒内就会让它打开，而在此之前对它的请求会自己让它打开（即上面的闸门刷新）；窗口的「新建 group」按钮是 `data-feature="groups"`。 | [`groups.go`](../../plugins/groups/groups.go)、[`routing.go`](../../plugins/routing/routing.go) |

## 网关扩展点（P3）：middleware

middleware 这个点负责在请求穿过网关时改动它、在回复返回时改动回复。它在 [`internal/gwmw`](../../internal/gwmw/gwmw.go) 里有两处，各自有自己的失败策略：

| 部分 | 作用 | 来源 |
| --- | --- | --- |
| `gwmw.Agent` | 某个 Agent 发出的请求上的 middleware，在 Agent 自己的 API 上、在网关路由它之前，以及在它拿到的回复上：`Begin(Info) Run`，以及这个 run 的 `Request`（一个 `Reject` 会以 Agent 的错误形态结束请求）、`Wrap` 和 `End`。`Use(a)` 注册一个直到它的 remove；`Agents` 按顺序列出它们。`serveAgent` 依次运行每一个；没有时它直接调用 `serve`。它失败放行，和 JS middleware 一直以来的做法一样。 | [`gwmw.go`](../../internal/gwmw/gwmw.go)、[`middleware.go`](../../internal/gateway/middleware.go) |
| `plugins/middleware` | 需要 `gateway`，Essential 的。Apply 注册 `middleware.Chain`，即 JS 网关 middleware（[网关 middleware](gateway-middleware.md)），它的位置、API、时限和失败放行策略都不变。`InUse` 是 `middleware.On()`，即 plugins.json 里开着的 middleware 数量。关闭时，Agent 的请求照原样发出，不调用任何 middleware，而 plugins.json 保持原样。网关的 `KeepFresh` 也盯着 plugins.json，所以从 CLI 或窗口添加的 middleware 2 秒内就会让它打开，而在此之前的请求会自己去问闸门的刷新（`gwmw.AgentsWanted`：没有注册任何 agent middleware 而 `middleware.On()` 大于 0 时，最多每秒一次），于是 middleware 在那个请求上运行，和它在成为 plugin 之前一秒内运行一样。 | [`middleware.go`](../../plugins/middleware/middleware.go)、[`run.go`](../../internal/middleware/run.go) |
| `gwmw.Masker` | 遮住星鸣发给厂商的东西，并把那些值放回收到的东西里：`JSON`、`Text`、`Restore` 和 `Writer`。`Mask(m)` 为一个 masker 提供服务直到它的 remove（最后注册的生效）。它失败关闭：`Masking(wanted)` 是某个请求的 masker，当设置要求脱敏却没有 masker 服务时是 `ErrUnmasked`，而 `Safely` 造出一个会把 panic 变成错误的 masker。网关的调用点（`serve`、count_tokens、embeddings 和 audio、transcriptions、images、video、搜索替身及其查询、Codex 后端）要么回答 503「……什么都没发送」，厂商什么都收不到。`Served` 在没有 masker 时会问闸门的刷新（网关的是 `pluginrt.Refresh`，由 `ListenAndServe` 中的 `gwmw.Refresh` 设置），最多每秒一次且绝不从一次查找内部发起，然后重新查找，于是刚刚打开的脱敏会为需要它的那个请求把 plugin 打开。 | [`gwmw.go`](../../internal/gwmw/gwmw.go)、[`redact.go`](../../internal/gateway/redact.go)、[`gateway.go`](../../internal/gateway/gateway.go) |
| `gwmw.Teller` | masker 除了遮蔽之外还能说的东西：`JSONHits` 和 `TextHits` 像 `JSON` 和 `Text` 那样遮蔽，并额外返回每个新建或复用的占位符（`Hit`：占位符、种类、查找方式、开头、长度），而 `Value` 是某个占位符所代表的内容，只要这个 masker 还记得它。可选：只有当提供服务的 masker 具备这一能力时网关才记录一个请求的脱敏（`masking.Masker` 具备），设置 → 隐私里的记录和值需要它；没有它的 masker 遮蔽是一样的，但不留任何记录。 | [`gwmw.go`](../../internal/gwmw/gwmw.go)、[`masked.go`](../../internal/gateway/masked.go)、[`masking.go`](../../internal/masking/masking.go) |
| `plugins/redact` | 需要 `settings` 和 `gateway`，Essential 的。Apply 提供 `masking.Masker`，它每次调用都读 settings.json，并像以前那样用 `internal/redact` 遮蔽（它住在 `internal/masking`，因为 `settings` import 了 `redact`）。`InUse` 是 `masking.Wanted`：*脱敏密钥*、*脱敏个人信息*或用户自己的某个词；settings.json 读不了就是错误，所以它保持开着。既然只要要求脱敏它就在用，那么只有在没有任何脱敏要求时它才可能是关的；那个 503 是给还没看到的刷新或手工改过的 plugins.yml 的。 | [`redact.go`](../../plugins/redact/redact.go)、[`masking.go`](../../internal/masking/masking.go) |

## 网关扩展点（P3）：工具

工具这个点是星鸣替某个 Agent 自己回答的工具，代替厂商。现在是联网搜索：给不能搜索的模型用的替身 `web_search`、搜索 MCP server，以及 ChatGPT 答不了时的 Codex 搜索（[网关路由](gateway-routing.md)）。自己能搜索的厂商从来不属于这个点：它自己的 `web_search` 就发给它，无论开关。

| 部分 | 作用 | 来源 |
| --- | --- | --- |
| `gwtool` | `Serve(name)` 为一个工具提供服务直到它的 remove；`Gate()` 让每个工具都等 `Serve` 直到它停下；`Served(name)` 表示它没有被闸门挡住，或已被服务。网关在提供服务期间持有 `Gate`（`ListenAndServe`，`TestServingGatesTools`），所以不提供网关服务的进程（CLI、TUI、写在那些地方的 Agent 配置）拥有全部工具，和这个点出现之前一样。`WebSearch` 是搜索的名字。 | [`gwtool.go`](../../internal/gwtool/gwtool.go)、[`gateway.go`](../../internal/gateway/gateway.go) |
| 关闭 | `searchServed` 闸住每条搜索路径都要经过的三处：`canSearch`（替身工具、`/v1/models` 里给另一个 starling 的搜索说明、各订阅的搜索、Agent 那些行用的 `CanSearch`）、`searchesOwn`（Kimi Code 或 WorkBuddy 计划自己的搜索服务、Google 登录后的 Gemini）以及 `webSearch` 本身，后者不找搜索者也不找搜索 API 就回答 `errSearchOff`。所以关闭时，带 `web_search` 的请求照没有配好搜索者时那样发给厂商，而 MCP server 和 Codex 的搜索会说该 plugin 已关闭（`TestSearchOnAndOff`）。 | [`search.go`](../../internal/gateway/search.go)、[`search_api.go`](../../internal/gateway/search_api.go)、[`search_kimi.go`](../../internal/gateway/search_kimi.go) |
| `plugins/search` | 需要 `settings` 和 `gateway`，Essential 的。Apply 提供 `gwtool.WebSearch`。`InUse` 统计配好的搜索 API（`provider.SearchAPIsSet`，只看文件）和在设置里选定的搜索者（既不是星鸣自己的选择，也不是关掉）；providers.json 或 settings.json 读不了就是错误，所以它保持开着。网关的 `KeepFresh` 盯着这两个文件，所以从 CLI 或窗口添加的搜索 API 2 秒内就会让它打开。 | [`search.go`](../../plugins/search/search.go) |

## 网关扩展点（P3）：预算

<a id="gateway-extension-points-p3-budget"></a>

预算这个点把一个网关 key 限制在它的所有者设定的额度内（#585、#2227），在该 key 的任何请求发给 provider 之前。它是 P3 几个点里的最后一个。没有 key 的请求，或属于一个没有限额的 key 的请求，都不会问它。

| 部分 | 作用 | 来源 |
| --- | --- | --- |
| `gwbudget.Limiter` | `Reserve(who, body, model, now) (release, refused)`：放行某个 key 的请求，把它预期会用到的部分预扣住直到 release，或者用一个 `Refusal`（`RetryAfter`、`ResetAt`）拒绝它。`Limit(l)` 为一个 limiter 提供服务直到它的 remove；重复注册会叠加，最后一个生效。 | [`gwbudget.go`](../../internal/gwbudget/gwbudget.go) |
| 失败关闭 | `Limiting(who)` 对没有 key 或 key 没有限额的请求是 nil，否则是提供服务的 limiter，而当 key 有限额却没有 limiter 服务时是 `ErrUnlimited`。没有时它先问刷新（网关的是 `pluginrt.Refresh`，由 `ListenAndServe` 中的 `gwbudget.Refresh` 设置），最多每秒一次且绝不从一次查找内部发起，然后重新查找，于是刚刚设下的限额会为需要它的那个请求把 plugin 打开。`reserveFor` 是网关唯一的入口，用于 chat 类接口（`keyLimited`）、搜索 MCP server 和视觉那个：拒绝是 429 `gateway_key_limit`，带 Retry-After 和 X-Starling-Limit-Reset；`ErrUnlimited` 是 503 `gateway_key_limit_off`、X-Should-Retry 为 false、什么都不发送，并且留下一条被拒的用量记录（`TestKeyLimitsOnAndOff`、`TestVisionMCPHeldToKeyLimit`）。 | [`gwbudget.go`](../../internal/gwbudget/gwbudget.go)、[`key_limit.go`](../../internal/gateway/key_limit.go) |
| `plugins/budget` | 需要 `settings` 和 `gateway`，Essential 的。Apply 提供 `budget.Limiter`，以及 `internal/budget` 的窗口和预扣，和以前一样。`InUse` 统计带限额的网关 key（`access.List`）；key 文件读不了就是错误，所以它保持开着。既然只要 key 有限额它就在用，那么只有在没有任何 key 有限额时它才可能是关的；那个 503 是给还没看到的刷新或手工改过的 plugins.yml 的。网关的 `KeepFresh` 也盯着 key 文件，所以从 CLI 或窗口设下的限额 2 秒内就会让它打开。 | [`budget.go`](../../plugins/budget/budget.go) |
| 留在核心里 | 一个 key 用了多少、还剩多少（`budget.All`、`budget.Binding`），供 `/v1/starling/limit`、网关页和 CLI 使用，以及 `budget.Append`，这样用量账本在 plugin 开着或关着时统计 key 的用量一样。key 编辑器的限额字段在 plugin 关闭时保留，因为设限额正是把它重新打开的动作。 | [`budget.go`](../../internal/budget/budget.go) |

## P4：每个点上的社区 plugin

<a id="p4-community-plugins-at-every-point"></a>

P4 把每一个扩展点（网关的和窗口的）向社区 plugin 开放（OpenCode plugin，plugin host 里的 npm 包）。网关的每个点仍藏在提供服务的内置 plugin 后面，所以关闭仍然是关闭；页面则归它自己的 plugin 开关。

| 点 | 社区 plugin 提供什么 | 由谁提供 | 参考 |
| --- | --- | --- | --- |
| 路由策略 | 模块的 `export const strategies = {name: (pool, seats, options) => order}`（或 `{order}` 对象）是路由 `<short name>/<name>`，在网关提供 `middleware.Points()`、plugin 开着且已加载时用 `routes.Register` 注册；关掉或重载时它们被移除。每个 20 毫秒，失败放行：出错、超时或返回的不是座位顺序，就用列出的顺序，并在插件页上计数。路由选择器提供被服务的那些（设置状态里的 `jsRoutings`）；保存过但 plugin 已不在的路由照原样显示，并按列出的顺序走。providers.json 会把这样的路由按保存的样子留下（`provider.PluginRouting`），所以关掉再打开的 plugin 仍会路由它的 group。 | `plugins/middleware`（“JS gateway plugins”，和它的 [middleware](gateway-middleware.md) 出自同一个模块） | [`points.go`](../../internal/middleware/points.go)、[`app.js`](../../internal/gui/assets/app.js)、[`client.js`](../../internal/gui/assets/plugins/routing/client.js)、[`groups_cli.go`](../../internal/cli/groups_cli.go)、[`routing.go`](../../internal/provider/routing.go) |
| 用量观察者 | `export function onUsage(row, options)` 拿到白名单里的每一条用量记录（没有请求体、提示词或 id），来自请求路径之外的有界队列：队列满了就丢弃并计数，请求永远不等待。每个 250 毫秒，失败放行。 | `plugins/middleware`（“JS gateway plugins”，和它的 [middleware](gateway-middleware.md) 出自同一个模块） | [`points.go`](../../internal/middleware/points.go) |
| 工具 | 它 OpenCode `tool` hook 里的工具，由星鸣的 MCP server 在 `/mcp/starling/tools` 提供给每个 Agent，每个命名为 `<plugin>_<tool>`，每次调用都受调用方 key 的限额约束，并在宿主里于 55 秒内运行。关闭时没有工具被列出，调用会说该 plugin 已关闭。 | `plugins/tools`（“Plugin tools”，Essential 的；`InUse` 统计开启并提供工具的 plugin，来自 `plugin-tools.json`）、`gwtool.PluginTools` | [Plugin tools](plugin-tools.md) |
| 窗口里的一个页面 | package.json 里的 `starling.page {title, file, icon}` 和 `starling.permissions`，像任何 plugin 一样从 npm、GitHub 或目录安装（`plugin.WindowPage`：`file` 是 plugin 内的 .html，图标在页面的目录里，读取限定在 `plugin.PageReads`）。一个开着的 plugin 的页面会在侧边栏的插件组里得到一个标签页，排在内置页面之后（`pluginpages.js`，一个内核 `page` part）。它在沙箱为 `allow-scripts` 的 iframe 里运行，它的文件从 `/_starling/plugin/<token>/` 提供（spec 的 HMAC，配上本次运行的密钥），策略是 `sandbox allow-scripts`，只允许它自己的目录，`connect-src 'none'`。框架通过 `postMessage` 与窗口对话：窗口只回答它自己的那个框架（`ev.source`），内容包括主题、它自己的选项（`/api/plugin-pages/options`，最多每秒一次，保存时不必为纯页面 plugin 重启宿主）以及它声明的读取（`/api/plugin-pages/read`，服务端会再检查一遍）：`usage:summary`（总数和前 20 个模型）和 `providers:list`（id 和名字）。没有秘密可读，因为页面仍然可以把它的框架发给别的站点。`noFramedAPI` 挡掉沙箱框架或别的站点能发的每一个 `/api/` 请求（`Origin: null`、`Sec-Fetch-Dest` 是框架、跨站子资源），因为桌面窗口的 API 没有 key；`webGuard` 放行该前缀但不放行框架从不发送的 cookie，于是 token 就是 key。抛出异常的页面，或者 10 秒内不说 hello 的页面，会经内核被判定为损坏（`ctx.broke`），并提供重试。插件页在 plugin 打开之前就在那一行显示这个页面和它的读取。只是一个页面的 plugin 仅在星鸣内部有效：不会为它启动 Bun host。发现页在页面分组下按 `kind: "page"` 列出市场条目并带上页面标签，而打了 `starling-plugin` 标记、且 package.json 里有 `starling.page.file` 的 GitHub 仓库则作为页面列出（`TaggedRepos`，它随后检查那个文件已提交，就像对 provider 条目那样；`TestTaggedRepos`、`plugin-kinds-discover.test.cjs`）。测试：`TestWindowPageStaysInside`、`TestPluginPageServedShutIn`、`TestPluginPageReadsWhatItDeclared`、`TestPluginPageNoAPI`、`TestPluginPageWebGuard`、`plugin-page.test.cjs`。示例：`internal/gui/testdata/page-plugin`。 | 窗口，不需内置 plugin：这个页面归该 plugin 自己的开关 | [`page.go`](../../internal/plugin/page.go)、[`pluginpage.go`](../../internal/gui/pluginpage.go)、[`pagebridge.js`](../../internal/gui/pagebridge.js)、[`pluginpages.js`](../../internal/gui/assets/pluginpages.js) |

现在 `plugins/middleware` 是「JS gateway plugins」。Apply 注册这条链和 `usage.Observe(middleware.Points())`；关闭时没有 JS middleware、策略或 `onUsage` 运行，由 JS 策略路由的池按它列出的顺序走。`InUse` 不变。（[`middleware.go`](../../plugins/middleware/middleware.go)）

## P5：星鸣自己页面里的面板，以及取代内置页面的页面

P5 让社区 plugin 进入星鸣自己画的页面，而不只是拥有自己的页面。两者都来自 package.json，由 `plugin.Window`（[`page.go`](../../internal/plugin/page.go)）读取；都不会启动 Bun host。

- **面板。** `starling.panels: [{slot, file, title}]`，一个 plugin 最多 8 个（`maxPanels`），每个 `slot` 是 `plugin.PanelSlots` 之一：`agents`（Agent 页的末尾）、`usage`（用量概览的末尾）、`settings`（设置里的一个插件标签页，只在它里面有面板时显示，在标签栏里，也在宽窗口的侧边栏里）和 `provider`（每个已保存 provider 的编辑器，会告知 `{provider: {id, name}}`）。`file` 是 plugin 目录里的一个 .html（`plugin.FileIn`）。`GET /api/plugin-panels` 按 plugins.json 的顺序列出一个开着的 plugin 的面板，每个面板有自己的 token（`pageToken(spec + "\x00panel\x00" + i)`），在和页面相同的策略下提供（[`pluginpage.go`](../../internal/gui/pluginpage.go)）。面板的框架和页面一样被沙箱为 `allow-scripts`，并通过同样的 `/api/plugin-pages/read` 和 `/options` 共享它那个 plugin 的读取（`starling.permissions`）和选项。桥（[`pagebridge.js`](../../internal/gui/pagebridge.js)）给它 `starling.context` 并上报它的高度（`size`），窗口把它限制在 0..2400px。抛出异常的面板，或者 10 秒内不说 hello 的面板，会被换成它自己的卡片（「{name} 停止工作了」、重试）；它周围的页面和旁边的面板照常。尺寸观察自己的「ResizeObserver loop」提示——WebKit 给沙箱页面的是一句裸的「Script error.」——不算抛出异常。
- **标题与外观。** `starling.page.titles` 和面板的 `titles`，一个语言代码到标题的对象（`plugin.titles`：去掉空白，空的和非对象的被丢弃），在窗口的语言里给出标签页和面板标题，`zh-TW` 回落到 `zh`，其他都回落到 `title`（[`pluginpages.js`](../../internal/gui/assets/pluginpages.js) 里的 `titled`，`<html lang>` 变化时重画）。桥会前置一份 `:where()` 规则、权重为零的样式表，针对字体、颜色、标题、按钮、输入框和表格，这样没有样式表的页面读起来和窗口一致，而页面自己的任何规则都算数。测试：`TestWindowTitles`、`plugin-page.test.cjs`。
- **策略里的 scheme。** 桌面窗口在 `wails://localhost` 上提供服务（macOS、Linux），而沙箱框架的 origin 是不透明的，所以一个不带 scheme 的 CSP source 在那里只匹配 http(s)。`ownWindow` 把原生 scheme 放到请求的 context 上，`pagePolicy` 指出它下面的目录（#2501）；web 版保持不带 scheme 的 source。`TestPluginPageWindowPolicy` 检查该策略，`TestPluginPageLoadsItsFilesInTheWindow` 在真实 WKWebView 里经 `ownWindow` 加载一个页面；Playwright 在 http 上看不到它。
- **slot。** 页面可以用标记里的 `<div class="plugin-slot" data-slot="…" hidden>` 指明一个 slot，也可以由 app.js 用 `starling.slot(name, context)` 画一个（[`pluginpages.js`](../../internal/gui/assets/pluginpages.js)）。面板到来时 slot 被填上，而且只在面板变了时才重填（一个开着的面板原样保留），内核之后写入某个页面的标记时也会填（`starling.fillSlots`，由 `add` 和重试触发）。
- **取代一个页面。** `starling.page.replaces`，取 `plugin.Replaceable` 之一（`sessions`、`library`、`routing`、`usage`），把 plugin 的页面放到内置页面在侧边栏里的位置：内置页面的按钮被隐藏（`data-replaced`），而对它的 `show()` 会打开 plugin 的页面（`starling.replacing`）。按 plugins.json 顺序第一个开着的 plugin 拿到一个页面；另一个要求同一页面的则拥有自己的页面。内置页面仍然在，plugin 关闭时会回来。
- **插件页**在那一行说明一个 plugin 的面板放在哪里、读些什么，以及它取代了哪个页面（[`plugins.js`](../../internal/gui/assets/plugins.js) 里的 `panelsLine`、`pageLine`），并带一个面板标签。

测试：`TestWindowPanels`、`TestPluginPanelsServedShutIn`、`TestPluginPanelsReadWhatTheyDeclared`、`TestPluginPageReplacesFirstWins`、`plugin-panels.test.cjs`。示例：`internal/gui/testdata/panel-plugin`。

## P6：星鸣菜单里的动作

P6 让 plugin 的页面可以往星鸣画的行菜单里加条目：provider 行（右键）和 Agent 行（它的句柄，或右键）。一个条目会打开该 plugin 的页面并告诉它选中了什么；活儿由页面来做。它和窗口清单的其余部分一起由 `plugin.Window` 读取（[`page.go`](../../internal/plugin/page.go)）；不启动 Bun host。

- **清单。** `starling.actions: [{id, menu, title, titles}]`。`menu` 是 `plugin.ActionMenus` 之一（`provider`、`agent`）；session 的菜单被略去，因为 session 的标题是用户的内容。`id` 要匹配 `actionID`（`^[a-z0-9][a-z0-9-]{0,39}$`），同一个 id 给出两次时取第一个，而一个 plugin 最多 8 个（`maxActions`）。没有页面的 plugin 没有动作：动作是进入它页面的方式。缺标题就用页面的标题，按各语言（[`pluginpage.go`](../../internal/gui/pluginpage.go) 里的 `windowParts`）；`/api/plugin-pages` 把它们带在页面上。
- **菜单。** `starling.actions(menu, context)`（[`pluginpages.js`](../../internal/gui/assets/pluginpages.js)）给出那些条目，在星鸣自己的条目和一个分隔线之后，由 `titled` 命名，按 plugin 写的样子绘制（`openRowMenu` 里的 `literal`，不经过 `t()`）。`openProviderMenu` 传 `{provider: {id, name}}`，`openAgentMenu` 传 `{agent: {id, name}}`，行的其余内容一概不传。
- **选中一个**会运行 `show(key)`（读者保留的未保存编辑会把它拦在那里），等页面的 hello，然后只向那个框架发送 `{event: "action", data: {id, ...context}}`。桥（[`pagebridge.js`](../../internal/gui/pagebridge.js)）把它交给 `starling.onAction(fn)`；早于 `fn` 给出前就到达的动作会等它，而 `fn` 里的抛出异常就是页面的错误（「停止工作了」）。
- **插件页**在那一行说明一个 plugin 给哪些菜单加了东西（[`plugins.js`](../../internal/gui/assets/plugins.js) 里的 `actionsLine`），依据是那一行的 `actions`（[`plugins.go`](../../internal/gui/plugins.go)），关着的 plugin 也照样说明。

测试：`TestWindowActions`、`TestPluginPageActions`、`plugin-actions.test.cjs`。

## 事实来源与组合

plugin 用什么运行是逐字段组合出来的（`Compose`），每个值都记着自己的那一层（`Source`）：

1. `default`：每个 plugin 都开着，用它的 `Manifest.Defaults`；
2. `preset`：Essentials 让每个不是 `Essential` 的 plugin 都关着；
3. `user`：plugins.yml 的那些行（`enabled`，以及逐 key 合并进默认值的 `config`，映射会深合并）；
4. `in use`：`InUse` 大于 0 的 plugin 不管那些行怎么说都是开着的，因为关掉它，它的东西会继续工作却没地方看。`InUse` 报错也算在用。

第 4 步之前，`Composed.Chosen` 就是 plugins.yml 的说法。`starling plugin dump`（`--json`）打印组合后的列表，附上每个 plugin 的状态和原因。

## 运行时路径

- **启动**（`internal/cli/main.go`，在 `settings.Migrate` 之后）：`pluginrt.Settle()` 在没有 plugins.yml 时写出第一个，内容是 `all`，新装也是如此：新用户能看到星鸣做的一切，并关掉自己不要的。低于 `settled: 1` 的文件会被更新一次：新装时从 10-13 写到 10-17 的 `essentials` 变成 `all`，用户自己的行保留。每次写入都设置 `settled`，所以用户之后选的 Essentials 仍是他的选择。然后 `pluginrt.Refresh` 建立这个进程的运行时并让它对齐。错误会被打印出来，星鸣照常启动。
- **每次读取**：`Refresh` 重新加载 plugins.yml、组合它并对齐。窗口的 `/api/state`、`/api/plugins/builtin` 和 boot.js 各自调用它（`builtinsState`），所以 CLI 做的改动在下一次读取时就能看到。
- **开关**：`Set(id, on)`（`POST /api/plugins/builtin`，`starling plugin builtin on|off <id>`）写入用户的那一行；`SetPreset` 设置预设（`starling plugin preset essentials|all`）。plugin 在用时不允许关闭它。
- **窗口**：见 [GUI 应用外壳](gui-shell.md) 的「内置 plugin」。
- **网关**：提供网关服务的进程运行 `pluginrt.KeepFresh`（每 2 秒一次），在 plugins.yml、settings.json 或 providers.json 变化时刷新（`middleware` 还包括 plugins.json），所以在窗口里或用 `starling plugin builtin` 开关的 plugin，无需重启就能在网关里生效或消失，无头的 `starling serve` 也会跟着文件走。

## 失败与兼容性

- 读不了、解析不了或写了别的预设的 plugins.yml 是一个错误，绝不会被当成空文件。`Refresh` 返回该错误并让运行时保持原样；在第一次成功读取之前，每个 plugin 都是开着的，和 plugin 存在之前一样。`Set` 和 `SetPreset` 拒绝覆盖它，窗口在列表上方说明它并禁用所有开关，CLI 则带着它退出。用户修好文件或删掉它。
- 写入持有一把对其他 starling 的 `filelock.Mutex`（`plugins.yml.lock`）和进程内的一把互斥锁，整体写入（临时文件、同步、重命名），并把被替换的文件保留为 `plugins.yml.bak`。
- 为另一个主版本 API 或更新的次版本写的 plugin 会带着原因成为 `failed`；其余照常运行。
- plugin 的页面代码抛出异常时只有那个页面坏掉（见上面的边界）；POST 一个 schema 不接受的配置会得到 400，什么都不写。
- 页面 plugin 的关闭只影响窗口：它的页面、板块和行从窗口里消失。网关、TUI 和 CLI 两种情况下服务的内容一样，用户做的东西不会被删除。网关 plugin 的关闭只影响网关：关掉 `otel` 时不导出任何东西，也不会为它完整保留任何回复。观察者不会让请求失败，它只读。关掉 `rotation` 时，设成它某个路由的池按列出的顺序尝试；失败的策略回落到同样的顺序，绝不变成错误。失败的 JS middleware 让请求保持原样；要求脱敏却做不到时什么也不发送（503）。Davsync 不带 plugins.yml。

## 验证

- `go test ./internal/pluginrt/`（`TestReconcile`、`TestCompatible`、`TestPluginsImportOnlyCore`、`TestCompose`、`TestSettleAndSet`、`TestSettleBringsUpAnEssentialsInstall`、`TestBrokenFileIsKept`）。
- `go test ./internal/pluginrt/ -run 'TestSchemaOf|TestValidate|TestSetConfig'`。
- `go test ./internal/gui/ -run 'TestBuiltin|TestBootSaysBuiltinsOff'`（`TestBuiltinConfig`、`TestBuiltinClientsLoaded`）。
- `i18n-keys`、`gui-de`、`gui-ja`、`gui-zh-tw`，以及搬走的页面自己的测试（`sessions-*`、`session-*`、`sess-*`；`library-*`、`skill-*`、`mcp-*`、`found-skill-*`），在 Chromium 和 WebKit 里。
- `internal/gui/tests/builtins.test.cjs`（设置表单）和 `plugin-kernel.test.cjs`（错误边界、重试、来自 boot.js 和来自一次改动的 `callsPage`），在 Chromium 和 WebKit 里，所有语言，宽 1100、800 和 420；网关页自己的测试（`gateway-*`、`recent-calls-*`）不变。
- P3 观察者：`go test ./internal/usage/ -run 'TestObserver|TestBodiesStayOutOfTheLog'`、`go test ./internal/gateway/ -run 'TestObserverPoint|OTel|Deferred'`、`go test ./internal/otlp/ ./plugins/otel/`（`TestOnExportsOffSendsNothing`、`TestInUseByEnvironment`）和 `go test ./internal/pluginrt/ -run TestKeepFreshFollowsPluginsYml`。
- P3 路由策略：`go test ./internal/routes/ ./plugins/rotation/`（`TestRegisterStacks`、`TestApplyFailsOpen`、各策略的顺序、`TestOnServesOffDoesNot`、`TestInUse`）、整个 `go test ./internal/gateway/`（它的 `TestMain` 注册那些策略，所以每个路由测试都是 plugin 开着的等价性运行；`TestRotationOnAndOff` 把它关掉），以及 `routing-save.test.cjs`（关掉时只剩 Smart，各种宽度和引擎）。
- P3 group：`go test ./internal/provider/ -run 'TestOwnGroupsWaitOnThePlugin|TestNamedGroupsUnreadable'`、`go test ./plugins/groups/`（`TestOnServesOffDoesNot`、`TestInUse`）和整个 `go test ./internal/gateway/`（它的 `TestMain` 持有 `ServeGroups`，是 plugin 开着的等价性运行；`TestGroupsOnAndOff` 把它关掉：用户的 group 离开 `/v1/models` 且不被服务，找到的那个两者都是）。
- P3 middleware：`go test ./internal/gwmw/ ./plugins/middleware/ ./plugins/redact/`（`TestMaskingFailsClosed`、`TestServedRefreshesOnce`、`TestSafelyTurnsAPanicIntoAnError`、`TestAgentsWantedRefreshesWhenWanted`、`TestOnServesOffDoesNot`、`TestInUse`）和整个 `go test ./internal/gateway/`（它的 `TestMain` 提供 `masking.Masker` 和 `middleware.Chain`，是两者都开着的等价性运行；`TestMaskingOnAndOff`、`TestMaskedQueryFailsClosed` 和 `TestMiddlewareOnAndOff` 把它们关掉：要求脱敏时是 503 且厂商什么都收不到，middleware 关闭时请求照原样发出）。
- P5：`go test ./internal/plugin/ -run TestWindowPanels`、`go test ./internal/gui/ -run 'TestPluginPanels|TestPluginPageReplaces'` 和 `plugin-panels.test.cjs`（每个 slot、一个坏掉的面板和重试、尺寸观察的循环提示、一个被取代的资源库、设置标签页的出现与消失），在 Chromium 和 WebKit 里，所有语言，宽 1100 和 420（插件页的那些行宽 980 和 420）。
- P6：`go test ./internal/plugin/ -run TestWindowActions`、`go test ./internal/gui/ -run TestPluginPageActions` 和 `plugin-actions.test.cjs`（两个菜单、一个迟迟才给出处理器的页面、一个什么都不加的 plugin、插件那一行），在 Chromium 和 WebKit 里，所有语言，宽 1100 和 420。
- 手工验证，在沙箱 HOME 下：`starling plugin builtin`、`starling plugin builtin on groups`、`starling plugin dump`，以及一个手工弄坏的 plugins.yml。
