package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuickStartMatchesSelectedApplications(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config Config
		want   []string
		absent []string
	}{
		{"default", Config{API: true, Web: true}, []string{"Go 1.23+", "Node.js 22", "make setup", "make run-api", "make run-web", "separate terminal", "Go API: hello", "make run` starts only the API"}, []string{"make run-ios"}},
		{"api", Config{API: true}, []string{"Go 1.23+", "make run-api", "8080/api/hello"}, []string{"Node.js", "make run-web"}},
		{"web", Config{Web: true}, []string{"Node.js 22", "make run-web", "127.0.0.1:5173"}, []string{"Go 1.23+", "make run-api", "Go API: hello"}},
		{"mobile", Config{IOS: true, Android: true}, []string{"make run-ios", "Xcode", "make run-android", "emulator"}, []string{"make run-api", "make run-web"}},
		{"miniapps", Config{WeChat: true, Alipay: true}, []string{"make run-wechat", "dist/weapp", "make run-alipay", "dist/alipay", "App ID"}, []string{"make run-api", "make run-web"}},
		{"plain", Config{}, []string{"No application was selected", "aistd doctor --dir ."}, []string{"make setup", "make run", "Requirements:"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.config.Name = "demo"
			guide := QuickStart(tc.config)
			for _, want := range tc.want {
				if !strings.Contains(guide, want) {
					t.Errorf("guide missing %q", want)
				}
			}
			for _, absent := range tc.absent {
				if strings.Contains(guide, absent) {
					t.Errorf("guide includes unselected requirement/command %q", absent)
				}
			}
			files, err := Files(tc.config)
			if err != nil || !strings.Contains(files["README.md"], guide) {
				t.Fatalf("README did not include shared quick start: %v", err)
			}
		})
	}
}

func TestInvalidNameCannotEscape(t *testing.T) {
	for _, name := range []string{"../escape", "/tmp/x", "", "Upper", "x/y"} {
		if _, err := Files(Config{Name: name}); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}
}

func TestAPITemplatesFollowSelection(t *testing.T) {
	for _, tc := range []struct {
		name     string
		api, web bool
	}{
		{"api", true, false}, {"apiweb", true, true}, {"web", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, err := Files(Config{Name: tc.name, API: tc.api, Web: tc.web})
			if err != nil {
				t.Fatal(err)
			}
			for _, rel := range []string{
				"cmd/server/main.go", "cmd/server/main_test.go",
				"internal/config/config.go", "internal/config/config_test.go",
				"internal/httpapi/handler.go", "internal/httpapi/handler_test.go",
				"internal/httpapi/server.go", "internal/httpapi/server_test.go", "Dockerfile", ".dockerignore",
			} {
				content, ok := files["apps/api/"+rel]
				if ok != tc.api {
					t.Fatalf("API template %s presence disagrees with selection", rel)
				}
				if strings.Contains(content, "example.com/PROJECT") {
					t.Fatalf("unexpanded module in %s", rel)
				}
			}
			if !tc.api {
				return
			}
			if !strings.Contains(files["apps/api/cmd/server/main.go"], "example.com/"+tc.name+"/api/internal/config") {
				t.Fatal("API imports do not match the project module")
			}
			if !strings.Contains(files["apps/api/Dockerfile"], "COPY . .") {
				t.Fatal("Docker build omits module files")
			}
			for _, excluded := range []string{".env\n", ".env.*\n", "!.env.example\n", "bin\n", "dist\n", ".git\n"} {
				if !strings.Contains(files["apps/api/.dockerignore"], excluded) {
					t.Fatalf("Docker context missing exclusion %q", excluded)
				}
			}
			if !strings.Contains(files["Makefile"], "gofmt -l apps/api/cmd apps/api/internal") ||
				!strings.Contains(files["scripts/check.sh"], "gofmt -l ./cmd ./internal") {
				t.Fatal("format checks omit internal packages")
			}
		})
	}
}

