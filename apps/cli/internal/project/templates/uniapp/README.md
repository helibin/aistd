# uni-app mini program starter

This Vue 3 + Vite starter follows the official `dcloudio/uni-preset-vue` vite-ts toolchain. DCloud packages use one pinned release series. Source: https://github.com/dcloudio/uni-preset-vue/tree/vite-ts; CLI documentation: https://uniapp.dcloud.net.cn/quickstart-cli.html.

Use Node.js 22 and install dependencies from the project root with `make setup`. The root npm workspace owns the lockfile. `make check` runs TypeScript checks; `make build` builds each selected mini program.

WeChat output is `dist/build/mp-weixin`; Alipay output is `dist/build/mp-alipay`. Watch commands use `dist/dev/mp-weixin` and `dist/dev/mp-alipay`. Import the corresponding output directory into the platform's developer tools. Replace `touristappid` in `src/manifest.json` with your WeChat App ID before platform testing; configure the actual Alipay App ID in its developer tools. Builds alone do not prove platform approval or device behavior.
