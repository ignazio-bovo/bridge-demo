import { EvmBatchProcessor, Log } from "@subsquid/evm-processor";
import { Store, TypeormDatabase } from "@subsquid/typeorm-store";
import { Logger, createLogger } from "@subsquid/logger";
import { MappingHandler } from "../mappings/handler";
import { events } from "../abi/Bridge.sol";

export class EthereumProcessor {
  private chainId = 1;
  private processor: EvmBatchProcessor<{ log: { transactionHash: true } }>;
  private logger: Logger;
  private rpcEndpoint = process.env.EVM_RPC_ENDPOINT || "http://127.0.0.1:8545";
  private contractAddress =
    process.env.EVM_CONTRACT_ADDRESS ||
    "0x261D8c5e9742e6f7f1076Fa1F560894524e19cad";

  private mappingHandler: MappingHandler;

  constructor() {
    this.mappingHandler = new MappingHandler();

    this.logger = createLogger("sqd:evm-processor");
    this.processor = new EvmBatchProcessor()
      .setBlockRange({ from: 0 })
      .setRpcEndpoint(this.rpcEndpoint)
      .setFinalityConfirmation(0)
      .addLog({
        address: [this.contractAddress],
      });

    this.logger.info(
      `Ethereum Processor initialized with rpc endpoint ${this.rpcEndpoint}`
    );
  }

  private decodeTokenWrapped(log: Log) {
    if (log.topics[0] !== events.TokenWrapped.topic) {
      return null;
    }
    return events.TokenWrapped.decode(log);
  }

  private decodeTransferRequestExecuted(log: Log) {
    try {
      const decoded = events.TransferRequestExecuted.decode(log);
      return decoded;
    } catch (error) {
      return null;
    }
  }

  private decodeNewTokenWhitelisted(log: Log) {
    try {
      const decoded = events.NewTokenWhitelisted.decode(log);
      return decoded;
    } catch (error) {
      return null;
    }
  }

  private decodeTransferRequested(log: Log) {
    try {
      const decoded = events.TransferRequested.decode(log);
      return decoded;
    } catch (error) {
      return null;
    }
  }

  async run(db: TypeormDatabase): Promise<void> {
    this.processor.run(db, async (ctx) => {
      for (let block of ctx.blocks) {
        for (let log of block.logs) {
          this.logger.info(`Processing log: ${log}`);
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

    const decodedTransferRequested = this.decodeTransferRequested(log);
    if (decodedTransferRequested) {
      await this.mappingHandler.handleTransferRequested(
        decodedTransferRequested,
        store
      );
      return;
    }

    const decodedWhitelist = this.decodeNewTokenWhitelisted(log);
    if (decodedWhitelist) {
      await this.mappingHandler.handleNewTokenWhitelisted(
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
