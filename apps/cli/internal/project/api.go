package project

import (
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"path"
	"strings"
)

//go:embed all:templates/api
var apiTemplates embed.FS

func apiFiles(name string) (map[string]string, error) {
	files := make(map[string]string)
	err := fs.WalkDir(apiTemplates, "templates/api", func(file string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, err := apiTemplates.ReadFile(file)
		if err != nil {
			return err
		}
		rel := strings.TrimSuffix(strings.TrimPrefix(file, "templates/api/"), ".tmpl")
		content = []byte(strings.ReplaceAll(string(content), "example.com/PROJECT/api", "example.com/"+name+"/api"))
		if strings.HasSuffix(rel, ".go") {
			content, err = format.Source(content)
			if err != nil {
				return fmt.Errorf("format API template %s: %w", rel, err)
			}
		}
		files[path.Join("apps/api", rel)] = string(content)
		return nil
	})
	return files, err
}
