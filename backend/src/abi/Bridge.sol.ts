import * as p from '@subsquid/evm-codec'
import { event, fun, viewFun, indexed, ContractBase } from '@subsquid/evm-abi'
import type { EventParams as EParams, FunctionArguments, FunctionReturn } from '@subsquid/evm-abi'

export const events = {
    Initialized: event("0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2", "Initialized(uint64)", {"version": p.uint64}),
    NewTokenWhitelisted: event("0x80150666a36e18684118023c777f2647e6b371e01dc410a29550745a709117c6", "NewTokenWhitelisted(bytes32,(address,bool,bool,bool),(string,string,uint8))", {"tokenKey": indexed(p.bytes32), "tokenInfo": p.struct({"tokenAddress": p.address, "managed": p.bool, "enabled": p.bool, "supported": p.bool}), "tokenMetadata": p.struct({"name": p.string, "symbol": p.string, "decimals": p.uint8})}),
    Paused: event("0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258", "Paused(address)", {"account": p.address}),
    RoleAdminChanged: event("0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff", "RoleAdminChanged(bytes32,bytes32,bytes32)", {"role": indexed(p.bytes32), "previousAdminRole": indexed(p.bytes32), "newAdminRole": indexed(p.bytes32)}),
    RoleGranted: event("0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d", "RoleGranted(bytes32,address,address)", {"role": indexed(p.bytes32), "account": indexed(p.address), "sender": indexed(p.address)}),
    RoleRevoked: event("0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b", "RoleRevoked(bytes32,address,address)", {"role": indexed(p.bytes32), "account": indexed(p.address), "sender": indexed(p.address)}),
    TokenWhitelistStatusUpdated: event("0xa158d459581858b2eda49d61c909897cae2b98e26fb0bacc1f6f89d42a372b16", "TokenWhitelistStatusUpdated(bytes32,(address,bool,bool,bool))", {"tokenKey": indexed(p.bytes32), "tokenInfo": p.struct({"tokenAddress": p.address, "managed": p.bool, "enabled": p.bool, "supported": p.bool})}),
    TokenWrapped: event("0x8350ea9d2ac1a64b7990b6abfefffa3e2fc4e00730d262032f8c555eb56da066", "TokenWrapped(bytes32,address,bool,(string,string,uint8))", {"tokenKey": indexed(p.bytes32), "wrappedToken": indexed(p.address), "isWhitelisted": indexed(p.bool), "tokenMetadata": p.struct({"name": p.string, "symbol": p.string, "decimals": p.uint8})}),
    TransferRequestExecuted: event("0xcbc66fcdc0aaadd449255e15a56335fbcc6fcdcf3987c59bf5591c1636d84ce2", "TransferRequestExecuted(uint256,uint64)", {"nonce": p.uint256, "srcChainId": p.uint64}),
    TransferRequested: event("0x3f9b5dddbaec56df245d81eed98595f8f27d3dca9554dfd6f8fe646aeb76bac6", "TransferRequested((uint256,address,address,bytes32,uint256,uint64,uint64))", {"request": p.struct({"nonce": p.uint256, "from": p.address, "to": p.address, "tokenKey": p.bytes32, "amount": p.uint256, "srcChainId": p.uint64, "destChainId": p.uint64})}),
    Unpaused: event("0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa", "Unpaused(address)", {"account": p.address}),
    Upgraded: event("0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b", "Upgraded(address)", {"implementation": indexed(p.address)}),
}

