# 同步与备份

<a id="sync-and-backup"></a>

`internal/backup` 把崇云科技可移植的配置收集成一个用口令封存的包。`internal/davsync` 把这个包放在一个 WebDAV 文件夹、一个 S3 桶或一个 GitHub 仓库里。它在电脑之间合并改动；它不克隆工作树，也不同步任意文件。

## 事实来源

<a id="sources-of-truth"></a>

- [`Config`、`Configure`、`Status` 和 `syncOnce`](../../internal/davsync/davsync.go) 把当前启用的服务器记在 `sync.json` 里，把最后的本地/远端哈希、版本、通知和撤销文件记在 `sync-state.json` 里。这些文件放在崇云科技的设置旁边。密钥只留在本机，绝不会出现在状态响应里。
- [`remote`](../../internal/davsync/dav.go) 接口读一个版本并有条件地替换它。WebDAV 和 S3 用 ETag；GitHub 用它的 Contents API 要求的 blob SHA。[`github.go`](../../internal/davsync/github.go) 解析 `github://owner/repo[/prefix]`，`Config.Branch` 选择分支，或者用仓库的默认分支。
- [`backup.go`](../../internal/backup/backup.go) 负责包的内容和封存。同步把配置分成 provider、设置、profile、agent 和库。provider 密钥、agent 模型、库和用量各有自己的包含开关。
- [`usage.go`](../../internal/davsync/usage.go) 通过远端的 `files` 接口共享分别封存的用量和配额文件。
- [`restore.go`](../../internal/davsync/restore.go) 用远端包里的部分替换本地的部分，先把本地配置留作 Undo。
- [`backup.go`](../../internal/gui/backup.go)、[`app.js`](../../internal/gui/assets/app.js) 里的设置表单和 [`webdav_cli.go`](../../internal/cli/webdav_cli.go) 触达的是同一批同步操作。

## 每台电脑上保留什么

<a id="what-stays-on-each-computer"></a>

设置这一部分带着用户的偏好，但不带那些只关于一台电脑的东西。[`Settings.KeepOwn`](../../internal/settings/settings.go) 在每次同步、加入、Restore、Undo 或备份文件把设置带进来时，都把这部分从本机的设置里放回去，而 `hashes` 把它们抹平，于是改掉其中一个不构成要推送的改动。它们是：窗口的大小以及是否最大化、字体、代理、网关的端口、局域网共享（`LAN`）以及放行进来的地址（`AllowIPs`）、崇云科技可以从哪些私有主机下载（`FetchPrivate`）、Dock、网关模式、保持唤醒、轻量窗口、WSL 查询、账户发现、何时读取额度、记录网关对话、session 恢复时所在的 Mac 终端（`SessionTerminal`）以及菜单栏显示什么。开机自启是系统自己的设置（`internal/autostart`），监听地址是 `MAGPIE_ADDR`，所以两者都不在包里。

在一台电脑上打开共享，绝不会在另一台上打开或关闭它（Lullaby 报告）。把备份恢复到一台新电脑上时，共享保持那台电脑本来的状态，在全新机器上就是关闭：共享会把网关开放给网络，而用户要在那台机器上手动打开它，就像端口和代理也是在那里设置的一样。默认的共享密钥则不同：它在网关密钥库里，随着设置这一部分一起带着，键名是 `LANKey` 和 `LANKeyID`，所以每台电脑拿到的都是同一个，而在第二台电脑上打开共享时是取用它而不是另造一个（`TestLANSharingStaysOnItsComputer`）。

## 运行时与状态

<a id="runtime-and-state"></a>

网关默认每三分钟跑一次同步。`SetAuto` 可以选一个更长的间隔，或者只手动同步。在设置里保存会立刻同步。每个操作都持着同步锁，所以配置改动不会和同步的状态写入抢跑。

第一台电脑会创建一个远端备份。另一台电脑加入时取走远端的配置，并把被顶掉的本地部分留在自己的同步文件夹里。之后的每次同步都把每一部分和上一次同步的哈希做比较：只有本地改动的往上走，只有远端改动的往下走，两边都改过的保留较新的版本，同时把被替换掉的那份留下。因为中途有另一台电脑提交而遭拒的有条件写入，会通过同一套合并再试一次。

