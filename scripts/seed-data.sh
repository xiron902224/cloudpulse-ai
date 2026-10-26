#!/bin/bash

# ============================================
# Test data for CloudPulse AI
# ============================================

# Colores
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo ""
echo "=========================================="
echo "  🌱 Loading test data"
echo "=========================================="
echo ""

# ---------- CONFIGURATION ----------
API_URL="http://localhost:8080"
COLLECTOR_URL="http://localhost:9090"

# ---------- VERIFICAR SERVICIOS ----------
echo "📋 Verificando servicios..."
echo ""

if curl -s "$API_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ API corriendo en $API_URL${NC}"
else
    echo -e "${RED}❌ API is not running at $API_URL${NC}"
    echo "   Start with: make dev-api"
    echo "   O: docker-compose up -d api"
    exit 1
fi

echo ""

# ---------- INSERT METRICS ----------
echo "📊 Generating test metrics..."
echo ""

for i in $(seq 1 50); do
    CPU_USAGE=$(shuf -i 10-90 -n 1)
    MEMORY_USAGE=$(shuf -i 20-80 -n 1)
    DISK_USAGE=$(shuf -i 30-70 -n 1)
    NETWORK_IN=$(shuf -i 100-900 -n 1)
    NETWORK_OUT=$(shuf -i 50-500 -n 1)
    
    TIMESTAMP=$(date -d "-$i minutes" +%s 2>/dev/null || date -v-"$i"M +%s)
    
    RESPONSE=$(curl -s -X POST "$API_URL/api/v1/metrics" \
        -H "Content-Type: application/json" \
        -d "{
            \"cpu_usage\": $CPU_USAGE,
            \"memory_usage\": $MEMORY_USAGE,
            \"disk_usage\": $DISK_USAGE,
            \"network_in\": $NETWORK_IN,
            \"network_out\": $NETWORK_OUT,
            \"timestamp\": $TIMESTAMP
        }" 2>/dev/null)
    
    if [ $((i % 10)) -eq 0 ]; then
        echo "   ✅ $i/50 metrics sent"
    fi
    
    sleep 0.1
done

echo ""
echo -e "${GREEN}✅ 50 metrics generated${NC}"
echo ""

# ---------- INSERT ALERTS ----------
echo "🚨 Generating test alerts..."
echo ""

ALERTS=(
    '{"severity":"critical","message":"CPU usage above 90% on server-1","status":"active"}'
    '{"severity":"warning","message":"Memory usage above 80% on server-2","status":"active"}'
    '{"severity":"info","message":"Disk usage above 70% on server-3","status":"resolved"}'
    '{"severity":"critical","message":"Network latency above 500ms","status":"active"}'
    '{"severity":"warning","message":"High number of connections","status":"resolved"}'
)

for alert in "${ALERTS[@]}"; do
    curl -s -X POST "$API_URL/api/v1/alerts" \
        -H "Content-Type: application/json" \
        -d "$alert" > /dev/null 2>&1
    sleep 0.1
done

echo -e "${GREEN}✅ 5 alerts generated${NC}"
echo ""

# ---------- FINAL RESULT ----------
echo "=========================================="
echo "  ✅ Test data loaded!"
echo "=========================================="
echo ""
echo "You can now:"
echo "  1. Abrir el frontend en http://localhost:3000"
echo "  2. View charts with data"
echo "  3. Test alerts"
echo ""
echo "To view the data:"
echo "  curl http://localhost:8080/api/v1/metrics"
echo ""