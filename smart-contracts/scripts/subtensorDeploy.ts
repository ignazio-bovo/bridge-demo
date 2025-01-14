import { ethers, upgrades } from "hardhat";
import fs from "fs";
import {
  privateKeys,
  fundSubtensorAccount,
  subtensorExtraConfig,
} from "./fundSubtensorAccounts";

const SUBTENSOR_GAS_LIMIT = 60000000;

async function main(): Promise<void> {
  const accounts = await ethers.getSigners();
  const deployer = accounts[3];
  const admin = accounts[0];
  const authority = accounts[1];

  const deployerAddress = await deployer.getAddress();
  const adminAddress = await admin.getAddress();
  const authorityAddress = await authority.getAddress();

  await fundSubtensorAccount(deployerAddress, 1000000);
  await fundSubtensorAccount(adminAddress, 1000000);
  await fundSubtensorAccount(authorityAddress, 1000000);

  // Deploy upgradeable bridge
  console.log("👉 Deploying bridge...");
  const bridgeFactory = await ethers.getContractFactory("Bridge");
  const bridge = await bridgeFactory.connect(deployer).deploy({
    gasLimit: SUBTENSOR_GAS_LIMIT,
  });
  await bridge.initialize(authorityAddress, adminAddress, {
    gasLimit: SUBTENSOR_GAS_LIMIT,
  });

  const bridgeAddress = await bridge.getAddress();
  await bridge.deploymentTransaction()?.wait();
  console.log("🚧 Bridge deployed at 🎯", bridgeAddress);

  // Deploy StakeManager
  const daturaHotkey =
    "0x0000000000000000000000000000000000000000000000000000000000000000";
  const stakePrecompile = "0x0000000000000000000000000000000000000801";
  const rate = 38051750380;

  console.log("👉 Deploying StakeManager...");
  const stakeManagerFactory = await ethers.getContractFactory("StakeManager");
  const stakeManager = await stakeManagerFactory.connect(deployer).deploy({
    gasLimit: SUBTENSOR_GAS_LIMIT,
  });

  await stakeManager.initialize(
    bridgeAddress,
    rate,
    daturaHotkey,
    stakePrecompile,
    {
      gasLimit: SUBTENSOR_GAS_LIMIT,
    }
  );
  const stakeManagerAddress = await stakeManager.getAddress();
  console.log("🚀 StakeManager deployed to:", stakeManagerAddress);
}

main().catch((error: Error) => {
  console.error(error);
  process.exitCode = 1;
});
