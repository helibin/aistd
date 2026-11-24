# 版本与变更记录

依据固定工程规范 `docs/versioning.md`，CLI 和网站独立维护 SemVer：`apps/cli/VERSION` 与 `apps/web/VERSION` 是各自唯一版本来源。CLI 可执行文件、生成项目的来源信息、官网所显示的 CLI 版本及发行路径读取 CLI VERSION；网站 package.json 与锁文件版本须等于网站 VERSION。固定规范 v1.0 与生成应用自身的初始版本独立于 CLI 版本。

每次提交前，先对受影响的应用升级版本并在对应 `apps/<app>/CHANGELOG.md` 顶部添加日期和面向用户的说明。新增兼容功能升 minor，兼容修复和文档维护升 patch，破坏兼容的变更升 major，并说明迁移。共享文件改动影响两个应用，因此两个版本及 changelog 都要更新。纯 CLI 或网站目录的变更只升级对应应用。

用户可见改动在 `changes/<app>/<unique-id>.md` 记录片段，包含 `type: added|changed|fixed|deprecated|removed|security`、`breaking: true|false`。准备提交时汇总到 changelog 并标明版本，保留原片段作为归档。纯重构无需用户可见片段，但仍遵守本项目每次提交升级版本和更新 changelog 的要求。历史发布记录不可静默重写。

运行 `make setup` 启用本仓库 `.githooks/pre-commit`，检查暂存区的版本升级、带日期的 changelog 条目和网站包版本一致性。CI 对每次 push/PR 中的新提交调用同一脚本；不能把多个未升级版本的提交藏在最终版本更新之前。只应暂存当前提交的文件，禁止 `git add .`。

`make check` 与 `make build` 检查当前代码；`make release` 生成当前版本的四平台二进制、SHA256SUMS 和带版本的安装脚本。部署包从相同版本路径取文件，发布前完成安装、重复升级和失败保留旧版验收。上述步骤不自动提交、打标签或发布公网。当前没有原子构建序号分配服务，不伪造开发构建序号；首次接入记录保留历史 `0.1.0` 基线，日期明确为补录日期。
