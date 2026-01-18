# https://just.systems

set windows-shell := ["pwsh.exe", "-NoLogo", "-Command"]
set shell := ["sh", "-c"]
set quiet := true
set dotenv-load := true

# Команда по умолчанию
default: dev

# Показать список команд
help:
    just --list

# Запуск сервера
[group("server")]
run:
    go run ./cmd/server/main.go

# Установка зависимостей
[group("server")]
install:
    go get -tool github.com/a-h/templ/cmd/templ@latest
    go get -tool github.com/jackc/tern/v2@latest
    go get -tool github.com/air-verse/air@latest

# Запуск всех тестов
[group("testing")]
test:
    go test ./... -v

# Запуск тестов с покрытием кода
[group("testing")]
test-coverage:
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage -o coverage.html

# Запуск тестов с детектором гонок
[group("testing")]
test-race:
    go test ./... -race

# Запуск бенчмарков
[group("testing")]
test-bench:
    go test ./... -bench=. -benchmem

# Запуск тестов с подсчетом покрытия в процентах
[group("testing")]
test-coverage-short:
    go test ./... -cover

# Запуск тестов конкретного пакета
[group("testing")]
test-pkg PACKAGE=".":
    go test ./{{ PACKAGE }} -v

# Запуск тестов с отображением только неудачных
[group("testing")]
test-fail:
    go test ./... | grep -E "(FAIL|PASS:|ok\s|^ok\s|^---)"

# Генерация шаблонов
[group("templ")]
gen-templ:
    go tool templ generate

# Генерация отчетов статического анализа
[group("quality")]
lint:
    go vet ./...

# Форматирование кода
[group("quality")]
fmt:
    go fmt ./...

# Сборка проекта
[group("build")]
build:
    go build -o ./bin/server ./cmd/server/main.go

# Сборка с оптимизациями
[group("build")]
build-release:
    go build -ldflags="-s -w" -o ./bin/server ./cmd/server/main.go

# Запуск в режиме разработки с автоперезагрузкой (если установлен air)
[group("development")]
dev:
    go tool air

# Создать новую миграцию
[group("database")]
new-table TABLE:
    go tool tern new {{ TABLE }}

# Миграции базы данных (пример для golang-migrate)
[group("database")]
migrate-up:
    go tool tern migrate

[group("database")]
migrate-down DOWN:
    go tool tern migrate --destination -{{ DOWN }}

#

# Запуск всех проверок перед коммитом
[group("ci")]
pre-commit: fmt lint test

# Просмотр зависимостей
[group("deps")]
deps:
    go list -m all

# Обновление зависимостей
[group("deps")]
update-deps:
    go get -u ./...
    go mod tidy

# Просмотр неиспользуемых зависимостей
[group("analysis")]
unused-deps:
    go mod tidy -v
