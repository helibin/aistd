package project

import (
	"aistd/internal/admintemplates"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminConfiguration(t *testing.T) {
	c, err := (Config{Name: "legacy"}).Normalize()
	if err != nil || c.Admin || c.AdminStack != "" {
		t.Fatalf("legacy configuration enabled Admin: %+v, %v", c, err)
	}
	c, err = (Config{Name: "admin", Admin: true}).Normalize()
	if err != nil || c.AdminStack != "vben" || c.API || c.Web {
		t.Fatalf("invalid independent Admin defaults: %+v, %v", c, err)
	}
	for _, c := range []Config{{Name: "admin", AdminStack: "vben"}, {Name: "admin", Admin: true, AdminStack: "native"}} {
		if _, err := c.Normalize(); err == nil {
			t.Fatalf("accepted invalid Admin configuration: %+v", c)
		}
	}
}

func TestAdminFilesAndRuntimeChain(t *testing.T) {
	for _, stack := range []string{"vben", "soybean"} {
		t.Run(stack, func(t *testing.T) {
			for _, web := range []bool{false, true} {
				c := Config{Name: "admin", Admin: true, AdminStack: stack, Web: web}
				files, err := Files(c)
				if err != nil {
					t.Fatal(err)
				}
				for _, path := range []string{"apps/admin/package.json", "apps/admin/pnpm-lock.yaml", "apps/admin/LICENSE", ".aistd/admin-template.json", "apps/admin/AISTD.Dockerfile", ".github/workflows/check.yml"} {
					if files[path] == "" {
						t.Fatalf("missing Admin source or runtime file: %s", path)
					}
				}
				var config Config
				if err := json.Unmarshal([]byte(files[".aistd/project.json"]), &config); err != nil || !config.Admin || config.AdminStack != stack {
					t.Fatalf("persisted Admin configuration differs: %+v, %v", config, err)
				}
				if strings.Contains(files["package.json"], "apps/admin") || (!web && files["package.json"] != "") {
					t.Fatal("independent pnpm template was included in root npm workspace")
				}
				for path, commands := range map[string][]string{
					"Makefile":                    {"CI=1 SKIP_INSTALL_SIMPLE_GIT_HOOKS=1 " + adminPNPM + " install --frozen-lockfile", adminCheck(c), adminBuild(c), adminRun(c), "docker-build-admin:"},
					"scripts/check.sh":            {adminCheck(c)},
					"scripts/build.sh":            {adminBuild(c)},
					"scripts/publish.sh":          {"apps/admin/AISTD.Dockerfile", "$REGISTRY/admin:$IMAGE_TAG"},
					"apps/admin/AISTD.Dockerfile": {"ENV CI=1 SKIP_INSTALL_SIMPLE_GIT_HOOKS=1"},
				} {
					for _, command := range commands {
						if !strings.Contains(files[path], command) {
							t.Fatalf("%s does not run Admin command %q", path, command)
						}
					}
				}
				if !web && !strings.Contains(files["Makefile"], "run: run-admin\n") {
					t.Fatal("Admin-only project has no default run target")
				}
				if !strings.Contains(files["deploy/compose.yaml"], "  admin:\n") || !strings.Contains(files["deploy/caddy/Caddyfile"], "@admin path /admin /admin/*") {
					t.Fatal("Admin static deployment is not routed independently")
				}
				if strings.Contains(files["deploy/caddy/Caddyfile"], "reverse_proxy web:80") != web {
					t.Fatal("Admin route changed Web selection")
				}
				quickStart := QuickStart(c)
				if !strings.Contains(quickStart, "Node.js 22.18+") || !strings.Contains(quickStart, "make run-admin") || strings.Contains(quickStart, "No application was selected") {
					t.Fatalf("Admin quick start missing: %s", quickStart)
				}
			}
		})
	}
}

func TestAdminCreatePreservesExistingFilesAndDryRun(t *testing.T) {
	for _, stack := range []string{"vben", "soybean"} {
		parent := t.TempDir()
		root := filepath.Join(parent, "admin")
		c := Config{Name: "admin", Admin: true, AdminStack: stack}
		if err := Create(root, c, true); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(root); !os.IsNotExist(err) {
			t.Fatal("Admin dry-run wrote project files")
		}
		if err := Create(root, c, false); err != nil {
			t.Fatal(err)
		}
		executables, err := admintemplates.ExecutablePaths(stack)
		if err != nil {
			t.Fatal(err)
		}
		for path := range executables {
			info, err := os.Stat(filepath.Join(root, path))
			if err != nil || info.Mode()&0111 == 0 {
				t.Fatalf("upstream executable mode lost for %s: %v", path, err)
			}
		}
		marker := filepath.Join(root, "marker")
		if err := os.WriteFile(marker, []byte("keep"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := Create(root, c, false); err == nil {
			t.Fatal("Admin creation overwrote an existing project")
		}
		if got, err := os.ReadFile(marker); err != nil || string(got) != "keep" {
			t.Fatalf("existing Admin project changed: %q, %v", got, err)
		}
		if err := Doctor(root); err != nil {
			t.Fatal(err)
		}
		if _, err := exec.LookPath("git"); err == nil {
			if output, err := exec.Command("git", "init", root).CombinedOutput(); err != nil {
				t.Fatalf("git init: %v: %s", err, output)
			}
			publicEnv := "apps/admin/.env"
			if stack == "vben" {
				publicEnv = "apps/admin/apps/web-naive/.env"
			}
			for _, path := range []string{publicEnv, publicEnv + ".production", publicEnv + ".example", ".env.production", ".env.example", "apps/admin/pnpm-lock.yaml"} {
				err := exec.Command("git", "-C", root, "check-ignore", "--no-index", path).Run()
				if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
					t.Fatalf("public Admin source %s is ignored or Git check failed: %v", path, err)
				}
			}
			for _, path := range []string{".env", ".env.local", ".env.production.local", "apps/admin/.env.local", "apps/admin/.env.production.local", publicEnv + ".local", publicEnv + ".production.local"} {
				if err := exec.Command("git", "-C", root, "check-ignore", path).Run(); err != nil {
					t.Fatalf("private environment configuration %s is no longer ignored: %v", path, err)
				}
			}
		}
	}
}
