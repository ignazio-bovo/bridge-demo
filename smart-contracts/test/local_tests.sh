#!/bin/bash

# Array to store PIDs
declare -a PIDS=()

# Function to cleanup and exit
cleanup() {
    echo "Cleaning up..."
    for pid in "${PIDS[@]}"; do
        kill $pid 2>/dev/null
    done
    exit
}

# Set up trap for cleanup
trap cleanup EXIT INT TERM

# Start Hardhat node in the background
start_hardhat_node() {
    local port=$1
    npx hardhat node --port $port &
    PIDS+=($!)
    sleep 2
}

# anvil
start_node() {
    local port=$1
    local chain_id=$2
    anvil \
        --accounts 11 \
        --balance 10000 \
        --port $port \
        --block-time 3 \
        --silent \
        --chain-id $chain_id &
    # --silent \
    PIDS+=($!)
}

start_node 8545 1
start_node 9944 945

# Deploy bridge
npx hardhat deploy:bridge --network localhost --v 1
npx hardhat deploy:bridge --network other --v 1
echo "⏳ Waiting for bridge to be deployed..."
sleep 5

# Whitelist Token
npx hardhat functions:whitelistToken --network localhost --address 0x0000000000000000000000000000000000000000
echo "⏳ Token whitelisted"
sleep 3

# Request Transfer
npx hardhat functions:requestTransfer --network localhost --to 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --amount 0.1 --destination 945 --native
echo "⏳ Waiting for transfer 1 -> 945 Native to be confirmed..."
sleep 3

# # Confirm Transfer Request
# npx hardhat functions:requestTransfer --network other --to 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --amount 0.1 --destination 31337 --native
# echo "⏳ Waiting for transfer 31338 -> 31337 ERC20 to be confirmed..."
# sleep 10

# # Request Transfer
# npx hardhat functions:requestTransfer --network other --to 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --amount 0.1 --destination 31337 --native
# echo "⏳ Waiting for transfer 31338 -> 31337 Native to be confirmed..."
# sleep 10

# # Confirm Transfer Request
# npx hardhat functions:requestTransfer --network localhost --to 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --amount 0.1 --destination 31338 --native
# echo "⏳ Waiting for transfer 31337 -> 31338 ERC20 to be confirmed..."
# sleep 10

# Wait for a moment to ensure all tasks are completed
# Confirm Transfer Request
npx hardhat functions:confirmTransferRequest --network localhost --sourcechainid 1 --to 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --amount 0.1
echo "⏳ transfer 1 -> 1 Native confirmed (for testing purposes)"
sleep 2

# # Confirm Transfer Request
npx hardhat functions:confirmTransferRequestTao --network localhost --sourcechainid 945 --to 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --amount 0.1
echo "⏳ transfer 1 -> 1 TAO confirmed (for testing purposes)"
sleep 2

sleep 90000

# Cleanup will be called automatically due to the trap
