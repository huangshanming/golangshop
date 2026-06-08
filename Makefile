.PHONY: run build seed docker-up docker-down tidy dev

run:
	go run ./cmd/server -config config.yaml

build:
	go build -o bin/goshop ./cmd/server
	go build -o bin/seed ./cmd/seed

seed:
	go run ./cmd/seed -config config.yaml

dev:
	./scripts/dev.sh

tidy:
	go mod tidy

docker-up:
	cd deploy && docker compose up -d mysql redis rabbitmq minio

docker-up-all:
	cd deploy && docker compose up -d

docker-down:
	cd deploy && docker compose down
