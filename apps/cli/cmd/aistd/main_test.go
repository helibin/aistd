package main

import (
	"aistd"
	"aistd/internal/project"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateNextStepsAndDryRun(t *testing.T) {
	for _, dry := range []bool{false, true} {
		t.Run(map[bool]string{false: "create", true: "dry-run"}[dry], func(t *testing.T) {
			parent := filepath.Join(t.TempDir(), "space ' quote $(false)")
			if err := os.Mkdir(parent, 0755); err != nil {
				t.Fatal(err)
			}
			output, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			original := os.Stdout
			os.Stdout = output
			t.Cleanup(func() { os.Stdout = original })
			args := []string{"create", "--name", "demo", "--dir", parent}
			if dry {
				args = append(args, "--dry-run")
			}
			err = run(args)
			os.Stdout = original
			if err != nil {
				t.Fatal(err)
			}
			content, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			text := string(content)
			for _, marker := range []string{"Created:", "Next, enter", "make setup", "make run-api", "make run-web", "See README.md"} {
				if strings.Contains(text, marker) == dry {
					t.Errorf("unexpected output marker %q (dry=%v)", marker, dry)
				}
			}
			root := filepath.Join(parent, "demo")
			if dry {
				if _, err := os.Stat(root); !os.IsNotExist(err) {
					t.Fatal("dry-run created a project")
				}
				return
			}
			for _, line := range strings.Split(text, "\n") {
				if strings.HasPrefix(line, "  cd -- ") {
					cmd := exec.Command("sh", "-c", line+" && test -f .aistd/project.json")
					if result, err := cmd.CombinedOutput(); err != nil {
						t.Fatalf("printed cd command failed: %v: %s", err, result)
					}
					return
				}
			}
			t.Fatal("missing cd command")
		})
	}
}

func TestUninstallRefusesDifferentExecutable(t *testing.T) {
	t.Setenv("AISTD_INSTALL_DIR", t.TempDir())
	if err := run([]string{"uninstall", "--yes"}); err == nil || !strings.Contains(err.Error(), "does not match installer path") {
		t.Fatalf("expected refusal for non-installed executable, got %v", err)
	}
	if err := run([]string{"uninstall", "extra"}); err == nil {
		t.Fatal("expected positional argument rejection")
	}
}

func TestCreateWithNameNeedsNoConfiguration(t *testing.T) {
	parent := t.TempDir()
	if err := run([]string{"create", "--name", "demo", "--dir", parent}); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"apps/api/go.mod", "apps/web/package.json", "design/tokens/brand.json"} {
		if _, err := os.Stat(filepath.Join(parent, "demo", rel)); err != nil {
			t.Fatalf("default project missing %s: %v", rel, err)
		}
	}
	generation, err := os.ReadFile(filepath.Join(parent, "demo", ".aistd/generation.json"))
	if err != nil {
		t.Fatal(err)
	}
	var metadata struct {
		CLI string `json:"cli"`
	}
	if err := json.Unmarshal(generation, &metadata); err != nil || metadata.CLI != aistd.Version {
		t.Fatalf("generated CLI version must match VERSION: %s (%v)", generation, err)
	}
	if err := run([]string{"doctor", "--dir", filepath.Join(parent, "demo")}); err != nil {
		t.Fatal(err)
	}
}

func TestCreateNameRejectsInvalidInputBeforeWriting(t *testing.T) {
	parent := t.TempDir()
	if err := run([]string{"create", "--name", "../outside", "--dir", parent}); err == nil {
		t.Fatal("expected invalid project name")
	}
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 0 {
		t.Fatalf("invalid name wrote to parent: %v", err)
	}
}

