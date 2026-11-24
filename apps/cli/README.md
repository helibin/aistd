# aistd CLI

维护者文档：Go 实现，构建 CLI 需要 Go 1.22 或以上；终端向导使用兼容 Go 1.22 的 `huh v0.6.0`，提供方向键、多选和取消处理，依赖固定在 `go.mod` / `go.sum`。最终用户只使用公网安装脚本和已编译的二进制，无需 Go 或源码。选择 Go API 的生成项目需要 Go 1.23 或以上；选择 NestJS、React/Vue、Expo、Taro/uniapp 需要 Node 22 和 npm；Flutter 需要独立的 Flutter SDK 及对应原生工具链。

```bash
make check
make build
./bin/aistd create --config examples/project.json --yes --dry-run
./bin/aistd create --config examples/project.json --yes
./bin/aistd create --name hello-aistd
./bin/aistd doctor --dir hello-aistd
./bin/aistd init --dir /path/to/existing-project --dry-run
./bin/aistd init --dir /path/to/existing-project
```

## 终端向导

在交互终端运行 `aistd` 或 `aistd create`：项目名称默认 `hello-aistd`，应用目标默认勾选 API 和 Web。使用方向键移动，空格切换多选，回车确认；随后只询问所选应用的技术栈和视觉主题草稿。每步保留选择摘要，最后显示目标目录和完整配置，确认后才创建文件。按 Ctrl+C 或 Esc 取消，尚未确认时不会写入项目；`aistd create --dry-run` 在确认后仅显示文件预览。

向导要求标准输入、标准输出均连接终端，且 `TERM` 已设置并非 `dumb`；管道、重定向或简易终端中的 `aistd create` 继续使用编号问答，无参数 `aistd` 显示帮助。设置 `NO_COLOR=1` 可关闭颜色。显式使用 `--name`、`--config` 或 `--yes` 时保留命令行及自动化流程。

## 创建后运行

`create` 成功后会按所选目标显示依赖要求、进入项目目录和启动命令；生成的 `README.md` 保留同一份快速开始说明。默认 API + Web 项目需要 Go 1.23+、Node.js 22、npm 和 make，首次运行：

```sh
aistd create --name hello-aistd
cd hello-aistd
make setup
make run-api
```

保留 API 终端，在第二个终端进入同一个 `hello-aistd` 目录并运行 `make run-web`。打开 Vite 打印的地址（通常为 `http://127.0.0.1:5173`），页面应显示 `Go API: hello`；API 可直接访问 `http://127.0.0.1:8080/api/hello`。如果先打开了 Web，再启动 API，需要刷新页面。

`make run` 只启动优先选中的一个目标：默认项目为 API，不会同时启动 Web。`make check` 与 `make build` 负责检查和构建，不会启动服务。停止开发时在各终端按 Ctrl+C。

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

交互模式 `aistd create` 只询问已选平台的技术栈。iOS/Android 共用一个移动应用，微信/支付宝共用一个小程序应用。JSON 对应字段为 `apiStack`、`webStack`、`adminStack`、`mobileStack`、`miniappStack`；不与 `--config` 混用这些参数，也不接受为未选择的平台指定技术栈。

生成过程离线运行；`make setup` 才安装所选栈的依赖。Go API 需要 Go 1.23+；NestJS、React、Vue、React Native/Expo、Taro、uniapp 需要 Node 22/npm；Flutter 需要 Flutter SDK 和对应原生构建工具，不属于 npm workspace。Flutter 的 Android 构建产出调试 APK，iOS 构建为未签名应用，需要 macOS/Xcode；Expo 构建导出 JS 资源。小程序构建后仍须使用开发者工具配置真实 App ID 并完成平台审核。

技术栈选择需使用包含本功能的新 CLI；官网配置器与 CLI 应一起发行。旧 JSON 缺少技术栈字段仍可用，缺少 `web` 时继续不生成 Web；旧 CLI 无法读取新增字段。切换已有项目技术栈和自动迁移不在 `create` 范围内。

## Admin 后台管理模板

Admin 是默认关闭的第七个独立目标；可只创建后台，也可与 API/Web 同时选择。`--admin-stack vben`（默认）使用 Vben Naive UI，`soybean` 使用 Soybean-admin，JSON 对应 `admin` 和 `adminStack`。Admin-only JSON 请显式设 `api: false`；未启用 Admin 却指定其技术栈会拒绝。

