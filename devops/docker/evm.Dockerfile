# This Dockerfile is used to build the EVM 
FROM ubuntu:20.04

SHELL ["/bin/bash", "-c"]

RUN apt update

RUN apt install -y curl git

# install foundry
RUN curl -L https://foundry.paradigm.xyz | bash

RUN /root/.foundry/bin/foundryup


# recreate the foundry directory structure inside the container
RUN mkdir /root/lib
RUN mkdir /root/script
RUN mkdir /root/src

WORKDIR /root
# Expose the default Anvil port
EXPOSE 8545

# Add environment variables with defaults
ENV ANVIL_PORT=8545
ENV ANVIL_CHAIN_ID=31337

# Update CMD to use environment variables
CMD ["/bin/bash", "-c", "/root/.foundry/bin/anvil --port $ANVIL_PORT --chain-id $ANVIL_CHAIN_ID --host 0.0.0.0", "--verbosity 5"]

# FROM node:18-slim

# WORKDIR /app

# RUN apt-get update && \
#     apt-get install -y python3 make g++ git && \
#     rm -rf /var/lib/apt/lists/*

# # Copy package files first to leverage Docker cache
# COPY smart-contracts/package*.json ./
# RUN npm install

# # Copy the rest of the files
# COPY smart-contracts ./

# EXPOSE 8545

# # Use host.docker.internal for the JSON-RPC server
# CMD ["npx", "hardhat", "node", "--hostname", "host.docker.internal"]