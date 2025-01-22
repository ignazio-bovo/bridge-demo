import { ethers } from "hardhat";
import {
  privateKeys,
  fundSubtensorAccount,
  subtensorExtraConfig,
} from "./subtensorUtils";

async function main(): Promise<void> {
  const accounts = await ethers.getSigners();
  const user = accounts[2];
  const address = await user.getAddress();

  await fundSubtensorAccount(address, 100000);

  const bridge = await ethers.getContractAt(
    "Bridge",
    "0x057ef64E23666F000b34aE31332854aCBd1c8544",
    user
  );

  const tokenKey = ethers.keccak256(ethers.toUtf8Bytes("DATURABRIDGE:TAO"));
  const amount = ethers.parseEther("0.1");

  const tx = await bridge.requestTransfer(
    tokenKey,
    "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
    amount,
    1,
    {
      gasLimit: 6000000,
      value: amount,
    }
  );

  const receipt = await tx.wait();
  if (receipt?.status === 0) {
    throw new Error(`Transaction failed: ${tx.hash}`);
  }
  console.log("🚀 Transaction sent:", tx.hash);
}

main().catch((error: Error) => {
  console.error(error);
  process.exitCode = 1;
});
