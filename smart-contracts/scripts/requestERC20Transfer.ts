import { ethers } from "hardhat";
import {
  privateKeys,
  fundSubtensorAccount,
  subtensorExtraConfig,
} from "./fundSubtensorAccounts";
import { Bridge, BridgedToken } from "../typechain-types";

async function main(): Promise<void> {
  const provider = new ethers.JsonRpcProvider(subtensorExtraConfig.httpUrl);
  const user = new ethers.Wallet(privateKeys[2], provider);
  const admin = new ethers.Wallet(privateKeys[0], provider);

  await fundSubtensorAccount(user.address, 10000);
  await fundSubtensorAccount(admin.address, 10000);

  const bridge = (await ethers.getContractFactory("Bridge"))
    .attach("0x8f8903DADc4316228C726C6e44dd34800860Fc62")
    .connect(admin) as Bridge;

  const token = (await ethers.getContractFactory("BridgedToken"))
    .attach("0x01A64AA532801026d4856e437A893ab1d7992c92")
    .connect(admin) as BridgedToken;

  await (
    await token.connect(admin).mint(user.address, ethers.parseEther("1"))
  ).wait();

  await (
    await token.connect(user).approve(bridge.target, ethers.parseEther("1"))
  ).wait();

  const tx = await bridge
    .connect(admin)
    .requestTransfer(user.address, ethers.parseEther("0.1"), 31337, false, {
      gasLimit: 3000000,
      gasPrice: ethers.parseUnits("50", "gwei"),
      value: ethers.parseEther("0.1"),
    });

  console.log("🚀 Transaction sent:", tx.hash);
  await tx.wait();
}

main().catch((error: Error) => {
  console.error(error);
  process.exitCode = 1;
});
