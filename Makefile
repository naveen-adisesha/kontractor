GOFLAGS := -trimpath -ldflags="-s -w"
BINDIR  := bin

.PHONY: build build-validate build-post-render clean test

build: build-validate build-post-render

build-validate:
	@mkdir -p $(BINDIR)
	go build $(GOFLAGS) -o $(BINDIR)/kontractor-validate ./cmd/kontractor-validate/

build-post-render:
	@mkdir -p $(BINDIR)
	go build $(GOFLAGS) -o $(BINDIR)/kontractor-post-render ./cmd/kontractor-post-render/

clean:
	rm -rf $(BINDIR)

test:
	go test ./...

install-plugin: build
	@echo "Installing Helm plugin..."
	helm plugin install $(shell pwd) || true
