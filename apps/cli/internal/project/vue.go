package project

func vueFiles(c Config) (map[string]string, error) {
	packageJSON, err := jsonFile(map[string]any{
		"name": c.Name + "-web", "private": true, "version": "0.1.0", "type": "module",
		"scripts":         map[string]string{"dev": "vite --host 127.0.0.1", "build": "vue-tsc --noEmit && vite build", "lint": "vue-tsc --noEmit", "test": "node --test tests/*.test.mjs"},
		"dependencies":    map[string]string{"vue": "3.5.45"},
		"devDependencies": map[string]string{"vite": "6.2.6", "@vitejs/plugin-vue": "5.2.4", "vue-tsc": "2.2.12", "typescript": "5.8.3"},
	})
	if err != nil {
		return nil, err
	}
	app := "<script setup lang=\"ts\">\n"
	if c.API {
		app += "import { ref, onMounted } from 'vue';\nconst message = ref('Connecting to API...');\nonMounted(() => { fetch('/api/hello').then(response => response.json()).then(value => { message.value = value.message; }).catch(() => { message.value = 'API unavailable'; }); });\n"
	}
	app += "</script>\n<template><main style=\"font-family: system-ui; max-width: 720px; margin: 4rem auto\"><h1>" + c.Name + "</h1>"
	if c.API {
		app += "<p>API: {{ message }}</p>"
	}
	app += "</main></template>\n"
	vite := "import { defineConfig } from 'vite';\nimport vue from '@vitejs/plugin-vue';\nexport default defineConfig({ plugins: [vue()]"
	if c.API {
		vite += ", server: { proxy: { '/api': 'http://127.0.0.1:8080' } }"
	}
	vite += " });\n"
	return map[string]string{
		"apps/web/package.json":         packageJSON,
		"apps/web/src/App.vue":          app,
		"apps/web/src/main.ts":          "import { createApp } from 'vue';\nimport App from './App.vue';\ncreateApp(App).mount('#root');\n",
		"apps/web/tsconfig.json":        "{\"compilerOptions\":{\"target\":\"ES2022\",\"module\":\"ESNext\",\"moduleResolution\":\"Bundler\",\"strict\":true,\"skipLibCheck\":true,\"lib\":[\"ES2022\",\"DOM\"]},\"include\":[\"src/**/*.ts\",\"src/**/*.vue\"]}\n",
		"apps/web/vite.config.ts":       vite,
		"apps/web/index.html":           "<!doctype html><html lang=\"en\"><head><meta charset=\"UTF-8\"/><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"/><title>" + c.Name + "</title></head><body><div id=\"root\"></div><script type=\"module\" src=\"/src/main.ts\"></script></body></html>\n",
		"apps/web/tests/smoke.test.mjs": "import { test } from 'node:test';\nimport assert from 'node:assert/strict';\nimport { readFileSync } from 'node:fs';\ntest('Vue app mounts into the HTML root', () => { assert.match(readFileSync('index.html', 'utf8'), /id=\"root\"/); assert.match(readFileSync('src/main.ts', 'utf8'), /mount\\('#root'\\)/); });\n",
		"apps/web/Dockerfile":           "FROM node:22-alpine AS build\nWORKDIR /src\nCOPY package*.json ./\nRUN npm install --no-audit --no-fund\nCOPY . .\nRUN npm run build\nFROM caddy:2.10.0-alpine\nCOPY --from=build /src/dist /usr/share/caddy\n",
		"apps/web/.dockerignore":        "node_modules\ndist\n.git\n.env\n.env.*\n",
	}, nil
}