export const functions = {
    AUTHORITY_ROLE: viewFun("0x4a3fba0e", "AUTHORITY_ROLE()", {}, p.bytes32),
    DEFAULT_ADMIN_ROLE: viewFun("0xa217fddf", "DEFAULT_ADMIN_ROLE()", {}, p.bytes32),
    MAX_CONFIRMATION_LENGTH: viewFun("0x46a0dd5c", "MAX_CONFIRMATION_LENGTH()", {}, p.uint256),
    PAUSER_ROLE: viewFun("0xe63ab1e9", "PAUSER_ROLE()", {}, p.bytes32),
    UPGRADE_INTERFACE_VERSION: viewFun("0xad3cb1cc", "UPGRADE_INTERFACE_VERSION()", {}, p.string),
    bridgeNonce: viewFun("0x1f87a250", "bridgeNonce()", {}, p.uint256),
    executeTransferRequests: fun("0x640c89ee", "executeTransferRequests((address,uint256,uint256,bytes32,uint64,(string,string,uint8))[])", {"batch": p.array(p.struct({"to": p.address, "amount": p.uint256, "nonce": p.uint256, "tokenKey": p.bytes32, "srcChainId": p.uint64, "tokenMetadata": p.struct({"name": p.string, "symbol": p.string, "decimals": p.uint8})}))}, ),
    getAdmin: viewFun("0x6e9960c3", "getAdmin()", {}, p.address),
    getAuthority: viewFun("0xe2b178a0", "getAuthority()", {}, p.address),
    getRoleAdmin: viewFun("0x248a9ca3", "getRoleAdmin(bytes32)", {"role": p.bytes32}, p.bytes32),
    grantRole: fun("0x2f2ff15d", "grantRole(bytes32,address)", {"role": p.bytes32, "account": p.address}, ),
    hasRole: viewFun("0x91d14854", "hasRole(bytes32,address)", {"role": p.bytes32, "account": p.address}, p.bool),
    initialize: fun("0x485cc955", "initialize(address,address)", {"_authority": p.address, "_admin": p.address}, ),
    isPaused: viewFun("0xb187bd26", "isPaused()", {}, p.bool),
    metadataForToken: viewFun("0x770eef9e", "metadataForToken(bytes32)", {"_0": p.bytes32}, {"name": p.string, "symbol": p.string, "decimals": p.uint8}),
    pause: fun("0x8456cb59", "pause()", {}, ),
    paused: viewFun("0x5c975abb", "paused()", {}, p.bool),
    processedTransfers: viewFun("0x5e33e5fd", "processedTransfers(uint64,uint256)", {"_0": p.uint64, "_1": p.uint256}, p.bool),
    proxiableUUID: viewFun("0x52d1902d", "proxiableUUID()", {}, p.bytes32),
    renounceRole: fun("0x36568abe", "renounceRole(bytes32,address)", {"role": p.bytes32, "callerConfirmation": p.address}, ),
    requestTransfer: fun("0x6c0652fb", "requestTransfer(bytes32,address,uint256,uint64)", {"tokenKey": p.bytes32, "to": p.address, "amount": p.uint256, "destChainId": p.uint64}, ),
    revokeRole: fun("0xd547741f", "revokeRole(bytes32,address)", {"role": p.bytes32, "account": p.address}, ),
    setAuthority: fun("0x7a9e5e4b", "setAuthority(address)", {"_authority": p.address}, ),
    setStakingManager: fun("0xb00bba6a", "setStakingManager(address)", {"_stakingManager": p.address}, ),
    stakingManager: viewFun("0x22828cc2", "stakingManager()", {}, p.address),
    supportsInterface: viewFun("0x01ffc9a7", "supportsInterface(bytes4)", {"interfaceId": p.bytes4}, p.bool),
    tokensInfo: viewFun("0xf0170150", "tokensInfo(bytes32)", {"_0": p.bytes32}, {"tokenAddress": p.address, "managed": p.bool, "enabled": p.bool, "supported": p.bool}),
    transferRequests: viewFun("0x03486c40", "transferRequests(uint256)", {"_0": p.uint256}, {"nonce": p.uint256, "from": p.address, "to": p.address, "tokenKey": p.bytes32, "amount": p.uint256, "srcChainId": p.uint64, "destChainId": p.uint64}),
    unpause: fun("0x3f4ba83a", "unpause()", {}, ),
    updateWhitelistStatus: fun("0x3cdf146e", "updateWhitelistStatus(bytes32,bool,address)", {"tokenKey": p.bytes32, "enabled": p.bool, "tokenAddress": p.address}, ),
    upgradeToAndCall: fun("0x4f1ef286", "upgradeToAndCall(address,bytes)", {"newImplementation": p.address, "data": p.bytes}, ),
    whitelistToken: fun("0x3227247a", "whitelistToken(bytes32,bool,address,string,string,uint8)", {"tokenKey": p.bytes32, "enabled": p.bool, "tokenAddress": p.address, "tokenSymbol": p.string, "tokenName": p.string, "tokenDecimals": p.uint8}, ),
}

export class Contract extends ContractBase {

    AUTHORITY_ROLE() {
        return this.eth_call(functions.AUTHORITY_ROLE, {})
    }

    DEFAULT_ADMIN_ROLE() {
        return this.eth_call(functions.DEFAULT_ADMIN_ROLE, {})
    }

    MAX_CONFIRMATION_LENGTH() {
        return this.eth_call(functions.MAX_CONFIRMATION_LENGTH, {})
    }

