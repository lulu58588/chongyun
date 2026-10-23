# GUI app 外壳

<a id="gui-app-shell"></a>

桌面 app 有一个托盘（菜单栏）图标，点它会落下一个紧凑的面板，另有一个常规窗口。两者都显示同一个小的 web UI，由内嵌的资源提供。页面通过同一个 handler 上的 JSON API 和 Go 通信，没有绑定生成器也没有打包器。`starling web` 把同一个页面提供给浏览器，用于没有桌面的机器。包 `internal/gui` 承载了这三者。

## 职责与事实来源

<a id="responsibilities-and-sources-of-truth"></a>

| 部分 | 职责 | 来源 |
| --- | --- | --- |
| 原生宿主 | `Run` 启动 Wails v3 app：托盘图标、面板、主窗口、单实例交接、`starling://` 导入链接和开机自启。`host` 为 API 实现 `Windows`：显示和隐藏、打开 URL 和文件夹、复制到剪贴板、调整面板的尺寸和色调，以及设置文字大小。在 Linux 上复制会先设 PRIMARY 再设 CLIPBOARD，因为在没有主选择的合成器上，GTK 3 在 PRIMARY 之后设置时会把 CLIPBOARD 报成空的（KDE Plasma 关闭中键粘贴，#1881）；在 Wayland 上复制是通过合成器读回来的，而读不完整的那个会让 `/api/copy` 回答 501，于是页面先试自己的剪贴板 API，再试复制命令，然后说复制失败并把要复制的文本给出手动复制（`copy-fallback.test.cjs`）。 | [`app.go`](../../internal/gui/app.go), [`clipboard_gtk3.go`](../../internal/gui/clipboard_gtk3.go), [`clipboard_gtk3.h`](../../internal/gui/clipboard_gtk3.h), [`mainwindow.go`](../../internal/gui/mainwindow.go), [`traymenu.go`](../../internal/gui/traymenu.go) |
| 后端 | `startBackend` 启动这个进程所服务的网关，除非已经有另一个星鸣在服务它。那种情况下这个进程只显示对方的状态（`gw` 是 nil）。`stopServing` 把在途的请求排空。 | [`backend.go`](../../internal/gui/backend.go) |
| API 与资源 | `Handler` 提供 `assets/`（用 `//go:embed assets` 内嵌）和 `/api/*` 路由，比如 `/api/state`、`/api/set` 和 `/api/agents/{action}/{id}`。`held` 让一个 API 请求的多次读共享同一次目录构建（`provider.Hold`），并在任何写之后把它丢弃。`revalidated` 和 `versionedPage` 按内容哈希给每个脚本命名，所以缓存绝不会在新页面下给出旧的 `app.js`；按当前哈希请求的文件（`?v=`）会被永久缓存（`immutable`），其它则由它的 ETag 重新验证。 | [`api.go`](../../internal/gui/api.go) |
| 浏览器模式 | `StartWeb` 为 `starling web` 提供页面和网关。浏览器用链接里的一个运行密钥签入。网关模式在一台只做网关的机器上显示 Providers、Gateway、Routing、Usage、Sessions、Plugins 和 Settings。`webHandler` 为接受 gzip 的浏览器压缩页面的文本文件和 API 的 JSON（`webCompressed`，带一个仍会回答 304 的 `-gzip` ETag）；app 自己的窗口在进程内读页面，拿到的是未压缩的。 | [`web.go`](../../internal/gui/web.go), [`gatewaymode.go`](../../internal/gui/gatewaymode.go) |
| 页面 | `app.js` 为每个视图保留一个状态对象并把它渲染成一个列表，没有框架。各视图有自己的文件：`plugins.js` 和 `context.js`，以及 plugin 在 `assets/plugins/<id>/` 下加的那些（Gateway、Sessions、Library、Routing 和 Usage 页面，见 [plugin runtime](plugin-runtime.md)）（Usage 的 Context 标签页，以及 Routing 页面在某个请求的故事下画出的 context 窗口卡片，由 `context.go` 里的 `/api/context` 供给，它给每个 agent 打分和打标；点某个部分、单元格或条目会从 `/api/context/text` 打开它的文本，也就是内存里最近 10 个请求的脱敏正文（`openCtxText`），每次渲染 100 个条目和 20000 个字符；它的 "Tuned for you" 卡片由 `tune.go` 里的 `/api/tune` 供给，见 [agent wiring](agent-wiring.md)）。字符串经过 `t()`，译文在 `i18n.js` 里，plugin 自己的在它的 `i18n.js` 里（zh、zh-TW、ja 和 de；英文文本是键也是兜底）。zh-TW 是用台湾用词的繁体中文；系统语言是 zh-TW、zh-HK、zh-MO 或任何 zh-Hant 时会选它，托盘菜单和通知也跟随（`trayLang`）。下拉菜单用 `openProtoMenu`。每个 app 菜单（`openProtoMenu`、`openRowMenu`、`openSessCombo`）都由 `placeMenu` 定位：在它的锚点下面，或者窗口下方不够时盖在它上面；两种都放不下就取空间较大的那个，高度被限制住，并在内部滚动，而菜单内部的滚动不会关掉它（#1437，`removed-groups-menu-scroll.test.cjs`）。任何浮层能去到的最高处是 `popTop`：Mac 窗口上是 `header.top` 下方 8px（那里交通灯画在页面之上），其它地方是离顶部 8px；`placeMenu`、字段选择器的 `placePop`（它也把 `#pop` 的高度限制在那个空间里）和 agent 的模型列表（`openAgentModels`）都用它（#1573，`popover-mac-chrome.test.cjs`）。滚动只由 `scrollOnPurpose(e)` 在读者事件在手时完成。 | [`assets/app.js`](../../internal/gui/assets/app.js), [`assets/i18n.js`](../../internal/gui/assets/i18n.js) |
| 窗口导航 | 在宽 1024 点的窗口里（`SIDE_AT`，和 app.css 里 `side-nav` 规则的宽度一样），`layoutNav` 把品牌、页面（`#nav`）和 Settings 移进侧边栏（`#side`）：页面分成四组（Work、Models、Insights、Extend；一个组里的页面一个都不显示时——比如网关模式——它的名字就不出现），网关地址在它底部（`sideGwRead`，只在有侧边栏时才从 `/api/providers` 读），而页头则放一条面包屑（`renderCrumb`）。一个头部只装自己工具的页面（Library、Plugins、Settings）把它们放在标题栏里、面包屑之后，`tools` 设定它和窗口自己的按钮之间的空间。放不下时这些工具只显示图标（`tools-tight`；每个仍保留自己的名字作为 `aria-label` 和提示文字），接着面包屑丢掉它前面的部分，最后页面的名字从末尾被截断，所以没有工具会压在标题上（#2625，`title-bar-tools.test.cjs`）。打开页面自己的标签（`NAV_SUBS`：Usage、Library、Plugins、Settings）在它下面列成 `.nav-s` 行（`role="tab"`，`data-of` 是那个页面标签的 id），而页面上的那一条被隐藏；点某一行会立刻点到那个标签，选中态（`.nav-hl`）在按下那一行时就已经移开（`navSubGlide`；如果按下时离开了那一行就退回），而它是这个列表唯一的形状：悬停或按下的行没有自己的底色，而 `MutationObserver` 会在那一条变化时重画列表。同一页面的标签共用一个滚动，所以挑另一个标签的行会把它开在顶部（带着行的点击调用 `scrollOnPurpose`）；已经打开的标签再挑一次，页面留在原处（`sidebar-tab-top.test.cjs`）。列表用 iOS 那样的一弹一合来开合（`navSubEnter`、`navSubLeave`；减弱动效下都没有），而正在合拢的列表立刻就够不着了。再窄一点，侧边栏旁边的页面会比 Usage 的请求表所需（768 点）还窄，所以它们退回头部那一行，就像 Mac 的侧边栏在窄窗口里让位一样。在任何宽度上，读者都可以从标题栏的按钮（`#sideToggle`，⌃⌘S 或 Ctrl+Shift+S；`toggleSideNav`）里选侧边栏或头部标签，保存在 localStorage 的 `starling.navLayout`（`side` 或 `top`）里，所以从第一帧起就是这个布局；没选过时侧边栏按 `SIDE_AT` 来。在窄窗口里选了侧边栏时它更窄（860 以下是 184 点），而在手机上的 starling web 保留自己的标签栏，没有那个按钮。两种布局在一次视图过渡里交换位置（`navMorph`，app.css 的 `.nav-morph`）：侧边栏滑走，每个页面的标签飞向另一个里的位置，而页面作为整体滑过去（它旧的图像被丢弃，所以不会有哪一行被打印两次）；没有视图过渡、减弱动效下，或在 Linux app 里，这个变化是立即的：关闭合成的 WebKitGTK（`WEBKIT_DISABLE_COMPOSITING_MODE`，`webkit_linux.go` 在 Wayland 下会设它）在视图过渡里会崩。别处，一个过了 `MORPH_MAX`（1.5s）还在进行的过渡会被跳过，所以一个永远结束不了的过渡不会把窗口留在旧布局的图上。见 `side-toggle.test.cjs`。窗口被调宽或调窄时，会丢掉某次点击按住的东西（`heldWidth`）：重新布局的页面不是那次点击造成的。在页面重新布局的那一秒里，读者自己的位置反而被保留（`relaid`，一个挂在打开页面各部分上的 `ResizeObserver` 调用 `backToReader`）：跨过 `SIDE_AT` 时浏览器会把滚动夹到它经过的每一种布局上，而页面会停在离底部还差一点的地方（#1672，`text-size.test.cjs`）。见 `sidebar.test.cjs`。比短窗口还长的侧边栏不会在它的页面旁留滚动条：`.side-scroll` 和 `.side-foot` 的原生滚动条被隐藏，而 `sideBars` 在每个盒子的边缘上铺一条细滚动条，滚动时出现，停下大约一秒后淡出（悬停或拖动时保留），可以拖动或点击翻页，而盒子放得下时它就消失，在 WebKit、WebView2 和 WebKitGTK 里都一样。见 `sidebar-overlay-scrollbar.test.cjs`。窗口开在读者上次离开的那一页上，就像 Mac app 回来时还是原样（molvqingtai 和 lc 报告）：`show` 把页面记在 localStorage 的 `starling.lastView` 里（app 的 webview 跨重启保留它，就像保留 `starling.navLayout` 一样），而 `keptView` 在启动时把它打开。地址里要求的页面（`?view=`、托盘的链接、一次重载）优先，Agents 也是；记住的名字已经不是一个页面时开第一页，在网关模式下没记住的则开 Providers（`show`）。开在它上面不会滚动任何东西。见 `last-page.test.cjs`。 | [`assets/app.js`](../../internal/gui/assets/app.js), [`assets/app.css`](../../internal/gui/assets/app.css) |
| Settings › Privacy | 页面开在从厂商那里保留下来被脱敏的东西上（app.js 里的 `drawMaskObs`，由 `/api/masking/summary` 供给）：7 天或 30 天里被脱敏的值，配一张按天的图、各个种类、只有本地模型发现的那些、agent 到 provider 的对应，以及被脱敏最多的值，每一个只在被指向时显示（`maskTip`、`/api/masking/value`，它会说重启之后就找不到了）。它的记录（`drawMaskRecs`、`/api/masking/records` 和 `/api/masking/record`）占据页面的位置，可以按种类、查找者、agent 或 provider 过滤，一次一天，每条都打开厂商看到的那段文本并标出它的占位符。Gateway 页面的最近调用和 Routing 页面的日志上的 `maskedBadge` 会打开某个请求自己的记录（`openMaskRecord`），用琥珀色，当找到的每个值都是 Warn 时计数按原样发出。卡片下面是 *What to mask*、*How it's found*（规则，只要还有东西被脱敏就一直开着，本地模型在它们之上）、*When a value is found*（按种类是 Mask、Warn 还是 Block，秘密类排前面，个人的那些藏在 *Show all kinds* 后面，还有 *Never mask* 列表；`renderMaskActions`）、*Try it*（`renderMaskTry`，用 `/api/masking/preview` 给出厂商会拿到的文本，不存在任何存储里）和 *Sharing*。记录了什么见 [gateway routing](gateway-routing.md)。 | [`assets/app.js`](../../internal/gui/assets/app.js), [`masking.go`](../../internal/gui/masking.go) |
| 内置 plugin | Plugins › Built-in 是页面的第一个标签（`plugins.js` 里的 `drawBuiltin`，行由 app.js 里的 `renderBuiltins` 画出），列出星鸣的内置 plugin，每一个带着名字、一行说明和一个开关：Routing groups、Library、Observability、Sessions、Warm-up and check-in，以及 Gateway 和 Routing 页面（Essential）。开着且有设置的 plugin 会在它那一行下面有各自的行，由内核的 `ui.form` 按它的 schema 画出，改动会 POST `{id, config}`。Gateway 页面由它 plugin 自己的脚本通过窗口的内核画出（`kernel.js`，`assets/plugins/gateway/`）；见 [plugin runtime](plugin-runtime.md) 的 "The window's kernel"。它们是注册在 [plugin runtime](plugin-runtime.md) 上的 `plugins/*`；`GET /api/plugins/builtin` 会说明每一个（`builtinJSON`：开着是生效的、用户自己的选择、它有什么在用），而 `POST {id, on}`（或 `{id, config}`）写入 plugins.yml。页面用 `data-feature="<id>"` 标记跟着某个 plugin 的东西。关掉时 `featureStyle` 把所有这样标记的都隐藏（之后才画出来的行也算），`navHides` 把它的页面标签从侧边栏和头部都去掉，`show` 为它保留的地址打开 Agents，而 `setSetTab`、`partShows` 和 `findSettings` 把它的 Settings 标签和那些行也去掉，搜索里同样不出现。在用的那个会一直开着：它的开关被禁用，那一行会说明什么在用它，而那个 POST 拒绝把它关掉。解析不过的 plugins.yml 会在列表上方说明，没有开关会动。`boot.js` 在第一次绘制前说明哪些是关着的（`featuresOff`）和它们的设置（`pluginConfig`），而 `settings.builtins` 在每次状态读取里带着这个列表。TUI 不跟随它们。见 `builtins.test.cjs`、`builtins_test.go`。 | [`builtins.go`](../../internal/gui/builtins.go), [`assets/plugins.js`](../../internal/gui/assets/plugins.js), [`assets/app.js`](../../internal/gui/assets/app.js) |
| 无 GUI 构建 | `-tags nogui` 构建不带 GUI 的星鸣：`hasGUI` 是 false，`runGUI` 会让人用 `starling tui`。`internal/gui` 里需要 Wails 的文件在 `!nogui` 下才构建。 | [`gui_off.go`](../../internal/cli/gui_off.go) |

