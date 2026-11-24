package project

const adminPNPM = "npx --yes pnpm@10.33.4 --dir apps/admin"

func adminLabel(c Config) string {
	if c.AdminStack == "soybean" {
		return "Soybean-admin"
	}
	return "Vben Naive UI"
}

func adminCheck(c Config) string {
	if c.AdminStack == "soybean" {
		return adminPNPM + " run typecheck"
	}
	return adminPNPM + " --filter @vben/web-naive run typecheck"
}

func adminBuild(c Config) string {
	if c.AdminStack == "soybean" {
		return adminPNPM + " run build"
	}
	return adminPNPM + " run build:naive"
}

func adminRun(c Config) string {
	if c.AdminStack == "soybean" {
		return adminPNPM + " run dev"
	}
	return adminPNPM + " run dev:naive"
}

func adminOpsFiles(c Config) map[string]string {
	build, output := "build:naive", "apps/web-naive/dist"
	if c.AdminStack == "soybean" {
		build, output = "build", "dist"
	}
	return map[string]string{
		"apps/admin/AISTD.Dockerfile":              "FROM node:22-alpine AS build\nWORKDIR /app\nCOPY . .\nENV CI=1 SKIP_INSTALL_SIMPLE_GIT_HOOKS=1\nRUN npx --yes pnpm@10.33.4 install --frozen-lockfile\nRUN npx --yes pnpm@10.33.4 run " + build + "\n\nFROM caddy:2.10.0-alpine\nCOPY --from=build /app/" + output + " /srv/admin\nCOPY AISTD.Caddyfile /etc/caddy/Caddyfile\nEXPOSE 80\n",
		"apps/admin/AISTD.Caddyfile":               ":80 {\n  root * /srv\n  encode zstd gzip\n  redir / /admin/ 302\n  try_files {path} /admin/index.html\n  file_server\n}\n",
		"apps/admin/AISTD.Dockerfile.dockerignore": "**/node_modules\n**/dist\n**/.turbo\n.git\n**/.env.local\n**/.env.*.local\n",
	}
}
