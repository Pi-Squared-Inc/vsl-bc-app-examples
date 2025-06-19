export interface BlockHeaderDetails {
  claim_id: string;
  chain: string;
  created_at: string;
  verification_time: number | null;
  error: string | null;
}

export interface BlockHeaderRecord {
  block_number: number;
  claim_details: BlockHeaderDetails;
}

export interface BlockHeaderRecordsResponse {
  block_header_records: BlockHeaderRecord[];
  total: number;
}
