# 2026-11-07 当前代码推送 GitHub

## 结果

- 用户明确要求将当前代码推送至 `https://github.com/iotwq/sub2api`。
- 目标：`origin` 的 `main` 分支，普通快进推送，没有强制覆盖远端。
- 推送前远端：`7775571f4bf513081ff651843b4553112f545554`。
- 代码提交：`618d706d47f5135dffaa3d39f8e15731661689e0`，`feat: preserve site customizations and v0.2.20 compatibility`。
- 独立读取远端 main 后确认与代码提交一致；核验时间：2026-11-07 15:27:48 CST。
- 同步此前尚未推送的提交历史，并将当前未提交的源码、测试及已有记录归入代码提交，共 367 个文件变更。上游 0.2.20 的 `6b4b37db43443b5e88e424fe3c69602cc3810dfb` 已包含在其历史中。

## 范围与验证

- 本次没有改动业务源码；提交内容与上一轮已验证的完整工作区逐项一致，暂存区的 359 个现存文件内容哈希也与验证快照一致，8 项既有删除一并提交。
- 包含 Basispoints/Excel、媒体转发、模型广场、渠道监控、近期前端调整和本次上游合并兼容修复；新增的源码和测试均已入库，本次合并清单一并跟踪。
- 沿用紧邻的合并验证：前端 416 文件 / 3416 项测试，后端完整 unit、vet、相关 race、隔离 PostgreSQL/Redis、浏览器及生产编译均通过；源码未变化，本次不重复业务测试。
- 暂存内容未包含本地环境配置、私钥或依赖/构建目录；未发现常见真实凭据标记。空白检查仅报告此前已有的 `docs/CHANNEL_MONITOR.md:100` 文件尾空行，未为发布改动无关内容。
- 根目录 `app-initial~app-main~onboarding-page~hotkey-window-thread-page~quick-chat-window-page~chatg~k0ede4gb-C17KDkOa.js` 和 `pelican-bike.html` 是此前保留的无关临时文件，未提交、未删除。其他忽略文件沿用现有规则。
- 本次只推送 Git，没有制作/推送镜像、操作生产部署或执行生产数据库迁移。新增 MCP 默认关闭及 249 迁移边界仍见 `upstream-sync-20261107-files.md`。
- 本文件和本次 progress 条目以单独记录提交跟随代码提交推送，不改变已验证的源码。

## 证据与恢复

证据目录：`/var/folders/p3/s4y4scsx14s5wr1r7693n_fw0000gn/T/sub2api-github-push-20261107-fuaxbm7k`。保存推送前 HEAD/索引/差异、暂存路径、校验结果、代码推送回执和远端核验；后续记录提交的回执也保存在此目录。

推送前远端恢复点为 `7775571f4bf513081ff651843b4553112f545554`。如需检查原远端代码，可以在独立目录恢复，不覆盖当前工作区或远端：

```sh
restore_dir=$(mktemp -d /tmp/sub2api-github-before-20261107.XXXXXX)
git worktree add --detach "$restore_dir" 7775571f4bf513081ff651843b4553112f545554
```

推送前本机完整代码内容已经保存在 `618d706d47f5135dffaa3d39f8e15731661689e0`，上游合并前的工作区恢复方式见 `upstream-sync-20261107-files.md`。不要把远端旧快照与推送前本机完整工作区混同；恢复源码不会回滚数据库或生产服务。
