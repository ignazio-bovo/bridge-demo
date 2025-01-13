import { ethers, upgrades } from "hardhat";
import fs from "fs";
import {
  privateKeys,
  fundSubtensorAccount,
  subtensorExtraConfig,
} from "./fundSubtensorAccounts";

interface DeployedContract {
  address: string;
  abi: any;
}

async function main(): Promise<void> {
  const provider = new ethers.JsonRpcProvider(subtensorExtraConfig.httpUrl);
  const accounts = await ethers.getSigners();
  const deployer = accounts[10];
  const admin = accounts[0];
  const authority = accounts[1];

  await fundSubtensorAccount(deployer.address, 10000);
  await fundSubtensorAccount(admin.address, 10000);
  await fundSubtensorAccount(authority.address, 10000);

  let nativeBalance = await provider.getBalance(authority.address);
  console.log("🪙 Authority native balance:", nativeBalance.toString(), "wei");

  // Deploy token
  const tokenFactory = await ethers.getContractFactory("BridgedToken");
  const token = await tokenFactory
    .connect(deployer)
    .deploy("Ported ETH", "dETH", admin.address);

  const tokenAddress = await token.getAddress();
  console.log("🪙 Token deployed at 🎯", tokenAddress);

  // Deploy upgradeable bridge
  const bridgeFactory = await ethers.getContractFactory("Bridge");
  const bridge = await upgrades.deployProxy(
    bridgeFactory.connect(authority),
    [authority.address, admin.address],
    {
      initializer: "initialize",
      timeout: 60000,
      kind: "uups",
    }
  );

  const bridgeAddress = await bridge.getAddress();
  await bridge.deploymentTransaction()?.wait();
  console.log("🚧 Bridge deployed at 🎯", bridgeAddress);

  // Deploy StakeManager
  const daturaHotkey = "0x5GP7c3fFazW9GXK8Up3qgu2DJBk8inu4aK9TZy3RuoSWVCMi";
  const stakePrecompile = "0x0000000000000000000000000000000000000801";
  const rate = 38051750380;

  const stakeManagerFactory = await ethers.getContractFactory("StakeManager");
  const stakeManagerProxy = await upgrades.deployProxy(
    stakeManagerFactory.connect(deployer),
    [bridgeAddress, rate, daturaHotkey, stakePrecompile]
  );
  const stakeManagerAddress = await stakeManagerProxy.getAddress();
  console.log("🚀 StakeManager deployed to:", stakeManagerAddress);

  // Save deployment addresses and ABIs
  const deployedContract: DeployedContract = {
    address: tokenAddress,
    abi: JSON.parse(token.interface.formatJson()),
  };
  fs.writeFileSync(
    "./deployed-contract.json",
    JSON.stringify(deployedContract)
  );
}

main().catch((error: Error) => {
  console.error(error);
  process.exitCode = 1;
});
