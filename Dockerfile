# Frontend Stage
FROM node:25-alpine AS frontend-build
WORKDIR /frontend

COPY frontend/package*.json ./

RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY frontend ./

ENV NODE_OPTIONS=--max-old-space-size=1024

RUN npm run build \
 && rm -rf node_modules

# Backend Stage
FROM golang:1.26.1-alpine AS backend-build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -trimpath -ldflags=-s -o /pingopher ./cmd/app

# Application Stage
FROM alpine:3.23

RUN apk add --no-cache tzdata

COPY --from=backend-build /pingopher /usr/local/bin/pingopher
COPY --from=frontend-build /frontend/dist ./frontend/dist

USER nobody:nobody
CMD ["pingopher"]