package project

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

// Native shells originate from Flutter 3.44.4; source and licenses ship with the generated app.
//
//go:embed all:templates/flutter
var flutterTemplates embed.FS

func flutterFiles(c Config) (map[string]string, error) {
	files := make(map[string]string)
	replace := strings.NewReplacer(
		"AISTD_DART_NAME", "app_"+strings.ReplaceAll(c.Name, "-", "_"),
		"AISTD_NATIVE_NAME", "app"+strings.ReplaceAll(c.Name, "-", ""),
		"AISTD_PROJECT_NAME", c.Name,
	)
	err := fs.WalkDir(flutterTemplates, "templates/flutter", func(file string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		rel := strings.TrimPrefix(file, "templates/flutter/")
		if strings.HasPrefix(rel, "android/") && !c.Android || strings.HasPrefix(rel, "ios/") && !c.IOS {
			return nil
		}
		content, err := flutterTemplates.ReadFile(file)
		if err != nil {
			return err
		}
		text := string(content)
		if ext := path.Ext(rel); ext != ".png" && ext != ".jar" {
			text = replace.Replace(text)
		}
		files[path.Join("apps/mobile", replace.Replace(rel))] = text
		return nil
	})
	return files, err
}
