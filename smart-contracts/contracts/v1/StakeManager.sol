// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;
import {IBridge} from "../interfaces/IBridge.sol";
import {IStakeManager} from "../interfaces/IStakeManager.sol";
import {IStakingPrecompile} from "../interfaces/IStaking.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {EnumerableSet} from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";
import {ReentrancyGuardUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";

contract StakeManager is
    IStakeManager,
    UUPSUpgradeable,
    OwnableUpgradeable,
    ReentrancyGuardUpgradeable
{
    address public stakePrecompile;

    uint64 public constant NORMALIZER = 1e18; // blocks produced in 5 years, if avg. block production is 1 block per 6 seconds

    bytes32 public daturaStakingHotkey;
    uint256 public rewardRate;
    IBridge public bridge;
    uint16 public constant ROOT_NET_UID = 0; // used in the stake precompile to identify the network
    uint64 public constant STAKE_INTERVAL = 360; // 360 blocks, constraint in the subtensor staking logic
    uint64 public nextStakingEpochId;

    mapping(address => Stake) public stakes;
    using EnumerableSet for EnumerableSet.AddressSet;
    EnumerableSet.AddressSet private rewardRecipients;
    uint64 public lastStakingBlock;
    mapping(uint64 => uint64) public stakingEpochIdToLastStakingBlock;

    modifier maybeRouteToStakePrecompile() {
        uint64 blocksPassedSinceLastStaking = uint64(block.number) -
            lastStakingBlock;
        if (blocksPassedSinceLastStaking >= STAKE_INTERVAL) {
            uint256 amount = address(this).balance;
            _callStakePrecompile(amount);
            lastStakingBlock = uint64(block.number);
            uint64 currentStakingEpochId = nextStakingEpochId;
            nextStakingEpochId++;
            stakingEpochIdToLastStakingBlock[
                currentStakingEpochId
            ] = lastStakingBlock;
            emit FundsStakedOnPrecompile(
                amount,
                lastStakingBlock,
                currentStakingEpochId
            );
        }
        _;
    }

    // simply triggers funding to stake precompile, it should be called by a trigger bot in case of inactivity
    function stakeToPrecompile() external maybeRouteToStakePrecompile {}

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() initializer {}

    function initialize(
        IBridge _bridge,
        uint256 _rewardRate,
        bytes32 _daturaHotkey,
        address _stakePrecompile
    ) external initializer {
        bridge = _bridge;
        rewardRate = _rewardRate;
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();
        daturaStakingHotkey = _daturaHotkey;
        address admin = bridge.getAdmin();
        __Ownable_init(admin);
        stakePrecompile = _stakePrecompile;
        lastStakingBlock = uint64(block.number);
        nextStakingEpochId = 1;
        __ReentrancyGuard_init();
    }

    modifier onlyBridge() {
        if (msg.sender != address(bridge)) revert NotBridge();
        _;
    }

    // called internally by the bridge to route amount for staking
    function _callStakePrecompile(uint256 amount) internal {
        IStakingPrecompile stakingPrecompile = IStakingPrecompile(
            stakePrecompile
        );
        (bool success, ) = stakePrecompile.call{value: amount}(
            abi.encodeWithSelector(
                stakingPrecompile.addStake.selector,
                daturaStakingHotkey,
                ROOT_NET_UID
            )
        );
        if (!success) revert StakingFailed();
    }

    function _callUnstakePrecompile(uint256 amount) internal {
        IStakingPrecompile stakingPrecompile = IStakingPrecompile(
            stakePrecompile
        );
        (bool success, ) = stakePrecompile.call(
            abi.encodeWithSelector(
                stakingPrecompile.removeStake.selector,
                daturaStakingHotkey,
                amount,
                ROOT_NET_UID
            )
        );
        if (!success) revert UnstakingFailed();
    }

    function stake(
        uint256 amount,
        address user
    ) external payable onlyBridge maybeRouteToStakePrecompile {
        stakes[user].amount += amount;
        stakes[user].stakingEpochId = nextStakingEpochId;
        emit TokensStaked(user, amount);
    }

    function unstake(
        uint256 amount,
        address user
    ) external onlyBridge maybeRouteToStakePrecompile nonReentrant {
        // CHECKS
        Stake storage stakePosition = stakes[user];
        if (stakePosition.amount < amount) revert InsufficientStakedBalance();

        // Calculate reward before any state changes
        uint256 reward = 0;
        if (stakePosition.stakingEpochId < nextStakingEpochId) {
            reward =
                (stakePosition.amount *
                    rewardRate *
                    uint256(
                        uint64(block.number) -
                            stakingEpochIdToLastStakingBlock[
                                stakePosition.stakingEpochId
                            ] -
                            1
                    )) /
                uint256(NORMALIZER);
            reward = (reward > amount) ? amount : reward;
        }

        // EFFECTS
        stakePosition.stakingEpochId = nextStakingEpochId;
        stakePosition.amount -= amount;
        emit TokensUnstaked(user, amount);

        // Calculate reward distribution (if any)
        uint256 rewardPerParticipant = 0;
        uint256 userReward = reward;
        if (reward > 0 && rewardRecipients.length() > 0) {
            rewardPerParticipant =
                (reward - reward / 2) /
                rewardRecipients.length();
            userReward =
                reward -
                (rewardPerParticipant * rewardRecipients.length());
        }

        // INTERACTIONS
        if (reward > 0) {
            _callUnstakePrecompile(amount + reward);
        }

        // Transfer funds to user
        (bool success, ) = user.call{value: amount + userReward}("");
        if (!success) revert UnstakingFailed();

        // Distribute rewards to participants
        if (reward > 0 && rewardRecipients.length() > 0) {
            for (uint256 i = 0; i < rewardRecipients.length(); i++) {
                (success, ) = rewardRecipients.at(i).call{
                    value: rewardPerParticipant
                }("");
                if (!success) revert UnstakingFailed();
            }
        }
    }

    function addBridgeParticipantReward(
        address rewardAddress
    ) external onlyOwner {
        if (rewardAddress == address(0)) revert InvalidAddress();
        if (rewardRecipients.contains(rewardAddress))
            revert RewardRecipientAlreadyExists();
        rewardRecipients.add(rewardAddress);
    }

    function removeBridgeParticipantReward(
        address rewardAddress
    ) external onlyOwner {
        if (!rewardRecipients.contains(rewardAddress))
            revert RewardRecipientNotFound();
        rewardRecipients.remove(rewardAddress);
    }

    function _authorizeUpgrade(address) internal override onlyOwner {}

    receive() external payable {}
}
