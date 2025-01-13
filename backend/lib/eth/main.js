"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const processor_1 = require("./processor");
const typeorm_store_1 = require("@subsquid/typeorm-store");
// start both processors in parallel with their respective databases
new processor_1.EthereumProcessor().run(new typeorm_store_1.TypeormDatabase({
    stateSchema: "squid_ethereum",
    supportHotBlocks: false,
}));
//# sourceMappingURL=main.js.map