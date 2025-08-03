#!/bin/bash
# Tilt Development Environment Setup Script
# Sets up the todo app development environment with all required dependencies

set -e

echo "🚀 Setting up Todo App Development Environment"
echo "============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if running on macOS or Linux
if [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
else
    echo -e "${RED}❌ Unsupported operating system: $OSTYPE${NC}"
    exit 1
fi

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install via Homebrew (macOS)
install_with_brew() {
    if ! command_exists brew; then
        echo -e "${YELLOW}⚠️  Homebrew not found. Installing...${NC}"
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
    brew install "$1"
}

# Function to check and install dependencies
check_and_install() {
    local cmd=$1
    local package=$2
    local description=$3
    
    if command_exists "$cmd"; then
        echo -e "${GREEN}✅ $description is already installed${NC}"
    else
        echo -e "${YELLOW}📦 Installing $description...${NC}"
        if [[ "$OS" == "macos" ]]; then
            install_with_brew "$package"
        else
            echo -e "${RED}❌ Please install $description manually on Linux${NC}"
            exit 1
        fi
    fi
}

echo -e "${BLUE}📋 Checking required dependencies...${NC}"

# Check Docker
if command_exists docker; then
    if docker info >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Docker is running${NC}"
    else
        echo -e "${RED}❌ Docker is installed but not running. Please start Docker Desktop${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ Docker not found. Please install Docker Desktop${NC}"
    exit 1
fi

# Check Kubernetes context
check_and_install kubectl kubernetes-cli "kubectl"
check_and_install minikube minikube "Minikube"
check_and_install helm helm "Helm"
check_and_install buf buf "Buf CLI"
check_and_install tilt tilt "Tilt"

# Check if Bun is installed
if ! command_exists bun; then
    echo -e "${YELLOW}📦 Installing Bun...${NC}"
    curl -fsSL https://bun.sh/install | bash
    export PATH="$HOME/.bun/bin:$PATH"
fi

# Check Go installation
if ! command_exists go; then
    echo -e "${YELLOW}📦 Installing Go...${NC}"
    if [[ "$OS" == "macos" ]]; then
        install_with_brew go
    else
        echo -e "${RED}❌ Please install Go manually on Linux${NC}"
        exit 1
    fi
fi

echo -e "${BLUE}🔧 Setting up Kubernetes environment...${NC}"

# Start Minikube if not running
if ! minikube status >/dev/null 2>&1; then
    echo -e "${YELLOW}🚀 Starting Minikube...${NC}"
    minikube start --memory=4096 --cpus=2 --disk-size=20gb
fi

# Enable required addons
echo -e "${YELLOW}🔌 Enabling Minikube addons...${NC}"
minikube addons enable ingress
minikube addons enable registry

echo -e "${BLUE}🗄️ Setting up local Docker registry...${NC}"

# Setup local registry for faster image pulls
if ! docker ps --filter "name=registry" --format "{{.Names}}" | grep -q registry; then
    echo -e "${YELLOW}🐳 Starting local Docker registry...${NC}"
    docker run -d --restart=always -p 5000:5000 --name registry registry:2
fi

echo -e "${BLUE}🌐 Setting up /etc/hosts entries...${NC}"

# Add hosts entries
HOSTS_ENTRY="127.0.0.1 todo.local api.todo.local"
if ! grep -q "todo.local" /etc/hosts; then
    echo -e "${YELLOW}📝 Adding hosts entries (requires sudo)...${NC}"
    echo "$HOSTS_ENTRY" | sudo tee -a /etc/hosts
    echo -e "${GREEN}✅ Added hosts entries${NC}"
else
    echo -e "${GREEN}✅ Hosts entries already exist${NC}"
fi

echo -e "${BLUE}📦 Installing project dependencies...${NC}"

# Install backend dependencies
echo -e "${YELLOW}🔧 Installing Go dependencies...${NC}"
cd backend
go mod download
cd ..

# Install frontend dependencies
echo -e "${YELLOW}🔧 Installing frontend dependencies...${NC}"
cd frontend
bun install
cd ..

echo -e "${BLUE}📋 Running initial setup tasks...${NC}"

# Generate protobuf code
echo -e "${YELLOW}🔧 Generating protocol buffer code...${NC}"
buf generate

# Create backup directory
mkdir -p backup

echo -e "${GREEN}🎉 Setup Complete!${NC}"
echo ""
echo -e "${BLUE}Quick Start Commands:${NC}"
echo -e "  ${YELLOW}tilt up${NC}                    # Start all services"
echo -e "  ${YELLOW}tilt down${NC}                  # Stop all services"
echo -e "  ${YELLOW}tilt trigger test-all${NC}      # Run all tests"
echo -e "  ${YELLOW}tilt trigger health-check${NC}  # Check service health"
echo ""
echo -e "${BLUE}Development URLs:${NC}"
echo -e "  ${YELLOW}http://todo.local${NC}          # Frontend application"
echo -e "  ${YELLOW}http://api.todo.local${NC}      # Backend API"
echo -e "  ${YELLOW}http://localhost:3000${NC}      # Frontend (direct)"
echo -e "  ${YELLOW}http://localhost:8080${NC}      # Backend API (direct)"
echo ""
echo -e "${GREEN}Ready to start development! Run 'tilt up' to begin.${NC}"