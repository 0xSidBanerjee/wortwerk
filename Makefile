.PHONY: install test lint build dev generate-types check-bundle-size go-test site serve

GO ?= go
NPM ?= npm

install:
	cd web && $(NPM) ci || $(NPM) install

go-test:
	$(GO) test ./...

web-test:
	cd web && $(NPM) run typecheck

test: go-test web-test

lint:
	$(GO) vet ./...
	cd web && $(NPM) run lint

generate-types:
	cd web && $(NPM) run generate-types

build:
	cd web && $(NPM) run build
	rsync -a content/ web/dist/content/

check-bundle-size: build
	cd web && $(NPM) run check-bundle-size

dev:
	cd web && $(NPM) run dev

serve:
	$(GO) run ./cmd/wortwerk serve

site: build
	@echo "Static site in web/dist (includes content/)"
