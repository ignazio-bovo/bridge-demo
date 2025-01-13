"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.EthereumProcessor = void 0;
const evm_processor_1 = require("@subsquid/evm-processor");
const logger_1 = require("@subsquid/logger");
const handler_1 = require("../mappings/handler");
class EthereumProcessor {
    constructor() {
        this.chainId = 1;
        this.CONTRACT_ADDRESS = process.env.CONTRACT_ADDRESS ||
            "0x71c95911e9a5d330f4d621842ec243ee1343292e";
        this.mappingHandler = new handler_1.MappingHandler();
        this.logger = (0, logger_1.createLogger)("sqd:evm-processor");
        this.processor = new evm_processor_1.EvmBatchProcessor()
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
        this.logger.info(`Ethereum Processor initialized with rpc endpoint 127.0.0.1:8545`);
    }
    async run(db) {
        this.processor.run(db, async (ctx) => {
            for (let block of ctx.blocks) {
                for (let log of block.logs) {
                    await this.runEtl(log, ctx.store);
                }
            }
        });
    }
    async runEtl(log, store) {
        await this.mappingHandler.handleTokenWhitelistStatusUpdated(log, store, this.chainId);
        await this.mappingHandler.handleTransferRequested(log, store);
        await this.mappingHandler.handleTransferRequestExecuted(log, store);
        await this.mappingHandler.tokenWrapped(log, store, this.chainId);
    }
}
exports.EthereumProcessor = EthereumProcessor;
//# sourceMappingURL=processor.js.map