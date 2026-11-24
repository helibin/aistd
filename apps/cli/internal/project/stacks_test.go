package project

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeStacks(t *testing.T) {
	c, err := (Config{Name: "demo", API: true, Web: true, IOS: true, WeChat: true}).Normalize()
	if err != nil || c.APIStack != "go" || c.WebStack != "react" || c.MobileStack != "react-native" || c.MiniappStack != "taro" {
		t.Fatalf("defaults: %+v %v", c, err)
	}
	plain, err := (Config{Name: "plain"}).Normalize()
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(plain)
	if strings.Contains(string(encoded), "Stack") {
		t.Fatal("disabled stacks must be omitted")
	}
	for _, c := range []Config{
		{Name: "demo", APIStack: "go"}, {Name: "demo", WebStack: "vue"}, {Name: "demo", MobileStack: "flutter"}, {Name: "demo", MiniappStack: "uniapp"},
		{Name: "demo", API: true, APIStack: "rust"}, {Name: "demo", Web: true, WebStack: "angular"}, {Name: "demo", IOS: true, MobileStack: "swift"}, {Name: "demo", WeChat: true, MiniappStack: "native"},
	} {
		if err := c.Validate(); err == nil {
			t.Fatalf("accepted invalid stack: %+v", c)
		}
	}
}

func TestStackGenerationMatrix(t *testing.T) {
	for _, tc := range []struct {
		name         string
		config       Config
		want, absent []string
	}{
		{"nestjsvue", Config{API: true, Web: true, APIStack: "nestjs", WebStack: "vue"}, []string{"apps/api/src/main.ts", "apps/api/test/http.test.cjs", "apps/web/src/App.vue"}, []string{"apps/api/go.mod", "apps/api/cmd/server/main.go", "apps/web/src/main.tsx"}},
		{"nestjsreact", Config{API: true, Web: true, APIStack: "nestjs"}, []string{"apps/api/src/main.ts", "apps/web/src/main.tsx"}, []string{"apps/api/go.mod", "apps/web/src/App.vue"}},
		{"govue", Config{API: true, Web: true, WebStack: "vue"}, []string{"apps/api/go.mod", "apps/web/src/App.vue"}, []string{"apps/api/src/main.ts", "apps/web/src/main.tsx"}},
		{"vueonly", Config{Web: true, WebStack: "vue"}, []string{"apps/web/src/App.vue"}, []string{"apps/api/go.mod", "apps/api/src/main.ts"}},
		{"flutter", Config{IOS: true, Android: true, MobileStack: "flutter"}, []string{"apps/mobile/pubspec.yaml"}, []string{"package.json", "apps/mobile/package.json"}},
		{"uniapp", Config{WeChat: true, Alipay: true, MiniappStack: "uniapp"}, []string{"apps/miniapp/package.json"}, []string{"apps/mobile/package.json"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.config
			c.Name = tc.name
			files, err := Files(c)
			if err != nil {
				t.Fatal(err)
			}
			for _, path := range tc.want {
				if _, ok := files[path]; !ok {
					t.Errorf("missing %s", path)
				}
			}
			for _, path := range tc.absent {
				if _, ok := files[path]; ok {
					t.Errorf("unexpected %s", path)
				}
			}
			var normalized Config
			if err := json.Unmarshal([]byte(files[".aistd/project.json"]), &normalized); err != nil {
				t.Fatal(err)
			}
			if err := normalized.Validate(); err != nil {
				t.Fatal(err)
			}
			makefile := files["Makefile"]
			if strings.Contains(makefile, "go vet") != (c.API && c.APIStack != "nestjs") {
				t.Fatal("Go requirement disagrees with backend")
			}
			var root struct {
				Workspaces []string `json:"workspaces"`
			}
			if content, ok := files["package.json"]; ok {
				if err := json.Unmarshal([]byte(content), &root); err != nil {
					t.Fatal(err)
				}
			}
			for _, app := range []string{"api", "web", "mobile", "miniapp"} {
				_, has := files["apps/"+app+"/package.json"]
				found := false
				for _, workspace := range root.Workspaces {
					if workspace == "apps/"+app {
						found = true
					}
				}
				if found != has {
					t.Errorf("workspace %s mismatches generated package", app)
				}
			}
			if c.APIStack == "nestjs" {
				if !strings.Contains(files["deploy/compose.yaml"], `"node", "scripts/healthcheck.cjs"`) || strings.Contains(files["README.md"], "Go API") {
					t.Fatal("NestJS healthcheck or copy is wrong")
				}
			}
			if c.WebStack == "vue" {
				if strings.Contains(files["apps/web/src/App.vue"], "fetch('/api/hello')") != c.API || strings.Contains(files["apps/web/vite.config.ts"], "proxy:") != c.API {
					t.Fatal("Vue API integration mismatches selection")
				}
			}
			if c.MobileStack == "flutter" {
				for _, command := range []string{"flutter pub get", "flutter analyze && flutter test", "flutter build apk --debug", "flutter build ios --debug --no-codesign", "flutter run"} {
					if !strings.Contains(makefile, command) {
						t.Errorf("missing Flutter command %s", command)
					}
				}
				if strings.Contains(files[".github/workflows/check.yml"], "setup-node") || !strings.Contains(files[".github/workflows/check.yml"], "macos-latest") {
					t.Fatal("Flutter CI toolchain wrong")
				}
				for _, tool := range []string{"setup-java@", "java-version: '17'", "setup-android@", "flutter-action@"} {
					if !strings.Contains(files[".github/workflows/check.yml"], tool) {
						t.Errorf("Flutter CI missing toolchain %s", tool)
					}
				}
			}
		})
	}
}
