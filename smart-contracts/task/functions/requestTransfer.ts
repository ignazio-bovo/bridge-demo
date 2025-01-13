import { task } from "hardhat/config";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { Bridge, BridgedToken } from "../../typechain-types";
import DeployedContracts from "../result/contract.json";
import { ethers } from "ethers";

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
  .addFlag("native", "Whether the transfer is native or bridged")
  .addParam("destination", "The destination chain ID")
  .setAction(async (taskArgs, hre: HardhatRuntimeEnvironment) => {
    const { to, amount: amountUnit, native, destination } = taskArgs;

    const signers = await hre.ethers.getSigners();
    const admin = signers[0];
    const fromSigner = signers[2];

    const bridgeProxyAddress =
      DeployedContracts[
        `Bridge_${hre.network.name}` as keyof typeof DeployedContracts
      ];
    const bridgedTokenAddress =
      DeployedContracts[
        `BridgedToken_${hre.network.name}` as keyof typeof DeployedContracts
      ];

    // Get the contract factory and attach to the deployed address
    const bridge = (await (await hre.ethers.getContractFactory("Bridge"))
      .attach(bridgeProxyAddress)
      .connect(fromSigner)) as Bridge;

    // Convert amount to Wei
    const amount = hre.ethers.parseEther(amountUnit);

    if (!native) {
      const bridgedToken = (await hre.ethers.getContractFactory("BridgedToken"))
        .attach(bridgedTokenAddress)
        .connect(admin) as BridgedToken;
      await bridgedToken.mint(fromSigner.address, amount);
      await bridgedToken
        .connect(fromSigner)
        .approve(bridgeProxyAddress, amount);
    }

    const ethers = hre.ethers;
    const tokenKey = ethers.keccak256(ethers.toUtf8Bytes("DATURABRIDGE:ETH"));

    const tx = await bridge.requestTransfer(
      tokenKey,
      to,
      amount,
      BigInt(destination),
      {
        value: native ? amount : 0n,
      }
    );
    await tx.wait();
  });
