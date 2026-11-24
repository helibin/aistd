package project

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

// Versions follow dcloudio/uni-preset-vue's vite-ts branch, inspected 2026-11-16.
const uniappVersion = "3.0.0-5030420261029002"

//go:embed all:templates/uniapp
var uniappTemplates embed.FS

func uniappFiles(c Config) (map[string]string, error) {
	files := make(map[string]string)
	err := fs.WalkDir(uniappTemplates, "templates/uniapp", func(file string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, err := uniappTemplates.ReadFile(file)
		if err != nil {
			return err
		}
		files[path.Join("apps/miniapp", strings.TrimPrefix(file, "templates/uniapp/"))] = strings.ReplaceAll(string(content), "AISTD_PROJECT_NAME", c.Name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	scripts := map[string]string{"check": "vue-tsc --noEmit"}
	dependencies := map[string]string{
		"@dcloudio/uni-app": uniappVersion, "@dcloudio/uni-components": uniappVersion,
		"vue": "3.4.21",
	}
	manifest := map[string]any{
		"name": c.Name, "appid": "", "versionName": "0.1.0", "versionCode": "100",
		"vueVersion": "3", "uniStatistics": map[string]bool{"enable": false},
	}
	if c.WeChat {
		dependencies["@dcloudio/uni-mp-weixin"] = uniappVersion
		scripts["build:weapp"], scripts["dev:weapp"] = "uni build -p mp-weixin", "uni -p mp-weixin"
		manifest["mp-weixin"] = map[string]any{"appid": "touristappid", "usingComponents": true, "setting": map[string]bool{"urlCheck": true}}
	}
	if c.Alipay {
		dependencies["@dcloudio/uni-mp-alipay"] = uniappVersion
		scripts["build:alipay"], scripts["dev:alipay"] = "uni build -p mp-alipay", "uni -p mp-alipay"
		manifest["mp-alipay"] = map[string]any{"usingComponents": true}
	}
	packageJSON, err := jsonFile(map[string]any{
		"name": c.Name + "-miniapp", "version": "0.1.0", "private": true,
		"scripts": scripts, "dependencies": dependencies,
		"devDependencies": map[string]string{
			"@dcloudio/types": "3.4.31", "@dcloudio/uni-cli-shared": uniappVersion,
			"@dcloudio/vite-plugin-uni": uniappVersion,
			"typescript":                "5.8.3", "vite": "5.2.8", "vue-tsc": "2.2.12",
		},
	})
	if err != nil {
		return nil, err
	}
	files["apps/miniapp/package.json"] = packageJSON
	files["apps/miniapp/src/manifest.json"], err = jsonFile(manifest)
	return files, err
}
