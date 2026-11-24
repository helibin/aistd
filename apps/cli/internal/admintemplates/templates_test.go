package admintemplates

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
)

func TestPinnedTemplates(t *testing.T) {
	for _, stack := range []string{"vben", "soybean"} {
		t.Run(stack, func(t *testing.T) {
			files, err := Load(stack)
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"apps/admin/LICENSE", "apps/admin/pnpm-lock.yaml", "apps/admin/pnpm-workspace.yaml", "apps/admin/AISTD-TEMPLATE.md", ".aistd/admin-template.json"} {
				if files[name] == "" {
					t.Fatalf("missing %s", name)
				}
			}
			var manifest source
			if err := json.Unmarshal([]byte(files[".aistd/admin-template.json"]), &manifest); err != nil || manifest.Commit != sources[stack].Commit || manifest.License != "MIT" {
				t.Fatalf("invalid provenance: %+v, %v", manifest, err)
			}
			if !strings.Contains(files["apps/admin/.npmrc"], "registry=https://registry.npmjs.org") {
				t.Fatal("unexpected local registry")
			}
			var pkg struct {
				PackageManager string `json:"packageManager"`
			}
			if err := json.Unmarshal([]byte(files["apps/admin/package.json"]), &pkg); err != nil || pkg.PackageManager != "pnpm@10.33.4" {
				t.Fatalf("unexpected package manager: %+v %v", pkg, err)
			}
			if stack == "vben" {
				for _, line := range strings.Split(files["apps/admin/playground/.env.development"], "\n") {
					if strings.HasPrefix(line, "VITE_GLOB_AUTH_DINGDING_") && !strings.HasSuffix(line, "=") {
						t.Fatal("public template must not ship DingTalk identifiers")
					}
				}
				for _, env := range []string{".env.development", ".env.production"} {
					if !strings.Contains(files["apps/admin/apps/web-naive/"+env], "VITE_BASE=/admin/\n") {
						t.Fatalf("wrong base URL in %s", env)
					}
				}
				if !strings.Contains(files["apps/admin/apps/web-naive/package.json"], "naive-ui") {
					t.Fatal("missing Naive UI entrypoint")
				}
				if !strings.Contains(files["apps/admin/apps/web-naive/src/preferences.ts"], "logo: { source: `${import.meta.env.BASE_URL}favicon.ico` }") || files["apps/admin/apps/web-naive/public/favicon.ico"] == "" {
					t.Fatal("missing bundled Vben logo")
				}
			} else if !strings.Contains(files["apps/admin/.env"], "VITE_BASE_URL=/admin/\n") || strings.Contains(files["apps/admin/vite.config.ts"], "open: true") {
				t.Fatal("wrong Soybean integration settings")
			} else if strings.Contains(files["apps/admin/src/service/request/index.ts"], "apifoxToken") {
				t.Fatal("public template must not ship an Apifox demo token")
			}
			executables, err := ExecutablePaths(stack)
			if err != nil || len(executables) == 0 {
				t.Fatalf("missing executable modes: %v", err)
			}
			for name := range executables {
				if files[name] == "" {
					t.Fatalf("executable missing from source: %s", name)
				}
			}
		})
	}
	if _, err := Load("unknown"); err == nil {
		t.Fatal("unknown template accepted")
	}
}

func TestRejectsUnsafeArchives(t *testing.T) {
	for _, names := range [][]string{{"/outside"}, {"root/../outside"}, {"root/a/../outside"}, {"root/a\\b"}, {"root/a", "other/b"}, {"root/a", "root/a"}, {"root"}} {
		data, sum := testArchive(t, names, false)
		if _, err := checkedArchive(data, sum); err == nil {
			t.Fatalf("accepted unsafe entries: %v", names)
		}
	}
	data, sum := testArchive(t, []string{"root/link"}, true)
	if _, err := checkedArchive(data, sum); err == nil {
		t.Fatal("accepted symlink")
	}
	data, sum = testArchive(t, []string{"root/valid"}, false)
	if _, err := checkedArchive(data, sum); err != nil {
		t.Fatal(err)
	}
	data[0] ^= 1
	if _, err := checkedArchive(data, sum); err == nil {
		t.Fatal("accepted modified archive")
	}
}

func testArchive(t *testing.T, names []string, symlink bool) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, name := range names {
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		h.SetMode(0644)
		if symlink {
			h.SetMode(fs.ModeSymlink | 0777)
		}
		f, err := w.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte("source")); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:])
}
