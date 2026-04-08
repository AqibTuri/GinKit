.PHONY: deps tidy run test migrate-up migrate-down docker-up docker-down lint swagger

deps:
	go mod download

tidy:
	go mod tidy

run:
	go run ./cmd/api

test:
	go test ./... -count=1 -race

# Install CLI once: go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate-up:
	migrate -path migrations -database "$${DATABASE_URL}" up

migrate-down:
	migrate -path migrations -database "$${DATABASE_URL}" down 1

docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down

lint:
	golangci-lint run ./...

# Regenerate OpenAPI docs under docs/ after changing // @Summary etc. on handlers
swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g cmd/api/main.go -o docs --parseInternal
