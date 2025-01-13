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
  const deployer = new ethers.Wallet(privateKeys[3], provider);
  const admin = new ethers.Wallet(privateKeys[0], provider);
  const authority = new ethers.Wallet(privateKeys[1], provider);

  await fundSubtensorAccount(deployer.address, 10000);
  await fundSubtensorAccount(admin.address, 10000);
  await fundSubtensorAccount(authority.address, 10000);

  let nativeBalance = await provider.getBalance(authority.address);
  console.log("🪙 Authority native balance:", nativeBalance.toString(), "wei");
  console.log("🔑 Authority private key:", privateKeys[1]);

  // Get deployer balance
  const tokenFactory = await ethers.getContractFactory("BridgedToken");
  const token = await tokenFactory
    .connect(deployer)
    .deploy("Ported ETH", "dETH", admin.address, {
      gasLimit: 3000000,
      gasPrice: ethers.parseUnits("50", "gwei"),
    });

  await token.waitForDeployment();

  console.log("🪙 Token deployed at 🎯", token.target);

  const bridgeFactory = await ethers.getContractFactory("Bridge");
  const bridge = await bridgeFactory.connect(deployer).deploy({
    gasLimit: 3000000,
    gasPrice: ethers.parseUnits("50", "gwei"),
  });
  await bridge.waitForDeployment();

  await bridge.initialize(authority.address, admin.address, {
    gasLimit: 3000000,
    gasPrice: ethers.parseUnits("50", "gwei"),
  });
  console.log("🚧 Bridge deployed at 🎯", bridge.target); // should be 0x8f8903DADc4316228C726C6e44dd34800860Fc62

  console.log("🔑 Admin role set for bridge");

  // Save deployment address and ABI
  const deployedContract: DeployedContract = {
    address: token.target.toString(),
    abi: JSON.parse(token.interface.formatJson()),
  };
  fs.writeFileSync(
    "./deployed-contract.json",
    JSON.stringify(deployedContract)
  );

  // Get authority signer balance
  const authorityBalance = await provider.getBalance(authority.address);
  console.log("👛 Authority balance:", authorityBalance.toString(), "wei");
}

main().catch((error: Error) => {
  console.error(error);
  process.exitCode = 1;
});
