import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "@openzeppelin/hardhat-upgrades";
import "hardhat-gas-reporter";
import dotenv from "dotenv";
dotenv.config();

import "./task/deployment/deploy";
import "./task/deployment/health";
import "./task/functions/requestTransfer";
import "./task/functions/confirmTransferRequest";
import "./task/functions/confirmTransferRequestTao";
import "./task/deployment/deployTestnet";
import "./task/functions/whitelistToken";

const etherscanApiKey = process.env.ETHERSCAN_API_KEY;
const sepoliaApiKey = process.env.SEPOLIA_API_KEY;
const privateKey = process.env.PRIVATE_KEY;

const config: HardhatUserConfig = {
  solidity: {
    version: "0.8.22",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200,
      },
    },
  },
  gasReporter: {
    enabled: true,
    currency: "USD",
  },
  networks: {
    localhost: {
      url: "http://127.0.0.1:8545",
      chainId: 1,
    },
    other: {
      url: "http://127.0.0.1:9944",
      chainId: 945,
    },
    sepolia: {
      url: `https://sepolia.infura.io/v3/${sepoliaApiKey}`,
      accounts: [privateKey!],
      chainId: 11155111,
    },
    subevm: {
      url: "https://evm-testnet.dev.opentensor.ai",
      accounts: [privateKey!],
      chainId: 945,
    },
  },
  paths: {
    sources: "./contracts", // Where your contracts live
    tests: "./test", // Where your tests live
    cache: "./cache", // Where compiled contracts are cached
    artifacts: "./artifacts", // Where compiled contracts are saved
  },
  etherscan: {
    apiKey: etherscanApiKey!,
  },
};

export default config;
