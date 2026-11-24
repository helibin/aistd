package project

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed all:templates/nestjs
var nestTemplates embed.FS

func nestjsFiles(c Config) (map[string]string, error) {
	files := map[string]string{}
	err := fs.WalkDir(nestTemplates, "templates/nestjs", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, err := nestTemplates.ReadFile(name)
		if err != nil {
			return err
		}
		files["apps/api/"+strings.TrimPrefix(name, "templates/nestjs/")] = strings.ReplaceAll(string(content), "PROJECT-api", c.Name+"-api")
		return nil
	})
	return files, err
}
