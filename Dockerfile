FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

FROM alpine

COPY --from=builder /app/main .