    PAUSER_ROLE() {
        return this.eth_call(functions.PAUSER_ROLE, {})
    }

    UPGRADE_INTERFACE_VERSION() {
        return this.eth_call(functions.UPGRADE_INTERFACE_VERSION, {})
    }

    bridgeNonce() {
        return this.eth_call(functions.bridgeNonce, {})
    }

    getAdmin() {
        return this.eth_call(functions.getAdmin, {})
    }

    getAuthority() {
        return this.eth_call(functions.getAuthority, {})
    }

    getRoleAdmin(role: GetRoleAdminParams["role"]) {
        return this.eth_call(functions.getRoleAdmin, {role})
    }

    hasRole(role: HasRoleParams["role"], account: HasRoleParams["account"]) {
        return this.eth_call(functions.hasRole, {role, account})
    }

    isPaused() {
        return this.eth_call(functions.isPaused, {})
    }

    metadataForToken(_0: MetadataForTokenParams["_0"]) {
        return this.eth_call(functions.metadataForToken, {_0})
    }

    paused() {
        return this.eth_call(functions.paused, {})
    }

    processedTransfers(_0: ProcessedTransfersParams["_0"], _1: ProcessedTransfersParams["_1"]) {
        return this.eth_call(functions.processedTransfers, {_0, _1})
    }

    proxiableUUID() {
        return this.eth_call(functions.proxiableUUID, {})
    }

    stakingManager() {
        return this.eth_call(functions.stakingManager, {})
    }

    supportsInterface(interfaceId: SupportsInterfaceParams["interfaceId"]) {
        return this.eth_call(functions.supportsInterface, {interfaceId})
    }

    tokensInfo(_0: TokensInfoParams["_0"]) {
        return this.eth_call(functions.tokensInfo, {_0})
    }

    transferRequests(_0: TransferRequestsParams["_0"]) {
        return this.eth_call(functions.transferRequests, {_0})
    }
}

/// Event types
export type InitializedEventArgs = EParams<typeof events.Initialized>
export type NewTokenWhitelistedEventArgs = EParams<typeof events.NewTokenWhitelisted>
export type PausedEventArgs = EParams<typeof events.Paused>
export type RoleAdminChangedEventArgs = EParams<typeof events.RoleAdminChanged>
export type RoleGrantedEventArgs = EParams<typeof events.RoleGranted>
export type RoleRevokedEventArgs = EParams<typeof events.RoleRevoked>
export type TokenWhitelistStatusUpdatedEventArgs = EParams<typeof events.TokenWhitelistStatusUpdated>
export type TokenWrappedEventArgs = EParams<typeof events.TokenWrapped>
export type TransferRequestExecutedEventArgs = EParams<typeof events.TransferRequestExecuted>
export type TransferRequestedEventArgs = EParams<typeof events.TransferRequested>
export type UnpausedEventArgs = EParams<typeof events.Unpaused>
export type UpgradedEventArgs = EParams<typeof events.Upgraded>

/// Function types
export type AUTHORITY_ROLEParams = FunctionArguments<typeof functions.AUTHORITY_ROLE>
export type AUTHORITY_ROLEReturn = FunctionReturn<typeof functions.AUTHORITY_ROLE>

export type DEFAULT_ADMIN_ROLEParams = FunctionArguments<typeof functions.DEFAULT_ADMIN_ROLE>
export type DEFAULT_ADMIN_ROLEReturn = FunctionReturn<typeof functions.DEFAULT_ADMIN_ROLE>

export type MAX_CONFIRMATION_LENGTHParams = FunctionArguments<typeof functions.MAX_CONFIRMATION_LENGTH>
export type MAX_CONFIRMATION_LENGTHReturn = FunctionReturn<typeof functions.MAX_CONFIRMATION_LENGTH>

export type PAUSER_ROLEParams = FunctionArguments<typeof functions.PAUSER_ROLE>
export type PAUSER_ROLEReturn = FunctionReturn<typeof functions.PAUSER_ROLE>

export type UPGRADE_INTERFACE_VERSIONParams = FunctionArguments<typeof functions.UPGRADE_INTERFACE_VERSION>
export type UPGRADE_INTERFACE_VERSIONReturn = FunctionReturn<typeof functions.UPGRADE_INTERFACE_VERSION>

export type BridgeNonceParams = FunctionArguments<typeof functions.bridgeNonce>
export type BridgeNonceReturn = FunctionReturn<typeof functions.bridgeNonce>