离线创建时安装固定 MIT 上游源码：Vben 5.7.0（`63a38dce49ba109f61607994e21ba921d8e970e9`）、Soybean 2.0.2（`5e40b850180804aa51573590eb2e058d1fd8eb06`）。`apps/admin` 保留各自 workspace/lockfile 和许可证；`.aistd/admin-template.json` 记录来源/哈希，`apps/admin/AISTD-TEMPLATE.md` 记录适配。Vben 保留完整 workspace，但只调用 Naive UI 入口。Admin 不属于根 npm workspaces，需要 Node 22.18+ 和 npm；根 `make setup` 通过 npx 调用固定 pnpm 10.33.4，无需全局安装，之后使用 `make check`、`make build`、`make run-admin`。Vben 本地入口为 `http://localhost:5888/admin/`，Soybean 为 `http://localhost:9527/admin/`，静态部署路径均为 `/admin/`。

Vben 开发保留本地 Nitro mock；Soybean 保留上游外部 Apifox 演示服务，生产默认配置也可能使用演示服务。尚未对接真实后台认证/权限，生成的 hello API 也不提供这些接口。升级 CLI 不迁移已有项目；`doctor` 只检查规范原文。维护源码的固定版本及哈希见 `internal/admintemplates/README.md`。

## 维护者验证入口

从仓库根先运行 `make setup`，下载 CLI 的固定模块依赖并启用提交检查，再运行以下命令（生成 API 的验证需 Go 1.23+；本地需设置可写的 `TMPDIR`，CI 使用 `runner.temp`）：

```sh
make -C apps/cli check                         # Go 格式、vet、单元测试
make -C apps/cli verify                        # 轻量离线：实际 CLI 生命周期 + API 检查/构建
make -C apps/cli verify-deep                    # 需网络：Go/NestJS/React/Vue/Expo/Taro/uniapp
make -C apps/cli verify-deep PLATFORMS=web       # 单独验证 Web
make -C apps/cli verify-deep PLATFORMS=expo      # iOS + Android JS 导出
make -C apps/cli verify-deep PLATFORMS=taro      # 微信 + 支付宝构建
make -C apps/cli verify-deep PLATFORMS="nestjs vue uniapp"
make -C apps/cli verify-deep PLATFORMS="vben soybean" # 固定 Admin 源码安装/类型检查/构建
make -C apps/cli verify-deep PLATFORMS=flutter-android # 需 Flutter/Android SDK/JDK
make -C apps/cli verify-deep PLATFORMS=flutter-ios     # 需 Flutter/macOS/Xcode
```

`verify` 禁用 Go 模块网络下载，不安装 npm 依赖；编译 CLI 时读取 `make setup` 已准备的模块缓存，生成 API 的检查与构建仍使用隔离的空模块缓存。使用临时编译的 CLI 验证 create→doctor、init→doctor、重复操作拒绝、防覆盖、两种 dry-run 不写入、篡改后 doctor 非零退出，并对生成的无外部模块 API 执行 `make setup check build`。它不表示其他技术栈可以构建。

`verify-deep` 每个平台分别创建真实项目，执行生成项目的 `make setup check build` 并断言产物存在；需 Go、Node 22、npm、make 和 npm registry 网络访问，可能耗时较长。`PLATFORMS` 可为 `api nestjs web vue expo taro uniapp flutter-android flutter-ios` 的任意非空组合；默认不运行 Flutter 原生构建，显式选择时需要对应 SDK。任何命令或产物检查失败都会非零退出，不跳过、不伪报通过。项目、HOME、构建缓存及日志均位于 `$TMPDIR/aistd-verify.*`（离线编译 CLI 时读取已有模块缓存），成功自动删除，失败保留并打印证据路径。不要将失败目录当作可发布产物。深层验证不默认接入 PR CI，以免离线检查隐式安装全部跨平台依赖；现有 CI 运行离线入口。

Flutter 深层验证同样隔离 HOME；Android 目标需显式设置 `ANDROID_HOME` 指向已有且完成许可配置的 SDK，例如 macOS 常见路径 `$HOME/Library/Android/sdk`，以免 SDK 自动发现依赖原 HOME。普通生成项目的 `make setup/check/build` 不隔离 HOME。

上述验证不运行模拟器、原生签名构建、开发者工具审核或 Docker 镜像构建；`doctor` 也不评估业务符合性。生成项目本身的明确入口仍是根目录 `make setup`（安装）、`make check`（检查）、`make build`（构建）。

## 安装与升级 CLI

安装及升级使用同一条命令，无需先卸载：

```sh
curl -fsSL https://aistd.vlb.cn/install.sh | sh
```

脚本默认安装到 `~/.local/bin`；新 CLI 下载、SHA256 校验和运行验证成功后，原子替换安装目录中的旧 CLI，失败保留旧版。符号链接、目录等异常安装目标会被拒绝。自定义安装目录时使用 `curl -fsSL https://aistd.vlb.cn/install.sh | AISTD_INSTALL_DIR=/path/to/bin sh`，之后升级沿用相同目录。升级仅替换 CLI，不迁移已有项目或规范文件。

## 卸载 CLI

