import { EventDecodingError } from "@subsquid/evm-abi";

import { EventEmptyTopicsError } from "@subsquid/evm-abi";

import { EventInvalidSignatureError } from "@subsquid/evm-abi";
import {
  decodeTokenMetadata,
  TokenMetadataDecodingError,
} from "../eth/decoding";
import { BridgeTxData, TokenWithChain } from "../model";
import { Token } from "../model";
import { events } from "../abi/Bridge.sol";
import { Chain } from "../model";
import { Log } from "@subsquid/evm-processor";
import { Store } from "@subsquid/typeorm-store";
import {
  checkNativeToken,
  EntityNotFoundError,
  PendingTxNotFoundError,
  stripHexPrefix,
  transferId,
  UpsertFailedError,
} from "../utils";
import { Logger, createLogger } from "@subsquid/logger";

export class MappingHandler {
  private logger: Logger;

  constructor() {
    this.logger = createLogger("sqd:mapping-handler");
  }

  async tokenWrapped(log: Log, store: Store, chainId: number): Promise<void> {
    const handlerLogger = this.logger.child({
      event: "TokenWrapped",
      blockNumber: log.block.height,
      transactionHash: log.transaction?.hash,
    });
    if (log.topics[0] !== events.TokenWrapped.topic) {
      return;
    }
    try {
      const decoded = events.TokenWrapped.decode(log);
      const chain = await store.get(Chain, chainId.toString()).then((chain) => {
        if (!chain) {
          throw new EntityNotFoundError("Chain", chainId.toString());
        }
        return chain;
      });
      const metadata = decoded.tokenMetadata;

      const token = new Token({
        id: stripHexPrefix(decoded.tokenKey),
        decimals: metadata.decimals,
        name: metadata.name,
        symbol: metadata.symbol,
      });
      const tokenWithChain = new TokenWithChain({
        id: `${stripHexPrefix(decoded.tokenKey)}-${chain.id}`,
        token: token,
        chain: chain,
        address: decoded.wrappedToken,
        native: checkNativeToken(decoded.tokenKey, chainId.toString()),
      });
      token.chains = [tokenWithChain];
      chain.tokens = [tokenWithChain];
      await store.upsert(token).catch((error) => {
        throw new UpsertFailedError("Token", decoded.tokenKey);
      });
      await store.upsert(chain).catch((error) => {
        throw new UpsertFailedError("Chain", chainId.toString());
      });
      await store.upsert(tokenWithChain).catch((error) => {
        throw new UpsertFailedError(
          "TokenWithChain",
          `${decoded.tokenKey}-${chainId}`
        );
      });
    } catch (error) {
      if (
        error instanceof EventDecodingError ||
        error instanceof EventEmptyTopicsError ||
        error instanceof EventInvalidSignatureError
      ) {
        handlerLogger.error(
          { error: error.name },
          `Error decoding token wrapped event: ${error.message}`
        );
      } else if (error instanceof EntityNotFoundError) {
        handlerLogger.error(
          { error: error.name },
          `Error with typeorm operation: ${error}`
        );
      } else if (error instanceof UpsertFailedError) {
        handlerLogger.error(
          { error: error.name },
          `Error with typeorm operation: ${error}`
        );
      } else {
        handlerLogger.error(`Error with typeorm operation: ${error}`);
      }
    }
  }

  async handleTransferRequestExecuted(log: Log, store: Store): Promise<void> {
    const handlerLogger = this.logger.child({
      event: "TransferRequestExecuted",
      blockNumber: log.block.height,
      transactionHash: log.transaction?.hash,
    });
    if (log.topics[0] !== events.TransferRequestExecuted.topic) {
      return;
    }
    try {
      const decoded = events.TransferRequestExecuted.decode(log);
      const id = transferId(Number(decoded.nonce), Number(decoded.srcChainId));
      const txData = await store.get(BridgeTxData, id);
      if (!txData) {
        throw new PendingTxNotFoundError(id);
      }
      txData.confirmed = true;
      txData.confirmedAtTimestamp = new Date();
      await store.upsert(txData).catch((error) => {
        throw new UpsertFailedError("BridgeTxData", id);
      });
    } catch (error) {
      if (
        error instanceof EventDecodingError ||
        error instanceof EventEmptyTopicsError ||
        error instanceof EventInvalidSignatureError
      ) {
        handlerLogger.error(
          { error: error.name },
          `Error decoding transfer request executed event: ${error.message}`
        );
      } else if (error instanceof PendingTxNotFoundError) {
        handlerLogger.error(
          { error: error.name },
          `Pending transaction not found for id: ${error.message}. Cannot continue processing transfer request execution.`
        );
        // TODO: Enable in production
        // throw error;
      } else {
        handlerLogger.error(`Error with typeorm operation: ${error}`);
      }
    }
  }

