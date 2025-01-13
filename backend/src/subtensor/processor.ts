import { TypeormDatabase } from "@subsquid/typeorm-store";
import { SubstrateBatchProcessor } from "@subsquid/substrate-processor";
import { Logger, createLogger } from "@subsquid/logger";
import { Event } from "@subsquid/substrate-processor";
import { EvmBatchProcessor, Log } from "@subsquid/evm-processor";

export class SubtensorProcessor {
  // private processor: SubstrateBatchProcessor;
  private processor: EvmBatchProcessor;
  private logger: Logger;

  constructor() {
    this.logger = createLogger("sqd:evm-processor");

    // First we configure data retrieval.

    this.processor = new EvmBatchProcessor()
      .setRpcEndpoint("http://127.0.0.1:8545")
      .setFinalityConfirmation(15)
      .addLog({ range: { from: 1 } });
    // this.processor = new SubstrateBatchProcessor()
    //   .setRpcEndpoint("http://127.0.0.1:9944")
    //   .setFinalityConfirmation(15)
    //   .addEthereumTransaction({})
    //   .addEvmLog({});

    this.logger.info(
      `Subtensor Processor initialized with rpc endpoint 127.0.0.1:9944`
    );
  }

  async runEtl<Store>(log: Log, store: Store): Promise<void> {
    this.logger!.info(
      `Processing log for block ${log.block.height} with topic ${log.topics[0]}`
    );
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
}
