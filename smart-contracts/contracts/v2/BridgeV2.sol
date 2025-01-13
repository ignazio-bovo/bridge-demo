// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.20;
import {Bridge} from "../v1/Bridge.sol";

contract BridgeV2 is Bridge {
    event Health(string version);
    function testUpgrade() external {
        emit Health("v2");
    }

    function upgradeInitialize(
        address _asset,
        address _authority,
        uint64 _chainId,
        address _admin
    ) external reinitializer(2) {}
}
