import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { ValidationRecord } from "@/types";
import { formatBytes } from "@/utils";

interface ClientProofCellProps {
  details: ValidationRecord["claim_details"];
  blockNumber: number;
  executionClient: string;
  onViewData: (
    title: string,
    blockNumber: number,
    type: "verification",
    executionClient: string,
    size?: number | null
  ) => void;
  fallbackVerificationSize?: number | null;
}

export function ClientProofCell({
  // details,
  blockNumber,
  executionClient,
  onViewData,
  fallbackVerificationSize,
}: ClientProofCellProps) {
  const verificationSize = 0;
  const effectiveSize =
    verificationSize !== undefined
      ? verificationSize
      : fallbackVerificationSize;
  const displaySize =
    effectiveSize !== undefined && effectiveSize !== null
      ? formatBytes(effectiveSize)
      : "—";
  const canView = effectiveSize !== undefined && effectiveSize !== null;

  return (
    <TableCell className="px-4 align-middle">
      <div className="flex items-center gap-3">
        <Button
          size="sm"
          onClick={() =>
            onViewData(
              "Proof",
              blockNumber,
              "verification",
              executionClient,
              effectiveSize
            )
          }
          className="flex-shrink-0 bg-pi2-accent-white text-pi2-accent-black hover:bg-pi2-accent-white/90"
          disabled={!canView}
        >
          View
        </Button>
        <div className="text-sm text-pi2-accent-white font-normal">
          {displaySize}
        </div>
      </div>
    </TableCell>
  );
}
