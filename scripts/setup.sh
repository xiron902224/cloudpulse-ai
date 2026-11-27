#!/bin/bash

# ============================================
# Initial setup for CloudPulse AI
# ============================================

# Message colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No color

echo ""
echo "=========================================="
echo "  🚀 CloudPulse AI setup"
echo "=========================================="
echo ""

# ---------- 1. CHECK TOOLS ----------
echo "📋 Checking installed tools..."
echo ""

# Verificar Go
if command -v go &> /dev/null; then
    echo -e "${GREEN}✅ Go instalado:${NC} $(go version)"
else
    echo -e "${RED}❌ Go is not installed${NC}"
    echo "   Download from: https://golang.org/dl/"
    exit 1
fi

# Verificar Rust
if command -v cargo &> /dev/null; then
    echo -e "${GREEN}✅ Rust instalado:${NC} $(rustc --version)"
else
    echo -e "${RED}❌ Rust is not installed${NC}"
    echo "   Install with: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
    exit 1
fi

# Verificar Node.js
if command -v node &> /dev/null; then
    echo -e "${GREEN}✅ Node.js instalado:${NC} $(node --version)"
else
    echo -e "${RED}❌ Node.js is not installed${NC}"
    echo "   Download from: https://nodejs.org/"
    exit 1
fi

# Verificar Docker
if command -v docker &> /dev/null; then
    echo -e "${GREEN}✅ Docker instalado:${NC} $(docker --version)"
else
    echo -e "${YELLOW}⚠️  Docker is not installed (optional)${NC}"
    echo "   You can work without Docker, but you will need:"
    echo "   - PostgreSQL local"
    echo "   - Redis local"
fi

echo ""

# ---------- 2. CREATE .env FILE ----------
echo "🔧 Creating .env file..."
echo ""

if [ -f .env ]; then
    echo -e "${YELLOW}⚠️  .env already exists, leaving it unchanged${NC}"
else
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${GREEN}✅ .env creado desde .env.example${NC}"
        echo ""
        echo -e "${YELLOW}⚠️  IMPORTANT: Edit .env and add your API keys${NC}"
        echo "   nano .env"
    else
        echo -e "${RED}❌ .env.example was not found${NC}"
        echo "   Create the .env.example file first"
    fi
fi

echo ""

# ---------- 3. CREATE DIRECTORIES ----------
echo "📁 Creating required directories..."
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

# ---------- 4. DOWNLOAD DEPENDENCIES ----------
echo "📦 Downloading dependencies..."
echo ""

if [ -d services/api-go ]; then
    echo "📦 Go dependencies..."
    cd services/api-go
    go mod download 2>/dev/null
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Go dependencies ready${NC}"
    else
        echo -e "${YELLOW}⚠️  Could not download Go dependencies${NC}"
        echo "   (go.mod may not exist yet)"
    fi
    cd ../..
else
    echo -e "${YELLOW}⚠️  services/api-go does not exist yet${NC}"
fi

echo ""

if [ -d services/collector-rust ]; then
    echo "📦 Rust dependencies..."
    cd services/collector-rust
    cargo fetch 2>/dev/null
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Rust dependencies ready${NC}"
    else
        echo -e "${YELLOW}⚠️  Could not download Rust dependencies${NC}"
        echo "   (Cargo.toml may not exist yet)"
    fi
    cd ../..
else
    echo -e "${YELLOW}⚠️  services/collector-rust does not exist yet${NC}"
fi

echo ""

if [ -d frontend ]; then
    echo "📦 Node dependencies..."
    cd frontend
    npm install 2>/dev/null
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Node dependencies ready${NC}"
    else
        echo -e "${YELLOW}⚠️  Could not download Node dependencies${NC}"
        echo "   (package.json may not exist yet)"
    fi
    cd ..
else
    echo -e "${YELLOW}⚠️  frontend does not exist yet${NC}"
fi

echo ""

# ---------- 5. FINAL INSTRUCTIONS ----------
echo "=========================================="
echo "  ✅ Setup completado!"
echo "=========================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Edit your .env file:"
echo "   nano .env"
echo ""
echo "2. Start the services:"
echo "   docker-compose up -d postgres redis"
echo ""
echo "3. Start the API:"
echo "   make dev-api"
echo ""
echo "4. Start the frontend:"
echo "   make dev-frontend"
echo ""
echo "5. Start the collector:"
echo "   make dev-rust"
echo ""
echo "To see all commands:"
echo "   make help"
echo ""