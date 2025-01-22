import { ethers } from "hardhat";
import {
  privateKeys,
  fundSubtensorAccount,
  subtensorExtraConfig,
} from "./subtensorUtils";

async function main(): Promise<void> {
  const accounts = await ethers.getSigners();
  const admin = accounts[0];
  const address = await admin.getAddress();

  await fundSubtensorAccount(address, 10000);
  const bridge = await ethers.getContractAt(
    "Bridge",
    "0x057ef64E23666F000b34aE31332854aCBd1c8544",
    admin
  );

  const tokenKey = ethers.keccak256(ethers.toUtf8Bytes("DATURABRIDGE:TAO"));

  const tx = await bridge.whitelistToken(
    tokenKey,
    true,
    "0x0000000000000000000000000000000000000000",
    "TAO",
    "Tao",
    9,
    {
      gasLimit: 6000000,
    }
  );

  console.log("🚀 Transaction sent:", tx.hash);
  await tx.wait();
}

main().catch((error: Error) => {
  console.error(error);
  process.exitCode = 1;
});
