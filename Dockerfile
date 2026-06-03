# Frontend Stage
FROM node:26-alpine AS frontend-build
WORKDIR /frontend

COPY frontend/package*.json ./

RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY frontend ./

ENV NODE_OPTIONS=--max-old-space-size=1024

RUN npm run build \
 && rm -rf node_modules

# Backend Stage
FROM golang:1.26-alpine AS backend-build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN go build -trimpath -ldflags=-s -o /pingopher ./cmd/app

# Application Stage
FROM gcr.io/distroless/static-debian12:latest
WORKDIR /app

COPY --from=backend-build /pingopher /app/bin/pingopher
COPY --from=frontend-build /frontend/dist /app/frontend/dist

ENTRYPOINT ["/app/bin/pingopher"]