export type ExecuteTransferRequestsParams = FunctionArguments<typeof functions.executeTransferRequests>
export type ExecuteTransferRequestsReturn = FunctionReturn<typeof functions.executeTransferRequests>

export type GetAdminParams = FunctionArguments<typeof functions.getAdmin>
export type GetAdminReturn = FunctionReturn<typeof functions.getAdmin>

export type GetAuthorityParams = FunctionArguments<typeof functions.getAuthority>
export type GetAuthorityReturn = FunctionReturn<typeof functions.getAuthority>

export type GetRoleAdminParams = FunctionArguments<typeof functions.getRoleAdmin>
export type GetRoleAdminReturn = FunctionReturn<typeof functions.getRoleAdmin>

export type GrantRoleParams = FunctionArguments<typeof functions.grantRole>
export type GrantRoleReturn = FunctionReturn<typeof functions.grantRole>

export type HasRoleParams = FunctionArguments<typeof functions.hasRole>
export type HasRoleReturn = FunctionReturn<typeof functions.hasRole>

export type InitializeParams = FunctionArguments<typeof functions.initialize>
export type InitializeReturn = FunctionReturn<typeof functions.initialize>

export type IsPausedParams = FunctionArguments<typeof functions.isPaused>
export type IsPausedReturn = FunctionReturn<typeof functions.isPaused>

export type MetadataForTokenParams = FunctionArguments<typeof functions.metadataForToken>
export type MetadataForTokenReturn = FunctionReturn<typeof functions.metadataForToken>

export type PauseParams = FunctionArguments<typeof functions.pause>
export type PauseReturn = FunctionReturn<typeof functions.pause>

export type PausedParams = FunctionArguments<typeof functions.paused>
export type PausedReturn = FunctionReturn<typeof functions.paused>

export type ProcessedTransfersParams = FunctionArguments<typeof functions.processedTransfers>
export type ProcessedTransfersReturn = FunctionReturn<typeof functions.processedTransfers>

export type ProxiableUUIDParams = FunctionArguments<typeof functions.proxiableUUID>
export type ProxiableUUIDReturn = FunctionReturn<typeof functions.proxiableUUID>

export type RenounceRoleParams = FunctionArguments<typeof functions.renounceRole>
export type RenounceRoleReturn = FunctionReturn<typeof functions.renounceRole>

export type RequestTransferParams = FunctionArguments<typeof functions.requestTransfer>
export type RequestTransferReturn = FunctionReturn<typeof functions.requestTransfer>

export type RevokeRoleParams = FunctionArguments<typeof functions.revokeRole>
export type RevokeRoleReturn = FunctionReturn<typeof functions.revokeRole>

export type SetAuthorityParams = FunctionArguments<typeof functions.setAuthority>
export type SetAuthorityReturn = FunctionReturn<typeof functions.setAuthority>

export type SetStakingManagerParams = FunctionArguments<typeof functions.setStakingManager>
export type SetStakingManagerReturn = FunctionReturn<typeof functions.setStakingManager>

export type StakingManagerParams = FunctionArguments<typeof functions.stakingManager>
export type StakingManagerReturn = FunctionReturn<typeof functions.stakingManager>

export type SupportsInterfaceParams = FunctionArguments<typeof functions.supportsInterface>
export type SupportsInterfaceReturn = FunctionReturn<typeof functions.supportsInterface>

export type TokensInfoParams = FunctionArguments<typeof functions.tokensInfo>
export type TokensInfoReturn = FunctionReturn<typeof functions.tokensInfo>

export type TransferRequestsParams = FunctionArguments<typeof functions.transferRequests>
export type TransferRequestsReturn = FunctionReturn<typeof functions.transferRequests>

export type UnpauseParams = FunctionArguments<typeof functions.unpause>
export type UnpauseReturn = FunctionReturn<typeof functions.unpause>

export type UpdateWhitelistStatusParams = FunctionArguments<typeof functions.updateWhitelistStatus>
export type UpdateWhitelistStatusReturn = FunctionReturn<typeof functions.updateWhitelistStatus>

export type UpgradeToAndCallParams = FunctionArguments<typeof functions.upgradeToAndCall>
export type UpgradeToAndCallReturn = FunctionReturn<typeof functions.upgradeToAndCall>

export type WhitelistTokenParams = FunctionArguments<typeof functions.whitelistToken>
export type WhitelistTokenReturn = FunctionReturn<typeof functions.whitelistToken>

