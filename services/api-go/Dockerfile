# ============================================
# Dockerfile para la API de CloudPulse AI
# ============================================

# ---------- ETAPA 1: COMPILAR ----------
FROM golang:1.27-alpine AS builder

# Instalar git (necesario para descargar dependencias)
RUN apk add --no-cache git

# Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -o /api cmd/main.go

# ---------- ETAPA 2: IMAGEN FINAL ----------
FROM alpine:latest

# Instalar certificados SSL (necesarios para HTTPS)
RUN apk --no-cache add ca-certificates

# Crear usuario no-root
RUN adduser -D -g '' appuser

# Establecer directorio de trabajo
WORKDIR /app

# Copiar el binario compilado
COPY --from=builder /api .

# Copiar documentación Swagger
COPY --from=builder /app/docs ./docs

# Cambiar a usuario no-root
USER appuser

# Exponer puerto
EXPOSE 8080

# Comando para ejecutar
CMD ["./api"]