func TestCreateRefusesExistingDirectory(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "keep")
	if err := os.WriteFile(p, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Create(root, Config{Name: "demo"}, false); err == nil {
		t.Fatal("expected refusal")
	}
	b, err := os.ReadFile(p)
	if err != nil || string(b) != "original" {
		t.Fatal("existing file changed", err)
	}
}

func TestDryRunDoesNotWrite(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo")
	if err := Create(root, Config{Name: "demo", API: true, Visual: true}, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("dry run wrote target")
	}
}

func TestCreateInstallsBothStandardsForEveryOption(t *testing.T) {
	for _, tc := range []struct {
		name   string
		api    bool
		web    bool
		visual bool
	}{
		{"plain", false, false, false},
		{"api", true, false, false},
		{"web", false, true, false},
		{"visual", false, false, true},
		{"apps", true, true, false},
		{"all", true, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), tc.name)
			if err := Create(root, Config{Name: tc.name, API: tc.api, Web: tc.web, Visual: tc.visual}, false); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"engineering", "visual"} {
				p := filepath.Join(root, standardsDir, name, "v1.0", "AGENTS.md")
				if _, err := os.Stat(p); err != nil {
					t.Fatalf("%s standard missing: %v", name, err)
				}
			}
			if err := Doctor(root); err != nil {
				t.Fatal(err)
			}
			plan, err := os.ReadFile(filepath.Join(root, "task.md"))
			if err != nil || !strings.Contains(string(plan), "## Planned") || !strings.Contains(string(plan), "## Blocked") {
				t.Fatalf("editable development plan missing: %v", err)
			}
			if err := os.WriteFile(filepath.Join(root, "task.md"), []byte("# My plan\n"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := Doctor(root); err != nil {
				t.Fatalf("doctor rejected editable task.md: %v", err)
			}
			for _, p := range []string{"apps/api/cmd/server/main.go", "contracts/openapi.yaml", ".github/workflows/check.yml"} {
				_, err := os.Stat(filepath.Join(root, p))
				want := tc.api
				if p == ".github/workflows/check.yml" {
					want = tc.api || tc.web
				}
				if (err == nil) != want {
					t.Fatalf("file %s presence disagrees with options: %v", p, err)
				}
			}
			for _, p := range []string{"apps/web/package.json", "apps/web/src/main.tsx", "apps/web/vite.config.ts"} {
				_, err := os.Stat(filepath.Join(root, p))
				if (err == nil) != tc.web {
					t.Fatalf("web file %s presence disagrees with option: %v", p, err)
				}
			}
			makefile, err := os.ReadFile(filepath.Join(root, "Makefile"))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(makefile), "npm run build --workspace apps/web") != tc.web {
				t.Fatalf("web build command disagrees with option: %s", makefile)
			}
			if tc.api || tc.web {
				if !strings.Contains(string(makefile), "run: run-") || !strings.Contains(string(makefile), "docker-build: docker-build-") {
					t.Fatalf("starter command aliases missing: %s", makefile)
				}
			}
			if tc.web {
				ignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
				if err != nil || !strings.Contains(string(ignore), "**/node_modules/") || !strings.Contains(string(ignore), "**/dist/") {
					t.Fatalf("web build directories are not ignored: %v", err)
				}
				main, err := os.ReadFile(filepath.Join(root, "apps/web/src/main.tsx"))
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(main), "fetch('/api/hello')") != tc.api {
					t.Fatalf("web API call disagrees with API option: %s", main)
				}
				viteConfig, err := os.ReadFile(filepath.Join(root, "apps/web/vite.config.ts"))
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(viteConfig), "proxy:") != tc.api {
					t.Fatalf("web API proxy disagrees with API option: %s", viteConfig)
				}
				workflow, err := os.ReadFile(filepath.Join(root, ".github/workflows/check.yml"))
				if err != nil || !strings.Contains(string(workflow), "actions/setup-node@v4") {
					t.Fatalf("web workflow missing Node setup: %v", err)
				}
			}
			_, err = os.Stat(filepath.Join(root, "design/tokens/brand.json"))
			if (err == nil) != tc.visual {
				t.Fatalf("visual token presence disagrees with option: %v", err)
			}
			lockJSON, err := os.ReadFile(filepath.Join(root, ".aistd/standards.lock.json"))
			if err != nil {
				t.Fatal(err)
			}
			var lock standardsLock
			if err := json.Unmarshal(lockJSON, &lock); err != nil {
				t.Fatal(err)
			}
			if lock.Status != "sources-installed" || len(lock.Engineering.Files) != 51 || len(lock.Visual.Files) != 12 {
				t.Fatalf("incomplete standards lock: %+v", lock)
			}
		})
	}
}

