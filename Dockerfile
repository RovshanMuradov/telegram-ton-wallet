FROM golang:1.22-alpine

WORKDIR /app

# Installing the required dependencies
RUN apk add --no-cache git

# Copying project files
COPY . .

# Installing Go Dependencies
RUN go mod download

# Assembling the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

# Installing migrate
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copying migrations
COPY migrations /migrations

EXPOSE 8080

RUN chmod -R 777 /migrations/migrations

CMD ["./main"]

