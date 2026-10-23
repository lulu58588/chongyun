# 代码规范

<a id="code-standards"></a>

这些规则来自 PR 评审里反复提出的要求。评审者会逐条检查，不满足哪一条就会留下
一条评论说明是哪一条。如何描述和评审某个子系统的改动，见
[subsystem references](subsystems/README.md)。

## 一次改动包含什么

<a id="what-a-change-contains"></a>

| 规则 | 评审者会看什么 |
| --- | --- |
| 一次改动只做它的目标需要的事。 | 行为上的每一处变化都有理由。无关的重构、额外功能和策略变更要放进单独的 PR。 |
| 一个修复恰好覆盖它要覆盖的场景。 | 哪些场景变了，哪些相邻场景保持不变。不同的错误不能当成一个来处理。 |
| 一次改动在整条路径上都站得住。 | 从用户输入一直跟到最终结果，穿过调用方、它改变的状态和副作用。一个函数正确不等于整个操作正确。例如，agent 的模型选择存的是一个 id，读出来却是另一个（#926）。一个更新按钮并没有更新真正在跑的那个二进制（#930）。 |
| 一个修复覆盖这个 bug 的每一个同类位置。 | grep 同样的模式：每一个 agent、同形状的 subcommand、guard，以及被改动判据的每一个调用方，都要在同一次改动里拿到这个修复。这包括同一操作的其他分支，以及提到同一个 endpoint 或规则的其它消息。2026-10-05 有六个修复需要再发一次版本，因为它们只覆盖了报告里那个场景。10-06，#933 的 Codex 模型在登录后隔了一小时才跟上。10-07，cd8d64a9（#1059）在一处提示里给出了 sub2api 的余额 URL，而余额读取失败后显示的错误仍然只提 new-api 的。Pi 对 mcp.json 的整文件移动直到 d40fee80（#1097）才有备份，尽管旁边的合并分支早就有了。 |
| 一次改动会重新推导所有建立在它之上东西。 | 当一次改动改变了公式、列表、布局断点或某个异步步骤的时序时，找出每一个假定了旧含义的调用方、保存的副本和测试，并跑一遍。12cba74a 在新的 max_tokens 下保留了 Zed 旧的 clamp。5d871723（#860）把 Requests 页面的列宽改到 1150px，而 click-scroll 在 10-07 大部分时间在 main 上都是红的（82a36a35）。ee2588c2 加了一次立即重读，测试中关于中间状态的断言在 Linux CI 上变成了竞态（3e38bd76）。 |
| 一次读取、解析或哈希失败意味着"未知"。 | 它绝不意味着"空"、"无"或"相等"。保留之前读到的内容，把解析不出来的文件先复制到一边再覆盖写入，并修好共享读取器，而不是在旁边另写一个更严格的。一个用于限制访问的列表（白名单、key 数量上限）在文件读不出来时必须失败关闭。用文件读不出来的情况来测试。readLogins 把解析错误写回成"没有账号"（b363327d），两天后 agent/stash.go 有了同样的 bug（270b5257）。10-12，7a9490638 的 IP 白名单把读不出来的 settings.json 当成空列表，放任每个地址进来（7d9382db2）。10-13，#2017 的补丁把读不出来的 codex-autoreset.json 读成"没有重置额度用掉"（e297a8ab2）。 |
| 别的 app 的取值来自那个 app。 | 当星鸣写另一个 app 的文件或代它说话时，每个字段的类型和取值都要取自那个 app 自己的数据、二进制、合并代码或文档，而不只是它的字段名。84554e57 把 updatedAt 写成毫秒，卡死了 Claude Desktop 的 skills 页面（#863）。#966 猜了 Codex 的 limit_reached，#996 在读了 ChatGPT.app 之后把它撤销了。10-13，accc8c090 抄了 Claude Code 的 OAuth client 名字，却发了自己的 redirect_uri（127.0.0.1，而 Claude Code 发的是 localhost）；e0ecf0161 读了二进制并修好了。 |
| 会阻塞的东西绝不在它的读取方需要的锁下运行。 | 绝不要在持有一个对方先拿的锁时往 channel 上发送或调用任何可能阻塞的东西。在交接（让另一个 goroutine 来看共享状态）之前先设好共享状态，而按值复制的 struct 里装着 map 并不等于复制。假服务器的 handler 跑在很多 goroutine 上：它记录的东西要用锁保护，并且只统计测试自己发起的调用，因为后台读取也会打到它。启动 goroutine 或 timer 的改动在推送前要跑 `-race`。85fd35ab 卡死了 macOS 的 -race CI（4fb8e887）。01e66e0f 的 marshal 往一个共享 map 里写（15afd988）。10-11，7cba316b9 的 turn-done timer 在自己的 goroutine 上读字段，在 -race 下变红（a2afc1828）。10-13，TestKeptLoginStandsAtItsPlace 的假 Codex 从两个 goroutine 追加（a75de4c21）。 |
| 星鸣的默认值绝不覆盖用户设过的值。 | 在写某个 key 之前，先列出 agent 或星鸣会从哪里读到这项设置，并在存在用户设定值的情况下做测试。会重写某项设置的迁移要列出那些自己设过它的用户，并保留他们的值，或者在发布说明里讲清楚为什么不保留。CLAUDE_CODE_AUTO_COMPACT_WINDOW 覆盖了用户自己的 autoCompactWindow 大约 150 个版本（602d4ec6）。10-10 的四个修复（603dc0c4、5aca7d09、c65d9fb4、d9bf7c1c）都是在一个持有用户设定值的测试都没有的情况下开工的。10-17，5a2414acd 把每一个未标记的 Essentials 预设都变成了 All，包括用户用 `starling plugin preset essentials` 选的那个（#2749）。 |
| 检查覆盖改动触及的边界。 | 对数字，检查边界值。对路径，检查每个平台（macOS、Linux、Windows）。对并发，检查竞态和过期状态。按风险挑选检查项。 |

