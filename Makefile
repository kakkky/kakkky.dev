COMPOSE := docker compose -f app/compose.yml
COMPOSE_EXEC := $(COMPOSE) exec app

.PHONY: dev.up dev.down dev.restart dev.build dev.logs dev.sh wire.gen

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
