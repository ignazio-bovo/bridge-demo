import dotenv from "dotenv";
import { task } from "hardhat/config";
import { saveItems } from "../utils/utils";
import DeployedContracts from "../result/contract.json";
import { Contract } from "ethers";

dotenv.config();

task("deploy:bridge", "deploy bridge")
  .addParam("v", "contract version")
  .setAction(async ({ v }, { ethers, network, upgrades }) => {
    const accounts = await ethers.getSigners();
    const deployer = accounts[3];
    const admin = await accounts[0].getAddress();
    const authority = await accounts[1].getAddress();
    const deployerAddress = await deployer.getAddress();

    console.log("👨‍💼 Admin:", admin);
    console.log("🔑 Authority:", authority);
    console.log("🚀 Deployer:", deployerAddress);
    console.log("🌐 Network:", network.name);

    const tokenFactory = await ethers.getContractFactory("BridgedToken");
    const bridgedToken = await tokenFactory.deploy("TestToken", "TST", admin);

    const factoryName = v === "1" ? "Bridge" : `BridgeV${v}`;
    const bridgeFactory = await ethers.getContractFactory(factoryName);

    let bridgeAddress: string;
    let bridge: Contract;
    const contractKey =
      `Bridge_${network.name}` as keyof typeof DeployedContracts;
    const deployedContract = DeployedContracts[contractKey];

    if (v === "1" || !deployedContract) {
      bridge = await upgrades.deployProxy(
        bridgeFactory.connect(accounts[1]), // Connect with authority account (accounts[1])
        [authority, admin],
        {
          initializer: "initialize",
          timeout: 60000,
          kind: "uups",
        }
      );
      bridgeAddress = await bridge.getAddress();
      await bridge.deploymentTransaction()?.wait();
      console.log("🚀 Bridge deployed to:", bridgeAddress);
    } else {
      console.log("Upgrading existing Bridge contract...");
      try {
        bridge = await upgrades.upgradeProxy(deployedContract, bridgeFactory, {
          kind: "uups",
          timeout: 60000,
          call: {
            fn: "upgradeInitialize",
            args: [authority, admin],
          },
        });
        bridgeAddress = await bridge.getAddress();
      } catch (error) {
        console.error("Error during upgrade:", error);
        throw error;
      }
    }

    // Verify the upgrade for v2 and above
    if (v !== "1") {
      try {
        await bridge.testUpgrade();
        console.log("Upgrade verified successfully");
      } catch (error) {
        console.error("Error verifying upgrade:", error);
      }
    }

    await saveItems([
      {
        title: `Bridge_${network.name}`,
        value: bridgeAddress,
      },
      {
        title: `BridgedToken_${network.name}`,
        value: await bridgedToken.getAddress(),
      },
    ]);

    const daturaHotkey =
      "0x0000000000000000000000000000000000000000000000000000000000000001"; // 32 bytes hex string
    const stakePrecompile = "0x0000000000000000000000000000000000000801";
    const rate = 38051750380;
    const stakeManagerFactory = await ethers.getContractFactory("StakeManager");
    const stakeManagerProxy = await upgrades.deployProxy(
      stakeManagerFactory.connect(deployer),
      [bridgeAddress, rate, daturaHotkey, stakePrecompile]
    );
    const stakeManagerAddress = await stakeManagerProxy.getAddress();
    console.log("🚀 StakeManager deployed to:", stakeManagerAddress);

    await saveItems([
      {
        title: `StakeManager_${network.name}`,
        value: stakeManagerAddress,
      },
    ]);
  });
