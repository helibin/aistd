package main

import (
	"aistd/internal/project"
	"path/filepath"
	"strings"
	"testing"
)

func TestWizardValidation(t *testing.T) {
	for _, name := range []string{"hello-aistd", "a", strings.Repeat("a", 63)} {
		if err := validateWizardName(name); err != nil {
			t.Fatalf("rejected valid name %q: %v", name, err)
		}
	}
	for _, name := range []string{"", "../outside", "Aistd", strings.Repeat("a", 64)} {
		if err := validateWizardName(name); err == nil {
			t.Fatalf("accepted invalid name %q", name)
		}
	}
	for _, targets := range [][]string{nil, {"api", "api"}, {"unknown"}} {
		if err := validateWizardTargets(targets); err == nil {
			t.Fatalf("accepted invalid targets %v", targets)
		}
	}
	if err := validateWizardTargets([]string{"api", "web", "ios", "android", "wechat", "alipay", "admin"}); err != nil {
		t.Fatal(err)
	}
}

func TestWizardSummaryMatchesNormalizedSelection(t *testing.T) {
	parent := t.TempDir()
	c, err := (project.Config{Name: "hello-aistd", API: true, Web: true, WebStack: "vue", IOS: true, Android: true, MobileStack: "flutter", WeChat: true, MiniappStack: "uniapp", Visual: true, OpenSpec: true, Admin: true, AdminStack: "soybean"}).Normalize()
	if err != nil {
		t.Fatal(err)
	}
	summary, err := wizardSummary(c, parent)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{filepath.Join(parent, "hello-aistd"), "API：go", "Web：vue", "iOS App：flutter", "Android App：flutter", "微信小程序：uniapp", "视觉主题草稿：是", "OpenSpec：是", "Admin：soybean"} {
		if !strings.Contains(summary, text) {
			t.Fatalf("summary missing %q: %s", text, summary)
		}
	}
	if strings.Contains(summary, "支付宝") {
		t.Fatalf("summary included an unselected target: %s", summary)
	}
}

func TestWizardFallsBackWithoutCapableTerminal(t *testing.T) {
	for _, terminal := range []string{"", "dumb"} {
		t.Setenv("TERM", terminal)
		if interactiveTerminal() {
			t.Fatalf("interactive wizard enabled for TERM=%q", terminal)
		}
	}
}
