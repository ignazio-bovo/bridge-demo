import {
  loadFixture,
  mine,
} from "@nomicfoundation/hardhat-toolbox/network-helpers";
import { expect } from "chai";
import { ethers, upgrades } from "hardhat";
import { HardhatEthersSigner } from "@nomicfoundation/hardhat-ethers/signers";
import {
  Bridge,
  BridgedToken,
  MockERC20,
  StakeManager,
  StakePrecompileMock,
} from "../typechain-types";
import { anyValue } from "@nomicfoundation/hardhat-chai-matchers/withArgs";

const createTokenKey = (tokenSymbol: string) => {
  const key = "DATURABRIDGETOKEN" + ":" + tokenSymbol;
  return ethers.keccak256(ethers.toUtf8Bytes(key));
};

export const createTokenMetadata = (
  tokenSymbol: string,
  tokenName: string,
  decimals: number
) => {
  return {
    name: tokenName,
    symbol: tokenSymbol,
    decimals: decimals,
  };
};

async function deployBridgeFixture() {
  const [admin, authority] = await ethers.getSigners();

  // Deploy Bridge implementation and proxy
  const BridgeFactory = await ethers.getContractFactory("Bridge");

  const bridgeProxy = await upgrades.deployProxy(
    BridgeFactory,
    [authority.address, admin.address],
    {
      initializer: "initialize",
      kind: "uups",
    }
  );

  const bridge = BridgeFactory.attach(await bridgeProxy.getAddress()) as Bridge;

  return {
    bridge,
    admin,
    authority,
  };
}

async function deployWithStakeManagerFixture() {
  const { bridge, admin, authority } = await loadFixture(deployBridgeFixture);
  const rewardRate = BigInt(1e17) / BigInt(5256000); // 10% APY
  const daturaHotkey = ethers.keccak256(ethers.toUtf8Bytes("datura"));
  const StakeManagerFactory = await ethers.getContractFactory("StakeManager");

  // deploy stake precompile mock
  const stakePrecompileFactory = await ethers.getContractFactory(
    "StakePrecompileMock"
  );
  const stakePrecompile = await stakePrecompileFactory.deploy();
  const stakeManagerProxy = await upgrades.deployProxy(
    StakeManagerFactory,
    [bridge.target, rewardRate, daturaHotkey, stakePrecompile.target],
    {
      initializer: "initialize",
      kind: "uups",
    }
  );

  const stakeManager = StakeManagerFactory.attach(
    await stakeManagerProxy.getAddress()
  ) as StakeManager;

  await bridge.setStakingManager(stakeManager.target);

  return {
    bridge,
    admin,
    authority,
    stakeManager,
    stakePrecompile,
  };
}

async function deployBridgeWithStakeManagerAndTokensFixture() {
  const { bridge, admin, authority, stakeManager, stakePrecompile } =
    await loadFixture(deployWithStakeManagerFixture);
  const [, , user, recipient] = await ethers.getSigners();

  // Deploy mock token
  const BridgedTokenFactory = await ethers.getContractFactory("BridgedToken");
  const dTao = await BridgedTokenFactory.deploy(
    "Datura TAO",
    "dTAO",
    admin.address
  );

  await dTao
    .connect(admin)
    .grantRole(await dTao.DEFAULT_ADMIN_ROLE(), bridge.target);

  const MaticFactory = await ethers.getContractFactory("MockERC20");
  const matic = (await MaticFactory.connect(admin).deploy(
    ethers.parseEther("1000000000000000000000000")
  )) as MockERC20;

  return {
    bridge: bridge,
    stakeManager,
    dTao,
    admin,
    user,
    authority,
    recipient,
    matic: matic.connect(admin) as MockERC20,
    stakePrecompile,
  };
}

async function deployBridgeWithTokensFixture() {
  const { bridge, admin, authority } = await loadFixture(deployBridgeFixture);
  const [, , user, recipient] = await ethers.getSigners();

  // Deploy mock token
  const BridgedTokenFactory = await ethers.getContractFactory("BridgedToken");
  const dTao = await BridgedTokenFactory.deploy(
    "Datura TAO",
    "dTAO",
    admin.address
  );

  await dTao
    .connect(admin)
    .grantRole(await dTao.DEFAULT_ADMIN_ROLE(), bridge.target);

  const MaticFactory = await ethers.getContractFactory("MockERC20");
  const matic = (await MaticFactory.connect(admin).deploy(
    ethers.parseEther("1000000000000000000000000")
  )) as MockERC20;

  return {
    bridge: bridge,
    dTao,
    admin,
    user,
    authority,
    recipient,
    matic: matic.connect(admin) as MockERC20,
  };
}

