// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IStakingPrecompile} from "../interfaces/IStaking.sol";
import "hardhat/console.sol";

contract StakePrecompileMock is IStakingPrecompile {
    function addStake(bytes32 hotkey, uint16 netUid) external payable {
        // msg.value is automatically transferred to this contract
    }

    function removeStake(
        bytes32 /* hotkey */,
        uint256 amount,
        uint16 /* _netUid */
    ) external {
        (bool success, ) = msg.sender.call{value: amount}("");
        if (!success) {
            console.log(
                "Transfer failed with amountToUnstake vs this.balance:",
                amount,
                address(this).balance
            );
        }
        require(success, "Transfer failed");
    }

    receive() external payable {}
}
