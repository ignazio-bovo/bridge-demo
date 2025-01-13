import { EventDecodingError } from "@subsquid/evm-abi";
import { EventEmptyTopicsError } from "@subsquid/evm-abi";
import { EventInvalidSignatureError } from "@subsquid/evm-abi";
import { TokenMetadataDecodingError } from "../eth/decoding";
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

  async handleTokenWrapped(
    decodedEvent: ReturnType<typeof events.TokenWrapped.decode>,
    store: Store,
    chainId: number
  ): Promise<void> {
    const handlerLogger = this.logger.child({
      event: "TokenWrapped",
    });
    try {
      const chain = await store.get(Chain, chainId.toString()).then((chain) => {
        if (!chain) {
          throw new EntityNotFoundError("Chain", chainId.toString());
        }
        return chain;
      });
      const metadata = decodedEvent.tokenMetadata;

      const token = new Token({
        id: stripHexPrefix(decodedEvent.tokenKey),
        decimals: metadata.decimals,
        name: metadata.name,
        symbol: metadata.symbol,
      });
      const tokenWithChain = new TokenWithChain({
        id: `${stripHexPrefix(decodedEvent.tokenKey)}-${chain.id}`,
        token: token,
        chain: chain,
        address: decodedEvent.wrappedToken,
        native: checkNativeToken(decodedEvent.tokenKey, chainId.toString()),
      });
      token.chains = [tokenWithChain];
      chain.tokens = [tokenWithChain];
      await store.upsert(token).catch((error) => {
        throw new UpsertFailedError("Token", decodedEvent.tokenKey);
      });
      await store.upsert(chain).catch((error) => {
        throw new UpsertFailedError("Chain", chainId.toString());
      });
      await store.upsert(tokenWithChain).catch((error) => {
        throw new UpsertFailedError(
          "TokenWithChain",
          `${decodedEvent.tokenKey}-${chainId}`
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

  async handleTransferRequestExecuted(
    decodedEvent: ReturnType<typeof events.TransferRequestExecuted.decode>,
    store: Store
  ): Promise<void> {
    const handlerLogger = this.logger.child({
      event: "TransferRequestExecuted",
    });
    try {
      const id = transferId(
        Number(decodedEvent.nonce),
        Number(decodedEvent.srcChainId)
      );
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

  async handleTransferRequested(
    decodedEvent: ReturnType<typeof events.TransferRequested.decode>,
    store: Store
  ): Promise<void> {
    const handlerLogger = this.logger.child({
      event: "TransferRequested",
    });
    try {
      const id = transferId(
        Number(decodedEvent.request.nonce),
        Number(decodedEvent.request.srcChainId)
      );
      const sourceChain = await store
        .get(Chain, decodedEvent.request.srcChainId.toString())
        .then((chain) => {
          if (!chain) {
            throw new EntityNotFoundError(
              "Chain",
              decodedEvent.request.srcChainId.toString()
            );
          }
          return chain;
        });
      const destinationChain = await store
        .get(Chain, decodedEvent.request.destChainId.toString())
        .then((chain) => {
          if (!chain) {
            throw new EntityNotFoundError(
              "Chain",
              decodedEvent.request.destChainId.toString()
            );
          }
          return chain;
        });
      const token = await store
        .get(Token, stripHexPrefix(decodedEvent.request.tokenKey))
        .then((token) => {
          if (!token) {
            throw new EntityNotFoundError(
              "Token",
              decodedEvent.request.tokenKey
            );
          }
          return token;
        });

      const txData = new BridgeTxData({
        id,
        sourceAddress: decodedEvent.request.from,
        destinationAddress: decodedEvent.request.to,
        amount: BigInt(decodedEvent.request.amount),
        nonce: BigInt(decodedEvent.request.nonce),
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
          decodedEvent.request.srcChainId.toString()
        );
      });
      await store.upsert(destinationChain).catch((error) => {
        throw new UpsertFailedError(
          "Chain",
          decodedEvent.request.destChainId.toString()
        );
      });
      await store.upsert(token).catch((error) => {
        throw new UpsertFailedError("Token", decodedEvent.request.tokenKey);
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
    decodedEvent: {
      tokenKey: string;
      tokenMetadata: { decimals: number; name: string; symbol: string };
    },
    store: Store,
    chainId: number
  ): Promise<void> {
    const handlerLogger = this.logger.child({
      event: "TokenWhitelistStatusUpdated",
    });
    try {
      const chain = await store.get(Chain, chainId.toString());
      if (!chain) {
        throw new Error(`Chain not found for id: ${chainId}`);
      }
      const token = new Token({
        id: stripHexPrefix(decodedEvent.tokenKey),
        decimals: decodedEvent.tokenMetadata.decimals,
        name: decodedEvent.tokenMetadata.name,
        symbol: decodedEvent.tokenMetadata.symbol,
      });
      const tokenWithChain = new TokenWithChain({
        id: `${stripHexPrefix(decodedEvent.tokenKey)}-${chain.id}`,
        token: token,
        chain: chain,
        address: decodedEvent.tokenKey,
        native: checkNativeToken(decodedEvent.tokenKey, chainId.toString()),
      });
      token.chains = [tokenWithChain];
      await store.upsert(token).catch(() => {
        throw new UpsertFailedError(
          "Token",
          stripHexPrefix(decodedEvent.tokenKey)
        );
      });
      await store.upsert(tokenWithChain).catch(() => {
        throw new UpsertFailedError(
          "TokenWithChain",
          `${decodedEvent.tokenKey}-${chain.id}`
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
