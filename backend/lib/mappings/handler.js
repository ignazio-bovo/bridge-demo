"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.MappingHandler = void 0;
const evm_abi_1 = require("@subsquid/evm-abi");
const evm_abi_2 = require("@subsquid/evm-abi");
const evm_abi_3 = require("@subsquid/evm-abi");
const decoding_1 = require("../eth/decoding");
const model_1 = require("../model");
const model_2 = require("../model");
const Bridge_sol_1 = require("../abi/Bridge.sol");
const model_3 = require("../model");
const utils_1 = require("../utils");
const logger_1 = require("@subsquid/logger");
class MappingHandler {
    constructor() {
        this.logger = (0, logger_1.createLogger)("sqd:mapping-handler");
    }
    async tokenWrapped(log, store, chainId) {
        const handlerLogger = this.logger.child({
            event: "TokenWrapped",
            blockNumber: log.block.height,
            transactionHash: log.transaction?.hash,
        });
        if (log.topics[0] !== Bridge_sol_1.events.TokenWrapped.topic) {
            return;
        }
        try {
            const decoded = Bridge_sol_1.events.TokenWrapped.decode(log);
            const chain = await store.get(model_3.Chain, chainId.toString()).then((chain) => {
                if (!chain) {
                    throw new utils_1.EntityNotFoundError("Chain", chainId.toString());
                }
                return chain;
            });
            const metadata = decoded.tokenMetadata;
            const token = new model_2.Token({
                id: (0, utils_1.stripHexPrefix)(decoded.tokenKey),
                decimals: metadata.decimals,
                name: metadata.name,
                symbol: metadata.symbol,
            });
            const tokenWithChain = new model_1.TokenWithChain({
                id: `${(0, utils_1.stripHexPrefix)(decoded.tokenKey)}-${chain.id}`,
                token: token,
                chain: chain,
                address: decoded.wrappedToken,
                native: (0, utils_1.checkNativeToken)(decoded.tokenKey, chainId.toString()),
            });
            token.chains = [tokenWithChain];
            chain.tokens = [tokenWithChain];
            await store.upsert(token).catch((error) => {
                throw new utils_1.UpsertFailedError("Token", decoded.tokenKey);
            });
            await store.upsert(chain).catch((error) => {
                throw new utils_1.UpsertFailedError("Chain", chainId.toString());
            });
            await store.upsert(tokenWithChain).catch((error) => {
                throw new utils_1.UpsertFailedError("TokenWithChain", `${decoded.tokenKey}-${chainId}`);
            });
        }
        catch (error) {
            if (error instanceof evm_abi_1.EventDecodingError ||
                error instanceof evm_abi_2.EventEmptyTopicsError ||
                error instanceof evm_abi_3.EventInvalidSignatureError) {
                handlerLogger.error({ error: error.name }, `Error decoding token wrapped event: ${error.message}`);
            }
            else if (error instanceof utils_1.EntityNotFoundError) {
                handlerLogger.error({ error: error.name }, `Error with typeorm operation: ${error}`);
            }
            else if (error instanceof utils_1.UpsertFailedError) {
                handlerLogger.error({ error: error.name }, `Error with typeorm operation: ${error}`);
            }
            else {
                handlerLogger.error(`Error with typeorm operation: ${error}`);
            }
        }
    }
    async handleTransferRequestExecuted(log, store) {
        const handlerLogger = this.logger.child({
            event: "TransferRequestExecuted",
            blockNumber: log.block.height,
            transactionHash: log.transaction?.hash,
        });
        if (log.topics[0] !== Bridge_sol_1.events.TransferRequestExecuted.topic) {
            return;
        }
        try {
            const decoded = Bridge_sol_1.events.TransferRequestExecuted.decode(log);
            const id = (0, utils_1.transferId)(Number(decoded.nonce), Number(decoded.srcChainId));
            const txData = await store.get(model_1.BridgeTxData, id);
            if (!txData) {
                throw new utils_1.PendingTxNotFoundError(id);
            }
            txData.confirmed = true;
            txData.confirmedAtTimestamp = new Date();
            await store.upsert(txData).catch((error) => {
                throw new utils_1.UpsertFailedError("BridgeTxData", id);
            });
        }
        catch (error) {
            if (error instanceof evm_abi_1.EventDecodingError ||
                error instanceof evm_abi_2.EventEmptyTopicsError ||
                error instanceof evm_abi_3.EventInvalidSignatureError) {
                handlerLogger.error({ error: error.name }, `Error decoding transfer request executed event: ${error.message}`);
            }
            else if (error instanceof utils_1.PendingTxNotFoundError) {
                handlerLogger.error({ error: error.name }, `Pending transaction not found for id: ${error.message}. Cannot continue processing transfer request execution.`);
                // TODO: Enable in production
                // throw error;
            }
            else {
                handlerLogger.error(`Error with typeorm operation: ${error}`);
            }
        }
    }
    async handleTransferRequested(log, store) {
        const handlerLogger = this.logger.child({
            event: "TransferRequested",
            blockNumber: log.block.height,
            transactionHash: log.transaction?.hash,
        });
        if (log.topics[0] !== Bridge_sol_1.events.TransferRequested.topic) {
            return;
        }
        try {
            const decoded = Bridge_sol_1.events.TransferRequested.decode(log);
            const id = (0, utils_1.transferId)(Number(decoded.request.nonce), Number(decoded.request.srcChainId));
            const sourceChain = await store
                .get(model_3.Chain, decoded.request.srcChainId.toString())
                .then((chain) => {
                if (!chain) {
                    throw new utils_1.EntityNotFoundError("Chain", decoded.request.srcChainId.toString());
                }
                return chain;
            });
            const destinationChain = await store
                .get(model_3.Chain, decoded.request.destChainId.toString())
                .then((chain) => {
                if (!chain) {
                    throw new utils_1.EntityNotFoundError("Chain", decoded.request.destChainId.toString());
                }
                return chain;
            });
            const token = await store
                .get(model_2.Token, (0, utils_1.stripHexPrefix)(decoded.request.tokenKey))
                .then((token) => {
                if (!token) {
                    throw new utils_1.EntityNotFoundError("Token", decoded.request.tokenKey);
                }
                return token;
            });
            const txData = new model_1.BridgeTxData({
                id,
                sourceAddress: decoded.request.from,
                destinationAddress: decoded.request.to,
                amount: decoded.request.amount,
                nonce: decoded.request.nonce,
                sourceChain: sourceChain,
                destinationChain: destinationChain,
                token: token,
                confirmed: false,
                createdAtTimestamp: new Date(),
            });
            sourceChain.outgoingTransfers = [
                ...(sourceChain.outgoingTransfers || []),
                txData,
            ];
            destinationChain.incomingTransfers = [
                ...(destinationChain.incomingTransfers || []),
                txData,
            ];
            token.transfers = [...(token.transfers || []), txData];
            await store.upsert(txData).catch((error) => {
                throw new utils_1.UpsertFailedError("BridgeTxData", id);
            });
            await store.upsert(sourceChain).catch((error) => {
                throw new utils_1.UpsertFailedError("Chain", decoded.request.srcChainId.toString());
            });
            await store.upsert(destinationChain).catch((error) => {
                throw new utils_1.UpsertFailedError("Chain", decoded.request.destChainId.toString());
            });
            await store.upsert(token).catch((error) => {
                throw new utils_1.UpsertFailedError("Token", decoded.request.tokenKey);
            });
        }
        catch (error) {
            if (error instanceof evm_abi_1.EventDecodingError ||
                error instanceof evm_abi_2.EventEmptyTopicsError ||
                error instanceof evm_abi_3.EventInvalidSignatureError) {
                handlerLogger.error({ error: error.name }, `Error decoding transfer requested event: ${error.message}`);
            }
            else {
                handlerLogger.error(`Error with typeorm operation: ${error}`);
            }
        }
    }
    async handleTokenWhitelistStatusUpdated(log, store, chainId) {
        const handlerLogger = this.logger.child({
            event: "TokenWhitelistStatusUpdated",
            blockNumber: log.block.height,
            transactionHash: log.transaction?.hash,
        });
        if (log.topics[0] !== Bridge_sol_1.events.NewTokenWhitelisted.topic) {
            return;
        }
        try {
            const decoded = Bridge_sol_1.events.NewTokenWhitelisted.decode(log);
            const tokenMetadata = decoded.tokenMetadata;
            const chain = await store.get(model_3.Chain, chainId.toString());
            if (!chain) {
                throw new Error(`Chain not found for id: ${chainId}`);
            }
            const token = new model_2.Token({
                id: (0, utils_1.stripHexPrefix)(decoded.tokenKey),
                decimals: tokenMetadata.decimals,
                name: tokenMetadata.name,
                symbol: tokenMetadata.symbol,
            });
            const tokenWithChain = new model_1.TokenWithChain({
                id: `${(0, utils_1.stripHexPrefix)(decoded.tokenKey)}-${chain.id}`,
                token: token,
                chain: chain,
                address: decoded.tokenKey,
                native: (0, utils_1.checkNativeToken)(decoded.tokenKey, chainId.toString()),
            });
            token.chains = [tokenWithChain];
            await store.upsert(token).catch(() => {
                throw new utils_1.UpsertFailedError("Token", (0, utils_1.stripHexPrefix)(decoded.tokenKey));
            });
            await store.upsert(tokenWithChain).catch(() => {
                throw new utils_1.UpsertFailedError("TokenWithChain", `${decoded.tokenKey}-${chain.id}`);
            });
        }
        catch (error) {
            if (error instanceof evm_abi_1.EventDecodingError ||
                error instanceof evm_abi_2.EventEmptyTopicsError ||
                error instanceof evm_abi_3.EventInvalidSignatureError) {
                handlerLogger.error({ error: error.name }, `Error decoding token whitelist status updated event: ${error.message}`);
            }
            else if (error instanceof decoding_1.TokenMetadataDecodingError) {
                handlerLogger.error({ error: error.name }, `Error decoding token metadata: ${error.message}`);
            }
            else {
                handlerLogger.error(`Error with typeorm operation: ${error}`);
            }
        }
    }
}
exports.MappingHandler = MappingHandler;
//# sourceMappingURL=handler.js.map