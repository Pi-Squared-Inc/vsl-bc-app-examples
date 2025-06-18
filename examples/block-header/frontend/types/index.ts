/**
 * Types index file
 * Re-exports all types from various categorized files
 */

// Removed export * from "./claim"; // Assuming types are defined here now

// Renamed Claim to ValidationRecord
export interface ValidationRecord {
  claim_hash: string;
  claim: string | null;
  proof: string | null;
  verifier_address: string | null;
  submitter_address: string | null;
  claim_size: number | null;
  proof_size: number | null;
  submitted_timestamp: string | null;
  verified_timestamp: string | null;
  validation_time_seconds: number | null;
  status: string;
}

// Removed ValidationRecordClientDetails type definition
// export type ValidationRecordClientDetails = { ... };
