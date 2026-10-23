# 插件工具

社区插件（即插件宿主里的 OpenCode 插件，一个 npm 包）可以提供工具：OpenCode 的 `tool` 钩子，`{tool: {name: tool({description, args, execute})}}`。在 OpenCode 里只有 OpenCode 自己的模型能调用它们。星鸣改为把它们提供给所有 Agent，作为自带的一个 MCP 服务器，地址在 `/mcp/starling/tools`（Streamable HTTP），任何能接入 MCP 服务器的 Agent 都可以被指派它。它就是网关的工具接入点（[插件运行时](plugin-runtime.md)的 P4）向社区插件开放的那一处。内置插件 `tools` 负责提供它。同一宿主里的供应商插件见[供应商与插件归属](provider-plugins.md)；网关中间件见[网关中间件](gateway-middleware.md)。

## 职责与事实来源

| 部件 | 职责 | 来源 |
| --- | --- | --- |
| 收集 | 每个已加载插件的工具：名称匹配 `^[A-Za-z0-9_-]{1,48}$` 且带 `execute` 的，同名以最后一个为准，与 OpenCode 一致。每个工具的 `args` 转成 JSON Schema：zod 形状（v3 或 v4，含 `.describe`、`.optional`、`.default`、枚举、数组、嵌套对象）或 zod 对象，而普通 JSON Schema 对象则原样保留。它们随插件列表一起返回（`Loaded.Tools`）。 | [`host.js`](../../internal/plugin/host.js) 中的 `toolsOf`、`jsonSchema`、`argsSchema`、`toolList`；[`tools.go`](../../internal/plugin/tools.go) 中的 `Tools` |
| 运行 | 宿主的 `tool` 调用：先用工具自己的 zod schema 校验参数（补齐默认值；参数错误会指明其路径），再把 `execute(args, ctx)` 与 `ToolTimeout`（55s，低于 MCP 客户端给一次调用的 60s）放在一起竞速。`ctx` 是 OpenCode 的：`agent` 是网关所知的调用方 Agent，`abort` 在超时时触发，`directory` 是宿主的。字符串应答就是文本，`{output}` 是它的输出，其它任何形式都是它的 JSON。 | `host.js` 中的 `parsedArgs`、`runTool`；`tools.go` 中的 `RunTool` |
| 记录 | 每次加载之后，宿主在 `plugin-tools.json`（设置目录）里记录哪些插件提供了哪些工具，整体写入且仅在有变化时写。`ToolsOn` 统计 plugins.json 里处于开启状态且在这里有工具的插件。无法读取或解析的记录算错误，插件因此保持开启。`ToolNames` 是插件宿主未运行时 Plugins 页面的那份记录。 | `tools.go` 中的 `noteTools`、`ToolsOn`、`ToolNames` |
| 提供 | `tools/list` 给每个工具命名为 `<plugin>_<tool>`：插件的短名，与它的 Plugins 行所写的相同（`plugin.ShortName`，它会去掉作用域前缀和一个开头的 `middleware-`、`opencode-`、`tools-`、`page-` 或 `agent-`：opencode-weather-tools → weather-tools，@magpie-community/tools-calc → calc；从文件夹添加的插件则用文件夹名），最多 24 个字符，只含 `[A-Za-z0-9_-]`，后面接工具名。每个名字最多 64 个字符，有两个名字会撞车时后者加 `_2`、`_3`。`tools/call` 运行其中一个；名字未知则报 -32602 错误。它与搜索 MCP 服务器一样受保护（`serveMCP`：网页发来的请求被拒绝，远程请求需要密钥），请求体上限 4 MiB。 | [`tools_mcp.go`](../../internal/gateway/tools_mcp.go) |
| 开关 | `plugins/tools`（"插件工具"，需要 `gateway`，Essential）提供 `gwtool.PluginTools`。`InUse` 就是 `plugin.ToolsOn`，所以只要有处于开启状态且提供工具的插件，它就是开的。网关的 `KeepFresh` 也会盯 `plugin-tools.json`。 | [`plugins/tools`](../../plugins/tools/tools.go)、[`gwtool.go`](../../internal/gwtool/gwtool.go)、[`gateway.go`](../../internal/gateway/gateway.go) |
| GUI | 插件处于开启状态且提供工具时，它的每个已安装行按 MCP 名字列出这些工具，悬停时各自显示描述，另有一个按钮可复制服务器地址（`mcpAddress`）。只有工具的插件挂 Tools 标签，而不是 Provider。Discover 会把 `kind: "tools"` 的市场条目列在 Agent 的 Tools 之下。 | [`plugins.go`](../../internal/gui/plugins.go) 中的 `pluginEntryJSON.Tools`；`plugins.js` 中的 `toolsLine`、`kindOf` 和 `KINDS.tl` |
| CLI | `starling plugin` 用同样的 MCP 名字列出插件的工具，插件页面则按标题和它可以读取的内容列出。宿主不加载处于关闭状态的插件，所以这类插件不列出任何工具。 | [`plugins_cli.go`](../../internal/cli/plugins_cli.go) 中的 `toolsLine`、`pageLine` |
| 用户文档 | 如何编写一个工具并把地址加入某个 Agent | [插件工具](plugin-tools.md) 与[插件运行时](plugin-runtime.md) |

