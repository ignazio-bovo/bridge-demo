// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IStakeManager} from "./IStakeManager.sol";

/// @title Interface for Bridge Contract
interface IBridge {
    // Custom errors
    /// @notice Thrown when a non-authority address attempts to perform an authority-only action
    error NotBridgeAuthority();
    /// @notice Thrown when there are insufficient funds for a transfer
    error InsufficientBalance();
    /// @notice Thrown when an ETH transfer fails
    error ETHTransferFailed();
    /// @notice Thrown when there are insufficient tokens locked in the contract
    error InsufficientTokensLocked();
    /// @notice Thrown when attempting to process a transfer that was already processed
    error TransferAlreadyProcessed();
    /// @notice Thrown when the transfer amount is invalid (zero or exceeds total supply)
    error InvalidAmount();
    /// @notice Thrown when an unauthorized transfer is attempted
    error UnauthorizedTransfer();
    /// @notice Thrown when the recipient address is invalid (zero address)
    error InvalidRecipient();
    /// @notice Thrown when the sender address is invalid (zero address)
    error InvalidSender();
    /// @notice Thrown when an invalid admin address is provided
    error InvalidAdmin();
    /// @notice Thrown when an invalid authority address is provided
    error InvalidAuthority();
    /// @notice Thrown when the allowance is insufficient for the transfer
    error InsufficientAllowance();
    /// @notice Thrown when the input is invalid
    error InvalidInput();
    /// @notice Thrown when the signature is invalid
    error InvalidSignature();
    /// @notice Thrown when attempting to transfer a token that is not whitelisted
    error TokenNotWhitelisted();
    /// @notice Thrown when a token transfer operation fails
    error TransferFailed();
    /// @notice Thrown when the destination chain ID is invalid or matches the current chain
    error InvalidDestinationChain();
    /// @notice Thrown when the bridge is paused
    error BridgePaused();
    /// @notice Thrown when the staking manager is not set
    error StakingManagerNotSet();
    /// @notice Thrown when the token metadata is invalid
    error InvalidTokenMetadata();
    /// @notice Token already added to the whitelist
    error TokenAlreadyWhitelisted();

    /// @notice Token metadata struct
    /// @param name The name of the token
    /// @param symbol The symbol of the token
    /// @param decimals The decimals of the token
    struct TokenMetadata {
        string name;
        string symbol;
        uint8 decimals;
    }

    /// @notice Parameters required to confirm a transfer request
    /// @param to Recipient address receiving the transfer
    /// @param amount Amount of tokens being transferred
    /// @param transferId Unique hash identifying the transfer
    /// @param sourceChainNative Whether the source chain is native
    /// @param tokenMetadata Metadata about the token (e.g. decimals, name, symbol) json serialized
    struct TransferExecutionDetails {
        address to;
        uint256 amount;
        uint256 nonce;
        bytes32 tokenKey;
        uint64 srcChainId;
        TokenMetadata tokenMetadata;
    }

    /// @notice Struct containing information about a token in the bridge
    /// @param token Address of the token contract
    /// @param enabled Whether the token is enabled
    struct TokenInfo {
        address tokenAddress;
        bool managed; // true if the token is managed by the bridge via mint/burn
        bool enabled; // true if the token is enabled for transfers by the bridge
        bool supported; // true if the token is supported by the bridge
    }

    /// @notice Struct containing information about a transfer request
    /// @param nonce Unique identifier for the transfer request
    /// @param from Address initiating the transfer
    /// @param to Address receiving the transfer
    /// @param tokenKey Key of the token being transferred
    /// @param amount Amount of tokens being transferred
    /// @param srcChainId Chain ID where the transfer originates
    /// @param destChainId Chain ID where the transfer is destined
    struct TransferRequest {
        uint256 nonce;
        address from;
        address to;
        bytes32 tokenKey;
        uint256 amount;
        uint64 srcChainId;
        uint64 destChainId;
    }

    // Events
    event TransferRequested(TransferRequest request);
    /// @notice Emitted when a new wrapped token is created or updated
    /// @param tokenKey The key of the token
    /// @param wrappedToken The address of the wrapped token on this chain
    /// @param isWhitelisted Whether the token is whitelisted for bridging
    /// @param tokenMetadata Metadata about the token (e.g. decimals, name, symbol) json serialized
    event TokenWrapped(
        bytes32 indexed tokenKey,
        address indexed wrappedToken,
        bool indexed isWhitelisted,
        TokenMetadata tokenMetadata
    );

    /// @notice Emitted when a single transfer request has been executed
    /// @param nonce The nonce of the transfer request
    /// @param srcChainId The chain ID where the transfer originated
    event TransferRequestExecuted(uint256 nonce, uint64 srcChainId);

    /// @notice Emitted when a token's whitelist status is updated
    /// @param tokenKey The key of the token
    /// @param tokenInfo The token information including its whitelist status
    event TokenWhitelistStatusUpdated(
        bytes32 indexed tokenKey,
        TokenInfo tokenInfo
    );

    /// @notice Emitted when a new token is whitelisted
    /// @param tokenKey The key of the token
    /// @param tokenInfo The token information including its whitelist status
    /// @param tokenMetadata The token metadata
    event NewTokenWhitelisted(
        bytes32 indexed tokenKey,
        TokenInfo tokenInfo,
        TokenMetadata tokenMetadata
    );

    /// @notice Requests a token transfer to another chain
    /// @param tokenKey Key of the token
    /// @param to Recipient address on the destination chain
    /// @param amount Amount of tokens to transfer
    /// @param destinationChainId ID of the destination chain
    function requestTransfer(
        bytes32 tokenKey,
        address to,
        uint256 amount,
        uint64 destinationChainId
    ) external payable;

    /// @notice Confirms a transfer request on the destination chain
    /// @param batch Array of transfer requests to confirm
    function executeTransferRequests(
        TransferExecutionDetails[] calldata batch
    ) external;

    /// @notice Sets the staking manager contract address
    /// @param stakingManager The address of the new staking manager contract
    function setStakingManager(IStakeManager stakingManager) external;

    /// @notice Returns the address of the admin role
    /// @return The address of the admin
    function getAdmin() external view returns (address);

    /// @notice Returns the address of the authority role
    /// @return The address of the authority
    function getAuthority() external view returns (address);

    /// @notice Updates the whitelist status of a token
    /// @param tokenKey The key of the token to whitelist
    /// @param enabled Whether to enable or disable the token
    /// @param tokenAddress The address of the token contract
    /// @param tokenSymbol The symbol of the token
    /// @param tokenName The name of the token
    /// @param tokenDecimals The decimals of the token
    function whitelistToken(
        bytes32 tokenKey,
        bool enabled,
        address tokenAddress,
        string memory tokenSymbol,
        string memory tokenName,
        uint8 tokenDecimals
    ) external;

    function updateWhitelistStatus(
        bytes32 tokenKey,
        bool enabled,
        address tokenAddress
    ) external;
}