func TestTargetSelection(t *testing.T) {
	parent := t.TempDir()
	if err := run([]string{"create", "--name", "mobile", "--targets", "ios,wechat", "--visual=false", "--dir", parent}); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"apps/mobile/App.tsx", "apps/miniapp/src/app.ts", "apps/miniapp/project.config.json"} {
		if _, err := os.Stat(filepath.Join(parent, "mobile", rel)); err != nil {
			t.Fatalf("selected target missing %s: %v", rel, err)
		}
	}
	for _, rel := range []string{"apps/api/go.mod", "apps/web/package.json", "design/tokens/brand.json"} {
		if _, err := os.Stat(filepath.Join(parent, "mobile", rel)); !os.IsNotExist(err) {
			t.Fatalf("unselected target created %s: %v", rel, err)
		}
	}
	if err := run([]string{"doctor", "--dir", filepath.Join(parent, "mobile")}); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"ios,3", "ios,unknown", "ios,", ""} {
		if err := selectTargets(&project.Config{}, input); err == nil {
			t.Fatalf("accepted invalid targets %q", input)
		}
	}
	if err := run([]string{"create", "--name", "invalid", "--targets", "ios,3", "--dir", parent}); err == nil {
		t.Fatal("accepted duplicate target aliases")
	}
	if _, err := os.Stat(filepath.Join(parent, "invalid")); !os.IsNotExist(err) {
		t.Fatal("invalid target selection wrote a project")
	}
	if err := run([]string{"create", "--config", "fixture.json", "--visual=true"}); err == nil {
		t.Fatal("accepted conflicting configuration flags")
	}
}

