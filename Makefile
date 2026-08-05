.PHONY: docs build-dev build-local build fmt fmt-check test python-test testacc vet staticcheck check

PKG ?=
TEST ?=
DEV_DIR ?= $(CURDIR)/build

docs:
	go generate ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

fmt-check:
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))" || \
		(echo "Go files need formatting:"; gofmt -l $$(find . -name '*.go' -not -path './vendor/*'); exit 1)

test:
	go test ./...

python-test:
	PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts -p 'test_*.py'

testacc:
ifdef PKG
	TF_ACC=1 go test -v -p 1 -timeout 120m $(if $(TEST),-run $(TEST)) ./internal/service/$(PKG)/...
else
	TF_ACC=1 go test -v -p 1 -timeout 120m $(if $(TEST),-run $(TEST)) ./...
endif

vet:
	go vet ./...

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 ./...

check: fmt-check test python-test vet staticcheck

build-dev:
	mkdir -p "$(DEV_DIR)"
	go build -o "$(DEV_DIR)/terraform-provider-opnsense" .

build-local:
	go build -o ~/.terraform.d/plugins/dev.io/biptec/opnsense/1.0.0/$$(go env GOOS)_$$(go env GOARCH)/terraform-provider-opnsense .

build:
	go build -o terraform-provider-opnsense .
