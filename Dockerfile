FROM golang:1.25.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./cmd/app/pingopher ./cmd/app

FROM alpine:3.22

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/cmd/app/pingopher /usr/local/bin/pingopher

USER nobody:nobody

ENTRYPOINT ["pingopher"]