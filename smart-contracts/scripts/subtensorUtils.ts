import { ApiPromise, WsProvider, Keyring } from "@polkadot/api";
import { ethers } from "ethers";
import { blake2AsU8a, encodeAddress } from "@polkadot/util-crypto";
import { hexToU8a } from "@polkadot/util";
import { SubmittableExtrinsic } from "@polkadot/api/types";
import { ISubmittableResult, AnyJson } from "@polkadot/types/types/";
import { AccountId, EventRecord } from "@polkadot/types/interfaces";
import { DispatchError } from "@polkadot/types/interfaces/system";
import { KeyringPair } from "@polkadot/keyring/types";
import AsyncLock from "async-lock";
import { assert } from "chai";
import BN from "bn.js";
import { evmToAddress } from "@polkadot/util-crypto";

export enum LogLevel {
  None,
  Debug,
  Verbose,
}

const nonceCacheByAccount = new Map<string, number>();

export class Sender {
  private readonly api: ApiPromise;
  static readonly asyncLock: AsyncLock = new AsyncLock();
  private readonly keyring: Keyring;

  constructor(api: ApiPromise, keyring: Keyring) {
    this.api = api;
    this.keyring = keyring;
  }

  // Synchronize all sending of transactions into mempool, so we can always safely read
  // the next account nonce taking mempool into account. This is safe as long as all sending of transactions
  // from same account occurs in the same process. Returns a promise of the Extrinsic Dispatch Result ISubmittableResult.
  // The promise resolves on tx finalization (For both Dispatch success and failure)
  // The promise is rejected if transaction is rejected by node.

  public async signAndSend(
    tx: SubmittableExtrinsic<"promise">,
    account: AccountId | string,
    tip?: BN | number
  ): Promise<ISubmittableResult> {
    const addr = this.keyring.encodeAddress(account);
    const senderKeyPair: KeyringPair = this.keyring.getPair(addr);

    let unsubscribe: () => void;
    let finalized: { (result: ISubmittableResult): void };
    const whenFinalized: Promise<ISubmittableResult> = new Promise(
      (resolve, reject) => {
        finalized = resolve;
      }
    );

    // saved human representation of the signed tx, will be set before it is submitted.
    // On error it is logged to help in debugging.
    let sentTx: AnyJson;

    const handleEvents = (result: ISubmittableResult) => {
      if (result.status.isFuture) {
        // Its virtually impossible for us to continue with tests
        // when this occurs and we don't expect the tests to handle this correctly
        // so just abort!
        console.error("Future Tx, aborting!");
        process.exit(-1);
      }

      if (!(result.status.isInBlock || result.status.isFinalized)) {
        return;
      }

      const success = result.findRecord("system", "ExtrinsicSuccess");
      const failed = result.findRecord("system", "ExtrinsicFailed");

      // Log failed transactions
      if (failed) {
        const record = failed as EventRecord;
        assert(record);
        const {
          event: { data },
        } = record;
        const err = data[0] as DispatchError;
        if (err.isModule) {
          try {
            this.api.registry.findMetaError(err.asModule);
          } catch (findmetaerror) {
            // example Error: findMetaError: Unable to find Error with index 0x1400/[{"index":20,"error":0}]
            // Happens for dispatchable calls that don't explicitly use `-> DispatchResult` return value even
            // if they return an error enum variant from the decl_error! macro
            console.log(
              "Dispatch Error (error details not found):",
              err.asModule.toHuman(),
              sentTx
            );
          }
        } else {
          console.log("Non-module Dispatch Error:", err.toHuman(), sentTx);
        }
      } else {
        assert(success);
      }

      // Always resolve irrespective of success or failure. Error handling should
      // be dealt with by caller.
      if (success || failed) {
        if (unsubscribe) {
          unsubscribe();
        }
        finalized(result);
      }
    };

    // We used to do this: Sender.asyncLock.acquire(`${senderKeyPair.address}` ...
    // Instead use a single lock for all calls, to force all transactions to be submitted in same order
    // of call to signAndSend. Otherwise it raises chance of race conditions.
    // It happens in rare cases and has lead some tests to fail occasionally in the past
    await Sender.asyncLock.acquire(
      ["tx-queue", `nonce-${account.toString()}`],
      async () => {
        // The node sometimes returns invalid account nonce at the exact time a new block is produced
        // For a split second the node will then not take "pending" transactions into account,
        // that's why we must partialy rely on cached nonce
        const nodeNonce = await this.api.rpc.system.accountNextIndex(
          senderKeyPair.address
        );

        const cachedNonce = nonceCacheByAccount.get(senderKeyPair.address);

        const nonce = BN.max(nodeNonce, new BN(cachedNonce || 0));

        const signedTx = await tx.signAsync(senderKeyPair, { nonce, tip });
        sentTx = signedTx.toHuman();
        const { method, section } = signedTx.method;
        try {
          unsubscribe = await signedTx.send(handleEvents);
          nonceCacheByAccount.set(account.toString(), nonce.toNumber() + 1);
        } catch (err) {
          throw err;
        }
      }
    );

    return whenFinalized;
  }
}

