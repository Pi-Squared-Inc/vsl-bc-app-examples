/**
 * Claim-related types
 */

/**
 * Represents a claim with its metadata and validation information
 */
export interface Claim {
  /** Unique identifier for the claim */
  id: string;

  /** Block number/height in the blockchain */
  block_number: number;

  /** ISO timestamp when the claim was created */
  created_at: string;

  /** Validation time in microseconds, null if not validated */
  validation_time: number | null;

  /** Error message if claim generation failed, null if successful */
  generation_error: string | null;

  /** Error message if claim validation failed, null if successful */
  validation_error: string | null;

  /** Size of the claim in bytes */
  claim_size: number | null;

  /** Size of the verification context in bytes */
  verification_context_size: number | null;
}
