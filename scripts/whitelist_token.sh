#!/bin/bash

cd "$(dirname "$0")/../smart-contracts"

# Whitelist Token
echo "🔑 Whitelisting token on Ethereum..."
npx hardhat functions:whitelistToken --network localhost --address 0x0000000000000000000000000000000000000000 --name "Ether" --symbol "ETH" --decimals 18
echo "✅ Token whitelisted on Ethereum"

echo "🔑 Whitelisting token on Subtensor..."
npx hardhat run scripts/whitelistToken.ts --network other_after_setup
echo "✅ Token whitelisted on Subtensor"