// hardhat network private keys being used in subtensor for local testing, DON'T USE THEM IN PRODUCTION
export const privateKeys = [
  "ac0974bec39a17e36ba4a6b4d23a8d103d857c13477d2c711b66614e1c7831a1", // account [0]
  "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a84a2f8e5924caa673d", // account [1]
  "5de4111afa1a4b949080356a392a688fdf4dda661ab95c820278f5f5ea0b5d57", // account [2]
  "7c852118294e51e653712a81e05800f419141751f6b64850b6c9593c3d2c6718", // account [3]
];

export const subtensorExtraConfig = {
  wsUrl: "ws://127.0.0.1:9944",
  httpUrl: "http://127.0.0.1:9944",
};

export async function fundSubtensorAccount(
  recipientEthereumAddress: string,
  taoAmount: number
) {
  const wsProvider = new WsProvider(subtensorExtraConfig.wsUrl);
  const api = await ApiPromise.create({ provider: wsProvider });
  const keyring = new Keyring({ type: "sr25519" });

  const sender = keyring.addFromUri("//Alice"); // Alice's default development account

  const ss58Address = evmToAddress(recipientEthereumAddress);

  // Amount to send: 1 TAO on Substrate side = 1*10^9
  const amount = BigInt(taoAmount) * 10n ** 9n;

  // Alice funds herself with 1M TAO
  const txSudoSetBalance = api.tx.sudo.sudo(
    api.tx.balances.forceSetBalance(sender.address, amount * 2n)
  );
  const txSender = new Sender(api, keyring);
  await txSender.signAndSend(txSudoSetBalance, sender.address);
  console.log("Balance force-set");

  // Create a transfer transaction
  const transfer = api.tx.balances.transferAllowDeath(ss58Address, amount);

  // Sign and send the transaction
  await txSender.signAndSend(transfer, sender.address);
  console.log(
    `Transfer sent to ${recipientEthereumAddress} (its ss58 mirror address is: ${ss58Address})`
  );

  await api.disconnect();
}

export async function sudoSetEvmChainId(chainId: number) {
  const wsProvider = new WsProvider(subtensorExtraConfig.wsUrl);
  const api = await ApiPromise.create({ provider: wsProvider });
  const keyring = new Keyring({ type: "sr25519" });

  const sender = keyring.addFromUri("//Alice"); // Alice's default development account

  const txSudoSetEvmChainId = api.tx.sudo.sudo(
    api.tx.adminUtils.sudoSetEvmChainId(chainId)
  );

  const txSender = new Sender(api, keyring);
  await txSender.signAndSend(txSudoSetEvmChainId, sender.address);
  console.log(`EVM chain ID set to ${chainId}`);

  await api.disconnect();
}
