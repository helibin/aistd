package project

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"aistd"
	"aistd/internal/admintemplates"
	"aistd/internal/standards"
)

const standardsDir = ".aistd/standards"

const taskTemplate = `# Development plan

This file tracks the project's next deliverables and decisions. Keep it current as work moves forward; it is an editable project file, not part of the pinned standards checked by aistd doctor.

## Goal and scope

- Goal: define the next deliverable and its acceptance criteria.
- Scope: identify what is included and explicitly excluded.

## Planned

- [ ] Define the first milestone and acceptance criteria.

## In progress

- None.

## Blocked

- None. For each blocker, record its owner and the decision needed.

## Done

- None. Move completed items here with a date and verification evidence.

## Working agreement

- Give each task a clear outcome, owner (when applicable), and checkable acceptance criteria.
- Move tasks between sections as status changes; break large tasks into small deliverables.
- Update this plan when priorities change and after each milestone; keep technical decisions in project docs.
`

type Config struct {
	Name         string `json:"name"`
	API          bool   `json:"api"`
	Web          bool   `json:"web"`
	Admin        bool   `json:"admin"`
	IOS          bool   `json:"ios"`
	Android      bool   `json:"android"`
	WeChat       bool   `json:"wechat"`
	Alipay       bool   `json:"alipay"`
	Visual       bool   `json:"visual"`
	OpenSpec     bool   `json:"openspec,omitempty"`
	APIStack     string `json:"apiStack,omitempty"`
	WebStack     string `json:"webStack,omitempty"`
	MobileStack  string `json:"mobileStack,omitempty"`
	MiniappStack string `json:"miniappStack,omitempty"`
	AdminStack   string `json:"adminStack,omitempty"`
}

type lockedStandard struct {
	Version       string            `json:"version"`
	ArchiveSHA256 string            `json:"archiveSha256"`
	Files         map[string]string `json:"files"`
}

type standardsLock struct {
	Schema      int            `json:"schema"`
	Status      string         `json:"status"`
	Engineering lockedStandard `json:"engineering"`
	Visual      lockedStandard `json:"visual"`
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)

func (c Config) Validate() error {
	_, err := c.Normalize()
	return err
}

func (c Config) Normalize() (Config, error) {
	if !namePattern.MatchString(c.Name) {
		return c, fmt.Errorf("name must start with a lowercase letter and contain only lowercase letters, numbers or hyphens (1-63 characters)")
	}
	for _, stack := range []struct {
		name                string
		enabled             bool
		value               *string
		defaults, alternate string
	}{
		{"apiStack", c.API, &c.APIStack, "go", "nestjs"},
		{"webStack", c.Web, &c.WebStack, "react", "vue"},
		{"mobileStack", c.IOS || c.Android, &c.MobileStack, "react-native", "flutter"},
		{"miniappStack", c.WeChat || c.Alipay, &c.MiniappStack, "taro", "uniapp"},
		{"adminStack", c.Admin, &c.AdminStack, "vben", "soybean"},
	} {
		if !stack.enabled && *stack.value != "" {
			return c, fmt.Errorf("%s requires an enabled target", stack.name)
		}
		if !stack.enabled {
			continue
		}
		if *stack.value == "" {
			*stack.value = stack.defaults
		}
		if *stack.value != stack.defaults && *stack.value != stack.alternate {
			return c, fmt.Errorf("%s must be %s or %s", stack.name, stack.defaults, stack.alternate)
		}
	}
	return c, nil
}

