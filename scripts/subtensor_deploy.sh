#!/bin/bash

cd "$(dirname "$0")/../smart-contracts"

# Deploy bridge
echo "🌉 Deploying subtensor bridge..."
ts-node scripts/subtensorDeploy.ts
echo "✅ Subtensor bridge deployed"
