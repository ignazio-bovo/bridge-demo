// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;
interface IStakingPrecompile {
    function addStake(bytes32 hotkey, uint16 netUid) external payable;
    function removeStake(
        bytes32 hotkey,
        uint256 amount,
        uint16 netUid
    ) external;
}
