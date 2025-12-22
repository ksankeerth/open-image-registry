.PHONY: run-server build-server start-dev-server lint-go test-server build-ui test-ui lint-webapp start-dev-ui

build-server:
	mkdir -p server/bin && cd server/cmd && go build -o ../bin/open-image-registry-server

build-ui: 
	cd webapp && npm run build

run-server: build-server
	cd server/bin && ./open-image-registry-server

tests: build-server
	cd server && go test -v ./... -cover


integration: build-server
	cd server && go test -v ./tests/integration/...

# Build images and then start all services defined in docker-compose.yml
.PHONY: compose-up-build
compose-up-build:
	docker compose up --build

# Stop and remove containers, networks created by Compose
.PHONY: compose-down
compose-down:
	docker compose down


# -------------------------------
# Manual Docker (non-compose) targets
# -------------------------------
.PHONY: docker-build-all
docker-build-all:
	docker build -t client-react -f ./webapp/Dockerfile.dev ./webapp
	docker build -t api-golang -f ./server/Dockerfile.dev ./server

.PHONY: docker-run-all
docker-run-all:
	echo "$$DOCKER_COMPOSE_NOTE"

	$(MAKE) docker-stop
	${MAKE} docker-rm

	docker network create open-image-registry-network

	docker run -d \
		--name api-golang \
		--network open-image-registry-network \
		-v $(PWD)/server:/usr/src/app/server \
		-p 8000:8000 \
		--restart unless-stopped \
		api-golang

	docker run -d \
		--name client-react \
		--network open-image-registry-network \
		-v ${PWD}/client-react/vite.config.js:/usr/src/app/vite.config.js \
		-p 3000:3000 \
		--restart unless-stopped \
		client-react

.PHONY: docker-stop

docker-stop:
	-docker stop api-golang
	-docker stop client-react

.PHONY: docker-rm
docker-rm:
	-docker container rm api-golang
	-docker container rm client-react
	-docker network rm open-image-registry-network