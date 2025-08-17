# Этап сборки
FROM golang:1.24 AS builder

WORKDIR /app

# Копируем зависимости и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект, включая миграции
COPY . .

# Сборка бинарника
RUN go build -o server ./cmd/

# Минимальный образ для запуска
FROM debian:bookworm-slim

WORKDIR /app

# Копируем бинарник и миграции
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

# Копируем .env (если хочешь встроить в контейнер)
COPY --from=builder /app/.env .

# Открываем порт
EXPOSE 8080

# Запуск сервера
CMD ["./server"]