"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SubtensorProcessor = void 0;
const logger_1 = require("@subsquid/logger");
const evm_processor_1 = require("@subsquid/evm-processor");
class SubtensorProcessor {
    constructor() {
        this.logger = (0, logger_1.createLogger)("sqd:evm-processor");
        // First we configure data retrieval.
        this.processor = new evm_processor_1.EvmBatchProcessor()
            .setRpcEndpoint("http://127.0.0.1:8545")
            .setFinalityConfirmation(15)
            .addLog({ range: { from: 1 } });
        // this.processor = new SubstrateBatchProcessor()
        //   .setRpcEndpoint("http://127.0.0.1:9944")
        //   .setFinalityConfirmation(15)
        //   .addEthereumTransaction({})
        //   .addEvmLog({});
        this.logger.info(`Subtensor Processor initialized with rpc endpoint 127.0.0.1:9944`);
    }
    async runEtl(log, store) {
        this.logger.info(`Processing log for block ${log.block.height} with topic ${log.topics[0]}`);
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
}
exports.SubtensorProcessor = SubtensorProcessor;
//# sourceMappingURL=processor.js.map