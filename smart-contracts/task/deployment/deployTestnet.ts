import dotenv from "dotenv";
import { task } from "hardhat/config";
import { saveItems } from "../utils/utils";
import DeployedContracts from "../result/contract.json";
import { Contract } from "ethers";

dotenv.config();

task("deploy:testnet", "deploy bridge")
  .addParam("v", "contract version")
  .setAction(async ({ v }, { ethers, network, upgrades }) => {
    const [privateKey] = await ethers.getSigners();
    const admin = await privateKey.getAddress();
    const authority = await privateKey.getAddress();

    if (network.name !== "sepolia" && network.name !== "subevm") {
      throw new Error("Network is not sepolia or subtensor");
    }

    const gasPrice = await ethers.provider.getFeeData();
    console.log(
      "⛽ Gas Price:",
      ethers.formatUnits(gasPrice.gasPrice || 0n, "gwei"),
      "gwei"
    );
    console.log(
      "🌊 Max Fee:",
      ethers.formatUnits(gasPrice.maxFeePerGas || 0n, "gwei"),
      "gwei"
    );
    console.log(
      "💨 Max Priority Fee:",
      ethers.formatUnits(gasPrice.maxPriorityFeePerGas || 0n, "gwei"),
      "gwei"
    );

    console.log("👨‍💼 Admin:", admin);
    console.log("🔑 Authority:", authority);
    console.log("🚀 Deployer:", privateKey.address);
    console.log("🌐 Network:", network.name);

    const tokenFactory = await ethers.getContractFactory("BridgedToken");
    const bridgedToken = await tokenFactory.deploy("TestToken", "TST", admin);
    const asset = await bridgedToken.getAddress();

    const factoryName = v === "1" ? "Bridge" : `BridgeV${v}`;
    const bridgeFactory = await ethers.getContractFactory(factoryName);

    let bridgeAddress: string;
    let bridge: Contract;
    const contractKey =
      `Bridge_${network.name}` as keyof typeof DeployedContracts;
    const deployedContract = DeployedContracts[contractKey];

    console.log("Deploying new Bridge contract...");
    if (v === "1" || !deployedContract) {
      // Add custom fee data before deployment
      const customFeeData = {
        maxFeePerGas: ethers.parseUnits("25", "gwei"),
        maxPriorityFeePerGas: ethers.parseUnits("2", "gwei"), // Increase priority fee significantly
      };

      console.log("Attempting deployment with custom gas settings...");
      bridge = await upgrades.deployProxy(
        bridgeFactory.connect(privateKey),
        [authority, admin],
        {
          initializer: "initialize",
          timeout: 300000, // 5 minutes
          kind: "uups",
          pollingInterval: 15000,
          // Add custom gas settings
          ...customFeeData,
        }
      );
      bridgeAddress = await bridge.getAddress();

      const deployTx = bridge.deploymentTransaction();
      if (!deployTx) {
        throw new Error("Deployment transaction failed");
      }

      console.log("Deployment transaction hash:", deployTx.hash);
      console.log("Waiting for confirmations (this may take a few minutes)...");

      try {
        // Wait for more confirmations and log progress
        const receipt = await deployTx.wait(3); // Wait for 3 confirmations
        console.log("Deployment confirmed in block:", receipt?.blockNumber);
      } catch (error: any) {
        if (error.message.includes("timeout")) {
          console.log(
            "\nTransaction is still pending. You can check its status at:"
          );
          console.log(`https://sepolia.etherscan.io/tx/${deployTx.hash}`);
          console.log(
            "\nIf the transaction succeeds, you can run the script again with the same parameters."
          );
        }
        throw error;
      }
    } else {
      console.log("Upgrading existing Bridge contract...");
      try {
        bridge = await upgrades.upgradeProxy(deployedContract, bridgeFactory, {
          kind: "uups",
          timeout: 60000,
          call: {
            fn: "initialize",
            args: [authority, admin],
          },
        });
        bridgeAddress = await bridge.getAddress();
      } catch (error) {
        console.error("Error during upgrade:", error);
        throw error;
      }
    }

    console.log("Bridge contract deployed at:", bridgeAddress);

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

    const daturaHotkey = "0x5GP7c3fFazW9GXK8Up3qgu2DJBk8inu4aK9TZy3RuoSWVCMi";
    const stakePrecompile = "0x0000000000000000000000000000000000000801";
    const rate = 38051750380;
    const stakeManagerFactory = await ethers.getContractFactory("StakeManager");
    const stakeManagerProxy = await upgrades.deployProxy(
      stakeManagerFactory.connect(privateKey),
      [bridgeAddress, rate, daturaHotkey, stakePrecompile]
    );
    const stakeManagerAddress = await stakeManagerProxy.getAddress();
    console.log("🚀 StakeManager deployed to:", stakeManagerAddress);
  });
