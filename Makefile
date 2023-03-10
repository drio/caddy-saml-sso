HOST=$(shell hostname)
MOD_NAME=caddy-saml-sso
PRJ_NAME=$(MOD_NAME)
BINS=caddy.arm64.osx caddy.amd64.linux caddy.amd64.windows

VERSION=$(shell cat version.go | tail -1| awk -F\" '{print $$2}')

ifeq ($(HOST), air)
include .env.dev
export $(shell sed 's/=.*//' .env.dev)
endif

dev:
	xcaddy run

test-env:
	@echo "saml_root_url=$$SAML_ROOT_URL"

.PHONY: build-all
build-all: $(BINS)

caddy.arm64.osx: xcaddy
	xcaddy build --with github.com/drio/$(MOD_NAME) --output $@

caddy.amd64.linux:
	GOARCH=amd64 GOOS=linux xcaddy build --with github.com/drio/$(MOD_NAME) --output $@

caddy.amd64.windows:
	GOARCH=amd64 GOOS=windows xcaddy build --with github.com/drio/$(MOD_NAME) --output $@

xcaddy:
	go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest

# This is useful if you want to play with the config file
# Use caddy reload to make Caddy reload the config
run: caddy Caddyfile
	./caddy run ./Caddyfile

clean:
	rm -f caddy $(BINS)

.PHONY: test single-run-test lint
test:
	@ls *.go | entr -c -s 'go test -v . && notify "💚" || notify "🛑"'

single-run-test:
	go test -v .

lint:
	golangci-lint run

saml-cert:
	mkdir saml-cert
	openssl req -x509 -newkey rsa:2048 \
		-keyout saml-cert/service.key \
		-out saml-cert/service.cert \
		-days 365 -nodes -subj "/CN=$(DOMAIN)"

.PHONY: metadata
metadata:
	@curl $$SAML_ROOT_URL/saml/metadata

docker/login:
	export CR_PAT=`cat ./.gh-token`;\
	echo $$CR_PAT | docker login ghcr.io -u USERNAME --password-stdin

docker/build:
	docker build -t ghcr.io/drio/caddy-saml-sso:$(VERSION) .

docker/publish: docker/build
	docker push ghcr.io/drio/caddy-saml-sso:$(VERSION)

docker/run:
	docker run -p 12000:12000 \
		--env-file=.env.dev \
		-v ./saml-cert:/saml-cert \
		-v ./Caddyfile:/etc/caddy/Caddyfile \
		ghcr.io/drio/caddy-saml-sso:$(VERSION) \
		/usr/bin/caddy run --config=/etc/caddy/Caddyfile

release:
	./release.sh
	make docker/publish

tunnel/%:
	./scripts/$*-tunnel.sh
