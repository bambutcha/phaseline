.PHONY: help up down logs api-test sqlc web-dev

help:
	@echo "PHASELINE — команды"
	@echo "  make up        docker compose up --build"
	@echo "  make down      docker compose down"
	@echo "  make logs      docker compose logs -f"
	@echo "  make sqlc      generate Go from SQL (apps/api)"
	@echo "  make api-test  go test ./... in apps/api"

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f

sqlc:
	cd apps/api && sqlc generate

api-test:
	cd apps/api && go test ./...
