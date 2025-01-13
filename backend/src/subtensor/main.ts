import { SubtensorProcessor } from "./processor";
import { TypeormDatabase } from "@subsquid/typeorm-store";

// start both processors in parallel with their respective databases
new SubtensorProcessor().run(
  new TypeormDatabase({
    stateSchema: "squid_subtensor",
    supportHotBlocks: false,
  })
);
