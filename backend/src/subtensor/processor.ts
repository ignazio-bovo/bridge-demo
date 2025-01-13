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
    "0x71c95911e9a5d330f4d621842ec243ee1343292e";

  constructor() {
    this.logger = createLogger("sqd:subtensor-processor");
    this.handler = new MappingHandler();
    this.chainId = 945;
    // First we configure data retrieval.
    this.processor = new SubstrateBatchProcessor()
      .setRpcEndpoint("http://127.0.0.1:9944")
      .setFinalityConfirmation(15)
      .addEthereumTransaction({})
      .addEvmLog({
        address: [this.CONTRACT_ADDRESS],
        range: { from: 1 },
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
    if (event.args.topics[0] !== events.TokenWrapped.topic) {
      return null;
    }
    return events.TokenWrapped.decode(event.args);
  }

  private decodeTransferRequestExecuted(event: Event) {
    if (event.args.topics[0] !== events.TransferRequestExecuted.topic) {
      return null;
    }
    return events.TransferRequestExecuted.decode(event.args);
  }

  private decodeTokenWhitelistStatusUpdated(event: Event) {
    if (event.args.topics[0] !== events.NewTokenWhitelisted.topic) {
      return null;
    }
    return events.NewTokenWhitelisted.decode(event.args);
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
