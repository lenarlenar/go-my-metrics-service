# cmd/server

В данной директории будет содержаться код Сервера, который скомпилируется в бинарное приложение

# Сборка и запуск

Сборка (из корня проекта):
go build -ldflags "-X main.buildVersion=1.0.2 -X main.buildDate=2025-05-06 -X main.buildCommit=abc123" -o agent cmd/agent/main.go

Запуск (из корня проекта):
./agent