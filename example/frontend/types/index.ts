/**
 * Types index file
 * Re-exports all types from various categorized files
 */

// Removed export * from "./claim"; // Assuming types are defined here now

// Renamed Claim to ValidationRecord
export interface ValidationRecord {
  id: number;
  claim_size: number | null;
  verification_context_size: number | null;
  validation_time: number | null;
  validation_error: string | null;
  created_at: string;
  claim?: string;
  proof?: string;
  result?: string | null;
}

// Removed ValidationRecordClientDetails type definition
// export type ValidationRecordClientDetails = { ... };

export interface VerificationRecord {
  id: string;
  created_at: string;
  status: string;
  type: string;
  claim_id?: string;
  result?: string;
  error?: string;
}