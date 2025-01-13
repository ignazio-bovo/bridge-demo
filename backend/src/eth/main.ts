import { EthereumProcessor } from "./processor";
import { TypeormDatabase } from "@subsquid/typeorm-store";

// start both processors in parallel with their respective databases
new EthereumProcessor().run(
  new TypeormDatabase({
    stateSchema: "squid_ethereum",
    supportHotBlocks: false,
  })
);