describe("Bridge Test Suite", function () {
  describe("Bridge Deployment", function () {
    let bridge: Bridge;
    let admin: HardhatEthersSigner;
    let authority: HardhatEthersSigner;
    beforeEach(async function () {
      const fixture = await loadFixture(deployBridgeFixture);
      bridge = fixture.bridge;
      admin = fixture.admin;
      authority = fixture.authority;
    });
    it("Should have non-zero address", async function () {
      expect(await bridge.getAddress()).to.not.equal(ethers.ZeroAddress);
    });

    it("Should set admin role correctly", async function () {
      const adminRole = await bridge.DEFAULT_ADMIN_ROLE();
      expect(await bridge.hasRole(adminRole, admin.address)).to.equal(true);
    });

    it("Should set authority correctly", async function () {
      const authorityRole = await bridge.AUTHORITY_ROLE();
      expect(await bridge.hasRole(authorityRole, authority.address)).to.equal(
        true
      );
    });

    it("Should set pauser role correctly", async function () {
      const pauserRole = await bridge.PAUSER_ROLE();
      expect(await bridge.hasRole(pauserRole, admin.address)).to.equal(true);
    });
  });

  describe("Bridged Token", function () {
    it("Should fail minting without admin role", async function () {
      const { dTao, user } = await loadFixture(deployBridgeWithTokensFixture);
      const amount = ethers.parseEther("100");

      await expect(
        dTao.connect(user).mint(user.address, amount)
      ).to.be.revertedWithCustomError(dTao, "AccessControlUnauthorizedAccount");
    });

    it("Should fail burning without admin role", async function () {
      const { dTao, user, admin } = await loadFixture(
        deployBridgeWithTokensFixture
      );
      const amount = ethers.parseEther("100");

      // First mint some tokens to burn
      await dTao.connect(admin).mint(user.address, amount);

      await expect(
        dTao.connect(user).burnFrom(user.address, amount)
      ).to.be.revertedWithCustomError(dTao, "AccessControlUnauthorizedAccount");
    });
    it("Should fail when bridge is paused", async function () {
      const { bridge, admin, dTao, user } = await loadFixture(
        deployBridgeWithTokensFixture
      );
      await bridge.connect(admin).pause();
      const amount = ethers.parseEther("100");
      const chainId = 945;

      await expect(
        bridge
          .connect(user)
          .requestTransfer(createTokenKey("ETH"), user.address, amount, chainId)
      ).to.be.revertedWithCustomError(bridge, "EnforcedPause");
    });

    it("Should mint tokens successfully with admin role", async function () {
      const { dTao, user, admin } = await loadFixture(
        deployBridgeWithTokensFixture
      );
      const amount = ethers.parseEther("100");
      const recipient = user.address;

      await dTao.connect(admin).mint(recipient, amount);
      expect(await dTao.balanceOf(recipient)).to.equal(amount);
    });

    it("Should burn tokens successfully with admin role", async function () {
      const { dTao, user, admin } = await loadFixture(
        deployBridgeWithTokensFixture
      );
      const amount = ethers.parseEther("100");
      const holder = user.address;

      // First mint tokens
      await dTao.connect(admin).mint(holder, amount);
      expect(await dTao.balanceOf(holder)).to.equal(amount);

      // Then burn them
      await dTao.connect(admin).burnFrom(holder, amount);
      expect(await dTao.balanceOf(holder)).to.equal(0);
    });

    it("Should correctly check admin role", async function () {
      const { dTao, admin, user } = await loadFixture(
        deployBridgeWithTokensFixture
      );

      expect(await dTao.isAdmin(admin.address)).to.be.true;
      expect(await dTao.isAdmin(user.address)).to.be.false;
    });
  });

  describe("Pausing", function () {
    let bridge: Bridge;
    let admin: HardhatEthersSigner;
    let user: HardhatEthersSigner;
    beforeEach(async function () {
      const fixture = await loadFixture(deployBridgeWithTokensFixture);
      bridge = fixture.bridge;
      admin = fixture.admin;
      user = fixture.user;
    });
    it("Should fail when non admin tries to pause", async function () {
      await expect(bridge.connect(user).pause()).to.be.revertedWithCustomError(
        bridge,
        "AccessControlUnauthorizedAccount"
      );
    });
    it("Should pause successfully", async function () {
      await expect(bridge.connect(admin).pause()).not.to.be.reverted;
    });
    it("Should fail when non admin tries to unpause", async function () {
      await expect(bridge.connect(admin).pause()).not.to.be.reverted;
      await expect(
        bridge.connect(user).unpause()
      ).to.be.revertedWithCustomError(
        bridge,
        "AccessControlUnauthorizedAccount"
      );
    });
    it("Should unpause successfully", async function () {
      await expect(bridge.connect(admin).pause()).not.to.be.reverted;
      await expect(bridge.connect(admin).unpause()).not.to.be.reverted;
    });
  });
  describe("Whitelisting", function () {
    let bridge: Bridge;
    let admin: HardhatEthersSigner;
    let user: HardhatEthersSigner;
    let token: string;

    beforeEach(async function () {
      const fixture = await loadFixture(deployBridgeWithTokensFixture);
      bridge = fixture.bridge;
      admin = fixture.admin;
      user = fixture.user;
      token = ethers.ZeroAddress;
    });

    it("Should fail when non admin tries to whitelist token", async function () {
      await expect(
        bridge
          .connect(user)
          .whitelistToken(
            createTokenKey("ETH"),
            true,
            ethers.ZeroAddress,
            "ETH",
            "Ether",
            18
          )
      ).to.be.revertedWithCustomError(
        bridge,
        "AccessControlUnauthorizedAccount"
      );
    });

    it("Should successfully whitelist token", async function () {
      await bridge
        .connect(admin)
        .whitelistToken(
          createTokenKey("ETH"),
          true,
          ethers.ZeroAddress,
          "ETH",
          "Ether",
          18
        );
      await expect(
        bridge.tokensInfo(createTokenKey("ETH"))
      ).to.eventually.deep.equal([ethers.ZeroAddress, false, true, true]);
    });
    it("Should fail when token is already whitelisted", async function () {
      await bridge
        .connect(admin)
        .whitelistToken(
          createTokenKey("ETH"),
          true,
          ethers.ZeroAddress,
          "ETH",
          "Ether",
          18
        );
    });
    it("Should emit TokenWhitelistStatusUpdated event when token is whitelisted", async function () {
      await expect(
        bridge
          .connect(admin)
          .whitelistToken(
            createTokenKey("ETH"),
            true,
            ethers.ZeroAddress,
            "ETH",
            "Ether",
            18
          )
      )
        .to.emit(bridge, "NewTokenWhitelisted")
        .withArgs(
          createTokenKey("ETH"),
          [ethers.ZeroAddress, false, true, true],
          ["Ether", "ETH", 18n]
        );
    });
    it("Should add token metadata to storage correctly", async function () {
      await expect(
        bridge
          .connect(admin)
          .whitelistToken(
            createTokenKey("ETH"),
            true,
            ethers.ZeroAddress,
            "ETH",
            "Ether",
            18
          )
      ).to.not.be.reverted;

      expect(
        await bridge.metadataForToken(createTokenKey("ETH"))
      ).to.deep.equal(["Ether", "ETH", 18n]);
    });
  });

  describe("Transfer Request Revert Cases", function () {
    let bridge: Bridge;
    let dTao: BridgedToken;
    let matic: MockERC20;
    let user: HardhatEthersSigner;
    let admin: HardhatEthersSigner;
    let chainId: number;
    const ethTokenKey = createTokenKey("ETH");
    const maticTokenKey = createTokenKey("MATIC");

    beforeEach(async function () {
      const fixture = await loadFixture(deployBridgeWithTokensFixture);
      bridge = fixture.bridge;
      dTao = fixture.dTao;
      matic = fixture.matic;
      user = fixture.user;
      admin = fixture.admin;
      chainId = 945;
    });

    it("Should revert when destination chain is current chain", async function () {
      const hardhatChainId = 31337; // hardhat chain id used in the testing environment
      const amount = ethers.parseEther("100");

      await expect(
        bridge
          .connect(user)
          .requestTransfer(ethTokenKey, user.address, amount, hardhatChainId)
      ).to.be.revertedWithCustomError(bridge, "InvalidDestinationChain");
    });

    it("Should revert when token is not whitelisted", async function () {
      const amount = ethers.parseEther("100");

      await expect(
        bridge
          .connect(user)
          .requestTransfer(ethTokenKey, user.address, amount, chainId)
      ).to.be.revertedWithCustomError(bridge, "TokenNotWhitelisted");
    });

    it("Should revert when recipient is zero address", async function () {
      const amount = ethers.parseEther("100");

      await bridge
        .connect(admin)
        .whitelistToken(
          ethTokenKey,
          true,
          ethers.ZeroAddress,
          "ETH",
          "Ether",
          18
        );

      await expect(
        bridge
          .connect(user)
          .requestTransfer(ethTokenKey, ethers.ZeroAddress, amount, chainId)
      ).to.be.revertedWithCustomError(bridge, "InvalidRecipient");
    });

    it("Should revert when amount is zero", async function () {
      const amount = ethers.parseEther("0");
      await bridge
        .connect(admin)
        .whitelistToken(
          ethTokenKey,
          true,
          ethers.ZeroAddress,
          "ETH",
          "Ether",
          18
        );

      await expect(
        bridge
          .connect(user)
          .requestTransfer(ethTokenKey, user.address, amount, chainId)
      ).to.be.revertedWithCustomError(bridge, "InvalidAmount");
    });

    it("Transfer fails with insufficient allowance for non bridged erc20 token", async function () {
      const amount = ethers.parseEther("100");
      await bridge.whitelistToken(
        maticTokenKey,
        true,
        matic.target,
        "MATIC",
        "Matic",
        18
      );
      await matic.connect(admin).transfer(user.address, amount);

      await expect(
        bridge
          .connect(user)
          .requestTransfer(maticTokenKey, user.address, amount, chainId)
      ).to.be.revertedWithCustomError(bridge, "InsufficientAllowance");
    });

    describe("Should revert when amount is less than specified", function () {
      it("When token is a bridged token", async function () {
        const taoTokenKey = createTokenKey("TAO");
        const amountInt = 100;
        const amount = ethers.parseEther(amountInt.toString());
        await bridge
          .connect(admin)
          .whitelistToken(taoTokenKey, true, dTao.target, "TAO", "Tao", 9);
        await dTao.connect(admin).mint(user.address, amount);
        await dTao.connect(user).approve(bridge.target, amount);

        await expect(
          bridge
            .connect(user)
            .requestTransfer(
              taoTokenKey,
              user.address,
              ethers.parseEther((amountInt + 100).toString()),
              chainId
            )
        ).to.be.reverted;
      });

      it("When token is a non bridged erc20 token", async function () {
        const amountInt = 100;
        const amount = ethers.parseEther(amountInt.toString());
        await bridge.whitelistToken(
          maticTokenKey,
          true,
          matic.target,
          "MATIC",
          "Matic",
          18
        );
        await matic.connect(admin).transfer(user.address, amount);
        await matic.connect(user).approve(bridge.target, amount);

        await expect(
          bridge
            .connect(user)
            .requestTransfer(
              maticTokenKey,
              user.address,
              ethers.parseEther((amountInt + 100).toString()),
              chainId
            )
        ).to.be.reverted;
      });

      it("When token is native", async function () {
        const amount = (await ethers.provider.getBalance(user.address)) + 1n;

        await bridge
          .connect(admin)
          .whitelistToken(
            ethTokenKey,
            true,
            ethers.ZeroAddress,
            "ETH",
            "Ether",
            18
          );

        await expect(
          bridge
            .connect(user)
            .requestTransfer(ethTokenKey, user.address, amount, chainId)
        ).to.be.reverted;
      });
    });
  });
  describe("Transfer Request With Hosted Tokens", function () {
    const amountMatic = ethers.parseEther("100");
    const amountEth = ethers.parseEther("100");
    const maticTokenKey = createTokenKey("MATIC");
    const ethTokenKey = createTokenKey("ETH");
    let bridge: Bridge;
    let matic: MockERC20;
    let user: HardhatEthersSigner;
    let admin: HardhatEthersSigner;
    let chainId: number;

    beforeEach(async function () {
      const fixture = await loadFixture(deployBridgeWithTokensFixture);
      bridge = fixture.bridge;
      matic = fixture.matic;
      user = fixture.user;
      admin = fixture.admin;
      chainId = 945;

      await bridge
        .connect(admin)
        .whitelistToken(
          maticTokenKey,
          true,
          matic.target,
          "MATIC",
          "Matic",
          18
        );
      await bridge
        .connect(admin)
        .whitelistToken(
          ethTokenKey,
          true,
          ethers.ZeroAddress,
          "ETH",
          "Ether",
          18
        );

      // move matic tokens to user
      await matic.connect(admin).transfer(user.address, amountMatic);
      await matic.connect(user).approve(bridge.target, amountMatic);
    });

    describe("Hosted ERC20 token transfers", function () {
      it("Should emit TransferRequested event with correct parameters", async function () {
        const request = [
          0,
          user.address,
          user.address,
          maticTokenKey,
          amountMatic,
          31337, // source chain id
          chainId,
        ];

        await expect(
          bridge
            .connect(user)
            .requestTransfer(maticTokenKey, user.address, amountMatic, chainId)
        )
          .to.emit(bridge, "TransferRequested")
          .withArgs(request);
      });

      describe("Accounting has lock semantics", function () {
        beforeEach(async function () {
          await expect(
            bridge
              .connect(user)
              .requestTransfer(
                maticTokenKey,
                user.address,
                amountMatic,
                chainId
              )
          )
            .to.emit(matic, "Transfer")
            .withArgs(user.address, bridge.target, amountMatic);
        });

        it("should decrease user token balance", async function () {
          await expect(matic.balanceOf(user.address)).to.eventually.equal(0);
        });

        it("should increase bridge token balance", async function () {
          await expect(matic.balanceOf(bridge.target)).to.eventually.equal(
            amountMatic
          );
        });
      });
    });

    describe("Native token transfers", function () {
      it("Should emit TransferRequested event with correct parameters", async function () {
        const request = [
          0,
          user.address,
          user.address,
          ethTokenKey,
          amountEth,
          31337, // source chain id
          chainId,
        ];

        await expect(
          bridge
            .connect(user)
            .requestTransfer(ethTokenKey, user.address, amountEth, chainId, {
              value: amountEth,
            })
        )
          .to.emit(bridge, "TransferRequested")
          .withArgs(request);
      });

      describe("Accounting has lock semantics", function () {
        let userBalancePre: bigint;
        let bridgeBalancePre: bigint;

        beforeEach(async function () {
          userBalancePre = await ethers.provider.getBalance(user.address);
          bridgeBalancePre = await ethers.provider.getBalance(bridge.target);

          await expect(
            bridge
              .connect(user)
              .requestTransfer(ethTokenKey, user.address, amountEth, chainId, {
                value: amountEth,
              })
          ).to.not.be.reverted;
        });

        it("should decrease user eth balance", async function () {
          const userBalancePost = await ethers.provider.getBalance(
            user.address
          );
          // accounts for gas fees
          expect(userBalancePost).to.be.lessThan(userBalancePre - amountEth);
        });

        it("should increase bridge eth balance", async function () {
          await expect(
            ethers.provider.getBalance(bridge.target)
          ).to.eventually.equal(bridgeBalancePre + amountEth);
        });
      });
    });

    it("Should increment bridge nonce for chainId after successful request", async function () {
      await admin.sendTransaction({
        to: user.address,
        value: amountEth,
      });

      expect(await bridge.bridgeNonce()).to.equal(0);

      await expect(
        bridge
          .connect(user)
          .requestTransfer(ethTokenKey, user.address, amountEth, chainId, {
            value: amountEth,
          })
      ).to.not.be.reverted;

      expect(await bridge.bridgeNonce()).to.equal(1);

      await expect(
        bridge
          .connect(user)
          .requestTransfer(ethTokenKey, user.address, amountEth, chainId, {
            value: amountEth,
          })
      ).to.not.be.reverted;

      expect(await bridge.bridgeNonce()).to.equal(2);
    });
  });
  describe("Execute Transfer Requests With Revert Cases", function () {
    let bridge: Bridge;
    let user: HardhatEthersSigner;
    let admin: HardhatEthersSigner;
    let authority: HardhatEthersSigner;
    let dEth: BridgedToken;
    let matic: MockERC20;

    beforeEach(async function () {
      const fixture = await loadFixture(deployBridgeWithTokensFixture);
      bridge = fixture.bridge;
      user = fixture.user;
      admin = fixture.admin;
      authority = fixture.authority;
      dEth = fixture.dTao; // simulate transfer from Bittensor EVM with ported srcToken = dETH
      matic = fixture.matic;
    });

    it("Should revert when called by non-authority", async function () {
      const batch = [
        {
          tokenKey: createTokenKey("ETH"),
          to: user.address,
          amount: ethers.parseEther("100"),
          nonce: 0,
          srcChainId: 945,
          tokenMetadata: createTokenMetadata("ETH", "Ethereum", 18),
        },
      ];
      await expect(bridge.connect(user).executeTransferRequests(batch)).to.be
        .reverted;
    });

    it("Should revert when transfer request is already processed", async function () {
      const batch = [
        {
          tokenKey: createTokenKey("TAO"),
          to: user.address,
          amount: ethers.parseEther("100"),
          nonce: 0,
          srcChainId: 945,
          tokenMetadata: createTokenMetadata("ETH", "Ethereum", 18),
        },
      ];
      await expect(bridge.connect(authority).executeTransferRequests(batch)).to
        .not.be.reverted;

      await expect(
        bridge.connect(authority).executeTransferRequests(batch)
      ).to.be.revertedWithCustomError(bridge, "TransferAlreadyProcessed");
    });

    it("Should revert when batch is empty", async function () {
      const batch: any[] = [];

      await expect(
        bridge.connect(authority).executeTransferRequests(batch)
      ).to.be.revertedWithCustomError(bridge, "InvalidInput");
    });

    it("Should revert when batch exceeds max length", async function () {
      const batch: any[] = [];
      for (let i = 0; i <= 100; i++) {
        batch.push({
          tokenKey: createTokenKey("ETH"),
          to: user.address,
          amount: ethers.parseEther("100"),
          nonce: i,
          srcChainId: 945,
          tokenMetadata: createTokenMetadata("ETH", "Ethereum", 18),
        });
      }

      await expect(
        bridge.connect(authority).executeTransferRequests(batch)
      ).to.be.revertedWithCustomError(bridge, "InvalidInput");
    });

    it("Cannot wrap a token that is already wrapped", async function () {
      const batchWithMatic = [
        {
          tokenKey: createTokenKey("MATIC"),
          to: user.address,
          amount: ethers.parseEther("100"),
          nonce: 0,
          srcChainId: 945,
          tokenMetadata: createTokenMetadata("MATIC", "Polygon", 18),
        },
      ];
      await expect(
        bridge.connect(authority).executeTransferRequests(batchWithMatic)
      ).to.not.be.reverted;
      batchWithMatic[0].nonce = 1; // retry changing the nonce
      await expect(
        bridge.connect(authority).executeTransferRequests(batchWithMatic)
      ).to.not.emit(bridge, "TokenWrapped");
    });
  });

  describe("Execute Transfer Requests With Hosted Tokens", function () {
    let bridge: Bridge;
    let user: HardhatEthersSigner;
    let admin: HardhatEthersSigner;
    let authority: HardhatEthersSigner;
    let matic: MockERC20;
    const maticAmount = ethers.parseEther("100");
    const maticKey = createTokenKey("MATIC");

    beforeEach(async function () {
      const fixture = await loadFixture(deployBridgeWithTokensFixture);
      bridge = fixture.bridge;
      user = fixture.user;
      admin = fixture.admin;
      authority = fixture.authority;
      matic = fixture.matic;
      expect(bridge).not.to.be.undefined;

      await bridge
        .connect(admin)
        .whitelistToken(maticKey, true, matic.target, "MATIC", "Matic", 18);
      await matic.connect(admin).transfer(user.address, maticAmount);
      await matic.connect(user).approve(bridge.target, maticAmount);
    });

    describe("Bridging Hosted Tokens Back to Source Chain", function () {
      const ethTokenKey = createTokenKey("ETH");
      let batchWithdEth: any[];
      describe("Accounting Hosted Native Token has unlock semantics", function () {
        const chainId = 945;
        const amount = ethers.parseEther("100");
        let balanceUserPreUnlock: bigint;
        let balanceBridgePreUnlock: bigint;
        beforeEach(async function () {
          // first user bridges 100 ETH to the bridge
          await bridge
            .connect(admin)
            .whitelistToken(
              ethTokenKey,
              true,
              ethers.ZeroAddress,
              "ETH",
              "Ether",
              18
            );
          await expect(
            bridge
              .connect(user)
              .requestTransfer(ethTokenKey, user.address, amount, chainId, {
                value: amount,
              })
          ).not.to.be.reverted;

          // on the other chain user requests to bridge Back to Bittensor EVM
          batchWithdEth = [
            {
              tokenKey: ethTokenKey,
              to: user.address,
              amount: amount,
              nonce: 0,
              srcChainId: 945,
              tokenMetadata: createTokenMetadata("ETH", "Ether", 18),
            },
          ];
          balanceUserPreUnlock = await ethers.provider.getBalance(user.address);
          balanceBridgePreUnlock = await ethers.provider.getBalance(
            bridge.target
          );
        });
        it("Should not emit TokenWrapped event", async function () {
          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdEth)
          ).to.not.emit(bridge, "TokenWrapped");
        });
        it("Should emit TransferRequestExecuted event", async function () {
          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdEth)
          )
            .to.emit(bridge, "TransferRequestExecuted")
            .withArgs(0, 945);
        });
        it("Should increase user eth balance by amount", async function () {
          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdEth)
          ).to.not.be.reverted;

          await expect(
            ethers.provider.getBalance(user.address)
          ).to.eventually.equal(balanceUserPreUnlock + amount);
        });
        it("Should decrease bridge eth balance by amount", async function () {
          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdEth)
          ).to.not.be.reverted;
          await expect(
            ethers.provider.getBalance(bridge.target)
          ).to.eventually.equal(balanceBridgePreUnlock - amount);
        });
      });
      describe("Accounting Hosted ERC20 has unlock semantics", function () {
        const chainId = 945;
        let balanceUserPreUnlock: bigint;
        let balanceBridgePreUnlock: bigint;
        let batchWithdMatic: any[];

        beforeEach(async function () {
          await expect(
            bridge
              .connect(user)
              .requestTransfer(maticKey, user.address, maticAmount, chainId)
          ).not.to.be.reverted;

          // on the other chain user requests to bridge Back to Bittensor EVM
          balanceUserPreUnlock = await matic.balanceOf(user.address);
          balanceBridgePreUnlock = await matic.balanceOf(bridge.target);
          batchWithdMatic = [
            {
              tokenKey: maticKey,
              to: user.address,
              amount: maticAmount,
              nonce: 0,
              srcChainId: 945,
              tokenMetadata: createTokenMetadata("MATIC", "Polygon", 18),
            },
          ];
        });
        it("Should not emit TokenWrapped event", async function () {
          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdMatic)
          ).to.not.emit(bridge, "TokenWrapped");
        });
        it("Should emit TransferRequestExecuted event", async function () {
          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdMatic)
          )
            .to.emit(bridge, "TransferRequestExecuted")
            .withArgs(0, 945);
        });
        it("Should increase user matic balance by amount", async function () {
          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdMatic)
          ).to.not.be.reverted;
          await expect(matic.balanceOf(user.address)).to.eventually.equal(
            balanceUserPreUnlock + maticAmount
          );
        });
        it("Should decrease bridge matic balance by amount", async function () {
          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdMatic)
          ).to.not.be.reverted;
          await expect(matic.balanceOf(bridge.target)).to.eventually.equal(
            balanceBridgePreUnlock - maticAmount
          );
        });
      });
    });
  });
  describe("Transferring new tokens into destination chain with wrapping", function () {
    const amount = ethers.parseEther("100");
    let bridge: Bridge;
    let user: HardhatEthersSigner;
    let authority: HardhatEthersSigner;
    const taoKey = createTokenKey("TAO");
    let batchWithTao: any[];

    beforeEach(async function () {
      const fixture = await loadFixture(deployBridgeWithTokensFixture);
      bridge = fixture.bridge;
      user = fixture.user;
      authority = fixture.authority;

      batchWithTao = [
        {
          tokenKey: taoKey,
          to: user.address,
          amount: amount,
          nonce: 0,
          srcChainId: 945,
          tokenMetadata: createTokenMetadata("TAO", "Tao", 9),
        },
      ];
    });
    it("Should emit TokenWrapped event", async function () {
      await expect(
        bridge.connect(authority).executeTransferRequests(batchWithTao)
      )
        .to.emit(bridge, "TokenWrapped")
        .withArgs(taoKey, anyValue, true, ["Tao", "TAO", 9n]);
    });
    it("Should emit TransferRequestExecuted event", async function () {
      await expect(
        bridge.connect(authority).executeTransferRequests(batchWithTao)
      )
        .to.emit(bridge, "TransferRequestExecuted")
        .withArgs(0, 945);
    });
    it("transaction should work", async function () {
      await expect(
        bridge.connect(authority).executeTransferRequests(batchWithTao)
      ).to.not.be.reverted;
    });
    describe("Wrapped token should be:", function () {
      beforeEach(async function () {
        await expect(
          bridge.connect(authority).executeTransferRequests(batchWithTao)
        ).to.not.be.reverted;
      });
      it("added to the tokensInfo", async function () {
        const tokenInfo = await bridge.tokensInfo(taoKey);
        expect(tokenInfo.supported).to.be.equal(true);
        expect(tokenInfo.managed).to.be.equal(true);
        expect(tokenInfo.tokenAddress).to.not.be.equal(ethers.ZeroAddress);
      });
      describe("Properly accounted for with mint semantics", function () {
        let dTao: BridgedToken;
        beforeEach(async function () {
          const taoInfo = await bridge.tokensInfo(taoKey);
          const bridgeTokenFactory =
            await ethers.getContractFactory("BridgedToken");
          dTao = bridgeTokenFactory.attach(
            taoInfo.tokenAddress
          ) as BridgedToken;
        });
        it("should increase user tao balance by amount", async function () {
          await expect(dTao.balanceOf(user.address)).to.eventually.equal(
            amount
          );
        });
        it("should leave the bridge tao balance unchanged", async function () {
          await expect(dTao.balanceOf(bridge.target)).to.eventually.equal(0);
        });
      });
    });
  });
  describe("Transfer request with Wrapped Token", function () {
    let bridge: Bridge;
    let user: HardhatEthersSigner;
    let chainId: number;
    let authority: HardhatEthersSigner;
    const taoKey = createTokenKey("TAO");
    const amount = ethers.parseEther("100");
    let dTao: BridgedToken;
    let userBalancePre: bigint;
    let bridgeBalancePre: bigint;
    beforeEach(async function () {
      const fixture = await loadFixture(deployBridgeWithTokensFixture);
      bridge = fixture.bridge;
      user = fixture.user;
      authority = fixture.authority;
      chainId = 945;

      const batchWithTao = [
        {
          tokenKey: taoKey,
          to: user.address,
          amount: amount,
          nonce: 0,
          srcChainId: 945,
          tokenMetadata: createTokenMetadata("TAO", "Tao", 9),
        },
      ];

      // simulate incoming tao from Bittensor EVM
      await expect(
        bridge.connect(authority).executeTransferRequests(batchWithTao)
      ).to.not.be.reverted;

      const taoInfo = await bridge.tokensInfo(taoKey);
      const bridgeTokenFactory =
        await ethers.getContractFactory("BridgedToken");
      dTao = bridgeTokenFactory.attach(taoInfo.tokenAddress) as BridgedToken;
      userBalancePre = await dTao.balanceOf(user.address);
      bridgeBalancePre = await dTao.balanceOf(bridge.target);

      await dTao.connect(user).approve(bridge.target, amount);
      expect(await dTao.balanceOf(user.address)).to.equal(amount);
    });
    it("Should emit TransferRequested event with correct parameters", async function () {
      const nonce = 0;
      const request = [
        nonce,
        user.address,
        user.address,
        taoKey,
        amount,
        31337, // source chain id
        chainId,
      ];

      await expect(
        bridge
          .connect(user)
          .requestTransfer(taoKey, user.address, amount, chainId)
      )
        .to.emit(bridge, "TransferRequested")
        .withArgs(request);
    });

    describe("Should perform proper accounting actions with burn semantics", async function () {
      beforeEach(async function () {
        await expect(
          bridge
            .connect(user)
            .requestTransfer(taoKey, user.address, amount, chainId)
        ).to.not.be.reverted;
      });
      it("should decrease user tao balance by amount", async function () {
        await expect(dTao.balanceOf(user.address)).to.eventually.equal(
          userBalancePre - amount
        );
      });
      it("should leave bridge tao balance unchanged", async function () {
        await expect(dTao.balanceOf(bridge.target)).to.eventually.equal(
          bridgeBalancePre
        );
      });
    });
  });
  describe("Bridge with Stake Manager", function () {
    let bridge: Bridge;
    let admin: HardhatEthersSigner;
    let authority: HardhatEthersSigner;
    let stakeManager: StakeManager;
    let user: HardhatEthersSigner;
    let taoKey: string;
    let amount: bigint;
    const sourceChainId = 31337;
    const destinationChainId = 1000; // dummy value non relevant for this test
    let stakePrecompile: StakePrecompileMock;

    describe("Deployment With Stake Manager", function () {
      beforeEach(async function () {
        const fixture = await loadFixture(
          deployBridgeWithStakeManagerAndTokensFixture
        );
        bridge = fixture.bridge;
        admin = fixture.admin;
        stakeManager = fixture.stakeManager;
        user = fixture.user;
        stakePrecompile = fixture.stakePrecompile;
        taoKey = createTokenKey("TAO"); // simulate native transfers on Bittensor EVM
        amount = ethers.parseEther("100");

        await bridge
          .connect(admin)
          .whitelistToken(taoKey, true, ethers.ZeroAddress, "TAO", "Tao", 9);
      });

      it("Should set staking manager correctly", async function () {
        expect(await bridge.stakingManager()).to.equal(stakeManager.target);
      });

      it("Should set the admin correctly on the stake manager", async function () {
        expect(await stakeManager.owner()).to.equal(admin.address);
      });

      it("Should set datura hotkey correctly on the stake manager", async function () {
        const daturaHotkey = ethers.keccak256(ethers.toUtf8Bytes("datura"));
        expect(await stakeManager.daturaStakingHotkey()).to.equal(daturaHotkey);
      });
      it("Non admin should not be able to add reward recipients", async function () {
        await expect(
          stakeManager.connect(user).addBridgeParticipantReward(user.address)
        )
          .to.be.revertedWithCustomError(
            stakeManager,
            "OwnableUnauthorizedAccount"
          )
          .withArgs(user.address);
      });
      it("Successful add reward recipient", async function () {
        await expect(
          stakeManager.connect(admin).addBridgeParticipantReward(user.address)
        ).to.not.be.reverted;
      });
      it("Successful remove reward recipient", async function () {
        await expect(
          stakeManager.connect(admin).addBridgeParticipantReward(user.address)
        ).to.not.be.reverted;
        await expect(
          stakeManager
            .connect(admin)
            .removeBridgeParticipantReward(user.address)
        ).to.not.be.reverted;
      });
      it("Should not be able to remove non existent reward recipient", async function () {
        await expect(
          stakeManager
            .connect(admin)
            .removeBridgeParticipantReward(user.address)
        ).to.be.revertedWithCustomError(
          stakeManager,
          "RewardRecipientNotFound"
        );
      });
      it("Should not be able to add address 0 as reward recipient", async function () {
        await expect(
          stakeManager
            .connect(admin)
            .addBridgeParticipantReward(ethers.ZeroAddress)
        ).to.be.revertedWithCustomError(stakeManager, "InvalidAddress");
      });
      it("Should not be able to add reward recipient twice", async function () {
        await expect(
          stakeManager.connect(admin).addBridgeParticipantReward(user.address)
        ).to.not.be.reverted;
        await expect(
          stakeManager.connect(admin).addBridgeParticipantReward(user.address)
        ).to.be.revertedWithCustomError(
          stakeManager,
          "RewardRecipientAlreadyExists"
        );
      });
      it("Non admin should not be able to remove reward recipient", async function () {
        await stakeManager
          .connect(admin)
          .addBridgeParticipantReward(user.address);
        await expect(
          stakeManager.connect(user).removeBridgeParticipantReward(user.address)
        )
          .to.be.revertedWithCustomError(
            stakeManager,
            "OwnableUnauthorizedAccount"
          )
          .withArgs(user.address);
      });
    });
    describe("Requesting transfers with stake manager", function () {
      beforeEach(async function () {
        const fixture = await loadFixture(
          deployBridgeWithStakeManagerAndTokensFixture
        );
        bridge = fixture.bridge;
        admin = fixture.admin;
        stakeManager = fixture.stakeManager;
        user = fixture.user;
        taoKey = createTokenKey("TAO"); // simulate native transfers on Bittensor EVM
        amount = ethers.parseEther("100");

        await bridge
          .connect(admin)
          .whitelistToken(taoKey, true, ethers.ZeroAddress, "TAO", "Tao", 9);
      });
      it("Should emit TransferRequested event with correct parameters", async function () {
        const nonce = 0;
        const request = [
          nonce,
          user.address,
          user.address,
          taoKey,
          amount,
          sourceChainId,
          destinationChainId,
        ];
        await expect(
          bridge
            .connect(user)
            .requestTransfer(taoKey, user.address, amount, destinationChainId, {
              value: amount,
            })
        )
          .to.emit(bridge, "TransferRequested")
          .withArgs(request);
      });
      it("Should emit Staked event", async function () {
        await expect(
          bridge
            .connect(user)
            .requestTransfer(taoKey, user.address, amount, destinationChainId, {
              value: amount,
            })
        )
          .to.emit(stakeManager, "TokensStaked")
          .withArgs(user.address, amount);
      });
      it("Should route funds to the staking manager", async function () {
        await expect(
          bridge
            .connect(user)
            .requestTransfer(taoKey, user.address, amount, destinationChainId, {
              value: amount,
            })
        ).to.not.be.reverted;

        await expect(
          ethers.provider.getBalance(stakeManager.target)
        ).to.eventually.equal(amount);
      });
      it("More than target stakes during interval are allowed", async function () {
        const stakeManagerBalancePre = await ethers.provider.getBalance(
          stakeManager.target
        );
        for (let i = 0; i < 10; i++) {
          await expect(
            bridge
              .connect(user)
              .requestTransfer(
                taoKey,
                user.address,
                amount,
                destinationChainId,
                {
                  value: amount,
                }
              )
          ).to.not.be.reverted;
        }

        const stakePosition = await stakeManager.stakes(user.address);
        expect(stakePosition.amount).to.be.equal(amount * 10n);
        expect(
          await ethers.provider.getBalance(stakeManager.target)
        ).to.be.equal(stakeManagerBalancePre + amount * 10n);
      });
      describe("Staking to precompile after interval", function () {
        let stakePrecompileBalancePre: bigint;
        let blocksToNextStakingBlock: bigint;

        beforeEach(async function () {
          await expect(
            bridge
              .connect(user)
              .requestTransfer(
                taoKey,
                user.address,
                amount,
                destinationChainId,
                {
                  value: amount,
                }
              )
          ).to.not.be.reverted;

          stakePrecompileBalancePre = await ethers.provider.getBalance(
            stakePrecompile.target
          );

          const stakeInterval = await stakeManager.STAKE_INTERVAL();
          const lastBlockStaked = await stakeManager.lastStakingBlock();
          blocksToNextStakingBlock =
            lastBlockStaked +
            stakeInterval -
            BigInt(await ethers.provider.getBlockNumber());

          await mine(Number(blocksToNextStakingBlock));
        });

        it("Should emit FundsStakedOnPrecompile event when staking to precompile", async function () {
          await expect(stakeManager.connect(user).stakeToPrecompile()).to.emit(
            stakeManager,
            "FundsStakedOnPrecompile"
          );
        });

        it("Should increase precompile balance by staked amount", async function () {
          await stakeManager.connect(user).stakeToPrecompile();
          const stakePrecompileBalancePost = await ethers.provider.getBalance(
            stakePrecompile.target
          );
          expect(stakePrecompileBalancePost).to.be.equal(
            stakePrecompileBalancePre + amount
          );
        });

        it("Should set correct staking epoch ID for user stake", async function () {
          await stakeManager.connect(user).stakeToPrecompile();
          const stakePosition = await stakeManager.stakes(user.address);
          const stakingEpochId = await stakeManager.nextStakingEpochId();
          expect(stakePosition.stakingEpochId).to.be.equal(stakingEpochId - 1n);
        });

        it("Should record correct staking block for epoch", async function () {
          const currentBlock = await ethers.provider.getBlockNumber();
          await stakeManager.connect(user).stakeToPrecompile(); // this next executed in currentBlock + 1
          const stakingEpochId = await stakeManager.nextStakingEpochId();
          const stakeEpochStartBlock =
            await stakeManager.stakingEpochIdToLastStakingBlock(
              stakingEpochId - 1n
            );
          expect(stakeEpochStartBlock).to.be.equal(currentBlock + 1);
        });
      });
    });
    describe("Unstaking with stake manager before subtensor stake interval", function () {
      let batchWithdrawTao: any[];
      let startBlock: number;
      const taoKey = createTokenKey("TAO"); // simulate native transfers on Bittensor EVM
      let stakePrecompileBalancePre: bigint;
      const amount = ethers.parseEther("100");
      const amountToUnstake = ethers.parseEther("10");

      beforeEach(async function () {
        const fixture = await loadFixture(
          deployBridgeWithStakeManagerAndTokensFixture
        );
        bridge = fixture.bridge;
        admin = fixture.admin;
        stakeManager = fixture.stakeManager;
        user = fixture.user;
        stakePrecompile = fixture.stakePrecompile;
        authority = fixture.authority;

        // user first requests to transfer tao and stakes it
        await bridge
          .connect(admin)
          .whitelistToken(taoKey, true, ethers.ZeroAddress, "TAO", "Tao", 9);

        await expect(
          bridge
            .connect(user)
            .requestTransfer(taoKey, user.address, amount, destinationChainId, {
              value: amount,
            })
        ).to.not.be.reverted;

        batchWithdrawTao = [
          {
            tokenKey: taoKey,
            to: user.address,
            amount: amountToUnstake,
            nonce: 0,
            srcChainId: 945,
            tokenMetadata: createTokenMetadata("TAO", "Tao", 9),
          },
        ];
        stakePrecompileBalancePre = await ethers.provider.getBalance(
          stakePrecompile.target
        );
      });
      it("Should be able to unstake without rewards", async function () {
        await mine(10);
        const stakeManagerBalancePre = await ethers.provider.getBalance(
          stakeManager.target
        );
        await expect(
          bridge.connect(authority).executeTransferRequests(batchWithdrawTao)
        ).to.not.be.reverted;

        await expect(
          ethers.provider.getBalance(stakeManager.target)
        ).to.eventually.equal(stakeManagerBalancePre - amountToUnstake);
      });
    });
    describe("Unstaking with stake manager after at least subtensor stake interval", function () {
      let batchWithdrawTao: any[];
      let stakePrecompileBalancePre: bigint;
      let stakeBlock: number;
      let externalPartner: HardhatEthersSigner;
      const numberOfBlocksToMine = 1000; // keep calculations simple
      // This function assumes that the unstake function is called immediately after,
      const mineAndIncreaseStakePrecompileBalance = async (
        numberOfBlocks: number
      ) => {
        const initialAmount = await ethers.provider.getBalance(
          stakePrecompile.target
        );
        await mine(numberOfBlocks);
        const endBlock = await ethers.provider.getBlockNumber();
        const blocksPassed = endBlock - stakeBlock + 1; // this accounts that the unstake function is called immediately after (+1 block)
        const normalizer = BigInt(await stakeManager.NORMALIZER());
        const rewardRate = await stakeManager.rewardRate();

        const num = initialAmount * rewardRate * BigInt(blocksPassed);
        const reward = num / normalizer;
        await admin.sendTransaction({
          to: stakePrecompile.target,
          value: reward,
        });

        stakePrecompileBalancePre = await ethers.provider.getBalance(
          stakePrecompile.target
        );
        return reward;
      };
      const taoKey = createTokenKey("TAO"); // simulate native transfers on Bittensor EVM

      const amount = ethers.parseEther("100");
      beforeEach(async function () {
        const fixture = await loadFixture(
          deployBridgeWithStakeManagerAndTokensFixture
        );
        bridge = fixture.bridge;
        admin = fixture.admin;
        stakeManager = fixture.stakeManager;
        user = fixture.user;
        stakePrecompile = fixture.stakePrecompile;
        authority = fixture.authority;

        // user first requests to transfer tao and stakes it
        await bridge
          .connect(admin)
          .whitelistToken(taoKey, true, ethers.ZeroAddress, "TAO", "Tao", 9);

        await expect(
          bridge
            .connect(user)
            .requestTransfer(taoKey, user.address, amount, destinationChainId, {
              value: amount,
            })
        ).to.not.be.reverted;

        const stakeInterval = await stakeManager.STAKE_INTERVAL();
        const lastBlockStaked = await stakeManager.lastStakingBlock();
        const blockInterval =
          lastBlockStaked +
          stakeInterval -
          BigInt(await ethers.provider.getBlockNumber()) -
          1n;
        await mine(Number(blockInterval)); // mine up until lastBlockStaked + INTERVAL - 1
        await expect(stakeManager.connect(user).stakeToPrecompile()).to.not.be
          .reverted; // stake at lastBlockStaked + INTERVAL
        stakeBlock = await ethers.provider.getBlockNumber();

        batchWithdrawTao = [
          {
            tokenKey: taoKey,
            to: user.address,
            amount: amount,
            nonce: 0,
            srcChainId: 945,
            tokenMetadata: createTokenMetadata("TAO", "Tao", 9),
          },
        ];
      });
      describe("When participants are not set", function () {
        it("Should decrease staking precompile balance by amount + rewards", async function () {
          const reward =
            await mineAndIncreaseStakePrecompileBalance(numberOfBlocksToMine);

          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdrawTao)
          ).to.not.be.reverted;

          await expect(
            ethers.provider.getBalance(stakePrecompile.target)
          ).to.eventually.equal(stakePrecompileBalancePre - amount - reward);
        });
        it("Should emit Unstaked event", async function () {
          await mineAndIncreaseStakePrecompileBalance(numberOfBlocksToMine);

          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdrawTao)
          )
            .to.emit(stakeManager, "TokensUnstaked")
            .withArgs(user.address, amount);
        });
        it("Should reward user with amount + rewards", async function () {
          const reward =
            await mineAndIncreaseStakePrecompileBalance(numberOfBlocksToMine);
          const balancePre = await ethers.provider.getBalance(user.address);

          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdrawTao)
          ).to.not.be.reverted;

          await expect(
            ethers.provider.getBalance(user.address)
          ).to.eventually.equal(balancePre + amount + reward);
        });
      });
      describe("When participants are set", function () {
        beforeEach(async function () {
          // add admin and external partner as reward recipients
          const accounts = await ethers.getSigners();
          externalPartner = accounts[8];
          await expect(
            stakeManager
              .connect(admin)
              .addBridgeParticipantReward(admin.address)
          ).to.not.be.reverted;
          await expect(
            stakeManager
              .connect(admin)
              .addBridgeParticipantReward(externalPartner.address)
          ).to.not.be.reverted;
        });
        it("Should decrease staking precompile balance by amount + rewards", async function () {
          const reward =
            await mineAndIncreaseStakePrecompileBalance(numberOfBlocksToMine);

          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdrawTao)
          ).to.not.be.reverted;

          await expect(
            ethers.provider.getBalance(stakePrecompile.target)
          ).to.eventually.equal(stakePrecompileBalancePre - amount - reward);
        });
        it("Should emit Unstaked event", async function () {
          await mineAndIncreaseStakePrecompileBalance(numberOfBlocksToMine);

          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdrawTao)
          )
            .to.emit(stakeManager, "TokensUnstaked")
            .withArgs(user.address, amount);
        });
        it("Should reward user with amount + rewards/2", async function () {
          const reward =
            await mineAndIncreaseStakePrecompileBalance(numberOfBlocksToMine);
          const balancePre = await ethers.provider.getBalance(user.address);
          const rewardToParticipants = reward / 4n;
          const rewardToUser = reward - 2n * rewardToParticipants;

          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdrawTao)
          ).to.not.be.reverted;

          await expect(
            ethers.provider.getBalance(user.address)
          ).to.eventually.equal(balancePre + amount + rewardToUser);
        });
        it("Should reward participants with rewards / 2 divided by number of participants", async function () {
          const reward =
            await mineAndIncreaseStakePrecompileBalance(numberOfBlocksToMine);
          const rewardPerParticipant = reward / 4n; // 25% of the reward (50% to user, 25% to participants)
          const adminBalancePre = await ethers.provider.getBalance(
            admin.address
          );
          const externalPartnerBalancePre = await ethers.provider.getBalance(
            externalPartner.address
          );

          await expect(
            bridge.connect(authority).executeTransferRequests(batchWithdrawTao)
          ).to.not.be.reverted;

          await expect(
            ethers.provider.getBalance(externalPartner.address)
          ).to.eventually.equal(
            externalPartnerBalancePre + rewardPerParticipant
          );
          await expect(
            ethers.provider.getBalance(admin.address)
          ).to.eventually.equal(adminBalancePre + rewardPerParticipant);
        });
      });
    });
  });
});
