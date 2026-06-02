VERSION ?= dev
COMMIT ?= local
LDFLAGS := -s -w -X xui-next/internal/version.Version=$(VERSION) -X xui-next/internal/version.Commit=$(COMMIT)

.PHONY: build build-web package

build:
	go build -ldflags "$(LDFLAGS)" -o dist/xuid ./cmd/xuid
	go build -ldflags "$(LDFLAGS)" -o dist/xui-agent ./cmd/xui-agent

build-web:
	cd web && npm install && npm run build

package: build
	mkdir -p dist/package
	cp dist/xuid dist/package/xuid
	cp dist/xui-agent dist/package/xui-agent
	cp scripts/xui-pro.sh dist/package/xui-pro
	cp -R reality dist/package/reality
	chmod +x dist/package/xuid dist/package/xui-agent dist/package/xui-pro
	tar -C dist/package -czf dist/xui-pro-linux-local.tar.gz .
