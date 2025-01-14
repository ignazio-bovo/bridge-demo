import { task } from "hardhat/config";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { Bridge } from "../../typechain-types";
import DeployedContracts from "../result/contract.json";
import { ethers, Signer } from "ethers";
import { expect } from "chai";

export const createTokenMetadata = (
  tokenSymbol: string,
  tokenName: string,
  decimals: number
) => {
  return {
    symbol: tokenSymbol,
    name: tokenName,
    decimals,
  };
};

task(
  "functions:confirmTransferRequest",
  "Confirms a transfer request on the Bridge contract"
)
  .addParam("sourcechainid", "The source chain id")
  .addParam("to", "The recipient address")
  .addParam("amount", "The amount to transfer (in ETH)")
  .setAction(async (taskArgs, hre: HardhatRuntimeEnvironment) => {
    const { sourcechainid: sourceChainId, to, amount } = taskArgs;
    // Deploy contracts
    const accounts = await hre.ethers.getSigners();
    const fromSigner = accounts[2];

    const authority = accounts[1];

    let deployedContract: string;
    if (hre.network.name === "localhost") {
      const contractKey =
        `Bridge_${hre.network.name}` as keyof typeof DeployedContracts;
      deployedContract = DeployedContracts[contractKey];
    } else {
      deployedContract = "0x057ef64E23666F000b34aE31332854aCBd1c8544";
    }

    // Get the contract factory and attach to the deployed address
    const Bridge = await hre.ethers.getContractFactory("Bridge");
    const bridge = Bridge.attach(deployedContract).connect(authority) as Bridge;

    // Convert amount to Wei
    const amountWei = hre.ethers.parseEther(amount);

    // Send the amountWei from admin to the bridge contract
    const bridgeAddress = await bridge.getAddress();
    const admin = accounts[0];
    await sendEth(admin, amountWei, bridgeAddress);

    const nonce = 0; // unstable: needs to restart the env every time

    const ethers = hre.ethers;
    const tokenKey = ethers.keccak256(ethers.toUtf8Bytes("DATURABRIDGE:ETH"));
    const from = fromSigner.address;
    const request = {
      amount: amountWei,
      nonce,
      from,
      to,
      tokenKey,
      tokenMetadata: createTokenMetadata("ETH", "Ethereum", 18),
      srcChainId: sourceChainId,
    };

    // Get authority balance before executing transfer
    const authorityBalance = await hre.ethers.provider.getBalance(
      authority.address
    );
    console.log(
      "💸 Authority balance:",
      hre.ethers.formatEther(authorityBalance),
      "ETH"
    );
    const tx = await bridge.executeTransferRequests([request]);

    console.log("🚀 Transaction sent:", tx.toJSON());

    const receipt = await tx.wait();
    expect(receipt, "Transaction failed");
    expect(
      receipt!.logs.length > 0,
      "At least wrapped token event should have been emitted"
    );
  });

async function sendEth(admin: Signer, amount: bigint, bridgeAddress: string) {
  const tx = await admin.sendTransaction({
    to: bridgeAddress,
    value: amount,
  });
  await tx.wait();
}
