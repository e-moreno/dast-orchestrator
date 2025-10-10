#!/bin/bash

# 🚀 Build AMD64 Docker Image on DigitalOcean Droplet
# Run this script to automatically create, build, and cleanup

set -e

echo "🌊 Building AMD64 image on DigitalOcean..."

# 1. Create temporary AMD64 droplet
echo "📦 Creating build droplet..."
DROPLET_ID=$(doctl compute droplet create dast-build-$(date +%s) \
  --image ubuntu-22-04-x64 \
  --size s-2vcpu-2gb \
  --region nyc3 \
  --wait \
  --format ID \
  --no-header)

echo "✅ Droplet created: $DROPLET_ID"

# Get droplet IP
DROPLET_IP=$(doctl compute droplet get $DROPLET_ID --format PublicIPv4 --no-header)
echo "🌐 Droplet IP: $DROPLET_IP"

# Wait for SSH to be ready
echo "⏳ Waiting for SSH to be ready..."
until ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no root@$DROPLET_IP "echo 'SSH Ready'" 2>/dev/null; do
    echo "   Still waiting..."
    sleep 5
done

echo "🔧 Installing Docker and dependencies..."
ssh -o StrictHostKeyChecking=no root@$DROPLET_IP << 'EOF'
# Install Docker
apt update -qq
apt install -y docker.io git curl
systemctl start docker
systemctl enable docker

# Clone repository (replace with your repo URL)
git clone https://github.com/YOUR_USERNAME/dast-orchestrator.git
cd dast-orchestrator/api

# Build AMD64 image
echo "🏗️ Building AMD64 image..."
docker build -t tommoreno/dast-api:1.0.3-amd64 .

# Login and push
echo "📤 Pushing to Docker Hub..."
docker login
docker push tommoreno/dast-api:1.0.3-amd64

echo "✅ Build complete!"
EOF

echo "🗑️ Cleaning up droplet..."
doctl compute droplet delete $DROPLET_ID --force

echo "🎉 AMD64 image built and pushed: tommoreno/dast-api:1.0.3-amd64"
echo "📝 Update your deployment to use: tommoreno/dast-api:1.0.3-amd64"
