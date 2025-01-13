import { EvmBatchProcessor, Log } from "@subsquid/evm-processor";
import { Store, TypeormDatabase } from "@subsquid/typeorm-store";
import { Logger, createLogger } from "@subsquid/logger";
import { MappingHandler } from "../mappings/handler";
import { events } from "../abi/Bridge.sol";

export class EthereumProcessor {
  private chainId = 1;
  private processor: EvmBatchProcessor<{ log: { transactionHash: true } }>;
  private logger: Logger;
  private CONTRACT_ADDRESS =
    process.env.CONTRACT_ADDRESS ||
    "0x71c95911e9a5d330f4d621842ec243ee1343292e";

  private mappingHandler: MappingHandler;

  constructor() {
    this.mappingHandler = new MappingHandler();

    this.logger = createLogger("sqd:evm-processor");
    this.processor = new EvmBatchProcessor()
      .setRpcEndpoint("http://127.0.0.1:8545")
      .setFinalityConfirmation(15)
      .addLog({
        address: [this.CONTRACT_ADDRESS],
        range: { from: 1 },
      })
      .setFields({
        log: {
          transactionHash: true,
          topics: true,
        },
      });

    this.logger.info(
      `Ethereum Processor initialized with rpc endpoint 127.0.0.1:8545`
    );
  }

  private decodeTokenWrapped(log: Log) {
    if (log.topics[0] !== events.TokenWrapped.topic) {
      return null;
    }
    return events.TokenWrapped.decode(log);
  }

  private decodeTransferRequestExecuted(log: Log) {
    if (log.topics[0] !== events.TransferRequestExecuted.topic) {
      return null;
    }
    return events.TransferRequestExecuted.decode(log);
  }

  private decodeTokenWhitelistStatusUpdated(log: Log) {
    if (log.topics[0] !== events.NewTokenWhitelisted.topic) {
      return null;
    }
    return events.NewTokenWhitelisted.decode(log);
  }

  async run(db: TypeormDatabase): Promise<void> {
    this.processor.run(db, async (ctx) => {
      for (let block of ctx.blocks) {
        for (let log of block.logs) {
          await this.runEtl(log, ctx.store);
        }
      }
    });
  }

  async runEtl(log: Log, store: Store): Promise<void> {
    const decodedTokenWrapped = this.decodeTokenWrapped(log);
    if (decodedTokenWrapped) {
      await this.mappingHandler.handleTokenWrapped(
        decodedTokenWrapped,
        store,
        this.chainId
      );
      return;
    }

    const decodedWhitelist = this.decodeTokenWhitelistStatusUpdated(log);
    if (decodedWhitelist) {
      await this.mappingHandler.handleTokenWhitelistStatusUpdated(
        decodedWhitelist,
        store,
        this.chainId
      );
      return;
    }

    const decodedTransferRequestExecuted =
      this.decodeTransferRequestExecuted(log);
    if (decodedTransferRequestExecuted) {
      await this.mappingHandler.handleTransferRequestExecuted(
        decodedTransferRequestExecuted,
        store
      );
      return;
    }
  }
}
