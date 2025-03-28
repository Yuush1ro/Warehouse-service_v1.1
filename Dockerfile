# Используем официальное изображение Go для сборки
FROM golang:1.24 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum и устанавливаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код в контейнер
COPY . .

# Компилируем бинарник
RUN go build -o warehouse-service ./cmd/server

# Устанавливаем migrate (теперь в builder!)
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Создаём финальный образ
FROM debian:bookworm-slim

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем собранный бинарник из предыдущего контейнера
COPY --from=builder /app/warehouse-service .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Указываем порт, который будет использоваться
EXPOSE 8080

# Запускаем приложение
CMD ["./warehouse-service"]
