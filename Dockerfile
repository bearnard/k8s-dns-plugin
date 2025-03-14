FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o k8sdns cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/k8sdns .

CMD ["./k8sdns"]