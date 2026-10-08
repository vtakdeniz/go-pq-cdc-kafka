.PHONY: up down logs tidy connector web

up:
	docker compose up -d

down:
	docker compose down -v

logs:
	docker compose logs -f

tidy:
	go mod tidy

connector:
	go run ./cmd/connector

web:
	go run ./cmd/web
