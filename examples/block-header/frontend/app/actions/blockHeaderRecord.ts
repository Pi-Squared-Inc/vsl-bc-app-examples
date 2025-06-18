// blockHeaderRecord.ts
import { BlockHeaderRecordsResponse } from "@/types/blockHeader";
import { createAlova } from "alova";
import adapterFetch from "alova/fetch";

export function newHTTPClient(url: string | undefined) {
  return createAlova({
    baseURL: url,
    requestAdapter: adapterFetch(),
    responded: (response) => {
      const json = response.json();
      return json;
    },
  });
}

export interface BlockHeaderRecordParams {
  page?: number;
  per_page?: number;
  chain?: string;
}

export function createBlockHeaderRecordsAPI(baseURL: string) {
  const client = newHTTPClient(baseURL);

  return {
    getBlockHeaderRecords: (params: BlockHeaderRecordParams = {}) => {
      const searchParams = new URLSearchParams();
      if (params.page !== undefined)
        searchParams.set("page", params.page.toString());
      if (params.per_page)
        searchParams.set("per_page", params.per_page.toString());
      if (params.chain) searchParams.set("chain", params.chain);

      const queryString = searchParams.toString();
      const endpoint = queryString
        ? `/block_header_records?${queryString}`
        : "/block_header_records";

      return client.Get<BlockHeaderRecordsResponse>(endpoint);
    },
  };
}
