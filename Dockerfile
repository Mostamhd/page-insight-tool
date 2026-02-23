# Stage 1 — Frontend Build
FROM node:20-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2 — Go Build
FROM golang:1.24-alpine AS go-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY --from=frontend-build /app/frontend/dist/ frontend/dist/
RUN go test ./...
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server

# Stage 3 — Runtime
FROM alpine:3.19
COPY --from=go-build /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
