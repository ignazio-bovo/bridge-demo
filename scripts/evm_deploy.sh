#!/bin/bash

cd "$(dirname "$0")/../smart-contracts"

# Deploy bridge
echo "🌉 Deploying EVM bridge..."
npx hardhat deploy:bridge --network localhost --v 1
echo "✅ Datura bridge contract suite deployed on EVM"

# # Whitelist Token
# npx hardhat functions:whitelistToken --network localhost --address 0x0000000000000000000000000000000000000000
# echo "⏳ Token whitelisted"
# sleep 3

# # Request Transfer
# npx hardhat functions:requestTransfer --network localhost --to 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --amount 0.1 --destination 945 --native
# echo "⏳ Waiting for transfer 1 -> 945 Native to be confirmed..."
# sleep 3
