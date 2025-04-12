.PHONY: up

up:
	docker build -t avitopvz .
	docker compose up -d app

down:
	docker compose down

unit:
	go test ./... -short

cover:
	go test -coverprofile="coverage.out" ./... -short
	go tool cover -func="coverage.out"

integration: up
	go test test/integration
	docker compose down