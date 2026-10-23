# 网关 middleware

plugin 也可以是 JS 网关 plugin：一段 JavaScript，星鸣在自己的网关里运行它，处理 Agent 发出的请求和拿回的回复（middleware）、一个池里账户或 key 的尝试顺序（`strategies`），以及每一条用量记录（`onUsage`），无论最终是哪个 provider 在服务它们。它与 [provider 与 plugin 归属](provider-plugins.md)里的 provider plugin 不同，后者跑在 plugin host 里的 Bun 上。middleware 跑在网关自己的进程里，用的是 [moejs](https://github.com/Calcium-Ion/moejs)——一个用 Go 写的 JavaScript 引擎，因为它处在每个请求和每个流式事件上。

## 职责与事实来源

| 部分 | 职责 | 来源 |
| --- | --- | --- |
| 发现 | 找到 plugin 的网关模块：package.json 里的 `starling.middleware`（或它的别名 `starling.gateway`），或者一个指向 `*.middleware.js` / `*.gateway.js`（或 `.mjs`）文件的路径 | [`internal/plugin/middleware.go`](../../internal/plugin/middleware.go) 里的 `plugin.Middleware` |
| Bun host | 跳过只做 middleware 的 plugin（没有 `main`，也没有 `exports`） | [`host.go`](../../internal/plugin/host.go) 里的 `Middleware(...)` 检查 |
| 加载与运行 | 编译、池化运行时、带时限调用 hook、统计调用与失败 | [`internal/middleware`](../../internal/middleware)：`middleware.go`、`run.go` |
| 网关 | 只在 Agent 的请求上运行这条链，作为网关 middleware 点上的一个 `gwmw.Agent`；内置 plugin `middleware` 注册它（`middleware.Chain`），这个 plugin 关闭时没有 middleware 运行 | [`internal/gateway/middleware.go`](../../internal/gateway/middleware.go) 里的 `serveAgent`，由 `gateway.go` 和 `codex_backend.go` 调用；[`internal/gwmw`](../../internal/gwmw/gwmw.go)、[`plugins/middleware`](../../plugins/middleware/middleware.go)；[plugin 运行时](plugin-runtime.md) |
| 路由与用量 | 模块的 `strategies` 就是路由 `<short name>/<name>`（`routes.Register`），它的 `onUsage` 拿到白名单内的用量记录；两者都在网关提供 `middleware.Points()`（`middleware` plugin 注册的一个 `usage.Observer`）期间生效 | [`points.go`](../../internal/middleware/points.go) |
| GUI 与 CLI | 展示每个 middleware 的 hook、调用次数、单次调用的 µs、失败次数和加载错误；它的路由及被请求的次数；它的用量记录以及被丢弃的那些。提供的路由是设置状态里的 `jsRoutings`，出现在每个路由选择器里；`starling group set <id> routing=<plugin>/<name>` 接受其中一个 | `middleware.States()`；[`internal/gui/plugins.go`](../../internal/gui/plugins.go) 里的 `pluginEntryJSON.Middleware`、`plugins.js` 里的 `mwLine`、[`plugins_cli.go`](../../internal/cli/plugins_cli.go) 里的 `listPlugins` |
| 选项 | plugins.json 里 middleware 的 `options`，作为 `ctx.options`；按名称、spec、短名或去掉作用域后的名称设置（`plugin.Called`：agent-aider，在 `agent-`、`tools-` 和 `page-` 也被去掉之前它就是短名）；包里的 `starling.options` 只是编辑器里的示例 | [`store.go`](../../internal/plugin/store.go) 里的 `SetOptions`、`ShortName`；`plugins_cli.go` 里的 `pluginOptions`；`plugins.js` 里的 `optionsEditor` |
| 类别 | 一个 plugin 是 provider、middleware 还是两者兼具，显示在每张卡片和每一行上，关闭的也显示 | `pluginEntryJSON` 里的 `isMiddleware`、`middlewareOnly`；市场条目里的 `kind: "middleware"`（`market.json`、社区的 `registry.json`）；`plugins.js` 里的 `kindChip` |
| 社区 middleware | 现成的包（`@magpie-community/middleware-<name>`），大多是 New API 的渠道设置 | [magpie-community/plugins](https://github.com/magpie-community/plugins) 里的 `packages/{param-override,model-map,system-prompt,word-guard,think-tags,antigravity-cloak}`，用 `scripts/middleware.mjs` 测试 |
| 用户文档 | 面向 plugin 作者的 hook、`ctx` 和一个示例 | [插件运行时](plugin-runtime.md) 与[网关中间件](gateway-middleware.md) |

plugins.json 是「装了哪些 middleware、哪些开着、按什么顺序」的唯一事实来源；关闭的 middleware 不会被加载。

## 运行时路径

1. `serveAgent` 确定[网关 session 身份和可选的录制](gateway-sessions.md)，通过 Bun 运行任何[异步请求 hook](gateway-request-hooks.md)，然后对每个已注册的 `gwmw.Agent`（JS 链的那个是 `middleware.Begin`）调用 `Begin`，带上请求的协议、模型、流式标志、路径和 Agent。没有 middleware 时它直接调用 `serve`。
2. `Run.Request` 把 Agent 的请求体按它自己的 API 形态依次交给每个 `onRequest`。返回的对象就是下一个 middleware 和网关拿到的请求体；返回 `undefined` 则字节保持原样。`ctx.reject(status, message)` 终止这条链，Agent 收到按它自己 API 形态表示的那个错误（`writeError`）。
3. `Run.Tried` 在 `serve` 每次尝试把请求发给某个 provider 时，都在那次尝试发出之前对该请求体运行每个 `onProvider`（`gateway.go` 的候选循环里的 `triedOn`；`serveAgent` 把属于 `gwmw.Tried` 的那些运行放到请求的 context 上）。`ctx.provider` 说明是哪一个：`id`、`name`、`account`（订阅的种类，按 `accountAgent` 给出的：`claude`、`antigravity`……；用 key 时没有这一项）和 `model`，即模型在那里的 id。每次尝试都从 `onRequest` 留下的请求体重新开始，所以发给某个 provider 的内容变化不会影响其他任何一次尝试：这就是 plugin 只对某一家厂商生效的方式，比如 Claude Desktop 这种模型名完全看不出请求去向的 Agent。在这里 `null` 和 `ctx.reject` 都拦不下什么；请求已经被路由了。
4. 当某个 middleware 有 `onEvent` 或 `onResponse` 时，`Run.Wrap` 会包住响应写入端。SSE 回复被拆成一个个事件。数据是 JSON 对象的每个事件会交给所有声明想要它名字的 `onEvent`（`export const events`）：返回对象就替换它，返回 `null` 就丢弃，返回 `undefined` 就写入原始字节。`data: [DONE]` 和其他非对象数据原样通过。完整的 JSON 回复会先缓冲起来交给 `onResponse`，并设置 `ctx.status`。带 `Content-Encoding` 的回复原样通过。
5. `Run.End` 把运行时归还各自的池。

每个 middleware 在整个请求期间拿到同一个运行时，所以它的各个 hook 共享 `ctx.state`。运行时来自每个 middleware 各自的 `sync.Pool`，模块每次加载只编译一次。编译 plugin 的模块（import 保持在 plugin 内部）用的是 [`internal/jsmod`](../../internal/jsmod/jsmod.go)，Agent plugin（[Agent 接线](agent-wiring.md)）也共用它。

只有 Agent 发出的请求才会走 `serveAgent`。星鸣自己的请求（session 标题、路由器的分类器、搜索和视觉的替身）直接调用 `serve`。middleware 看到的是脱敏之前的请求体，脱敏仍然在上游通过 `redact` plugin 的 masker 生效；要求脱敏却没有 masker 提供服务时什么也不发送（[plugin 运行时](plugin-runtime.md)）。

## 路由与用量记录

两者共用网关的观察者 `Points()`（`plugins/middleware` 用 `usage.Observe` 注册它；只有网关进程会提供它）。在提供期间，它每秒看一次已加载的模块，把每个 strategy 注册为 `<short name>/<name>`：包的短名，或者去掉 `.middleware`/`.gateway` 和扩展名后的文件名。重载或模块关闭会把它们移除，而一个由没有任何 plugin 提供的路由来路由的池，按它列出的顺序尝试（`routingOf`）。

- 一个 strategy 就是 `strategies[name]`，要么是一个函数 `(pool, seats, options)`，要么是一个带 `order` 的对象。`seats` 是 `routes.Seat` 的 camelCase JSON（`key`、`account`、`known`、`learns`、`used`、`soon`（unix 毫秒）、`pace`、`weight`、`resting`、`tokens`）。它返回座位索引的一个顺序。
- 每个模块为它的 strategies 和 `onUsage` 各有一个运行时，所以它们共享模块状态；同一时刻只跑一个调用。
- `onUsage(row, options)` 拿到 `usageRow` 里白名单的一条记录：时间、操作、Agent、provider 及其账户和 key、调用方的 key、模型、请求量、服务量、token 数、毫秒、ttftMs、状态、errType、类别、是否被拒、尝试次数和费用。没有请求体、提示词、错误文本、session 或请求 id。记录先进队列（1024），由一个 worker 取出，在请求路径之外处理；队列满了就丢弃该记录并计数。
- 这种形式的路由（`provider.PluginRouting`：一个由 `A-Za-z0-9._-` 组成的名称、一个斜杠、另一个名称）会按保存的样子留在 providers.json 里，无论是 provider 的还是 group 的 Save 都一样，不管当下有没有 plugin 提供服务，这样把 plugin 关掉再打开，由它路由的 group 仍然保持。文件的基本名会被做成这样的名称（其他字符变成 `-`）。`starling group set <id> routing=<plugin>/<name>` 接受同样的形式。

## 约束与失败行为

- 时限：`onRequest`、每个 `onProvider` 和 `onResponse` 各 250 毫秒，每个 `onEvent` 50 毫秒，加载 1 秒。超过时限的 hook 会被中断。
- 失败放行（fail open）：抛出异常、返回既不是对象也不是 `null`/`undefined` 的东西（`onProvider` 返回 `null` 也一样），或者被中断的 hook，都让请求或事件保持原样。该失败会被计数并保留为最近一次错误，日志最多每分钟 60 行。
- 一个 middleware 只 import 相对路径的文件；依赖的包必须打包进去。
- 重载：对 plugins.json 或某个 middleware 文件的改动，最多一秒后就会被下一个请求读到（`stamp`）。
- 一个 strategy 有 20 毫秒，含等待运行时的开销；抛出异常、超时或返回的不是座位顺序的，会让座位保持列出的顺序（fail open），并被计为失败、记入日志。`onUsage` 有 250 毫秒，同样失败放行。请求永远不会等 `onUsage`。
- 统计覆盖整个调用，包括把 JSON 解析进运行时再写回去。调用在运行网关的进程里计数，所以 `starling plugin` 只显示 hook 和加载错误。

用真实账户和真实 Claude Code 测得：每个事件约 1–1.6µs；一个解析并重写 Claude Code 约 80KB 请求体的 `onRequest` 约 0.2–0.7ms。`BenchmarkEvent` 约 0.9µs，11 次分配。

## 验证

```sh
go test -tags nogui ./internal/middleware   # TestStrategy*, TestOnUsage*, TestDocsExample: points.go
go test -tags nogui -run 'TestParseRouting|TestPluginListSaysMiddleware' ./internal/cli
go test -tags nogui -run TestPluginRoutingKept ./internal/provider
go test -tags nogui -run TestOnServesStrategies ./plugins/middleware
node --test internal/gui/tests/routing-js.test.cjs
go test -tags nogui -run 'TestMiddlewareThroughGateway|TestMiddlewareOnProviderPerTry' ./internal/gateway
go test -tags nogui -run TestPluginListSaysMiddleware .
go test -tags nogui -run "TestSetOptionsShortName|TestBuiltinMarket" ./internal/plugin
go test -tags nogui -bench Event ./internal/middleware
node --test internal/gui/tests/plugin-middleware.test.cjs
# in magpie-community/plugins: each package runs its cases.json through the hooks
bun test packages/think-tags packages/model-map packages/param-override packages/system-prompt packages/word-guard packages/antigravity-cloak
```
