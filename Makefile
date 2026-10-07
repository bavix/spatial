.PHONY: check check-min modernize lint lint-fix fmt

GO_MIN_VERSION := go1.26.0

GOLANGCI_LINT_VERSION := v2.14.0
GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

check:
	go build ./...
	go vet ./...

check-min:
	GOTOOLCHAIN=$(GO_MIN_VERSION) go build ./...
	GOTOOLCHAIN=$(GO_MIN_VERSION) go vet ./...

modernize:
	go fix ./...

lint:
	$(GOLANGCI_LINT) run $(args)

lint-fix:
	$(MAKE) lint args=--fix

fmt:
	$(GOLANGCI_LINT) fmt
