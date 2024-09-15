FROM golang:1.23 AS builder
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /usr/local/bin/app cmd/server/main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /usr/local/bin/app .
COPY migrations ./migrations/

EXPOSE 8080

CMD ["/app/app"]
