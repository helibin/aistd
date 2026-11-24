package project

import (
	"encoding/json"
	"fmt"
	"strings"
)

func jsonFile(value any) (string, error) {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(encoded) + "\n", nil
}

func mobileFiles(c Config) (map[string]string, error) {
	platforms := []string{}
	scripts := map[string]string{
		"start": "expo start",
		"check": "tsc --noEmit",
	}
	expo := map[string]any{
		"name": c.Name, "slug": c.Name, "version": "1.0.0",
		"orientation": "portrait", "userInterfaceStyle": "light",
	}
	nativeID := strings.ReplaceAll(c.Name, "-", "")
	if c.IOS {
		platforms = append(platforms, "ios")
		expo["ios"] = map[string]any{"bundleIdentifier": "com.example." + nativeID + ".ios", "supportsTablet": true}
		scripts["ios"] = "expo start --ios"
		scripts["build:ios"] = "expo export --platform ios --output-dir dist/ios"
	}
	if c.Android {
		platforms = append(platforms, "android")
		expo["android"] = map[string]any{"package": "com.example." + nativeID + ".android"}
		scripts["android"] = "expo start --android"
		scripts["build:android"] = "expo export --platform android --output-dir dist/android"
	}
	expo["platforms"] = platforms
	packageJSON, err := jsonFile(map[string]any{
		"name": c.Name + "-mobile", "version": "1.0.0", "private": true,
		"main": "index.ts", "scripts": scripts,
		"dependencies": map[string]string{
			"expo": "~58.0.19", "react": "19.3.0", "react-native": "0.88.0",
		},
		"devDependencies": map[string]string{
			"@types/react": "~19.3.0", "typescript": "~6.0.3",
		},
	})
	if err != nil {
		return nil, err
	}
	appJSON, err := jsonFile(map[string]any{"expo": expo})
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"apps/mobile/package.json":  packageJSON,
		"apps/mobile/app.json":      appJSON,
		"apps/mobile/tsconfig.json": "{\n  \"extends\": \"expo/tsconfig.base\",\n  \"compilerOptions\": {\"strict\": true}\n}\n",
		"apps/mobile/index.ts":      "import { registerRootComponent } from 'expo';\nimport App from './App';\nregisterRootComponent(App);\n",
		"apps/mobile/App.tsx":       fmt.Sprintf("import { StyleSheet, Text, View } from 'react-native';\n\nexport default function App() {\n  return <View style={styles.root}><Text style={styles.title}>%s</Text></View>;\n}\n\nconst styles = StyleSheet.create({\n  root: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: '#fff' },\n  title: { fontSize: 28, color: '#17211b' },\n});\n", c.Name),
	}, nil
}

func miniappFiles(c Config) (map[string]string, error) {
	scripts := map[string]string{"check": "tsc --noEmit"}
	dependencies := map[string]string{
		"@babel/runtime":     "^7.24.4",
		"@tarojs/components": "4.3.1", "@tarojs/helper": "4.3.1",
		"@tarojs/plugin-framework-react": "4.3.1", "@tarojs/react": "4.3.1",
		"@tarojs/runtime": "4.3.1", "@tarojs/shared": "4.3.1", "@tarojs/taro": "4.3.1",
		"react": "18.3.1", "react-dom": "18.3.1",
	}
	if c.WeChat {
		dependencies["@tarojs/plugin-platform-weapp"] = "4.3.1"
		scripts["build:weapp"] = "taro build --type weapp"
		scripts["dev:weapp"] = "taro build --type weapp --watch"
	}
	if c.Alipay {
		dependencies["@tarojs/plugin-platform-alipay"] = "4.3.1"
		scripts["build:alipay"] = "taro build --type alipay"
		scripts["dev:alipay"] = "taro build --type alipay --watch"
	}
	packageJSON, err := jsonFile(map[string]any{
		"name": c.Name + "-miniapp", "version": "1.0.0", "private": true,
		"scripts": scripts, "dependencies": dependencies,
		"devDependencies": map[string]string{
			"@babel/core": "^7.24.4", "@babel/preset-react": "^7.24.1", "@tarojs/cli": "4.3.1",
			"@tarojs/webpack5-runner": "4.3.1", "@types/node": "^22.0.0",
			"@types/react": "^18.3.0", "babel-preset-taro": "4.3.1",
			"postcss": "^8.5.6", "typescript": "^5.4.5",
		},
	})
	if err != nil {
		return nil, err
	}
	files := map[string]string{
		"apps/miniapp/package.json":              packageJSON,
		"apps/miniapp/config/index.ts":           fmt.Sprintf("import { defineConfig } from '@tarojs/cli';\n\nexport default defineConfig<'webpack5'>(async () => ({\n  projectName: '%s',\n  designWidth: 750,\n  sourceRoot: 'src',\n  outputRoot: `dist/${process.env.TARO_ENV || 'weapp'}`,\n  framework: 'react',\n  compiler: 'webpack5',\n  mini: {},\n}));\n", c.Name),
		"apps/miniapp/babel.config.js":           "module.exports = { presets: [['taro', { framework: 'react', ts: true, compiler: 'webpack5' }]] };\n",
		"apps/miniapp/tsconfig.json":             "{\n  \"compilerOptions\": {\n    \"target\": \"es2017\",\n    \"module\": \"commonjs\",\n    \"moduleResolution\": \"node\",\n    \"jsx\": \"react-jsx\",\n    \"strict\": true,\n    \"skipLibCheck\": true,\n    \"esModuleInterop\": true\n  },\n  \"include\": [\"src\", \"config\", \"types\"]\n}\n",
		"apps/miniapp/types/global.d.ts":         "/// <reference types=\"@tarojs/taro\" />\ndeclare module '*.css';\n",
		"apps/miniapp/src/app.ts":                "import type { PropsWithChildren } from 'react';\nimport './app.css';\n\nexport default function App({ children }: PropsWithChildren) {\n  return children;\n}\n",
		"apps/miniapp/src/app.config.ts":         fmt.Sprintf("export default defineAppConfig({\n  pages: ['pages/index/index'],\n  window: { navigationBarTitleText: '%s' },\n});\n", c.Name),
		"apps/miniapp/src/app.css":               "page { background: #fff; }\n",
		"apps/miniapp/src/pages/index/index.tsx": fmt.Sprintf("import { Text, View } from '@tarojs/components';\nimport './index.css';\n\nexport default function Index() {\n  return <View className='root'><Text>%s</Text></View>;\n}\n", c.Name),
		"apps/miniapp/src/pages/index/index.css": ".root { padding: 32px; font-size: 20px; }\n",
	}
	if c.WeChat {
		projectJSON, err := jsonFile(map[string]any{
			"miniprogramRoot": "./dist/weapp", "projectname": c.Name,
			"appid": "touristappid", "compileType": "miniprogram",
			"setting": map[string]bool{"urlCheck": true},
		})
		if err != nil {
			return nil, err
		}
		files["apps/miniapp/project.config.json"] = projectJSON
	}
	return files, nil
}
