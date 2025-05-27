.PHONY: run-server build-server start-dev-server lint-go test-server build-ui test-ui lint-webapp start-dev-ui

build-server:
	mkdir -p bin && cd server && go build -o ../bin

build-ui: 
	cd webapp && npm run build

run-server: build-server build-ui
	cd bin && ./open-image-registry --webapp-build-path=$(shell pwd)/webapp/build
	