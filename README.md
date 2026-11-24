# aistd

aistd 把工程和视觉规范带进项目初始化流程，生成可以直接启动的项目骨架。CLI 支持离线生成，依赖在 `make setup` 时安装。

## 安装并运行第一个项目

macOS / Linux（amd64、arm64）：

```sh
curl -fsSL https://aistd.vlb.cn/install.sh | sh
```

默认安装到 `~/.local/bin`，请确保该目录在 `PATH` 中。重复执行可升级；卸载用 `aistd uninstall`。

安装 CLI 后，默认 API + Web 项目需要 Go 1.23+、Node.js 22、npm 和 make：

```sh
aistd create --name demo
cd demo
make setup
make run-api
```

保持 API 终端运行，在第二个终端进入同一个 `demo` 目录：

```sh
make run-web
```

打开 Vite 输出的本地地址（默认 `http://127.0.0.1:5173`），页面应显示 `Go API: hello`。API 默认监听 8080，Web 通过 Vite 代理访问 `/api/hello`。`make run` 在同时选择 API 和 Web 时只启动 API；两个服务需要分别启动。修改后运行 `make check` 和 `make build` 做检查与构建，这两个命令不会启动服务。

只选择 Web 时无需 Go，运行 `make setup` 后执行 `make run-web` 即可。其他平台的启动步骤见生成项目的 `README.md`。

## 技术栈选择

先选目标平台，再选每一类应用的技术栈。省略技术栈参数时保持原默认组合；移动端 React 指 React Native / Expo。

| 应用 | 参数 | 可选值（默认在前） |
| --- | --- | --- |
| 后端 API | `--api-stack` | `go`、`nestjs` |
| Web | `--web-stack` | `react`、`vue` |
| Admin 后台管理 | `--admin-stack` | `vben`（Naive UI）、`soybean` |
| iOS / Android | `--mobile-stack` | `react-native`、`flutter` |
| 微信 / 支付宝小程序 | `--miniapp-stack` | `taro`、`uniapp` |

```sh
aistd create --name demo --targets api,web --api-stack nestjs --web-stack vue
aistd create --name mobile-demo --targets ios,android --mobile-stack flutter
aistd create --name mini-demo --targets wechat,alipay --miniapp-stack uniapp
aistd create --name hello-aistd --targets admin --admin-stack vben
aistd create --name soybean-demo --targets api,admin --admin-stack soybean
```

交互终端运行 `aistd` 或 `aistd create`，名称默认 `hello-aistd`，API 与 Web 默认勾选；方向键移动、空格多选、回车确认，只询问已选平台的技术栈。每步显示选择摘要，最后确认目标目录与配置后才创建文件；Ctrl+C 或 Esc 可取消。管道、重定向或 `TERM` 未设置/为 `dumb` 时，`aistd create` 保留编号问答，无参数 `aistd` 显示帮助；`--name`、`--config`、`--yes` 保留原有非交互行为。iOS/Android 共用一个移动应用，微信/支付宝共用一个小程序应用。JSON 对应字段为 `apiStack`、`webStack`、`adminStack`、`mobileStack`、`miniappStack`；不与 `--config` 混用这些参数，也不接受为未选择的平台指定技术栈。

生成过程离线运行；`make setup` 才安装所选栈的依赖。Go API 需要 Go 1.23+；NestJS、React、Vue、React Native/Expo、Taro、uniapp 需要 Node 22/npm；Flutter 需要 Flutter SDK 和对应原生构建工具，不属于 npm workspace。Flutter 的 Android 构建产出调试 APK，iOS 构建为未签名应用，需要 macOS/Xcode；Expo 构建导出 JS 资源。小程序构建后仍须使用开发者工具配置真实 App ID 并完成平台审核。

技术栈选择需使用包含本功能的新 CLI；官网配置器与 CLI 应一起发行。旧 JSON 缺少技术栈字段仍可用，缺少 `web` 时继续不生成 Web；旧 CLI 无法读取新增字段。切换已有项目技术栈和自动迁移不在 `create` 范围内。

## Admin 后台管理模板

Admin 是独立于 Web 的第七个目标，默认关闭。选择 `--targets admin` 后，`--admin-stack vben` 使用 Vben Naive UI，`--admin-stack soybean` 使用 Soybean-admin；也可与 API、Web 或其他目标组合。旧 JSON 省略 `admin` 时保持关闭；Admin-only JSON 须显式设置 `api: false`，避免旧配置的 API 默认值。

CLI 内嵌基于固定上游版本的 MIT 模板：Vben 5.7.0、Soybean 2.0.2，创建时无需网络。来源、随 CLI 分发的归档 SHA256 和入口记录在生成项目的 `.aistd/admin-template.json`，许可证和集成说明保留在 `apps/admin/`。Vben 保留上游 workspace 结构，但运行、检查、构建只选择 Naive UI。两个模板都是独立 pnpm 子树，不加入根 npm workspaces。

生成后需要 Node.js 22.18+、npm 和 make。在项目根目录运行 `make setup`，通过 npx 调用固定 pnpm 10.33.4 安装依赖，无需全局安装 pnpm；随后运行 `make check`、`make build`、`make run-admin`。Vben 开发地址为 `http://localhost:5888/admin/`，Soybean 为 `http://localhost:9527/admin/`；以启动输出为准。部署静态入口使用 `/admin/`，可与 Web 的 `/` 共存。

Vben 开发使用本地 Nitro mock；Soybean 默认配置仍指向上游 Apifox 演示服务，但公开版不内置演示令牌，请在使用前对接自有后端。现有 Go/NestJS hello API 不提供后台认证或权限接口。`doctor` 只校验固定规范原文，不验证可编辑 Admin 源码。升级 CLI 不会迁移已有项目。

