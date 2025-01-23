#!/bin/bash

cd "$(dirname "$0")/../smart-contracts"

# Request Transfer
echo "💸 Requesting transfer in ETH from Ethereum to Subtensor..."
npx hardhat functions:requestTransfer --network localhost --to 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --amount 0.1 --destination 0
echo "✅ Ethereum -> Subtensor ETH transfer requested"

echo "💸 Requesting transfer in TAO from Subtensor to Ethereum..."
npx hardhat run scripts/requestTransfer.ts --network other_after_setup
echo "✅ Subtensor -> Ethereum TAO transfer requested"