func TestDoctorRejectsStandardChanges(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(string) error
	}{
		{"changed file", func(root string) error {
			return os.WriteFile(filepath.Join(root, standardsDir, "visual/v1.0/docs/VISUAL-STANDARD.md"), []byte("changed"), 0644)
		}},
		{"missing file", func(root string) error {
			return os.Remove(filepath.Join(root, standardsDir, "engineering/v1.0/docs/STANDARD.md"))
		}},
		{"extra file", func(root string) error {
			return os.WriteFile(filepath.Join(root, standardsDir, "visual/v1.0/extra.md"), []byte("extra"), 0644)
		}},
		{"changed lock", func(root string) error {
			p := filepath.Join(root, ".aistd/standards.lock.json")
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			return os.WriteFile(p, []byte(strings.Replace(string(b), "sources-installed", "not-installed", 1)), 0644)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "demo")
			if err := Create(root, Config{Name: "demo"}, false); err != nil {
				t.Fatal(err)
			}
			if err := tc.edit(root); err != nil {
				t.Fatal(err)
			}
			if err := Doctor(root); err == nil {
				t.Fatal("doctor accepted modified standard sources")
			}
		})
	}
}

func TestRootOpsFilesFollowSelection(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config Config
	}{
		{"plain", Config{}},
		{"api", Config{API: true}},
		{"web", Config{Web: true}},
		{"apiweb", Config{API: true, Web: true}},
		{"apps", Config{API: true, Web: true, IOS: true, Android: true, WeChat: true, Alipay: true}},
		{"mobileonly", Config{IOS: true, WeChat: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.config
			c.Name = "ops" + strings.ReplaceAll(tc.name, "-", "")
			files, err := Files(c)
			if err != nil {
				t.Fatal(err)
			}
			editor, ok := files[".editorconfig"]
			if !ok || !strings.Contains(editor, "root = true") {
				t.Fatalf("editor config missing or invalid: %v", err)
			}
			rootPackage, hasRoot := files["package.json"]
			wantJS := c.Web || c.IOS || c.Android || c.WeChat || c.Alipay
			if hasRoot != wantJS {
				t.Fatalf("root package.json presence disagrees with JS apps")
			}
			makefile := files["Makefile"]
			if strings.Contains(makefile, "npm --prefix") {
				t.Fatalf("makefile uses per-package npm instead of workspaces: %s", makefile)
			}
			if strings.Contains(makefile, "npm install --no-audit --no-fund") != wantJS {
				t.Fatalf("workspace setup disagrees with JS apps: %s", makefile)
			}
			if wantJS {
				var parsed struct {
					Name       string   `json:"name"`
					Private    bool     `json:"private"`
					Workspaces []string `json:"workspaces"`
				}
				if err := json.Unmarshal([]byte(rootPackage), &parsed); err != nil {
					t.Fatal(err)
				}
				if parsed.Name != c.Name || !parsed.Private {
					t.Fatalf("invalid root package.json: %s", rootPackage)
				}
				var want []string
				if c.Web {
					want = append(want, "apps/web")
				}
				if c.IOS || c.Android {
					want = append(want, "apps/mobile")
				}
				if c.WeChat || c.Alipay {
					want = append(want, "apps/miniapp")
				}
				if strings.Join(parsed.Workspaces, ",") != strings.Join(want, ",") {
					t.Fatalf("workspaces disagree with selection: %v", parsed.Workspaces)
				}
			}
			ops := c.API || c.Web
			if _, ok := files["deploy/compose.yaml"]; ok != ops {
				t.Fatalf("compose presence disagrees with selection")
			}
			if _, ok := files["scripts/size-report.sh"]; ok != ops {
				t.Fatalf("size report presence disagrees with selection")
			}
			if strings.Contains(makefile, "docker compose --env-file deploy/.env") != ops {
				t.Fatalf("compose targets disagree with selection: %s", makefile)
			}

			if strings.Contains(makefile, "release-image") != ops {
				t.Fatalf("release target disagrees with selection: %s", makefile)
			}
			if !ops {
				for _, p := range []string{"deploy/.env.example", "deploy/caddy/Caddyfile", "scripts/check.sh", "scripts/build.sh", "scripts/publish.sh"} {
					if _, ok := files[p]; ok {
						t.Fatalf("unexpected root ops file without api/web: %s", p)
					}
				}
				return
			}
			compose := files["deploy/compose.yaml"]
			if !strings.Contains(compose, "name: "+c.Name) {
				t.Fatalf("compose project name missing: %s", compose)
			}
			if strings.Contains(compose, "  api:") != c.API || strings.Contains(compose, "  web:") != c.Web {
				t.Fatalf("compose services disagree with selection: %s", compose)
			}
			if !strings.Contains(compose, "  edge:") || strings.Contains(compose, "healthcheck") != c.API {
				t.Fatalf("compose edge/healthcheck disagree with selection: %s", compose)
			}
			caddy := files["deploy/caddy/Caddyfile"]
			if strings.Contains(caddy, "/api/*") != (c.API && c.Web) {
				t.Fatalf("caddy api route disagrees with selection: %s", caddy)
			}
			if strings.Contains(caddy, "reverse_proxy api:8080") != c.API || strings.Contains(caddy, "reverse_proxy web:80") != c.Web {
				t.Fatalf("caddy upstreams disagree with selection: %s", caddy)
			}
			if env := files["deploy/.env.example"]; !strings.Contains(env, "PUBLIC_PORT=") {
				t.Fatalf("compose env example missing port: %s", env)
			}
			check := files["scripts/check.sh"]
			if strings.Contains(check, "gofmt -l ./cmd") != c.API || strings.Contains(check, "npm run lint --workspace apps/web") != c.Web {
				t.Fatalf("check script disagrees with selection: %s", check)
			}
			if strings.Contains(check, "npm run check --workspace apps/mobile") != (c.IOS || c.Android) || strings.Contains(check, "npm run check --workspace apps/miniapp") != (c.WeChat || c.Alipay) {
				t.Fatalf("check script platforms disagree with selection: %s", check)
			}
			build := files["scripts/build.sh"]
			if strings.Contains(build, "-o ../../dist/api") != c.API || strings.Contains(build, "npm run build --workspace apps/web") != c.Web {
				t.Fatalf("build script disagrees with selection: %s", build)
			}
			if strings.Contains(build, "build:ios") != c.IOS || strings.Contains(build, "build:android") != c.Android || strings.Contains(build, "build:weapp") != c.WeChat || strings.Contains(build, "build:alipay") != c.Alipay {
				t.Fatalf("build script platforms disagree with selection: %s", build)
			}
			publish := files["scripts/publish.sh"]
			if strings.Contains(publish, "\"$REGISTRY/api:$IMAGE_TAG\"") != c.API || strings.Contains(publish, "\"$REGISTRY/web:$IMAGE_TAG\"") != c.Web {
				t.Fatalf("publish script disagrees with selection: %s", publish)
			}
			readme := files["README.md"]
			if !strings.Contains(readme, "make up") || !strings.Contains(readme, "deploy/.env.example") {
				t.Fatalf("readme orchestration section missing: %s", readme)
			}
		})
	}
}

