"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.TokenMetadataDecodingError = void 0;
exports.decodeTokenMetadata = decodeTokenMetadata;
const ethers_1 = require("ethers");
class TokenMetadataDecodingError extends Error {
    constructor(metadata, message) {
        super(`Error decoding token metadata:\n\n${metadata}\n\n${message}`);
        this.name = "TokenMetadataDecodingError";
    }
}
exports.TokenMetadataDecodingError = TokenMetadataDecodingError;
function decodeTokenMetadata(encodedMetadata) {
    try {
        const decoded = ethers_1.ethers.AbiCoder.defaultAbiCoder().decode(["string", "string", "uint8"], encodedMetadata);
        // You can access the values like this:
        const [symbol, name, decimals] = decoded;
        return {
            symbol,
            name,
            decimals,
        };
    }
    catch (error) {
        throw new TokenMetadataDecodingError(encodedMetadata, "Error decoding token metadata");
    }
}
//# sourceMappingURL=decoding.js.map