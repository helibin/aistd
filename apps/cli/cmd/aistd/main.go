package main

import (
	"aistd"
	"aistd/internal/project"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
func run(args []string) error {
	if len(args) == 0 && interactiveTerminal() {
		args = []string{"create"}
	}
	if len(args) == 0 {
		fmt.Println("aistd create|init|doctor|uninstall|version\nUse aistd create -h for options.")
		return nil
	}
	if args[0] == "version" {
		fmt.Printf("aistd %s\n", aistd.Version)
		return nil
	}
	if args[0] == "uninstall" {
		fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
		yes := fs.Bool("yes", false, "confirm removal without prompting")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return fmt.Errorf("uninstall accepts no positional arguments")
		}
		return uninstall(*yes, os.Stdin, os.Stdout)
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	config := fs.String("config", "", "JSON configuration path")
	name := fs.String("name", "", "project name; use default Go API, React Web and visual draft")
	targets := fs.String("targets", "", "comma-separated targets: api,web,ios,android,wechat,alipay,admin")
	apiStack := fs.String("api-stack", "", "API stack: go or nestjs")
	webStack := fs.String("web-stack", "", "Web stack: react or vue")
	mobileStack := fs.String("mobile-stack", "", "mobile stack: react-native or flutter")
	miniappStack := fs.String("miniapp-stack", "", "miniapp stack: taro or uniapp")
	adminStack := fs.String("admin-stack", "", "admin template: vben (Naive UI) or soybean")
	visual := fs.Bool("visual", true, "include visual theme draft")
	openSpec := fs.Bool("openspec", false, "initialize official OpenSpec CodeBuddy workflow (requires openspec CLI)")
	dry := fs.Bool("dry-run", false, "preview without writing")
	yes := fs.Bool("yes", false, "disable interactive questions")
	dir := fs.String("dir", ".", "parent directory for create; project directory for init/doctor")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments; use --dir to select a directory")
	}
	visualSet := false
	stackSet := false
	fs.Visit(func(f *flag.Flag) {
		if strings.HasSuffix(f.Name, "-stack") {
			stackSet = true
		}
		if f.Name == "visual" {
			visualSet = true
		}
	})
	switch args[0] {
	case "init":
		return project.AdoptWithOpenSpec(*dir, *dry, *openSpec)
	case "doctor":
		return project.Doctor(*dir)
	case "create":
		c := project.Config{API: true, Visual: true, OpenSpec: *openSpec}
		guided := false
		if *config != "" && (*targets != "" || visualSet || stackSet) {
			return fmt.Errorf("--targets, --visual and stack flags cannot be used with --config")
		}
		if *targets != "" && *name == "" {
			return fmt.Errorf("--targets requires --name")
		}
		if visualSet && *name == "" && *config == "" {
			return fmt.Errorf("--visual requires --name")
		}
		if *config != "" && *name != "" {
			return fmt.Errorf("--name and --config cannot be used together")
		}
		if *config != "" {
			f, err := os.Open(*config)
			if err != nil {
				return err
			}
			defer f.Close()
			d := json.NewDecoder(f)
			d.DisallowUnknownFields()
			if err = d.Decode(&c); err != nil {
				return err
			}
			var extra any
			if err = d.Decode(&extra); err != io.EOF {
				return fmt.Errorf("configuration must contain exactly one JSON object")
			}
		} else if *name != "" {
			c.Name = *name
			c.Web = true
			c.Visual = *visual
			c.APIStack, c.WebStack, c.MobileStack, c.MiniappStack, c.AdminStack = *apiStack, *webStack, *mobileStack, *miniappStack, *adminStack
			if *targets != "" {
				if err := selectTargets(&c, *targets); err != nil {
					return err
				}
			}
		} else if *yes {
			return fmt.Errorf("--yes requires --name or --config")
		} else {
			if stackSet {
				return fmt.Errorf("stack flags require --name")
			}
			if interactiveTerminal() {
				var err error
				c, err = createWizard(c, *dir, *dry)
				if err != nil {
					return err
				}
				guided = true
			} else {
				c.Web = true
				reader := bufio.NewReader(os.Stdin)
				ask := func(prompt string) (string, error) {
					fmt.Print(prompt)
					s, err := reader.ReadString('\n')
					return strings.TrimSpace(s), err
				}
				var err error
				c.Name, err = ask("Project name: ")
				if err != nil {
					return err
				}
				fmt.Println("Select one or more targets: 1 API, 2 Web, 3 iOS App, 4 Android App, 5 WeChat Mini Program, 6 Alipay Mini Program, 7 Admin")
				s, err := ask("Targets [1,2]: ")
				if err != nil {
					return err
				}
				if s == "" {
					s = "1,2"
				}
				if err := selectTargets(&c, s); err != nil {
					return err
				}
				for _, stack := range []struct {
					enabled               bool
					prompt, first, second string
					value                 *string
				}{
					{c.API, "API stack: 1 Go, 2 NestJS [1]: ", "go", "nestjs", &c.APIStack},
					{c.Web, "Web stack: 1 React, 2 Vue [1]: ", "react", "vue", &c.WebStack},
					{c.IOS || c.Android, "Mobile stack: 1 React Native, 2 Flutter [1]: ", "react-native", "flutter", &c.MobileStack},
					{c.WeChat || c.Alipay, "Miniapp stack: 1 Taro, 2 uni-app [1]: ", "taro", "uniapp", &c.MiniappStack},
					{c.Admin, "Admin template: 1 Vben Naive UI, 2 Soybean-admin [1]: ", "vben", "soybean", &c.AdminStack},
				} {
					if !stack.enabled {
						continue
					}
					choice, err := ask(stack.prompt)
					if err != nil {
						return err
					}
					switch choice {
					case "", "1":
						*stack.value = stack.first
					case "2":
						*stack.value = stack.second
					default:
						return fmt.Errorf("expected stack choice 1 or 2")
					}
				}
				s, err = ask("Include visual theme draft? [Y/n]: ")
				if err != nil {
					return err
				}
				switch strings.ToLower(s) {
				case "", "y", "yes":
					c.Visual = true
				case "n", "no":
					c.Visual = false
				default:
					return fmt.Errorf("expected yes or no")
				}
			}
		}
		if *openSpec {
			c.OpenSpec = true
		}
		c, err := c.Normalize()
		if err != nil {
			return err
		}
		root := filepath.Join(*dir, c.Name)
		if guided && !*dry {
			fmt.Println("│ 正在生成项目并安装规范原文…")
		}
		if err := project.Create(root, c, *dry); err != nil {
			if guided {
				fmt.Println("└ 未完成，请根据错误信息检查目标目录。")
			}
			return err
		}
		if guided {
			if *dry {
				fmt.Println("└ 预览完成，未写入文件。")
			} else {
				fmt.Println("└ 项目创建完成。")
			}
		}
		if !*dry {
			fmt.Println("Created:", root)
			fmt.Printf("\nNext, enter the project:\n\n  cd -- '%s'\n\n", strings.ReplaceAll(root, "'", "'\"'\"'"))
			fmt.Print(project.QuickStart(c))
			fmt.Println("\nSee README.md for the full project guide.")
		}
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// uninstall only removes the running executable when it lives in the installer's directory.
// Project files and the installation directory are never removed.
func uninstall(yes bool, input io.Reader, output io.Writer) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate running executable: %w", err)
	}
	installDir := os.Getenv("AISTD_INSTALL_DIR")
	if installDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("locate home directory: %w", err)
		}
		installDir = filepath.Join(home, ".local", "bin")
	}
	installDir, err = filepath.Abs(installDir)
	if err != nil {
		return err
	}
	installDir, err = filepath.EvalSymlinks(installDir)
	if err != nil {
		return fmt.Errorf("resolve installation directory: %w", err)
	}
	target := filepath.Join(installDir, "aistd")
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return fmt.Errorf("resolve running executable: %w", err)
	}
	if executable != target {
		return fmt.Errorf("running executable %q does not match installer path %q; set AISTD_INSTALL_DIR to the installation directory if customized", executable, target)
	}
	info, err := os.Lstat(target)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("refusing to remove non-regular file %q", target)
	}
	if !yes {
		fmt.Fprintf(output, "Remove CLI executable %s? [y/N]: ", target)
		answer, err := bufio.NewReader(input).ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if strings.ToLower(strings.TrimSpace(answer)) != "y" && strings.ToLower(strings.TrimSpace(answer)) != "yes" {
			return fmt.Errorf("uninstall cancelled")
		}
	}
	current, err := os.Lstat(target)
	if err != nil {
		return err
	}
	if !os.SameFile(info, current) || !current.Mode().IsRegular() {
		return fmt.Errorf("installation changed during confirmation; refusing to remove %q", target)
	}
	if err := os.Remove(target); err != nil {
		return fmt.Errorf("remove CLI executable: %w", err)
	}
	fmt.Fprintf(output, "Uninstalled: %s\nProject files and installation directory were not removed.\n", target)
	return nil
}

func selectTargets(c *project.Config, input string) error {
	c.API, c.Web, c.IOS, c.Android, c.WeChat, c.Alipay, c.Admin = false, false, false, false, false, false, false
	seen := make(map[string]bool)
	for _, item := range strings.Split(input, ",") {
		key := strings.ToLower(strings.TrimSpace(item))
		var target string
		switch key {
		case "1", "api":
			target = "api"
			c.API = true
		case "2", "web":
			target = "web"
			c.Web = true
		case "3", "ios":
			target = "ios"
			c.IOS = true
		case "4", "android":
			target = "android"
			c.Android = true
		case "5", "wechat":
			target = "wechat"
			c.WeChat = true
		case "6", "alipay":
			target = "alipay"
			c.Alipay = true
		case "7", "admin":
			target = "admin"
			c.Admin = true
		default:
			return fmt.Errorf("unknown target %q; choose api,web,ios,android,wechat,alipay,admin", key)
		}
		if seen[target] {
			return fmt.Errorf("duplicate target: %s", target)
		}
		seen[target] = true
	}
	return nil
}
