import { SubmittableExtrinsic } from "@polkadot/api/types";
import { KeyringPair } from "@polkadot/keyring/types";
import { Keyring } from "@polkadot/keyring";
import { ApiPromise, WsProvider } from "@polkadot/api";
import { ethers } from "ethers";
import { evmToAddress } from "@polkadot/util-crypto";

export async function SubtensorFaucet(
  to: string,
  amount: number
): Promise<string> {
  if (!ethers.isAddress(to)) {
    throw new Error("Invalid address");
  }
  const wsProvider = new WsProvider("ws://127.0.0.1:9944");
  const api = await ApiPromise.create({ provider: wsProvider });
  const tk = getTestKeys();
  console.log("keying backend:", tk.alice.address);
  const realAmount = BigInt(amount * 10 ** 12); // Convert to the smallest unit
  const transfer = api.tx.balances.transferAllowDeath(
    evmToAddress(to),
    realAmount
  );
  const txHash = await sendTransaction(api, transfer, tk.alice);
  return txHash.toString();
}

function sendTransaction(
  api: ApiPromise,
  call: SubmittableExtrinsic<"promise">,
  signer: KeyringPair
): Promise<import("@polkadot/types/interfaces").BlockHash> {
  return new Promise((resolve, reject) => {
    let unsubscribed = false;
    const unsubscribe = call
      .signAndSend(signer, ({ status, events, dispatchError }) => {
        const safelyUnsubscribe = () => {
          if (!unsubscribed) {
            unsubscribed = true;
            unsubscribe
              .then(() => {})
              .catch((error) => console.error("Failed to unsubscribe:", error));
          }
        };

        // Check for transaction errors
        if (dispatchError) {
          let errout = dispatchError.toString();
          if (dispatchError.isModule) {
            // for module errors, we have the section indexed, lookup
            const decoded = api.registry.findMetaError(dispatchError.asModule);
            const { docs, name, section } = decoded;
            errout = `${section}.${name}: ${docs.join(" ")}`;
          }
          safelyUnsubscribe();
          reject(errout);
        }
        // Log and resolve when the transaction is included in a block
        console.log(`Transaction status: ${status}`);
        if (status.isInBlock) {
          safelyUnsubscribe();
          resolve(status.asInBlock);
        }
      })
      .catch((error) => {
        reject(error);
      });
  });
}

function getTestKeys() {
  const keyring = new Keyring({ type: "sr25519" });
  return {
    alice: keyring.addFromUri("//Alice"),
    aliceHot: keyring.addFromUri("//AliceHot"),
    bob: keyring.addFromUri("//Bob"),
    bobHot: keyring.addFromUri("//BobHot"),
    charlie: keyring.addFromUri("//Charlie"),
    charlieHot: keyring.addFromUri("//CharlieHot"),
    dave: keyring.addFromUri("//Dave"),
    daveHot: keyring.addFromUri("//DaveHot"),
    eve: keyring.addFromUri("//Eve"),
    zari: keyring.addFromUri("//Zari"),
  };
}
