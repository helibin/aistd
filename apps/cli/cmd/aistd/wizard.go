package main

import (
	"aistd"
	"aistd/internal/project"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/x/term"
)

var errWizardCancelled = errors.New("已取消，未创建项目")

func interactiveTerminal() bool {
	return os.Getenv("TERM") != "dumb" && os.Getenv("TERM") != "" &&
		term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stdout.Fd())
}

func wizardField(field huh.Field) error {
	keys := huh.NewDefaultKeyMap()
	keys.Quit.SetKeys("ctrl+c", "esc")
	err := huh.NewForm(huh.NewGroup(field)).
		WithInput(os.Stdin).WithOutput(os.Stdout).
		WithTheme(huh.ThemeBase()).WithKeyMap(keys).WithShowHelp(false).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return errWizardCancelled
	}
	return err
}

func validateWizardName(name string) error {
	if err := (project.Config{Name: name}).Validate(); err != nil {
		return errors.New("名称需以小写字母开头，只含小写字母、数字或连字符，共 1–63 个字符")
	}
	return nil
}

func validateWizardTargets(targets []string) error {
	if len(targets) == 0 {
		return errors.New("请至少选择一个应用目标")
	}
	return selectTargets(&project.Config{}, strings.Join(targets, ","))
}

func createWizard(c project.Config, dir string, dry bool) (project.Config, error) {
	fmt.Printf("┌ aistd %s · 创建项目\n│\n", aistd.Version)
	c.Name = "hello-aistd"
	if err := wizardField(huh.NewInput().Title("项目名称").
		Description("回车继续 · Esc 取消").Value(&c.Name).Validate(validateWizardName)); err != nil {
		return c, err
	}
	fmt.Printf("◇ 项目名称 · %s\n│\n", c.Name)
	targets := []string{"api", "web"}
	if err := wizardField(huh.NewMultiSelect[string]().Title("选择应用目标").
		Description("↑↓ 移动 · 空格选择 · 回车确认 · Esc 取消").Filtering(false).
		Options(huh.NewOption("API", "api"), huh.NewOption("Web", "web"),
			huh.NewOption("iOS App", "ios"), huh.NewOption("Android App", "android"),
			huh.NewOption("微信小程序", "wechat"), huh.NewOption("支付宝小程序", "alipay"),
			huh.NewOption("Admin 后台管理", "admin")).
		Value(&targets).Validate(validateWizardTargets)); err != nil {
		return c, err
	}
	if err := selectTargets(&c, strings.Join(targets, ",")); err != nil {
		return c, err
	}
	fmt.Printf("◇ 应用目标 · %s\n│\n", strings.Join(targets, ", "))
	for _, stack := range []struct {
		enabled        bool
		title          string
		labels, values [2]string
		value          *string
	}{
		{c.API, "API 技术栈", [2]string{"Go", "NestJS"}, [2]string{"go", "nestjs"}, &c.APIStack},
		{c.Web, "Web 技术栈", [2]string{"React", "Vue"}, [2]string{"react", "vue"}, &c.WebStack},
		{c.IOS || c.Android, "移动端技术栈", [2]string{"React Native / Expo", "Flutter"}, [2]string{"react-native", "flutter"}, &c.MobileStack},
		{c.WeChat || c.Alipay, "小程序技术栈", [2]string{"Taro", "uni-app"}, [2]string{"taro", "uniapp"}, &c.MiniappStack},
		{c.Admin, "Admin 模板", [2]string{"Vben Naive UI", "Soybean-admin"}, [2]string{"vben", "soybean"}, &c.AdminStack},
	} {
		if !stack.enabled {
			continue
		}
		*stack.value = stack.values[0]
		if err := wizardField(huh.NewSelect[string]().Title(stack.title).
			Description("↑↓ 移动 · 回车确认 · Esc 取消").Filtering(false).
			Options(huh.NewOption(stack.labels[0], stack.values[0]), huh.NewOption(stack.labels[1], stack.values[1])).
			Value(stack.value)); err != nil {
			return c, err
		}
		label := stack.labels[0]
		if *stack.value == stack.values[1] {
			label = stack.labels[1]
		}
		fmt.Printf("◇ %s · %s\n│\n", stack.title, label)
	}
	if err := wizardField(huh.NewSelect[bool]().Title("包含视觉主题草稿？").
		Description("↑↓ 移动 · 回车确认 · Esc 取消").
		Options(huh.NewOption("包含", true), huh.NewOption("不包含", false)).Value(&c.Visual)); err != nil {
		return c, err
	}
	fmt.Printf("◇ 视觉主题草稿 · %s\n│\n", yesNo(c.Visual))
	c, err := c.Normalize()
	if err != nil {
		return c, err
	}
	summary, err := wizardSummary(c, dir)
	if err != nil {
		return c, err
	}
	fmt.Print(summary)
	confirmed := true
	action := "确认创建项目？"
	if dry {
		action = "确认预览？（不写入文件）"
	}
	if err := wizardField(huh.NewConfirm().Title(action).
		Description("←→ 选择 · 回车确认 · Esc 取消").
		Affirmative("确认").Negative("取消").Value(&confirmed)); err != nil {
		return c, err
	}
	if !confirmed {
		return c, errWizardCancelled
	}
	return c, nil
}

func yesNo(value bool) string {
	if value {
		return "是"
	}
	return "否"
}

func wizardSummary(c project.Config, dir string) (string, error) {
	root, err := filepath.Abs(filepath.Join(dir, c.Name))
	if err != nil {
		return "", err
	}
	var summary strings.Builder
	fmt.Fprintf(&summary, "◇ 创建摘要\n│ 目录：%s\n", root)
	for _, target := range []struct {
		enabled      bool
		label, stack string
	}{
		{c.API, "API", c.APIStack}, {c.Web, "Web", c.WebStack},
		{c.IOS, "iOS App", c.MobileStack}, {c.Android, "Android App", c.MobileStack},
		{c.WeChat, "微信小程序", c.MiniappStack}, {c.Alipay, "支付宝小程序", c.MiniappStack},
		{c.Admin, "Admin", c.AdminStack},
	} {
		if target.enabled {
			fmt.Fprintf(&summary, "│ %s：%s\n", target.label, target.stack)
		}
	}
	fmt.Fprintf(&summary, "│ 视觉主题草稿：%s\n│ OpenSpec：%s\n│\n", yesNo(c.Visual), yesNo(c.OpenSpec))
	return summary.String(), nil
}
