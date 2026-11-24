# 发行与构建

安装 Go 1.23+、Node.js 22、npm 和 make，运行：

```sh
npm ci
make setup
make check
make build
```

`make build` 将本机 CLI 写入 `apps/cli/bin/aistd`，将官网写入 `dist/`。`make -C apps/cli verify` 可验证离线生成、规范完整性和模板行为。

`make release` 构建 macOS 和 Linux 的 amd64/arm64 CLI、SHA256 清单和网站。GitHub 的发行构建工作流只构建制品，不部署官网。对外分发二进制时应随附根 MIT 许可证和适用的第三方声明，见[许可范围](license-scope.md)。
