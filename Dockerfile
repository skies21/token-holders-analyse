FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
COPY ./cmd ./
RUN go build -o main .

FROM alpine

COPY --from=builder /app/main .

CMD ["./main"]
