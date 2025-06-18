export interface ValidationRecord {
  block_number: number;
  claim_details?: {
    [clientName: string]: ValidationRecordClientDetails;
  };
}

export type ValidationRecordClientDetails = {
  created_at?: string;
  claim_id?: string;
  verification_time?: number;
  error?: string | null;
};
