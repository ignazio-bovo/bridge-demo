import { task } from "hardhat/config";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { Bridge, BridgedToken } from "../../typechain-types";
import DeployedContracts from "../result/contract.json";
import { ethers } from "ethers";
import { fundSubtensorAccount } from "../../scripts/fundSubtensorAccounts";

export const createTokenMetadata = (
  tokenSymbol: string,
  tokenName: string,
  decimals: number
) => {
  const metadata = ethers.AbiCoder.defaultAbiCoder().encode(
    ["string", "string", "uint8"],
    [tokenSymbol, tokenName, decimals]
  );
  return metadata;
};

task(
  "functions:requestTransfer",
  "Requests a transfer from the Bridge contract"
)
  .addParam("to", "The recipient address")
  .addParam("amount", "The amount to transfer (in ETH)")
  .addParam("destination", "The destination chain ID")
  .setAction(async (taskArgs, hre: HardhatRuntimeEnvironment) => {
    const { to, amount: amountUnit, destination } = taskArgs;

    const signers = await hre.ethers.getSigners();
    const fromSigner = signers[2];
    const fromAddress = await fromSigner.getAddress();
    const ethers = hre.ethers;

    const bridgeProxyAddress =
      DeployedContracts[
        `Bridge_${hre.network.name}` as keyof typeof DeployedContracts
      ];

    let bridgedTokenAddress: string;
    let tokenKey: string;
    if (hre.network.name !== "localhost") {
      bridgedTokenAddress = "0x057ef64E23666F000b34aE31332854aCBd1c8544";
      tokenKey = ethers.keccak256(ethers.toUtf8Bytes("DATURABRIDGE:TAO"));
      await fundSubtensorAccount(fromAddress, 10000);
    } else {
      bridgedTokenAddress =
        DeployedContracts[
          `BridgedToken_${hre.network.name}` as keyof typeof DeployedContracts
        ];
      tokenKey = ethers.keccak256(ethers.toUtf8Bytes("DATURABRIDGE:ETH"));
    }

    // Get the contract factory and attach to the deployed address
    const bridge = (await hre.ethers.getContractFactory("Bridge"))
      .attach(bridgeProxyAddress)
      .connect(fromSigner) as Bridge;

    // Convert amount to Wei
    const amount = hre.ethers.parseEther(amountUnit);

    const tx = await bridge
      .connect(fromSigner)
      .requestTransfer(tokenKey, to, amount, BigInt(destination), {
        value: amount,
      });

    await tx.wait();
    console.log("🚀 Transaction sent:", tx.hash);
  });
