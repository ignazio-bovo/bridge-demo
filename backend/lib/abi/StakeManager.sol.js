"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.Contract = exports.functions = exports.events = void 0;
const p = __importStar(require("@subsquid/evm-codec"));
const evm_abi_1 = require("@subsquid/evm-abi");
exports.events = {
    FundsStakedOnPrecompile: (0, evm_abi_1.event)("0xcb35838c2d3b529d682bc65aac48be02411f8eb3b9146368db6feead7e1f6e2c", "FundsStakedOnPrecompile(uint256,uint64,uint64)", { "amount": p.uint256, "lastStakingBlock": p.uint64, "stakingEpochId": p.uint64 }),
    Initialized: (0, evm_abi_1.event)("0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2", "Initialized(uint64)", { "version": p.uint64 }),
    OwnershipTransferred: (0, evm_abi_1.event)("0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0", "OwnershipTransferred(address,address)", { "previousOwner": (0, evm_abi_1.indexed)(p.address), "newOwner": (0, evm_abi_1.indexed)(p.address) }),
    TokensStaked: (0, evm_abi_1.event)("0xb539ca1e5c8d398ddf1c41c30166f33404941683be4683319b57669a93dad4ef", "TokensStaked(address,uint256)", { "user": (0, evm_abi_1.indexed)(p.address), "amount": p.uint256 }),
    TokensUnstaked: (0, evm_abi_1.event)("0x9845e367b683334e5c0b12d7b81721ac518e649376fa65e3d68324e8f34f2679", "TokensUnstaked(address,uint256)", { "user": (0, evm_abi_1.indexed)(p.address), "amount": p.uint256 }),
    Upgraded: (0, evm_abi_1.event)("0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b", "Upgraded(address)", { "implementation": (0, evm_abi_1.indexed)(p.address) }),
};
exports.functions = {
    NORMALIZER: (0, evm_abi_1.viewFun)("0x07e72f45", "NORMALIZER()", {}, p.uint64),
    ROOT_NET_UID: (0, evm_abi_1.viewFun)("0x230cde7b", "ROOT_NET_UID()", {}, p.uint16),
    STAKE_INTERVAL: (0, evm_abi_1.viewFun)("0xec8b70d1", "STAKE_INTERVAL()", {}, p.uint64),
    UPGRADE_INTERFACE_VERSION: (0, evm_abi_1.viewFun)("0xad3cb1cc", "UPGRADE_INTERFACE_VERSION()", {}, p.string),
    addBridgeParticipantReward: (0, evm_abi_1.fun)("0xb12f7998", "addBridgeParticipantReward(address)", { "rewardAddress": p.address }),
    bridge: (0, evm_abi_1.viewFun)("0xe78cea92", "bridge()", {}, p.address),
    daturaStakingHotkey: (0, evm_abi_1.viewFun)("0xd544f674", "daturaStakingHotkey()", {}, p.bytes32),
    initialize: (0, evm_abi_1.fun)("0x294e8d0e", "initialize(address,uint256,bytes32,address)", { "_bridge": p.address, "_rewardRate": p.uint256, "_daturaHotkey": p.bytes32, "_stakePrecompile": p.address }),
    lastStakingBlock: (0, evm_abi_1.viewFun)("0x44f98990", "lastStakingBlock()", {}, p.uint64),
    nextStakingEpochId: (0, evm_abi_1.viewFun)("0xc347e4e8", "nextStakingEpochId()", {}, p.uint64),
    owner: (0, evm_abi_1.viewFun)("0x8da5cb5b", "owner()", {}, p.address),
    proxiableUUID: (0, evm_abi_1.viewFun)("0x52d1902d", "proxiableUUID()", {}, p.bytes32),
    removeBridgeParticipantReward: (0, evm_abi_1.fun)("0xf82b45be", "removeBridgeParticipantReward(address)", { "rewardAddress": p.address }),
    renounceOwnership: (0, evm_abi_1.fun)("0x715018a6", "renounceOwnership()", {}),
    rewardRate: (0, evm_abi_1.viewFun)("0x7b0a47ee", "rewardRate()", {}, p.uint256),
    stake: (0, evm_abi_1.fun)("0x7acb7757", "stake(uint256,address)", { "amount": p.uint256, "user": p.address }),
    stakePrecompile: (0, evm_abi_1.viewFun)("0xaf03528b", "stakePrecompile()", {}, p.address),
    stakeToPrecompile: (0, evm_abi_1.fun)("0xee8aae30", "stakeToPrecompile()", {}),
    stakes: (0, evm_abi_1.viewFun)("0x16934fc4", "stakes(address)", { "_0": p.address }, { "amount": p.uint256, "stakingEpochId": p.uint64 }),
    stakingEpochIdToLastStakingBlock: (0, evm_abi_1.viewFun)("0x65f5f670", "stakingEpochIdToLastStakingBlock(uint64)", { "_0": p.uint64 }, p.uint64),
    transferOwnership: (0, evm_abi_1.fun)("0xf2fde38b", "transferOwnership(address)", { "newOwner": p.address }),
    unstake: (0, evm_abi_1.fun)("0x8381e182", "unstake(uint256,address)", { "amount": p.uint256, "user": p.address }),
    upgradeToAndCall: (0, evm_abi_1.fun)("0x4f1ef286", "upgradeToAndCall(address,bytes)", { "newImplementation": p.address, "data": p.bytes }),
};
class Contract extends evm_abi_1.ContractBase {
    NORMALIZER() {
        return this.eth_call(exports.functions.NORMALIZER, {});
    }
    ROOT_NET_UID() {
        return this.eth_call(exports.functions.ROOT_NET_UID, {});
    }
    STAKE_INTERVAL() {
        return this.eth_call(exports.functions.STAKE_INTERVAL, {});
    }
    UPGRADE_INTERFACE_VERSION() {
        return this.eth_call(exports.functions.UPGRADE_INTERFACE_VERSION, {});
    }
    bridge() {
        return this.eth_call(exports.functions.bridge, {});
    }
    daturaStakingHotkey() {
        return this.eth_call(exports.functions.daturaStakingHotkey, {});
    }
    lastStakingBlock() {
        return this.eth_call(exports.functions.lastStakingBlock, {});
    }
    nextStakingEpochId() {
        return this.eth_call(exports.functions.nextStakingEpochId, {});
    }
    owner() {
        return this.eth_call(exports.functions.owner, {});
    }
    proxiableUUID() {
        return this.eth_call(exports.functions.proxiableUUID, {});
    }
    rewardRate() {
        return this.eth_call(exports.functions.rewardRate, {});
    }
    stakePrecompile() {
        return this.eth_call(exports.functions.stakePrecompile, {});
    }
    stakes(_0) {
        return this.eth_call(exports.functions.stakes, { _0 });
    }
    stakingEpochIdToLastStakingBlock(_0) {
        return this.eth_call(exports.functions.stakingEpochIdToLastStakingBlock, { _0 });
    }
}
exports.Contract = Contract;
//# sourceMappingURL=StakeManager.sol.js.map