import { EvmBatchProcessor, Log } from "@subsquid/evm-processor";
import { Store, TypeormDatabase } from "@subsquid/typeorm-store";
import { Logger, createLogger } from "@subsquid/logger";
import { MappingHandler } from "../mappings/handler";
import { events } from "../abi/Bridge.sol";
import { NetworkConfig, parseYamlConfig } from "../utils";

export class EthereumProcessor {
  private config: NetworkConfig;
  private processor: EvmBatchProcessor<{ log: { transactionHash: true } }>;
  private chainId: number;
  private logger: Logger;

  private mappingHandler: MappingHandler;

  constructor() {
    this.mappingHandler = new MappingHandler();
    this.config = parseYamlConfig().ethereum;
    this.chainId = this.config.chain_id;
    this.logger = createLogger("sqd:evm-processor");

    this.logger.info(
      `Ethereum Processor initialized with:\n${JSON.stringify(
        this.config,
        null,
        2
      )}`
    );

    this.processor = new EvmBatchProcessor()
      .setBlockRange({ from: 0 })
      .setFinalityConfirmation(this.config.finality_confirmations)
      .addLog({
        address: [this.config.contract_address],
      });

    if (this.config.rpc_settings) {
      this.processor = this.processor
        .setRpcEndpoint({
          url: this.config.rpc_settings.endpoint_url,
          rateLimit: this.config.rpc_settings.rate_limit,
        })
        .setRpcDataIngestionSettings({
          headPollInterval: this.config.rpc_settings.head_poll_interval_s,
        });
    } else {
      if (!this.config.gateway_settings) {
        throw new Error(
          "Gateway settings are required because Gateway facilities not provider yet for Ethereum"
        );
      }
      this.processor = this.processor.setGateway(
        this.config.gateway_settings!.endpoint_url
      );
    }

    this.logger.info(
      `Ethereum Processor initialized with rpc endpoint ${this.config.chain_id}`
    );
  }

  private decodeTokenWrapped(log: Log) {
    try {
      const decoded = events.TokenWrapped.decode(log);
      return decoded;
    } catch (error) {
      return null;
    }
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
        Number(this.chainId)
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
        Number(this.chainId)
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