当远端文件是空的、损坏的，或是别的应用的格式时，同步会停下来，并通过 `ServerFile` 描述它。设置和后端的 `upload` 命令可以用本机的配置把它替换掉，先把旧文件保留下来。HTML 和 XML 不提供替换。GitHub 的诊断信息会点名仓库和 `chongyun github upload`，而不是 WebDAV 那条恢复路径。

一个存在但读不了的本地文件是未知，不是空，也不是默认值。`backup.Collect` 会在读不了的 providers.json（连同它的分组和顺序）、settings.json（`settings.LoadChecked`：读不了，或者解析不过且没有可用的 `.bak`）或 library.json 上停下来，也会在导出失败的网关密钥库上停下来。备份、导出和同步都会带着这个错误结束，什么都不封存、什么都不推送；同步的状态会显示它。缺失或空的 settings.json 仍然算没有（`TestCollectStopsOnUnreadableSettings`、`TestSyncStopsOnUnreadableProviders`、`TestSyncStopsOnUnreadableSettings`，#2269）。

切换服务时，所有未启用的种类都留在 `Config.Servers` 里。`Other` 保留为最近离开的那个服务器，以兼容旧的已保存配置和客户端。只有当前启用的服务器会被使用。空的凭据字段只会为它对应的账户复用已保存的凭据；GitHub 把复用限制在同一个 owner/repo 上。该仓库内改分支或改前缀可以继续用同一个 token。分支是同步状态键的一部分。

## GitHub 的限制与错误

<a id="github-constraints-and-errors"></a>

- 仓库必须存在而且是私有的：仓库读取没有给出 `"private": true` 时，`prepare` 会拒绝它，因为公开的历史等于把每一个封存版本交给任何人去猜口令（`TestGitHubPublicRepositoryRefused`）。在空仓库里，第一份加密备份通过 Contents API 初始化它的默认分支。选了默认分支以外的分支时，那个分支必须已经存在。token 需要仓库 Contents 的读和写权限。
- 备份放在 `<prefix>/magpie/magpie.magpie-backup`；可选的用量文件在它们旁边的 `usage/` 下。改动会创建 commit，而没有变化的配置不会再创建一个备份 commit。
- 这里的 token 与库的 GitHub token 是分开的，只发往 `https://api.github.com`，并且必须与加密口令不同。仓库历史里保留着更早的加密版本。
- 仓库和分支的访问权限在文件的 404 可能意味着"还没有备份"之前就先检查好。读取失败和元数据格式错误都会让同步失败。
- 只有 ref 端点明确返回 `409: Git Repository is empty.` 才允许初始化，而且是在仓库访问权限和默认分支都检查过之后。首次写入不带 `branch`，这样 GitHub 会创建它的默认分支；其他冲突和缺失的分支不当作空仓库处理。
- 大文件按 blob SHA 下载，保留被读到的那个版本。同步文件上限是 64 MiB。列出用量文件时撞上 Contents API 的 1000 个文件上限就失败，而不是把被截断的列表当作完整的。
- 过期的 SHA 或抢跑的首次创建会返回 `errChanged`。其他 422 响应仍然是可操作的错误。认证、权限和分支的失败与文件缺失仍然区分开。速率限制用同步的退避。
- 请求归档仍然是 S3 独有的功能；它的对象与加密的配置和用量文件是分开的。

## 验证

<a id="verification"></a>

```sh
go test -tags nogui ./internal/davsync
go test -tags nogui ./internal/cli -run 'TestGitHubCmd|TestWebdavCmd|TestS3Cmd'
node --test internal/gui/tests/github-sync.test.cjs internal/gui/tests/s3-sync.test.cjs internal/gui/tests/sync-other-kind.test.cjs
```

`TestGitHubEmptyRepository` 覆盖了第一份加密备份初始化一个空仓库，以及第二次同步没有变化。`TestGitHubEmptyRepositoryFailures` 覆盖了缺失的分支、无关的 ref 冲突，以及另一台电脑赢下首次写入。`TestGitHubNotBackup` 覆盖了仓库里放着一个空文件或外来文件时后端特有的恢复建议。

Go 测试用 `testenv` 的 home 和 HTTP 夹具。浏览器测试在 Chromium 和 WebKit 里用隔离的 API 夹具。一次真实的 GitHub 写入需要一个测试仓库和 token；夹具测试并不能证明某个具体账户的权限或仓库规则允许提交。