func TestCreateWritesExecutableOpsScripts(t *testing.T) {
	root := filepath.Join(t.TempDir(), "demo")
	if err := Create(root, Config{Name: "demo", API: true, Web: true}, false); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"scripts/check.sh", "scripts/build.sh", "scripts/publish.sh", "scripts/size-report.sh"} {
		info, err := os.Stat(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&0111 == 0 {
			t.Fatalf("%s is not executable", rel)
		}
	}
	for _, rel := range []string{".editorconfig", "deploy/compose.yaml", "deploy/.env.example", "deploy/caddy/Caddyfile"} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("%s missing: %v", rel, err)
		}
	}
	if err := Doctor(root); err != nil {
		t.Fatal(err)
	}
}

func TestPlatformFilesFollowSelection(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config Config
		want   []string
		absent []string
	}{
		{"ios", Config{IOS: true}, []string{"apps/mobile/App.tsx", "build:ios"}, []string{"build:android", "apps/miniapp/package.json"}},
		{"android", Config{Android: true}, []string{"apps/mobile/App.tsx", "build:android"}, []string{"build:ios", "apps/miniapp/package.json"}},
		{"wechat", Config{WeChat: true}, []string{"apps/miniapp/project.config.json", "build:weapp"}, []string{"build:alipay", "apps/mobile/App.tsx"}},
		{"alipay", Config{Alipay: true}, []string{"apps/miniapp/package.json", "build:alipay"}, []string{"build:weapp", "apps/miniapp/project.config.json"}},
		{"all", Config{IOS: true, Android: true, WeChat: true, Alipay: true}, []string{"build:ios", "build:android", "build:weapp", "build:alipay"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.config
			c.Name = tc.name
			files, err := Files(c)
			if err != nil {
				t.Fatal(err)
			}
			all := files["Makefile"] + files["apps/mobile/package.json"] + files["apps/miniapp/package.json"]
			for _, item := range tc.want {
				if _, ok := files[item]; !ok && !strings.Contains(all, item) {
					t.Fatalf("missing selected target %s", item)
				}
			}
			for _, item := range tc.absent {
				if _, ok := files[item]; ok || strings.Contains(all, item) {
					t.Fatalf("included unselected target %s", item)
				}
			}
		})
	}
}

