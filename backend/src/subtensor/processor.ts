import { Store, TypeormDatabase } from "@subsquid/typeorm-store";
import { SubstrateBatchProcessor, Event } from "@subsquid/substrate-processor";
import { Logger, createLogger } from "@subsquid/logger";
import { MappingHandler } from "../mappings/handler";
import { events } from "../abi/Bridge.sol";

export class SubtensorProcessor {
  private processor: SubstrateBatchProcessor;
  private logger: Logger;
  private handler: MappingHandler;
  private chainId: number;
  private CONTRACT_ADDRESS =
    process.env.CONTRACT_ADDRESS ||
    "0x057ef64E23666F000b34aE31332854aCBd1c8544";

  constructor() {
    this.logger = createLogger("sqd:subtensor-processor");
    this.handler = new MappingHandler();
    this.chainId = 945;

    this.processor = new SubstrateBatchProcessor()
      .setBlockRange({ from: 0 })
      .setRpcEndpoint("http://127.0.0.1:9944")
      .setFinalityConfirmation(0)
      .addEvmLog({
        address: [this.CONTRACT_ADDRESS],
      });

    this.logger.info(
      `Subtensor Processor initialized with rpc endpoint 127.0.0.1:9944`
    );
  }

  async runEtl(event: Event, store: Store): Promise<void> {
    const decodedTokenWrapped = this.decodeTokenWrapped(event);
    if (decodedTokenWrapped) {
      await this.handler.handleTokenWrapped(
        decodedTokenWrapped,
        store,
        this.chainId
      );
      return;
    }

    const decodedWhitelist = this.decodeTokenWhitelistStatusUpdated(event);
    if (decodedWhitelist) {
      await this.handler.handleTokenWhitelistStatusUpdated(
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

  private decodeTokenWrapped(event: Event) {
    try {
      return events.TokenWrapped.decode(event.args);
    } catch (error) {
      return null;
    }
  }

  private decodeTransferRequestExecuted(event: Event) {
    try {
      return events.TransferRequestExecuted.decode(event.args);
    } catch (error) {
      return null;
    }
  }

  private decodeTokenWhitelistStatusUpdated(event: Event) {
    try {
      return events.NewTokenWhitelisted.decode(event.args);
    } catch (error) {
      return null;
    }
  }

  async run(db: TypeormDatabase): Promise<void> {
    this.processor.run(db, async (ctx) => {
      for (let block of ctx.blocks) {
        for (let event of block.events) {
          if (event.name === "EVM.Log") {
            await this.runEtl(event, ctx.store);
          }
        }
      }
    });
  }
}
