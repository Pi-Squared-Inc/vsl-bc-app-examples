import { ByteArray, hashMessage, Hex, SignableMessage, toRlp } from "viem";

// Defines the structure of the signature components that will be returned.
export interface SignatureComponents {
  hash: `0x${string}`;
  r: `0x${string}`;
  s: `0x${string}`;
  v: number;
}

export type RecursiveArray<T> = (T | RecursiveArray<T>)[];

/**
 * Generates a signature and its components for a structured message using RLP encoding.
 * This function orchestrates the process of RLP-encoding an array, hashing it
 * according to EIP-191, signing it via a provided signMessage function,
 * and finally deconstructing the signature into r, s, and v.
 *
 * @param signMessageAsync The async signMessage function obtained from wagmi's useSignMessage hook.
 * @param message The array of data to be RLP-encoded and signed. This can be a nested array of Hex strings or ByteArrays.
 * @returns A promise that resolves to an object containing the messageHash, signature, and its r, s, v components.
 */
export async function generateSignatureComponents(
  signMessageAsync: (args: { message: SignableMessage }) => Promise<Hex>,
  message: RecursiveArray<ByteArray> | RecursiveArray<Hex>
): Promise<SignatureComponents> {
  // 1. RLP-encode the array.
  const rlpEncoded = toRlp(message);

  // 2. Calculate the EIP-191 hash of the RLP-encoded bytes. This must match what the wallet signs.
  // The signing function will perform its own hashing internally.
  const hash = hashMessage({ raw: rlpEncoded });

  // 3. Sign the RLP-encoded data by calling the provided mutation function.
  // The wallet will handle the EIP-191 hashing internally.
  const signature = await signMessageAsync({ message: { raw: rlpEncoded } });

  // 4. Deconstruct the signature into r, s, and v components.
  const r = `0x${signature.substring(2, 66)}` as const;
  const s = `0x${signature.substring(66, 130)}` as const;
  // The V value from personal_sign is already 27 or 28, but in hex.
  const vHex = signature.substring(130, 132);
  const v = parseInt(vHex, 16);

  return {
    hash,
    r,
    s,
    v,
  };
}

export async function hashStringSha256(message: string): Promise<string> {
  const textEncoder = new TextEncoder();
  const data = textEncoder.encode(message);
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);
  const byteArray = Array.from(new Uint8Array(hashBuffer));
  const hexHash = byteArray.map(b => b.toString(16).padStart(2, '0')).join('');
  return hexHash;
}

export async function hashFileSha256(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = async (event) => {
      try {
        const arrayBuffer = event.target?.result as ArrayBuffer;
        const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        const hexHash = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
        resolve(hexHash);
      } catch (error) {
        reject(error);
      }
    };
    reader.onerror = (error) => {
      reject(error);
    };
    reader.readAsArrayBuffer(file);
  });
}