安装脚本默认将 `aistd` 放在 `~/.local/bin`。运行 `aistd uninstall`，检查提示中的可执行文件路径并输入 `y` 确认；非交互环境可运行 `aistd uninstall --yes`。若安装时设置过 `AISTD_INSTALL_DIR`，卸载时必须设置相同目录，例如 `AISTD_INSTALL_DIR=/path/to/bin /path/to/bin/aistd uninstall`。命令拒绝删除安装目录以外的可执行文件或符号链接，只移除当前 CLI 文件，保留安装目录及已生成项目、已接入规范、锁文件和 `task.md`。

参数放在子命令之后。`--dir` 在 `create` 中表示父目录，在 `init` 和 `doctor` 中表示已有项目目录。

- `create`：可选 `--openspec`，使用官方 OpenSpec CLI 生成 CodeBuddy 工作流；默认关闭。`--name` 无需配置，默认生成 Go API、React Web、视觉草稿和根目录 `task.md` 开发计划；交互终端使用方向键与空格选择七种应用目标，非交互或简易终端使用编号问答，也可使用 `--name demo --targets ios,wechat --visual=false` 或 JSON 配置。目标目录存在时拒绝创建，`--dry-run` 不写文件。无论是否选择视觉主题，都安装工程与视觉 v1.0 原文到 `.aistd/standards/`。
- 规范快照：两个原始 ZIP 内嵌在 CLI，生成项目保留其完整目录树。`.aistd/standards.lock.json` 记录源 ZIP SHA256 和各文件 SHA256；快照内的参考项目与概念图不是活动应用代码。
- `api=true`、`apiStack=go`（默认）：生成 Go 标准库后端模板，按 `cmd/server`（进程入口）、`internal/config`（配置读取与校验）、`internal/httpapi`（路由、请求日志、健康检查及服务生命周期）组织，附带包级测试、OpenAPI、Dockerfile、Make 目标和 GitHub PR 检查。`HTTP_ADDR` 默认 `:8080`，通过环境变量注入，`.env.example` 不会自动加载；`--healthcheck` 使用同一地址访问 HTTP 健康端点。支持有界优雅退出，日志包含请求 ID、响应状态和耗时。固定 v1.0 原文保持不变，新模板只用于 `create`，不迁移已有项目；数据库、鉴权、生产部署和业务逻辑按需实现。
- `api=true`、`apiStack=nestjs`：生成 NestJS 后端、配置/健康检查、请求日志及 HTTP 回归测试，与 Go 共享三个 API 契约和默认端口；使用 Node/npm，不要求 Go。
- `web=true`：按 `webStack` 生成 React 或 Vue + Vite Web 应用；可独立于 API 使用，有 API 时代理 `/api`。旧 JSON 配置若未写 `web`，保持不生成 Web 应用。
- `ios` / `android`、`mobileStack=react-native`（默认）：按目标生成一个 Expo 应用。`make build` 导出所选平台的 JS 资源，不生成签名 IPA/AAB；使用 Node 22，原生发布需要另行配置标识、凭据和构建环境。
- `ios` / `android`、`mobileStack=flutter`：生成共享 Flutter/Dart 应用及所选平台原生壳；创建时不下载 SDK 或依赖，setup 执行 pub get，check 执行分析与测试。build 生成 Android 调试 APK / iOS 未签名应用，发布需要独立签名配置。
- `wechat` / `alipay`：按 `miniappStack` 生成 Taro React 或 uniapp Vue 3 小程序。`make build` 在独立目录输出所选平台文件；发布仍需对应平台的开发工具、App ID 和审核。
- `visual=true`：生成产品视觉方向文档和空白 Token 草稿。视觉规范中的 TrailGo 仅是示例，不会作为新产品的默认配色；不生成 App UI。
- `doctor`：将锁文件和受管规范文件与 CLI 内置快照逐项比对，拒绝缺失、修改、额外文件或不同版本。通过仅说明规范原文完整，不代表产品符合工程标准或通过视觉验收。
- `init`：可选 `--openspec`，仅当项目无 `openspec/` 与 `.codebuddy/` 时调用官方初始化；已有内容不覆盖。运行前需安装 OpenSpec CLI（Node.js >= 20.19；`npm install -g @fission-ai/openspec@latest`）。OpenSpec 管需求与变更，`task.md` 管项目级里程碑、阻塞和验收；`doctor` 不校验 OpenSpec。 在已有项目中新增工程与视觉规范原文、锁文件和 `AISTD-ADOPTION.md` 审计清单；根目录没有 `task.md` 时创建计划模板，已有时原样保留；检测到其他受管目标文件已存在时拒绝写入。不会修改已有代码、配置或指令文件。之后可运行 `doctor --dir .` 验证规范文件完整性。

新增技术栈起点由 aistd 提供，固定规范 v1.0 快照保持原样；它们不代表产品已符合规范。尚未实现规范升级、已有项目自动重构、数据库与生产发布流程。
