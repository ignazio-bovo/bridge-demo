import { keccak256, ethers } from "ethers";
import * as fs from "fs";
import path from "path";
import { parse } from "yaml";

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
  const ethMainnetChainId = "1";
  const ethLocalChainId = "31337";
  const ethSepoliaChainId = "11155111";
  const subtensorChainId = "945";
  const subtensorLocalChainId = "3138";
  const subtensorTestnetChainId = "945";
  return (
    (stripHexPrefix(tokenKey) ===
      "414713f6f005b72ddaf8b453df42fc9c641dd532bc024d2b77ebd084408e2c7e" &&
      (chainId === ethMainnetChainId ||
        chainId === ethLocalChainId ||
        chainId === ethSepoliaChainId)) ||
    (stripHexPrefix(tokenKey) ===
      "3a636391d72d0aec588d3a7908f5b3950fa7ac843ef46bf86ed066ba010044de" &&
      (chainId === subtensorTestnetChainId ||
        chainId === subtensorLocalChainId ||
        chainId === subtensorChainId))
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

export interface NetworkConfig {
  rpc_settings?: {
    endpoint_url: string;
    rate_limit: number;
    head_poll_interval_s: number;
  };
  gateway_settings?: {
    endpoint_url: string;
  };
  contract_address: string;
  chain_id: number;
  start_block: number;
  finality_confirmations: number;
}

export interface YamlConfig {
  ethereum: NetworkConfig;
  subtensor: NetworkConfig;
}

export function parseYamlConfig(): YamlConfig {
  const file = fs.readFileSync(
    path.join(__dirname, "..", "config", "networks.yaml"),
    "utf8"
  );
  try {
    return parse(file);
  } catch (error) {
    throw new Error("Failed to parse YAML config");
  }
}
