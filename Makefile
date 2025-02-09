ALL_MODULES := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort )
TOOLS_MOD_DIR := ./internal/tools

.PHONY: install-tools
install-tools:
	cd $(TOOLS_MOD_DIR) && go install github.com/securego/gosec/v2/cmd/gosec
	cd $(TOOLS_MOD_DIR) && go install github.com/mgechev/revive

.PHONY: for-all
for-all:
	@set -e; for dir in $(ALL_MODULES); do \
	  (cd "$${dir}" && $${CMD} ); \
	done

.PHONY: tidy
tidy:
	$(MAKE) for-all CMD="go mod tidy"

.PHONY: secure
secure:
	gosec -exclude-dir internal/tools ./...

.PHONY: lint
lint:
	revive ./... 

.PHONY: test
test:
	go test -race ./...

.PHONY: release-test
release-test:
	goreleaser release --snapshot --skip=publish --clean
