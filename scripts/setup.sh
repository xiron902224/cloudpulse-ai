#!/bin/bash

# ============================================
# Setup inicial para CloudPulse AI
# ============================================

# Colores para mensajes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # Sin color

echo ""
echo "=========================================="
echo "  🚀 Setup de CloudPulse AI"
echo "=========================================="
echo ""

# ---------- 1. VERIFICAR HERRAMIENTAS ----------
echo "📋 Verificando herramientas instaladas..."
echo ""

# Verificar Go
if command -v go &> /dev/null; then
    echo -e "${GREEN}✅ Go instalado:${NC} $(go version)"
else
    echo -e "${RED}❌ Go no está instalado${NC}"
    echo "   Descarga desde: https://golang.org/dl/"
    exit 1
fi

# Verificar Rust
if command -v cargo &> /dev/null; then
    echo -e "${GREEN}✅ Rust instalado:${NC} $(rustc --version)"
else
    echo -e "${RED}❌ Rust no está instalado${NC}"
    echo "   Instala con: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
    exit 1
fi

# Verificar Node.js
if command -v node &> /dev/null; then
    echo -e "${GREEN}✅ Node.js instalado:${NC} $(node --version)"
else
    echo -e "${RED}❌ Node.js no está instalado${NC}"
    echo "   Descarga desde: https://nodejs.org/"
    exit 1
fi

# Verificar Docker
if command -v docker &> /dev/null; then
    echo -e "${GREEN}✅ Docker instalado:${NC} $(docker --version)"
else
    echo -e "${YELLOW}⚠️  Docker no está instalado (opcional)${NC}"
    echo "   Puedes trabajar sin Docker, pero necesitarás:"
    echo "   - PostgreSQL local"
    echo "   - Redis local"
fi

echo ""

# ---------- 2. CREAR ARCHIVO .env ----------
echo "🔧 Creando archivo .env..."
echo ""

if [ -f .env ]; then
    echo -e "${YELLOW}⚠️  .env ya existe, no se modifica${NC}"
else
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${GREEN}✅ .env creado desde .env.example${NC}"
        echo ""
        echo -e "${YELLOW}⚠️  IMPORTANTE: Edita .env y agrega tus API keys${NC}"
        echo "   nano .env"
    else
        echo -e "${RED}❌ No se encontró .env.example${NC}"
        echo "   Crea el archivo .env.example primero"
    fi
fi

echo ""

# ---------- 3. CREAR CARPETAS ----------
echo "📁 Creando carpetas necesarias..."
echo ""

if [ ! -d logs ]; then
    mkdir -p logs
    echo -e "${GREEN}✅ logs/ creada${NC}"
fi

if [ ! -d data ]; then
    mkdir -p data
    echo -e "${GREEN}✅ data/ creada${NC}"
fi

if [ ! -d ml/models ]; then
    mkdir -p ml/models
    echo -e "${GREEN}✅ ml/models/ creada${NC}"
fi

echo ""

# ---------- 4. DESCARGAR DEPENDENCIAS ----------
echo "📦 Descargando dependencias..."
echo ""

if [ -d services/api-go ]; then
    echo "📦 Go dependencies..."
    cd services/api-go
    go mod download 2>/dev/null
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Dependencias de Go listas${NC}"
    else
        echo -e "${YELLOW}⚠️  No se pudieron descargar dependencias de Go${NC}"
        echo "   (Puede que go.mod no exista aún)"
    fi
    cd ../..
else
    echo -e "${YELLOW}⚠️  services/api-go no existe aún${NC}"
fi

echo ""

if [ -d services/collector-rust ]; then
    echo "📦 Rust dependencies..."
    cd services/collector-rust
    cargo fetch 2>/dev/null
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Dependencias de Rust listas${NC}"
    else
        echo -e "${YELLOW}⚠️  No se pudieron descargar dependencias de Rust${NC}"
        echo "   (Puede que Cargo.toml no exista aún)"
    fi
    cd ../..
else
    echo -e "${YELLOW}⚠️  services/collector-rust no existe aún${NC}"
fi

echo ""

if [ -d frontend ]; then
    echo "📦 Node dependencies..."
    cd frontend
    npm install 2>/dev/null
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Dependencias de Node listas${NC}"
    else
        echo -e "${YELLOW}⚠️  No se pudieron descargar dependencias de Node${NC}"
        echo "   (Puede que package.json no exista aún)"
    fi
    cd ..
else
    echo -e "${YELLOW}⚠️  frontend no existe aún${NC}"
fi

echo ""

# ---------- 5. INSTRUCCIONES FINALES ----------
echo "=========================================="
echo "  ✅ Setup completado!"
echo "=========================================="
echo ""
echo "Siguientes pasos:"
echo ""
echo "1. Edita tu archivo .env:"
echo "   nano .env"
echo ""
echo "2. Inicia los servicios:"
echo "   docker-compose up -d postgres redis"
echo ""
echo "3. Inicia la API:"
echo "   make dev-api"
echo ""
echo "4. Inicia el frontend:"
echo "   make dev-frontend"
echo ""
echo "5. Inicia el collector:"
echo "   make dev-rust"
echo ""
echo "Para ver todos los comandos:"
echo "   make help"
echo ""