// SPDX-License-Identifier: MIT

pragma solidity ^0.8.20;
import "hardhat/console.sol";

import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {AccessControlUpgradeable} from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IBridge} from "../interfaces/IBridge.sol";
import {BridgedToken} from "./BridgedToken.sol";
import {IStakeManager} from "../interfaces/IStakeManager.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import {ReentrancyGuardUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";

/// @title Bridge Contract for Cross-Chain Token Transfers
/// @notice This contract enables the transfer of tokens between different blockchain networks.
contract Bridge is
    IBridge,
    UUPSUpgradeable,
    AccessControlUpgradeable,
    PausableUpgradeable,
    ReentrancyGuardUpgradeable
{
    using SafeERC20 for IERC20;

    uint256 public constant MAX_CONFIRMATION_LENGTH = 100;
    // Roles
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE"); // can pause the bridge
    bytes32 public constant AUTHORITY_ROLE = keccak256("AUTHORITY_ROLE"); // can execute transfers

    uint256 public bridgeNonce;
    address private authority;
    address private admin;
    bool public isPaused;
    IStakeManager public stakingManager;

    mapping(uint64 => mapping(uint256 => bool)) public processedTransfers;
    mapping(uint256 => TransferRequest) public transferRequests;
    mapping(bytes32 => TokenInfo) public tokensInfo;
    mapping(bytes32 => TokenMetadata) public metadataForToken;

    function initialize(
        address _authority,
        address _admin
    ) external initializer {
        if (_admin == address(0)) revert InvalidAdmin();
        if (_authority == address(0)) revert InvalidAuthority();

        authority = _authority;
        admin = _admin;

        __UUPSUpgradeable_init();
        __AccessControl_init();
        __Pausable_init();

        _grantRole(DEFAULT_ADMIN_ROLE, _admin);
        _grantRole(PAUSER_ROLE, _admin);
        _grantRole(AUTHORITY_ROLE, _authority);
        __ReentrancyGuard_init();
    }

    function _authorizeUpgrade(
        address newImplementation
    ) internal override onlyRole(DEFAULT_ADMIN_ROLE) {}

    function pause() external onlyRole(PAUSER_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(PAUSER_ROLE) {
        _unpause();
    }

    function requestTransfer(
        bytes32 tokenKey,
        address to,
        uint256 amount,
        uint64 destChainId
    ) external payable whenNotPaused nonReentrant {
        // CHECKS
        if (destChainId == block.chainid) {
            revert InvalidDestinationChain();
        }
        if (to == address(0)) revert InvalidRecipient();
        if (amount == 0) revert InvalidAmount();
        if (msg.sender == address(0)) revert InvalidSender();

        TokenInfo memory tokenInfo = tokensInfo[tokenKey];
        if (!tokenInfo.supported) revert TokenNotWhitelisted();

        if (!tokenInfo.managed && tokenInfo.tokenAddress == address(0)) {
            if (msg.value < amount) revert InsufficientBalance();
        } else if (!tokenInfo.managed) {
            if (
                IERC20(tokenInfo.tokenAddress).allowance(
                    msg.sender,
                    address(this)
                ) < amount
            ) {
                revert InsufficientAllowance();
            }
        }

        // EFFECTS
        uint64 chainId = uint64(block.chainid);
        TransferRequest memory request = TransferRequest({
            nonce: bridgeNonce,
            from: msg.sender,
            to: to,
            tokenKey: tokenKey,
            amount: amount,
            srcChainId: chainId,
            destChainId: destChainId
        });
        transferRequests[bridgeNonce] = request;
        bridgeNonce++;

        // INTERACTIONS
        if (tokenInfo.managed) {
            BridgedToken(tokenInfo.tokenAddress).burnFrom(msg.sender, amount);
        } else if (tokenInfo.tokenAddress == address(0)) {
            if (stakingManager != IStakeManager(address(0))) {
                stakingManager.stake{value: amount}(amount, msg.sender);
            }
        } else {
            IERC20(tokenInfo.tokenAddress).transferFrom(
                msg.sender,
                address(this),
                amount
            );
        }

        emit TransferRequested(request);
    }

    function executeTransferRequests(
        TransferExecutionDetails[] calldata batch
    ) external onlyRole(AUTHORITY_ROLE) whenNotPaused nonReentrant {
        if (batch.length == 0 || batch.length > MAX_CONFIRMATION_LENGTH)
            revert InvalidInput();

        for (uint256 i = 0; i < batch.length; ++i) {
            _processTransferRequest(batch[i]);
            emit TransferRequestExecuted(batch[i].nonce, batch[i].srcChainId);
        }
    }

    function _processTransferRequest(
        TransferExecutionDetails calldata request
    ) internal {
        // CHECKS
        if (processedTransfers[request.srcChainId][request.nonce])
            revert TransferAlreadyProcessed();

        TokenInfo memory tokenInfo = tokensInfo[request.tokenKey];
        BridgedToken token;
        TokenInfo memory newTokenInfo;

        // EFFECTS
        processedTransfers[request.srcChainId][request.nonce] = true;

        if (!tokenInfo.supported) {
            // Create new token contract
            token = new BridgedToken(
                request.tokenMetadata.name,
                request.tokenMetadata.symbol,
                address(this)
            );
            // Update state
            newTokenInfo = TokenInfo({
                tokenAddress: address(token),
                enabled: true,
                managed: true,
                supported: true
            });
            tokensInfo[request.tokenKey] = newTokenInfo;
            metadataForToken[request.tokenKey] = request.tokenMetadata;

            emit TokenWrapped(
                request.tokenKey,
                address(token),
                true,
                request.tokenMetadata
            );
        }

        // INTERACTIONS
        if (!tokenInfo.supported) {
            token.mint(request.to, request.amount);
        } else if (tokenInfo.managed) {
            BridgedToken(tokenInfo.tokenAddress).mint(
                request.to,
                request.amount
            );
        } else if (tokenInfo.tokenAddress == address(0)) {
            if (stakingManager != IStakeManager(address(0))) {
                stakingManager.unstake(request.amount, request.to);
            } else {
                // native token
                (bool success, ) = request.to.call{value: request.amount}("");
                if (!success) revert TransferFailed();
            }
        } else {
            // ERC20 token
            IERC20(tokenInfo.tokenAddress).transfer(request.to, request.amount);
        }
    }

    function whitelistToken(
        bytes32 tokenKey,
        bool enabled,
        address tokenAddress,
        string memory tokenSymbol,
        string memory tokenName,
        uint8 tokenDecimals
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        // avoid unbounded string length
        if (bytes(tokenSymbol).length > 80 || bytes(tokenName).length > 80) {
            revert InvalidTokenMetadata();
        }

        TokenInfo storage tokenInfo = tokensInfo[tokenKey];
        if (tokenInfo.supported) {
            revert TokenAlreadyWhitelisted();
        }

        tokenInfo.tokenAddress = tokenAddress;
        tokenInfo.supported = true;
        tokenInfo.enabled = enabled;
        tokenInfo.supported = true;

        TokenMetadata memory tokenMetadata = TokenMetadata({
            symbol: tokenSymbol,
            name: tokenName,
            decimals: tokenDecimals
        });
        metadataForToken[tokenKey] = tokenMetadata;

        emit NewTokenWhitelisted(tokenKey, tokenInfo, tokenMetadata);
    }

    function updateWhitelistStatus(
        bytes32 tokenKey,
        bool enabled,
        address tokenAddress
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        TokenInfo memory tokenInfo = tokensInfo[tokenKey];
        tokenInfo.tokenAddress = tokenAddress;
        tokenInfo.managed = false;
        tokenInfo.supported = enabled;
        emit TokenWhitelistStatusUpdated(tokenKey, tokenInfo);
    }

    // TODO: to be removed in production mainnet
    function setAuthority(
        address _authority
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        authority = _authority;
    }

    function getAuthority() external view returns (address) {
        return authority;
    }

    function getAdmin() external view returns (address) {
        return admin;
    }

    function setStakingManager(
        IStakeManager _stakingManager
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        stakingManager = _stakingManager;
    }

    receive() external payable {}
    fallback() external payable {}
}