func TestAdoptNeverOverwrites(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "AGENTS.md")
	if err := os.WriteFile(existing, []byte("keep me\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Adopt(root, true); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "AGENTS.md" {
		t.Fatalf("dry run changed directory entries: %v", entries)
	}
	if err := Adopt(root, false); err != nil {
		t.Fatal(err)
	}
	if err := Doctor(root); err != nil {
		t.Fatal(err)
	}
	plan, err := os.ReadFile(filepath.Join(root, "task.md"))
	if err != nil || !strings.Contains(string(plan), "## In progress") {
		t.Fatalf("adoption did not create development plan: %v", err)
	}
	content, err := os.ReadFile(existing)
	if err != nil || string(content) != "keep me\n" {
		t.Fatalf("existing instructions changed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".aistd/project.json")); !os.IsNotExist(err) {
		t.Fatal("adoption created a new-project configuration")
	}
	if err := Adopt(root, false); err == nil {
		t.Fatal("expected exclusive create failure")
	}
	if err := Doctor(root); err != nil {
		t.Fatalf("rejected adoption changed installed standards: %v", err)
	}
	content, err = os.ReadFile(existing)
	if err != nil || string(content) != "keep me\n" {
		t.Fatalf("rejected adoption changed existing instructions: %v", err)
	}
}

func TestAdoptPreservesExistingTaskPlan(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "task.md")
	original := "# Existing roadmap\n\n- [ ] Ship feature\n"
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Adopt(root, true); err != nil {
		t.Fatal(err)
	}
	if err := Adopt(root, false); err != nil {
		t.Fatal(err)
	}
	plan, err := os.ReadFile(path)
	if err != nil || string(plan) != original {
		t.Fatalf("existing task.md changed: %v", err)
	}
	if err := Doctor(root); err != nil {
		t.Fatalf("doctor rejected existing plan: %v", err)
	}
}

