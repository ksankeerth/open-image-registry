.PHONY: run-server build-server start-dev-server lint-go test-server build-ui test-ui lint-webapp start-dev-ui

build-server:
	mkdir -p server/bin && cd server/cmd && go build -o ../bin/open-image-registry-server

build-ui: 
	cd webapp && npm run build

run-server: build-server build-ui
	cd server/bin && ./open-image-registry-server --webapp-build-path=$(shell pwd)/webapp/build
	