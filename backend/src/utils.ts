import { keccak256 } from "ethers";

import { ethers } from "ethers";

export function stripHexPrefix(hex: string): string {
  if (hex.startsWith("0x")) {
    return hex.slice(2);
  }
  return hex;
}
export function transferId(nonce: number, srcChainId: number): string {
  return keccak256(ethers.toUtf8Bytes(`${nonce}:${srcChainId}`));
}
export function checkNativeToken(tokenKey: string, chainId: string): boolean {
  return (
    (stripHexPrefix(tokenKey) ===
      "414713f6f005b72ddaf8b453df42fc9c641dd532bc024d2b77ebd084408e2c7e" &&
      chainId === "1") ||
    (stripHexPrefix(tokenKey) ===
      "3a636391d72d0aec588d3a7908f5b3950fa7ac843ef46bf86ed066ba010044de" &&
      chainId === "945")
  );
}

export class EntityNotFoundError extends Error {
  constructor(entityName: string, id: string) {
    super(`${entityName} not found with id: ${id}`);
    this.name = "EntityNotFoundError";
  }
}

export class UpsertFailedError extends Error {
  constructor(entityName: string, id: string) {
    super(`Failed to upsert ${entityName} with id: ${id}`);
    this.name = "UpsertFailedError";
  }
}

export class PendingTxNotFoundError extends Error {
  constructor(id: string) {
    super(`Pending transaction not found with id: ${id}`);
    this.name = "PendingTxNotFoundError";
  }
}
