import { createAlova } from "alova";
import adapterFetch from "alova/fetch";

export function newHTTPClient(url: string | undefined) {
  return createAlova({
    baseURL: url,
    requestAdapter: adapterFetch(),
    responded: (response) => response.json(),
  });
}
