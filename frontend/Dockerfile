# ============================================
# Dockerfile para CloudPulse AI - Frontend
# Multi-stage build: Node build → Nginx serve
# ============================================

# ---------- ETAPA 1: COMPILAR ----------
FROM node:18-alpine AS builder

# Directorio de trabajo
WORKDIR /app

# Copiar manifiestos de dependencias primero (aprovechar cache)
COPY package.json package-lock.json* ./

# Instalar dependencias
# --frozen-lockfile: requiere que package-lock.json esté sincronizado
# --no-cache: no guarda cache de npm (imagen más limpia)
RUN npm ci --no-cache

# Copiar código fuente
COPY . .

# Argumento de build (variable en build time)
ARG VITE_API_URL=http://localhost:8080
ARG VITE_WS_URL=ws://localhost:8080
ARG VITE_APP_NAME="CloudPulse AI"
ARG VITE_APP_VERSION=0.1.0

# Pasar como env vars a Vite
ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_WS_URL=${VITE_WS_URL}
ENV VITE_APP_NAME=${VITE_APP_NAME}
ENV VITE_APP_VERSION=${VITE_APP_VERSION}

# Compilar el frontend
RUN npm run build

# ---------- ETAPA 2: SERVIR ----------
FROM nginx:1.25-alpine

# Instalar herramientas útiles
RUN apk add --no-cache curl

# Copiar configuración de Nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copiar el build compilado desde la etapa builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Timezone
ENV TZ=America/Havana

# Exponer puerto HTTP
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl --fail http://localhost/health || exit 1

# Nginx corre en foreground por defecto
CMD ["nginx", "-g", "daemon off;"]