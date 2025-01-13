// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IStakeManager {
    function stake(uint256 amount, address user) external payable;
    function unstake(uint256 amount, address user) external;

    /// @notice Thrown when unstaking fails
    error UnstakingFailed();
    /// @notice Thrown when staking fails
    error StakingFailed();
    /// @notice Thrown when the caller is not the bridge
    error NotBridge();
    /// @notice Thrown when the caller is not the admin
    error NotAdmin();
    /// @notice Thrown when the staked balance is insufficient
    error InsufficientStakedBalance();
    /// @notice Thrown when adding existing reward recipient
    error RewardRecipientAlreadyExists();
    /// @notice Thrown when removing non existent reward recipient
    error RewardRecipientNotFound();
    /// @notice Thrown when the address is invalid
    error InvalidAddress();

    /// @notice Emitted when tokens are staked by a user
    /// @param user The address of the user who staked tokens
    /// @param amount The amount of tokens staked
    event TokensStaked(address indexed user, uint256 amount);

    /// @notice Emitted when tokens are unstaked by a user
    /// @param user The address of the user who unstaked tokens
    /// @param amount The amount of tokens unstaked
    event TokensUnstaked(address indexed user, uint256 amount);

    /// @notice Emitted when funds are staked on the precompile contract
    /// @param amount The amount of funds staked on the precompile
    /// @param lastStakingBlock The block number when the staking occurred
    /// @param stakingEpochId The staking epoch id
    event FundsStakedOnPrecompile(
        uint256 amount,
        uint64 lastStakingBlock,
        uint64 stakingEpochId
    );
    struct Stake {
        uint256 amount;
        uint64 stakingEpochId;
    }

    function addBridgeParticipantReward(address rewardAddress) external;
    function removeBridgeParticipantReward(address rewardAddress) external;
}