## 运行时路径

<a id="runtime-path"></a>

1. 不带参数的 `starling`、`starling app` 或 `starling tray` 会调用 `runGUI`，后者调用 `gui.Run`。`nogui` 构建改跑 TUI。第二次启动会交给已经在跑的那个实例。在 Windows 上，在造出任何窗口之前，`webView2Check` 会问 Wails WebView2 运行时在不在（`Env.Info().PlatformInfo["WebView2"]`，查不到时是 ""；Wails 没设的键算作在）。没有它，Wails 会在第一个窗口就退出，屏幕上什么都不剩，所以星鸣会记下这件事，用读者的语言弹一个 Windows 消息框，点名这个运行时、微软的下载页（`webView2Page`）和 `starling web` / `starling tui`，选 Yes 就打开那个页面，然后 `Run` 结束（#1806，`webview2.go`，Windows 上的 `TestWebView2MissingFromWails`）。这个框不能由 `ErrorHandler` 来弹，它跑在 Wails 创建窗口的那个线程上。
2. `startBackend` 启动网关，或者找到另一个在服务它的星鸣。
3. 面板和窗口从 `Handler` 加载 `index.html`。`boot.js` 在第一次绘制前把保存的语言、主题、色盲模式、文字大小和桌面字体选择传进去。
4. 页面读 `GET /api/state`，通过 `POST` 路由写入。每次写入都回答新的状态，在 `held` 下构建。
5. 关闭窗口会把它隐藏。在 Mac 上，全屏窗口会先退出全屏（`closeStep`）。系统对面板的关闭（标题栏的 X、Alt+F4、窗口管理器的关闭键）像 Escape 那样隐藏它，所以托盘图标又能把它打开；被放开轻量模式、不再是 `h.panel` 的面板则关闭（`makePanel`、`TestPanelCloseHidesIt`）。轻量模式只放开以这种方式关掉的窗口：被别的 app 窗口盖住的、缩到最小化的、或在星鸣隐藏时打开的，仍然是打开的（`windowUp`，不是 Wails 的 `IsVisible`，后者是遮挡状态），窗口开着时对 Dock 也是这样，重启时也是（#1381，`TestCoveredWindowStaysUp`）。在 Mac 上，托盘图标和一个用量单元格用同一个判据关闭打开的面板（`togglePanelNow` 和 `trayCellClick` 里的 `windowOpen`，不是 Wails 的 `ToggleWindow`），所以在盖住整个屏幕的窗口下面——比如远程桌面服务器的隐私帘子——图标仍然会关掉面板，而不是把它再显示出来（#1661，`TestTrayCellClickReleasedPanel`）。
6. 窗口开在上次离开时的样子。它稳定下来的大小以及是否最大化记在 `settings.Window` 和 `settings.WindowMaximised` 里（`settle`；按机器算，见 `KeepOwn`）。最大化的窗口保留着它恢复时的大小。`makeMain` 按那个大小打开它（`openSize`）。第一次显示时，`placeMain` 在 Mac 和 Windows 上会再次最大化它（Windows 上是等页面出来之后）。在 Linux 上，`makeMain` 用 `StartState` 让它最大化。在 Windows 上，比屏幕工作区大的尺寸会被调整并居中（`fitRoom`），而那个更大的尺寸仍然保留。窗口的位置不保留。
7. 页面的头部就是窗口的标题栏。`plainTitlebar` 在 app 启动后设置一次，又在窗口重新做出轻量模式时设置一次（`madeAgain`）。在 Linux 上它藏起 GTK 的标题栏，保留阴影和缩放边缘；GTK 3 让圆角保持直角。在 Mac 上交通灯嵌进头部（`MacTitleBarHiddenInset`）。在 macOS 26 的设计里，带工具栏的窗口圆角比纯标题栏的大，所以 `insetLights`（[`titlebar_darwin.go`](../../internal/gui/titlebar_darwin.go)）把工具栏去掉，而 `StarlingLights` 把交通灯和标题栏的高度留在工具栏原本所在的地方：缩放之后、换标题之后、明暗变化之后、对比度更高或透明度更低之后，以及退出全屏之后。面板（比如文件夹选择器）仍然从标题栏下方开始。在别的 app 在前面时首次按下，仍能从那个标题栏的任意位置拖动窗口（`firstPress`），就像 AppKit 带工具栏时那样；AppKit 自己现在只在纯标题栏顶部 32pt 内这样做。AppKit 从 macOS 26 起给程序这个设计，条件是程序链接的是 macOS 26 SDK 或更晚的，而且它的 Info.plist 没有要求旧设计（`UIDesignRequiresCompatibility`）；`designOf26` 在运行时读这两样，因为 SDK 取决于构建星鸣的工具链。工具栏在旧设计里保留——那里它不改变圆角、交通灯在右边（从右往左的语言时）、或者标题栏没有按预期布局时；这一点在窗口设置时决定一次。全屏时标题栏是 AppKit 自己的。Windows 保留自己的标题栏。
8. 在 Mac 上，Settings › Show in Dock 可以让星鸣从不进 Dock（默认：不在 Dock 里，因而也不在 Command-Tab 里）、窗口开着时进（`DockWindow`）或一直进（`Dock`）；全屏窗口无论这个设置如何都加入（`inDock`、`dockOnFullscreen`）。`openMain` 在显示窗口之前先加入（`setDock` 立即在主线程上跑）。一个从 Dock 之外被激活的 app（快捷面板的焦点）随后加入 Dock，虽然它在前面，却仍排在 Command-Tab 顺序的最后，所以要加到前面就得借 Dock 激活星鸣（`dockStep`、`activateAgain`；#1381，`TestDockStep`、`TestDockJoinLeadsCommandTab`）。它这么做时不会问 NSApp 星鸣是不是活跃：Mac 会立刻把它变成前台 app，而 NSApp 只有在主线程把这件事接收进去之后才会这么说，那是在繁忙的 Mac 上加入之后（#1651）。

