import { Store, TypeormDatabase } from "@subsquid/typeorm-store";
import { SubstrateBatchProcessor } from "@subsquid/substrate-processor";
import { Logger, createLogger } from "@subsquid/logger";
import { MappingHandler } from "../mappings/handler";
import { events } from "../abi/Bridge.sol";
import { EventRecord } from "@subsquid/evm-abi";
import { NetworkConfig, parseYamlConfig } from "../utils";

export class SubtensorProcessor {
  private config: NetworkConfig;
  private processor: SubstrateBatchProcessor;
  private logger: Logger;
  private chainId: number;
  private handler: MappingHandler;

  constructor() {
    this.logger = createLogger("sqd:subtensor-processor");
    this.handler = new MappingHandler();
    this.config = parseYamlConfig().subtensor;
    this.chainId = this.config.chain_id;

    if (!this.config.rpc_settings) {
      throw new Error(
        "RPC settings are required because Gateway facilities not provider yet for Bittensor EVM"
      );
    }

    this.logger.info(
      `Subtensor Processor initialized with:\n${JSON.stringify(
        this.config,
        null,
        2
      )}`
    );

    this.processor = new SubstrateBatchProcessor()
      .setBlockRange({ from: this.config.start_block })
      .setFinalityConfirmation(this.config.finality_confirmations)
      .setRpcEndpoint(this.config.rpc_settings.endpoint_url)
      .setRpcDataIngestionSettings({
        headPollInterval: this.config.rpc_settings.head_poll_interval_s,
      })
      .addEvmLog({
        address: [this.config.contract_address],
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
      `Subtensor Processor initialized with rpc endpoint ${this.config.rpc_settings.endpoint_url}`
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
