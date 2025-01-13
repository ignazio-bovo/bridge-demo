"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.BridgeTxData = void 0;
const typeorm_store_1 = require("@subsquid/typeorm-store");
const chain_model_1 = require("./chain.model");
const token_model_1 = require("./token.model");
let BridgeTxData = class BridgeTxData {
    constructor(props) {
        Object.assign(this, props);
    }
};
exports.BridgeTxData = BridgeTxData;
__decorate([
    (0, typeorm_store_1.PrimaryColumn)(),
    __metadata("design:type", String)
], BridgeTxData.prototype, "id", void 0);
__decorate([
    (0, typeorm_store_1.Index)(),
    (0, typeorm_store_1.ManyToOne)(() => chain_model_1.Chain, { nullable: true }),
    __metadata("design:type", chain_model_1.Chain)
], BridgeTxData.prototype, "sourceChain", void 0);
__decorate([
    (0, typeorm_store_1.Index)(),
    (0, typeorm_store_1.ManyToOne)(() => chain_model_1.Chain, { nullable: true }),
    __metadata("design:type", chain_model_1.Chain)
], BridgeTxData.prototype, "destinationChain", void 0);
__decorate([
    (0, typeorm_store_1.StringColumn)({ nullable: false }),
    __metadata("design:type", String)
], BridgeTxData.prototype, "sourceAddress", void 0);
__decorate([
    (0, typeorm_store_1.StringColumn)({ nullable: false }),
    __metadata("design:type", String)
], BridgeTxData.prototype, "destinationAddress", void 0);
__decorate([
    (0, typeorm_store_1.BigIntColumn)({ nullable: false }),
    __metadata("design:type", BigInt)
], BridgeTxData.prototype, "amount", void 0);
__decorate([
    (0, typeorm_store_1.BigIntColumn)({ nullable: false }),
    __metadata("design:type", BigInt)
], BridgeTxData.prototype, "nonce", void 0);
__decorate([
    (0, typeorm_store_1.Index)(),
    (0, typeorm_store_1.ManyToOne)(() => token_model_1.Token, { nullable: true }),
    __metadata("design:type", token_model_1.Token)
], BridgeTxData.prototype, "token", void 0);
__decorate([
    (0, typeorm_store_1.BooleanColumn)({ nullable: false }),
    __metadata("design:type", Boolean)
], BridgeTxData.prototype, "confirmed", void 0);
__decorate([
    (0, typeorm_store_1.DateTimeColumn)({ nullable: false }),
    __metadata("design:type", Date)
], BridgeTxData.prototype, "createdAtTimestamp", void 0);
__decorate([
    (0, typeorm_store_1.DateTimeColumn)({ nullable: true }),
    __metadata("design:type", Object)
], BridgeTxData.prototype, "confirmedAtTimestamp", void 0);
exports.BridgeTxData = BridgeTxData = __decorate([
    (0, typeorm_store_1.Entity)(),
    __metadata("design:paramtypes", [Object])
], BridgeTxData);
//# sourceMappingURL=bridgeTxData.model.js.map