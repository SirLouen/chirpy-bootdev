include .env

GOOSE := goose -dir sql/schema postgres $(DB_URL)

.PHONY: migrate-up migrate-down migrate-status migrate-reset migrate-create

migrate-up:
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-status:
	$(GOOSE) status

migrate-reset:
	$(GOOSE) reset

migrate-create:
	@test -n "$(name)" || (echo "Usage: make migrate-create name=<migration_name>" && exit 1)
	$(GOOSE) create $(name) sql
