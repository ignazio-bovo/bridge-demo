#!/bin/bash

cd "$(dirname "$0")/../smart-contracts"

# Deploy bridge
echo "🌉 Deploying EVM bridge..."
npx hardhat deploy:bridge --network localhost --v 1
echo "✅ Datura bridge contract suite deployed on EVM"
