# --- Estágio 1: build do frontend (Vite emite para ../public) ---
FROM node:24-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# --- Estágio 2: build do backend ---
FROM golang:1.25-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/migrations ./cmd/migrations

# --- Estágio 3: imagem final ---
FROM alpine:3.21
# ca-certificates: chamadas HTTPS (Mercado Pago, SMTP); tzdata: horários corretos
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /out/server /out/migrations ./
COPY --from=frontend /app/public ./public
# o runner de migrations lê db/schema relativo ao CWD
COPY db/schema ./db/schema
EXPOSE 8080
CMD ["/bin/sh", "-c", "./migrations up && exec ./server"]
