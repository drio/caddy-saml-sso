HOST=$(shell hostname)
MOD_NAME=caddy-saml-sso
PRJ_NAME=$(MOD_NAME)

ifeq ($(HOST), air)
include .env.dev
export $(shell sed 's/=.*//' .env.dev)
endif

dev:
	xcaddy run

test-env:
	@echo "saml_root_url=$$SAML_ROOT_URL"

build: xcaddy
	xcaddy build --with github.com/drio/$(MOD_NAME)

xcaddy: 
	go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest

run: caddy Caddyfile
	./caddy run ./Caddyfile

clean:
	rm -f caddy

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
