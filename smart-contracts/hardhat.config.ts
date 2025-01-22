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
    hardhat: {
      chainId: 31337,
    },
    localhost: {
      url: "http://127.0.0.1:8545",
      chainId: 31337,
    },
    other: {
      url: "http://127.0.0.1:9944",
      accounts: [
        "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
        "0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d",
        "0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a",
        "0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6",
      ],

      chainId: 31338,
    },
    sepolia: {
      url: `https://sepolia.infura.io/v3/${sepoliaApiKey}`,
      accounts: [privateKey!],
      chainId: 11155111,
    },
    subevmTestnet: {
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
    customChains: [
      {
        network: "subevmTestnet",
        chainId: 945,
        urls: {
          apiURL: "https://evm-testscan.dev.opentensor.ai/api",
          browserURL: "https://evm-testscan.dev.opentensor.ai",
        },
      },
    ],
  },
};

export default config;
