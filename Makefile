.PHONY: all build build-arm64 clean help

CGO_CFLAGS=-mmacosx-version-min=13.1
GIT_COMMIT = $(shell git rev-parse HEAD 2> /dev/null || true)
GO_LDFLAGS = -ldflags
BUILD_FLAGS = \
	$(if $(GIT_COMMIT),-X bauklotze/pkg/machine/define.GitCommit=$(GIT_COMMIT),)
CODESIGN_IDENTITY ?= -

all: help

##@
##@ Build commands
##@
build: ##@ Build binaries for all architectures
	@echo "Build commit:  ${GIT_COMMIT}"
	@echo "GO_LDFLAGS: ${GO_LDFLAGS}"
	@$(MAKE) out/bin/ovm-arm64

build-arm64: ##@ Build arm64 binary
	@$(MAKE) out/bin/ovm-arm64

out/bin/ovm-arm64: out/bin/ovm-%:
	@mkdir -p $(@D)
	CGO_ENABLED=1 CGO_CFLAGS=$(CGO_CFLAGS) CGO_CFLAGS=$(CGO_CFLAGS) GOOS=darwin GOARCH=$* go build $(GO_LDFLAGS) '$(BUILD_FLAGS)' -o $@ bauklotze/cmd
	codesign --force --options runtime --sign $(CODESIGN_IDENTITY) $@

##@
##@ Clean commands
##@
clean: ##@ Clean up build artifacts
	$(RM) -rf out/bin/

##@
##@ Misc commands
##@

lint: ##@ Run linter
	@golangci-lint run

help: ##@ (Default) Print listing of key targets with their descriptions
	@printf "\nUsage: make <command>\n"
	@grep -F -h "##@" $(MAKEFILE_LIST) | grep -F -v grep -F | sed -e 's/\\$$//' | awk 'BEGIN {FS = ":*[[:space:]]*##@[[:space:]]*"}; \
	{ \
		if($$2 == "") \
			printf ""; \
		else if($$0 ~ /^#/) \
			printf "\n%s\n", $$2; \
		else if($$1 == "") \
			printf "     %-20s%s\n", "", $$2; \
		else \
			printf "\n    \033[34m%-20s\033[0m %s\n", $$1, $$2; \
	}'