## 约束与失败行为

<a id="constraints-and-failure-behavior"></a>

### 额度卡片刷新

<a id="allowance-card-refreshes"></a>

一次卡片刷新只应用 API 完整卡片列表里被请求的那个账户。在那次刷新完成之前开始的整页读取，在它那个较旧的响应到达时会保留较新的卡片；其它卡片照常更新，而之后的一次整页读取可以替换或移除这张卡片。这也避免了两次不同的卡片刷新互相撤销对方的结果。如果一次刷新先给出账户的名字，它那个 provider 的卡片会被替换，而更新的卡片刷新和其它 provider 保持不变。被替换的键也受保护：较旧的整页响应不能恢复一张没有名字的卡片，也不能抹掉它那张刚有了名字的替代品。只有一次完成的整页读取才会让用不到的顺序标记退役。远端星鸣的汇总同样可以变成它带命名空间的子卡片；刷新它会保留本地的 provider 和其它远端星鸣。 这个发布顺序由 [`app.js`](../../internal/gui/assets/app.js) 里的 `keepRefreshedQuotas` 和 `quotaRefresh` 掌管；`quota-refresh-order.test.cjs` 覆盖被挂起的响应和失败，`usage-card-refresh.test.cjs` 在两个浏览器引擎和五种语言里检查真正的按钮。缓存的 `readDeferred` 卡片会说明这次没有请求刷新；读失败的会保留它的失败提示。托盘面板的 Allowances 卡片用同一个 `quotaRefresh`，位置是账户下面那一行的开头（重置截止时间、错误，或某个余额的名字），在卡片被悬停或按钮被聚焦时打开。那一行会点名截止时间（"Resets 14:26"），所以不会只用一个光秃秃的 `↻` 来表示；`renderPanelQuota` 在重画之后把焦点放回按钮上。`panel-quota-refresh.test.cjs` 在两个引擎、五种语言、440 和 360px 下都覆盖了它。 重画卡片会保留它们的刷新控件和任何正在跑的回合。应答会立刻更新读数；图标随后把当前这一整圈转完才停下（减弱动效下不转）。


### 流式输出速度

<a id="streamed-output-speed"></a>

Usage 的行和汇总用 [`internal/usage/usage.go`](../../internal/usage/usage.go) 里的 `Record.Decode` 和 `DecodeOf`。Routing 用 [`plugins/routing/client.js`](../../internal/gui/assets/plugins/routing/client.js) 里的 `decodeOf`；Usage 的行用 [`plugins/usage/page.js`](../../internal/gui/assets/plugins/usage/page.js) 里的 `ledDecode`。

当答案跟在推理后面时，只从第一段答案文本开始计时答案 token。没有答案文本时，只有当网关记录到有节奏的可见推理并且有可用的内容流时，才对所有输出 token 计时。[`ttft.go`](../../internal/gateway/ttft.go) 里的 `firstToken.reasoningMs` 要求推理出现在第一个内容处，随后的推理增量至少晚 100 ms。它排除原生的 Responses 和 Gemini——它们的汇总不测量自己报告的推理 token——也排除被脱敏的思考，包括翻译过的封闭块。客户端收到翻译时，上游协议会被保留。这些是接收侧的计时检查，不是去读厂商的生成时钟：一个把缓冲好的应答匀速放出来的中继看上去仍然像有节奏。没有上游保证时，这只是一个观测到的输出速率估计。汇总仍然被排除，而不是把它可见部分的时间扩展到一个没被观测到的推理 token 数上。

`Record.FirstReasoning`（`first_reasoning_ms`）和对应的 `Call`、`Route`、`Try` 字段（`firstReasoning`）用和它们的耗时同样的来源记录这份证据。历史记录没有证据，保留它们原来的速度。请求页面的缓存和压缩快照保留这个可选字段；打包的行把 `ReasoningHidden` 放在它自己的 `Flags` 位里（`rowReasoningHidden`），和定价的 `swapKnown` 分开——后者每个有价格的行都带着。`ReasoningHidden`（账本里的 `reasoning_hidden`、trace 里的 `reasoningHidden`）会压掉没有文字的速度，哪怕被脱敏/加密的思考没有报告推理 token 数也压掉。Responses 的汇总会这样标记，在它自己的登录路径上也会，在端点回退之后也会——那时这次尝试的协议仍然写着第一个端点，而标记来自作出应答的上游（[`continuation.go`](../../internal/gateway/continuation.go) 里的 `summaryUpstream`）。可选的 OTel 属性带着这份证据而不带响应文本。`DecodeOf` 里已有的突发和保留内容上限照常生效，被隐藏的推理不进入平均速度。

一个流式请求的耗时（`Millis`，它结束每一个速度窗口）结束于它最后一次尝试的应答结束之时，而不是 `serve` 返回之时：网关在应答之后写下的东西（`answered` 的粘滞记录，`NoteServed` 里 provider 最后的答案）要落盘，而在繁忙的磁盘上要花上几秒，这会让有节奏的应答看起来像是被压住，读不出速度（#1939，`TestSpeedEndsWithTheReply`）。一个决策组的行计时方式相同。

`TestVisibleReasoningSpeed`、`TestReasoningSpeedEvidence`、`TestReasoningSpeedThroughGateway`、`TestMixedRedactedReasoning`、`TestHiddenReasoningThroughTranslation`、`TestSlowReleasedReasoningIsUntimed`、`TestPricedRowKeepsReasoningEvidence`、`TestSummaryAfterEndpointFallbackStaysHidden` 和 [`reasoning-speed.test.cjs`](../../internal/gui/tests/reasoning-speed.test.cjs) 覆盖了这段计算、缓存往返、翻译过的流和两个页面。

### 路由用途筛选

<a id="routing-purpose-filter"></a>

请求标题上的用途菜单在 [`plugins/routing/client.js`](../../internal/gui/assets/plugins/routing/client.js) 里，用的是 `openProtoMenu` 的实时勾选：每勾一次就立刻显示匹配任一所选用途的请求，菜单保持打开。空选、All purposes 或 Clear filter 会恢复所有用途，而不改当天或请求/session 的视图。实时菜单在它重画时把列表滚动夹住的情况下仍保持打开；读者在菜单外滚动仍会把它关掉。选 All purposes 会关掉菜单并把键盘焦点还给用途按钮，不滚动；取消最后一个勾选则让菜单保持打开并聚焦在那个勾选框上。 Claude Code 子 agent 的回合带 Subagent 标记（`tagKind`、`Route.SubagentID`）而没有 kind，所以它留在 Unmarked 下面，和 Usage 页面上一样。计数、当前故事和回放都跟着同一份过滤后的列表。某一天没有匹配的请求时，所选的用途仍然可用。从另一个页面打开某个请求时，只有当那个请求被排除在外才会清掉筛选。Usage 的用途选择器仍是单选，发送给账本 API。见 `purpose-filter.test.cjs` 和 `routing-purpose-state.test.cjs`。


### 路由模式标记

<a id="routing-mode-chip"></a>

舞台上的标记和 [`plugins/routing/client.js`](../../internal/gui/assets/plugins/routing/client.js) 里的 Accounts and keys 标题，会说明一条路由是按什么方式路由的，但只在那个能改的地方：一个组，或者 `/api/groups` 的 `pools` 里的一个 provider（开了两个或更多账户或密钥）。在那里它们是 `.rt-go` 控件（`goTo`、`goRouting`），通过 `scrollOnPurpose` 把那个组的卡片或那个 provider 的行在 Several accounts or keys 下面带进视野、给它做标记（`rt-found`）并聚焦它的路由。只开了一个的 provider 不会说出任何路由，模式那一段会说明怎么才能有一个（`LONE`）。组进来时标题会重画（`headOf`）。见 `routing-mode-goes.test.cjs`。

### 路由密钥与账户折叠

<a id="routing-key-and-account-folds"></a>

同一个 provider 的三个或更多 API 密钥、或者三个或更多账户，在同一个模型下并排站着（同样的固定 effort、fallback 和组中组标题）时，会折成舞台上一行 `li.rt-fold`（[`plugins/routing/client.js`](../../internal/gui/assets/plugins/routing/client.js) 里的，`FOLD_AT`、`rebuild`）。它把数量和它们的状态放在一起显示：哪个是亮着的、它在做什么、有多少可用、有多少在休息（`together`）。落到被折起来的席位上的请求会飞向那个折（`shownRow`）。点它（或按 Enter / Space）会把每一个就地展开，并从折里接上线路。Accounts and keys 把一个 provider 的密钥和它的账户各自折成一行 `.rt-keys`，并给出它们的合计。

组中组标题里的一个组（`li.rt-sub`、`subNode`）用同样的方式折叠（`toggleSub`）：折起来时它说明自己路由到多少个席位以及它们一起做什么，发给下面任一座位的请求会飞向那个标题（`row.shut`），而它的线路是最后画的那条（`wiresTo`）。

