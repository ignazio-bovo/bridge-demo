import * as p from '@subsquid/evm-codec'
import { event, fun, viewFun, indexed, ContractBase } from '@subsquid/evm-abi'
import type { EventParams as EParams, FunctionArguments, FunctionReturn } from '@subsquid/evm-abi'

export const events = {
    FundsStakedOnPrecompile: event("0xcb35838c2d3b529d682bc65aac48be02411f8eb3b9146368db6feead7e1f6e2c", "FundsStakedOnPrecompile(uint256,uint64,uint64)", {"amount": p.uint256, "lastStakingBlock": p.uint64, "stakingEpochId": p.uint64}),
    Initialized: event("0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2", "Initialized(uint64)", {"version": p.uint64}),
    OwnershipTransferred: event("0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0", "OwnershipTransferred(address,address)", {"previousOwner": indexed(p.address), "newOwner": indexed(p.address)}),
    TokensStaked: event("0xb539ca1e5c8d398ddf1c41c30166f33404941683be4683319b57669a93dad4ef", "TokensStaked(address,uint256)", {"user": indexed(p.address), "amount": p.uint256}),
    TokensUnstaked: event("0x9845e367b683334e5c0b12d7b81721ac518e649376fa65e3d68324e8f34f2679", "TokensUnstaked(address,uint256)", {"user": indexed(p.address), "amount": p.uint256}),
    Upgraded: event("0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b", "Upgraded(address)", {"implementation": indexed(p.address)}),
}

export const functions = {
    NORMALIZER: viewFun("0x07e72f45", "NORMALIZER()", {}, p.uint64),
    ROOT_NET_UID: viewFun("0x230cde7b", "ROOT_NET_UID()", {}, p.uint16),
    STAKE_INTERVAL: viewFun("0xec8b70d1", "STAKE_INTERVAL()", {}, p.uint64),
    UPGRADE_INTERFACE_VERSION: viewFun("0xad3cb1cc", "UPGRADE_INTERFACE_VERSION()", {}, p.string),
    addBridgeParticipantReward: fun("0xb12f7998", "addBridgeParticipantReward(address)", {"rewardAddress": p.address}, ),
    bridge: viewFun("0xe78cea92", "bridge()", {}, p.address),
    daturaStakingHotkey: viewFun("0xd544f674", "daturaStakingHotkey()", {}, p.bytes32),
    initialize: fun("0x294e8d0e", "initialize(address,uint256,bytes32,address)", {"_bridge": p.address, "_rewardRate": p.uint256, "_daturaHotkey": p.bytes32, "_stakePrecompile": p.address}, ),
    lastStakingBlock: viewFun("0x44f98990", "lastStakingBlock()", {}, p.uint64),
    nextStakingEpochId: viewFun("0xc347e4e8", "nextStakingEpochId()", {}, p.uint64),
    owner: viewFun("0x8da5cb5b", "owner()", {}, p.address),
    proxiableUUID: viewFun("0x52d1902d", "proxiableUUID()", {}, p.bytes32),
    removeBridgeParticipantReward: fun("0xf82b45be", "removeBridgeParticipantReward(address)", {"rewardAddress": p.address}, ),
    renounceOwnership: fun("0x715018a6", "renounceOwnership()", {}, ),
    rewardRate: viewFun("0x7b0a47ee", "rewardRate()", {}, p.uint256),
    stake: fun("0x7acb7757", "stake(uint256,address)", {"amount": p.uint256, "user": p.address}, ),
    stakePrecompile: viewFun("0xaf03528b", "stakePrecompile()", {}, p.address),
    stakeToPrecompile: fun("0xee8aae30", "stakeToPrecompile()", {}, ),
    stakes: viewFun("0x16934fc4", "stakes(address)", {"_0": p.address}, {"amount": p.uint256, "stakingEpochId": p.uint64}),
    stakingEpochIdToLastStakingBlock: viewFun("0x65f5f670", "stakingEpochIdToLastStakingBlock(uint64)", {"_0": p.uint64}, p.uint64),
    transferOwnership: fun("0xf2fde38b", "transferOwnership(address)", {"newOwner": p.address}, ),
    unstake: fun("0x8381e182", "unstake(uint256,address)", {"amount": p.uint256, "user": p.address}, ),
    upgradeToAndCall: fun("0x4f1ef286", "upgradeToAndCall(address,bytes)", {"newImplementation": p.address, "data": p.bytes}, ),
}

export class Contract extends ContractBase {

    NORMALIZER() {
        return this.eth_call(functions.NORMALIZER, {})
    }

    ROOT_NET_UID() {
        return this.eth_call(functions.ROOT_NET_UID, {})
    }

    STAKE_INTERVAL() {
        return this.eth_call(functions.STAKE_INTERVAL, {})
    }

    UPGRADE_INTERFACE_VERSION() {
        return this.eth_call(functions.UPGRADE_INTERFACE_VERSION, {})
    }

    bridge() {
        return this.eth_call(functions.bridge, {})
    }

    daturaStakingHotkey() {
        return this.eth_call(functions.daturaStakingHotkey, {})
    }

    lastStakingBlock() {
        return this.eth_call(functions.lastStakingBlock, {})
    }

    nextStakingEpochId() {
        return this.eth_call(functions.nextStakingEpochId, {})
    }

    owner() {
        return this.eth_call(functions.owner, {})
    }

    proxiableUUID() {
        return this.eth_call(functions.proxiableUUID, {})
    }

