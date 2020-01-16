GO=go

SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

VERSION := v0.1.0
TARGETS := demo
project=github/fezho/k8s-examples

all: check build

build: $(TARGETS) $(TEST_TARGETS)

$(TARGETS): $(SRC)
	$(GO) build $(project)/01-leader-election/cmd/$@

image:
	docker build -t docker.pkg.github.com/fezho/k8s-examples/kube-leaderelection-demo:$(VERSION) .

.PHONY: clean all build check image

lint:
	@golangci-lint run --deadline=5m

packages = $(shell go list ./...|grep -v /vendor/)
test: check
	$(GO) test ${packages}

cov: check
	gocov test $(packages) | gocov-html > coverage.html

check:
	@echo skip go vet

clean:
	rm -f $(TARGETS)
