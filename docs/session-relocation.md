# 迁移 Claude Code 的项目会话

项目文件夹被改名或移动之后，Claude Code 保存的会话仍留在其 `projects` 目录下的旧编码路径里。在 **会话** 中选中 **Claude Code**，然后在旧文件夹旁边点击 **迁移项目**。

1. 关闭 Claude Code 及其编辑器会话。等它们最后一次写入之后至少一分钟。
2. 输入新的、已存在的项目目录的绝对路径。
3. 预览会话数量、大小，以及源路径和目标路径。
4. 确认移动。记下对话框里显示的备份路径。
5. 在编辑器中重新打开新文件夹，检查它的 Claude Code 会话列表。

该操作移动整个已存储的项目，包括子代理、工具结果和项目记忆。会话 ID 和消息内容都保留。旧项目之下顶层记录的 `cwd` 字段会被映射到新路径；对其他项目和外部 worktree 的引用保持不变。已存在的 `sessions-index.json` 里的项目和记录路径会被更新。消息、工具参数和记忆文本中的历史路径不会被改写。

在 Windows 上，WSL 项目的目标路径会在同一个正在运行的发行版内解析。CLI 也可以直接在 WSL 里运行：

```sh
starling sessions relocate-claude --from /home/me/old-project --to /home/me/new-project
starling sessions relocate-claude --from /home/me/old-project --to /home/me/new-project --confirm <token-from-preview>
```

第一条命令只做预览并打印 JSON。确认令牌把这次操作绑定到那些路径和源文件上；有改动就需要重新预览。为 CLI 设置 `CLAUDE_CONFIG_DIR` 时，它会选定 CLI 使用的 Claude Code 存储位置。

## 范围与安全

- 目标 Claude 项目存储必须不存在。不支持合并项目、挑选单个会话和跨发行版移动。
- 最近修改过的文件、格式错误的记录、符号链接、有歧义的编码路径，以及过长的哈希项目名都会被拒绝。
- 迁移前关闭客户端。一分钟的写入检查发现不了仍保持着某个会话打开的空闲客户端。
- Claude Desktop 独立的存储、账号设置、全局提示词历史，以及按会话 ID 索引的全局文件都原样保留。

## 恢复

原始项目文件按字节原样保存在 `<Claude config>/starling-relocations/project-*/original` 下，位于会话扫描范围之外。`manifest.json` 记录源路径、目标路径和备份路径。暂存文件在发布之前放在该备份旁边。

要撤销一次已完成的迁移，关闭客户端，把目标项目存储挪到 `projects` 之外保留下来（尤其是对话还在继续的情况），然后把备份里的 `original` 目录移到清单的 `from` 路径。不要覆盖已存在的目录，也不要在 `projects` 下留下同一会话的两份副本。

操作被中断之后，恢复之前先检查清单以及源、目标与 `original` 目录。发布失败会恢复原始源；如果这次恢复也失败，错误信息会指明备份位置。备份会一直保留，直到你删掉它们。