  async handleTransferRequested(log: Log, store: Store): Promise<void> {
    const handlerLogger = this.logger.child({
      event: "TransferRequested",
      blockNumber: log.block.height,
      transactionHash: log.transaction?.hash,
    });
    if (log.topics[0] !== events.TransferRequested.topic) {
      return;
    }
    try {
      const decoded = events.TransferRequested.decode(log);
      const id = transferId(
        Number(decoded.request.nonce),
        Number(decoded.request.srcChainId)
      );
      const sourceChain = await store
        .get(Chain, decoded.request.srcChainId.toString())
        .then((chain) => {
          if (!chain) {
            throw new EntityNotFoundError(
              "Chain",
              decoded.request.srcChainId.toString()
            );
          }
          return chain;
        });
      const destinationChain = await store
        .get(Chain, decoded.request.destChainId.toString())
        .then((chain) => {
          if (!chain) {
            throw new EntityNotFoundError(
              "Chain",
              decoded.request.destChainId.toString()
            );
          }
          return chain;
        });
      const token = await store
        .get(Token, stripHexPrefix(decoded.request.tokenKey))
        .then((token) => {
          if (!token) {
            throw new EntityNotFoundError("Token", decoded.request.tokenKey);
          }
          return token;
        });

      const txData = new BridgeTxData({
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
        throw new UpsertFailedError("BridgeTxData", id);
      });
      await store.upsert(sourceChain).catch((error) => {
        throw new UpsertFailedError(
          "Chain",
          decoded.request.srcChainId.toString()
        );
      });
      await store.upsert(destinationChain).catch((error) => {
        throw new UpsertFailedError(
          "Chain",
          decoded.request.destChainId.toString()
        );
      });
      await store.upsert(token).catch((error) => {
        throw new UpsertFailedError("Token", decoded.request.tokenKey);
      });
    } catch (error) {
      if (
        error instanceof EventDecodingError ||
        error instanceof EventEmptyTopicsError ||
        error instanceof EventInvalidSignatureError
      ) {
        handlerLogger.error(
          { error: error.name },
          `Error decoding transfer requested event: ${error.message}`
        );
      } else {
        handlerLogger.error(`Error with typeorm operation: ${error}`);
      }
    }
  }

  async handleTokenWhitelistStatusUpdated(
    log: Log,
    store: Store,
    chainId: number
  ): Promise<void> {
    const handlerLogger = this.logger.child({
      event: "TokenWhitelistStatusUpdated",
      blockNumber: log.block.height,
      transactionHash: log.transaction?.hash,
    });
    if (log.topics[0] !== events.NewTokenWhitelisted.topic) {
      return;
    }
    try {
      const decoded = events.NewTokenWhitelisted.decode(log);
      const tokenMetadata = decoded.tokenMetadata;

      const chain = await store.get(Chain, chainId.toString());
      if (!chain) {
        throw new Error(`Chain not found for id: ${chainId}`);
      }
      const token = new Token({
        id: stripHexPrefix(decoded.tokenKey),
        decimals: tokenMetadata.decimals,
        name: tokenMetadata.name,
        symbol: tokenMetadata.symbol,
      });
      const tokenWithChain = new TokenWithChain({
        id: `${stripHexPrefix(decoded.tokenKey)}-${chain.id}`,
        token: token,
        chain: chain,
        address: decoded.tokenKey,
        native: checkNativeToken(decoded.tokenKey, chainId.toString()),
      });
      token.chains = [tokenWithChain];
      await store.upsert(token).catch(() => {
        throw new UpsertFailedError("Token", stripHexPrefix(decoded.tokenKey));
      });
      await store.upsert(tokenWithChain).catch(() => {
        throw new UpsertFailedError(
          "TokenWithChain",
          `${decoded.tokenKey}-${chain.id}`
        );
      });
    } catch (error) {
      if (
        error instanceof EventDecodingError ||
        error instanceof EventEmptyTopicsError ||
        error instanceof EventInvalidSignatureError
      ) {
        handlerLogger.error(
          { error: error.name },
          `Error decoding token whitelist status updated event: ${error.message}`
        );
      } else if (error instanceof TokenMetadataDecodingError) {
        handlerLogger.error(
          { error: error.name },
          `Error decoding token metadata: ${error.message}`
        );
      } else {
        handlerLogger.error(`Error with typeorm operation: ${error}`);
      }
    }
  }
}
