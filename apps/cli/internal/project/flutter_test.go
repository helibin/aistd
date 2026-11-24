package project

import (
	"bytes"
	"io/fs"
	"path"
	"strings"
	"testing"
)

func TestFlutterFilesSelectNativePlatforms(t *testing.T) {
	for _, tc := range []Config{
		{Name: "switch", IOS: true},
		{Name: "my-app", Android: true},
		{Name: "my-app", IOS: true, Android: true},
	} {
		files, err := flutterFiles(tc)
		if err != nil {
			t.Fatal(err)
		}
		for _, rel := range []string{"pubspec.yaml", "lib/main.dart", "test/widget_test.dart", "FLUTTER-LICENSE", "GRADLE-LICENSE"} {
			if files["apps/mobile/"+rel] == "" {
				t.Fatalf("missing Flutter file %s", rel)
			}
		}
		if !strings.Contains(files["apps/mobile/pubspec.yaml"], "name: app_"+strings.ReplaceAll(tc.Name, "-", "_")) {
			t.Fatal("invalid Dart package name")
		}
		for path, content := range files {
			if strings.Contains(path, "AISTD_") || strings.Contains(content, "AISTD_") || strings.Contains(content, "/Users/") || strings.Contains(content, "DEVELOPMENT_TEAM =") {
				t.Fatalf("unexpanded placeholder or machine-specific content in %s", path)
			}
			if strings.HasSuffix(path, "local.properties") || strings.HasSuffix(path, ".iml") || strings.Contains(path, "/.idea/") || strings.HasSuffix(path, "/pubspec.lock") {
				t.Fatalf("generated or machine-specific file included: %s", path)
			}
			if strings.HasPrefix(path, "apps/mobile/android/") && !tc.Android || strings.HasPrefix(path, "apps/mobile/ios/") && !tc.IOS {
				t.Fatalf("unselected native platform: %s", path)
			}
		}
		_, android := files["apps/mobile/android/app/src/main/AndroidManifest.xml"]
		_, ios := files["apps/mobile/ios/Runner.xcodeproj/project.pbxproj"]
		if android != tc.Android || ios != tc.IOS {
			t.Fatal("native platform selection mismatch")
		}
		if tc.Android && !strings.HasPrefix(files["apps/mobile/android/gradle/wrapper/gradle-wrapper.jar"], "PK") {
			t.Fatal("Gradle wrapper binary missing or corrupted")
		}
	}
}

func TestFlutterBinaryAssetsPreserved(t *testing.T) {
	files, err := flutterFiles(Config{Name: "binary-app", Android: true, IOS: true})
	if err != nil {
		t.Fatal(err)
	}
	err = fs.WalkDir(flutterTemplates, "templates/flutter", func(file string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		if ext := path.Ext(file); ext != ".png" && ext != ".jar" {
			return nil
		}
		original, err := flutterTemplates.ReadFile(file)
		if err != nil {
			return err
		}
		generated := files["apps/mobile/"+strings.TrimPrefix(file, "templates/flutter/")]
		if !bytes.Equal(original, []byte(generated)) {
			t.Errorf("binary asset changed: %s", file)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