## 测试

<a id="tests"></a>

| 规则 | 评审者会看什么 |
| --- | --- |
| 一个修复要配一个没有它就会失败的测试。 | 评审者把旧代码放回去跑测试，这个测试必须失败。PR 描述要写明它以什么方式失败。一个以某条保证命名的测试，在那条保证被手动破坏时必须变红：#1318 的 TestHistoryWriteHoldsNoWaiter 在 waiter 被持有时依然通过（#1358）。10-12，bbe8825c0 的 memo 测试在淘汰逻辑被反过来时依然通过（9dd5917a4）；10-13，e21b1dbc1 的测试在 is_error 被去掉时依然通过（cc93b566e）。 |
| 测试绝不触网。 | 厂商的页面或 API 由 httptest 提供，用从真实环境保存下来的字节构建。依赖包来自 lock 文件，而不是 registry 上的最新版。破坏这次改动，然后观察测试的请求去了哪里。9755201b 打开了 ChatGPT，卡在它的 Cloudflare 挑战页上。10-11，一个库测试在图标代理关掉的情况下访问了真实 GitHub（f04792f9）。10-13，TestPiPackage 测试了 npm 当时的最新 pi（f230fa23b），plugin 和 cli 测试在真实的 models.dev 上等了 30 秒（bde11c778）。 |
| 测试用真实输入。 | fixture 看起来像真实请求、真实序列化输出和磁盘上的真实文件。这包括缺失字段、空值和默认值。 |
| 测试检查行为，而不是实现。 | 断言用户或调用方需要的东西。格式改动和内部重构不该弄坏测试，除非要求的就是精确字节。 |
| 在 `main` 上失败的测试就是要马上修的 bug。 | "在 origin/main 上也一样失败"不是基线，"整套跑时失败、单独跑时通过"也不是。修掉根因，或者开一个 issue 点名根因，并且用单独的提交来做，之后才能在它之上继续发布。从 10-05 到 10-07，有几十个提交在红灯测试之上发布了。有些是真 bug：一次数据竞态（a45b3e09）、测试顺序（bfe8bcaf），以及自 5d871723 起就一直红的 click-scroll，而 dcbae2ef、76d82528 和 b2345078 都在它之上发布了。在把一个 Go 测试叫作 flaky 之前，先跑 `go test -race -count=20 -run X`。 |
| 测试不碰真实机器。 | 在临时 HOME 下运行它们，并把 Go 的缓存固定住（见 [provider-plugins.md](subsystems/provider-plugins.md)）。绝不碰任何真实 agent 的配置或 `~/.config/starling`。agent 的变量要放进 `agentenv.Vars`，这样沙箱才会清掉它。 |
| 测试启动的东西，测试结束时也要结束。 | 测试放出去的假进程、goroutine 或异步刷新，必须在测试返回前被停掉或等完，否则它会写进下一个测试的目录（"directory not empty"）或者空转好几天。一个在自己 `TestMain` 里隔离自己、而不是用 `testenv.Run` 的包，要设置 testenv 设置的那些东西，包括 `testenv.Marker`。TestClaudeSignInByPaste 的假进程轮询了 2 天（38a0f470）。10-11，一个比测试活得久的 CLI 身份请求弄坏了下一个测试（#1524）：4ee3a76b 只在 internal/gateway 里修了它，9c377231 则通过 marker 为每个包修了它，而 gateway 从没设过这个 marker（a349de82）。 |
| 计时测试两边都要留余量。 | 绝不要用一个负载满满的 runner 可能超掉的挂钟时间来限制测试，并且让阈值分类器的 fixture 远离它的每一个阈值。复现一次计时失败要靠注入延迟（一次被挂起的写、一次迟到的首次到达），而不是反复重跑到它失败为止；修一个测试时把里面每一个计时假设都列出来。319b72c7 修了两个共用 2000 行上限的循环中的一个，另一个在 9 小时后变红（229ce19a）。10-11，TestMaxRPM 的 flaky 其实是个真漏洞：一个迟到抵达厂商的首次请求，让厂商在它那一分钟里多看到一次（ce1ecaa4）；而排队流的测试在 CPU 饥饿下把客户端时钟当成了权威（3f4287e6）。 |

