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
exports.TokenWithChain = void 0;
const typeorm_store_1 = require("@subsquid/typeorm-store");
const token_model_1 = require("./token.model");
const chain_model_1 = require("./chain.model");
let TokenWithChain = class TokenWithChain {
    constructor(props) {
        Object.assign(this, props);
    }
};
exports.TokenWithChain = TokenWithChain;
__decorate([
    (0, typeorm_store_1.PrimaryColumn)(),
    __metadata("design:type", String)
], TokenWithChain.prototype, "id", void 0);
__decorate([
    (0, typeorm_store_1.ManyToOne)(() => token_model_1.Token, { nullable: true }),
    __metadata("design:type", token_model_1.Token)
], TokenWithChain.prototype, "token", void 0);
__decorate([
    (0, typeorm_store_1.Index)(),
    (0, typeorm_store_1.ManyToOne)(() => chain_model_1.Chain, { nullable: true }),
    __metadata("design:type", chain_model_1.Chain)
], TokenWithChain.prototype, "chain", void 0);
__decorate([
    (0, typeorm_store_1.StringColumn)({ nullable: false }),
    __metadata("design:type", String)
], TokenWithChain.prototype, "address", void 0);
__decorate([
    (0, typeorm_store_1.BooleanColumn)({ nullable: false }),
    __metadata("design:type", Boolean)
], TokenWithChain.prototype, "native", void 0);
exports.TokenWithChain = TokenWithChain = __decorate([
    (0, typeorm_store_1.Index)(["token", "chain"], { unique: false }),
    (0, typeorm_store_1.Entity)(),
    __metadata("design:paramtypes", [Object])
], TokenWithChain);
//# sourceMappingURL=tokenWithChain.model.js.map