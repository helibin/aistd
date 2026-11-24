.PHONY: build check verify verify-deep
build:
	mkdir -p bin
	go build -o bin/aistd ./cmd/aistd
check:
	@gofmt_files=$$(gofmt -l version.go cmd internal); test -z "$$gofmt_files" || { printf 'Run gofmt first:\n%s\n' "$$gofmt_files" >&2; exit 1; }
	go vet ./...
	go test ./...

# Offline: CLI lifecycle plus generated API checks/build, only under TMPDIR.
verify:
	sh scripts/verify.sh offline

# Network/dependency-heavy: override PLATFORMS="web", "expo", "vben soybean", etc.
PLATFORMS ?= api nestjs web vue expo taro uniapp
verify-deep:
	sh scripts/verify.sh deep $(PLATFORMS)
