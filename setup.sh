#!/bin/bash
set -e

echo "🚀 DevSmith Modular Platform - One-Command Setup"
echo "=================================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Step 1: Check prerequisites
echo "📋 Step 1/8: Checking prerequisites..."
bash scripts/check-prerequisites.sh
echo -e "${GREEN}✓ Prerequisites verified${NC}\n"

# Step 2: Create databases
echo "🗄️  Step 2/8: Creating databases and schemas..."
bash scripts/create-databases.sh
echo -e "${GREEN}✓ Databases created${NC}\n"

# Step 3: Run migrations
echo "📊 Step 3/8: Running database migrations..."
bash scripts/run-migrations.sh
echo -e "${GREEN}✓ Migrations completed${NC}\n"

# Step 4: Setup environment files
echo "⚙️  Step 4/8: Setting up environment configuration..."
if [ ! -f .env ]; then
  cp .env.example .env
  echo -e "${YELLOW}⚠️  Please edit .env with your GitHub OAuth credentials${NC}"
  echo "   Then run ./setup.sh again"
  exit 1
fi
echo -e "${GREEN}✓ Environment configured${NC}\n"

# Step 5: Build services
echo "🔨 Step 5/8: Building all service binaries..."
bash scripts/build-services.sh
echo -e "${GREEN}✓ Services built${NC}\n"

# Step 6: Setup Ollama and select model based on RAM
echo "🤖 Step 6/8: Setting up Ollama and AI model..."
if ! pgrep -x "ollama" > /dev/null 2>&1; then
  echo "Starting Ollama..."
  ollama serve > /dev/null 2>&1 &
  sleep 3
fi

# Detect RAM and recommend model
TOTAL_RAM=$(free -g 2>/dev/null | awk '/^Mem:/{print $2}' || echo "0")

if [ -z "$TOTAL_RAM" ] || [ "$TOTAL_RAM" -lt 8 ]; then
  echo -e "${RED}⚠️  Unable to detect RAM or less than 8GB${NC}"
  echo "   Recommend: deepseek-coder:1.5b (minimal)"
  DEFAULT_MODEL="deepseek-coder:1.5b"
elif [ "$TOTAL_RAM" -lt 24 ]; then
  echo "✓ ${TOTAL_RAM}GB RAM detected"
  echo "   Recommend: deepseek-coder:6.7b (good balance)"
  DEFAULT_MODEL="deepseek-coder:6.7b"
else
  echo "✓ ${TOTAL_RAM}GB RAM detected"
  echo "   Recommend: deepseek-coder-v2:16b (best quality)"
  DEFAULT_MODEL="deepseek-coder-v2:16b"
fi

echo ""
echo "Available models:"
echo "  1) deepseek-coder:1.5b (8GB RAM, ~1GB download, fastest)"
echo "  2) deepseek-coder:6.7b (16GB RAM, ~4GB download, recommended)"
echo "  3) deepseek-coder-v2:16b (32GB RAM, ~9GB download, best quality)"
echo "  4) qwen2.5-coder:7b (16GB RAM, ~4GB download, alternative)"
echo ""

read -p "Select model [2]: " MODEL_CHOICE
MODEL_CHOICE=${MODEL_CHOICE:-2}

case $MODEL_CHOICE in
  1) CHOSEN_MODEL="deepseek-coder:1.5b" ;;
  2) CHOSEN_MODEL="deepseek-coder:6.7b" ;;
  3) CHOSEN_MODEL="deepseek-coder-v2:16b" ;;
  4) CHOSEN_MODEL="qwen2.5-coder:7b" ;;
  *) CHOSEN_MODEL=$DEFAULT_MODEL ;;
esac

echo "Selected model: $CHOSEN_MODEL"

# Pull model if not already present
if ! ollama list 2>/dev/null | grep -q "$CHOSEN_MODEL"; then
  echo "Pulling $CHOSEN_MODEL (this may take 5-15 minutes depending on model size)..."
  ollama pull "$CHOSEN_MODEL"
else
  echo "✓ Model $CHOSEN_MODEL already downloaded"
fi

# Update .env with chosen model
if [ -f .env ]; then
  if grep -q "OLLAMA_MODEL=" .env; then
    sed -i.bak "s|OLLAMA_MODEL=.*|OLLAMA_MODEL=$CHOSEN_MODEL|" .env
  else
    echo "OLLAMA_MODEL=$CHOSEN_MODEL" >> .env
  fi
  echo "✓ Updated .env with OLLAMA_MODEL=$CHOSEN_MODEL"
fi

echo -e "${GREEN}✓ Ollama ready${NC}\n"

# Step 7: Start services
echo "🚀 Step 7/8: Starting all services..."
bash scripts/start-services.sh
echo -e "${GREEN}✓ Services started${NC}\n"

# Step 8: Health checks
echo "🏥 Step 8/8: Running health checks..."
sleep 5  # Give services time to start
bash scripts/health-checks.sh
echo -e "${GREEN}✓ All services healthy${NC}\n"

echo "=================================================="
echo -e "${GREEN}✅ Setup complete!${NC}"
echo ""
echo "🌐 Platform URLs:"
echo "   Portal:    http://localhost:8080"
echo "   Review:    http://localhost:8081"
echo "   Logs:      http://localhost:8082"
echo "   Analytics: http://localhost:8083"
echo ""
echo "📝 Next steps:"
echo "   1. Open http://localhost:8080"
echo "   2. Log in with GitHub OAuth"
echo "   3. Start reviewing code!"
echo ""
echo "🛑 To stop all services: ./teardown.sh"
echo "🔍 To verify setup: ./verify-setup.sh"