func TestAdoptRejectsExistingManagedFilesBeforeWriting(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".aistd"), 0755); err != nil {
		t.Fatal(err)
	}
	lock := filepath.Join(root, ".aistd/standards.lock.json")
	if err := os.WriteFile(lock, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Adopt(root, false); err == nil {
		t.Fatal("expected managed file conflict")
	}
	content, err := os.ReadFile(lock)
	if err != nil || string(content) != "existing" {
		t.Fatalf("existing lock changed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "AISTD-ADOPTION.md")); !os.IsNotExist(err) {
		t.Fatal("adoption wrote files before checking conflicts")
	}
}

func TestOptionalOpenSpecIntegration(t *testing.T) {
	bin := t.TempDir()
	invocations := filepath.Join(t.TempDir(), "calls")
	stub := filepath.Join(bin, "openspec")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$OPENSPEC_CALLS\"\nmkdir -p \"$2/openspec\" \"$2/.codebuddy\"\nprintf 'schema: spec-driven\\n' > \"$2/openspec/config.yaml\"\n"
	if err := os.WriteFile(stub, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("OPENSPEC_CALLS", invocations)
	parent := t.TempDir()
	plain := filepath.Join(parent, "plain")
	if err := Create(plain, Config{Name: "plain"}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(plain, "openspec")); !os.IsNotExist(err) {
		t.Fatal("default unexpectedly initialized OpenSpec")
	}
	root := filepath.Join(parent, "enabled")
	if err := Create(root, Config{Name: "enabled", OpenSpec: true}, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("dry run wrote project")
	}
	if err := Create(root, Config{Name: "enabled", OpenSpec: true}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "openspec/config.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := Doctor(root); err != nil {
		t.Fatal(err)
	}
	adopt := filepath.Join(parent, "existing")
	if err := os.Mkdir(adopt, 0755); err != nil {
		t.Fatal(err)
	}
	if err := AdoptWithOpenSpec(adopt, true, true); err != nil {
		t.Fatal(err)
	}
	if entries, err := os.ReadDir(adopt); err != nil || len(entries) != 0 {
		t.Fatalf("adopt dry run wrote files: %v", err)
	}
	if err := AdoptWithOpenSpec(adopt, false, true); err != nil {
		t.Fatal(err)
	}
	calls, err := os.ReadFile(invocations)
	if err != nil || string(calls) != "init "+root+" --tools codebuddy\ninit "+adopt+" --tools codebuddy\n" {
		t.Fatalf("unexpected official CLI calls: %q: %v", calls, err)
	}
}

func TestAdoptOpenSpecConflictsBeforeWriting(t *testing.T) {
	for _, rel := range []string{"openspec", ".codebuddy"} {
		t.Run(rel, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, rel), 0755); err != nil {
				t.Fatal(err)
			}
			if err := AdoptWithOpenSpec(root, false, true); err == nil {
				t.Fatal("existing integration was accepted")
			}
			if _, err := os.Stat(filepath.Join(root, "AISTD-ADOPTION.md")); !os.IsNotExist(err) {
				t.Fatal("conflict wrote adoption files")
			}
		})
	}
}

func TestMissingOpenSpecCLILeavesProjectUntouched(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	parent := t.TempDir()
	root := filepath.Join(parent, "demo")
	if err := Create(root, Config{Name: "demo", OpenSpec: true}, false); err == nil {
		t.Fatal("missing CLI accepted")
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("missing CLI created project")
	}
	if err := AdoptWithOpenSpec(parent, false, true); err == nil {
		t.Fatal("missing CLI accepted for existing project")
	}
	if entries, err := os.ReadDir(parent); err != nil || len(entries) != 0 {
		t.Fatalf("missing CLI wrote files: %v", err)
	}
}

func TestAdoptAcceptsRelativeDirectory(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "existing")
	if err := os.Mkdir(root, 0755); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(cwd, root)
	if err != nil {
		t.Fatal(err)
	}
	if err := Adopt(rel, false); err != nil {
		t.Fatal(err)
	}
	if err := Doctor(rel); err != nil {
		t.Fatal(err)
	}
}