plugins.json 是判断哪些插件开启的事实来源。宿主只加载这些插件，所以只有它们的工具会被列出。

## 运行路径

1. Agent 的 MCP 客户端向 `/mcp/starling/tools` 发起 `tools/list`。当这个接入点未提供、或没有插件开启时，`servedTools` 返回空，插件宿主（Bun）也不会为此启动。否则它向宿主索取各插件的工具（`plugin.Tools`，必要时自行启动宿主）并给出名字（`namedTools`）。
2. `tools/call` 在同一份列表里按 MCP 名字找到该工具。接入点关闭时它直接回答 `isError` "starling's Plugin tools plugin is off…"，不询问宿主。
3. `toolsMCPRun` 通过 `reserveFor` 把这次调用计入调用方网关密钥的限额，与聊天 API 和搜索 MCP 服务器的方式相同（[插件运行时](plugin-runtime.md)，预算接入点）。被拒绝，或限额已满且"密钥限额"插件关闭，都是 `isError` 应答，工具不会运行。
4. `plugin.RunTool` 让宿主运行它。应答，或错误（`<tool>'s arguments are wrong: …`、`<tool> took longer than 55s`，或工具抛出的内容），作为这次调用的文本回给 Agent，出错时带上 `isError`。

工具的应答回到调用它的那个 Agent，无论它在本机还是在一台带密钥的共享电脑上。它绝不会发给厂商，所以脱敏与它无关。Agent 接下来发给模型的内容，则像任何请求一样会被脱敏。

## 约束与失败行为

- 关闭就是关闭：不列出任何工具，调用会说明插件已关闭，并且不询问宿主（`TestToolsMCPOff`）。
- 插件的工具在插件宿主里以用户权限运行，和它的供应商代码一样。安装插件就等于信任它。
- 一个宿主服务所有插件。永不返回的工具会在 `ToolTimeout` 处得到应答，它在宿主里占的那个调用槽位留给它自己的中止来结束；宿主继续提供服务（`TestPluginTools`）。
- 对给定的一组开启插件，名字是稳定的。若在 plugins.json 里更靠前的位置开启了一个提供同名 MCP 工具的插件，后面那个会被重新编号。
- 接线由用户负责：Plugins 页面负责复制地址，星鸣不会为此写任何 Agent 的 MCP 配置。目前还没有对应的资源库条目。

## 验证

```sh
go test ./internal/plugin/ -run Tool
go test ./internal/gateway/ -run 'ToolsMCP|ServingGates|OwnMCPRejects' -timeout 45m
go test ./plugins/tools/ ./internal/pluginrt/
go test ./internal/gui/ -run 'Builtin|Plugins'
node --test internal/gui/tests/plugin-tools.test.cjs internal/gui/tests/builtins.test.cjs
```

测试加载的示例插件是 [`internal/plugin/testdata/tools`](../../internal/plugin/testdata/tools/index.js)：带 zod 形状的工具、原样 JSON Schema、一个会睡过超时的、一个会抛错的，以及两个被排除在外的。
