.PHONY: up down sqlc backend-test

up:
	docker compose up -d postgres

down:
	docker compose down

sqlc:
	docker run --rm --user $$(id -u):$$(id -g) -v "$(CURDIR)/backend:/src" -w /src sqlc/sqlc generate

backend-test:
	cd backend && go test ./...