    rewardRate() {
        return this.eth_call(functions.rewardRate, {})
    }

    stakePrecompile() {
        return this.eth_call(functions.stakePrecompile, {})
    }

    stakes(_0: StakesParams["_0"]) {
        return this.eth_call(functions.stakes, {_0})
    }

    stakingEpochIdToLastStakingBlock(_0: StakingEpochIdToLastStakingBlockParams["_0"]) {
        return this.eth_call(functions.stakingEpochIdToLastStakingBlock, {_0})
    }
}

/// Event types
export type FundsStakedOnPrecompileEventArgs = EParams<typeof events.FundsStakedOnPrecompile>
export type InitializedEventArgs = EParams<typeof events.Initialized>
export type OwnershipTransferredEventArgs = EParams<typeof events.OwnershipTransferred>
export type TokensStakedEventArgs = EParams<typeof events.TokensStaked>
export type TokensUnstakedEventArgs = EParams<typeof events.TokensUnstaked>
export type UpgradedEventArgs = EParams<typeof events.Upgraded>

/// Function types
export type NORMALIZERParams = FunctionArguments<typeof functions.NORMALIZER>
export type NORMALIZERReturn = FunctionReturn<typeof functions.NORMALIZER>

export type ROOT_NET_UIDParams = FunctionArguments<typeof functions.ROOT_NET_UID>
export type ROOT_NET_UIDReturn = FunctionReturn<typeof functions.ROOT_NET_UID>

export type STAKE_INTERVALParams = FunctionArguments<typeof functions.STAKE_INTERVAL>
export type STAKE_INTERVALReturn = FunctionReturn<typeof functions.STAKE_INTERVAL>

export type UPGRADE_INTERFACE_VERSIONParams = FunctionArguments<typeof functions.UPGRADE_INTERFACE_VERSION>
export type UPGRADE_INTERFACE_VERSIONReturn = FunctionReturn<typeof functions.UPGRADE_INTERFACE_VERSION>

export type AddBridgeParticipantRewardParams = FunctionArguments<typeof functions.addBridgeParticipantReward>
export type AddBridgeParticipantRewardReturn = FunctionReturn<typeof functions.addBridgeParticipantReward>

export type BridgeParams = FunctionArguments<typeof functions.bridge>
export type BridgeReturn = FunctionReturn<typeof functions.bridge>

export type DaturaStakingHotkeyParams = FunctionArguments<typeof functions.daturaStakingHotkey>
export type DaturaStakingHotkeyReturn = FunctionReturn<typeof functions.daturaStakingHotkey>

export type InitializeParams = FunctionArguments<typeof functions.initialize>
export type InitializeReturn = FunctionReturn<typeof functions.initialize>

export type LastStakingBlockParams = FunctionArguments<typeof functions.lastStakingBlock>
export type LastStakingBlockReturn = FunctionReturn<typeof functions.lastStakingBlock>

export type NextStakingEpochIdParams = FunctionArguments<typeof functions.nextStakingEpochId>
export type NextStakingEpochIdReturn = FunctionReturn<typeof functions.nextStakingEpochId>

export type OwnerParams = FunctionArguments<typeof functions.owner>
export type OwnerReturn = FunctionReturn<typeof functions.owner>

export type ProxiableUUIDParams = FunctionArguments<typeof functions.proxiableUUID>
export type ProxiableUUIDReturn = FunctionReturn<typeof functions.proxiableUUID>

export type RemoveBridgeParticipantRewardParams = FunctionArguments<typeof functions.removeBridgeParticipantReward>
export type RemoveBridgeParticipantRewardReturn = FunctionReturn<typeof functions.removeBridgeParticipantReward>

export type RenounceOwnershipParams = FunctionArguments<typeof functions.renounceOwnership>
export type RenounceOwnershipReturn = FunctionReturn<typeof functions.renounceOwnership>

export type RewardRateParams = FunctionArguments<typeof functions.rewardRate>
export type RewardRateReturn = FunctionReturn<typeof functions.rewardRate>

export type StakeParams = FunctionArguments<typeof functions.stake>
export type StakeReturn = FunctionReturn<typeof functions.stake>

export type StakePrecompileParams = FunctionArguments<typeof functions.stakePrecompile>
export type StakePrecompileReturn = FunctionReturn<typeof functions.stakePrecompile>

export type StakeToPrecompileParams = FunctionArguments<typeof functions.stakeToPrecompile>
export type StakeToPrecompileReturn = FunctionReturn<typeof functions.stakeToPrecompile>

export type StakesParams = FunctionArguments<typeof functions.stakes>
export type StakesReturn = FunctionReturn<typeof functions.stakes>

export type StakingEpochIdToLastStakingBlockParams = FunctionArguments<typeof functions.stakingEpochIdToLastStakingBlock>
export type StakingEpochIdToLastStakingBlockReturn = FunctionReturn<typeof functions.stakingEpochIdToLastStakingBlock>

export type TransferOwnershipParams = FunctionArguments<typeof functions.transferOwnership>
export type TransferOwnershipReturn = FunctionReturn<typeof functions.transferOwnership>

export type UnstakeParams = FunctionArguments<typeof functions.unstake>
export type UnstakeReturn = FunctionReturn<typeof functions.unstake>

export type UpgradeToAndCallParams = FunctionArguments<typeof functions.upgradeToAndCall>
export type UpgradeToAndCallReturn = FunctionReturn<typeof functions.upgradeToAndCall>

