# Используем минимальный образ Go
FROM golang:1.21-alpine

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем остальные файлы
COPY . .

# Собираем приложение
RUN go build -o main cmd/main.go

# Указываем порт (если необходимо)
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]