func digest(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func standardFiles() (map[string]string, standardsLock, map[string]standards.Bundle, error) {
	bundles, err := standards.Load()
	if err != nil {
		return nil, standardsLock{}, nil, err
	}
	files := make(map[string]string)
	lock := standardsLock{Schema: 1, Status: "sources-installed"}
	for _, name := range []string{"engineering", "visual"} {
		bundle, ok := bundles[name]
		if !ok {
			return nil, standardsLock{}, nil, fmt.Errorf("missing embedded %s standard", name)
		}
		entry := lockedStandard{Version: bundle.Version, ArchiveSHA256: bundle.ArchiveSHA256, Files: make(map[string]string)}
		for rel, content := range bundle.Files {
			files[filepath.Join(standardsDir, name, "v"+bundle.Version, filepath.FromSlash(rel))] = string(content)
			entry.Files[rel] = digest(content)
		}
		if name == "engineering" {
			lock.Engineering = entry
		} else {
			lock.Visual = entry
		}
	}
	return files, lock, bundles, nil
}

func QuickStart(c Config) string {
	c, _ = c.Normalize()
	js := hasJS(c)
	if !c.API && !js && !c.Admin && c.MobileStack != "flutter" {
		return "No application was selected. Read AGENTS.md and update task.md before adding an application. Verify the installed sources with `aistd doctor --dir .`.\n"
	}
	var out strings.Builder
	out.WriteString("Requirements: make")
	if c.APIStack == "go" {
		out.WriteString(", Go 1.23+")
	}
	if c.MobileStack == "flutter" {
		out.WriteString(", Flutter 3.44.4 SDK (Android SDK; macOS/Xcode for iOS)")
	}
	if js || c.Admin {
		if c.Admin {
			out.WriteString(", Node.js 22.18+ and npm (Admin uses pinned pnpm via npx)")
		} else {
			out.WriteString(", Node.js 22 and npm")
		}
	}
	out.WriteString(". Install dependencies from the project root:\n\n```sh\nmake setup\n```\n\nStart the desired target below from the project root. Keep its terminal open; Ctrl+C stops it.\n")
	if c.API && c.Web {
		out.WriteString("Run the API and Web in separate terminals, both from the same project root.\n")
	}
	if c.API {
		out.WriteString("\n" + apiLabel(c) + ":\n\n```sh\nmake run-api\n```\n\nCheck http://127.0.0.1:8080/api/hello (expected: {\"message\":\"hello\"}).\n")
	}
	if c.Web {
		out.WriteString("\nReact Web:\n\n```sh\nmake run-web\n```\n\nOpen the URL printed by Vite (normally http://127.0.0.1:5173).\n")
		if c.API {
			out.WriteString("Start the API first, then open or reload the Web page; it should show `Go API: hello`. `make run` starts only the API, so use both commands above for the full stack.\n")
		}
	}
	if c.Admin {
		port := "5888"
		if c.AdminStack == "soybean" {
			port = "9527"
		}
		out.WriteString("\n" + adminLabel(c) + ":\n\n```sh\nmake run-admin\n```\n\nOpen http://127.0.0.1:" + port + "/admin/ (or the URL printed by the development server). The upstream frontend demo is not connected to this project's API authentication.\n")
	}
	if c.IOS {
		out.WriteString("\niOS (requires macOS with Xcode and an iOS simulator):\n\n```sh\nmake run-ios\n```\n")
	}
	if c.IOS && c.Android {
		out.WriteString("\niOS and Android share one Expo application; choose the target command for the device you are testing.\n")
	}
	if c.Android {
		out.WriteString("\nAndroid (requires a configured Android emulator or connected device):\n\n```sh\nmake run-android\n```\n")
	}
	if c.WeChat {
		out.WriteString("\nWeChat:\n\n```sh\nmake run-wechat\n```\n\nAfter compilation, import apps/miniapp/dist/weapp in WeChat Developer Tools and configure your App ID.\n")
	}
	if c.Alipay {
		out.WriteString("\nAlipay:\n\n```sh\nmake run-alipay\n```\n\nAfter compilation, import apps/miniapp/dist/alipay in Alipay Developer Tools and configure your App ID.\n")
	}
	if c.WeChat && c.Alipay {
		out.WriteString("\nWeChat and Alipay share one Taro application; choose the target command for the platform you are testing.\n")
	}
	out.WriteString("\nRun `make check` and `make build` separately to validate and build the selected applications; these commands do not start them.\n")
	result := out.String()
	if c.APIStack == "nestjs" {
		result = strings.ReplaceAll(result, "Go API: hello", "NestJS API: hello")
	}
	if c.WebStack == "vue" {
		result = strings.ReplaceAll(result, "React Web", "Vue Web")
		result = strings.ReplaceAll(result, apiLabel(c)+": hello", "API: hello")
	}
	if c.MobileStack == "flutter" {
		result = strings.ReplaceAll(result, "Expo application", "Flutter application")
		result += "\nFlutter run commands prompt for a connected target; select the requested iOS/Android device. Builds are debug APKs or unsigned iOS apps, not store releases.\n"
	}
	if c.MiniappStack == "uniapp" {
		result = strings.ReplaceAll(result, "Taro application", "uni-app application")
		result = strings.ReplaceAll(result, "apps/miniapp/dist/weapp", "apps/miniapp/dist/dev/mp-weixin")
		result = strings.ReplaceAll(result, "apps/miniapp/dist/alipay", "apps/miniapp/dist/dev/mp-alipay")
	}
	return result
}

func Files(c Config) (map[string]string, error) {
	c, err := c.Normalize()
	if err != nil {
		return nil, err
	}
	files, lock, bundles, err := standardFiles()
	if err != nil {
		return nil, err
	}
	config, _ := json.MarshalIndent(c, "", "  ")
	lockJSON, _ := json.MarshalIndent(lock, "", "  ")
	files["README.md"] = "# " + c.Name + "\n\nGenerated by aistd " + aistd.Version + ". Engineering and visual standard v1.0 source snapshots are installed under `.aistd/standards/`; run `aistd doctor --dir .` to verify their integrity. Installation does not certify this product against either standard.\n\n## Quick start\n\n" + QuickStart(c) + "\nSee `AGENTS.md` and the installed standards before adding applications or UI. Maintain the editable `task.md` for milestones, progress, blockers, and acceptance criteria; it is not checked by `doctor`.\n"
	files["AGENTS.md"] = "# Project rules\n\nRead `.aistd/standards/engineering/v1.0/AGENTS.md` and `.aistd/standards/engineering/v1.0/docs/STANDARD.md` before engineering work. For UI, also read `.aistd/standards/visual/v1.0/AGENTS.md` and `.aistd/standards/visual/v1.0/docs/VISUAL-STANDARD.md`. Keep product-specific decisions in project docs and tokens. Maintain `task.md` as the editable development plan; update task status and verification evidence as work progresses. Run applicable checks. `aistd doctor` verifies source integrity only; assess implementation and visual QA separately before claiming compliance.\n"
	files["task.md"] = taskTemplate
	if c.OpenSpec {
		files["README.md"] += "\n## OpenSpec (optional)\n\nOpenSpec was initialized with the official CLI for CodeBuddy. Use OpenSpec proposals, specs and change tasks to define and review behavior; keep `task.md` for project milestones, blockers and verification evidence. OpenSpec is not checked by `aistd doctor`. See https://openspec.dev/docs/ for the workflow.\n"
		files["AGENTS.md"] += "\nFor behavior changes, use the OpenSpec workflow and review the proposal before implementation; keep `task.md` for project-level progress.\n"
	}
	files[".aistd/project.json"] = string(config) + "\n"
	files[".aistd/standards.lock.json"] = string(lockJSON) + "\n"
	files[".aistd/generation.json"] = "{\"cli\":\"" + aistd.Version + "\",\"template\":\"standards-v1\"}\n"
	files[".gitignore"] = ".env\n.env.local\n.env.*.local\n/bin/\n**/node_modules/\n**/dist/\n**/coverage/\napps/mobile/.expo/\napps/miniapp/.temp/\n"
	if c.API && c.APIStack == "go" {
		api := bundles["engineering"].Files
		files["apps/api/go.mod"] = "module example.com/" + c.Name + "/api\n\ngo 1.23\n"
		for _, rel := range []string{".env.example", "CHANGELOG.md", "VERSION"} {
			content, ok := api["apps/api/"+rel]
			if !ok {
				return nil, fmt.Errorf("missing engineering reference file: apps/api/%s", rel)
			}
			files[filepath.Join("apps/api", rel)] = string(content)
		}
		activeAPI, err := apiFiles(c.Name)
		if err != nil {
			return nil, err
		}
		for path, content := range activeAPI {
			files[path] = content
		}
		contract, ok := api["contracts/openapi.yaml"]
		if !ok {
			return nil, fmt.Errorf("missing engineering reference file: contracts/openapi.yaml")
		}
		files["contracts/openapi.yaml"] = string(contract)
		files["README.md"] += "\n## Go API starter\n\n`apps/api` keeps the engineering v1.0 HTTP contract in `contracts/openapi.yaml`: `/api/health/live`, `/api/health/ready`, and `/api/hello`. Its active starter is organized into `cmd/server` (assembly and signals), `internal/config` (validated environment configuration), and `internal/httpapi` (routes, request logs, health checks and server lifecycle). The pinned standard sources remain unchanged.\n\nRun `make run-api`; the default listen address is `:8080`. Set `HTTP_ADDR` explicitly to change it, for example `HTTP_ADDR=127.0.0.1:9090 make run-api`. `.env.example` documents configuration; the application does not load .env files. After `make build`, `HTTP_ADDR=127.0.0.1:9090 ./bin/api --healthcheck` verifies the live HTTP response at that address. The Web development proxy uses port 8080 by default; update it if you change the API port.\n\nResponses include `X-Request-ID`; access logs record the generated request ID, method, path, status and duration. HTTP read-header/read/write/idle timeouts are 5/10/15/60 seconds. SIGINT/SIGTERM stop new connections and allow active requests up to five seconds to finish, then force-close connections and exit with an error if the deadline expires. `make check` runs configuration, route, health-check and shutdown tests. `make docker-build-api` builds a local image. This starter has no database or business logic; review security, observability, release and conditional requirements before production deployment.\n"
	}
	if c.API && c.APIStack == "nestjs" {
		generated, err := nestjsFiles(c)
		if err != nil {
			return nil, err
		}
		for path, content := range generated {
			files[path] = content
		}
		files["contracts/openapi.yaml"] = string(bundles["engineering"].Files["contracts/openapi.yaml"])
		files["README.md"] += "\n## NestJS API starter\n\n`apps/api/src` contains the module, controller, configuration and bootstrap. Run `make run-api` (compiles once before starting); restart it after code changes. `HTTP_ADDR` defaults to `:8080`; .env files are not automatically loaded. The three `/api/health/live`, `/api/health/ready`, `/api/hello` responses retain `contracts/openapi.yaml`. Responses/logs include a generated request ID. SIGINT/SIGTERM allow five seconds to close active requests before forcing exit. Run `npm run healthcheck --workspace apps/api` after build with the same HTTP_ADDR to verify real HTTP health. `make check` includes real HTTP tests, and `make docker-build-api` builds a standalone Node image. This is a starter without database or business logic.\n"
	}
	if c.Web && c.WebStack == "vue" {
		generated, err := vueFiles(c)
		if err != nil {
			return nil, err
		}
		for path, content := range generated {
			files[path] = content
		}
		files["README.md"] += "\n## Vue web starter\n\n`apps/web` uses Vue 3, Vite and TypeScript. Run `make run-web`; `make check` type-checks Vue files and `make build` writes `apps/web/dist`. An API-enabled project proxies `/api` to port 8080 during development.\n"
	}
	if c.Web && c.WebStack == "react" {
		web := bundles["engineering"].Files
		for _, rel := range []string{".dockerignore", ".env.example", "CHANGELOG.md", "Dockerfile", "VERSION", "index.html", "package.json", "src/main.tsx", "tests/smoke.test.mjs", "tsconfig.json", "vite.config.ts"} {
			content, ok := web["apps/web/"+rel]
			if !ok {
				return nil, fmt.Errorf("missing engineering reference file: apps/web/%s", rel)
			}
			files[filepath.Join("apps/web", rel)] = string(content)
		}
		var packageJSON map[string]any
		if err := json.Unmarshal([]byte(files["apps/web/package.json"]), &packageJSON); err != nil {
			return nil, fmt.Errorf("invalid engineering web package: %w", err)
		}
		packageJSON["name"] = c.Name + "-web"
		encoded, err := json.MarshalIndent(packageJSON, "", "  ")
		if err != nil {
			return nil, err
		}
		files["apps/web/package.json"] = string(encoded) + "\n"
		files["apps/web/index.html"] = "<!doctype html><html lang=\"en\"><head><meta charset=\"UTF-8\"/><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"/><title>" + c.Name + "</title></head><body><div id=\"root\"></div><script type=\"module\" src=\"/src/main.tsx\"></script></body></html>\n"
		if !c.API {
			files["apps/web/vite.config.ts"] = "import { defineConfig } from 'vite';\nimport react from '@vitejs/plugin-react';\nexport default defineConfig({ plugins: [react()] });\n"
		}
		main := "import React from 'react';\nimport { createRoot } from 'react-dom/client';\n\nfunction App() {\n"
		if c.API {
			main += "  const [message, setMessage] = React.useState('Connecting to API...');\n  React.useEffect(() => {\n    fetch('/api/hello').then(response => response.json()).then(value => setMessage(value.message)).catch(() => setMessage('API unavailable'));\n  }, []);\n"
		}
		main += "  return <main style={{ fontFamily: 'system-ui', maxWidth: 720, margin: '4rem auto' }}><h1>" + c.Name + "</h1>"
		if c.API {
			main += "<p>" + apiLabel(c) + ": {message}</p>"
		}
		main += "</main>;\n}\n\ncreateRoot(document.getElementById('root')!).render(<App />);\n"
		files["apps/web/src/main.tsx"] = main
		files["README.md"] += "\n## React web starter\n\n`apps/web` is adapted from the engineering v1.0 React + Vite reference starter. Run `make run-web` for local development and `make docker-build-web` for a static image. Dependencies install into the root workspace; commit the root `package-lock.json` before CI or release. This is a starter screen; define the product UI and complete visual QA before shipping.\n"
	}
	if c.IOS || c.Android {
		var mobile map[string]string
		var err error
		if c.MobileStack == "flutter" {
			mobile, err = flutterFiles(c)
		} else {
			mobile, err = mobileFiles(c)
		}
		if err != nil {
			return nil, err
		}
		for path, content := range mobile {
			files[path] = content
		}
		if c.MobileStack == "flutter" {
			files["README.md"] += "\n## Flutter app starter\n\n`apps/mobile` contains a Flutter app with the selected native platform shells. Use Flutter 3.44.4, run `make setup` to fetch Dart packages, and commit `apps/mobile/pubspec.lock`. `make check` runs flutter analyze and flutter test. Android builds produce a debug APK; iOS builds produce an unsigned Runner.app and require macOS/Xcode. `make run-ios` and `make run-android` invoke flutter run; choose the corresponding device when prompted. These are not store release packages: configure identifiers, signing and platform distribution before release.\n"
		} else {
			files["README.md"] += "\n## Mobile app starter\n\n`apps/mobile` is an Expo starter for the selected iOS and/or Android targets. Run `make setup`, `make check`, and `make build`; use `make run-ios` or `make run-android` for development. Build exports are JavaScript bundles, not signed IPA/AAB packages. Configure real identifiers, credentials and native builds before release. Commit the root package-lock.json after setup.\n"
		}
	}
	if c.WeChat || c.Alipay {
		var miniapp map[string]string
		var err error
		if c.MiniappStack == "uniapp" {
			miniapp, err = uniappFiles(c)
		} else {
			miniapp, err = miniappFiles(c)
		}
		if err != nil {
			return nil, err
		}
		for path, content := range miniapp {
			files[path] = content
		}
		files["README.md"] += "\n## Miniapp starter\n\n`apps/miniapp` is a Taro React starter for the selected WeChat and/or Alipay targets. Run `make setup`, `make check`, and `make build`; import `apps/miniapp/dist/weapp` or `apps/miniapp/dist/alipay` into the matching platform developer tools. Configure real app IDs, permissions and platform review before release. Commit the root `package-lock.json` after setup.\n"
	}
	if c.Admin {
		admin, err := admintemplates.Load(c.AdminStack)
		if err != nil {
			return nil, err
		}
		for path, content := range admin {
			files[path] = content
		}
		// Upstream .env files contain public template defaults and must survive a Git clone.
		for _, path := range Paths(admin) {
			if strings.HasPrefix(path, "apps/admin/") && strings.HasSuffix(path, "/.env") {
				files[".gitignore"] += "!/" + path + "\n"
			}
		}
		for path, content := range adminOpsFiles(c) {
			files[path] = content
		}
		files["README.md"] += "\n## Admin template\n\n`apps/admin` is an independent upstream pnpm workspace (not part of the root npm workspaces). Run `make setup`, `make check`, `make build`, and `make run-admin` from this project root. Node 22.18+ and npm are required; commands use pinned pnpm 10.33.4 through npx. The complete upstream workspace and MIT license are retained; Vben commands select only the Naive UI application. Source revisions and archive hashes are recorded in `.aistd/admin-template.json`. Admin is served under `/admin/`, alongside the optional Web frontend at `/`. These are upstream demonstration frontends; configure a real backend and authentication before production. The generated hello API does not provide their login or permission endpoints.\n"
	}
	jsWorkspaces := []string{}
	for _, app := range []string{"api", "web", "mobile", "miniapp"} {
		if _, ok := files["apps/"+app+"/package.json"]; ok {
			jsWorkspaces = append(jsWorkspaces, "apps/"+app)
		}
	}
	if len(jsWorkspaces) > 0 {
		rootPackage, err := jsonFile(map[string]any{"name": c.Name, "version": "0.1.0", "private": true, "workspaces": jsWorkspaces})
		if err != nil {
			return nil, err
		}
		files["package.json"] = rootPackage
		files["README.md"] += "\n## npm workspaces\n\nJS packages share the root package.json workspace. `make setup` installs them once; commit the generated root package-lock.json before CI or release.\n"
	}
	if c.MobileStack == "flutter" {
		files[".gitignore"] += "apps/mobile/.dart_tool/\napps/mobile/build/\napps/mobile/android/.gradle/\napps/mobile/android/local.properties\napps/mobile/ios/Pods/\n"
	}
	if c.MiniappStack == "uniapp" {
		files["README.md"] = strings.ReplaceAll(files["README.md"], "Taro React starter", "uni-app Vue starter")
		files["README.md"] = strings.ReplaceAll(files["README.md"], "apps/miniapp/dist/weapp", "apps/miniapp/dist/build/mp-weixin")
		files["README.md"] = strings.ReplaceAll(files["README.md"], "apps/miniapp/dist/alipay", "apps/miniapp/dist/build/mp-alipay")
	}
	if err := rootOpsFiles(c, bundles["engineering"].Files, files); err != nil {
		return nil, err
	}
	files["Makefile"] = makefile(c)
	if c.API || c.Web || c.Admin || c.IOS || c.Android || c.WeChat || c.Alipay {
		files[".github/workflows/check.yml"] = checkWorkflow(c)
	}
	if c.Visual {
		files["docs/design/theme.md"] = "# Product visual identity (draft)\n\nBefore building screens, record the audience, platforms, usage context, 3-5 brand traits, two contrasting directions, rejected patterns, typography and layout rules. Use `.aistd/standards/visual/v1.0/docs/VISUAL-STANDARD.md`, `.aistd/standards/visual/v1.0/docs/VISUAL-QA.md`, and `.aistd/standards/visual/v1.0/prompts/new-screen.md` as the source requirements. The token file is a blank project scaffold; the TrailGo example in the standard archive is illustrative and must not be adopted as this product's palette. Record screenshot and contrast results before visual signoff.\n"
		files["design/tokens/brand.json"] = "{\n  \"status\": \"draft\",\n  \"primitive\": {},\n  \"semantic\": {\"light\": {}},\n  \"component\": {},\n  \"type\": {},\n  \"space\": {}\n}\n"
		files["README.md"] += "\n## Visual design\n\nFill `docs/design/theme.md` and `design/tokens/brand.json` for this product, then complete the installed visual QA checklist for each screen family. These draft files are not an approved design system.\n"
	}
	return files, nil
}

// rootOpsFiles materializes the engineering reference's root-level operations
// files for generated projects: an editor config for every project, plus local
// Compose orchestration, ops scripts, and a Caddy edge adapted to the selected
// API/Web/Admin targets. These are project copies; the pinned v1.0 snapshots under
// .aistd/standards stay untouched.
func rootOpsFiles(c Config, engineering map[string][]byte, files map[string]string) error {
	editorconfig, ok := engineering[".editorconfig"]
	if !ok {
		return fmt.Errorf("missing engineering reference file: .editorconfig")
	}
	files[".editorconfig"] = string(editorconfig)
	if !c.API && !c.Web && !c.Admin {
		return nil
	}
	envExample, ok := engineering["deploy/.env.example"]
	if !ok {
		return fmt.Errorf("missing engineering reference file: deploy/.env.example")
	}
	sizeReport, ok := engineering["scripts/size-report.sh"]
	if !ok {
		return fmt.Errorf("missing engineering reference file: scripts/size-report.sh")
	}
	files["deploy/compose.yaml"] = composeFile(c)
	files["deploy/.env.example"] = string(envExample)
	files["deploy/caddy/Caddyfile"] = caddyFile(c)
	for path, content := range scriptFiles(c) {
		files[path] = content
	}
	files["scripts/size-report.sh"] = string(sizeReport)
	files["README.md"] += "\n## Local orchestration and ops scripts\n\n`deploy/compose.yaml` runs the selected starters behind a local Caddy edge bound to 127.0.0.1. Copy `deploy/.env.example` to `deploy/.env`, then run `make config`, `make up`, `make logs`, and `make down`. `scripts/` provides the engineering v1.0 reference operations entry points: `check.sh`, `build.sh`, `publish.sh` (requires `REGISTRY` and an immutable `IMAGE_TAG`), and `size-report.sh`. This is local reference orchestration, not a production deployment; review the deployment and operations guidance in the installed standard before shipping.\n"
	return nil
}

func composeFile(c Config) string {
	var out strings.Builder
	out.WriteString("name: " + c.Name + "\nservices:\n")
	if c.API {
		out.WriteString("  api:\n    build:\n      context: ../apps/api\n    environment:\n      HTTP_ADDR: ':8080'\n    restart: unless-stopped\n    read_only: true\n    security_opt: [no-new-privileges:true]\n    cap_drop: [ALL]\n    healthcheck:\n      test: [\"CMD\", \"/api\", \"--healthcheck\"]\n      interval: 30s\n      timeout: 3s\n      retries: 3\n")
	}
	if c.Web {
		out.WriteString("  web:\n    build:\n      context: ../apps/web\n    restart: unless-stopped\n    read_only: true\n    security_opt: [no-new-privileges:true]\n    cap_drop: [ALL]\n")
	}
	if c.Admin {
		out.WriteString("  admin:\n    build:\n      context: ../apps/admin\n      dockerfile: AISTD.Dockerfile\n    restart: unless-stopped\n    read_only: true\n    security_opt: [no-new-privileges:true]\n    cap_drop: [ALL]\n    tmpfs:\n      - /data\n      - /config\n")
	}
	out.WriteString("  edge:\n    image: caddy:2.10.0-alpine\n    ports:\n      - '127.0.0.1:${PUBLIC_PORT:-8088}:80'\n    volumes:\n      - ./caddy/Caddyfile:/etc/caddy/Caddyfile:ro\n    depends_on:\n")
	if c.API {
		out.WriteString("      - api\n")
	}
	if c.Web {
		out.WriteString("      - web\n")
	}
	if c.Admin {
		out.WriteString("      - admin\n")
	}
	out.WriteString("    restart: unless-stopped\n    security_opt: [no-new-privileges:true]\n    cap_drop: [ALL]\n")
	result := out.String()
	if c.APIStack == "nestjs" {
		result = strings.Replace(result, `["CMD", "/api", "--healthcheck"]`, `["CMD", "node", "scripts/healthcheck.cjs"]`, 1)
	}
	return result
}

func caddyFile(c Config) string {
	var out strings.Builder
	out.WriteString(":80 {\n  encode zstd gzip\n")
	if c.API && (c.Web || c.Admin) {
		out.WriteString("  handle /api/* {\n    reverse_proxy api:8080\n  }\n")
	}
	if c.Admin {
		out.WriteString("  @admin path /admin /admin/*\n  handle @admin {\n    reverse_proxy admin:80\n  }\n")
		if !c.API && !c.Web {
			out.WriteString("  handle {\n    redir / /admin/ 302\n    respond 404\n  }\n}\n")
			return out.String()
		}
	}
	out.WriteString("  handle {\n    reverse_proxy ")
	if c.API && !c.Web {
		out.WriteString("api:8080")
	} else {
		out.WriteString("web:80")
	}
	out.WriteString("\n  }\n}\n")
	return out.String()
}

func scriptFiles(c Config) map[string]string {
	scripts := make(map[string]string)
	var check strings.Builder
	check.WriteString("#!/bin/sh\nset -eu\n")
	if c.API && c.APIStack == "go" {
		check.WriteString("(cd apps/api && test -z \"$(gofmt -l ./cmd ./internal)\" && go vet ./... && go test ./...)\n")
	}
	if c.APIStack == "nestjs" {
		check.WriteString("npm run check --workspace apps/api\n")
	}
	if c.Web {
		check.WriteString("npm run lint --workspace apps/web\nnpm test --workspace apps/web\n")
	}
	if c.IOS || c.Android {
		if c.MobileStack == "flutter" {
			check.WriteString("(cd apps/mobile && flutter analyze && flutter test)\n")
		} else {
			check.WriteString("npm run check --workspace apps/mobile\n")
		}
	}
	if c.WeChat || c.Alipay {
		check.WriteString("npm run check --workspace apps/miniapp\n")
	}
	if c.Admin {
		check.WriteString(adminCheck(c) + "\n")
	}
	check.WriteString("printf 'check: success\\n'\n")
	scripts["scripts/check.sh"] = check.String()

	var build strings.Builder
	build.WriteString("#!/bin/sh\nset -eu\n")
	if c.API && c.APIStack == "go" {
		build.WriteString("mkdir -p dist\n(cd apps/api && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o ../../dist/api ./cmd/server)\n")
	}
	if c.APIStack == "nestjs" {
		build.WriteString("npm run build --workspace apps/api\n")
	}
	if c.Web {
		build.WriteString("npm run build --workspace apps/web\n")
	}
	if c.IOS {
		if c.MobileStack == "flutter" {
			build.WriteString("(cd apps/mobile && flutter build ios --debug --no-codesign)\n")
		} else {
			build.WriteString("npm run build:ios --workspace apps/mobile\n")
		}
	}
	if c.Android {
		if c.MobileStack == "flutter" {
			build.WriteString("(cd apps/mobile && flutter build apk --debug)\n")
		} else {
			build.WriteString("npm run build:android --workspace apps/mobile\n")
		}
	}
	if c.WeChat {
		build.WriteString("npm run build:weapp --workspace apps/miniapp\n")
	}
	if c.Alipay {
		build.WriteString("npm run build:alipay --workspace apps/miniapp\n")
	}
	if c.Admin {
		build.WriteString(adminBuild(c) + "\n")
	}
	build.WriteString("printf 'build: success\\n'\n")
	scripts["scripts/build.sh"] = build.String()

	var publish strings.Builder
	publish.WriteString("#!/bin/sh\nset -eu\n: \"${REGISTRY:?REGISTRY required, e.g. registry.example.com/team}\"\n: \"${IMAGE_TAG:?IMAGE_TAG required; use immutable tag such as commit SHA}\"\ncase \"$IMAGE_TAG\" in latest|dev) echo 'Refusing mutable release tag' >&2; exit 2;; esac\n")
	if c.API {
		publish.WriteString("docker build -f apps/api/Dockerfile -t \"$REGISTRY/api:$IMAGE_TAG\" apps/api\ndocker push \"$REGISTRY/api:$IMAGE_TAG\"\n")
	}
	if c.Web {
		publish.WriteString("docker build -f apps/web/Dockerfile -t \"$REGISTRY/web:$IMAGE_TAG\" apps/web\ndocker push \"$REGISTRY/web:$IMAGE_TAG\"\n")
	}
	if c.Admin {
		publish.WriteString("docker build -f apps/admin/AISTD.Dockerfile -t \"$REGISTRY/admin:$IMAGE_TAG\" apps/admin\ndocker push \"$REGISTRY/admin:$IMAGE_TAG\"\n")
	}
	publish.WriteString("printf 'published images; resolve and record image digests before deployment\\n'\n")
	scripts["scripts/publish.sh"] = publish.String()
	return scripts
}

func makefile(c Config) string {
	var out strings.Builder
	phony := ".PHONY: setup check build run run-api run-web run-admin run-ios run-android run-wechat run-alipay docker-build docker-build-api docker-build-web docker-build-admin"
	if c.API || c.Web || c.Admin {
		phony += " config up down logs release-image size-report"
	}
	out.WriteString(phony + "\nsetup:\n")
	if c.API && c.APIStack == "go" {
		out.WriteString("\t@go version\n")
	}
	js := hasJS(c)
	if js {
		out.WriteString("\t@npm --version\n\tnpm install --no-audit --no-fund\n")
	}
	if c.Admin {
		out.WriteString("\tCI=1 SKIP_INSTALL_SIMPLE_GIT_HOOKS=1 " + adminPNPM + " install --frozen-lockfile\n")
	}
	if !c.API && !js && !c.Admin && c.MobileStack != "flutter" {
		out.WriteString("\t@printf 'No application dependencies yet\\n'\n")
	}
	if c.MobileStack == "flutter" {
		out.WriteString("\tcd apps/mobile && flutter pub get\n")
	}
	out.WriteString("check:\n")
	if c.APIStack == "nestjs" {
		out.WriteString("\tnpm run check --workspace apps/api\n")
	}
	if c.API && c.APIStack == "go" {
		out.WriteString("\t@test -z \"$$(gofmt -l apps/api/cmd apps/api/internal)\"\n\tcd apps/api && go vet ./... && go test ./...\n")
	}
	if c.Web {
		out.WriteString("\tnpm run lint --workspace apps/web\n\tnpm test --workspace apps/web\n")
	}
	if c.IOS || c.Android {
		if c.MobileStack == "flutter" {
			out.WriteString("\tcd apps/mobile && flutter analyze && flutter test\n")
		} else {
			out.WriteString("\tnpm run check --workspace apps/mobile\n")
		}
	}
	if c.WeChat || c.Alipay {
		out.WriteString("\tnpm run check --workspace apps/miniapp\n")
	}
	if c.Admin {
		out.WriteString("\t" + adminCheck(c) + "\n")
	}
	if !c.API && !js && !c.Admin && c.MobileStack != "flutter" {
		out.WriteString("\t@printf 'No application checks yet; run aistd doctor --dir . for standards integrity\\n'\n")
	}
	out.WriteString("build:\n")
	if c.APIStack == "nestjs" {
		out.WriteString("\tnpm run build --workspace apps/api\n")
	}
	if c.API && c.APIStack == "go" {
		out.WriteString("\tmkdir -p bin\n\tcd apps/api && go build -o ../../bin/api ./cmd/server\n")
	}
	if c.Web {
		out.WriteString("\tnpm run build --workspace apps/web\n")
	}
	if c.IOS {
		if c.MobileStack == "flutter" {
			out.WriteString("\tcd apps/mobile && flutter build ios --debug --no-codesign\n")
		} else {
			out.WriteString("\tnpm run build:ios --workspace apps/mobile\n")
		}
	}
	if c.Android {
		if c.MobileStack == "flutter" {
			out.WriteString("\tcd apps/mobile && flutter build apk --debug\n")
		} else {
			out.WriteString("\tnpm run build:android --workspace apps/mobile\n")
		}
	}
	if c.WeChat {
		out.WriteString("\tnpm run build:weapp --workspace apps/miniapp\n")
	}
	if c.Alipay {
		out.WriteString("\tnpm run build:alipay --workspace apps/miniapp\n")
	}
	if c.Admin {
		out.WriteString("\t" + adminBuild(c) + "\n")
	}
	if !c.API && !js && !c.Admin && c.MobileStack != "flutter" {
		out.WriteString("\t@printf 'No application build yet\\n'\n")
	}
	switch {
	case c.API:
		out.WriteString("run: run-api\n")
	case c.Web:
		out.WriteString("run: run-web\n")
	case c.Admin:
		out.WriteString("run: run-admin\n")
	case c.IOS:
		out.WriteString("run: run-ios\n")
	case c.Android:
		out.WriteString("run: run-android\n")
	case c.WeChat:
		out.WriteString("run: run-wechat\n")
	case c.Alipay:
		out.WriteString("run: run-alipay\n")
	}
	if c.API && c.APIStack == "go" {
		out.WriteString("docker-build: docker-build-api\nrun-api:\n\tcd apps/api && go run ./cmd/server\ndocker-build-api:\n\tdocker build -t local/" + c.Name + "-api:dev apps/api\n")
	}
	if c.APIStack == "nestjs" {
		out.WriteString("docker-build: docker-build-api\nrun-api:\n\tnpm run dev --workspace apps/api\ndocker-build-api:\n\tdocker build -t local/" + c.Name + "-api:dev apps/api\n")
	}
	if c.Web {
		if !c.API {
			out.WriteString("docker-build: docker-build-web\n")
		}
		out.WriteString("run-web:\n\tnpm run dev --workspace apps/web\ndocker-build-web:\n\tdocker build -t local/" + c.Name + "-web:dev apps/web\n")
	}
	if c.Admin {
		if !c.API && !c.Web {
			out.WriteString("docker-build: docker-build-admin\n")
		}
		out.WriteString("run-admin:\n\t" + adminRun(c) + "\ndocker-build-admin:\n\tdocker build -f apps/admin/AISTD.Dockerfile -t local/" + c.Name + "-admin:dev apps/admin\n")
	}
	if c.IOS {
		if c.MobileStack == "flutter" {
			out.WriteString("run-ios:\n\tcd apps/mobile && flutter run\n")
		} else {
			out.WriteString("run-ios:\n\tnpm run ios --workspace apps/mobile\n")
		}
	}
	if c.Android {
		if c.MobileStack == "flutter" {
			out.WriteString("run-android:\n\tcd apps/mobile && flutter run\n")
		} else {
			out.WriteString("run-android:\n\tnpm run android --workspace apps/mobile\n")
		}
	}
	if c.WeChat {
		out.WriteString("run-wechat:\n\tnpm run dev:weapp --workspace apps/miniapp\n")
	}
	if c.Alipay {
		out.WriteString("run-alipay:\n\tnpm run dev:alipay --workspace apps/miniapp\n")
	}
	if c.API || c.Web || c.Admin {
		out.WriteString("REGISTRY ?= example.invalid/team\nIMAGE_TAG ?= dev\nconfig:\n\tdocker compose --env-file deploy/.env -f deploy/compose.yaml config\nup:\n\tdocker compose --env-file deploy/.env -f deploy/compose.yaml up -d --build\ndown:\n\tdocker compose --env-file deploy/.env -f deploy/compose.yaml down\nlogs:\n\tdocker compose --env-file deploy/.env -f deploy/compose.yaml logs -f\nrelease-image:\n\tREGISTRY='$(REGISTRY)' IMAGE_TAG='$(IMAGE_TAG)' ./scripts/publish.sh\nsize-report:\n\t./scripts/size-report.sh\n")
	}
	return out.String()
}

func checkWorkflow(c Config) string {
	var out strings.Builder
	out.WriteString("name: check\non:\n  push:\n  pull_request:\njobs:\n  check:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n")
	if c.API && c.APIStack == "go" {
		out.WriteString("      - uses: actions/setup-go@v5\n        with:\n          go-version-file: apps/api/go.mod\n")
	}
	if hasJS(c) || c.Admin {
		out.WriteString("      - uses: actions/setup-node@v4\n        with:\n          node-version: '22'\n")
	}
	if c.MobileStack == "flutter" {
		if c.Android {
			out.WriteString("      - uses: actions/setup-java@v4\n        with:\n          distribution: temurin\n          java-version: '17'\n      - uses: android-actions/setup-android@v3\n")
		}
		out.WriteString("      - uses: subosito/flutter-action@v2\n        with:\n          flutter-version: '3.44.4'\n          channel: stable\n")
	}
	out.WriteString("      - run: make setup\n      - run: make check\n      - run: make build\n")
	result := out.String()
	if c.MobileStack == "flutter" && c.IOS {
		result = strings.Replace(result, "ubuntu-latest", "macos-latest", 1)
	}
	return result
}

func Paths(files map[string]string) []string {
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

func Create(root string, c Config, dry bool) error {
	c, err := c.Normalize()
	if err != nil {
		return err
	}
	files, err := Files(c)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(root); err == nil {
		return fmt.Errorf("target already exists: %s", root)
	} else if !os.IsNotExist(err) {
		return err
	}
	if dry {
		for _, p := range Paths(files) {
			fmt.Println(filepath.Join(root, p))
		}
		if c.OpenSpec {
			fmt.Println("OpenSpec: openspec init", root, "--tools codebuddy (requires installed OpenSpec CLI)")
		}
		return nil
	}
	var adminExecutables map[string]bool
	if c.Admin {
		adminExecutables, err = admintemplates.ExecutablePaths(c.AdminStack)
		if err != nil {
			return err
		}
	}
	if c.OpenSpec {
		if _, err := openSpecCLI(); err != nil {
			return err
		}
	}
	if err := os.Mkdir(root, 0755); err != nil {
		return err
	}
	complete := false
	defer func() {
		if !complete {
			os.RemoveAll(root)
		}
	}()
	for _, p := range Paths(files) {
		target := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		mode := fs.FileMode(0644)
		if (strings.HasPrefix(p, "scripts/") && strings.HasSuffix(p, ".sh")) || p == "apps/mobile/android/gradlew" || adminExecutables[p] {
			mode = 0755
		}
		if err := os.WriteFile(target, []byte(files[p]), mode); err != nil {
			return err
		}
	}
	if c.OpenSpec {
		if err := initOpenSpec(root); err != nil {
			return err
		}
	}
	complete = true
	return nil
}

func openSpecCLI() (string, error) {
	binary, err := exec.LookPath("openspec")
	if err != nil {
		return "", fmt.Errorf("OpenSpec requested but CLI is not installed: install @fission-ai/openspec, then retry: %w", err)
	}
	return binary, nil
}

func initOpenSpec(root string) error {
	binary, err := openSpecCLI()
	if err != nil {
		return err
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	cmd := exec.Command(binary, "init", absolute, "--tools", "codebuddy")
	cmd.Dir = absolute
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("OpenSpec initialization failed (project may contain partial OpenSpec files): %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// Adopt adds pinned sources and a review checklist without touching existing project files.
func Adopt(root string, dry bool) error {
	return AdoptWithOpenSpec(root, dry, false)
}

// AdoptWithOpenSpec optionally initializes the official OpenSpec CodeBuddy integration.
func AdoptWithOpenSpec(root string, dry bool, openSpec bool) error {
	var err error
	root, err = filepath.Abs(root)
	if err != nil {
		return err
	}
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("target is not a directory")
	}
	files, lock, _, err := standardFiles()
	if err != nil {
		return err
	}
	lockJSON, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	files[".aistd/standards.lock.json"] = string(lockJSON) + "\n"
	files["AISTD-ADOPTION.md"] = "# Standards adoption review\n\nStatus: pending project audit. Pinned engineering and visual v1.0 sources are installed under `.aistd/standards/`. Run `aistd doctor --dir .` to verify their integrity; this does not assess project conformance.\n\nReview the installed engineering and visual standards against the existing architecture, public contracts, CI, UI, and release process. Record decisions, gaps, accepted exceptions, and visual QA evidence here before claiming compliance. Maintain the project's development plan in `task.md` (created only if absent). Existing application and instruction files were not modified by init.\n"
	// A plan is user-owned: preserve an existing task.md instead of treating it as a managed file.
	if _, err := os.Lstat(filepath.Join(root, "task.md")); os.IsNotExist(err) {
		files["task.md"] = taskTemplate
	} else if err != nil {
		return err
	}
	for _, rel := range append(Paths(files), ".aistd/standards") {
		if _, err := os.Lstat(filepath.Join(root, rel)); err == nil {
			return fmt.Errorf("target already exists: %s", filepath.Join(root, rel))
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if entry, err := os.Lstat(filepath.Join(root, ".aistd")); err == nil && !entry.IsDir() {
		return fmt.Errorf(".aistd is not a directory")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	if openSpec {
		// The upstream init may refresh existing tool files; refuse to touch either tree.
		for _, rel := range []string{"openspec", ".codebuddy"} {
			if _, err := os.Lstat(filepath.Join(root, rel)); err == nil {
				return fmt.Errorf("OpenSpec integration would touch existing %s; initialize it manually or omit --openspec", rel)
			} else if !os.IsNotExist(err) {
				return err
			}
		}
	}
	if dry {
		for _, rel := range Paths(files) {
			fmt.Println(filepath.Join(root, rel))
		}
		if openSpec {
			fmt.Println("OpenSpec: openspec init", root, "--tools codebuddy (requires installed OpenSpec CLI)")
		}
		return nil
	}
	if openSpec {
		if _, err := openSpecCLI(); err != nil {
			return err
		}
	}
	var createdFiles, createdDirs []string
	complete := false
	defer func() {
		if complete {
			return
		}
		for i := len(createdFiles) - 1; i >= 0; i-- {
			os.Remove(createdFiles[i])
		}
		for i := len(createdDirs) - 1; i >= 0; i-- {
			os.Remove(createdDirs[i])
		}
	}()
	for _, rel := range Paths(files) {
		target := filepath.Join(root, rel)
		var missing []string
		for dir := filepath.Dir(target); dir != root; dir = filepath.Dir(dir) {
			if _, err := os.Lstat(dir); os.IsNotExist(err) {
				missing = append(missing, dir)
			} else if err != nil {
				return err
			}
		}
		for i := len(missing) - 1; i >= 0; i-- {
			if err := os.Mkdir(missing[i], 0755); err != nil {
				return err
			}
			createdDirs = append(createdDirs, missing[i])
		}
		f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return err
		}
		createdFiles = append(createdFiles, target)
		if _, err := f.WriteString(files[rel]); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	if openSpec {
		if err := initOpenSpec(root); err != nil {
			return err
		}
	}
	complete = true
	return nil
}

func Doctor(root string) error {
	var err error
	root, err = filepath.Abs(root)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(root, ".aistd/standards.lock.json")); err != nil {
		return fmt.Errorf("missing .aistd/standards.lock.json: %w", err)
	}
	lockJSON, err := os.ReadFile(filepath.Join(root, ".aistd/standards.lock.json"))
	if err != nil {
		return err
	}
	expectedFiles, expectedLock, _, err := standardFiles()
	if err != nil {
		return err
	}
	expectedLockJSON, err := json.MarshalIndent(expectedLock, "", "  ")
	if err != nil {
		return err
	}
	if !bytes.Equal(lockJSON, append(expectedLockJSON, '\n')) {
		return fmt.Errorf("standards lock differs from the CLI's pinned v1.0 sources")
	}
	seen := make(map[string]bool)
	base := filepath.Join(root, standardsDir)
	if err := filepath.WalkDir(base, func(p string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		expected, ok := expectedFiles[rel]
		if !ok || !entry.Type().IsRegular() {
			return fmt.Errorf("unexpected standard file: %s", rel)
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if digest(content) != digest([]byte(expected)) {
			return fmt.Errorf("modified standard file: %s", rel)
		}
		seen[rel] = true
		return nil
	}); err != nil {
		return err
	}
	for rel := range expectedFiles {
		if !seen[rel] {
			return fmt.Errorf("missing standard file: %s", rel)
		}
	}
	fmt.Println("Pinned engineering and visual v1.0 source files verified. Project conformance and visual QA are not assessed.")
	return nil
}
