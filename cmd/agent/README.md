# cmd/agent

В данной директории будет содержаться код Агента, который скомпилируется в бинарное приложение

# Сборка и запуск

Сборка (из корня проекта):
go build -ldflags "-X main.buildVersion=1.0.2 -X main.buildDate=2025-05-06 -X main.buildCommit=abc123" -o server cmd/server/main.go

Запуск (из корня проекта):
./server