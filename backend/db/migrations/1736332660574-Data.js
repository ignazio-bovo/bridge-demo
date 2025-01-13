module.exports = class Data1736332660574 {
    name = 'Data1736332660574'

    async up(db) {
        await db.query(`CREATE TABLE "bridge_tx_data" ("id" character varying NOT NULL, "source_address" text NOT NULL, "destination_address" text NOT NULL, "amount" numeric NOT NULL, "nonce" numeric NOT NULL, "confirmed" boolean NOT NULL, "created_at_timestamp" TIMESTAMP WITH TIME ZONE NOT NULL, "confirmed_at_timestamp" TIMESTAMP WITH TIME ZONE, "source_chain_id" character varying, "destination_chain_id" character varying, "token_id" character varying, CONSTRAINT "PK_acaad2663a9de02abb07f5ef65a" PRIMARY KEY ("id"))`)
        await db.query(`CREATE INDEX "IDX_f0daf72849d48ca5cf52b9b725" ON "bridge_tx_data" ("source_chain_id") `)
        await db.query(`CREATE INDEX "IDX_940bc82a28a16003c38a85ad15" ON "bridge_tx_data" ("destination_chain_id") `)
        await db.query(`CREATE INDEX "IDX_cb96530bf77115fde100729ea4" ON "bridge_tx_data" ("token_id") `)
        await db.query(`CREATE TABLE "chain" ("id" character varying NOT NULL, "name" text NOT NULL, CONSTRAINT "PK_8e273aafae283b886672c952ecd" PRIMARY KEY ("id"))`)
        await db.query(`CREATE TABLE "token_with_chain" ("id" character varying NOT NULL, "address" text NOT NULL, "native" boolean NOT NULL, "token_id" character varying, "chain_id" character varying, CONSTRAINT "PK_3103e309e216bca308a5cda79bf" PRIMARY KEY ("id"))`)
        await db.query(`CREATE INDEX "IDX_09aafb872652b431435ce038cf" ON "token_with_chain" ("chain_id") `)
        await db.query(`CREATE INDEX "IDX_6ea7626820fd50075198409920" ON "token_with_chain" ("token_id", "chain_id") `)
        await db.query(`CREATE TABLE "token" ("id" character varying NOT NULL, "name" text, "symbol" text, "decimals" integer NOT NULL, CONSTRAINT "PK_82fae97f905930df5d62a702fc9" PRIMARY KEY ("id"))`)
        await db.query(`ALTER TABLE "bridge_tx_data" ADD CONSTRAINT "FK_f0daf72849d48ca5cf52b9b725e" FOREIGN KEY ("source_chain_id") REFERENCES "chain"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`)
        await db.query(`ALTER TABLE "bridge_tx_data" ADD CONSTRAINT "FK_940bc82a28a16003c38a85ad150" FOREIGN KEY ("destination_chain_id") REFERENCES "chain"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`)
        await db.query(`ALTER TABLE "bridge_tx_data" ADD CONSTRAINT "FK_cb96530bf77115fde100729ea41" FOREIGN KEY ("token_id") REFERENCES "token"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`)
        await db.query(`ALTER TABLE "token_with_chain" ADD CONSTRAINT "FK_9ef60856e5445dbb2ac08d3bca7" FOREIGN KEY ("token_id") REFERENCES "token"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`)
        await db.query(`ALTER TABLE "token_with_chain" ADD CONSTRAINT "FK_09aafb872652b431435ce038cfb" FOREIGN KEY ("chain_id") REFERENCES "chain"("id") ON DELETE NO ACTION ON UPDATE NO ACTION`)
    }

    async down(db) {
        await db.query(`DROP TABLE "bridge_tx_data"`)
        await db.query(`DROP INDEX "public"."IDX_f0daf72849d48ca5cf52b9b725"`)
        await db.query(`DROP INDEX "public"."IDX_940bc82a28a16003c38a85ad15"`)
        await db.query(`DROP INDEX "public"."IDX_cb96530bf77115fde100729ea4"`)
        await db.query(`DROP TABLE "chain"`)
        await db.query(`DROP TABLE "token_with_chain"`)
        await db.query(`DROP INDEX "public"."IDX_09aafb872652b431435ce038cf"`)
        await db.query(`DROP INDEX "public"."IDX_6ea7626820fd50075198409920"`)
        await db.query(`DROP TABLE "token"`)
        await db.query(`ALTER TABLE "bridge_tx_data" DROP CONSTRAINT "FK_f0daf72849d48ca5cf52b9b725e"`)
        await db.query(`ALTER TABLE "bridge_tx_data" DROP CONSTRAINT "FK_940bc82a28a16003c38a85ad150"`)
        await db.query(`ALTER TABLE "bridge_tx_data" DROP CONSTRAINT "FK_cb96530bf77115fde100729ea41"`)
        await db.query(`ALTER TABLE "token_with_chain" DROP CONSTRAINT "FK_9ef60856e5445dbb2ac08d3bca7"`)
        await db.query(`ALTER TABLE "token_with_chain" DROP CONSTRAINT "FK_09aafb872652b431435ce038cfb"`)
    }
}
