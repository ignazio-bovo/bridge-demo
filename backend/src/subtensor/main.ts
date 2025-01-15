import { SubtensorProcessor } from "./processor";
import { TypeormDatabase } from "@subsquid/typeorm-store";

new SubtensorProcessor().run(
  new TypeormDatabase({
    stateSchema: "squid_subtensor",
    supportHotBlocks: false,
  })
);
