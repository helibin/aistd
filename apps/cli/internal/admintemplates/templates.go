// Package admintemplates installs pinned upstream Admin sources without network access.
package admintemplates

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
)

//go:embed assets/*.zip
var archives embed.FS

type source struct {
	Repository    string `json:"repository"`
	Version       string `json:"version"`
	Commit        string `json:"commit"`
	ArchiveSHA256 string `json:"archiveSha256"`
	License       string `json:"license"`
	Entrypoint    string `json:"entrypoint"`
	archive       string
}

var sources = map[string]source{
	"vben": {
		Repository: "https://github.com/vbenjs/vue-vben-admin", Version: "5.7.0",
		Commit:        "63a38dce49ba109f61607994e21ba921d8e970e9",
		ArchiveSHA256: "afcd73a7fe5857f8dc2b44022350f6495ec6f0ad4b061914714ba257d0597fad",
		License:       "MIT", Entrypoint: "apps/admin/apps/web-naive", archive: "vben-v5.7.0.zip",
	},
	"soybean": {
		Repository: "https://github.com/soybeanjs/soybean-admin", Version: "2.0.2",
		Commit:        "5e40b850180804aa51573590eb2e058d1fd8eb06",
		ArchiveSHA256: "2007bc56ac5779f2a3da20e8720159277cc67567ee08af808458e070d99c9606",
		License:       "MIT", Entrypoint: "apps/admin", archive: "soybean-v2.0.2.zip",
	},
}

func openSource(stack string) (source, *zip.Reader, error) {
	s, ok := sources[stack]
	if !ok {
		return s, nil, fmt.Errorf("unknown Admin template %q", stack)
	}
	data, err := archives.ReadFile("assets/" + s.archive)
	if err != nil {
		return s, nil, err
	}
	r, err := checkedArchive(data, s.ArchiveSHA256)
	return s, r, err
}

func checkedArchive(data []byte, expected string) (*zip.Reader, error) {
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != expected {
		return nil, fmt.Errorf("Admin template archive checksum mismatch")
	}
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	root := ""
	seen := make(map[string]bool)
	for _, f := range r.File {
		parts := strings.SplitN(f.Name, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[0] == "." || parts[0] == ".." || strings.Contains(f.Name, "\\") {
			return nil, fmt.Errorf("unsafe Admin archive path %q", f.Name)
		}
		if root == "" {
			root = parts[0]
		}
		rel := strings.TrimSuffix(parts[1], "/")
		if parts[0] != root || (rel != "" && (path.Clean(rel) != rel || path.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, "../"))) {
			return nil, fmt.Errorf("unsafe Admin archive path %q", f.Name)
		}
		if f.FileInfo().IsDir() {
			continue
		}
		if rel == "" || !f.Mode().IsRegular() || seen[rel] {
			return nil, fmt.Errorf("invalid Admin archive entry %q", f.Name)
		}
		seen[rel] = true
	}
	if len(seen) == 0 {
		return nil, fmt.Errorf("empty Admin template archive")
	}
	return r, nil
}

// ExecutablePaths preserves the executable modes of the pinned upstream sources.
func ExecutablePaths(stack string) (map[string]bool, error) {
	_, r, err := openSource(stack)
	if err != nil {
		return nil, err
	}
	paths := make(map[string]bool)
	for _, f := range r.File {
		if !f.FileInfo().IsDir() && f.Mode()&0111 != 0 {
			paths["apps/admin/"+strings.SplitN(f.Name, "/", 2)[1]] = true
		}
	}
	return paths, nil
}

// Load returns project-relative files, including source provenance and integration notes.
func Load(stack string) (map[string]string, error) {
	s, r, err := openSource(stack)
	if err != nil {
		return nil, err
	}
	files := make(map[string]string)
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		reader, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		files["apps/admin/"+strings.SplitN(f.Name, "/", 2)[1]] = string(data)
	}
	// Keep the upstream dependency graph and lockfile, adapting only integration settings.
	files["apps/admin/.npmrc"] = strings.ReplaceAll(files["apps/admin/.npmrc"], "https://registry.npmmirror.com", "https://registry.npmjs.org")
	if stack == "vben" {
		for _, name := range []string{".env.development", ".env.production"} {
			key := "apps/admin/apps/web-naive/" + name
			files[key] = strings.ReplaceAll(files[key], "VITE_BASE=/", "VITE_BASE=/admin/")
		}
		key := "apps/admin/apps/web-naive/src/preferences.ts"
		files[key] = strings.Replace(files[key], "  // overrides\n", "  // overrides\n  logo: { source: `${import.meta.env.BASE_URL}favicon.ico` },\n", 1)
	} else {
		files["apps/admin/.env"] = strings.ReplaceAll(files["apps/admin/.env"], "VITE_BASE_URL=/", "VITE_BASE_URL=/admin/")
		files["apps/admin/vite.config.ts"] = strings.ReplaceAll(files["apps/admin/vite.config.ts"], "open: true", "open: false")
		var pkg map[string]any
		if err := json.Unmarshal([]byte(files["apps/admin/package.json"]), &pkg); err != nil {
			return nil, err
		}
		pkg["packageManager"] = "pnpm@10.33.4"
		data, err := json.MarshalIndent(pkg, "", "  ")
		if err != nil {
			return nil, err
		}
		files["apps/admin/package.json"] = string(data) + "\n"
	}
	manifest, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, err
	}
	files[".aistd/admin-template.json"] = string(manifest) + "\n"
	files["apps/admin/AISTD-TEMPLATE.md"] = fmt.Sprintf(`# Admin template

Source: %s, version %s, commit %s (MIT).
Bundled archive SHA256: %s. The upstream LICENSE is preserved here.
The upstream workspace structure is retained to preserve its dependency graph;
the selected entrypoint is %s. Vben uses only its Naive UI application.

This is an independent pnpm workspace, not part of the root npm workspaces.
Requirements: Node.js 22.18+ and npm. From the generated project root use
make setup, make check, make build and make run-admin; pnpm 10.33.4 is invoked
through npx without requiring a global pnpm installation.

Integration changes: the selected application uses /admin/ as its base URL;
Vben's selected application uses its bundled favicon as its logo, avoiding a CDN dependency;
the template-local registry is registry.npmjs.org; Soybean records the fixed pnpm
version and disables automatic browser opening. Source dependencies and lockfiles
are retained. Vben's other UI applications remain upstream source references.

Vben development uses its local Nitro mock. Soybean still points to the upstream
Apifox demo service by default, but no demo token is bundled. Configure a backend
you control before use. The generated Go/NestJS hello API has no Admin authentication
or permissions endpoints.
doctor verifies pinned standard sources, not Admin template integrity or compliance.
`, s.Repository, s.Version, s.Commit, s.ArchiveSHA256, s.Entrypoint)
	return files, nil
}
