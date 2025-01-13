import { ethers } from "ethers";

export interface TokenMetadata {
  symbol: string;
  name: string;
  decimals: number;
}

export class TokenMetadataDecodingError extends Error {
  constructor(metadata: string, message: string) {
    super(`Error decoding token metadata:\n\n${metadata}\n\n${message}`);
    this.name = "TokenMetadataDecodingError";
  }
}

export function decodeTokenMetadata(encodedMetadata: string): TokenMetadata {
  try {
    const decoded = ethers.AbiCoder.defaultAbiCoder().decode(
      ["string", "string", "uint8"],
      encodedMetadata
    );

    // You can access the values like this:
    const [symbol, name, decimals] = decoded;

    return {
      symbol,
      name,
      decimals,
    };
  } catch (error) {
    throw new TokenMetadataDecodingError(
      encodedMetadata,
      "Error decoding token metadata"
    );
  }
}