密钥一开始是折起的。账户和组中组标题在装的少于 `LONG`（6）个时一开始是展开的，否则折起（Aiirobyte 报告：一组账号太多把页面撑开了）。读者展开的记在 localStorage 的 `starling.routingKeysOpen` 里，折起的记在 `starling.routingFolded` 里（`keepOpen`、`isOpen`），键是舞台上密钥用 `provider/model`、账户用 `provider/model@`，列表里是 `provider/*` 和 `provider/@`，标题是 `group/<id>`。只影响显示：网关的顺序和席位选择都不变。两个仍然是两行。见 `routing-keys-fold.test.cjs` 和 `routing-group-fold.test.cjs`。

### 路由账户窗口

<a id="routing-account-windows"></a>

Accounts and keys 里一个订阅账户的那一行，会在路由所用的那个数字下面列出它的每个窗口，连同每个何时重置（"5 hours 40% used · resets in 2h"），和 Usage 页面上读它们的方式一样（#1833）。[`plugins/routing/client.js`](../../internal/gui/assets/plugins/routing/client.js) 里的 `cardOf` 把路由的账户和 app.js 的 `quotas` 对上：是本机的一张 `subscription` 卡片（没有 `from`），它的 provider 是该账户的 `agent`（一个 plugin 的 provider id）或它的 provider，而它的 user 是该账户的 user，忽略大小写。`winsOf` 用 `familyWindows`、`quotaText` 和 `resetInText` 点名它们。不会读任何新的东西：打开 Routing 时按常规方式问 `/api/usage/quotas`（`show`，在读数缺失或超过一分钟时），绝不用 `?asked=1`。记住的那一天的列表不显示任何东西，因为那些请求衡量的读数早就没了。见 `routing-account-windows.test.cjs`。

Accounts and keys 标题旁边的刷新（`#rtReadAllowances`，#1834）会手动再读一次：`loadQuotas(true)`，就是 Usage 页面上 Refresh now 做的那种带 ask 的读取。读取在路上时它会转起来，这时再点就不会多问任何东西。页面上没有任何东西会自己再读。它只显示实时的读数，订阅账户也列在里面。见 `routing-read-allowances.test.cjs`。

在舞台上，一个账户的席位在它的提示里说明同样的窗口，写在它的 id 下面（`render`，#1835）。那一行本身保留一行：空闲时是占用最多的窗口，作答时是它正在做什么。见 `routing-seat-windows.test.cjs`。

### WSL 中保存的 agent 会话

<a id="stored-agent-sessions-in-wsl"></a>

Sessions 页面用 `sessions.ListAgent`；Usage → Sessions 通过 [`sessions.go`](../../internal/gui/sessions.go) 和 [`sessions_manage.go`](../../internal/gui/sessions_manage.go) 里的处理器用 `sessions.List` 和 `StatsFor`。`ListAgent` 也会列出一个什么都没说过的 session，好把它清掉；`List` 把这样的 session 排除在外。某一行打开的那些路由（对话、Markdown、Terminal）用 `sessions.Find` 查找那个 session。`Find` 能找到 `ListAgent` 列出的每一个 session，空的也包括，所以这样的行不会回答 "no such session"（#2273，`TestFindEmptySession`、`TestEmptySessionOpens`）。 在 Windows 上，[`internal/sessions/wsl.go`](../../internal/sessions/wsl.go) 通过 WSL 的共享路径读取跑着的发行版。停着的发行版保留它缓存的列表，不会被打开；跑着的发行版最多每 30 秒重新列一次。

对 omp，[`ompRootsIn`](../../internal/sessions/omp.go) 同时包含 `~/.omp/agent/sessions` 和 `~/.omp/profiles/<name>/agent/sessions`。子 agent 和 advisor 文件算在它们的父 session 里。某个 profile 的 session 的 Resume 命令会用 `omp --profile <name> --resume <id>` 选中它自己的 profile，通过 `wsl.exe` 恢复时也一样。Delete 和 Restore 会保留它原来的 profile 路径和产物。Windows 不会去发现由发行版的 `PI_CODING_AGENT_DIR`、XDG 变量或 `--session-dir` 选中的任意存储。


### 桌面字体

<a id="desktop-fonts"></a>

Settings → General 提供互相独立的界面字体和代码字体，每一种都带已安装字族的实际字形。`GET /api/fonts` 调用 [`fonts.List`](../../internal/fonts/fonts.go)：Windows 上是 DirectWrite，macOS 上是 CoreText，Linux 上是 Fontconfig。发现是惰性的，缓存在处理器里，由菜单刷新。它返回名字和 CSS 特征，绝不返回字体文件或路径。刷新失败会回答 503 并保留上一次的集合；一次成功的空集合和失败是两回事。`starling web` 和 `nogui` 不提供这份桌面目录。

`settings.UIFont` 和 `settings.CodeFont` 记着字族、字形名、OpenType 字重、CSS 倾斜度和宽度百分比。缺失的选择保留平台默认值。一次设置 POST 可以用显式的 null 恢复那个默认值；省略这个字段会保留它，浏览器模式的每次保存也一样。`KeepOwn` 在同步和备份恢复期间保留本机的选择。无效的特征会拒绝这次保存，不替换掉之前的设置。

[`fonts.js`](../../internal/gui/assets/fonts.js) 在启动时和通过 `applyPrefs` 时应用选择。字体改成功之后，`host.fontsChanged` 会通过 `receiveFonts` 通知两个已有的 webview，不会等另一个窗口获得焦点；被放开的轻量 webview 重建时使用启动时的设置。正在进行的本地保存会忽略这个通知，走完自己的队列。 选中的字族排在现有的平台回退栈之前。界面和代码的特征各自独立生效；显式的分量字重仍然标示强调。一个选择优先于 Omarchy 的字体，而恢复默认后又重新跟随主题。如果发现某个已保存的字形找不到了，页面保留它的名字，解释回退方式，并在该字形可用或用户另选一个之前使用默认值。一次发现错误绝不会清掉已保存的选择。浏览器自身的排版不变。

这些菜单支持搜索、键盘选择和刷新。字形先按字重排序，然后是 normal、italic 和 oblique；剩下的平局用宽度和名字来分。更换字族时会尽量保留一个匹配的字形，否则用最接近的常规字形；具名的可变实例会提供出来，但没有任意的轴控制。

### 最近调用的正文

<a id="recent-call-bodies"></a>

[`providers.go`](../../internal/gui/providers.go) 里的 `providersState` 发送最近调用的元数据，不带请求和响应正文，provider 写入之后也一样（#1521：四十次调用的正文让 `/api/providers` 变成大约 20 MB）。网关那个 40 个调用的环里，每次调用都有一个 `id`，在 `Server.record` 里编号。在 Gateway 上打开一个最近调用，会从 `GET /api/gateway/call?id=…` 取它捕获的预览（`Server.RecentCall`），响应带 `Cache-Control: no-store`。这条路由和 API 其余部分在同一个 mux 上，所以 `starling web` 的密钥守卫也管着它。正文是环里保留的那些，按请求时的样子脱敏过（`redacted`）；这条路由不会去取 S3，也不改变捕获、导出或归档。被环挤掉的调用，或者由另一个星鸣服务的网关，返回 404。页面按那个 id 给行做键（`callRow`），在它们的调用还在列表里时把取回的正文留在内存中，跨汇总刷新保留它们，并在读失败后提供重试。见 `TestRecentCallBodiesOnDemand` 和 `recent-call-bodies.test.cjs`。

Routing 页面上的一个请求会打开同样的那些正文（Fate 报告）：它故事的头部有 "Request and response"，只有按下时才会去取 `GET /api/gateway/call?route=…`（`Server.RouteCall`）。一次调用带着它所属的那个 Routing 请求（`Call.Route`，也就是 trace 的 `Route.ID`，在 `serve`、Codex 后端和被拒的标题记录它的地方设置）；路由 id 不是调用 id。当环里已经没有它时，如果那个路由指明了归档（`Route.Archive`、`call.archiveName()`，只在那次请求开了归档时才设置），故事会通过 `archivePanel` 显示请求归档里的那份，否则就说明怎么打开归档。捕获的东西不比以前多。见 `TestRoutingRequestOpensItsBodies`、`TestArchiveWholeBodies` 和 `routing-bodies.test.cjs`。

### Plugin 签到开关

<a id="plugin-check-in-switches"></a>

plugin 的 Usage 卡片上的 Auto check-in 按钮会先读已保存的 `settings.pluginCheckins[provider]`（[`app.js`](../../internal/gui/assets/app.js) 里的 `pluginCheckinOn`）。`/api/state` 返回那张表，不带 `/api/settings` 包含的派生列表 `checkinPlugins`。显式保存的 false 优先于那个列表；没有保存的选择时，以列表里解析出的厂商默认值为兜底，而没有列表时（一次状态刷新），那个 plugin 接管的内置签到的开关状态就是（`PLUGIN_CHECKIN_INHERITS`，也就是 `provider.PluginCheckinOn` 点名的那些 id；一个测试会读这两处）。这让一个保存为开的开关在一次聚焦刷新后仍然是开的，而下一次点击可以把它关掉。[`plugin-checkin.test.cjs`](../../internal/gui/tests/plugin-checkin.test.cjs) 覆盖了两种响应形状、刷新以及随后的切换。

### Plugin 选项

<a id="plugin-options"></a>

