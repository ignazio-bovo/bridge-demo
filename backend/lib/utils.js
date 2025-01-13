"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.PendingTxNotFoundError = exports.UpsertFailedError = exports.EntityNotFoundError = void 0;
exports.stripHexPrefix = stripHexPrefix;
exports.transferId = transferId;
exports.checkNativeToken = checkNativeToken;
const ethers_1 = require("ethers");
const ethers_2 = require("ethers");
function stripHexPrefix(hex) {
    if (hex.startsWith("0x")) {
        return hex.slice(2);
    }
    return hex;
}
function transferId(nonce, srcChainId) {
    return (0, ethers_1.keccak256)(ethers_2.ethers.toUtf8Bytes(`${nonce}-${srcChainId}`));
}
function checkNativeToken(tokenKey, chainId) {
    return ((stripHexPrefix(tokenKey) ===
        "414713f6f005b72ddaf8b453df42fc9c641dd532bc024d2b77ebd084408e2c7e" &&
        chainId === "1") ||
        (stripHexPrefix(tokenKey) ===
            "3a636391d72d0aec588d3a7908f5b3950fa7ac843ef46bf86ed066ba010044de" &&
            chainId === "945"));
}
class EntityNotFoundError extends Error {
    constructor(entityName, id) {
        super(`${entityName} not found with id: ${id}`);
        this.name = "EntityNotFoundError";
    }
}
exports.EntityNotFoundError = EntityNotFoundError;
class UpsertFailedError extends Error {
    constructor(entityName, id) {
        super(`Failed to upsert ${entityName} with id: ${id}`);
        this.name = "UpsertFailedError";
    }
}
exports.UpsertFailedError = UpsertFailedError;
class PendingTxNotFoundError extends Error {
    constructor(id) {
        super(`Pending transaction not found with id: ${id}`);
        this.name = "PendingTxNotFoundError";
    }
}
exports.PendingTxNotFoundError = PendingTxNotFoundError;
//# sourceMappingURL=utils.js.map