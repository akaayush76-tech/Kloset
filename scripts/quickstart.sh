#!/bin/bash

# Quick Start Script for Fittingly Go Backend with Docker
# This script helps users get started with the application quickly

set -e

echo "================================================"
echo "Fittingly Go Backend - Docker Quick Start"
echo "================================================"
echo ""

# Check Docker installation
echo "✓ Checking Docker installation..."
if ! command -v docker &> /dev/null; then
    echo "✗ Docker is not installed. Please install Docker from https://www.docker.com/products/docker-desktop"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "✗ Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi

echo "✓ Docker $(docker --version | grep -oP 'version \K[^,]*')"
echo "✓ Docker Compose $(docker-compose --version | grep -oP '\d+\.\d+\.\d+')"
echo ""

# Check if .env file exists
echo "✓ Checking environment configuration..."
if [ ! -f .env ]; then
    if [ -f .env.development ]; then
        echo "  Creating .env from .env.development..."
        cp .env.development .env
        echo "  ✓ .env created"
    else
        echo "  ✗ No .env file found. Please create one based on .env.docker"
        exit 1
    fi
else
    echo "  ✓ .env file exists"
fi
echo ""

# Check ports
echo "✓ Checking port availability..."
PORTS=(8080 27017 8081)
for port in "${PORTS[@]}"; do
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "  ⚠ Port $port is in use. This may cause issues."
    else
        echo "  ✓ Port $port is available"
    fi
done
echo ""

# Start services
echo "================================================"
echo "Starting Docker containers..."
echo "================================================"
echo ""

# Build images
echo "📦 Building Docker images..."
docker-compose build --quiet
echo "✓ Docker images built"
echo ""

# Start services
echo "🚀 Starting services..."
docker-compose up -d
echo "✓ Services starting..."
echo ""

# Wait for services to be healthy
echo "⏳ Waiting for services to be ready..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if docker-compose ps api | grep -q "healthy"; then
        echo "✓ API is ready"
        break
    fi
    echo -n "."
    sleep 1
    ((attempt++))
done

if [ $attempt -eq $max_attempts ]; then
    echo ""
    echo "✗ API failed to start. Checking logs..."
    docker-compose logs api | tail -20
    exit 1
fi
echo ""

# Display service information
echo "================================================"
echo "✓ All services are running!"
echo "================================================"
echo ""
echo "📍 Service URLs:"
echo "   API Server:       http://localhost:8080"
echo "   MongoDB Express:  http://localhost:8081"
echo ""
echo "📊 MongoDB Connection:"
echo "   URI: mongodb://root:rootpassword@localhost:27017/"
echo "   Database: kloset_dev"
echo ""
echo "🧪 Quick Test:"
echo "   curl http://localhost:8080/api/products"
echo ""
echo "📖 Documentation:"
echo "   See DOCKER_SETUP.md for detailed instructions"
echo ""
echo "💡 Useful Commands:"
echo "   make help           - Show all available commands"
echo "   make docker-logs    - View service logs"
echo "   make docker-down    - Stop services"
echo "   make docker-clean   - Remove containers and volumes"
echo ""
echo "Happy coding! 🚀"
