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
    Initialized: (0, evm_abi_1.event)("0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2", "Initialized(uint64)", { "version": p.uint64 }),
    NewTokenWhitelisted: (0, evm_abi_1.event)("0x80150666a36e18684118023c777f2647e6b371e01dc410a29550745a709117c6", "NewTokenWhitelisted(bytes32,(address,bool,bool,bool),(string,string,uint8))", { "tokenKey": (0, evm_abi_1.indexed)(p.bytes32), "tokenInfo": p.struct({ "tokenAddress": p.address, "managed": p.bool, "enabled": p.bool, "supported": p.bool }), "tokenMetadata": p.struct({ "name": p.string, "symbol": p.string, "decimals": p.uint8 }) }),
    Paused: (0, evm_abi_1.event)("0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258", "Paused(address)", { "account": p.address }),
    RoleAdminChanged: (0, evm_abi_1.event)("0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff", "RoleAdminChanged(bytes32,bytes32,bytes32)", { "role": (0, evm_abi_1.indexed)(p.bytes32), "previousAdminRole": (0, evm_abi_1.indexed)(p.bytes32), "newAdminRole": (0, evm_abi_1.indexed)(p.bytes32) }),
    RoleGranted: (0, evm_abi_1.event)("0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d", "RoleGranted(bytes32,address,address)", { "role": (0, evm_abi_1.indexed)(p.bytes32), "account": (0, evm_abi_1.indexed)(p.address), "sender": (0, evm_abi_1.indexed)(p.address) }),
    RoleRevoked: (0, evm_abi_1.event)("0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b", "RoleRevoked(bytes32,address,address)", { "role": (0, evm_abi_1.indexed)(p.bytes32), "account": (0, evm_abi_1.indexed)(p.address), "sender": (0, evm_abi_1.indexed)(p.address) }),
    TokenWhitelistStatusUpdated: (0, evm_abi_1.event)("0xa158d459581858b2eda49d61c909897cae2b98e26fb0bacc1f6f89d42a372b16", "TokenWhitelistStatusUpdated(bytes32,(address,bool,bool,bool))", { "tokenKey": (0, evm_abi_1.indexed)(p.bytes32), "tokenInfo": p.struct({ "tokenAddress": p.address, "managed": p.bool, "enabled": p.bool, "supported": p.bool }) }),
    TokenWrapped: (0, evm_abi_1.event)("0x8350ea9d2ac1a64b7990b6abfefffa3e2fc4e00730d262032f8c555eb56da066", "TokenWrapped(bytes32,address,bool,(string,string,uint8))", { "tokenKey": (0, evm_abi_1.indexed)(p.bytes32), "wrappedToken": (0, evm_abi_1.indexed)(p.address), "isWhitelisted": (0, evm_abi_1.indexed)(p.bool), "tokenMetadata": p.struct({ "name": p.string, "symbol": p.string, "decimals": p.uint8 }) }),
    TransferRequestExecuted: (0, evm_abi_1.event)("0xcbc66fcdc0aaadd449255e15a56335fbcc6fcdcf3987c59bf5591c1636d84ce2", "TransferRequestExecuted(uint256,uint64)", { "nonce": p.uint256, "srcChainId": p.uint64 }),
    TransferRequested: (0, evm_abi_1.event)("0x3f9b5dddbaec56df245d81eed98595f8f27d3dca9554dfd6f8fe646aeb76bac6", "TransferRequested((uint256,address,address,bytes32,uint256,uint64,uint64))", { "request": p.struct({ "nonce": p.uint256, "from": p.address, "to": p.address, "tokenKey": p.bytes32, "amount": p.uint256, "srcChainId": p.uint64, "destChainId": p.uint64 }) }),
    Unpaused: (0, evm_abi_1.event)("0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa", "Unpaused(address)", { "account": p.address }),
    Upgraded: (0, evm_abi_1.event)("0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b", "Upgraded(address)", { "implementation": (0, evm_abi_1.indexed)(p.address) }),
};
exports.functions = {
    AUTHORITY_ROLE: (0, evm_abi_1.viewFun)("0x4a3fba0e", "AUTHORITY_ROLE()", {}, p.bytes32),
    DEFAULT_ADMIN_ROLE: (0, evm_abi_1.viewFun)("0xa217fddf", "DEFAULT_ADMIN_ROLE()", {}, p.bytes32),
    MAX_CONFIRMATION_LENGTH: (0, evm_abi_1.viewFun)("0x46a0dd5c", "MAX_CONFIRMATION_LENGTH()", {}, p.uint256),
    PAUSER_ROLE: (0, evm_abi_1.viewFun)("0xe63ab1e9", "PAUSER_ROLE()", {}, p.bytes32),
    UPGRADE_INTERFACE_VERSION: (0, evm_abi_1.viewFun)("0xad3cb1cc", "UPGRADE_INTERFACE_VERSION()", {}, p.string),
    bridgeNonce: (0, evm_abi_1.viewFun)("0x1f87a250", "bridgeNonce()", {}, p.uint256),
    executeTransferRequests: (0, evm_abi_1.fun)("0x640c89ee", "executeTransferRequests((address,uint256,uint256,bytes32,uint64,(string,string,uint8))[])", { "batch": p.array(p.struct({ "to": p.address, "amount": p.uint256, "nonce": p.uint256, "tokenKey": p.bytes32, "srcChainId": p.uint64, "tokenMetadata": p.struct({ "name": p.string, "symbol": p.string, "decimals": p.uint8 }) })) }),
    getAdmin: (0, evm_abi_1.viewFun)("0x6e9960c3", "getAdmin()", {}, p.address),
    getAuthority: (0, evm_abi_1.viewFun)("0xe2b178a0", "getAuthority()", {}, p.address),
    getRoleAdmin: (0, evm_abi_1.viewFun)("0x248a9ca3", "getRoleAdmin(bytes32)", { "role": p.bytes32 }, p.bytes32),
    grantRole: (0, evm_abi_1.fun)("0x2f2ff15d", "grantRole(bytes32,address)", { "role": p.bytes32, "account": p.address }),
    hasRole: (0, evm_abi_1.viewFun)("0x91d14854", "hasRole(bytes32,address)", { "role": p.bytes32, "account": p.address }, p.bool),
    initialize: (0, evm_abi_1.fun)("0x485cc955", "initialize(address,address)", { "_authority": p.address, "_admin": p.address }),
    isPaused: (0, evm_abi_1.viewFun)("0xb187bd26", "isPaused()", {}, p.bool),
    metadataForToken: (0, evm_abi_1.viewFun)("0x770eef9e", "metadataForToken(bytes32)", { "_0": p.bytes32 }, { "name": p.string, "symbol": p.string, "decimals": p.uint8 }),
    pause: (0, evm_abi_1.fun)("0x8456cb59", "pause()", {}),
    paused: (0, evm_abi_1.viewFun)("0x5c975abb", "paused()", {}, p.bool),
    processedTransfers: (0, evm_abi_1.viewFun)("0x5e33e5fd", "processedTransfers(uint64,uint256)", { "_0": p.uint64, "_1": p.uint256 }, p.bool),
    proxiableUUID: (0, evm_abi_1.viewFun)("0x52d1902d", "proxiableUUID()", {}, p.bytes32),
    renounceRole: (0, evm_abi_1.fun)("0x36568abe", "renounceRole(bytes32,address)", { "role": p.bytes32, "callerConfirmation": p.address }),
    requestTransfer: (0, evm_abi_1.fun)("0x6c0652fb", "requestTransfer(bytes32,address,uint256,uint64)", { "tokenKey": p.bytes32, "to": p.address, "amount": p.uint256, "destChainId": p.uint64 }),
    revokeRole: (0, evm_abi_1.fun)("0xd547741f", "revokeRole(bytes32,address)", { "role": p.bytes32, "account": p.address }),
    setAuthority: (0, evm_abi_1.fun)("0x7a9e5e4b", "setAuthority(address)", { "_authority": p.address }),
    setStakingManager: (0, evm_abi_1.fun)("0xb00bba6a", "setStakingManager(address)", { "_stakingManager": p.address }),
    stakingManager: (0, evm_abi_1.viewFun)("0x22828cc2", "stakingManager()", {}, p.address),
    supportsInterface: (0, evm_abi_1.viewFun)("0x01ffc9a7", "supportsInterface(bytes4)", { "interfaceId": p.bytes4 }, p.bool),
    tokensInfo: (0, evm_abi_1.viewFun)("0xf0170150", "tokensInfo(bytes32)", { "_0": p.bytes32 }, { "tokenAddress": p.address, "managed": p.bool, "enabled": p.bool, "supported": p.bool }),
    transferRequests: (0, evm_abi_1.viewFun)("0x03486c40", "transferRequests(uint256)", { "_0": p.uint256 }, { "nonce": p.uint256, "from": p.address, "to": p.address, "tokenKey": p.bytes32, "amount": p.uint256, "srcChainId": p.uint64, "destChainId": p.uint64 }),
    unpause: (0, evm_abi_1.fun)("0x3f4ba83a", "unpause()", {}),
    updateWhitelistStatus: (0, evm_abi_1.fun)("0x3cdf146e", "updateWhitelistStatus(bytes32,bool,address)", { "tokenKey": p.bytes32, "enabled": p.bool, "tokenAddress": p.address }),
    upgradeToAndCall: (0, evm_abi_1.fun)("0x4f1ef286", "upgradeToAndCall(address,bytes)", { "newImplementation": p.address, "data": p.bytes }),
    whitelistToken: (0, evm_abi_1.fun)("0x3227247a", "whitelistToken(bytes32,bool,address,string,string,uint8)", { "tokenKey": p.bytes32, "enabled": p.bool, "tokenAddress": p.address, "tokenSymbol": p.string, "tokenName": p.string, "tokenDecimals": p.uint8 }),
};
class Contract extends evm_abi_1.ContractBase {
    AUTHORITY_ROLE() {
        return this.eth_call(exports.functions.AUTHORITY_ROLE, {});
    }
    DEFAULT_ADMIN_ROLE() {
        return this.eth_call(exports.functions.DEFAULT_ADMIN_ROLE, {});
    }
    MAX_CONFIRMATION_LENGTH() {
        return this.eth_call(exports.functions.MAX_CONFIRMATION_LENGTH, {});
    }
    PAUSER_ROLE() {
        return this.eth_call(exports.functions.PAUSER_ROLE, {});
    }
    UPGRADE_INTERFACE_VERSION() {
        return this.eth_call(exports.functions.UPGRADE_INTERFACE_VERSION, {});
    }
    bridgeNonce() {
        return this.eth_call(exports.functions.bridgeNonce, {});
    }
    getAdmin() {
        return this.eth_call(exports.functions.getAdmin, {});
    }
    getAuthority() {
        return this.eth_call(exports.functions.getAuthority, {});
    }
    getRoleAdmin(role) {
        return this.eth_call(exports.functions.getRoleAdmin, { role });
    }
    hasRole(role, account) {
        return this.eth_call(exports.functions.hasRole, { role, account });
    }
    isPaused() {
        return this.eth_call(exports.functions.isPaused, {});
    }
    metadataForToken(_0) {
        return this.eth_call(exports.functions.metadataForToken, { _0 });
    }
    paused() {
        return this.eth_call(exports.functions.paused, {});
    }
    processedTransfers(_0, _1) {
        return this.eth_call(exports.functions.processedTransfers, { _0, _1 });
    }
    proxiableUUID() {
        return this.eth_call(exports.functions.proxiableUUID, {});
    }
    stakingManager() {
        return this.eth_call(exports.functions.stakingManager, {});
    }
    supportsInterface(interfaceId) {
        return this.eth_call(exports.functions.supportsInterface, { interfaceId });
    }
    tokensInfo(_0) {
        return this.eth_call(exports.functions.tokensInfo, { _0 });
    }
    transferRequests(_0) {
        return this.eth_call(exports.functions.transferRequests, { _0 });
    }
}
exports.Contract = Contract;
//# sourceMappingURL=Bridge.sol.js.map