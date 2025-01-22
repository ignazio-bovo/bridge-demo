import { Store, TypeormDatabase } from "@subsquid/typeorm-store";
import { SubstrateBatchProcessor, Event } from "@subsquid/substrate-processor";
import { Logger, createLogger } from "@subsquid/logger";
import { MappingHandler } from "../mappings/handler";
import { events } from "../abi/Bridge.sol";
import { EventRecord } from "@subsquid/evm-abi";

export class SubtensorProcessor {
  private processor: SubstrateBatchProcessor;
  private logger: Logger;
  private handler: MappingHandler;
  private chainId: number;
  private rpcEndpoint =
    process.env.SUBTENSOR_RPC_ENDPOINT || "http://127.0.0.1:9944";
  private contractAddress =
    process.env.SUBTENSOR_CONTRACT_ADDRESS ||
    "0x057ef64E23666F000b34aE31332854aCBd1c8544";

  constructor() {
    this.logger = createLogger("sqd:subtensor-processor");
    this.handler = new MappingHandler();
    this.chainId = Number(process.env.SUBTENSOR_CHAIN_ID) || 945;

    this.processor = new SubstrateBatchProcessor()
      .setBlockRange({ from: 0 })
      .setRpcEndpoint(this.rpcEndpoint)
      .setFinalityConfirmation(0)
      .addEvmLog({
        address: [this.contractAddress],
      })
      .setFields({
        event: {
          name: true,
          args: true,
          topics: true,
          data: true,
        },
      });

    this.logger.info(
      `Subtensor Processor initialized with rpc endpoint ${this.rpcEndpoint}`
    );
  }

  async runEtl(event: EventRecord, store: Store): Promise<void> {
    const decodedTokenWrapped = this.decodeTokenWrapped(event);
    if (decodedTokenWrapped) {
      await this.handler.handleTokenWrapped(
        decodedTokenWrapped,
        store,
        this.chainId
      );
      return;
    }

    const decodedTransferRequested = this.decodeTransferRequested(event);
    if (decodedTransferRequested) {
      await this.handler.handleTransferRequested(
        decodedTransferRequested,
        store
      );
      return;
    }

    const decodedWhitelist = this.decodeNewTokenWhitelisted(event);
    if (decodedWhitelist) {
      await this.handler.handleNewTokenWhitelisted(
        decodedWhitelist,
        store,
        this.chainId
      );
      return;
    }

    const decodedTransferRequestExecuted =
      this.decodeTransferRequestExecuted(event);
    if (decodedTransferRequestExecuted) {
      await this.handler.handleTransferRequestExecuted(
        decodedTransferRequestExecuted,
        store
      );
      return;
    }
  }

  private decodeTransferRequested<E extends EventRecord>(event: E) {
    try {
      return events.TransferRequested.decode(event);
    } catch (error) {
      return null;
    }
  }

  private decodeTokenWrapped<E extends EventRecord>(event: E) {
    try {
      return events.TokenWrapped.decode(event);
    } catch (error) {
      return null;
    }
  }

  private decodeTransferRequestExecuted<E extends EventRecord>(event: E) {
    try {
      return events.TransferRequestExecuted.decode(event);
    } catch (error) {
      return null;
    }
  }

  private decodeNewTokenWhitelisted<E extends EventRecord>(event: E) {
    try {
      return events.NewTokenWhitelisted.decode(event);
    } catch (error) {
      return null;
    }
  }

  async run(db: TypeormDatabase): Promise<void> {
    this.processor.run(db, async (ctx) => {
      for (let block of ctx.blocks) {
        for (let event of block.events) {
          if (event.name === "EVM.Log") {
            await this.runEtl(event.args.log, ctx.store);
          }
        }
      }
    });
  }
}
