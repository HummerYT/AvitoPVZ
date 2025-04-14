#!/bin/bash
set -e

echo "Запуск тестов..."
go test -coverprofile=coverage.out ./... -short

echo "Фильтрация покрытия (исключаем моки)..."
grep -v "/mocks/" coverage.out > coverage_filtered.out

echo "Результат покрытия (без моков):"
go tool cover -func=coverage_filtered.out