## 验收标准

<a id="acceptance-criteria"></a>

一次改动只有做到下面每一条才算合并或发布。要说明跑了哪几条、它们打印了什么。
没跑的检查要列为未跑，绝不能略去。

1. **一个没有这次改动就会失败的测试。** 把旧代码放回去跑新测试。它必须以
   一个点名了行为的真实 `FAIL` 失败。编译错误或被跳过的测试都不算。要说明
   哪个测试失败、它以什么方式失败。然后恢复改动，确认测试通过。
2. **在沙箱里跑测试。** 在临时 HOME 下跑 Go 测试，并固定 Go 的缓存：换掉
   HOME 之前先从 `go env` 取出 `GOPATH`、`GOMODCACHE` 和 `GOCACHE`。之后
   用 `chmod -R u+w "$h"; rm -rf "$h"` 移除 HOME，因为其中的模块缓存是只读
   的。片段在 [provider-plugins.md](subsystems/provider-plugins.md#verification) 里。
   不固定缓存的话，每次运行都会往新的临时 HOME 里下载 1–2 GB。测试绝不读取
   或写入运行中的 agent 配置（`~/.codex`、`~/.claude`、`~/.claude.json`、
   `~/.gemini` 等等）或 `~/.config/starling`。各个包在它们的 `TestMain` 里设置
   这些，大多数用 testenv。
3. **构建和测试套件。** 对改动过的文件跑 `gofmt -l`（它不输出任何东西）、
   `go vet ./...`、`GOOS=linux go vet -tags nogui ./...`、
   `GOOS=windows go vet ./...`、`go build`、
   `GOOS=linux go build -tags nogui`、`GOOS=windows go build` 和
   `go test -tags nogui ./...`。跨 OS 的构建不会编译 `_test.go` 文件；vet 会。
   10-08，bbf99143 的测试调用了 `syscall.Getsid`，Linux 的 syscall 包没有这个
   函数，而 bb8a3e30 和 e7de897f 是没格式化就发布的。在一串命令里，设置
   `set -o pipefail`，这样 `go test | tail && git push` 才会在失败时停下。跑
   `./...`，而不是你以为你改过的那些包：仓库的守护测试就住在你这次改动没碰
   到的包里。当新的 agent 变量不在 `agentenv.Vars` 里时，TestFolderVarsAreListed
   （internal/agent）会失败；当代码调用 `exec.Command` 而不是 `proc.Command` 时，
   TestNoCommandBypassesProc（internal/proc）会失败；而一个输出格式改动会弄坏
   另一个包那个钉住了旧文本的测试。10-13，有四个提交只跑了挑出来的包，就这样把
   main 弄红了：6fa3109ea（494609866）、43d8a5633（24146c031）、5617281f5 和
   cdb7b5ff7（c6de0d2a5）。
4. **两个引擎里都跑 GUI 测试。** 对 `internal/gui/assets` 的改动，要把它触及的
   Playwright 测试在 Chromium 和 WebKit 里都跑一遍（`make test-ui`，或
   `BROWSER=webkit node --test …`）。测试套件还必须在中文、英文，以及
   `gui-zh-tw`/`gui-ja`/`gui-de` 下通过（每条字符串都有它的繁体中文、日文和
   德文版本，占位符相同）。CI 不跑这个套件，所以要在本地跑。当改动新增了 GUI
   会调用的 API、其他页面共用的选择器或 class、较老 WebKit 不支持的 CSS 特性
   （`:has()`），或者一个布局断点时，要跑整个套件，而不只是被触及页面的测试。
   10-07，479865bc、134d388b、61838f7c 和 51626258 修好了四个被那些只跑了自己
   页面测试的提交留成红色的测试。10-08，aff4f9f2 的 zh-TW 是在 685fa5dc 和
   9dc0121f 的字符串之前构建的（b1e795e5 在 10 分钟后又加了四条），而
   usage-ledger 在 586f2bce 之后红了大约 7 小时。当你改写一条字符串时，在
   `internal/gui/tests` 下 `git grep` 旧文本，并测试空的情况（一个没有列表价的
   模型，一个没有用量的账号）。
5. **尽可能做真实使用验证。** 对 provider、订阅或 agent 的改动也要对照真实
   情况检查：真实账号或 key、agent 的真实配置格式、厂商的真实回复。在沙箱 HOME
   里、用凭证的副本做这件事。绝不碰运行中的网关或正在跑的星鸣。如果做不到
   真实检查，就直说。没有用真实东西试过的部分不会被建立在上面，也不会被当作
   可用来告诉用户。平台报告（Windows、WSL、Linux）要在那个平台上检查；Windows
   那台机器和它的 WSL2 发行版都在，而 GOOS 构建不是测试。网关 guard 的测试要
   走真实服务器栈或构建出来的二进制（b2345078 调用 Handler 时没有带 lanGuard，
   于是每次这样的调用都拿到 401）。当对某个厂商的猜测不一致时，设计一个能把
   它们区分开的检查，或者请报告者去跑。10-08，bbf99143 说"没在 Linux 上试过"
   然后弄坏了 Linux 的 vet；#1185 只在 Linux 上出现的 reap 失败只在 macOS 上
   被评审。a24458ca（#1063）、5d0c1dca（MSIX 没试，说得很清楚）和 #1328（在
   Windows 那台机器上，六个破坏性检查）是范本。
6. **[GUI](#gui) 下的 UI 规则**成立：`t()` 字符串、用应用自己的菜单而不是
   `<select>`、点击不滚动、不用彩色左边框条纹。

一个 PR 只有在维护者把它拉下来、合并到当前的 `main` 上、并在结果上跑过上面
的检查之后才算合并。合并时要点名被评审的那个提交：
`gh pr merge --match-head-commit <reviewed sha>`。评审期间发生的推送会随后
挡住合并，而不是发布没人跑过的代码。评审之后被推上来的 head 要重新评审，评论
里要点名新的 head（#974 就是没这么做就被合并了）。在 PR 上写明跑了什么、在
哪个 head 上跑的，然后再合并：一次没有留下"跑了什么"记录的合并事后无法核查
（#981、#1000、#1030、#1037）。10-07 的 #1045、#1046、#1047、#1049、#1057、
#1060、#1061 和 #948 是范本。它们各自点名被评审的 head 和跑了什么，并且发在
合并之前。Draft PR 在被标记为 Ready for review 之前不能合并。

一次发布只有在下面全部成立时才算数：

- 在被打 tag 的提交上（或包含它的更晚的提交上）有一次 CI Test 运行跑完且是
  绿的。被下一次推送取消掉的运行不算通过。
- 标签 `vX.Y.Z` 在 `main` 上，是 `origin/main` 的祖先。
- `yetone/magpie-releases` 里的那次发布有全部 16 个资源。
- 那次发布被标为 Latest，也就是版本号最高的那个。

先检查 `git tag` 找下一个空闲的版本号，因为发布可能从不止一个地方做。
`git fetch --tags` 并确认这个提交还没有被打过 tag：10-12 有两个会话相隔一分钟
给 396a6a3d4 打了 tag，结果版本号较低的 v0.1.1192 成了 Latest。tag 不是发布：
在把版本号告诉任何人之前，先确认那次发布存在并且有 16 个资源；构建失败的发布
要报出来，而不是让下一个 tag 把它盖过去。v0.1.1213（hdiutil "Resource busy"）
和 v0.1.1218（公证时断网）从未发布，而且第一个 1.5 小时里没人发现（#1846）。
plugin 的版本号同理：在迁移者的 `min` 或某个回复点名它之前，先跑
`npm view <package>@<version> --prefer-online`。只增加测试的 PR 不会得到自己
的发布。

## 评审一个 PR

<a id="reviewing-a-pr"></a>

这些模式来自近期的评审。评审者会逐条检查其中适用的：

- **跑一遍。** 把 head 拉下来，合并到当前的 `main` 上（需要时解决冲突），然后
  跑验收检查。只读 diff 不算评审。
- **故意把它弄坏。** 回退这个修复，或者用几种方式破坏它守着的代码，确认新
  测试能抓到每一种破坏（#901 用四种方式破坏了 `backup.go`）。评审者也可以
  加自己的评审测试（#829–#831）。
- **查厂商，而不是 PR 对它的说法。** 读上游源码来了解厂商的行为（#870 读了
  Codex 的 `uses_openai_actor_authorization`）。把这个改动拿真实的试：
  #853 发现大多数 Chat 上游不会接着处理一条结尾的 assistant 消息。
- **检查每个平台和架构。** 一个在 arm64 上安全的 `int` 转换，在
  `GOARCH=amd64` 上可能溢出（#877）。Windows 把被拒绝的连接报告为
  `WSAECONNREFUSED`，而 `syscall.ECONNREFUSED` 匹配不上（#805）。
- **跟完整条路径。** 检查一个模型选择是以同一个 id 存下并读出的（#926）。检查
  一个按钮在真正在跑的那个东西上有效（#930）。检查别名和 User-Agent 不与另一个
  agent 的冲突（#926）。`agentenv.Vars` 里的每一个名字都被当作文件夹路径，
  除非它在 `agentenv.NotPaths` 里（#922）。
- **窄化分类。** 一个标记或错误类别只覆盖它描述的那种情况：#873 把一轮里的
  每次失败都算成"被拒"。
- **预览不写任何东西。** 一次 dry run、plan 或 preview 没有副作用（#827）。
- **不要推翻深思熟虑的决定。** 滚动 guard、按钮光标和其他产品决定是维护者才能
  改的（#820、#921）。改了其中之一的 PR 留给维护者处理。
- **GUI 测试套件。** 和 `main` 比较通过数。在 `main` 上也失败的测试不是基线
  （见 [Tests](#tests)）。一个说自己的失败"在 origin/main 上也一样失败"的 PR，
  要去核实这一点，并在合并前把这些失败修掉或报成 issue。
- **文档和源码一致。** 一处引用要对照它链接到的代码来检查，而且每个链接都必须
  能解析（#889）。
- **让 PR 守住它的目标。** 无关的测试和重构要放进单独的 PR。注释点名函数，而
  不是行号。fixture 的名字要和它代表的东西对得上（#901）。

## Go

<a id="go"></a>

- 写 `any`，不要写 `interface{}`。
- 用 `appdir.Getenv` 读取那些指向文件夹的环境变量，它会忽略相对路径。
  `agentenv.Vars` 里那些不指向文件夹的变量要放进 `agentenv.NotPaths`。
- 注释用一两句话说明代码是干什么的。点名函数，而不是给行号，行号会过时。

## GUI

<a id="gui"></a>

- 用户看到的每一条字符串都要经过 `t()`，并在 `i18n.js` 里拿到它的翻译。测试
  在中文和英文下运行。
- 不要用原生的 `<select>`。下拉菜单要用应用自己的菜单（`openProtoMenu`）。
- 点击绝不滚动页面（`scrollOnPurpose`）。
- 不要用彩色左边框条纹。用一个点或一个色块来标记状态。

## 提交

<a id="commits"></a>

提交信息先点名它改动的领域，然后说用户现在看到什么，用平白的说法写：
`gateway: a 429 that says the balance is spent is out of credit, not a rate limit`。
给出 issue 编号或报告者。正文说明它是怎么被验证的，以及没有这次改动就会失败
的是哪个测试。

"跑了什么"的记录写在提交正文里，以便之后的评审读到它：不是写在推送备注、聊天
里或者只写在 PR 上。一个没有正文的大提交事后无法评审：586f2bce 改了 5133 行，
只有一个标题。10-12，fjrtdk 的 ea3d480ed（+237）和 liubing 的 1cd814004
（+505）在没有任何记录的情况下被合入，而它们的评审提交在它们里面发现了真 bug
（1f0ff81df、9935119f1）。10-13，3ccbe6bc1（plugin runtime P0，2137 行）把
它的全套测试结果写在"推送备注"里；它留下了两个 GUI 测试是红的，当天晚些时候
由 2dcc51417 发现。当一个社区提交在没有正文的情况下被合入时，评审提交要说明
针对那位作者的代码跑了什么。

一个 checkout、一个 worktree 和里面的日志属于创建它们的会话。在你自己的
worktree 里工作（`git worktree add --detach`），绝不 checkout、提交或清理别的
会话的；提交之后检查 `git show --stat HEAD` 列出的正是你的文件、且只有你的文件。
10-09，2018882e 带着 davsync 的标题，却装着另一个会话的 GTK 修复，取自一个共享
的 index。10-12，3e88fc09d 带着 auto-mode 的标题，却装着 Get started 卡片
（#1825）。10-17，另一个会话在 7660399a1 的 worktree 运行中途 checkout 了
origin/main，另一个会话还删掉了 17557132c 的 GUI 日志的一部分。
