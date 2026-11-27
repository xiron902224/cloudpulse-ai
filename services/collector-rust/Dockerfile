# ============================================
# Dockerfile para CloudPulse Collector (Rust)
# Multi-stage build para imagen mínima
# ============================================

# ---------- ETAPA 1: COMPILAR ----------
FROM rust:1.75-alpine AS builder

# Instalar dependencias del sistema necesarias para compilar
RUN apk add --no-cache \
    musl-dev \
    openssl-dev \
    openssl-libs-static \
    pkgconfig

# Establecer directorio de trabajo
WORKDIR /app

# Copiar manifiestos de dependencias primero (para aprovechar cache)
COPY Cargo.toml ./

# Crear estructura mínima para compilar dependencias
RUN mkdir -p src && \
    echo "fn main() {}" > src/main.rs && \
    cargo build --release && \
    rm -rf src

# Copiar el código fuente real
COPY src ./src

# Compilar la aplicación
RUN cargo build --release

# ---------- ETAPA 2: IMAGEN FINAL ----------
FROM alpine:3.19

# Instalar certificados SSL y timezone
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# Configurar timezone
ENV TZ=America/Havana

# Crear usuario no-root
RUN adduser -D -g '' appuser

# Establecer directorio de trabajo
WORKDIR /app

# Copiar el binario compilado desde la etapa builder
COPY --from=builder /app/target/release/cloudpulse-collector .

# Cambiar a usuario no-root
USER appuser

# Exponer puerto HTTP
EXPOSE 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9090/health || exit 1

# Comando de inicio
CMD ["./cloudpulse-collector"]