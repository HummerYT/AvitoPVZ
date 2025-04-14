.PHONY: up

up:
	docker build -t avitopvz .
	docker compose up -d app

down:
	docker compose down

unit:
	go test ./... -short

cover:
	sh run_tests.sh

integration: up
	go test test/integration
	docker compose down