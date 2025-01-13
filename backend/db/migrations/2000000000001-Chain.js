module.exports = class Chain2000000000001 {
    name = 'Chain2000000000001'

    async up(db) {
        // set up chains
        await db.query(`INSERT INTO "chain" ("id", "name") VALUES ('1', 'Ethereum')`);
        await db.query(`INSERT INTO "chain" ("id", "name") VALUES ('945', 'Subtensor')`);
        // // set up tokens using key = keccak256(DATURABRIDGE:tokenSymbol)
        // await db.query(`INSERT INTO "token" ("id", "name", "symbol", "decimals") VALUES ('3a636391d72d0aec588d3a7908f5b3950fa7ac843ef46bf86ed066ba010044de', 'TAO', 'TAO', '9')`);
        // await db.query(`INSERT INTO "token" ("id", "name", "symbol", "decimals") VALUES ('414713f6f005b72ddaf8b453df42fc9c641dd532bc024d2b77ebd084408e2c7e', 'Ether', 'ETH', '18')`);
        // // create token and chain relationship
        // await db.query(`INSERT INTO "token_with_chain" ("id", "address", "native", "token_id", "chain_id") VALUES ('3a636391d72d0aec588d3a7908f5b3950fa7ac843ef46bf86ed066ba010044de-945', '0x0000000000000000000000000000000000000000', true, '3a636391d72d0aec588d3a7908f5b3950fa7ac843ef46bf86ed066ba010044de', '945')`);
        // await db.query(`INSERT INTO "token_with_chain" ("id", "address", "native", "token_id", "chain_id") VALUES ('414713f6f005b72ddaf8b453df42fc9c641dd532bc024d2b77ebd084408e2c7e-1', '0x0000000000000000000000000000000000000000', true, '414713f6f005b72ddaf8b453df42fc9c641dd532bc024d2b77ebd084408e2c7e', '1')`);
    }
    async down(db) {
        await db.query(`DELETE * FROM "chain" `);
        await db.query(`DELETE * FROM "token" `);
    }
}
