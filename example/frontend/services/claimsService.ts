import axios from "axios";

export const fetchClaimDetails = async (
  blockNumber: number,
  executionClient: string
) => {
  try {
    const response = await axios.get(
      `${process.env.NEXT_PUBLIC_API_URL}/claims/${blockNumber}/${executionClient}/claim`
    );
    return response.data;
  } catch (error) {
    console.error(
      `Failed to fetch claim details for Block ${blockNumber}:`,
      error
    );
    return null;
  }
};

export const fetchVerificationContext = async (
  blockNumber: number,
  executionClient: string
) => {
  try {
    const response = await axios.get(
      `${process.env.NEXT_PUBLIC_API_URL}/claims/${blockNumber}/${executionClient}/verification_context`
    );
    return response.data;
  } catch (error) {
    console.error(
      `Failed to fetch verification context for Block ${blockNumber}:`,
      error
    );
    return null;
  }
};
