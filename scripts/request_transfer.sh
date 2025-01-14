#!/bin/bash

cd "$(dirname "$0")/../smart-contracts"

# Whitelist Token
npx hardhat functions:whitelistToken --network localhost --address 0x0000000000000000000000000000000000000000
echo "✅ Token whitelisted"
sleep 3

# Request Transfer
npx hardhat functions:requestTransfer --network localhost --to 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --amount 0.1 --destination 945 --native
echo "✅ 1 -> 945 Native transfer requested"