func TestOpenSpecFlagAndJSONConfig(t *testing.T) {
	parent := t.TempDir()
	if err := run([]string{"create", "--name", "demo", "--openspec", "--dry-run", "--dir", parent}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(parent, "demo")); !os.IsNotExist(err) {
		t.Fatal("OpenSpec dry run wrote project")
	}
	config := filepath.Join(parent, "project.json")
	if err := os.WriteFile(config, []byte(`{"name":"configured","openspec":true}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"create", "--config", config, "--dry-run", "--dir", parent}); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"init", "--dir", parent, "--openspec", "--dry-run"}); err != nil {
		t.Fatal(err)
	}
	if entries, err := os.ReadDir(parent); err != nil || len(entries) != 1 {
		t.Fatalf("dry run modified parent: %v", err)
	}
}

func TestInteractiveTargetSelection(t *testing.T) {
	parent := t.TempDir()
	input := filepath.Join(parent, "answers.txt")
	if err := os.WriteFile(input, []byte("interactive\n3,4,5,6\n\n\nn\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	original := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = original }()
	if err := run([]string{"create", "--dir", parent}); err != nil {
		t.Fatal(err)
	}
	makefile, err := os.ReadFile(filepath.Join(parent, "interactive", "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"build:ios", "build:android", "build:weapp", "build:alipay"} {
		if !strings.Contains(string(makefile), target) {
			t.Fatalf("missing %s in Makefile", target)
		}
	}
}

func TestStackFlagsAndInvalidCombinations(t *testing.T) {
	parent := t.TempDir()
	if err := run([]string{"create", "--name", "stacked", "--api-stack", "nestjs", "--web-stack", "vue", "--dir", parent}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"apps/api/src/main.ts", "apps/web/src/App.vue"} {
		if _, err := os.Stat(filepath.Join(parent, "stacked", path)); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{
		{"--name", "invalid", "--targets", "web", "--api-stack", "nestjs"},
		{"--name", "invalid", "--web-stack", "wrong"},
		{"--config", "missing.json", "--api-stack", ""},
		{"--config", "missing.json", "--web-stack", "vue"},
	} {
		args = append([]string{"create", "--dir", parent}, args...)
		if err := run(args); err == nil {
			t.Fatalf("accepted invalid stack flags %v", args)
		}
		if _, err := os.Stat(filepath.Join(parent, "invalid")); !os.IsNotExist(err) {
			t.Fatal("invalid flags wrote project")
		}
	}
}

func TestInteractiveAlternativeStacks(t *testing.T) {
	parent := t.TempDir()
	input := filepath.Join(parent, "answers")
	if err := os.WriteFile(input, []byte("alternative\n1,2\n2\n2\nn\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	original := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = original }()
	if err := run([]string{"create", "--dir", parent}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"apps/api/src/main.ts", "apps/web/src/App.vue"} {
		if _, err := os.Stat(filepath.Join(parent, "alternative", path)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestJSONStackDefaultsAndPersistence(t *testing.T) {
	for _, tc := range []struct {
		name, input        string
		api, web           bool
		apiStack, webStack string
	}{
		{"legacy", `{"name":"legacy"}`, true, false, "go", ""},
		{"webonly", `{"name":"webonly","api":false,"web":true}`, false, true, "", "react"},
		{"alternate", `{"name":"alternate","apiStack":"nestjs","web":true,"webStack":"vue"}`, true, true, "nestjs", "vue"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parent := t.TempDir()
			config := filepath.Join(parent, "input.json")
			if err := os.WriteFile(config, []byte(tc.input), 0644); err != nil {
				t.Fatal(err)
			}
			if err := run([]string{"create", "--config", config, "--dir", parent}); err != nil {
				t.Fatal(err)
			}
			content, err := os.ReadFile(filepath.Join(parent, tc.name, ".aistd/project.json"))
			if err != nil {
				t.Fatal(err)
			}
			var got project.Config
			if err := json.Unmarshal(content, &got); err != nil {
				t.Fatal(err)
			}
			if got.API != tc.api || got.Web != tc.web || got.APIStack != tc.apiStack || got.WebStack != tc.webStack || !got.Visual {
				t.Fatalf("unexpected persisted config: %+v", got)
			}
		})
	}
	for _, flag := range []string{"--api-stack=", "--web-stack=vue", "--mobile-stack=flutter", "--miniapp-stack=uniapp", "--admin-stack=vben"} {
		if err := run([]string{"create", "--config", "missing.json", flag}); err == nil || !strings.Contains(err.Error(), "cannot be used with --config") {
			t.Fatalf("stack flag must conflict before reading JSON: %s: %v", flag, err)
		}
	}
}

func TestAdminFlagsAndJSON(t *testing.T) {
	for _, stack := range []string{"vben", "soybean"} {
		t.Run(stack, func(t *testing.T) {
			parent := t.TempDir()
			if err := run([]string{"create", "--name", "admin", "--targets", "admin", "--admin-stack", stack, "--dir", parent}); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(parent, "admin", ".aistd/project.json"))
			if err != nil {
				t.Fatal(err)
			}
			var got project.Config
			if err := json.Unmarshal(data, &got); err != nil || !got.Admin || got.AdminStack != stack || got.API || got.Web {
				t.Fatalf("Admin CLI flags differ from persisted configuration: %+v, %v", got, err)
			}
			config := filepath.Join(parent, "admin.json")
			if err := os.WriteFile(config, []byte(`{"name":"json-admin","api":false,"admin":true,"adminStack":"`+stack+`"}`), 0644); err != nil {
				t.Fatal(err)
			}
			if err := run([]string{"create", "--config", config, "--dir", parent, "--dry-run"}); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(parent, "json-admin")); !os.IsNotExist(err) {
				t.Fatal("Admin JSON dry-run created files")
			}
		})
	}
	for _, args := range [][]string{
		{"--name", "invalid", "--targets", "web", "--admin-stack", "vben"},
		{"--name", "invalid", "--targets", "admin", "--admin-stack", "native"},
	} {
		if err := run(append([]string{"create", "--dir", t.TempDir()}, args...)); err == nil {
			t.Fatalf("accepted invalid Admin flags: %v", args)
		}
	}
	c := project.Config{API: true, Web: true}
	if err := selectTargets(&c, "7"); err != nil || !c.Admin || c.API || c.Web {
		t.Fatalf("Admin numeric target did not reset old selection: %+v, %v", c, err)
	}
	if err := selectTargets(&c, "7,admin"); err == nil {
		t.Fatal("accepted duplicate Admin aliases")
	}
}

func TestPipedAdminTemplateSelection(t *testing.T) {
	parent := t.TempDir()
	input := filepath.Join(parent, "answers")
	if err := os.WriteFile(input, []byte("admin-piped\n7\n2\nn\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	original := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = original }()
	if err := run([]string{"create", "--dir", parent}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(parent, "admin-piped", ".aistd/project.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got project.Config
	if err := json.Unmarshal(data, &got); err != nil || !got.Admin || got.AdminStack != "soybean" || got.API || got.Web || got.Visual {
		t.Fatalf("piped Admin choice differs from generated project: %+v, %v", got, err)
	}
}
