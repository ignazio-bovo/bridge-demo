module.exports = class Chain2000000000001 {
    name = 'Chain2000000000001'

    async up(db) {
        // set up chains
        await db.query(`INSERT INTO "chain" ("id", "name") VALUES ('1', 'Ethereum')`);
        await db.query(`INSERT INTO "chain" ("id", "name") VALUES ('31337', 'Ethereum')`);
        await db.query(`INSERT INTO "chain" ("id", "name") VALUES ('0', 'Subtensor')`);
        await db.query(`INSERT INTO "chain" ("id", "name") VALUES ('945', 'Subtensor')`);
    }
    async down(db) {
        await db.query(`DELETE * FROM "chain" `);
        await db.query(`DELETE * FROM "token" `);
    }
}