一个 plugin provider 的编辑器会显示它的 plugin 声明的选项（[`app.js`](../../internal/gui/assets/app.js) 里的 `pluginOptionsField`；见 [provider-plugins.md](provider-plugins.md#a-plugins-own-options)）。plugin 的 label、description 和 `confirm` 文本都作为文本进去，绝不作为 HTML。一次改动会留在 `draft.pluginOptions` 里直到 Save，而 Save 会在 provider 自己保存之前把每一个 POST 到 `POST /api/settings/plugin-option`。被改回原处的选项会从草稿里去掉，所以它不会被 POST，也不会让编辑器变脏。[`plugin-options.test.cjs`](../../internal/gui/tests/plugin-options.test.cjs) 覆盖了它。

### Plugin 自己的页面

<a id="a-plugins-own-page"></a>

声明了 `starling.page` 的社区 plugin 会在侧边栏底部的一个 Plugins 组里得到一个标签（`pluginpages.js`，在 `plugins.js` 之后加载），从 `GET /api/plugin-pages` 列出，Plugins 每次重新加载它的列表时也再列一次（`window.loadPluginPages`）。标签的标题是 plugin 自己的，不翻译；`renderCrumb` 和子导航的标签读的是它那个普通的 `<span>`，而内置标签用的是 `span[data-t]`。框架在页面第一次打开时创建，并填满视图（`.view.plugin-page`、`.plugin-frame`）。`:root` 上的 `data-theme`、`lang`、`class` 或 `style` 每次变化，主题和语言都会传给它。一个被关掉或删掉的页面会失去它的标签，而窗口若正停在它上面会回到 Plugins。安全模型、桥接和读取见 [plugin-runtime.md](plugin-runtime.md#p4-community-plugins-at-every-point)。[`plugin-page.test.cjs`](../../internal/gui/tests/plugin-page.test.cjs) 在两个引擎、每种语言和窄窗口下都覆盖了它。

### 共享的 UI 规则

<a id="shared-ui-rules"></a>

- Settings → General 的布局照着 macOS 系统设置来（1ooogle 在 X 上说的）：名称本身够说清的行只写名称（Appearance、Colorblind mode、Language、Text size、两种字体、托盘点击、Open at login、Show in Dock），而名称本身说不清或带有后果的行保留一行简短说明（Lightweight mode、Prevent sleep、Session terminal）。Codex 和 Claude Code 自己的设置（Codex subagents、Long conversations、background compaction、thread titles and descriptions、composer predictions、自动评审和记忆的模型）不在这里：它们在 Agents 页面上那个 agent 展开后的行里（`AGENT_PREFS`、`connectPanel`），面板里的那一行有一个 Settings 项会用一个面板把它们打开（`openAgentPrefs`），agent 的菜单也是，Settings 的搜索则用一行打开同一个面板（`#agentPrefLinks`）。每一行都是在它的各个位置之间搬动的同一个元素，所以它的处理器和状态也跟着走（the owner, 10-20；`agent-prefs.test.cjs`）。字体行的说明只在它的字体加载或失败时显示；文字大小的按键是各个选项的提示文字，而 Colorblind mode 的说明是它的 Off/On 控件的（`#colorblindSegs`）。色盲模式（`settings.Colorblind`，`:root` 上的 `data-colorblind="1"`，#2073）会把图表的系列（`--c1`…`--c7`）在明暗两种主题下都换成 Okabe–Ito 调色板，相邻系列在色相和明度上都不同。在额度曲线上，每个窗口（`data-series`，它的下标对 7 取模，只由 `drawQuotaPlot` 和 `quotaCurve` 的图例设置）在 `path.qc-line` 上有自己的一段虚线，它的图例键会显示那条线的一个样本（`.qc-sample`）来画那个点，而均匀的节奏（`line.qc-even`）是点线；曲线的提示里会说明这一点。余额曲线不设 `data-series`，保留它的实线、虚线拟合和点。Context 标签页的 "good" 等级把 `--c2` 当颜色而不是系列用，它保留蓝色（`--c1`）。语言是一个菜单按钮，显示选中的语言，在 app 的菜单里打开（`renderLangPick`）。见 `settings-general.test.cjs` 和 `quota-curve.test.cjs`。Open at login 读 `autostart.Enabled`，在 macOS 13 及以后它还会向 Background Task Management 询问星鸣的 launch agent（`SMAppService.statusForLegacyURL:`，按 plist 的 Label 而不是路径匹配）；用户在 System Settings › General › Login Items 里关掉的记录会被读成需要允许，而那一行保持 On 并用一行说明这件事，还有一个按钮能打开 Login Items（`/api/open` 允许那一个 `x-apple.systempreferences` URL）。窗口回到前台时会再读一次状态，那行说明就清掉了。在 macOS 12、没有 cgo 的构建里，或者系统问不到时，就只由 plist 决定，和以前一样（#2120，`login-allow.test.cjs`、`login_allow_test.go`）。
- Settings 的搜索 provider 选择器提供 **Off**，保存为 `searcher: "off"`。它的帮助说明只用 API 搜索，并在关闭时压掉 provider 回退和自带搜索的提示；搜索优先级那一行被隐藏，因为没有 provider 和这些 APIs 竞争。见 [gateway search behavior](gateway-routing.md#turning-provider-search-off)、[`renderSearcher`](../../internal/gui/assets/app.js)、`TestSettingsSearchProviderOff` 和 `searcher-pick.test.cjs`。
- Settings 的搜索 API 按列出的顺序被依次询问：每个 API 的那一行（连同它的 key 那一行）作为一个整体拖动，或用 Alt+↑/↓ 移动，并且有一个开关能在不删除它的 key 的情况下关掉它（#2286）。见[搜索 API 的顺序](gateway-routing.md#the-search-apis-order-and-one-switched-off)、[`app.js`](../../internal/gui/assets/app.js) 里的 `searchArrange` 和 `search-order.test.cjs`。
- Settings 的 Search APIs 新增行在每个 API 的 key 下面有一个 base URL 字段（#1580）：厂商自己的地址是它的占位符，选一个已经加过的 API 会填上它保存的 URL，所以只填新 key 也会保留它，而不带 key 的 Add 只改 URL。见[搜索 API 的 base URL](gateway-routing.md#a-search-apis-base-url)、[`renderSearch`](../../internal/gui/assets/app.js) 和 `search-api.test.cjs`。
- 每一条用户可见的字符串都有 zh、zh-TW、ja 和 de 的译文，占位符相同。`gui-zh-tw.test.cjs`、`gui-ja.test.cjs` 和 `gui-de.test.cjs` 会在缺一条时失败；`gui-zh-tw.test.cjs` 还会在某条 zh-TW 字符串里留着简体字时失败。
- 点击绝不会移动页面。代码只在读者的事件在手时才滚动视图（`scrollOnPurpose`）；`click-scroll.test.cjs` 守着这一点。点击会在读者离开视图时所在的位置生效：就在它之前发生的、不属于读者的滚动（浏览器只在下一帧、也就是点击之后才发它的 scroll 事件）会先被放回去（点击守卫里的 `backToReader`，#1554）。被点的东西会留在原地（`hold`）；点在顶部时视图仍在顶部，而在它上面进来的东西会把它往下移（#2134，`update-auto.test.cjs`），只要那东西移动的距离不到它自身高度的一半、因而仍在指针之下。上面来的更大的东西——比如在 Routing 上选一天时带进来的那个舞台——会被滚过去，而那个控件留在被点击的地方（`routing-days-fit.test.cjs`、`routing-failed-open.test.cjs`）。
- 输入同样绝不会移动页面。会随文本变高的 textarea（plugins/library/client.js 里 Library 的 `autosize`：一个集合的文本、一个 agent 自己的文本、一个 skill 的描述、一个子 agent 的字段）在高度为 auto 时测量，而装着它的东西保持它的高度，所以页面不会有一瞬间变短：变短的页面其视图的滚动会被浏览器在每次输入时夹住，输入法的每一个组字步骤也算，而守卫把它放回去时会把一个比视图还高的 textarea 拉到它的一条边上（#2636，`library-ime-scroll.test.cjs`）。
- 没有原生的 `<select>` 元素，也没有彩色的左边框条纹。状态用一个点或一个色块表示。
- Usage 页面的卡片和托盘面板的 Allowances 标签共用一个顺序，`settings.UsageOrder`（`byUsageOrder`）。在面板里，只含余额的卡片就站在那个顺序给它的位置，而挨在一起的余额共用一个 *Balances* 网格（`panelBalances`，#1510）；没有哪个被收拢到套餐后面。面板的 *Arrange*（`panelArrange`）在面板里移动各行，并且只对那个标签隐藏订阅（`settings.PanelUsageHidden`）；两者都通过 `POST /api/usage/arrange` 保存，而那个接口只改动它被发送的那个字段，Settings 的保存保留两者。隐藏只影响显示：菜单栏单元格的卡片即便被隐藏也会显示（`panelPeek`）。见 `TestUsageArrangePanelHidden`、`panel-arrange.test.cjs` 和 `panel-balance-order.test.cjs`。
- Usage 页面的 Agents 列表里的某个 client 可以从那个列表里隐藏（ikunbots 在 X 上说的）：它那一行的 "…" 按钮或右键会打开行菜单，里面有 *Hide from this list*，按 client id 存在 `settings.UsageAgentsHidden` 里，走的是同一个 `POST /api/usage/arrange`（`agentsHidden`），并由 Settings 的保存保留。账本、合计和图表仍然会算上被隐藏的 client。列表下面有一行安静的 "{n} hidden"，点开它们，每个都带 *Unhide*（`hideUsageAgent`）。见 `TestUsageArrangeAgentsHidden` 和 `usage-agents-hide.test.cjs`。
- Usage › Overview 的 Agents、Models、Accounts 和 Gateway keys 各自把它的折叠放在标题下面（#2378）：标题是一个带箭头的 `.fold` 按钮，和 Gateway 页面上一样，而点击不会移动页面（`backToReader`）。折叠状态按查看者存在 localStorage 的 `starling.usageFolded` 里（`mountUsageFolds`、`showUsageFolds`）。合计照旧不把缓存算进去，除非 token 磁贴的标签——一个弹出的按钮（`usageTokensPick`，app 菜单）——被设成 *With cache*：这时磁贴、图表和每一行的计数与占比都会加上缓存的读取和写入（`usageTokens`），而磁贴会说明缓存的计数；这个设置存在 localStorage 的 `starling.usageWithCache` 里。Requests 标签的合计本来就计入每一个 token（`allTokens`）。分段的 token 行（Requests 那一栏的 Tokens 卡片、Overview 的 token 磁贴）会在各段计数之间换行，并保持每段完整（`usageParts`、`.parts`）；只有一个计数单独就比它的卡片还宽时才会被截断。见 `usage-overview-fold.test.cjs`。
- Usage 账本里一个请求的详情会在每种宽度下说明它的 token（#1839，`ledDetail`）："{i} in · {o} out · {r} cache read · {w} cache creation"，0 就写成 0，并且说明 5 分钟和 1 小时各写了多少（当它这么说时）。什么都没花掉的失败没有 Tokens 行。见 `ledger-cache-tokens.test.cjs` 和 `usage-ledger-detail.test.cjs`。
- 星鸣自己发出的请求（路由分类器的、联网搜索的、图片描述的）在 Usage 页面上和 agent 的请求区分开（#1839）：agents 之中星鸣自己的那一行、以及账本里它的那些行都带一个"星鸣自己"的标记（`ownBadge`），某个账户、密钥或模型的那一行会从汇总的 `own` 加上 "{n} by starling itself"，而某个账户的卡片会说明这一期间有多少请求是星鸣自己发的（`ownLine`，在汇总排在额度之后时会重画）。见 `own-requests.test.cjs` 和 `TestSummaryCountsStarlingsOwnCallsApart`。
- 打开额度就是要求看它们（#1588，`app.js` 里的 `quotasOpened`）：从导航打开 Usage 页面或选中它的 Overview 标签、窗口回到前台时停在 Overview、选中托盘面板的 Allowances 标签或面板在它上面被拉起来，以及点击菜单栏的单元格，都会加载 `usage/quotas?asked=1`，和 Refresh now 一样。一次带 ask 的读取会让后端把每个账户再读一遍（`AskClaudeUsage`），所以在上一次带 ask 的读取之后 30s（`QUOTAS_OPEN_GAP`）内的打开会改发一次普通读取，它会并入路上的一次；Refresh now 总是带 ask。面板在被收起时（隐藏着，或没有焦点）加载的页面不会带 ask。定时刷新不变，也从不带 ask。见 `allowance-open-read.test.cjs`。
- 菜单栏可以在额度旁边显示星鸣自己的合计（#1984）：`settings.TrayUsages` 里的 `usage:today` 和 `usage:7d`，在同一个选择器里排在卡片之后。`trayUsageCards` 用 `usage.QueryPage(p, Filter{}, 0, 1).Sum` 把它们加起来（Requests 标签和面板缓存的账本合计，网关和 session 文件的调用合在一起），而只选合计时不问任何厂商；`trayTotalCard` 汇总每一个 token（输入、输出和缓存，如 `allTokens`），标上 `Σ` 以免被读成输入量，再叠上以 Settings 里货币计的费用，有调用没有价格时加 `+`，一个都没有价格时就不显示费用那一行。Mac 把它们画成朴素的单元格，旁边的鸟说明这是谁的。点击会打开对应时间段的 Requests 标签（`quotaView`、`?tab=requests&period=`）或面板的 Usage 标签（`panelUseFocus`）。没有 3 天的合计：Usage 页面上没有这样的窗口。见 `TestTrayTotals` 和 `tray-totals.test.cjs`。Settings › *Menu bar cost*（`settings.TrayNoCost`，默认开着，每个系统上都一样，按机器由 `KeepOwn` 决定）关掉之后就不显示钱的部分（BussanQ 报告）：`trayMoney` 会在 `trayTotalCard` 和没有卡片时的提示 `trayIdleTip` 看到它之前，把合计里的费用和它对未定价调用的计数都去掉，于是这些卡片、它们的提示和空闲时的提示只说 token 和请求。中转或订阅的余额卡片本身就是钱，会留着。见 `TestTrayNoCost` 和 `tray-cost.test.cjs`。
- 托盘的图标是那只鸟本身，而栏上别的图标是字形：Mac 的菜单栏把 `tray.png` 当作模板图像，Linux 的栏（StatusNotifierItem 的 `IconPixmap`）拿到的是同一个剪影，单色画在透明上（[`trayglyph.go`](../../internal/gui/trayglyph.go) 里的 `trayIconFor`、`linuxTrayIcon`），所以按 alpha 给图标上色的栏染的是那只鸟而不是一块方形图标（ylorn 报告）。不上色的栏会显示浅色的它，或者在 portal 的 `org.freedesktop.appearance` 的 `color-scheme` 偏好浅色时显示深色的（`traySchemeLight`），而这个变化时会重画（`watchTrayScheme`）。没有设置 `IconName`：某个宿主如果找到一个 `-symbolic` 名字并且不上色地画它，就会在深色栏上显示一只黑鸟，而且没有 pixmap 可以退回到。Windows 的托盘保留 app 的彩色图标。菜单栏里没有卡片时，提示是星鸣加上今天的合计（`trayIdleTip`：每一个 token，按今天那张卡片的提示那样拆成输入、输出和缓存，然后是费用和请求数），它随卡片的刷新而更新；托盘的菜单仍是 Open、版本和 Quit。Wails 只在 Windows 上设置提示：在 Linux 上 `setTrayTip` 设置 StatusNotifierItem 的 `ToolTip` 并发送 `NewToolTip`（[`traytip_linux.go`](../../internal/gui/traytip_linux.go)），在 Mac 上则是状态项按钮的 `toolTip`（[`trayimage_darwin.go`](../../internal/gui/trayimage_darwin.go) 里的 `trayTipNative`；NextCommit 报告：那里悬停不会显示任何东西）。见 `TestTrayIconPerOS`、`TestTrayIdleTip` 和 `TestMacTrayTipIsSet`。
- 一个窗口的 Pace 是它卡片重置行上的一个数字（fflake33 报告）：已用份额减去已经过去的那个窗口的份额，以点为单位，均匀节奏之上是 "Pace +12%"、之下是 "−8%"，在它显示还剩多少时跟着那个仪表转，而用得比均匀节奏更多时是琥珀色（`quotaPace`、`paceLine`）。窗口的起点由网关给出，即 `provider.WithWindowStarts` 里的 `startsAt`（它的重置减去它的跨度，`QuotaWindow.Bounds`）；跨度未知的窗口、没有上限的窗口，或已经过去的周期，都没有。Settings › *Menu bar pace*（`settings.TrayPace`，默认关闭，在 Windows 上隐藏，因为它的托盘不显示文字）会把同一个数字放在菜单栏的每个数字旁边以及它的提示里（`trayPace`、`QuotaWindow.Pace`）。见 `TestWindowPace`、`TestTrayUsagePace` 和 `quota-pace.test.cjs`。
- 在 Usage → Allowances 里，Trends → Cycle 会在每个账户的曲线旁边给出它自己记录下来的周期菜单（[`app.js`](../../internal/gui/assets/app.js) 里的 `quotaCurve`、`quotaRecordedCycles`）。最长的窗口（通常是周）拥有这个菜单；较短的窗口会画在选中的日期上，并带上它们各自的重置断点。`/api/usage/quotas/history?days=45` 提供所有保留的读数，包括远端卡片的。已完成的周期来自观察到的边界，而不是七天算术；提前的重置如果落在两次读数之间，就把上一个周期结束在厂商的新起点上，否则结束在第一次新读数处。每个菜单项都给出它观察到的时长，所以提前的重置不会被标成完整的七天。选中的仪表、图形、图例、提示和虚线节奏都用那个周期里最后的读数，绝不用今天的百分比。选中的区间里没有读数的窗口会直接说明，而不是显示一个当前值。本地账户的过去周期每次选择只问一次 `/api/usage/quotas/history/holds?provider=&user=&from=&to=`（`quotaCycleHoldsOf`），它不问任何厂商（#2009，NorethSea）：[`window_holds_history.go`](../../internal/usage/window_holds_history.go) 里的 `HistoricalWindowHolds` 按每个窗口在该周期里自己最后一次读数来推算它整个窗口的值，依据是从请求日志到那次读数为止，所以 5 小时那个数字是该周期最后一个 5 小时窗口，而不是它那些短窗口加起来。它的提示会点名那个窗口的起点和读数的时间。答复到来之前它显示 "Loading…"。算不出来的窗口会像当前卡片那样说 "not enough data" 并给出原因；在另一个周期被选中之后才回来的答复会被丢弃，而卡片的 Refresh 或一次带 ask 的读取会重新推算。历史仪表显示周期结束时间和采样时间；当前的账户操作、余额和策略会留到回到当前视图时才出现。各账户彼此独立，一个选择在同一页面的卡片刷新后仍然保留。Trends → 2 days 会临时显示最近的读数；Off 移除曲线；托盘始终显示它当前的周期。见 `quota-curve.test.cjs` 和[历史上限与重置推断](providers-accounts.md#constraints-and-failure-behavior)。
- Usage 页面的卡片被打包进若干列，绝不会被全都拉成一行那么高（`packUsage`，#860）：按那个顺序，每张卡片进入结束位置最高的那个最左边的列，放在 1px 的网格行上，所以矮卡片会挨着高卡片堆起来。然后每一列的最后一张卡片会向下伸到最高那一列的结束处（一个 `min-height`，在重新测量卡片之前清掉），所以并排在一行里的卡片彼此一样高，而一摞卡片末尾会有那张高卡片陪着（ikunbots 在 X 上说的）；内容保持在每张卡片的顶部。页面的顺序（DOM、拖动、Alt+方向键）仍然是 `UsageOrder`。列数是能放下的最多的 260px 列，当那会留下多得多的空白时就少一列；只有当卡片变了或宽度能放下另一个列数时才重新选择，所以一张被展开或折起的卡片留在原处，即便滚动条这时进来。见 `usage-pack.test.cjs`。
- Providers 页面上方有一个覆盖列表的搜索框（#1614，[`app.js`](../../internal/gui/assets/app.js) 里的 `findProviders`）：输入的词会保留那些文字或数据里全都提到它们的行，开着的和关着的都算：名字、id、preset、地址、每个模型的 id 和名字、账户的 user 和 plan 以及每个登录的、账户名和 key 名。搜索时 Turned off 的折叠保持展开；按 Escape 或清空搜索框会显示每一行，折叠也恢复原样。一个开着的行可以通过拖动它的 logo、Alt+↑/↓、Alt+Shift+↑/↓（最上、最下）或它的右键菜单（`openProviderMenu`：Move to top、Move up、Move down、Move to bottom）重新排序。搜索时，一次移动或一次放下会把它在显示出来的行之间移动（它占据它越过的那个行的位置，`stepProvider`、`shownToOn`），而隐藏的行保持各自的顺序；Move to top 和 Move to bottom 会到整个列表的两端。每次移动都通过 `POST /api/providers/arrange` 保存整个顺序。见 `provider-find.test.cjs` 和 `provider-order.test.cjs`。
- 新用户的 Agents 页面（窗口的第一页和面板的 Agents 标签）在欢迎语的位置上、agents 上方有一张 *Get started* 卡片（wangye1644294 在 X 上说的，[`app.js`](../../internal/gui/assets/app.js) 里的 `guideCard`）。它的三个步骤是：添加一个模型来源；连接一个 agent，其中星鸣找到的、可以连接的 agent 会被点名，每一个一点就连上（`agents/connect`）；以及可选地建一个路由组，最无害的入门模板可以用用户自己的模型做出来（`provider.Templates`，Never stuck 排第一），通过 `POST /api/groups/template` 一点就加上。每一个勾都来自状态：星鸣提供的模型（`/api/state` 里的 `served`，不含组）、一个已接线的 agent，或者一个不是自动的组。没有一个勾是因为这张卡片被看过。前两步做完，卡片就消失。它的关闭按钮通过 `POST /api/settings/guide` 保存 `settings.GuideHidden`，所以面板、窗口和一次重载说法一致，Settings 的保存也会保留它。见 `getting-started.test.cjs`。
- 窗口的每一个页面都保持同一个宽度，最多 1200px 的内容居中，在窄窗口里留 16px 的边距（每个 `#view-…` 上的 `padding: max(16px, calc(50% - 600px))`，#860）。Routing、Usage 和 Sessions 也在这个规则里（面条报告），所以大屏幕不会把它们拉满到边，而其它页面则待在中间；它们的列会跟着分到的空间走（`byWidth`）。见 `page-width.test.cjs`。
- 页面必须能在 Chromium（Windows 的 WebView2）和 WebKit（macOS，以及 Linux 上的 WebKitGTK）里工作。GUI 测试在两个引擎里都跑。
- 在 macOS 26 的设计里，主窗口有纯标题栏窗口那样的圆角，而它的交通灯留在工具栏原本所在的地方。`TestMainWindowCorners`（[`titlebar_darwin_test.go`](../../internal/gui/titlebar_darwin_test.go)）在它自己的进程里、用 `windowshape` 这个包测量这一点：换过标题和尺寸、换过系统外观之后（按 AppKit 画那种外观的方式），在别的 app 在前面时首次按下会从哪里开始拖动窗口，以及面板从哪里开始。它把那个进程分别按链接到 macOS 26 SDK 和 macOS 15 SDK 各跑一次，不管工具链本身是哪个 SDK（`runAppKitSDK` 和 `stampSDK` 给出它运行的那份所链接的 SDK，并重新签名）。在旧设计里、在更老的 macOS 上或用 macOS 15 的 SDK 时，它检查窗口是否保留工具栏；`TestMainWindowRightToLeft` 检查保留工具栏时交通灯是否在右边。两者都在 CI 的 macOS 任务里跑。没有测试覆盖 `UIDesignRequiresCompatibility`：AppKit 是从一个 app bundle 的 Info.plist 里读它的，而 LaunchServices 会注册每一个运行过的 bundle，从这样一个 bundle 起来的进程内部发出的按下并不总会开始拖动窗口。
- 在 Linux 上，面板没有边框，主窗口有隐藏的标题栏（`plainTitlebar`）。GTK 3 会向 KDE 的 Wayland 会话（KWin 的装饰协议）要一个由 KWin 画的边框，用在所有没有客户端侧装饰的窗口上，无装饰的窗口也算，所以 `ownFrame` 告诉 KWin 面板在每次 realize 时自己画边框（#1283）。GTK 4 的构建自己就不要。没有 KWin 标题栏时，而且当 Wayland 窗口由合成器而不是由图标定位时，面板的头部在 Linux 上就是它的拖动手柄，和窗口的一样（`app.js` 里的 `--wails-draggable`；`panel-drag-linux.test.cjs`）；在 Mac 和 Windows 上面板不会被拖动（#1430）。
- 在 Linux 上，面板是一个实心窗口，和主窗口一样，它的页面把 `--bg` 整片画满（`panelOptions`、`app.css` 里的 `body.linux.panel`）。Mac 和 Windows 的面板保持半透明。一个能看穿的面板是透明 webview 下一个 RGBA 的、可由 app 绘制的 GTK 窗口，而在 KDE Plasma 6 上它那个以 93% 绘制的页面在滚动时会留下一份更淡的旧滚动位置的残影（Wayland 和 XWayland，合成模式开或关都是），而主窗口从来不会（#1803，`TestPanelSolidOnLinux`、`panel-linux-opaque.test.cjs`）。在 WebKitGTK 2.54 上带 picom 的 Xvfb 两种方式都看不到那个残影，所以这个修复的依据是报告者的机器。
- Wayland 窗口没法给自己定位，所以在 Hyprland 和 niri 下要请合成器把面板放到栏下面。在 Hyprland 下 `placePanel` 给它一条位于光标处的窗口规则（`omarchy.PlacePanel`），并在点击外面时关掉它（`omarchy.ReportClicks`）。那条规则点名光标所在的显示器，并从那个显示器的角落把面板移到 x、y（`omarchy.PanelAt`）：Hyprland 会把窗口所在显示器的位置加到一个 `move` 上，所以一个横跨整个布局的位置会把面板放到某台显示器之外而不是 0,0（#1785，`TestPanelAtMonitorOffScreenOrigin`）。`omarchy.ClickedOutside` 拿来比较的光标位置和客户端位置，都是横跨整个布局来算的。niri（`omarchy.Niri`，设置了 `NIRI_SOCKET`）在运行时不接受规则，也不报告光标。在那里 `omarchy.NiriPanel` 监听 niri 的事件流，每次面板打开就移动它：星鸣那个 pid 的、标题为 `panelTitle` 的窗口，用 `MoveFloatingWindow` 移动（如果是平铺的，先 `MoveWindowToFloating`）。它落在工作区顶部下方 6px 处，水平居中于托盘宿主传给 `Activate(x, y)` 的那个 x。Wails 在它的 Linux 托盘里保留着那个 x（`trayClick`、`TestTrayClickIsWails`），而这个 x 会横跨整个栏，因为 Wayland 的栏只知道这么多。没有点击时（`starling panel`），或点击落在输出之外时，它会到右上角。Wails 的 `HideOnFocusLost` 会关掉它，因为点另一个窗口会拿走 niri 的焦点。它仍然在中间打开并向上滑，而位于底部或侧边的栏不在考虑之内（#1670）。
- Wails 既会在点击图标时（`Activate`）调用托盘的点击处理器，也会在托盘宿主打开图标菜单时调用它（dbusmenu 的 "opened" 事件）。KDE 在右键时打开菜单，而那次点击打开的面板在 Wayland 上拿走了焦点并关掉了菜单。在 Linux 上，当这次调用来自菜单打开时，处理器会忽略它（[`trayclick.go`](../../internal/gui/trayclick.go) 里的 `menuOpening`，它在调用者里寻找 Wails 的 `(*linuxSystemTray).Event`；如果 Wails 改名或不再在那里调用处理器，`TestTrayMenuOpenFrameIsWails` 会失败）（#1430）。
- 在 GTK 3 的构建上，app 会在第一个 webview 之前给 GDK 一个字体 DPI，如果它自己没有的话（`fontDPI`，[`fontdpi_gtk3.h`](../../internal/gui/fontdpi_gtk3.h)）。GDK 的未设置值是 -1，而 WebKitGTK 2.54 的 GTK 3 设置读取器会不加检查地把它乘以 102，于是每个页面都以负宽度和 9000000px 的字体布局（#1371）。未设置的 DPI 取 `gtk-xft-dpi`，或 96，再乘上一个有效的 `GDK_DPI_SCALE`；正数则原样保留。Wails 只在 `Run` 内部初始化 GTK（它的 GtkApplication），并且在主循环上用一个默认空闲优先级的回调创建每个窗口，所以紧接在 `Run` 之前调用的 `fontDPI` 会把这项检查排成一个 `G_PRIORITY_HIGH` 的空闲任务：它在 GTK 自己启动之后、第一个窗口之前，在主线程上运行。`TestFontDPIUnset` 通过 `starling_font_dpi` 打开一个真实的 WebKitGTK 页面，需要一个显示器（Xvfb）。
- 测试绝不碰真实的 agent 配置。这个包的 `TestMain` 在 `testenv` 自己的 home 下运行。一个会显示窗口的 macOS 测试通过 `runAppKit` 在它自己的进程里运行它们（[`appkit_darwin_test.go`](../../internal/gui/appkit_darwin_test.go)）。AppKit 和 WebKit 的 home 取自账户而不是 HOME，所以 `runAppKit` 用 `CFFIXED_USER_HOME` 给它们一个临时的；外观、对比度和语言仍然来自那个账户。当那个进程在真实的 `~/Library/WebKit` 或 `~/Library/Caches` 里留下一个文件夹时，测试会失败。Playwright 测试用隔离的 `/api` 夹具提供 `assets/`。
- `POST /api/settings/codex-auto-review` 接受空值（Codex 自己的选择），或者 `provider.Served` 里精确的模型/组 ID，包括未列出的 provider。已知的 provider 配上未知或空的模型名会被拒绝，而不会改动保存的评审模型或目录标签。网关请求解析仍然是宽松的；它不是这个设置的校验者。见 [`api.go`](../../internal/gui/api.go) 里的 `Handler` 和 [`codex_auto_review_test.go`](../../internal/gui/codex_auto_review_test.go) 里的 `TestCodexAutoReviewSetting`。
- `Settings -> Sync and backup` 和 `/api/davsync/{action}` 配置并运行加密的 WebDAV、S3 或 GitHub 配置同步。它们共用 CLI 的那些操作，只暴露凭据是否存在，绝不暴露密钥本身。状态、合并、失败和兼容性行为见 [Sync and backup](sync-backup.md)。

编辑器刚打开时 `openModal` 会让背景变得不可交互，但确认框和菜单在重画过程中仍可交互。`closeModal` 在关闭动画一开始就释放那个背景，所以点击能在编辑器淡出时到达页面。在那段动画期间打开编辑器会再次保护背景。`ux-safety.test.cjs` 检查这些过渡。Routing 的 New group 按钮共用 `newGroup`，它会在替换一个未保存的组草稿之前先问；Cancel 保留草稿，Discard 打开新的编辑器。`newGroupWith` 在从一个模型创建组时施加同样的守卫，并等 `show` 接受这次导航之后才创建它的草稿。

搜索有结果时，Plugins 页面会更新 npm 的搜索结果那一节，保留推荐的 plugin 卡片不动，这样在结果到来之前就开始的点击仍会打开想要的那个 plugin。`plugin-market.test.cjs` 把搜索响应卡在 pointerdown 和 pointerup 之间，在 Chromium 和 WebKit 里检查这一点。

### 由提供网关服务的那个 Starling 展示的路由

<a id="routing-shown-by-the-starling-that-serves-the-gateway"></a>

本机上那个不提供网关服务的 Starling，会在它的 Routing 页面上显示正在服务的那个的路由，而不是那条"这里看不到它的路由"的提示（#2606）。触发条件是 `served.Load()` 为 nil，而网关地址上有一个 Starling 在应答。接着 `/api/gateway/trace` 通过回环读取正在服务的那个 Starling 的 `GET /v1/starling/trace`（[`OwnerTrace`](../../internal/gateway/trace.go)），并把它当作自己的来回答（`mine` 为 true，所以页面会播放它），而 `/api/lanes` 读取 `GET /v1/starling/concurrency`（[`OwnerLanes`](../../internal/gateway/trace.go)），这样 provider 的那些行会显示同样的药丸。两者都只经回环读取（`ownerURL`）：另一个网络接口或另一台主机上的网关地址绝不会被问到，不管 `STARLING_ADDR` 怎么写，而一个监听所有地址的则在 127.0.0.1 上读取。`OwnerTrace` 允许一个卡住的 Starling 超过 25 秒的长轮询再等 5 秒，或者在页面停止询问时结束；`OwnerLanes` 等 700ms，和 `ServedBy` 一样，而且只有当有 Starling 应答时才会被问到。状态读取带着正在服务的那个的 `serves` 和 `pid`（[`gatewayJSON`](../../internal/gui/providers.go)），trace 则把它们作为 `by` 带着。Routing 页面的标题从 `by` 里点名那个 Starling（"serve (pid 49217) is routing · 127.0.0.1:3525"），所以名字是跟着 trace 走的：当页面的 trace 来自另一个和之前不同的 Starling 时（无论哪种方向的接管），它会丢掉自己原有的那些路由，去读新的那个的整份 trace。什么都不写；提供服务和接管都不变。提供网关服务的那个 Starling 保留自己的 trace，而比那个端点更旧的会像以前一样保留那条提示（一次被拒绝或读不到的读取属于未知，绝不是空：`OwnerTrace` 用它的第二个结果这么说）。见 [`trace_peer_test.go`](../../internal/gui/trace_peer_test.go) 里的 `TestConsoleShowsTheServingStarlingsRouting` 和 `TestConsoleKeepsTheNoticeFromAnOlderGateway`，以及 [`trace_peer_test.go`](../../internal/gateway/trace_peer_test.go) 里的 `TestGatewayTraceIsLoopbackOnly`。

## 验证

<a id="verification"></a>

session 读取器包括 Cursor CLI 的聊天存储、Cursor 编辑器的项目 transcript 以及更老编辑器的 `cursorDiskKV` 对话。同一个 session ID 同时有项目 transcript 时，CLI 存储优先。编辑器的对话只读，在删除 API 上也是如此，而且不能被 cursor-agent 恢复。项目 transcript 不含消息时间戳或计费数据；它们现有的文件修改时间提供活动日期，而 token 和费用不会被推断出来。数据库里的气泡保留自己的时间戳。编辑器数据库每次列出会被打开成一个只读快照，由它的各个聊天共享，列完就移除。元数据变化会让缓存的汇总失效，即便项目 transcript 没有变。见 [`cursor_editor.go`](../../internal/sessions/cursor_editor.go)、[`cursor_editor_test.go`](../../internal/sessions/cursor_editor_test.go) 里的 `TestCursorEditorTranscript` 和 `TestCursorEditorLiveDatabaseAndMetadataRefresh`。

没有 token 也没有活动时间估计的消息，在概览里仍然算作 session。Cursor 和 OpenCode 的对话可以通过现有的 transcript API 打开，包括 CLI 的推理和工具调用。Kimi Code 的主 agent 也可以，从它的 `wire.jsonl` 里读：一个回合的 `turn.prompt` 以及从协议 1.4 起的循环事件（这样还在跑的回合也显示出来），以及从 kimi-cli 导入的 session 那些没有计时的 `context.append_message` 记录；Kimi Code 自己带计时的消息会重复这些，所以被跳过（[`transcript_kimi.go`](../../internal/sessions/transcript_kimi.go)、`TestKimiCodeTranscript`）。对话里的每一条回复都会点名它的模型（#2253），在它的第一部分之后、以及它在一条回复内部发生变化的地方再次说明：一部分的 `model` 是 agent 文件里它自己那一行写的名字，Claude Code 的 `message.model`（作出应答的模型；它自己的 `<synthetic>` 行没有），Codex 的 `turn_context.model`（这一回合被发往的模型），Pi 和 omp 的 `message.model`，OpenCode 的 `modelID` 或 OpenCode 2 的 `model.id`。Kimi Code（它的模型只在回复之后的 `usage.record` 上）、Cursor 和网关 transcript（它记录的是被请求的模型，不是作出应答的那个）都不点名，也绝不在它们的位置上猜。Usage 页面按请求的内容不会重复它。见 `TestTranscriptModel*` 和 `sessions-talk-model.test.cjs`。当选中的日期范围是空的但存在匹配的历史时，Usage 会提供 **All sessions**，并保留 agent、模型、文件夹和搜索筛选。见 `sessions-empty-range.test.cjs`。

```sh
go test -tags nogui ./internal/gui
go test -run TestMainWindow ./internal/gui
GOARCH=amd64 CGO_ENABLED=1 go test -run TestMainWindow ./internal/gui  # as for an Intel Mac
go test -run TestWebView2 ./internal/gui   # on Windows: Wails' runtime lookup, missing and present
go test -v ./internal/fonts         # native discovery on each desktop OS
make test-ui                      # every internal/gui/tests/*.test.cjs, Chromium and WebKit
BROWSER=webkit node --test internal/gui/tests/click-scroll.test.cjs
```

`make test-ui` 需要 Node.js 和 Playwright（`playwright install chromium webkit`）。Playwright 装在仓库之外时要设置 `NODE_PATH`。CI 里的 Test 工作流不跑这套测试，所以改动 GUI 必须在本地跑一遍并报告结果。见 [`tests/README.md`](../../internal/gui/tests/README.md)。

`TestMainWindowCorners` 和 `TestMainWindowRightToLeft` 在有桌面会话的 Mac 上运行，就像 CI 的 macOS 任务那样，并会显示几个窗口若干秒。在 macOS 26 上它们检查两种设计，不管工具链链接的是哪个 SDK。CI 为 arm64 构建它们。为 amd64 构建时（在 Apple silicon 上通过 Rosetta），它们检查的是一个在 `LC_VERSION_MIN_MACOSX` 里写出自己 SDK 的程序，就像 go test 为 Intel Mac 链接的那个一样。它们发出的按下都在测试自己的进程内完成，不会改动任何系统设置。

### Agent 可见性管理

<a id="agent-visibility-management"></a>

Agents 页面的 Manage 按钮在列表的页脚里、折叠控件旁边，什么都没被折叠时它也一直可用。托盘面板保留它原有的折叠控件。Manage agents 对话框包含被折叠和被隐藏的 agent。搜索匹配名字和 ID；全选和反选作用于搜索结果，而 Select unconnected agents 会把选择换成匹配那些没有接在星鸣上的 agent（既没接线、没添加，也没在漂移）。选择在搜索变化后仍然保留。Hide selected 和 Unhide selected 各自通过现有的 `POST /api/agents/arrange` 保存一份排布，保留顺序和操作范围之外的 agent。只有一次成功的保存才会改变列表；失败会保留选择以便重试。保存期间，Done、Escape 和背景遮罩都不能关掉管理器、露出冲突的行操作。无论成功还是失败，关闭能力都会恢复。一次完成的保存只让更早的在途状态读取在排布字段上失效。新的读取仍会接受其它窗口的可见性改动；批量取消隐藏不会加上持久的客户端覆盖。这些竞态由 `agent-visibility-saving.test.cjs` 覆盖。 管理器打开时，状态刷新会就地更新它的可见性标签和操作可用性，保留搜索、选择和键盘焦点。这些操作绝不连接、断开或编辑某个 agent 的文件。取消隐藏仍然遵循现有的 Not set up 折叠规则。对话框会说明恢复可见性并不会把一个未配置的 agent 钉进主列表。见 `agent-bulk-visibility.test.cjs`，它在 Chromium 和 WebKit 里覆盖持久化、部分选择、失败的保存、窄布局和全部五种语言。
