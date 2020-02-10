GO=go

SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

VERSION := v0.1.0
TARGETS := scheduler
project=github.com/fezho/k8s-examples

.PHONY: all
all: check build

.PHONY: build
build: $(TARGETS) $(TEST_TARGETS)

$(TARGETS): $(SRC)
	$(GO) build $(project)/03-custom-scheduler/cmd/$@

.PHONY: image
image:
	docker build -t docker.pkg.github.com/fezho/k8s-examples/kube-custom-scheduler:$(VERSION) .

.PHONY: lint
lint:
	@golangci-lint run --deadline=5m

packages = $(shell go list ./...|grep -v /vendor/)
test: check
	$(GO) test ${packages}

.PHONY: check
check:
	@echo skip go vet

.PHONY: clean
clean:
	rm -f $(TARGETS)

