import { ethers } from "hardhat";
import {
  privateKeys,
  fundSubtensorAccount,
  subtensorExtraConfig,
} from "./fundSubtensorAccounts";

async function main(): Promise<void> {
  const provider = new ethers.JsonRpcProvider(subtensorExtraConfig.httpUrl);
  const user = new ethers.Wallet(privateKeys[2], provider);

  await fundSubtensorAccount(user.address, 10000);

  const bridge = await ethers.getContractAt(
    "Bridge",
    "0x8f8903DADc4316228C726C6e44dd34800860Fc62"
  );

  const tokenKey = ethers.keccak256(ethers.toUtf8Bytes("DATURABRIDGE:ETH"));

  const tx = await bridge
    .connect(user)
    .requestTransfer(
      tokenKey,
      "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
      ethers.parseEther("0.1"),
      1,
      {
        gasLimit: 3000000,
        gasPrice: ethers.parseUnits("50", "gwei"),
        value: ethers.parseEther("0.1"),
      }
    );

  console.log("🚀 Transaction sent:", tx.hash);

  await tx.wait();
}

main().catch((error: Error) => {
  console.error(error);
  process.exitCode = 1;
});
