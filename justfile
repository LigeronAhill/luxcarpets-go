# https://just.systems

set windows-shell := ["pwsh.exe", "-NoLogo", "-Command"]
set shell := ["sh", "-c"]
set quiet := true
set dotenv-load := true

# Команда по умолчанию
default: help

# Показать список команд
help:
    just --list

# Запуск сервера
[group("server")]
run:
    go run ./cmd/server/main.go
