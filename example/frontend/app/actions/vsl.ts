import { SignatureComponents } from "../../utils/signature";
import { newHTTPClient } from "./client";

export function newVSLClient() {
  return newHTTPClient(process.env.NEXT_PUBLIC_VSL_API_URL);
}

interface VSLError {
  code: number;
  message: string;
  data?: unknown;
}

interface VSLCallResponse<T> {
  jsonrpc: string;
  id: number;
  result?: T;
  error?: VSLError;
}

export async function call<T>(method: string, params: unknown): Promise<T> {
  const httpClient = newVSLClient();
  const response = await httpClient.Post<VSLCallResponse<T>>(`/`, {
    jsonrpc: "2.0",
    method,
    params,
    id: 1,
  });

  if (response.error) {
    throw new Error(
      `RPC Error: ${response.error.message} (code: ${response.error.code})` +
        (response.error.data
          ? `\nData: ${JSON.stringify(response.error.data)}`
          : "")
    );
  }

  if (response.result === undefined) {
    throw new Error("Invalid RPC response: missing 'result' field.");
  }

  return response.result;
}

export function getBalance(address: string) {
  return call<string>("vsl_getBalance", { account_id: address });
}

export function getAccountNonce(address: string) {
  return call<number>("vsl_getAccountNonce", { account_id: address });
}

export function pay(
  from: string,
  to: string,
  amount: string,
  nonce: string,
  signatureComponents: SignatureComponents
) {
  return call<string>("vsl_pay", {
    payment: {
      from,
      to,
      amount,
      nonce,
      ...signatureComponents,
    },
  });
}
