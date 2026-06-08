#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "==> starting infrastructure (mysql redis rabbitmq minio)..."
docker compose -f deploy/docker-compose.yml up -d mysql redis rabbitmq minio

echo "==> waiting for mysql..."
for i in $(seq 1 30); do
  if docker compose -f deploy/docker-compose.yml exec -T mysql mysqladmin ping -h localhost -ugoshop -pgoshop123 --silent 2>/dev/null; then
    break
  fi
  sleep 2
done

echo "==> seeding database..."
go run ./cmd/seed -config config.yaml

echo "==> starting server..."
go run ./cmd/server -config config.yaml
