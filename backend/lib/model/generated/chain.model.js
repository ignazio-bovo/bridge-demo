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
exports.Chain = void 0;
const typeorm_store_1 = require("@subsquid/typeorm-store");
const tokenWithChain_model_1 = require("./tokenWithChain.model");
const bridgeTxData_model_1 = require("./bridgeTxData.model");
let Chain = class Chain {
    constructor(props) {
        Object.assign(this, props);
    }
};
exports.Chain = Chain;
__decorate([
    (0, typeorm_store_1.PrimaryColumn)(),
    __metadata("design:type", String)
], Chain.prototype, "id", void 0);
__decorate([
    (0, typeorm_store_1.StringColumn)({ nullable: false }),
    __metadata("design:type", String)
], Chain.prototype, "name", void 0);
__decorate([
    (0, typeorm_store_1.OneToMany)(() => tokenWithChain_model_1.TokenWithChain, e => e.chain),
    __metadata("design:type", Array)
], Chain.prototype, "tokens", void 0);
__decorate([
    (0, typeorm_store_1.OneToMany)(() => bridgeTxData_model_1.BridgeTxData, e => e.destinationChain),
    __metadata("design:type", Array)
], Chain.prototype, "incomingTransfers", void 0);
__decorate([
    (0, typeorm_store_1.OneToMany)(() => bridgeTxData_model_1.BridgeTxData, e => e.sourceChain),
    __metadata("design:type", Array)
], Chain.prototype, "outgoingTransfers", void 0);
exports.Chain = Chain = __decorate([
    (0, typeorm_store_1.Entity)(),
    __metadata("design:paramtypes", [Object])
], Chain);
//# sourceMappingURL=chain.model.js.map