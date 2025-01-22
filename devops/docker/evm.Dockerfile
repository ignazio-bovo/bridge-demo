FROM node:18-slim

WORKDIR /app

# Install basic dependencies
RUN apt-get update && \
    apt-get install -y python3 make g++ && \
    rm -rf /var/lib/apt/lists/*

COPY ../../smart-contracts ./

RUN npm install
EXPOSE 8545

CMD ["npx", "hardhat", "node", "--network", "hardhat"]