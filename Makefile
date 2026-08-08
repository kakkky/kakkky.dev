COMPOSE := docker compose -f app/compose.yml
COMPOSE_EXEC := $(COMPOSE) exec app

.PHONY: dev.up dev.down dev.restart dev.build dev.logs dev.sh wire.gen \
	migrate.up migrate.down migrate.version migrate.create schema.dump \
	seed

# ---- docker for development ----
dev.up:
	$(COMPOSE) up -d

dev.build:
	$(COMPOSE) build app

dev.down:
	$(COMPOSE) down

dev.restart:
	$(COMPOSE) restart app

dev.logs:
	$(COMPOSE) logs -f app

dev.sh:
	$(COMPOSE_EXEC) bash

# ---- wire (DI) ----
wire.gen:
	$(COMPOSE_EXEC) go run github.com/google/wire/cmd/wire@latest gen ./...

# ---- migration ----
MIGRATIONS_DIR := ./app/driver/db/migrations
SCHEMA_FILE := ./app/driver/db/schema/schema.sql
LOCAL_DATABASE_URL := postgres://dev-user:pswd@localhost:5432/dev-db?sslmode=disable

migrate.up:
	migrate -path=$(MIGRATIONS_DIR) -database="$(LOCAL_DATABASE_URL)" up
	$(MAKE) schema.dump

migrate.down:
	migrate -path=$(MIGRATIONS_DIR) -database="$(LOCAL_DATABASE_URL)" down 1
	$(MAKE) schema.dump

migrate.version:
	migrate -path=$(MIGRATIONS_DIR) -database="$(LOCAL_DATABASE_URL)" version

migrate.create:
	@[ -n "$(NAME)" ] || (echo "Usage: make migrate.create NAME=<name>"; exit 1)
	@mkdir -p $(MIGRATIONS_DIR)
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

schema.dump:
	atlas schema inspect \
		-u "$(LOCAL_DATABASE_URL)" \
		--exclude "schema_migrations" \
		--format '{{ sql . }}' \
		> $(SCHEMA_FILE)

# ---- seed (開発環境専用) ----
SEED_FILE := ./app/driver/db/seeds/seed.sql

seed:
	psql "$(LOCAL_DATABASE_URL)" -f $(SEED_FILE)
