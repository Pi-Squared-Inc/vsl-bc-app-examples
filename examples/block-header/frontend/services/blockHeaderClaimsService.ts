import { BlockHeaderRecordsResponse } from "@/types/blockHeader";
import axios from "axios";

export const fetchBlockHeaderRecords = async (
  page: number = 0,
  pageSize: number = 25,
  chain?: string
): Promise<BlockHeaderRecordsResponse | null> => {
  try {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    });

    if (chain) {
      params.append("chain", chain);
    }

    const response = await axios.get(
      `${process.env.NEXT_PUBLIC_API_URL}/block_header_records?${params}`
    );
    return response.data;
  } catch (error) {
    console.error("Failed to fetch block header records:", error);
    return null;
  }
};

export interface ClaimData {
  [key: string]: any;
}

export const fetchClaimDetails = async (
  claimId: string
): Promise<ClaimData | null> => {
  try {
    const response = await axios.get(
      `${process.env.NEXT_PUBLIC_API_URL}/block_header_records/${claimId}/claim`
    );
    return response.data;
  } catch (error) {
    console.error(`Failed to fetch claim details for Claim ${claimId}:`, error);
    return null;
  }
};

export const fetchVerificationContext = async (
  claimId: string
): Promise<any | null> => {
  try {
    const response = await axios.get(
      `${process.env.NEXT_PUBLIC_API_URL}/block_header_records/${claimId}/verification_context`
    );
    return response.data;
  } catch (error) {
    console.error(
      `Failed to fetch verification context for Claim ${claimId}:`,
      error
    );
    return null;
  }
};
