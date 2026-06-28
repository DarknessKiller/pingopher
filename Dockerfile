# Frontend Stage
FROM node:26-alpine AS frontend-build
WORKDIR /frontend
COPY frontend/package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci
COPY frontend ./
RUN npm run build

# Backend Stage
FROM golang:1.26-alpine AS backend-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o pingopher ./cmd/app

# Application Stage
FROM scratch
WORKDIR /app

COPY --from=backend-build /app/pingopher /app/bin/pingopher
COPY --from=frontend-build /frontend/dist /app/frontend/dist

ENV ENV=production
ENV PINGOPHER_HOST=0.0.0.0
ENV PINGOPHER_PORT=4112
ENV PINGOPHER_MAX_RETRY_INTERVAL=900

EXPOSE ${PINGOPHER_PORT}
ENTRYPOINT ["/app/bin/pingopher"]
