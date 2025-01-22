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
  .addParam("name", "The name of the token")
  .addParam("symbol", "The symbol of the token")
  .addParam("decimals", "The decimals of the token")
  .setAction(async (taskArgs, hre: HardhatRuntimeEnvironment) => {
    const { address: tokenAddress, name, symbol, decimals } = taskArgs;
    const admin = (await hre.ethers.getSigners())[0];

    let deployedContract: string;
    if (hre.network.name !== "localhost") {
      deployedContract = "0x057ef64E23666F000b34aE31332854aCBd1c8544";
    } else {
      const deployedContractKey =
        `Bridge_${hre.network.name}` as keyof typeof DeployedContracts;
      deployedContract = DeployedContracts[deployedContractKey];
    }

    // Get the contract factory and attach to the deployed address
    const Bridge = await hre.ethers.getContractFactory("Bridge");
    const bridge = Bridge.attach(deployedContract).connect(admin) as Bridge;

    const ethers = hre.ethers;
    const capSymbol = symbol.toUpperCase();
    const tokenKey = ethers.keccak256(
      ethers.toUtf8Bytes(`DATURABRIDGE:${capSymbol}`)
    );
    const tx = await bridge.whitelistToken(
      tokenKey,
      true,
      tokenAddress,
      capSymbol,
      name,
      decimals
    );

    await tx.wait();

    console.log(`👉 Token whitelisted ${name} (${symbol})`);
  });
