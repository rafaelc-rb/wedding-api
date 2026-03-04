-include .env
export

.PHONY: build run test clean setup migrate-up migrate-down seed-dev docker-build docker-run docker-stop postman-push

BINARY=bin/api
MAIN=cmd/api/main.go
IMAGE=weddo-api
VERSION?=latest

build:
	go build -o $(BINARY) $(MAIN)

run:
	go run $(MAIN)

test:
	go test ./... -v

clean:
	rm -rf bin/

setup:
	go mod tidy
	@test -f .env || cp .env.example .env
	@echo "Setup concluído. Edite o .env se necessário."

migrate-up:
	go run $(MAIN) -migrate-up

migrate-down:
	go run $(MAIN) -migrate-down

seed-dev:
	go run $(MAIN) -seed-dev

docker-build:
	docker build -t $(IMAGE):$(VERSION) .

docker-run:
	docker run -d --name $(IMAGE) \
		--env-file .env \
		-p 8080:8080 \
		$(IMAGE):$(VERSION)

docker-stop:
	docker stop $(IMAGE) && docker rm $(IMAGE)

postman-push:
	@test -n "$(POSTMAN_API_KEY)" || (echo "Erro: POSTMAN_API_KEY não definida. Preencha no .env ou exporte." && exit 1)
	cd postman && postman login --with-api-key "$(POSTMAN_API_KEY)" && postman workspace push --yes