## Go 后端模板

生成的 API 使用 Go 标准库，按实际职责组织：`cmd/server` 负责进程入口，`internal/config` 负责运行配置，`internal/httpapi` 负责路由、请求日志、健康检查和 HTTP 服务生命周期。各包附带测试；跨端 OpenAPI 仍位于根目录 `contracts/`。

`HTTP_ADDR` 默认 `:8080`，服务与 `--healthcheck` 使用同一配置。运行时通过环境变量注入，`.env.example` 是配置说明，不会自动加载。配置错误会阻止启动；停止服务时等待在途请求，并设置退出时限。请求日志包含请求 ID、响应状态和耗时。

这些文件属于可编辑的应用模板，固定 v1.0 规范原文保留在 `.aistd/standards/`。本次模板更新只作用于新建项目，不迁移或覆盖已有项目。业务模块按需增加；数据库、鉴权及生产发布配置需按产品需求实现。

## 本地检查

`apps/cli/` 是 Go CLI，`apps/web/` 是 React + Vite 官网。维护者的构建和部署步骤见 [发行与部署](docs/releasing.md)。

```bash
npm ci
make setup
make check
make release
```

`make setup` 下载 CLI 的固定 Go 模块依赖并启用提交检查；离线验证读取已准备的 CLI 模块缓存。发行后的 CLI 创建项目仍离线运行，无需为向导安装依赖。

`make release` 分别生成 `dist/` 官网和 `apps/cli/release/v<VERSION>/` CLI 发行文件。用户不需要执行这些维护者命令。

运行 `npm run dev:web` 可在 `http://localhost:5173/` 预览官网。Vite 预览未配置公网安装地址，不提供 `/install.sh`；正式安装入口由部署容器生成。

## 安装、升级与接入说明

安装脚本支持 macOS/Linux 的 amd64/arm64，默认安装到 `~/.local/bin`。**升级时重复执行同一条命令即可，无需先卸载。** 新 CLI 下载、SHA256 校验及运行验证成功后，自动替换安装目录中的旧 CLI；失败保留旧版。符号链接、目录等异常安装目标会被拒绝。自定义目录使用 `curl -fsSL https://aistd.vlb.cn/install.sh | AISTD_INSTALL_DIR=/path/to/bin sh`，升级时沿用相同目录。CLI 使用不需要 Go。Windows 安装暂未实现。

卸载 CLI 运行 `aistd uninstall`，确认显示的可执行文件路径后输入 `y`；自动化场景可使用 `aistd uninstall --yes`。如果安装时设置了 `AISTD_INSTALL_DIR`，卸载时也要设置为同一目录，例如 `AISTD_INSTALL_DIR=/path/to/bin /path/to/bin/aistd uninstall`。命令仅接受当前运行的可执行文件位于安装目录且名称为 `aistd` 的情况，只移除该文件，不删除整个安装目录。卸载 CLI 不会删除任何通过 `create` 或 `init` 创建/接入的项目文件，包括 `.aistd/`、`task.md` 和规范原文。

安装 CLI 后，新项目直接运行 `aistd create --name hello-aistd`；默认生成 Go API、React Web、产品视觉草稿，并安装两份 v1.0 规范原文。运行 `aistd` 或 `aistd create` 可在终端向导中使用方向键和空格选择 API、Web、iOS App、Android App、微信小程序和支付宝小程序，并为选中的应用选择技术栈，查看摘要后确认创建。管道或简易终端保留编号问答。非交互模式可运行 `aistd create --name hello-aistd --targets ios,wechat --visual=false`；官网配置器会生成对应命令，JSON 配置仍可选用。iOS/Android 共用 Expo 或 Flutter 项目，微信/支付宝共用 Taro 或 uniapp 项目；这些活动模板不改写固定规范 ZIP，生成的起点不代表完成原生签名或平台审核。已有项目在根目录运行 `aistd init --dir .`，新增两份规范原文、锁文件和 `AISTD-ADOPTION.md` 接入清单；若根目录没有 `task.md`，还会创建可编辑的开发计划，已有的 `task.md` 原样保留，不覆盖项目已有文件。两条路径均可运行 `aistd doctor --dir .` 校验规范来源和逐文件 SHA256。`task.md` 用于跟踪目标、待办、进行中、阻塞、完成项及验收依据，是可编辑的项目文件，不属于 `doctor` 校验的固定规范快照。原文完整性验证不等于项目符合规范，产品视觉方向、运行与发布验收仍须单独完成。

可选接入 OpenSpec：先安装官方 CLI，再用 `aistd create --name hello-aistd --openspec` 或 `aistd init --dir . --openspec` 初始化 CodeBuddy 工作流。默认不启用，也不会覆盖已有的 `openspec/` 或 `.codebuddy/`。

## 当前限制

iOS/Android 原生签名包、小程序开发工具及平台审核仍待验证。Go API 是工程规范参考骨架，不包含数据库、生产部署或完整业务逻辑。`doctor` 仅验证规范快照完整性，不评定工程或视觉合规性。

## 许可证

aistd 自有代码使用 [MIT License](LICENSE)。内嵌规范和第三方模板不由根许可证统一授权，见 [许可范围](docs/license-scope.md)。

## 版本与提交

CLI 当前版本见 `apps/cli/VERSION`，官网版本见 `apps/web/VERSION`；各自维护 `CHANGELOG.md`。每次提交升级受影响应用的 SemVer 并更新 changelog，用户可见变更保留 `changes/<app>/` 片段。运行 `make setup` 启用暂存区提交检查，CI 对每个新提交复核。完整规则见 `docs/versioning.md`。下载目录中的 `<VERSION>` 由 CLI VERSION 自动替换；安装脚本由 `make release` 生成，不直接执行源码模板。
