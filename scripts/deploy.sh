#!/bin/bash

# ============================================
# CloudPulse AI deployment
# ============================================

# Colores
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo ""
echo "=========================================="
echo "  🚀 Desplegando CloudPulse AI"
echo "=========================================="
echo ""

# ---------- 1. VERIFICAR HERRAMIENTAS ----------
echo "📋 Checking deployment tools..."
echo ""

if command -v vercel &> /dev/null; then
    echo -e "${GREEN}✅ Vercel CLI instalado${NC}"
else
    echo -e "${YELLOW}⚠️  Vercel CLI no instalado${NC}"
    echo "   Install with: npm install -g vercel"
    echo ""
    echo "   ¿Quieres instalarlo ahora? (s/n)"
    read -r respuesta
    if [ "$respuesta" = "s" ]; then
        npm install -g vercel
    fi
fi

echo ""

if command -v railway &> /dev/null; then
    echo -e "${GREEN}✅ Railway CLI instalado${NC}"
else
    echo -e "${YELLOW}⚠️  Railway CLI no instalado${NC}"
    echo "   Install with: npm install -g @railway/cli"
    echo ""
    echo "   ¿Quieres instalarlo ahora? (s/n)"
    read -r respuesta
    if [ "$respuesta" = "s" ]; then
        npm install -g @railway/cli
    fi
fi

echo ""

if command -v flyctl &> /dev/null; then
    echo -e "${GREEN}✅ Flyctl instalado${NC}"
else
    echo -e "${YELLOW}⚠️  Flyctl no instalado${NC}"
    echo "   Install with: curl -L https://fly.io/install.sh | sh"
fi

echo ""

# ---------- 2. EJECUTAR TESTS ----------
echo "🧪 Running tests before deployment..."
echo ""

echo "📦 Go tests..."
cd services/api-go
go test ./... 2>/dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Go tests passed${NC}"
else
    echo -e "${RED}❌ Go tests failed${NC}"
    echo "   Fix the errors before deployment"
    exit 1
fi
cd ../..

echo ""

echo "📦 Rust tests..."
cd services/collector-rust
cargo test 2>/dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Rust tests passed${NC}"
else
    echo -e "${RED}❌ Rust tests failed${NC}"
    echo "   Fix the errors before deployment"
    exit 1
fi
cd ../..

echo ""

# ---------- 3. DESPLEGAR FRONTEND ----------
echo "📦 Desplegando frontend a Vercel..."
echo ""

cd frontend
if [ -f package.json ]; then
    vercel --prod
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Frontend desplegado en Vercel${NC}"
    else
        echo -e "${RED}❌ Error desplegando frontend${NC}"
        echo "   Verify that you are logged in to Vercel:"
        echo "   vercel login"
    fi
else
    echo -e "${YELLOW}⚠️  frontend/package.json no existe${NC}"
    echo "   Crea primero el frontend"
fi
cd ..

echo ""

# ---------- 4. DESPLEGAR API ----------
echo "📦 Desplegando API a Railway..."
echo ""

cd services/api-go
if [ -f go.mod ]; then
    railway up
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ API desplegada en Railway${NC}"
    else
        echo -e "${RED}❌ Error desplegando API${NC}"
        echo "   Verify that you are logged in to Railway:"
        echo "   railway login"
    fi
else
    echo -e "${YELLOW}⚠️  services/api-go/go.mod no existe${NC}"
    echo "   Crea primero la API"
fi
cd ../..

echo ""

# ---------- 5. DESPLEGAR COLLECTOR ----------
echo "📦 Desplegando collector a Fly.io..."
echo ""

cd services/collector-rust
if [ -f Cargo.toml ]; then
    flyctl deploy
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Collector desplegado en Fly.io${NC}"
    else
        echo -e "${RED}❌ Error desplegando collector${NC}"
        echo "   Verify that you are logged in to Fly.io:"
        echo "   flyctl auth login"
    fi
else
    echo -e "${YELLOW}⚠️  services/collector-rust/Cargo.toml no existe${NC}"
    echo "   Crea primero el collector"
fi
cd ../..

echo ""

# ---------- 6. RESULTADO FINAL ----------
echo "=========================================="
echo "  ✅ Despliegue completado!"
echo "=========================================="
echo ""
echo "Production URLs:"
echo "  Frontend: https://cloudpulse-ai.vercel.app"
echo "  API:      https://cloudpulse-api.railway.app"
echo "  Collector: https://cloudpulse-collector.fly.dev"
echo ""
echo "Next steps:"
echo "  1. Configure environment variables for each service"
echo "  2. Verify that everything works in production"
echo "  3. Configura tu dominio personalizado"
echo ""