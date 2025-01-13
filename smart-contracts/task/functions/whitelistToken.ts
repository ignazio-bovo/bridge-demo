import { task } from "hardhat/config";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { Bridge } from "../../typechain-types";
import DeployedContracts from "../result/contract.json";
import { expect } from "chai";
import { BytesLike, ethers } from "ethers";

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

task("functions:whitelistToken", "Whitelists a token on the Bridge contract")
  .addParam("address", "The address of the token to whitelist")
  .setAction(async (taskArgs, hre: HardhatRuntimeEnvironment) => {
    const { address: tokenAddress } = taskArgs;
    const admin = (await hre.ethers.getSigners())[0];

    const contractKey =
      `Bridge_${hre.network.name}` as keyof typeof DeployedContracts;
    const deployedContract = DeployedContracts[contractKey];

    // Get the contract factory and attach to the deployed address
    const Bridge = await hre.ethers.getContractFactory("Bridge");
    const bridge = Bridge.attach(deployedContract).connect(admin) as Bridge;

    const ethers = hre.ethers;
    const tokenKey = ethers.keccak256(ethers.toUtf8Bytes("DATURABRIDGE:ETH"));
    const tx = await bridge.whitelistToken(
      tokenKey,
      true,
      tokenAddress,
      "ETH",
      "Ether",
      18
    );

    await tx.wait();

    console.log("👉 Token whitelisted");
  });
