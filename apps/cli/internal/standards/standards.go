package standards

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"strings"
)

//go:embed assets/*.zip
var archives embed.FS

type Bundle struct {
	Name          string
	Version       string
	ArchiveSHA256 string
	Files         map[string][]byte
}

var sources = []struct {
	name, asset, sha256, prefix string
}{
	{"engineering", "assets/engineering-v1.0.zip", "6cad0f6013e8ce7c925041b9d7ad8913d404e184e2828d711bf03edbaadf6cc6", ""},
	{"visual", "assets/visual-v1.0.zip", "f18ea62a1cb552a43b31e560d03519757523dab819784748eb597bb1b8c757e5", "AI-Product-Visual-Design-Standard-v1.0/"},
}

func Load() (map[string]Bundle, error) {
	bundles := make(map[string]Bundle, len(sources))
	for _, source := range sources {
		data, err := archives.ReadFile(source.asset)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != source.sha256 {
			return nil, fmt.Errorf("embedded %s standard archive differs from pinned SHA256", source.name)
		}
		archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, fmt.Errorf("open %s standard: %w", source.name, err)
		}
		files := make(map[string][]byte)
		for _, entry := range archive.File {
			if entry.FileInfo().IsDir() {
				continue
			}
			if !strings.HasPrefix(entry.Name, source.prefix) {
				return nil, fmt.Errorf("unexpected path in %s standard: %q", source.name, entry.Name)
			}
			rel := strings.TrimPrefix(entry.Name, source.prefix)
			if rel == "" || path.IsAbs(rel) || path.Clean(rel) != rel || rel == ".." || strings.HasPrefix(rel, "../") || !entry.Mode().IsRegular() {
				return nil, fmt.Errorf("unsafe path in %s standard: %q", source.name, entry.Name)
			}
			if _, exists := files[rel]; exists {
				return nil, fmt.Errorf("duplicate path in %s standard: %q", source.name, rel)
			}
			reader, err := entry.Open()
			if err != nil {
				return nil, err
			}
			content, readErr := io.ReadAll(reader)
			closeErr := reader.Close()
			if readErr != nil {
				return nil, readErr
			}
			if closeErr != nil {
				return nil, closeErr
			}
			files[rel] = content
		}
		bundles[source.name] = Bundle{Name: source.name, Version: "1.0", ArchiveSHA256: source.sha256, Files: files}
	}
	return bundles, nil
}
