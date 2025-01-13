#!/bin/bash

cd "$(dirname "$0")/../smart-contracts"

# Deploy bridge
echo "🌉 Deploying subtensor bridge..."
npx hardhat run scripts/subtensorDeploy.ts --network other
echo "✅ Datura bridge contract suite deployed on Subtensor"
