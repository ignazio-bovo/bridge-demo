#!/bin/bash

cd "$(dirname "$0")/../smart-contracts"

# Whitelist Token
echo "🔑 Whitelisting token..."
npx hardhat functions:whitelistToken --network localhost --address 0x0000000000000000000000000000000000000000
echo "✅ Token whitelisted"
