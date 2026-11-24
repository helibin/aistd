.PHONY: setup check build release
setup:
	@go version
	@node --version
	cd apps/cli && go mod download
	git config core.hooksPath .githooks
check:
	node --test scripts/check-version.test.mjs
	$(MAKE) -C apps/cli check
	npm test --workspace @aistd/web
	npm run build:web
build:
	$(MAKE) -C apps/cli build
	npm run build:web
release:
	sh apps/cli/scripts/release.sh
	npm run build:web
