# ============================================
# Dockerfile para la API de CloudPulse AI
# Multi-stage build optimizado
# ============================================

# ---------- ETAPA 1: BUILDER ----------
FROM golang:1.27-alpine AS builder

# Metadata de la imagen
LABEL stage=builder
LABEL description="Builder para CloudPulse AI API"

# Instalar dependencias necesarias
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata
# ↑ --no-cache: no guarda índice de paquetes (imagen más pequeña)
# ↑ git: necesario para go mod download
# ↑ ca-certificates: certificados SSL
# ↑ tzdata: zonas horarias

# Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos de dependencias PRIMERO (para aprovechar caché)
COPY go.mod go.sum ./
# ↑ Al copiar solo go.mod/go.sum primero:
# ↑ Si no cambian, Docker reutiliza la capa de go mod download
# ↑ Builds mucho más rápidos cuando solo cambia código

# Descargar dependencias
RUN go mod download
# ↑ Descarga todas las dependencias al caché de Go

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -ldflags="-w -s" \
    -o /api \
    cmd/main.go
# ↑ CGO_ENABLED=0: compilación estática (sin dependencias C)
# ↑ GOOS=linux: para Linux
# ↑ -ldflags="-w -s":
#   -w: elimina información de debug
#   -s: elimina tabla de símbolos
#   Resultado: binario ~30% más pequeño
# ↑ -o /api: binario de salida
# ↑ cmd/main.go: archivo a compilar

# ---------- ETAPA 2: RUNTIME ----------
FROM alpine:3.19

# Metadata
LABEL maintainer="xiron9090 <tu@email.com>"
LABEL description="CloudPulse AI API"
LABEL version="0.1.0"

# Instalar dependencias mínimas
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget
# ↑ ca-certificates: para HTTPS
# ↑ tzdata: para zonas horarias
# ↑ wget: para el healthcheck

# Crear usuario no-root
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app
# ↑ addgroup: crea grupo "app" con GID 1000
# ↑ adduser: crea usuario "app" con UID 1000
# ↑ -D: sin contraseña
# ↑ -G app: grupo "app"

# Establecer directorio de trabajo
WORKDIR /app

# Copiar binario desde la etapa builder
COPY --from=builder /api .
# ↑ --from=builder: copia desde la etapa anterior
# ↑ /api: archivo a copiar
# ↑ .: destino (WORKDIR actual)

# Copiar documentación Swagger
COPY --from=builder /app/docs ./docs
# ↑ Swagger UI necesita estos archivos

# Copiar archivos estáticos si los hay
# COPY --from=builder /app/static ./static

# Cambiar propietario de los archivos
RUN chown -R app:app /app
# ↑ Asegura que el usuario "app" pueda leer los archivos

# Cambiar a usuario no-root
USER app
# ↑ IMPORTANTE: no correr como root
# ↑ Si hay una vulnerabilidad, el atacante tiene permisos limitados

# Exponer puerto
EXPOSE 8080
# ↑ Documenta qué puerto usa el contenedor
# ↑ No abre el puerto, solo lo documenta

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
# ↑ --interval=30s: cada 30 segundos
# ↑ --timeout=10s: máximo 10 segundos
# ↑ --start-period=5s: espera 5 segundos antes del primer check
# ↑ --retries=3: falla después de 3 intentos
# ↑ wget --spider: verifica sin descargar
# ↑ exit 1: si falla, Docker marca el contenedor como unhealthy

# Comando por defecto
CMD ["./api"]
# ↑ Ejecuta el binario
# ↑ Usar CMD (no ENTRYPOINT) permite override

# ---------- METADATA ----------
# Variables de entorno por defecto
ENV PORT=8080
ENV ENVIRONMENT=production
# ↑ Valores por defecto que se pueden sobrescribir