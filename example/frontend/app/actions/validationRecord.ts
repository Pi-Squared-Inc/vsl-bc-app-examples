import { newHTTPClient } from "./client";

export function checkCanSubmit(
  address: string,
  compute_type: string
) {
  const httpClient = newHTTPClient(process.env.NEXT_PUBLIC_API_URL);
  return httpClient.Get(`/can_submit/`+compute_type+'/'+address)
}

export function getBackendAddress() {
  const httpClient = newHTTPClient(process.env.NEXT_PUBLIC_API_URL);
  return httpClient.Get(`/get_backend_address`);
}
