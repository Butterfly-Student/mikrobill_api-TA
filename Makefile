include .env
export

APP_NAME=backend-mikrobill
MIGRATIONS_DIR=migrations

# =========================
# DOCKER
# =========================
docker-up:
	podman-compose up -d postgres

docker-down:
	podman-compose down

docker-logs:
	podman logs -f app_postgres

# =========================
# GO COMMANDS
# =========================
run:
	go run ./cmd/server

build:
	go build -o bin/$(APP_NAME) ./cmd/server

test:
	go test ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

# =========================
# MIGRATIONS
# =========================
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq -digits 3 $(name) 

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-force:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(version)

migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

# =========================
# HELP
# =========================
help:
	@echo "Available commands:"
	@echo "  make run"
	@echo "  make build"
	@echo "  make test"
	@echo "  make fmt"
	@echo "  make tidy"
	@echo "  make migrate-create name=create_users"
	@echo "  make migrate-up"
	@echo "  make migrate-down"
	@echo "  make migrate-version"
