package project

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUniappFilesSelectPlatforms(t *testing.T) {
	for _, tc := range []Config{
		{Name: "mini", WeChat: true},
		{Name: "mini", Alipay: true},
		{Name: "mini", WeChat: true, Alipay: true},
	} {
		files, err := uniappFiles(tc)
		if err != nil {
			t.Fatal(err)
		}
		var pkg struct {
			Scripts         map[string]string `json:"scripts"`
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if err := json.Unmarshal([]byte(files["apps/miniapp/package.json"]), &pkg); err != nil {
			t.Fatal(err)
		}
		for _, platform := range []struct {
			name, key string
			enabled   bool
		}{
			{"weapp", "mp-weixin", tc.WeChat}, {"alipay", "mp-alipay", tc.Alipay},
		} {
			command, ok := pkg.Scripts["build:"+platform.name]
			if ok != platform.enabled || ok && command != "uni build -p "+platform.key {
				t.Fatalf("build command mismatch for %s", platform.name)
			}
			_, ok = pkg.Dependencies["@dcloudio/uni-"+platform.key]
			if ok != platform.enabled {
				t.Fatalf("dependency mismatch for %s", platform.name)
			}
			var manifest map[string]any
			if err := json.Unmarshal([]byte(files["apps/miniapp/src/manifest.json"]), &manifest); err != nil {
				t.Fatal(err)
			}
			_, ok = manifest[platform.key]
			if ok != platform.enabled {
				t.Fatalf("manifest mismatch for %s", platform.name)
			}
		}
		for name, version := range pkg.Dependencies {
			if strings.HasPrefix(name, "@dcloudio/") && version != uniappVersion {
				t.Fatalf("mixed DCloud release: %s", name)
			}
		}
		if pkg.Scripts["check"] != "vue-tsc --noEmit" {
			t.Fatal("missing type check")
		}
		if !strings.Contains(files["apps/miniapp/src/pages/index/index.vue"], ">mini<") {
			t.Fatal("missing project name")
		}
	}
}

func TestUniappCompilerMatchesVueWorkspace(t *testing.T) {
	miniapp, err := uniappFiles(Config{Name: "mixed", WeChat: true})
	if err != nil {
		t.Fatal(err)
	}
	web, err := vueFiles(Config{Name: "mixed", Web: true})
	if err != nil {
		t.Fatal(err)
	}
	var miniappPackage, webPackage struct {
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal([]byte(miniapp["apps/miniapp/package.json"]), &miniappPackage); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(web["apps/web/package.json"]), &webPackage); err != nil {
		t.Fatal(err)
	}
	// Old vue-tsc breaks when npm hoists a newer TypeScript from the web workspace.
	for _, dependency := range []string{"typescript", "vue-tsc"} {
		if miniappPackage.DevDependencies[dependency] != webPackage.DevDependencies[dependency] {
			t.Errorf("mixed Vue/uniapp workspaces need compatible %s versions", dependency)
		}
	}
}
