import dotenv from "dotenv";
import { task } from "hardhat/config";
import { saveItems } from "../utils/utils";
import DeployedContracts from "../result/contract.json";
dotenv.config();

task("deploy:health", "deploy health test").setAction(async ({ v }, { ethers, network, upgrades }) => {
  const accounts = await ethers.getSigners();
  const deployer = accounts[0];
  console.log("deployer:", deployer.address);
  // AtomicSwap contract deploy
  const erc20Factory = await ethers.getContractFactory("MockERC20");
  const erc20 = await erc20Factory.deploy();
  console.log("erc20: token", await erc20.getAddress());
});
